package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"ariand/internal/db"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.Transaction, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "transaction id is required")
	}
	txn, err := s.services.Transactions.Get(ctx, req.GetId())
	if err != nil {
		return nil, handleError(err)
	}
	return toProtoTransaction(txn), nil
}

func (s *Server) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	// Mirror REST defaults and bounds.
	queryLimit := int(req.GetLimit())
	if queryLimit <= 0 {
		queryLimit = 25
	} else if queryLimit > 100 {
		queryLimit = 100
	}

	opts := db.ListOpts{
		Limit:             queryLimit + 1, // over-fetch for pagination
		AccountIDs:        req.GetAccountIds(),
		Categories:        req.GetCategories(),
		Direction:         req.GetDirection(),
		MerchantSearch:    req.GetMerchantSearch(),
		DescriptionSearch: req.GetDescriptionSearch(),
		Currency:          req.GetCurrency(),
	}

	// Optional numeric filters
	if req.AmountMin != nil {
		opts.AmountMin = req.AmountMin
	}
	if req.AmountMax != nil {
		opts.AmountMax = req.AmountMax
	}

	// Optional date/cursor filters
	if ts := req.GetStartDate(); ts != nil && ts.IsValid() {
		t := ts.AsTime()
		opts.Start = &t
	}
	if ts := req.GetEndDate(); ts != nil && ts.IsValid() {
		t := ts.AsTime()
		opts.End = &t
	}
	if cur := req.GetCursor(); cur != nil {
		if ts := cur.GetDate(); ts != nil && ts.IsValid() {
			t := ts.AsTime()
			opts.CursorDate = &t
		}
		if cur.GetId() > 0 {
			id := cur.GetId()
			opts.CursorID = &id
		}
	}

	txns, err := s.services.Transactions.List(ctx, opts)
	if err != nil {
		return nil, handleError(err)
	}

	// Build next cursor and trim to requested page size.
	var nextCursor *pb.Cursor
	if len(txns) > queryLimit {
		last := txns[queryLimit] // the over-fetched item
		id := last.ID
		nextCursor = &pb.Cursor{
			Date: toProtoTimestamp(last.TxDate),
			Id:   &id,
		}
		txns = txns[:queryLimit]
	}

	out := make([]*pb.Transaction, len(txns))
	for i := range txns {
		out[i] = toProtoTransaction(&txns[i])
	}

	return &pb.ListTransactionsResponse{
		Transactions: out,
		NextCursor:   nextCursor,
	}, nil
}

func (s *Server) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	domainTxn := fromProtoCreateTransactionRequest(req)
	id, err := s.services.Transactions.Create(ctx, domainTxn)
	if err != nil {
		return nil, handleError(err)
	}
	return &pb.CreateTransactionResponse{Id: id}, nil
}

func (s *Server) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.Transaction, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "transaction id is required")
	}
	if req.GetUpdateMask() == nil {
		return nil, status.Error(codes.InvalidArgument, "update_mask is required")
	}

	fields := fieldsFromUpdateMask(req.GetUpdateMask(), req.GetUpdates())
	if len(fields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "update_mask contains no valid fields")
	}

	if err := s.services.Transactions.Update(ctx, req.GetId(), fields); err != nil {
		return nil, handleError(err)
	}

	updatedTxn, err := s.services.Transactions.Get(ctx, req.GetId())
	if err != nil {
		return nil, handleError(err)
	}
	return toProtoTransaction(updatedTxn), nil
}

func (s *Server) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "transaction id is required")
	}
	if err := s.services.Transactions.Delete(ctx, req.GetId()); err != nil {
		return nil, handleError(err)
	}
	return &pb.DeleteTransactionResponse{}, nil
}

func (s *Server) CategorizeTransaction(ctx context.Context, req *pb.CategorizeTransactionRequest) (*pb.CategorizeTransactionResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "transaction id is required")
	}
	if err := s.services.Transactions.CategorizeTransaction(ctx, req.GetId()); err != nil {
		return nil, handleError(err)
	}
	return &pb.CategorizeTransactionResponse{}, nil
}

func (s *Server) IdentifyMerchant(ctx context.Context, req *pb.IdentifyMerchantRequest) (*pb.IdentifyMerchantResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "transaction id is required")
	}
	if err := s.services.Transactions.IdentifyMerchantForTransaction(ctx, req.GetId()); err != nil {
		return nil, handleError(err)
	}
	return &pb.IdentifyMerchantResponse{}, nil
}
