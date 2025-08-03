package service

import (
	sqlc "ariand/internal/db/sqlc"
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DashboardService interface {
	BalanceForUser(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	DebtForUser(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	TrendsForUser(ctx context.Context, params sqlc.GetDashboardTrendsForUserParams) ([]sqlc.GetDashboardTrendsForUserRow, error)
	SummaryForUser(ctx context.Context, params sqlc.GetDashboardSummaryForUserParams) (*sqlc.GetDashboardSummaryForUserRow, error)
	MonthlyComparisonForUser(ctx context.Context, params sqlc.GetMonthlyComparisonForUserParams) ([]sqlc.GetMonthlyComparisonForUserRow, error)
	TopCategoriesForUser(ctx context.Context, params sqlc.GetTopCategoriesForUserParams) ([]sqlc.GetTopCategoriesForUserRow, error)
}

type dashSvc struct{ queries *sqlc.Queries }

func newDashSvc(queries *sqlc.Queries) DashboardService { return &dashSvc{queries} }

func (s *dashSvc) BalanceForUser(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	balance, err := s.queries.GetDashboardBalanceForUser(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(balance), nil
}

func (s *dashSvc) DebtForUser(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	debt, err := s.queries.GetDashboardDebtForUser(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(debt), nil
}

func (s *dashSvc) TrendsForUser(ctx context.Context, params sqlc.GetDashboardTrendsForUserParams) ([]sqlc.GetDashboardTrendsForUserRow, error) {
	return s.queries.GetDashboardTrendsForUser(ctx, params)
}

func (s *dashSvc) SummaryForUser(ctx context.Context, params sqlc.GetDashboardSummaryForUserParams) (*sqlc.GetDashboardSummaryForUserRow, error) {
	summary, err := s.queries.GetDashboardSummaryForUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (s *dashSvc) MonthlyComparisonForUser(ctx context.Context, params sqlc.GetMonthlyComparisonForUserParams) ([]sqlc.GetMonthlyComparisonForUserRow, error) {
	return s.queries.GetMonthlyComparisonForUser(ctx, params)
}

func (s *dashSvc) TopCategoriesForUser(ctx context.Context, params sqlc.GetTopCategoriesForUserParams) ([]sqlc.GetTopCategoriesForUserRow, error) {
	return s.queries.GetTopCategoriesForUser(ctx, params)
}
