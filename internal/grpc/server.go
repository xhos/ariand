package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"ariand/internal/db"
	"ariand/internal/service"
	"errors"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements all gRPC services in one type.
// (No duplicate method names after proto rename.)
type Server struct {
	pb.UnimplementedTransactionServiceServer
	pb.UnimplementedAccountServiceServer
	pb.UnimplementedCategoryServiceServer
	pb.UnimplementedDashboardServiceServer
	pb.UnimplementedReceiptServiceServer

	services *service.Services
	log      *log.Logger
}

func NewServer(services *service.Services, logger *log.Logger) *Server {
	return &Server{
		services: services,
		log:      logger,
	}
}

func (s *Server) RegisterServices(grpcServer *grpc.Server) {
	pb.RegisterTransactionServiceServer(grpcServer, s)
	pb.RegisterAccountServiceServer(grpcServer, s)
	pb.RegisterCategoryServiceServer(grpcServer, s)
	pb.RegisterDashboardServiceServer(grpcServer, s)
	pb.RegisterReceiptServiceServer(grpcServer, s)
}

// handleError maps domain/store errors to gRPC codes.
func handleError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, db.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, db.ErrConflict) {
		return status.Error(codes.AlreadyExists, err.Error())
	}
	return status.Errorf(codes.Internal, "internal error: %v", err)
}
