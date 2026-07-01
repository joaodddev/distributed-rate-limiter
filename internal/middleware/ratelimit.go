package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

// Limiter é o subconjunto da interface ratelimiter.Limiter que o
// middleware precisa — evita import direto do pacote ratelimiter
// e mantém o middleware desacoplado da implementação concreta.
type Limiter interface {
	Allow(ctx context.Context, key string) (Result, error)
}

// Result espelha ratelimiter.Result para evitar acoplamento direto.
type Result struct {
	Allowed    bool
	Remaining  int64
	Limit      int64
	RetryAfter int64 // em segundos, arredondado
}

// Config configura o comportamento do middleware de rate limit.
type Config struct {
	Limiter      Limiter
	KeyExtractor KeyExtractor
	// OnDenied permite customizar a resposta quando a requisição é
	// negada. Se nil, usa a resposta padrão em JSON.
	OnDenied func(w http.ResponseWriter, r *http.Request, res Result)
}

// RateLimit retorna um middleware net/http que aplica rate limiting
// usando o Limiter e KeyExtractor configurados.
func RateLimit(cfg Config) func(http.Handler) http.Handler {
	if cfg.KeyExtractor == nil {
		cfg.KeyExtractor = ByIP()
	}
	if cfg.OnDenied == nil {
		cfg.OnDenied = defaultOnDenied
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.KeyExtractor(r)

			res, err := cfg.Limiter.Allow(r.Context(), key)
			if err != nil {
				// falha no backend do limiter não deve derrubar a API;
				// loga e deixa passar (fail-open) é a escolha mais segura
				// pra disponibilidade — fail-closed pode ser configurável depois.
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(res.Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(res.Remaining, 10))

			if !res.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(res.RetryAfter, 10))
				cfg.OnDenied(w, r, res)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func defaultOnDenied(w http.ResponseWriter, r *http.Request, res Result) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":       "rate limit exceeded",
		"retry_after": res.RetryAfter,
	})
}
