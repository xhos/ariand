package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"ariand/internal/db"
	"context"
)

func (s *Server) GetBalance(ctx context.Context, _ *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	v, err := s.services.Dashboard.Balance(ctx)
	if err != nil {
		return nil, handleError(err)
	}
	// Assuming dashboard currency is CAD. A more robust solution might involve a user profile setting.
	return &pb.GetBalanceResponse{Balance: float64ToMoney(v, "CAD")}, nil
}

func (s *Server) GetDebt(ctx context.Context, _ *pb.GetDebtRequest) (*pb.GetDebtResponse, error) {
	v, err := s.services.Dashboard.Debt(ctx)
	if err != nil {
		return nil, handleError(err)
	}
	// Assuming dashboard currency is CAD.
	return &pb.GetDebtResponse{Debt: float64ToMoney(v, "CAD")}, nil
}

func (s *Server) GetTrends(ctx context.Context, req *pb.GetTrendsRequest) (*pb.GetTrendsResponse, error) {
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
		d, err := stringToDate(tp.Date)
		if err != nil {
			return nil, handleError(err) // Or log and continue
		}
		out[i] = &pb.TrendPoint{
			// Assuming dashboard currency is CAD.
			Date:     d,
			Income:   float64ToMoney(tp.Income, "CAD"),
			Expenses: float64ToMoney(tp.Expenses, "CAD"),
		}
	}

	return &pb.GetTrendsResponse{Trends: out}, nil
}
