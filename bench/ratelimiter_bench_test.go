package bench

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joaodddev/distributed-rate-limiter/internal/ratelimiter"
	"github.com/joaodddev/distributed-rate-limiter/internal/redisstore"
)

func BenchmarkInMemory_SingleKey(b *testing.B) {
	limiter := ratelimiter.NewSlidingWindowLimiter(1_000_000, time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = limiter.Allow(ctx, "bench-key")
	}
}

func BenchmarkInMemory_ParallelSingleKey(b *testing.B) {
	limiter := ratelimiter.NewSlidingWindowLimiter(1_000_000, time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = limiter.Allow(ctx, "bench-key")
		}
	})
}

func BenchmarkInMemory_ParallelMultiKey(b *testing.B) {
	limiter := ratelimiter.NewSlidingWindowLimiter(1_000_000, time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d", i%100)
			_, _ = limiter.Allow(ctx, key)
			i++
		}
	})
}

func BenchmarkRedis_SingleKey(b *testing.B) {
	store := redisstore.NewStore(redisstore.Config{Addr: "localhost:6380"})
	defer store.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Eval(ctx, "bench-key-redis", 1_000_000, time.Minute, 1)
	}
}

func BenchmarkRedis_ParallelSingleKey(b *testing.B) {
	store := redisstore.NewStore(redisstore.Config{Addr: "localhost:6380"})
	defer store.Close()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Eval(ctx, "bench-key-redis-parallel", 1_000_000, time.Minute, 1)
		}
	})
}

func BenchmarkRedis_ParallelMultiKey(b *testing.B) {
	store := redisstore.NewStore(redisstore.Config{Addr: "localhost:6380"})
	defer store.Close()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-redis-%d", i%100)
			_, _ = store.Eval(ctx, key, 1_000_000, time.Minute, 1)
			i++
		}
	})
}
