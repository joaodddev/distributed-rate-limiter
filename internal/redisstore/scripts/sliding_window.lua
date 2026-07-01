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
local seq_key = key .. ":seq"

redis.call("ZREMRANGEBYSCORE", key, "-inf", cutoff)

local count = redis.call("ZCARD", key)

if count + n > limit then
  local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
  local retry_after_ms = 0
  if #oldest > 0 then
    local oldest_score = tonumber(oldest[2])
    retry_after_ms = (oldest_score + window_ms) - now_ms
  end

  return {0, limit - count, retry_after_ms}
end

-- INCR garante um sufixo único por entrada, mesmo quando múltiplas
-- chamadas concorrentes caem no mesmo timestamp (mesmo milissegundo).
for i = 1, n do
  local seq = redis.call("INCR", seq_key)
  redis.call("ZADD", key, now_ms, now_ms .. "-" .. seq)
end

redis.call("PEXPIRE", key, window_ms)
redis.call("PEXPIRE", seq_key, window_ms)

local remaining = limit - (count + n)
return {1, remaining, 0}
