package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"bytes"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxFileSize = 10 << 20 // 10 MB

func (s *Server) UploadReceipt(stream pb.ReceiptService_UploadReceiptServer) error {
	req, err := stream.Recv()
	if err != nil {
		return handleError(err)
	}

	info := req.GetInfo()
	if info == nil {
		return status.Error(codes.InvalidArgument, "first message must be of type Info")
	}
	if info.GetTransactionId() <= 0 {
		return status.Error(codes.InvalidArgument, "transaction_id is required")
	}

	var buf bytes.Buffer
	size := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return handleError(err)
		}

		chunk := req.GetChunk()
		size += len(chunk)
		if size > maxFileSize {
			return status.Error(codes.ResourceExhausted, "file size exceeds the 10MB limit")
		}
		if _, err := buf.Write(chunk); err != nil {
			return status.Error(codes.Internal, "failed to buffer file chunk")
		}
	}

	provider := fromProtoReceiptEngine(info.GetEngine())
	receipt, err := s.services.Receipts.LinkManual(stream.Context(), info.GetTransactionId(), &buf, info.GetFilename(), provider)
	if err != nil {
		return handleError(err)
	}

	if err := stream.SendAndClose(toProtoReceipt(receipt)); err != nil {
		return handleError(err)
	}
	return nil
}

func (s *Server) MatchReceipt(stream pb.ReceiptService_MatchReceiptServer) error {
	req, err := stream.Recv()
	if err != nil {
		return handleError(err)
	}

	info := req.GetInfo()
	if info == nil {
		return status.Error(codes.InvalidArgument, "first message must be of type Info")
	}

	var buf bytes.Buffer
	size := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return handleError(err)
		}

		chunk := req.GetChunk()
		size += len(chunk)
		if size > maxFileSize {
			return status.Error(codes.ResourceExhausted, "file size exceeds the 10MB limit")
		}
		if _, err := buf.Write(chunk); err != nil {
			return status.Error(codes.Internal, "failed to buffer file chunk")
		}
	}

	provider := fromProtoReceiptEngine(info.GetEngine())
	receipt, err := s.services.Receipts.MatchAndSuggest(stream.Context(), &buf, info.GetFilename(), provider)
	if err != nil {
		return handleError(err)
	}

	if err := stream.SendAndClose(toProtoReceipt(receipt)); err != nil {
		return handleError(err)
	}
	return nil
}
