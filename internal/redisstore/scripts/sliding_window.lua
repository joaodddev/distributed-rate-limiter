-- KEYS[1] = chave do rate limit (ex: "ratelimit:user-123")
-- ARGV[1] = limite máximo de requisições na janela
-- ARGV[2] = duração da janela em milissegundos
-- ARGV[3] = timestamp atual em milissegundos
-- ARGV[4] = quantidade de requisições a consumir (para AllowN)

local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local n = tonumber(ARGV[4])

local cutoff = now_ms - window_ms

-- remove entradas fora da janela (sliding window log)
redis.call("ZREMRANGEBYSCORE", key, "-inf", cutoff)

local count = redis.call("ZCARD", key)

if count + n > limit then
  -- pega a entrada mais antiga pra calcular retry_after
  local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
  local retry_after_ms = 0
  if #oldest > 0 then
    local oldest_score = tonumber(oldest[2])
    retry_after_ms = (oldest_score + window_ms) - now_ms
  end

  return {0, limit - count, retry_after_ms}
end

-- adiciona n entradas com score = timestamp atual
-- usa um member único por entrada (timestamp + índice) pra evitar colisão
for i = 1, n do
  redis.call("ZADD", key, now_ms, now_ms .. "-" .. i)
end

-- expira a chave automaticamente após a janela (evita memory leak no Redis)
redis.call("PEXPIRE", key, window_ms)

local remaining = limit - (count + n)
return {1, remaining, 0}
