package interceptors

import (
	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
)

// InterceptorConfig holds configuration for all interceptors
type InterceptorConfig struct {
	Logger            *log.Logger
	APIKey            string
	EnableAuth        bool
	EnableRateLimit   bool
	RateLimitRPS      float64 // requests per second for rate limiting
	RateLimitCapacity int
	// Methods that should skip authentication (e.g., health checks, public endpoints)
	PublicMethods []string
}

// SetupInterceptors creates and chains all interceptors based on configuration
func SetupInterceptors(config InterceptorConfig) (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	var unaryInterceptors []UnaryInterceptor
	var streamInterceptors []StreamInterceptor

	// Recovery should be first (outermost) to catch all panics
	unaryInterceptors = append(unaryInterceptors, NewRecoveryInterceptor(RecoveryConfig{
		Logger: config.Logger,
	}))
	streamInterceptors = append(streamInterceptors, NewRecoveryStreamInterceptor(RecoveryConfig{
		Logger: config.Logger,
	}))

	// Logging should be early to log all requests
	unaryInterceptors = append(unaryInterceptors, NewLoggingInterceptor(LoggingConfig{
		Logger: config.Logger,
	}))
	streamInterceptors = append(streamInterceptors, NewLoggingStreamInterceptor(LoggingConfig{
		Logger: config.Logger,
	}))

	// Authentication comes next
	if config.EnableAuth {
		unaryInterceptors = append(unaryInterceptors, NewAuthInterceptor(AuthConfig{
			Logger:      config.Logger,
			ExpectedKey: config.APIKey,
			SkipMethods: config.PublicMethods,
		}))
		streamInterceptors = append(streamInterceptors, NewAuthStreamInterceptor(AuthConfig{
			Logger:      config.Logger,
			ExpectedKey: config.APIKey,
			SkipMethods: config.PublicMethods,
		}))
	}

	// Rate limiting comes after auth (so authenticated requests can potentially skip it)
	if config.EnableRateLimit {
		limiter := NewTokenBucketLimiter(config.RateLimitRPS, config.RateLimitCapacity)
		unaryInterceptors = append(unaryInterceptors, NewRateLimitInterceptor(RateLimitConfig{
			Logger:            config.Logger,
			Limiter:           limiter,
			SkipAuthenticated: true, // Skip rate limiting for authenticated requests
		}))
		streamInterceptors = append(streamInterceptors, NewRateLimitStreamInterceptor(RateLimitConfig{
			Logger:            config.Logger,
			Limiter:           limiter,
			SkipAuthenticated: true,
		}))
	}

	return ChainUnaryInterceptors(unaryInterceptors...),
		ChainStreamInterceptors(streamInterceptors...)
}
