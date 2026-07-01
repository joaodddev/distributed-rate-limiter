package redisstore

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed scripts/sliding_window.lua
var slidingWindowScript string

// Store encapsula o acesso ao Redis para operações de rate limiting.
type Store struct {
	client *redis.Client
	script *redis.Script
}

// Config define os parâmetros de conexão com o Redis.
type Config struct {
	Addr     string
	Password string
	DB       int
}

// NewStore cria um novo Store conectado ao Redis, com o script Lua
// já preparado (mas não necessariamente carregado no servidor ainda).
func NewStore(cfg Config) *Store {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Store{
		client: client,
		script: redis.NewScript(slidingWindowScript),
	}
}

// EvalResult representa o retorno bruto do script Lua.
type EvalResult struct {
	Allowed    bool
	Remaining  int64
	RetryAfter time.Duration
}

// Eval executa o script de sliding window de forma atômica no Redis.
func (s *Store) Eval(ctx context.Context, key string, limit int64, window time.Duration, n int64) (EvalResult, error) {
	now := time.Now().UnixMilli()
	windowMs := window.Milliseconds()

	raw, err := s.script.Run(ctx, s.client, []string{key}, limit, windowMs, now, n).Result()
	if err != nil {
		return EvalResult{}, fmt.Errorf("redisstore: eval sliding window script: %w", err)
	}

	return parseEvalResult(raw)
}

func parseEvalResult(raw interface{}) (EvalResult, error) {
	vals, ok := raw.([]interface{})
	if !ok || len(vals) != 3 {
		return EvalResult{}, fmt.Errorf("redisstore: unexpected script result shape: %v", raw)
	}

	allowed, ok1 := vals[0].(int64)
	remaining, ok2 := vals[1].(int64)
	retryAfterMs, ok3 := vals[2].(int64)
	if !ok1 || !ok2 || !ok3 {
		return EvalResult{}, fmt.Errorf("redisstore: unexpected script result types: %v", vals)
	}

	return EvalResult{
		Allowed:    allowed == 1,
		Remaining:  remaining,
		RetryAfter: time.Duration(retryAfterMs) * time.Millisecond,
	}, nil
}

// Close encerra a conexão com o Redis.
func (s *Store) Close() error {
	return s.client.Close()
}
