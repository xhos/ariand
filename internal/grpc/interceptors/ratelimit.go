package interceptors

import (
	"context"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RateLimiter interface for pluggable rate limiting strategies
type RateLimiter interface {
	Allow(key string) bool
}

// TokenBucketLimiter implements a simple token bucket rate limiter
type TokenBucketLimiter struct {
	buckets         map[string]*bucket
	mu              sync.RWMutex
	rate            time.Duration // time between tokens
	capacity        int           // max tokens
	cleanupInterval time.Duration // cleanup interval for old buckets
	lastCleanup     time.Time
}

type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(requestsPerSecond float64, capacity int) *TokenBucketLimiter {
	rate := time.Duration(float64(time.Second) / requestsPerSecond)
	return &TokenBucketLimiter{
		buckets:         make(map[string]*bucket),
		rate:            rate,
		capacity:        capacity,
		cleanupInterval: time.Minute * 5, // cleanup old buckets every 5 minutes
	}
}

func (tbl *TokenBucketLimiter) Allow(key string) bool {
	tbl.mu.RLock()
	b, exists := tbl.buckets[key]
	tbl.mu.RUnlock()

	if !exists {
		tbl.mu.Lock()
		// Double-check after acquiring write lock
		if b, exists = tbl.buckets[key]; !exists {
			b = &bucket{
				tokens:     tbl.capacity,
				lastRefill: time.Now(),
			}
			tbl.buckets[key] = b
		}
		tbl.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	// Refill tokens based on time elapsed
	tokensToAdd := int(now.Sub(b.lastRefill) / tbl.rate)
	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > tbl.capacity {
			b.tokens = tbl.capacity
		}
		b.lastRefill = now
	}

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes old unused buckets
func (tbl *TokenBucketLimiter) cleanup() {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	if time.Since(tbl.lastCleanup) < tbl.cleanupInterval {
		return
	}

	cutoff := time.Now().Add(-time.Hour) // Remove buckets unused for 1 hour
	for key, b := range tbl.buckets {
		b.mu.Lock()
		if b.lastRefill.Before(cutoff) {
			delete(tbl.buckets, key)
		}
		b.mu.Unlock()
	}

	tbl.lastCleanup = time.Now()
}

// RateLimitConfig holds configuration for the rate limiting interceptor
type RateLimitConfig struct {
	Logger  *log.Logger
	Limiter RateLimiter
	// If true, skip rate limiting for authenticated requests
	SkipAuthenticated bool
}

// NewRateLimitInterceptor creates a new rate limiting interceptor
func NewRateLimitInterceptor(config RateLimitConfig) UnaryInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip rate limiting for authenticated requests if configured
		if config.SkipAuthenticated {
			if auth := ctx.Value(Authenticated); auth != nil && auth.(bool) {
				return handler(ctx, req)
			}
		}

		// Get client IP for rate limiting key
		key := getClientIP(ctx)
		if key == "" {
			key = "unknown"
		}

		// Check rate limit
		if !config.Limiter.Allow(key) {
			config.Logger.Warn("rate limit exceeded", "client", key, "method", info.FullMethod)
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// NewRateLimitStreamInterceptor creates a new rate limiting interceptor for streaming RPCs
func NewRateLimitStreamInterceptor(config RateLimitConfig) StreamInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Skip rate limiting for authenticated requests if configured
		if config.SkipAuthenticated {
			if auth := ctx.Value(Authenticated); auth != nil && auth.(bool) {
				return handler(srv, ss)
			}
		}

		// Get client IP for rate limiting key
		key := getClientIP(ctx)
		if key == "" {
			key = "unknown"
		}

		// Check rate limit
		if !config.Limiter.Allow(key) {
			config.Logger.Warn("rate limit exceeded", "client", key, "method", info.FullMethod)
			return status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(srv, ss)
	}
}

// getClientIP extracts the client IP address from the context
func getClientIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}
