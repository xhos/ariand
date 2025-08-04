package service

import (
	sqlc "ariand/internal/db/sqlc"
	"ariand/internal/receiptparser"
	"bytes"
	"context"
	"database/sql"
	"errors"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/genproto/googleapis/type/money"
)

// protoMoneyToMoney converts proto money to google.type.Money
func protoMoneyToMoney(pm *money.Money) *money.Money {
	if pm == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: pm.CurrencyCode,
		Units:        pm.Units,
		Nanos:        pm.Nanos,
	}
}

type ReceiptService interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.Receipt, error)
	GetForUser(ctx context.Context, params sqlc.GetReceiptForUserParams) (*sqlc.Receipt, error)
	Create(ctx context.Context, params sqlc.CreateReceiptParams) (*sqlc.Receipt, error)
	Update(ctx context.Context, params sqlc.UpdateReceiptParams) error
	DeleteForUser(ctx context.Context, params sqlc.DeleteReceiptForUserParams) error

	ListItemsForReceipt(ctx context.Context, receiptID int64) ([]sqlc.ReceiptItem, error)
	GetItem(ctx context.Context, id int64) (*sqlc.ReceiptItem, error)
	CreateItem(ctx context.Context, params sqlc.CreateReceiptItemParams) (*sqlc.ReceiptItem, error)
	UpdateItem(ctx context.Context, params sqlc.UpdateReceiptItemParams) (*sqlc.ReceiptItem, error)
	DeleteItem(ctx context.Context, id int64) error
	BulkCreateItems(ctx context.Context, items []sqlc.BulkCreateReceiptItemsParams) error
	DeleteItemsByReceipt(ctx context.Context, receiptID int64) error

	GetUnlinked(ctx context.Context, limit *int32) ([]sqlc.GetUnlinkedReceiptsRow, error)
	GetMatchCandidates(ctx context.Context) ([]sqlc.GetReceiptMatchCandidatesRow, error)

	UploadReceipt(ctx context.Context, userID uuid.UUID, imageData []byte, provider string) (*sqlc.Receipt, error)
	ParseReceipt(ctx context.Context, receiptID int64, provider string) (*sqlc.Receipt, error)
	SearchReceipts(ctx context.Context, userID uuid.UUID, query string, limit *int32) ([]sqlc.Receipt, error)
	GetReceiptsByTransaction(ctx context.Context, transactionID int64) ([]sqlc.Receipt, error)
}

type receiptSvc struct {
	queries *sqlc.Queries
	parser  receiptparser.Client
	log     *log.Logger
}

func newReceiptSvc(queries *sqlc.Queries, parser receiptparser.Client, lg *log.Logger) ReceiptService {
	return &receiptSvc{queries: queries, parser: parser, log: lg}
}

func (s *receiptSvc) ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.Receipt, error) {
	receipts, err := s.queries.ListReceiptsForUser(ctx, userID)
	if err != nil {
		return nil, wrapErr("ReceiptService.ListForUser", err)
	}
	return receipts, nil
}

func (s *receiptSvc) GetForUser(ctx context.Context, params sqlc.GetReceiptForUserParams) (*sqlc.Receipt, error) {
	receipt, err := s.queries.GetReceiptForUser(ctx, params)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wrapErr("ReceiptService.GetForUser", ErrNotFound)
	}
	if err != nil {
		return nil, wrapErr("ReceiptService.GetForUser", err)
	}
	return &receipt, nil
}

func (s *receiptSvc) Create(ctx context.Context, params sqlc.CreateReceiptParams) (*sqlc.Receipt, error) {
	receipt, err := s.queries.CreateReceipt(ctx, params)
	if err != nil {
		return nil, wrapErr("ReceiptService.Create", err)
	}
	return &receipt, nil
}

func (s *receiptSvc) Update(ctx context.Context, params sqlc.UpdateReceiptParams) error {
	_, err := s.queries.UpdateReceipt(ctx, params)
	if err != nil {
		return wrapErr("ReceiptService.Update", err)
	}
	return nil
}

func (s *receiptSvc) DeleteForUser(ctx context.Context, params sqlc.DeleteReceiptForUserParams) error {
	_, err := s.queries.DeleteReceiptForUser(ctx, params)
	if err != nil {
		return wrapErr("ReceiptService.DeleteForUser", err)
	}
	return nil
}

func (s *receiptSvc) ListItemsForReceipt(ctx context.Context, receiptID int64) ([]sqlc.ReceiptItem, error) {
	items, err := s.queries.ListReceiptItemsForReceipt(ctx, receiptID)
	if err != nil {
		return nil, wrapErr("ReceiptService.ListItemsForReceipt", err)
	}
	return items, nil
}

