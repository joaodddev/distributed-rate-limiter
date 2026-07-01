package middleware

import (
	"context"

	"github.com/joaodddev/distributed-rate-limiter/internal/ratelimiter"
)

// LimiterAdapter adapta um ratelimiter.Limiter para a interface
// Limiter esperada pelo middleware, convertendo entre os tipos Result.
type LimiterAdapter struct {
	limiter ratelimiter.Limiter
}

// NewLimiterAdapter cria um adapter em torno de um ratelimiter.Limiter.
func NewLimiterAdapter(l ratelimiter.Limiter) *LimiterAdapter {
	return &LimiterAdapter{limiter: l}
}

func (a *LimiterAdapter) Allow(ctx context.Context, key string) (Result, error) {
	res, err := a.limiter.Allow(ctx, key)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Allowed:    res.Allowed,
		Remaining:  res.Remaining,
		Limit:      res.Limit,
		RetryAfter: int64(res.RetryAfter.Seconds()),
	}, nil
}
