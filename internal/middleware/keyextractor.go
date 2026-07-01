package middleware

import (
	"net"
	"net/http"
)

// KeyExtractor extrai a chave de identificação usada para aplicar
// o rate limit a partir de uma requisição HTTP.
type KeyExtractor func(r *http.Request) string

// ByIP extrai o IP remoto do cliente como chave, considerando
// X-Forwarded-For quando presente (ex: atrás de load balancer/proxy).
func ByIP() KeyExtractor {
	return func(r *http.Request) string {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			return fwd
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr
		}
		return host
	}
}

// ByHeader extrai o valor de um header específico como chave,
// útil para rate limit por API key ou tenant ID.
func ByHeader(name string) KeyExtractor {
	return func(r *http.Request) string {
		return r.Header.Get(name)
	}
}
