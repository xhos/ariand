package db

import (
	"ariand/internal/domain"
	"context"
	"errors"
	"time"
)

// ListOpts specifies filtering and pagination options for listing transactions
type ListOpts struct {
	// --- cursor for pagination ---
	// based on the last transaction from the previous page to fetch the next one
	CursorID   *int64
	CursorDate *time.Time

	// --- filtering ---
	Start             *time.Time
	End               *time.Time
	Accounts          []string
	Categories        []string
	Direction         string  // "in" or "out"
	MerchantSearch    *string // case-insensitive search in merchant
	DescriptionSearch *string // case-insensitive search in description
	AmountMin         *float64
	AmountMax         *float64
	Currency          *string
	TimeOfDayStart    *string // "HH:MM:SS" format
	TimeOfDayEnd      *string // "HH:MM:SS" format

	// --- pagination Limit ---
	Limit int
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
	SetAccountAnchor(ctx context.Context, accountID int64, balance float64) error
	GetAccountBalance(ctx context.Context, accountID int64) (float64, error)
}
