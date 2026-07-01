package ratelimiter

import (
	"context"
	"time"
)

// SlidingWindowLimiter implementa rate limiting usando o algoritmo
// sliding window log: mantém um timestamp por requisição dentro da janela.
type SlidingWindowLimiter struct {
	limit  int64
	window time.Duration
	log    map[string][]time.Time
	now    func() time.Time
}

// NewSlidingWindowLimiter cria um limiter que permite até `limit`
// requisições por `window` de tempo.
func NewSlidingWindowLimiter(limit int64, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:  limit,
		window: window,
		log:    make(map[string][]time.Time),
		now:    time.Now,
	}
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (Result, error) {
	return l.AllowN(ctx, key, 1)
}

func (l *SlidingWindowLimiter) AllowN(ctx context.Context, key string, n int64) (Result, error) {
	now := l.now()
	cutoff := now.Add(-l.window)

	entries := l.log[key]
	entries = dropBefore(entries, cutoff)

	count := int64(len(entries))
	if count+n > l.limit {
		retryAfter := time.Duration(0)
		if len(entries) > 0 {
			retryAfter = entries[0].Add(l.window).Sub(now)
		}
		l.log[key] = entries
		return Result{
			Allowed:    false,
			Remaining:  max64(l.limit-count, 0),
			Limit:      l.limit,
			RetryAfter: retryAfter,
			ResetAt:    now.Add(l.window),
		}, nil
	}

	for i := int64(0); i < n; i++ {
		entries = append(entries, now)
	}
	l.log[key] = entries

	return Result{
		Allowed:   true,
		Remaining: l.limit - int64(len(entries)),
		Limit:     l.limit,
		ResetAt:   now.Add(l.window),
	}, nil
}

func dropBefore(entries []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for i < len(entries) && entries[i].Before(cutoff) {
		i++
	}
	return entries[i:]
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
