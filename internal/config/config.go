package config

import (
	"os"
	"strconv"
	"time"
)

// Config centraliza as variáveis de ambiente do servidor.
type Config struct {
	Port       string
	RedisAddr  string
	RateLimit  int64
	RateWindow time.Duration
}

// Load carrega a configuração a partir de variáveis de ambiente,
// aplicando defaults sensatos para desenvolvimento local.
func Load() Config {
	return Config{
		Port:       getEnv("PORT", "8080"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6380"),
		RateLimit:  getEnvInt64("RATE_LIMIT", 10),
		RateWindow: getEnvDuration("RATE_WINDOW", time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