func (s *receiptSvc) GetItem(ctx context.Context, id int64) (*sqlc.ReceiptItem, error) {
	item, err := s.queries.GetReceiptItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wrapErr("ReceiptService.GetItem", ErrNotFound)
	}
	if err != nil {
		return nil, wrapErr("ReceiptService.GetItem", err)
	}
	return &item, nil
}

func (s *receiptSvc) CreateItem(ctx context.Context, params sqlc.CreateReceiptItemParams) (*sqlc.ReceiptItem, error) {
	item, err := s.queries.CreateReceiptItem(ctx, params)
	if err != nil {
		return nil, wrapErr("ReceiptService.CreateItem", err)
	}
	return &item, nil
}

func (s *receiptSvc) UpdateItem(ctx context.Context, params sqlc.UpdateReceiptItemParams) (*sqlc.ReceiptItem, error) {
	item, err := s.queries.UpdateReceiptItem(ctx, params)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wrapErr("ReceiptService.UpdateItem", ErrNotFound)
	}
	if err != nil {
		return nil, wrapErr("ReceiptService.UpdateItem", err)
	}
	return &item, nil
}

func (s *receiptSvc) DeleteItem(ctx context.Context, id int64) error {
	_, err := s.queries.DeleteReceiptItem(ctx, id)
	if err != nil {
		return wrapErr("ReceiptService.DeleteItem", err)
	}
	return nil
}

func (s *receiptSvc) BulkCreateItems(ctx context.Context, items []sqlc.BulkCreateReceiptItemsParams) error {
	_, err := s.queries.BulkCreateReceiptItems(ctx, items)
	if err != nil {
		return wrapErr("ReceiptService.BulkCreateItems", err)
	}
	return nil
}

func (s *receiptSvc) DeleteItemsByReceipt(ctx context.Context, receiptID int64) error {
	_, err := s.queries.DeleteReceiptItemsByReceipt(ctx, receiptID)
	if err != nil {
		return wrapErr("ReceiptService.DeleteItemsByReceipt", err)
	}
	return nil
}

func (s *receiptSvc) GetUnlinked(ctx context.Context, limit *int32) ([]sqlc.GetUnlinkedReceiptsRow, error) {
	receipts, err := s.queries.GetUnlinkedReceipts(ctx, limit)
	if err != nil {
		return nil, wrapErr("ReceiptService.GetUnlinked", err)
	}
	return receipts, nil
}

func (s *receiptSvc) GetMatchCandidates(ctx context.Context) ([]sqlc.GetReceiptMatchCandidatesRow, error) {
	candidates, err := s.queries.GetReceiptMatchCandidates(ctx)
	if err != nil {
		return nil, wrapErr("ReceiptService.GetMatchCandidates", err)
	}
	return candidates, nil
}

func (s *receiptSvc) UploadReceipt(ctx context.Context, userID uuid.UUID, imageData []byte, provider string) (*sqlc.Receipt, error) {
	// Parse the receipt using the gRPC service
	parsedReceipt, err := s.parser.Parse(ctx,
		bytes.NewReader(imageData),
		"receipt.jpg", // filename - could be passed as parameter
		"image/jpeg",  // content type - could be detected or passed as parameter
		nil,           // engine - could be passed as parameter
	)
	if err != nil {
		s.log.Warn("failed to parse receipt", "error", err)
		// Continue with creating the receipt even if parsing fails
	}

	// Create receipt record
	parseStatus := int16(1) // Pending
	linkStatus := int16(1)  // Unlinked

	params := sqlc.CreateReceiptParams{
		Engine:      1, // Default engine
		ParseStatus: &parseStatus,
		LinkStatus:  &linkStatus,
	}

	if parsedReceipt != nil {
		successStatus := int16(2) // Success
		params.ParseStatus = &successStatus

		if parsedReceipt.Merchant != nil && *parsedReceipt.Merchant != "" {
			params.Merchant = parsedReceipt.Merchant
		}

		if parsedReceipt.TotalAmount != nil {
			// Convert money to decimal for the parameter (SQLC param types aren't mapped)
			totalAmountDecimal := decimal.NewFromFloat(float64(parsedReceipt.TotalAmount.Units) + float64(parsedReceipt.TotalAmount.Nanos)/1e9)
			params.TotalAmount = &totalAmountDecimal
			params.Currency = &parsedReceipt.TotalAmount.CurrencyCode
		}
		// Parse date if provided
		// TODO: Parse purchase date from receipt
	}

	receipt, err := s.queries.CreateReceipt(ctx, params)
	if err != nil {
		return nil, wrapErr("ReceiptService.UploadReceipt", err)
	}

	// Create receipt items if parsed successfully
	if parsedReceipt != nil && len(parsedReceipt.Items) > 0 {
		var itemParams []sqlc.BulkCreateReceiptItemsParams
		for i, item := range parsedReceipt.Items {
			qtyDecimal := decimal.NewFromFloat(item.Quantity)
			lineNo := int32(i + 1)

			itemParams = append(itemParams, sqlc.BulkCreateReceiptItemsParams{
				ReceiptID: receipt.ID,
				LineNo:    &lineNo,
				Name:      item.Name,
				Qty:       &qtyDecimal,
				UnitPrice: item.UnitPrice,
				LineTotal: item.LineTotal,
			})
		}

		if err := s.BulkCreateItems(ctx, itemParams); err != nil {
			s.log.Warn("failed to create receipt items", "error", err)
		}
	}

	return &receipt, nil
}

