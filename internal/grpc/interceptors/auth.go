package interceptors

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authCtxKey struct{}

var Authenticated = authCtxKey{}

var (
	ErrMissingMetadata = errors.New("missing metadata")
	ErrInvalidToken    = errors.New("invalid token")
	ErrUnauthorized    = errors.New("unauthorized")
)

// AuthConfig holds configuration for the auth interceptor
type AuthConfig struct {
	Logger      *log.Logger
	ExpectedKey string
	// Methods that don't require authentication
	SkipMethods []string
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(config AuthConfig) UnaryInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if this method should skip authentication
		for _, method := range config.SkipMethods {
			if method == info.FullMethod {
				return handler(ctx, req)
			}
		}

		if config.ExpectedKey == "" {
			config.Logger.Error("API_KEY not set on server")
			return nil, status.Error(codes.Internal, "server configuration error")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			config.Logger.Error("missing metadata")
			return nil, status.Error(codes.Unauthenticated, ErrMissingMetadata.Error())
		}

		// Get authorization header
		auth := md.Get("authorization")
		if len(auth) == 0 {
			config.Logger.Error("missing authorization header")
			return nil, status.Error(codes.Unauthenticated, ErrUnauthorized.Error())
		}

		token := auth[0]
		const prefix = "Bearer "
		if !strings.HasPrefix(token, prefix) {
			config.Logger.Error("invalid token format")
			return nil, status.Error(codes.Unauthenticated, ErrInvalidToken.Error())
		}

		apiKey := strings.TrimPrefix(token, prefix)
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(config.ExpectedKey)) != 1 {
			config.Logger.Error("invalid api key")
			return nil, status.Error(codes.Unauthenticated, ErrUnauthorized.Error())
		}

		// Add authentication info to context
		authCtx := context.WithValue(ctx, Authenticated, true)
		return handler(authCtx, req)
	}
}

// NewAuthStreamInterceptor creates a new auth interceptor for streaming RPCs
func NewAuthStreamInterceptor(config AuthConfig) StreamInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Check if this method should skip authentication
		for _, method := range config.SkipMethods {
			if method == info.FullMethod {
				return handler(srv, ss)
			}
		}

		if config.ExpectedKey == "" {
			config.Logger.Error("API_KEY not set on server")
			return status.Error(codes.Internal, "server configuration error")
		}

		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			config.Logger.Error("missing metadata")
			return status.Error(codes.Unauthenticated, ErrMissingMetadata.Error())
		}

		// Get authorization header
		auth := md.Get("authorization")
		if len(auth) == 0 {
			config.Logger.Error("missing authorization header")
			return status.Error(codes.Unauthenticated, ErrUnauthorized.Error())
		}

		token := auth[0]
		const prefix = "Bearer "
		if !strings.HasPrefix(token, prefix) {
			config.Logger.Error("invalid token format")
			return status.Error(codes.Unauthenticated, ErrInvalidToken.Error())
		}

		apiKey := strings.TrimPrefix(token, prefix)
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(config.ExpectedKey)) != 1 {
			config.Logger.Error("invalid api key")
			return status.Error(codes.Unauthenticated, ErrUnauthorized.Error())
		}

		// Create a new context with authentication info
		authCtx := context.WithValue(ctx, Authenticated, true)
		wrapped := &wrappedServerStream{ServerStream: ss, ctx: authCtx}
		return handler(srv, wrapped)
	}
}

// wrappedServerStream wraps a grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
