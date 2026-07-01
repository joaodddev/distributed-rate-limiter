package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubLimiter struct {
	result Result
	err    error
}

func (s stubLimiter) Allow(ctx context.Context, key string) (Result, error) {
	return s.result, s.err
}

func TestRateLimit_AllowsRequest(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := RateLimit(Config{
		Limiter: stubLimiter{result: Result{Allowed: true, Remaining: 4, Limit: 5}},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if !handlerCalled {
		t.Fatal("next handler should have been called when allowed")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-RateLimit-Remaining") != "4" {
		t.Fatalf("expected X-RateLimit-Remaining=4, got %s", rec.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimit_DeniesRequest(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	mw := RateLimit(Config{
		Limiter: stubLimiter{result: Result{Allowed: false, Remaining: 0, Limit: 5, RetryAfter: 3}},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if handlerCalled {
		t.Fatal("next handler should NOT have been called when denied")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") != "3" {
		t.Fatalf("expected Retry-After=3, got %s", rec.Header().Get("Retry-After"))
	}
}

func TestRateLimit_FailsOpenOnLimiterError(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := RateLimit(Config{
		Limiter: stubLimiter{err: context.DeadlineExceeded},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if !handlerCalled {
		t.Fatal("next handler should be called (fail-open) when limiter errors")
	}
}

func TestByIP_PrefersForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")

	key := ByIP()(req)
	if key != "203.0.113.5" {
		t.Fatalf("expected X-Forwarded-For to take precedence, got %s", key)
	}
}
