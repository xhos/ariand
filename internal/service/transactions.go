package service

import (
	"ariand/internal/ai"
	sqlc "ariand/internal/db/sqlc"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionService interface {
	// User-scoped operations
	ListForUser(ctx context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.ListTransactionsForUserRow, error)
	GetForUser(ctx context.Context, params sqlc.GetTransactionForUserParams) (*sqlc.GetTransactionForUserRow, error)
	CreateForUser(ctx context.Context, params sqlc.CreateTransactionForUserParams) (int64, error)
	Update(ctx context.Context, params sqlc.UpdateTransactionParams) error
	DeleteForUser(ctx context.Context, params sqlc.DeleteTransactionForUserParams) (int64, error)
	BulkDeleteForUser(ctx context.Context, params sqlc.BulkDeleteTransactionsForUserParams) error
	BulkCategorizeForUser(ctx context.Context, params sqlc.BulkCategorizeTransactionsForUserParams) error

	// Analytics and search
	GetTransactionCountByAccountForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.GetTransactionCountByAccountForUserRow, error)
	FindCandidateTransactionsForUser(ctx context.Context, params sqlc.FindCandidateTransactionsForUserParams) ([]sqlc.FindCandidateTransactionsForUserRow, error)

	// AI-powered operations
	CategorizeTransaction(ctx context.Context, userID uuid.UUID, txID int64) error
	IdentifyMerchantForTransaction(ctx context.Context, userID uuid.UUID, txID int64) error
}

type txnSvc struct {
	queries *sqlc.Queries
	log     *log.Logger
	catSvc  CategoryService
	aiMgr   *ai.Manager
}

func newTxnSvc(queries *sqlc.Queries, lg *log.Logger, catSvc CategoryService, aiMgr *ai.Manager) TransactionService {
	return &txnSvc{queries: queries, log: lg, catSvc: catSvc, aiMgr: aiMgr}
}

type categorizationResult struct {
	CategorySlug string
	Status       string
	Suggestions  []string
}

func (s *txnSvc) ListForUser(ctx context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.ListTransactionsForUserRow, error) {
	return s.queries.ListTransactionsForUser(ctx, params)
}

func (s *txnSvc) GetForUser(ctx context.Context, params sqlc.GetTransactionForUserParams) (*sqlc.GetTransactionForUserRow, error) {
	row, err := s.queries.GetTransactionForUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *txnSvc) CreateForUser(ctx context.Context, params sqlc.CreateTransactionForUserParams) (int64, error) {
	if err := s.validateCreateParams(params); err != nil {
		return 0, err
	}
	return s.queries.CreateTransactionForUser(ctx, params)
}

func (s *txnSvc) Update(ctx context.Context, params sqlc.UpdateTransactionParams) error {
	_, err := s.queries.UpdateTransaction(ctx, params)
	return err
}

func (s *txnSvc) DeleteForUser(ctx context.Context, params sqlc.DeleteTransactionForUserParams) (int64, error) {
	return s.queries.DeleteTransactionForUser(ctx, params)
}

func (s *txnSvc) BulkDeleteForUser(ctx context.Context, params sqlc.BulkDeleteTransactionsForUserParams) error {
	_, err := s.queries.BulkDeleteTransactionsForUser(ctx, params)
	return err
}

func (s *txnSvc) BulkCategorizeForUser(ctx context.Context, params sqlc.BulkCategorizeTransactionsForUserParams) error {
	_, err := s.queries.BulkCategorizeTransactionsForUser(ctx, params)
	return err
}

func (s *txnSvc) GetTransactionCountByAccountForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.GetTransactionCountByAccountForUserRow, error) {
	return s.queries.GetTransactionCountByAccountForUser(ctx, userID)
}

func (s *txnSvc) FindCandidateTransactionsForUser(ctx context.Context, params sqlc.FindCandidateTransactionsForUserParams) ([]sqlc.FindCandidateTransactionsForUserRow, error) {
	return s.queries.FindCandidateTransactionsForUser(ctx, params)
}

func (s *txnSvc) CategorizeTransaction(ctx context.Context, userID uuid.UUID, txID int64) error {
	tx, err := s.queries.GetTransactionForUser(ctx, sqlc.GetTransactionForUserParams{
		UserID: userID,
		ID:     txID,
	})
	if err != nil {
		return fmt.Errorf("getting transaction for categorization: %w", err)
	}

	// Convert GetTransactionForUserRow to Transaction for determineCategory
	txForCategory := &sqlc.Transaction{
		ID:           tx.ID,
		AccountID:    tx.AccountID,
		EmailID:      tx.EmailID,
		TxDate:       tx.TxDate,
		TxAmount:     tx.TxAmount,
		TxCurrency:   tx.TxCurrency,
		TxDirection:  tx.TxDirection,
		TxDesc:       tx.TxDesc,
		BalanceAfter: tx.BalanceAfter,
		Merchant:     tx.Merchant,
		CategoryID:   tx.CategoryID,
		CatStatus:    tx.CatStatus,
		Suggestions:  tx.Suggestions,
		UserNotes:    tx.UserNotes,
		CreatedAt:    tx.CreatedAt,
		UpdatedAt:    tx.UpdatedAt,
	}

	result, err := s.determineCategory(ctx, userID, txForCategory)
	if err != nil {
		return fmt.Errorf("determining category: %w", err)
	}

	var categoryID *int64
	if result.CategorySlug != "" {
		category, err := s.catSvc.BySlug(ctx, result.CategorySlug)
		if err != nil {
			return fmt.Errorf("finding category by slug %s: %w", result.CategorySlug, err)
		}
		categoryID = &category.ID
	}

	params := sqlc.UpdateTransactionParams{
		ID:          txID,
		UserID:      userID,
		CategoryID:  categoryID,
		CatStatus:   int16Ptr(2), // AI categorization status
		Suggestions: result.Suggestions,
	}

	_, err = s.queries.UpdateTransaction(ctx, params)
	return err
}

