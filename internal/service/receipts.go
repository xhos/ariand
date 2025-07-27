package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"ariand/internal/db"
	"ariand/internal/domain"
	"ariand/internal/receiptparser"

	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx/types"
)

// MatchResult represents a scored transaction candidate
type MatchResult struct {
	Tx         *domain.TransactionWithScore
	FinalScore float64
}

type ReceiptService interface {
	LinkManual(ctx context.Context, transactionID int64, file io.Reader, filename string, provider domain.ReceiptProvider) (*domain.Receipt, error)
	MatchAndSuggest(ctx context.Context, file io.Reader, filename string, provider domain.ReceiptProvider) (*domain.Receipt, error)
}

type receiptSvc struct {
	store  db.Store
	parser receiptparser.Client
	log    *log.Logger
}

func newReceiptSvc(store db.Store, parser receiptparser.Client, lg *log.Logger) ReceiptService {
	return &receiptSvc{store: store, parser: parser, log: lg}
}

// LinkManual manually links a receipt to a specific transaction
func (s *receiptSvc) LinkManual(ctx context.Context, transactionID int64, file io.Reader, filename string, provider domain.ReceiptProvider) (*domain.Receipt, error) {
	data, imageHash, err := s.readAndStoreImage(ctx, file, filename)
	if err != nil {
		return nil, err
	}

	receipt := &domain.Receipt{
		TransactionID: &transactionID,
		Provider:      provider,
		ImageSHA256:   imageHash,
		LinkStatus:    domain.LinkStatusMatched,
	}

	parsed, raw, parseErr := s.parser.Parse(ctx, bytes.NewReader(data), filename, provider)
	receipt.RawPayload = raw

	if parseErr != nil {
		s.log.Warn("parser failed on manual link", "err", parseErr, "txID", transactionID)
		receipt.ParseStatus = domain.StatusFailed
	} else {
		s.log.Info("parser succeeded on manual link", "merchant", parsed.Merchant, "txID", transactionID)
		receipt.ParseStatus = domain.StatusParsed
		s.populateReceiptFromParsedData(receipt, parsed)
	}

	return s.createAndLink(ctx, receipt, true)
}

// MatchAndSuggest parses a receipt, finds the best transaction matches, and links the top one
func (s *receiptSvc) MatchAndSuggest(ctx context.Context, file io.Reader, filename string, provider domain.ReceiptProvider) (*domain.Receipt, error) {
	data, imageHash, err := s.readAndStoreImage(ctx, file, filename)
	if err != nil {
		return nil, err
	}

	receipt := &domain.Receipt{
		Provider:    provider,
		ImageSHA256: imageHash,
	}

	parsed, raw, parseErr := s.parser.Parse(ctx, bytes.NewReader(data), filename, provider)
	receipt.RawPayload = raw

	if parseErr != nil {
		s.log.Warn("parser failed on auto match", "err", parseErr)
		receipt.ParseStatus = domain.StatusFailed
		receipt.LinkStatus = domain.LinkStatusUnlinked
		return s.createAndLink(ctx, receipt, false)
	}

	s.log.Info("parser succeeded, finding candidates", "merchant", parsed.Merchant)
	receipt.ParseStatus = domain.StatusParsed
	s.populateReceiptFromParsedData(receipt, parsed)

	s.assignBestMatch(ctx, receipt, parsed)

	return s.createAndLink(ctx, receipt, false)
}

func (s *receiptSvc) readAndStoreImage(ctx context.Context, file io.Reader, filename string) ([]byte, []byte, error) {
	hasher := sha256.New()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, io.TeeReader(file, hasher)); err != nil {
		return nil, nil, fmt.Errorf("reading & hashing file: %w", err)
	}
	hash := hasher.Sum(nil)

	if err := s.preserveImage(ctx, buf.Bytes(), filename, hash); err != nil {
		s.log.Warn("could not preserve image", "err", err)
	}

	return buf.Bytes(), hash, nil
}

func (s *receiptSvc) preserveImage(ctx context.Context, data []byte, filename string, hash []byte) error {
	// TODO: write data to persistent storage and set receipt.ImageURL
	return nil
}

