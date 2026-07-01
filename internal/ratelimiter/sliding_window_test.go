package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestSlidingWindowLimiter_AllowsWithinLimit(t *testing.T) {
	l := NewSlidingWindowLimiter(3, time.Second)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		res, err := l.Allow(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Allowed {
			t.Fatalf("request %d should be allowed, got denied", i+1)
		}
	}
}

func TestSlidingWindowLimiter_DeniesOverLimit(t *testing.T) {
	l := NewSlidingWindowLimiter(2, time.Second)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if _, err := l.Allow(ctx, "user-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	res, err := l.Allow(ctx, "user-1")
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

func TestSlidingWindowLimiter_ResetsAfterWindow(t *testing.T) {
	fakeNow := time.Now()
	l := NewSlidingWindowLimiter(1, 100*time.Millisecond)
	l.now = func() time.Time { return fakeNow }
	ctx := context.Background()

	res, _ := l.Allow(ctx, "user-1")
	if !res.Allowed {
		t.Fatal("first request should be allowed")
	}

	res, _ = l.Allow(ctx, "user-1")
	if res.Allowed {
		t.Fatal("second request within window should be denied")
	}

	// avança o relógio fake além da janela
	fakeNow = fakeNow.Add(150 * time.Millisecond)
	res, _ = l.Allow(ctx, "user-1")
	if !res.Allowed {
		t.Fatal("request after window reset should be allowed")
	}
}

func TestSlidingWindowLimiter_IsolatesKeys(t *testing.T) {
	l := NewSlidingWindowLimiter(1, time.Second)
	ctx := context.Background()

	res1, _ := l.Allow(ctx, "user-1")
	res2, _ := l.Allow(ctx, "user-2")

	if !res1.Allowed || !res2.Allowed {
		t.Fatal("different keys should have independent limits")
	}
}

func TestSlidingWindowLimiter_AllowN_Burst(t *testing.T) {
	l := NewSlidingWindowLimiter(5, time.Second)
	ctx := context.Background()

	res, err := l.AllowN(ctx, "user-1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Fatal("burst of 5 within limit of 5 should be allowed")
	}

	res, _ = l.Allow(ctx, "user-1")
	if res.Allowed {
		t.Fatal("request after exact burst should be denied")
	}
}
