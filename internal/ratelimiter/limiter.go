package ratelimiter

import (
	"context"
	"time"
)

// Result representa o resultado de uma verificação de rate limit.
type Result struct {
	// Allowed indica se a requisição pode prosseguir.
	Allowed bool

	// Remaining é o número de requisições restantes na janela atual.
	Remaining int64

	// Limit é o limite máximo configurado para a janela.
	Limit int64

	// RetryAfter indica quanto tempo esperar antes de tentar novamente.
	// Só é relevante quando Allowed == false.
	RetryAfter time.Duration

	// ResetAt indica quando a janela atual será resetada.
	ResetAt time.Time
}

// Limiter define o contrato para qualquer implementação de rate limiting,
// seja em memória, distribuída via Redis, ou outra estratégia.
type Limiter interface {
	// Allow verifica se uma requisição identificada por key pode prosseguir,
	// consumindo uma unidade da cota se permitido.
	Allow(ctx context.Context, key string) (Result, error)

	// AllowN verifica se n requisições podem prosseguir de uma vez,
	// útil para operações em lote.
	AllowN(ctx context.Context, key string, n int64) (Result, error)
}
