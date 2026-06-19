package api

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)
	ip := "192.168.1.1"

	allowed, rem, _ := rl.Allow(ip)
	if !allowed {
		t.Fatal("expected allow (1st)")
	}
	if rem != 1 {
		t.Fatalf("expected remaining 1, got %d", rem)
	}

	allowed, rem, _ = rl.Allow(ip)
	if !allowed {
		t.Fatal("expected allow (2nd)")
	}
	if rem != 0 {
		t.Fatalf("expected remaining 0, got %d", rem)
	}

	allowed, rem, _ = rl.Allow(ip)
	if allowed {
		t.Fatal("expected deny (3rd)")
	}
	if rem != 0 {
		t.Fatalf("expected remaining 0, got %d", rem)
	}
}

func TestRateLimiterNoLimit(t *testing.T) {
	rl := NewRateLimiter(0, 1*time.Minute)
	for range 100 {
		allowed, rem, _ := rl.Allow("test")
		if !allowed {
			t.Fatal("expected allow (unlimited)")
		}
		if rem != -1 {
			t.Fatalf("expected remaining -1 for unlimited, got %d", rem)
		}
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, 1*time.Minute)

	allowed, _, _ := rl.Allow("10.0.0.1")
	if !allowed {
		t.Fatal("expected allow")
	}
	allowed, _, _ = rl.Allow("10.0.0.2")
	if !allowed {
		t.Fatal("expected allow for different IP")
	}
	allowed, _, _ = rl.Allow("10.0.0.1")
	if allowed {
		t.Fatal("expected deny for first IP")
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	allowed, _, _ := rl.Allow("ip")
	if !allowed {
		t.Fatal("expected allow")
	}
	allowed, _, _ = rl.Allow("ip")
	if allowed {
		t.Fatal("expected deny")
	}

	time.Sleep(60 * time.Millisecond)

	allowed, _, _ = rl.Allow("ip")
	if !allowed {
		t.Fatal("expected allow after window reset")
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute)

	var wg sync.WaitGroup
	results := make(chan bool, 50)

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _, _ := rl.Allow("10.0.0.1")
			results <- allowed
		}()
	}

	wg.Wait()
	close(results)

	allowed := 0
	for r := range results {
		if r {
			allowed++
		}
	}
	if allowed != 10 {
		t.Fatalf("expected exactly 10 allowed, got %d", allowed)
	}
}

func TestRateLimitMiddlewareNoLimiter(t *testing.T) {
	srv := &Server{rateLimiter: nil}
	handler := srv.rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimitMiddlewareBlocks(t *testing.T) {
	srv := &Server{rateLimiter: NewRateLimiter(1, 1*time.Minute)}
	handler := srv.rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}
	if w1.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("expected X-RateLimit-Remaining: 0, got %q", w1.Header().Get("X-RateLimit-Remaining"))
	}
	if w1.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatal("expected X-RateLimit-Reset header")
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
	if w2.Header().Get("Retry-After") != "60" {
		t.Fatalf("expected Retry-After: 60, got %q", w2.Header().Get("Retry-After"))
	}
	if w2.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("expected X-RateLimit-Remaining: 0, got %q", w2.Header().Get("X-RateLimit-Remaining"))
	}
	if w2.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatal("expected X-RateLimit-Reset header")
	}
	limit := w2.Header().Get("X-RateLimit-Limit")
	if limit != "1" {
		t.Fatalf("expected X-RateLimit-Limit: 1, got %q", limit)
	}
}

func TestRateLimitHeaders(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Minute)
	allowed, rem, reset := rl.Allow("10.0.0.1")
	if !allowed {
		t.Fatal("expected allow")
	}
	if rem != 4 {
		t.Fatalf("expected remaining 4, got %d", rem)
	}
	if reset <= time.Now().Unix() {
		t.Fatalf("expected future reset, got %d", reset)
	}
	if strconv.FormatInt(reset, 10) == "" {
		t.Fatal("expected non-empty reset string")
	}
}
