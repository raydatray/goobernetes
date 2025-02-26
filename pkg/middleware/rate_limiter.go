package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrInvalidRateLimit    = errors.New("invalid rate limit: value must be positive")
	ErrInvalidWindowSize   = errors.New("invalid window size: duration must be positive")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrRateLimiterNotFound = errors.New("rate limiter not found")
)

type RateLimitResponse struct {
	Error     string `json:"error,omitempty"`
	Limit     int64  `json:"limit"`
	Remaining int64  `json:"remaining"`
	Reset     int64  `json:"reset"` // Unix timestamp
}

type RateLimiter struct {
	requests   atomic.Int64
	limit      int64
	windowSize time.Duration
	mu         sync.RWMutex
	lastReset  time.Time
}

func NewRateLimiter(requestLimit int64, windowSize time.Duration) (*RateLimiter, error) {
	if requestLimit <= 0 {
		return nil, ErrInvalidRateLimit
	}

	if windowSize <= 0 {
		return nil, ErrInvalidWindowSize
	}

	rateLimiter := &RateLimiter{
		limit:      requestLimit,
		windowSize: windowSize,
		lastReset:  time.Now(),
	}
	return rateLimiter, nil
}

func (rateLimiter *RateLimiter) TryAcquire() bool {
	rateLimiter.mu.Lock()
	defer rateLimiter.mu.Unlock()

	now := time.Now()
	if now.Sub(rateLimiter.lastReset) >= rateLimiter.windowSize {
		rateLimiter.requests.Store(0)
		rateLimiter.lastReset = now
	}

	currentRequests := rateLimiter.requests.Add(1)
	if currentRequests > rateLimiter.limit {
		rateLimiter.requests.Add(-1)
		return false
	}

	return true
}

func (rateLimiter *RateLimiter) Reset() {
	rateLimiter.mu.Lock()
	defer rateLimiter.mu.Unlock()

	rateLimiter.requests.Store(0)
	rateLimiter.lastReset = time.Now()
}

func (rateLimiter *RateLimiter) GetCurrentLimit() int64 {
	rateLimiter.mu.RLock()
	defer rateLimiter.mu.RUnlock()
	return rateLimiter.limit
}

func (rateLimiter *RateLimiter) GetRemainingRequests() int64 {
	rateLimiter.mu.RLock()
	defer rateLimiter.mu.RUnlock()
	return rateLimiter.limit - rateLimiter.requests.Load()
}

func (rateLimiter *RateLimiter) GetWindowSize() time.Duration {
	rateLimiter.mu.RLock()
	defer rateLimiter.mu.RUnlock()
	return rateLimiter.windowSize
}

func (rateLimiter *RateLimiter) GetResetTime() int64 {
	return time.Now().Add(rateLimiter.GetWindowSize()).Unix()
}

func writeRateLimitResponse(responseWriter http.ResponseWriter, statusCode int, response RateLimitResponse) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	json.NewEncoder(responseWriter).Encode(response)
}

// Common function to create a middleware with error handling
func createRateLimitMiddleware(createLimiterFunc func() (interface{}, error), requestsPerSecond int64,
	handleRequestFunc func(interface{}, http.ResponseWriter, *http.Request, http.HandlerFunc)) Middleware {

	limiter, err := createLimiterFunc()
	if err != nil {
		return func(next http.HandlerFunc) http.HandlerFunc {
			return func(responseWriter http.ResponseWriter, request *http.Request) {
				response := RateLimitResponse{
					Error: "Rate limiter misconfigured",
					Limit: requestsPerSecond,
				}
				writeRateLimitResponse(responseWriter, http.StatusServiceUnavailable, response)
			}
		}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(responseWriter http.ResponseWriter, request *http.Request) {
			handleRequestFunc(limiter, responseWriter, request, next)
		}
	}
}

// Handles rate limiting for a single RateLimiter
func handleBasicRateLimiting(limiter *RateLimiter, responseWriter http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
	if !limiter.TryAcquire() {
		response := RateLimitResponse{
			Error:     ErrRateLimitExceeded.Error(),
			Limit:     limiter.GetCurrentLimit(),
			Remaining: 0,
			Reset:     limiter.GetResetTime(),
		}
		writeRateLimitResponse(responseWriter, http.StatusTooManyRequests, response)
		return
	}

	next(responseWriter, request)
}

func RateLimiterMiddleware(requestsPerSecond int64) Middleware {
	return createRateLimitMiddleware(
		func() (interface{}, error) {
			return NewRateLimiter(requestsPerSecond, time.Second)
		},
		requestsPerSecond,
		func(limiter interface{}, responseWriter http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
			handleBasicRateLimiting(limiter.(*RateLimiter), responseWriter, request, next)
		},
	)
}

type IPRateLimiter struct {
	limiters sync.Map
	limit    int64
	window   time.Duration
}

func NewIPRateLimiter(requestLimit int64, windowSize time.Duration) (*IPRateLimiter, error) {
	if requestLimit <= 0 {
		return nil, ErrInvalidRateLimit
	}

	if windowSize <= 0 {
		return nil, ErrInvalidWindowSize
	}

	return &IPRateLimiter{
		limit:  requestLimit,
		window: windowSize,
	}, nil
}

func (ipLimiter *IPRateLimiter) GetLimiter(ipAddress string) (*RateLimiter, error) {
	limiterInterface, exists := ipLimiter.limiters.Load(ipAddress)
	if exists {
		limiter, ok := limiterInterface.(*RateLimiter)
		if !ok {
			return nil, fmt.Errorf("invalid limiter type for IP %s", ipAddress)
		}
		return limiter, nil
	}

	newLimiter, err := NewRateLimiter(ipLimiter.limit, ipLimiter.window)
	if err != nil {
		return nil, fmt.Errorf("failed to create limiter for IP %s: %w", ipAddress, err)
	}

	actualLimiter, loaded := ipLimiter.limiters.LoadOrStore(ipAddress, newLimiter)
	if loaded {
		limiter, ok := actualLimiter.(*RateLimiter)
		if !ok {
			return nil, fmt.Errorf("invalid stored limiter type for IP %s", ipAddress)
		}
		return limiter, nil
	}
	return newLimiter, nil
}

// Handles rate limiting based on IP address
func handleIPRateLimiting(ipLimiter *IPRateLimiter, responseWriter http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
	ipAddress := request.RemoteAddr
	if forwardedFor := request.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ipAddress = strings.Split(forwardedFor, ",")[0]
	}

	limiter, err := ipLimiter.GetLimiter(ipAddress)
	if err != nil {
		response := RateLimitResponse{
			Error: "Rate limiter error",
			Limit: ipLimiter.limit,
		}
		writeRateLimitResponse(responseWriter, http.StatusServiceUnavailable, response)
		return
	}

	handleBasicRateLimiting(limiter, responseWriter, request, next)
}

func IPRateLimiterMiddleware(requestsPerSecond int64) Middleware {
	return createRateLimitMiddleware(
		func() (interface{}, error) {
			return NewIPRateLimiter(requestsPerSecond, time.Second)
		},
		requestsPerSecond,
		func(limiter interface{}, responseWriter http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
			handleIPRateLimiting(limiter.(*IPRateLimiter), responseWriter, request, next)
		},
	)
}
