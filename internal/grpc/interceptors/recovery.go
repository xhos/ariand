package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryConfig holds configuration for the recovery interceptor
type RecoveryConfig struct {
	Logger *log.Logger
}

// NewRecoveryInterceptor creates a new recovery interceptor that catches panics
func NewRecoveryInterceptor(config RecoveryConfig) UnaryInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				config.Logger.Error("panic recovered",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()),
				)

				// Return internal error to client
				err = status.Error(codes.Internal, "internal server error")
				resp = nil
			}
		}()

		return handler(ctx, req)
	}
}

// NewRecoveryStreamInterceptor creates a new recovery interceptor for streaming RPCs
func NewRecoveryStreamInterceptor(config RecoveryConfig) StreamInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				config.Logger.Error("panic recovered in stream",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()),
				)

				// Return internal error to client
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}
