package db

import (
	"ariand/internal/domain"
	"context"
	"errors"
	"time"
)

type ListOpts struct {
	// cursor for pagination
	CursorID   *int64
	CursorDate *time.Time

	// filtering
	Start             *time.Time
	End               *time.Time
	AccountIDs        []int64
	Categories        []string
	Direction         string
	MerchantSearch    string
	DescriptionSearch string
	AmountMin         *float64
	AmountMax         *float64
	Currency          string
	TimeOfDayStart    *string
	TimeOfDayEnd      *string

	// pagination limit
	Limit int
}

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store interface {
	// transaction methods
	ListTransactions(ctx context.Context, opts ListOpts) ([]domain.Transaction, error)
	GetTransaction(ctx context.Context, id int64) (*domain.Transaction, error)
	CreateTransaction(ctx context.Context, t *domain.Transaction) (int64, error)
	UpdateTransaction(ctx context.Context, id int64, fields map[string]any) error
	DeleteTransaction(ctx context.Context, id int64) error

	// dashboard methods
	GetDashboardBalance(ctx context.Context) (float64, error)
	GetDashboardDebt(ctx context.Context) (float64, error)
	GetDashboardTrends(ctx context.Context, opts ListOpts) ([]domain.TrendPoint, error)

	// account methods
	ListAccounts(ctx context.Context) ([]domain.Account, error)
	GetAccount(ctx context.Context, id int64) (*domain.Account, error)
	CreateAccount(ctx context.Context, acc *domain.Account) (*domain.Account, error)
	UpdateAccount(ctx context.Context, id int64, fields map[string]any) error
	DeleteAccount(ctx context.Context, id int64) error
	SetAccountAnchor(ctx context.Context, accountID int64, balance float64) error
	GetAccountBalance(ctx context.Context, accountID int64) (float64, error)

	// category methods
	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategory(ctx context.Context, id int64) (*domain.Category, error)
	CreateCategory(ctx context.Context, slug, label, colour string) (int64, error)
	UpdateCategory(ctx context.Context, id int64, fields map[string]any) error
	DeleteCategory(ctx context.Context, id int64) error
	ListCategorySlugs(ctx context.Context) ([]string, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error)

	// receipt methods
	GetReceipt(ctx context.Context, id int64) (*domain.Receipt, error)
	CreateReceipt(ctx context.Context, rec *domain.Receipt) (*domain.Receipt, error)
	UpdateReceipt(ctx context.Context, id int64, fields map[string]any) error
	DeleteReceipt(ctx context.Context, id int64) error
	LinkReceiptToTransaction(ctx context.Context, receiptID, transactionID int64) error
}
