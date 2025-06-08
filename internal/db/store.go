package db

import (
	"ariand/internal/domain"
	"context"
	"errors"
	"time"
)

// ListOpts specifies filtering and pagination options for listing transactions
type ListOpts struct {
	Start     *time.Time
	End       *time.Time
	Accounts  []string
	Direction string
	Limit     int
	Offset    int
}

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// Store defines the interface for all database operations
type Store interface {
	// Transaction methods
	ListTransactions(ctx context.Context, opts ListOpts) ([]domain.Transaction, error)
	GetTransaction(ctx context.Context, id int64) (*domain.Transaction, error)
	CreateTransaction(ctx context.Context, t *domain.Transaction) (int64, error)
	UpdateTransaction(ctx context.Context, id int64, fields map[string]any) error
	DeleteTransaction(ctx context.Context, id int64) error

	// Dashboard methods
	GetDashboardBalance(ctx context.Context) (float64, error)
	GetDashboardDebt(ctx context.Context) (float64, error)
	GetDashboardTrends(ctx context.Context, opts ListOpts) ([]domain.TrendPoint, error)

	// Account methods
	ListAccounts(ctx context.Context) ([]domain.Account, error)
	GetAccount(ctx context.Context, id int64) (*domain.Account, error)
	SetAccountAnchor(ctx context.Context, accountID int64, date time.Time, balance float64) error
	GetAccountBalance(ctx context.Context, accountID int64) (float64, error)
}