func (s *receiptSvc) ParseReceipt(ctx context.Context, receiptID int64, provider string) (*sqlc.Receipt, error) {
	return nil, wrapErr("ReceiptService.ParseReceipt", ErrUnimplemented)
}

func (s *receiptSvc) SearchReceipts(ctx context.Context, userID uuid.UUID, query string, limit *int32) ([]sqlc.Receipt, error) {
	return nil, wrapErr("ReceiptService.SearchReceipts", ErrUnimplemented)
}

func (s *receiptSvc) GetReceiptsByTransaction(ctx context.Context, transactionID int64) ([]sqlc.Receipt, error) {
	return nil, wrapErr("ReceiptService.GetReceiptsByTransaction", ErrUnimplemented)
}

// type receiptSvc struct {
// 	store  db.Store
// 	parser receiptparser.Client
// 	log    *log.Logger
// }

// func newReceiptSvc(store db.Store, parser receiptparser.Client, lg *log.Logger) ReceiptService {
// 	return &receiptSvc{store: store, parser: parser, log: lg}
// }

// temporary conversion functions - to be removed when services are fully migrated
// func domainToSqlcReceipt(dr *domain.Receipt) *sqlc.Receipt {
// 	return &sqlc.Receipt{
// 		ID:          dr.ID,
// 		Engine:      ariand.ReceiptEngine(ariand.ReceiptEngine_value[string(dr.Engine)]),
// 		ParseStatus: ariand.ReceiptParseStatus(ariand.ReceiptParseStatus_value[string(dr.ParseStatus)]),
// 		LinkStatus:  ariand.ReceiptLinkStatus(ariand.ReceiptLinkStatus_value[string(dr.LinkStatus)]),
// 		MatchIds:    []int64(dr.MatchIDs),
// 		Merchant:    dr.Merchant,
// 		// Add other fields as needed
// 	}
// }

// func sqlcToDomainReceipt(sr *sqlc.Receipt) *domain.Receipt {
// 	return &domain.Receipt{
// 		ID:          sr.ID,
// 		Engine:      domain.ReceiptEngine(sr.Engine.String()),
// 		ParseStatus: domain.ReceiptParseStatus(sr.ParseStatus.String()),
// 		LinkStatus:  domain.ReceiptLinkStatus(sr.LinkStatus.String()),
// 		MatchIDs:    sr.MatchIds,
// 		Merchant:    sr.Merchant,
// 		// Add other fields as needed
// 	}
// }

// // LinkManual manually links a receipt to a specific transaction
// func (s *receiptSvc) LinkManual(ctx context.Context, transactionID int64, file io.Reader, filename string, provider domain.ReceiptEngine) (*domain.Receipt, error) {
// 	data, imageHash, err := s.readAndStoreImage(ctx, file, filename)
// 	if err != nil {
// 		return nil, err
// 	}

// 	receipt := &domain.Receipt{
// 		Engine:      provider,
// 		ImageSHA256: imageHash,
// 		LinkStatus:  domain.LinkMatched,
// 	}

// 	parsed, raw, parseErr := s.parser.Parse(ctx, bytes.NewReader(data), filename, provider)
// 	receipt.RawPayload = raw

// 	if parseErr != nil {
// 		s.log.Warn("parser failed on manual link", "err", parseErr, "txID", transactionID)
// 		receipt.ParseStatus = domain.ParseFailed
// 	} else {
// 		s.log.Info("parser succeeded on manual link", "merchant", parsed.Merchant, "txID", transactionID)
// 		receipt.ParseStatus = domain.ParseSuccess
// 		s.populateReceiptFromParsedData(receipt, parsed)
// 	}

