package service

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
)

type DashboardService interface {
	Balance(ctx context.Context) (float64, error)
	Debt(ctx context.Context) (float64, error)
	Trends(ctx context.Context, opts db.ListOpts) ([]domain.TrendPoint, error)
}

type dashSvc struct{ store db.Store }

func newDashSvc(store db.Store) DashboardService { return &dashSvc{store} }

func (s *dashSvc) Balance(ctx context.Context) (float64, error) {
	return s.store.GetDashboardBalance(ctx)
}

func (s *dashSvc) Debt(ctx context.Context) (float64, error) {
	return s.store.GetDashboardDebt(ctx)
}

func (s *dashSvc) Trends(ctx context.Context, opts db.ListOpts) ([]domain.TrendPoint, error) {
	return s.store.GetDashboardTrends(ctx, opts)
}
