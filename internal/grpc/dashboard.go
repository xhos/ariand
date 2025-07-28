package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"ariand/internal/db"
	"context"
)

func (s *Server) GetBalance(ctx context.Context, _ *pb.DashboardBalanceRequest) (*pb.DashboardBalanceResponse, error) {
	v, err := s.services.Dashboard.Balance(ctx)
	if err != nil {
		return nil, handleError(err)
	}
	return &pb.DashboardBalanceResponse{Balance: v}, nil
}

func (s *Server) GetDebt(ctx context.Context, _ *pb.DashboardDebtRequest) (*pb.DashboardDebtResponse, error) {
	v, err := s.services.Dashboard.Debt(ctx)
	if err != nil {
		return nil, handleError(err)
	}
	return &pb.DashboardDebtResponse{Debt: v}, nil
}

func (s *Server) GetTrends(ctx context.Context, req *pb.DashboardTrendsRequest) (*pb.DashboardTrendsResponse, error) {
	opts := db.ListOpts{}
	if ts := req.GetStart(); ts != nil && ts.IsValid() {
		t := ts.AsTime()
		opts.Start = &t
	}
	if ts := req.GetEnd(); ts != nil && ts.IsValid() {
		t := ts.AsTime()
		opts.End = &t
	}

	trends, err := s.services.Dashboard.Trends(ctx, opts)
	if err != nil {
		return nil, handleError(err)
	}

	out := make([]*pb.TrendPoint, len(trends))
	for i, tp := range trends {
		out[i] = &pb.TrendPoint{
			Date:     tp.Date,
			Income:   tp.Income,
			Expenses: tp.Expenses,
		}
	}

	return &pb.DashboardTrendsResponse{Trends: out}, nil
}
