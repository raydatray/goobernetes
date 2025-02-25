package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RateLimiter struct {
	requests   atomic.Int64
	limit      int64
	windowSize time.Duration
	lastReset  atomic.Int64
}

var (
	ErrBadRequest = errors.New("400 Bad Request")
)

func NewRateLimiter(limit int64, windowSize time.Duration) (*RateLimiter, error) {
	if limit < 0 {
		return nil, ErrBadRequest
	}

	if windowSize < 0 {
		return nil, ErrBadRequest
	}

	rl := &RateLimiter{
		limit:      limit,
		windowSize: windowSize,
	}
	rl.lastReset.Store(time.Now().UnixNano())
	return rl, nil
}

func (rl *RateLimiter) TryAcquire() bool {
	now := time.Now().UnixNano()
	windowStart := rl.lastReset.Load()

	if time.Duration(now-windowStart) >= rl.windowSize {
		rl.lastReset.Store(now)
		rl.requests.Store(0)
	}

	current := rl.requests.Add(1)
	if current > rl.limit {
		rl.requests.Add(-1)
		return false
	}

	return true
}

func (rl *RateLimiter) Reset() {
	rl.requests.Store(0)
	rl.lastReset.Store(time.Now().UnixNano())
}

func RateLimiterMiddleware(requestsPerSecond int64) Middleware {
	rateLimiter, err := NewRateLimiter(requestsPerSecond, time.Second)

	if err != nil {
		panic(fmt.Sprintf("Failed to create rate limiter: %v", err))
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.TryAcquire() {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerSecond))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

type IPRateLimiter struct {
	limiters sync.Map
	limit    int64
	window   time.Duration
}

func NewIPRateLimiter(limit int64, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		limit:  limit,
		window: window,
	}
}

func (irl *IPRateLimiter) getLimiter(ip string) *RateLimiter {
	limiter, exists := irl.limiters.Load(ip)

	if !exists {
		newLimiter, err := NewRateLimiter(irl.limit, irl.window)
		if err != nil {
			panic(fmt.Sprintf("Failed to create rate limiter: %v", err))
		}
		actual, loaded := irl.limiters.LoadOrStore(ip, newLimiter)
		if loaded {
			return actual.(*RateLimiter)
		}
		return newLimiter
	}
	return limiter.(*RateLimiter)
}

func IPRateLimiterMiddleware(requestsPerSecond int64) Middleware {
	ipRateLimiter := NewIPRateLimiter(requestsPerSecond, time.Second)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				ip = strings.Split(forwardedFor, ",")[0]
			}

			limiter := ipRateLimiter.getLimiter(ip)
			if !limiter.TryAcquire() {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerSecond))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}
