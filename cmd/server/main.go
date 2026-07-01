package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/joaodddev/distributed-rate-limiter/internal/config"
	"github.com/joaodddev/distributed-rate-limiter/internal/middleware"
	"github.com/joaodddev/distributed-rate-limiter/internal/ratelimiter"
	"github.com/joaodddev/distributed-rate-limiter/internal/redisstore"
)

// redisLimiterAdapter conecta o Store (Redis) diretamente à interface
// ratelimiter.Limiter, sem passar pela implementação in-memory.
type redisLimiterAdapter struct {
	store  *redisstore.Store
	limit  int64
	window time.Duration
}

func (a *redisLimiterAdapter) Allow(ctx context.Context, key string) (ratelimiter.Result, error) {
	return a.AllowN(ctx, key, 1)
}

func (a *redisLimiterAdapter) AllowN(ctx context.Context, key string, n int64) (ratelimiter.Result, error) {
	res, err := a.store.Eval(ctx, key, a.limit, a.window, n)
	if err != nil {
		return ratelimiter.Result{}, err
	}
	return ratelimiter.Result{
		Allowed:    res.Allowed,
		Remaining:  res.Remaining,
		Limit:      a.limit,
		RetryAfter: res.RetryAfter,
		ResetAt:    time.Now().Add(a.window),
	}, nil
}

func main() {
	cfg := config.Load()

	store := redisstore.NewStore(redisstore.Config{Addr: cfg.RedisAddr})
	defer store.Close()

	limiter := &redisLimiterAdapter{
		store:  store,
		limit:  cfg.RateLimit,
		window: cfg.RateWindow,
	}

	rateLimitMW := middleware.RateLimit(middleware.Config{
		Limiter:      middleware.NewLimiterAdapter(limiter),
		KeyExtractor: middleware.ByIP(),
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})

	handler := rateLimitMW(mux)

	log.Printf("server listening on :%s (limit=%d/%s)\n", cfg.Port, cfg.RateLimit, cfg.RateWindow)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
