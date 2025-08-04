package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggingConfig holds configuration for the logging interceptor
type LoggingConfig struct {
	Logger *log.Logger
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(config LoggingConfig) UnaryInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		logFields := []interface{}{
			"method", info.FullMethod,
			"duration", duration,
		}

		// Add peer information if available
		if p, ok := peer.FromContext(ctx); ok {
			logFields = append(logFields, "remote", p.Addr.String())
		}

		// Add request ID if available
		if reqID := getRequestID(ctx); reqID != "" {
			logFields = append(logFields, "req_id", reqID)
		}

		// Add error information if any
		if err != nil {
			st := status.Convert(err)
			logFields = append(logFields,
				"error", true,
				"code", st.Code().String(),
				"message", st.Message(),
			)
			config.Logger.Error("grpc.request", logFields...)
		} else {
			logFields = append(logFields, "error", false)
			config.Logger.Info("grpc.request", logFields...)
		}

		return resp, err
	}
}

// NewLoggingStreamInterceptor creates a new logging interceptor for streaming RPCs
func NewLoggingStreamInterceptor(config LoggingConfig) StreamInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)

		logFields := []interface{}{
			"method", info.FullMethod,
			"duration", duration,
			"stream", true,
		}

		// Add peer information if available
		if p, ok := peer.FromContext(ss.Context()); ok {
			logFields = append(logFields, "remote", p.Addr.String())
		}

		// Add request ID if available
		if reqID := getRequestID(ss.Context()); reqID != "" {
			logFields = append(logFields, "req_id", reqID)
		}

		// Add error information if any
		if err != nil {
			st := status.Convert(err)
			logFields = append(logFields,
				"error", true,
				"code", st.Code().String(),
				"message", st.Message(),
			)
			config.Logger.Error("grpc.stream", logFields...)
		} else {
			logFields = append(logFields, "error", false)
			config.Logger.Info("grpc.stream", logFields...)
		}

		return err
	}
}

// getRequestID extracts request ID from gRPC metadata
func getRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Check for X-Request-Id header
	if reqIDs := md.Get("x-request-id"); len(reqIDs) > 0 {
		return reqIDs[0]
	}

	// Check for Request-Id header (alternative)
	if reqIDs := md.Get("request-id"); len(reqIDs) > 0 {
		return reqIDs[0]
	}

	return ""
}

// isSensitiveField checks if a field contains sensitive information
func isSensitiveField(key string) bool {
	sensitive := []string{"password", "token", "api_key", "secret", "auth", "credential"}
	lowerKey := strings.ToLower(key)
	for _, s := range sensitive {
		if strings.Contains(lowerKey, s) {
			return true
		}
	}
	return false
}