func (s *txnSvc) IdentifyMerchantForTransaction(ctx context.Context, userID uuid.UUID, txID int64) error {
	tx, err := s.queries.GetTransactionForUser(ctx, sqlc.GetTransactionForUserParams{
		UserID: userID,
		ID:     txID,
	})
	if err != nil {
		return fmt.Errorf("getting transaction for merchant identification: %w", err)
	}

	if tx.TxDesc == nil || *tx.TxDesc == "" {
		return errors.New("transaction has no description to analyze")
	}

	if s.aiMgr == nil {
		return errors.New("AI manager not available")
	}

	// Get a provider from the manager (we could make this configurable)
	provider, err := s.aiMgr.GetProvider("openai", "gpt-4o-mini") // or whatever default you want
	if err != nil {
		return fmt.Errorf("getting AI provider: %w", err)
	}

	merchant, err := provider.ExtractMerchant(ctx, *tx.TxDesc)
	if err != nil {
		return fmt.Errorf("extracting merchant: %w", err)
	}

	if merchant == "" {
		return nil // No merchant identified
	}

	params := sqlc.UpdateTransactionParams{
		ID:       txID,
		UserID:   userID,
		Merchant: &merchant,
	}

	_, err = s.queries.UpdateTransaction(ctx, params)
	return err
}

// determineCategory analyzes a transaction to suggest a category
func (s *txnSvc) determineCategory(ctx context.Context, userID uuid.UUID, tx *sqlc.Transaction) (*categorizationResult, error) {
	// 1. Try similarity search (rule-based)
	if tx.TxDesc != nil {
		params := sqlc.ListTransactionsForUserParams{
			UserID: userID,
			DescQ:  tx.TxDesc,
			Limit:  int32Ptr(10),
		}
		rows, err := s.queries.ListTransactionsForUser(ctx, params)
		if err == nil {
			for _, potentialMatch := range rows {
				if potentialMatch.ID == tx.ID || potentialMatch.CategoryID == nil || potentialMatch.TxDesc == nil {
					continue
				}

				descSim := similarity(strings.ToLower(*tx.TxDesc), strings.ToLower(*potentialMatch.TxDesc))
				if descSim >= 0.7 && amountClose(tx.TxAmount, potentialMatch.TxAmount, 0.2) {
					s.log.Info("Found similar transaction for auto-categorization", "txID", tx.ID, "similarTxID", potentialMatch.ID)
					return &categorizationResult{
						CategorySlug: *potentialMatch.CategorySlug,
						Status:       "auto",
					}, nil
				}
			}
		}
	}

	// 2. Fallback to AI if available
	if s.aiMgr != nil {
		provider, err := s.aiMgr.GetProvider("openai", "gpt-4o-mini")
		if err == nil {
			s.log.Info("Falling back to AI for categorization", "txID", tx.ID)
			slugs, err := s.catSvc.ListSlugs(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list category slugs for AI: %w", err)
			}

			categorySlug, _, suggestions, err := provider.CategorizeTransaction(ctx, *tx, slugs)
			if err != nil {
				return nil, fmt.Errorf("AI categorization failed: %w", err)
			}

			return &categorizationResult{
				CategorySlug: categorySlug,
				Suggestions:  suggestions,
				Status:       "ai",
			}, nil
		}
	}

	// 3. Return empty result if no AI available
	return &categorizationResult{
		CategorySlug: "",
		Status:       "failed",
		Suggestions:  []string{},
	}, nil
}

// validateCreateParams validates transaction creation parameters
func (s *txnSvc) validateCreateParams(params sqlc.CreateTransactionForUserParams) error {
	if params.TxAmount.IsZero() {
		return errors.New("tx_amount cannot be zero")
	}
	switch params.TxDirection {
	case 1, 2: // DIRECTION_INCOMING, DIRECTION_OUTGOING
		// valid
	default:
		return errors.New("tx_direction must be 1 (DIRECTION_INCOMING) or 2 (DIRECTION_OUTGOING)")
	}
	return nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int16Ptr(i int16) *int16 {
	return &i
}

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

func amountClose(a, b decimal.Decimal, pct float64) bool {
	if a.Equal(b) {
		return true
	}

	diff := a.Sub(b).Abs()
	max := decimal.Max(a.Abs(), b.Abs())
	threshold := max.Mul(decimal.NewFromFloat(pct))

	return diff.LessThanOrEqual(threshold)
}
