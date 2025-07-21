package service

import (
	"ariand/internal/ai"
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/log"
)

type TransactionService interface {
	List(ctx context.Context, opts db.ListOpts) ([]domain.Transaction, error)
	Get(ctx context.Context, id int64) (*domain.Transaction, error)
	Create(ctx context.Context, tx *domain.Transaction) (int64, error)
	Update(ctx context.Context, id int64, fields map[string]any) error
	Delete(ctx context.Context, id int64) error
	CategorizeTransaction(ctx context.Context, txID int64) error
	IdentifyMerchantForTransaction(ctx context.Context, txID int64) error
}

type txnSvc struct {
	store  db.Store
	log    *log.Logger
	catSvc CategoryService
}

func newTxnSvc(store db.Store, lg *log.Logger, catSvc CategoryService) TransactionService {
	return &txnSvc{store: store, log: lg, catSvc: catSvc}
}

type categorizationResult struct {
	CategorySlug string
	Status       string
	Suggestions  []string
}

func (s *txnSvc) List(ctx context.Context, opts db.ListOpts) ([]domain.Transaction, error) {
	return s.store.ListTransactions(ctx, opts)
}

func (s *txnSvc) Get(ctx context.Context, id int64) (*domain.Transaction, error) {
	return s.store.GetTransaction(ctx, id)
}

func (s *txnSvc) Create(ctx context.Context, tx *domain.Transaction) (int64, error) {
	if err := validateTx(tx); err != nil {
		return 0, err
	}
	return s.store.CreateTransaction(ctx, tx)
}

func (s *txnSvc) Update(ctx context.Context, id int64, fields map[string]any) error {
	// cheap guard to prevent accidentally nulling monetary data
	if _, ok := fields["tx_amount"]; ok {
		if v, ok := fields["tx_amount"].(float64); !ok || v == 0 {
			return errors.New("tx_amount must be non-zero")
		}
	}
	return s.store.UpdateTransaction(ctx, id, fields)
}

func (s *txnSvc) Delete(ctx context.Context, id int64) error {
	return s.store.DeleteTransaction(ctx, id)
}

func (s *txnSvc) CategorizeTransaction(ctx context.Context, txID int64) error {
	tx, err := s.store.GetTransaction(ctx, txID)
	if err != nil {
		return fmt.Errorf("getting transaction for categorization: %w", err)
	}

	res, err := s.determineCategory(ctx, tx)
	if err != nil {
		return fmt.Errorf("determining category: %w", err)
	}

	var catID *int64
	if res.CategorySlug != "" && res.CategorySlug != "other" {
		category, err := s.catSvc.BySlug(ctx, res.CategorySlug)
		if err != nil {
			s.log.Warn("determined category slug not found in db, defaulting to null", "slug", res.CategorySlug, "txID", txID, "err", err)
		} else {
			catID = &category.ID
		}
	}

	fields := map[string]any{
		"category_id": catID,
		"cat_status":  res.Status,
		"suggestions": res.Suggestions,
	}
	return s.store.UpdateTransaction(ctx, txID, fields)
}

func (s *txnSvc) IdentifyMerchantForTransaction(ctx context.Context, txID int64) error {
	merchant, err := s.determineMerchant(ctx, txID)
	if err != nil {
		return fmt.Errorf("determining merchant: %w", err)
	}
	if merchant == "" {
		return nil // Nothing to update
	}
	return s.store.UpdateTransaction(ctx, txID, map[string]any{"merchant": merchant})
}

func (s *txnSvc) determineCategory(ctx context.Context, tx *domain.Transaction) (*categorizationResult, error) {
	// 1. Try similarity search (rule-based)
	if tx.TxDesc != nil {
		sim, err := s.store.ListTransactions(ctx, db.ListOpts{
			DescriptionSearch: *tx.TxDesc,
			Limit:             10,
		})
		if err == nil {
			for _, cand := range sim {
				if cand.ID == tx.ID || cand.CategorySlug == nil || cand.TxDesc == nil {
					continue
				}
				descSim := similarity(strings.ToLower(*tx.TxDesc), strings.ToLower(*cand.TxDesc))
				if descSim >= 0.7 && amountClose(tx.TxAmount, cand.TxAmount, 0.2) {
					s.log.Info("Found similar transaction for auto-categorization", "txID", tx.ID, "similarTxID", cand.ID)
					return &categorizationResult{
						CategorySlug: *cand.CategorySlug,
						Status:       "auto",
					}, nil
				}
			}
		}
	}

	// 2. Fallback to AI
	s.log.Info("Falling back to AI for categorization", "txID", tx.ID)
	slugs, err := s.catSvc.ListSlugs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list category slugs for AI: %w", err)
	}

	prov, err := ai.GetManager().GetProvider("openai", "gpt-4o")
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	categorySlug, _, suggestions, err := prov.CategorizeTransaction(ctx, *tx, slugs)
	if err != nil {
		return nil, fmt.Errorf("AI categorization failed: %w", err)
	}

	return &categorizationResult{
		CategorySlug: categorySlug,
		Suggestions:  suggestions,
		Status:       "ai",
	}, nil
}

func (s *txnSvc) determineMerchant(ctx context.Context, txID int64) (string, error) {
	tx, err := s.store.GetTransaction(ctx, txID)
	if err != nil {
		return "", err
	}
	if tx.TxDesc == nil || *tx.TxDesc == "" {
		return "", errors.New("transaction has no description")
	}

	// 1. Pre-AI Check: Look for similar, already-identified transactions
	similarTxs, err := s.store.ListTransactions(ctx, db.ListOpts{
		DescriptionSearch: *tx.TxDesc,
		Limit:             10,
	})
	if err == nil {
		for _, similarTx := range similarTxs {
			if similarTx.ID != tx.ID && similarTx.Merchant != nil && *similarTx.Merchant != "" {
				if similarity(strings.ToLower(*tx.TxDesc), strings.ToLower(*similarTx.TxDesc)) > 0.8 {
					s.log.Info("Found similar transaction for merchant extraction", "txID", txID, "similarTxID", similarTx.ID)
					return *similarTx.Merchant, nil
				}
			}
		}
	}

	// 2. AI Fallback
	s.log.Info("Falling back to AI for merchant extraction", "txID", txID)
	prov, err := ai.GetManager().GetProvider("openai", "gpt-4o")
	if err != nil {
		return "", err
	}

	return prov.ExtractMerchant(ctx, *tx.TxDesc)
}

func validateTx(t *domain.Transaction) error {
	if t.TxAmount == 0 {
		return errors.New("tx_amount cannot be zero")
	}
	if dir := strings.ToLower(t.TxDirection); dir != "in" && dir != "out" {
		return errors.New("tx_direction must be 'in' or 'out'")
	}
	return nil
}

// similarity calculates a Jaccard index on word sets.
func similarity(a, b string) float64 {
	aa := strings.Fields(a)
	bb := strings.Fields(b)
	set := map[string]bool{}
	for _, w := range aa {
		set[w] = true
	}
	inter := 0
	for _, w := range bb {
		if set[w] {
			inter++
		}
	}
	union := len(aa) + len(bb) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

// amountClose checks if two float64 values are within a certain percentage.
func amountClose(a, b, pct float64) bool {
	if a == b {
		return true
	}
	diff := math.Abs(a - b)
	return diff <= pct*math.Max(a, b)
}
