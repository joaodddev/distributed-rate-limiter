package main

import (
	"context"
	"log"
	"time"

	"github.com/joaodddev/distributed-rate-limiter/internal/redisstore"
)

func main() {
	ctx := context.Background()

	store := redisstore.NewStore(redisstore.Config{
		Addr: "localhost:6380", // porta mapeada no docker-compose
	})
	defer store.Close()

	const key = "ratelimit:manual-test"
	const limit = int64(3)
	const window = time.Second

	log.Printf("testando sliding window: limit=%d window=%s\n", limit, window)

	// dispara 5 requisições seguidas — as 3 primeiras devem passar,
	// as 2 últimas devem ser negadas
	for i := 1; i <= 5; i++ {
		res, err := store.Eval(ctx, key, limit, window, 1)
		if err != nil {
			log.Fatalf("erro na requisição %d: %v", i, err)
		}

		status := "ALLOWED"
		if !res.Allowed {
			status = "DENIED"
		}

		log.Printf("req %d -> %s | remaining=%d retry_after=%s\n",
			i, status, res.Remaining, res.RetryAfter)
	}

	log.Println("aguardando janela expirar (1.2s)...")
	time.Sleep(1200 * time.Millisecond)

	res, err := store.Eval(ctx, key, limit, window, 1)
	if err != nil {
		log.Fatalf("erro na requisição pós-janela: %v", err)
	}
	log.Printf("req pós-janela -> allowed=%v remaining=%d\n", res.Allowed, res.Remaining)
}