// 	return s.createAndLink(ctx, receipt, &transactionID, true)
// }

// // MatchAndSuggest parses a receipt, finds the best transaction matches, and links the top one
// func (s *receiptSvc) MatchAndSuggest(ctx context.Context, file io.Reader, filename string, provider domain.ReceiptEngine) (*domain.Receipt, error) {
// 	data, imageHash, err := s.readAndStoreImage(ctx, file, filename)
// 	if err != nil {
// 		return nil, err
// 	}

// 	receipt := &domain.Receipt{
// 		Engine:      provider,
// 		ImageSHA256: imageHash,
// 	}

// 	parsed, raw, parseErr := s.parser.Parse(ctx, bytes.NewReader(data), filename, provider)
// 	receipt.RawPayload = raw

// 	if parseErr != nil {
// 		s.log.Warn("parser failed on auto match", "err", parseErr)
// 		receipt.ParseStatus = domain.ParseFailed
// 		receipt.LinkStatus = domain.LinkUnlinked
// 		return s.createAndLink(ctx, receipt, nil, false)
// 	}

// 	s.log.Info("parser succeeded, finding candidates", "merchant", parsed.Merchant)
// 	receipt.ParseStatus = domain.ParseSuccess
// 	s.populateReceiptFromParsedData(receipt, parsed)

// 	bestTransactionID := s.assignBestMatch(ctx, receipt, parsed)

// 	return s.createAndLink(ctx, receipt, bestTransactionID, false)
// }

// func (s *receiptSvc) readAndStoreImage(ctx context.Context, file io.Reader, filename string) ([]byte, []byte, error) {
// 	hasher := sha256.New()
// 	buf := new(bytes.Buffer)
// 	if _, err := io.Copy(buf, io.TeeReader(file, hasher)); err != nil {
// 		return nil, nil, fmt.Errorf("reading & hashing file: %w", err)
// 	}
// 	hash := hasher.Sum(nil)

// 	if err := s.preserveImage(ctx, buf.Bytes(), filename, hash); err != nil {
// 		s.log.Warn("could not preserve image", "err", err)
// 	}

// 	return buf.Bytes(), hash, nil
// }

// func (s *receiptSvc) preserveImage(_ context.Context, _ []byte, _ string, _ []byte) error {
// 	// TODO: write data to persistent storage and set receipt.ImageURL
// 	return nil
// }

// func (s *receiptSvc) createAndLink(ctx context.Context, receipt *domain.Receipt, transactionID *int64, manual bool) (*domain.Receipt, error) {
// 	created, err := s.store.CreateReceipt(ctx, domainToSqlcReceipt(receipt))
// 	if err != nil {
// 		return nil, fmt.Errorf("storing receipt: %w", err)
// 	}

// 	if transactionID == nil {
// 		receipt, err := s.store.GetReceipt(ctx, created.Id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return sqlcToDomainReceipt(receipt), nil
// 	}

// 	linkErr := s.store.SetTransactionReceipt(ctx, *transactionID, created.Id)
// 	if linkErr == nil {
// 		receipt, err := s.store.GetReceipt(ctx, created.Id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return sqlcToDomainReceipt(receipt), nil
// 	}

// 	if manual && errors.Is(linkErr, db.ErrConflict) {
// 		s.log.Warn("manual link conflict – cleaning up", "txID", *transactionID, "receiptID", created.Id)
// 		if delErr := s.store.DeleteReceipt(context.Background(), created.Id); delErr != nil {
// 			s.log.Error("failed to delete orphaned receipt", "receiptID", created.Id, "err", delErr)
// 		}
// 		return nil, linkErr
// 	}

// 	if !errors.Is(linkErr, db.ErrConflict) {
// 		s.log.Error("failed to link receipt", "receiptID", created.ID, "err", linkErr)
// 	}

// 	receipt, err := s.store.GetReceipt(ctx, created.Id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return sqlcToDomainReceipt(receipt), nil
// }

// func (s *receiptSvc) assignBestMatch(ctx context.Context, receipt *domain.Receipt, parsed *receiptparser.ParsedReceipt) *int64 {
// 	cands, err := s.store.FindCandidateTransactions(ctx, parsed.Merchant, *receipt.PurchaseDate, parsed.Total)
// 	if err != nil {
// 		s.log.Error("querying candidates failed", "err", err)
// 		receipt.LinkStatus = domain.LinkUnlinked
// 		return nil
// 	}

