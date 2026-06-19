package api

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(ip string) (allowed bool, remaining int, resetAt int64) {
	if rl.limit <= 0 {
		return true, -1, 0
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	windowStart := now.Add(-rl.window)
	entries := rl.requests[ip]
	var active []time.Time
	for _, t := range entries {
		if t.After(windowStart) {
			active = append(active, t)
		}
	}
	if len(active) >= rl.limit {
		if len(active) > 0 {
			rl.requests[ip] = active
			resetAt = active[0].Add(rl.window).Unix()
		} else {
			delete(rl.requests, ip)
			resetAt = now.Add(rl.window).Unix()
		}
		return false, 0, resetAt
	}
	remaining = rl.limit - len(active) - 1
	if len(active) > 0 {
		resetAt = active[0].Add(rl.window).Unix()
	} else {
		resetAt = now.Add(rl.window).Unix()
	}
	active = append(active, now)
	rl.requests[ip] = active
	return true, remaining, resetAt
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.rateLimiter == nil {
			next.ServeHTTP(w, r)
			return
		}
		ip := stripPort(r.RemoteAddr)
		allowed, remaining, resetAt := s.rateLimiter.Allow(ip)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(s.rateLimiter.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))
		if !allowed {
			w.Header().Set("Retry-After", "60")
			writeError(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func stripPort(addr string) string {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return addr
	}
	return addr[:idx]
}
