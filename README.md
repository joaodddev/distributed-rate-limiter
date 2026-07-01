# distributed-rate-limiter

Rate limiter distribuído em Go, usando o algoritmo sliding window log
com execução atômica via Lua script no Redis — sem race conditions,
sem locks distribuídos, sem round-trips extras.

## Stack

Go 1.22 · Redis 7 · go-redis/v9 · net/http puro (sem framework)

## Por que sliding window log + Lua

Fixed window sofre com bursts na fronteira entre janelas (2x o limite
em picos). Sliding window log resolve isso mantendo um registro
preciso de timestamps, mas exige leitura+escrita atômica — daí o Lua
script: uma única chamada `EVALSHA` executa toda a lógica (ZREMRANGEBYSCORE,
ZCARD, ZADD) sem expor uma janela de tempo entre leitura e escrita
onde outra goroutine/instância poderia interferir.

## Arquitetura

\`\`\`
HTTP Request → middleware.RateLimit → ratelimiter.Limiter → redisstore.Store → Redis (Lua atômico)
\`\`\`

- `internal/ratelimiter` — interface + implementação in-memory (referência/testes)
- `internal/redisstore` — client Redis + script Lua embutido via go:embed
- `internal/middleware` — middleware net/http plugável, com key extractor configurável
- `internal/config` — configuração via variáveis de ambiente
