package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryInterceptor is a function type for unary RPC interceptors
type UnaryInterceptor func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)

// StreamInterceptor is a function type for streaming RPC interceptors
type StreamInterceptor func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error

// ChainUnaryInterceptors creates a single interceptor from multiple unary interceptors
// The first interceptor is the outermost (executed first), and the last is the innermost (executed last)
func ChainUnaryInterceptors(interceptors ...UnaryInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return chainUnary(interceptors, 0, ctx, req, info, handler)
	}
}

// ChainStreamInterceptors creates a single interceptor from multiple stream interceptors
func ChainStreamInterceptors(interceptors ...StreamInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return chainStream(interceptors, 0, srv, ss, info, handler)
	}
}

func chainUnary(interceptors []UnaryInterceptor, current int, ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if current == len(interceptors) {
		return handler(ctx, req)
	}

	return interceptors[current](ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return chainUnary(interceptors, current+1, ctx, req, info, handler)
	})
}

func chainStream(interceptors []StreamInterceptor, current int, srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if current == len(interceptors) {
		return handler(srv, ss)
	}

	return interceptors[current](srv, ss, info, func(srv interface{}, ss grpc.ServerStream) error {
		return chainStream(interceptors, current+1, srv, ss, info, handler)
	})
}