// 	matches := s.scoreAndSelectBestMatch(cands, receipt)
// 	if len(matches) == 0 {
// 		receipt.LinkStatus = domain.LinkUnlinked
// 		return nil
// 	}

// 	best := matches[0]
// 	transactionID := &best.Tx.ID
// 	if best.FinalScore == 1.0 {
// 		receipt.LinkStatus = domain.LinkMatched
// 	} else {
// 		receipt.LinkStatus = domain.LinkNeedsVerification
// 	}

// 	if len(matches) > 1 {
// 		sugs := make([]int64, 0, len(matches)-1)
// 		for _, m := range matches[1:] {
// 			sugs = append(sugs, m.Tx.ID)
// 		}
// 		receipt.MatchIDs = sugs
// 	}

// 	return transactionID
// }

// func (s *receiptSvc) scoreAndSelectBestMatch(candidates []*domain.TransactionWithScore, receipt *domain.Receipt) []MatchResult {
// 	if len(candidates) == 0 {
// 		return nil
// 	}

// 	const threshold = 0.7

// 	var results []MatchResult
// 	for _, c := range candidates {
// 		amountScore := scoreAmount(c.TxAmount, deref(receipt.TotalAmount))
// 		dateScore := dateScore(c.TxDate, *receipt.PurchaseDate)
// 		final := amountScore*0.45 + dateScore*0.35 + c.MerchantScore*0.2
// 		if final >= threshold {
// 			results = append(results, MatchResult{Tx: c, FinalScore: final})
// 		}
// 	}

// 	sort.Slice(results, func(i, j int) bool {
// 		return results[i].FinalScore > results[j].FinalScore
// 	})

// 	return results
// }

// // scoreAmount gives 1.0 for exact matches, allows up to 20% diff with linear decay
// func scoreAmount(txAmount, receiptTotal float64) float64 {
// 	if math.Abs(txAmount-receiptTotal) < 0.01 {
// 		return 1.0
// 	}

// 	if txAmount < receiptTotal {
// 		return 0
// 	}

// 	maxDiff := receiptTotal * 0.20
// 	diff := math.Abs(txAmount - receiptTotal)
// 	if diff > maxDiff {
// 		return 0
// 	}

// 	return 0.9 * (1.0 - (diff / maxDiff))
// }

// // dateScore gives 1.0 for exact matches, allows up to 30 days diff with linear decay
// func dateScore(txDate, receiptDate time.Time) float64 {
// 	d1 := time.Date(txDate.Year(), txDate.Month(), txDate.Day(), 0, 0, 0, 0, time.UTC)
// 	d2 := time.Date(receiptDate.Year(), receiptDate.Month(), receiptDate.Day(), 0, 0, 0, 0, time.UTC)
// 	days := math.Abs(d1.Sub(d2).Hours() / 24)

// 	const maxDays = 30.0
// 	if days >= maxDays {
// 		return 0
// 	}

// 	return 1.0 - (days / maxDays)
// }

// // populateReceiptFromParsedData fills the receipt fields from the parser result
// func (s *receiptSvc) populateReceiptFromParsedData(r *domain.Receipt, p *receiptparser.ParsedReceipt) {
// 	r.Merchant = &p.Merchant
// 	if p.Total > 0 {
// 		r.TotalAmount = &p.Total
// 	}

// 	if p.Date != "" {
// 		if t, err := time.Parse("2006-01-02", p.Date); err == nil {
// 			r.PurchaseDate = &t
// 		} else {
// 			s.log.Warn("could not parse date from receipt", "date", p.Date, "error", err)
// 		}
// 	} else {
// 		now := time.Now()
// 		r.PurchaseDate = &now
// 	}

// 	if len(p.Items) > 0 {
// 		r.Items = make([]domain.ReceiptItem, len(p.Items))
// 		for i, item := range p.Items {
// 			r.Items[i] = domain.ReceiptItem{
// 				Name:      item.Name,
// 				Qty:       &item.Qty,
// 				LineTotal: &item.Price,
// 			}
// 		}
// 	}

// 	if canon, err := json.Marshal(p); err == nil {
// 		r.CanonicalData = types.JSONText(canon)
// 	} else {
// 		s.log.Error("failed to marshal canonical data", "error", err)
// 	}
// }

// // deref is a helper to safely dereference a pointer to any type.
// func deref[T any](p *T) T {
// 	if p == nil {
// 		var zero T
// 		return zero
// 	}
// 	return *p
// }
