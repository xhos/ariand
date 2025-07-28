package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListAccounts(ctx context.Context, _ *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	accounts, err := s.services.Accounts.List(ctx)
	if err != nil {
		return nil, handleError(err)
	}

	out := make([]*pb.Account, len(accounts))
	for i := range accounts {
		out[i] = toProtoAccount(&accounts[i])
	}
	return &pb.ListAccountsResponse{Accounts: out}, nil
}

func (s *Server) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.Account, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "account id is required")
	}
	account, err := s.services.Accounts.Get(ctx, req.GetId())
	if err != nil {
		return nil, handleError(err)
	}
	return toProtoAccount(account), nil
}

func (s *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.Account, error) {
	if req.GetName() == "" || req.GetBank() == "" {
		return nil, status.Error(codes.InvalidArgument, "name and bank are required")
	}
	domainAcc := fromProtoCreateAccountRequest(req)
	createdAcc, err := s.services.Accounts.Create(ctx, domainAcc)
	if err != nil {
		return nil, handleError(err)
	}
	return toProtoAccount(createdAcc), nil
}

func (s *Server) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "account id is required")
	}
	if err := s.services.Accounts.Delete(ctx, req.GetId()); err != nil {
		return nil, handleError(err)
	}
	return &pb.DeleteAccountResponse{}, nil
}

func (s *Server) SetAnchor(ctx context.Context, req *pb.SetAnchorRequest) (*pb.SetAnchorResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "account id is required")
	}
	if err := s.services.Accounts.SetAnchor(ctx, req.GetId(), req.GetBalance()); err != nil {
		return nil, handleError(err)
	}
	return &pb.SetAnchorResponse{}, nil
}

// NOTE: matches the renamed RPC in the proto
func (s *Server) GetAccountBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "account id is required")
	}
	balance, err := s.services.Accounts.Balance(ctx, req.GetId())
	if err != nil {
		return nil, handleError(err)
	}
	return &pb.GetBalanceResponse{Balance: balance}, nil
}