func (s *receiptSvc) createAndLink(ctx context.Context, receipt *domain.Receipt, manual bool) (*domain.Receipt, error) {
	created, err := s.store.CreateReceipt(ctx, receipt)
	if err != nil {
		return nil, fmt.Errorf("storing receipt: %w", err)
	}

	if receipt.TransactionID == nil {
		return s.store.GetReceipt(ctx, created.ID)
	}

	linkErr := s.store.SetTransactionReceipt(ctx, *receipt.TransactionID, created.ID)
	if linkErr == nil {
		return s.store.GetReceipt(ctx, created.ID)
	}

	if manual && errors.Is(linkErr, db.ErrConflict) {
		s.log.Warn("manual link conflict – cleaning up", "txID", *receipt.TransactionID, "receiptID", created.ID)
		if delErr := s.store.DeleteReceipt(context.Background(), created.ID); delErr != nil {
			s.log.Error("failed to delete orphaned receipt", "receiptID", created.ID, "err", delErr)
		}
		return nil, linkErr
	}

	if !errors.Is(linkErr, db.ErrConflict) {
		s.log.Error("failed to link receipt", "receiptID", created.ID, "err", linkErr)
	}

	return s.store.GetReceipt(ctx, created.ID)
}

func (s *receiptSvc) assignBestMatch(ctx context.Context, receipt *domain.Receipt, parsed *receiptparser.ParsedReceipt) {
	cands, err := s.store.FindCandidateTransactions(ctx, parsed.Merchant, *receipt.PurchaseDate, parsed.Total)
	if err != nil {
		s.log.Error("querying candidates failed", "err", err)
		receipt.LinkStatus = domain.LinkStatusUnlinked
		return
	}

	matches := s.scoreAndSelectBestMatch(cands, receipt)
	if len(matches) == 0 {
		receipt.LinkStatus = domain.LinkStatusUnlinked
		return
	}

	best := matches[0]
	receipt.TransactionID = &best.Tx.ID
	if best.FinalScore == 1.0 {
		receipt.LinkStatus = domain.LinkStatusMatched
	} else {
		receipt.LinkStatus = domain.LinkStatusNeedsVerification
	}

	if len(matches) > 1 {
		sugs := make([]int64, 0, len(matches)-1)
		for _, m := range matches[1:] {
			sugs = append(sugs, m.Tx.ID)
		}
		receipt.MatchSuggestions = sugs
	}
}

func (s *receiptSvc) scoreAndSelectBestMatch(candidates []*domain.TransactionWithScore, receipt *domain.Receipt) []MatchResult {
	if len(candidates) == 0 {
		return nil
	}

	const threshold = 0.7

	var results []MatchResult
	for _, c := range candidates {
		amountScore := scoreAmount(c.TxAmount, *receipt.TotalAmount)
		dateScore := dateScore(c.TxDate, *receipt.PurchaseDate)
		final := amountScore*0.45 + dateScore*0.35 + c.MerchantScore*0.2
		if final >= threshold {
			results = append(results, MatchResult{Tx: c, FinalScore: final})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].FinalScore > results[j].FinalScore
	})

	return results
}

// scoreAmount gives 1.0 for exact matches, allows up to 20% diff with linear decay
func scoreAmount(txAmount, receiptTotal float64) float64 {
	if math.Abs(txAmount-receiptTotal) < 0.01 {
		return 1.0
	}

	if txAmount < receiptTotal {
		return 0
	}

	maxDiff := receiptTotal * 0.20
	diff := math.Abs(txAmount - receiptTotal)
	if diff > maxDiff {
		return 0
	}

	return 0.9 * (1.0 - (diff / maxDiff))
}

// dateScore gives 1.0 for exact matches, allows up to 30 days diff with linear decay
func dateScore(txDate, receiptDate time.Time) float64 {
	d1 := time.Date(txDate.Year(), txDate.Month(), txDate.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(receiptDate.Year(), receiptDate.Month(), receiptDate.Day(), 0, 0, 0, 0, time.UTC)
	days := math.Abs(d1.Sub(d2).Hours() / 24)

	const maxDays = 30.0
	if days >= maxDays {
		return 0
	}

	return 1.0 - (days / maxDays)
}

// populateReceiptFromParsedData fills the receipt fields from the parser result
func (s *receiptSvc) populateReceiptFromParsedData(r *domain.Receipt, p *receiptparser.ParsedReceipt) {
	r.Merchant = &p.Merchant
	if p.Total > 0 {
		r.TotalAmount = &p.Total
	}

	if p.Date != "" {
		if t, err := time.Parse("2006-01-02", p.Date); err == nil {
			r.PurchaseDate = &t
		} else {
			s.log.Warn("could not parse date from receipt", "date", p.Date, "error", err)
		}
	} else {
		now := time.Now()
		r.PurchaseDate = &now
	}

	if len(p.Items) > 0 {
		r.Items = make([]domain.ReceiptItem, len(p.Items))
		for i, item := range p.Items {
			r.Items[i] = domain.ReceiptItem{
				Name:      item.Name,
				Qty:       &item.Qty,
				LineTotal: &item.Price,
			}
		}
	}

	if canon, err := json.Marshal(p); err == nil {
		r.CanonicalData = types.JSONText(canon)
	} else {
		s.log.Error("failed to marshal canonical data", "error", err)
	}
}
