// go:build integration

package redisstore

import (
	"context"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store := NewStore(Config{Addr: "localhost:6380"})
	t.Cleanup(func() {
		store.Close()
	})
	return store
}

func TestStore_Eval_AllowsWithinLimit(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	key := "test:allows-within-limit"

	for i := 0; i < 3; i++ {
		res, err := store.Eval(ctx, key, 3, time.Second, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestStore_Eval_DeniesOverLimit(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	key := "test:denies-over-limit"

	for i := 0; i < 2; i++ {
		if _, err := store.Eval(ctx, key, 2, time.Second, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	res, err := store.Eval(ctx, key, 2, time.Second, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Allowed {
		t.Fatal("3rd request should be denied when limit is 2")
	}
	if res.RetryAfter <= 0 {
		t.Fatal("RetryAfter should be positive when denied")
	}
}

func TestStore_Eval_ResetsAfterWindow(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	key := "test:resets-after-window"

	if _, err := store.Eval(ctx, key, 1, 300*time.Millisecond, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, _ := store.Eval(ctx, key, 1, 300*time.Millisecond, 1)
	if res.Allowed {
		t.Fatal("second request within window should be denied")
	}

	time.Sleep(400 * time.Millisecond)

	res, err := store.Eval(ctx, key, 1, 300*time.Millisecond, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Fatal("request after window reset should be allowed")
	}
}

func TestStore_Eval_ConcurrentSameKey(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	key := "test:concurrent-same-key"

	const goroutines = 50
	results := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			res, err := store.Eval(ctx, key, 10, time.Second, 1)
			if err != nil {
				results <- false
				return
			}
			results <- res.Allowed
		}()
	}

	allowedCount := 0
	for i := 0; i < goroutines; i++ {
		if <-results {
			allowedCount++
		}
	}

	if allowedCount != 10 {
		t.Fatalf("expected exactly 10 allowed requests out of %d concurrent, got %d", goroutines, allowedCount)
	}
}
