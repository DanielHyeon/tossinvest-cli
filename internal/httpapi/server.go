package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	// MaxRequestBodyBytes is deliberately small because every mutation accepts a
	// finite, server-owned preset. Read resources and the stream accept no body.
	MaxRequestBodyBytes int64 = 256 << 10

	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultReadTimeout       = 5 * time.Second
	DefaultMaxHeaderBytes    = 32 << 10
)

// LimitRequestBody applies the process-wide input ceiling before any route can
// decode a request. It does not interpret the body or broaden the router's
// mutation allowlist.
func LimitRequestBody(next http.Handler) http.Handler {
	if next == nil {
		next = http.NotFoundHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Preserve the standard no-body sentinel. Read and stream routes use its
		// identity to reject actual bodies, and wrapping it would turn every
		// legitimate GET/HEAD into BODY_NOT_SUPPORTED.
		if r.Body != nil && r.Body != http.NoBody {
			r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// NewServer pins the resource limits that are part of the public daemon
// contract. TLS and listener ownership remain with cmd/tossctl so startup can
// validate the private network boundary before opening a socket.
func NewServer(addr string, handler http.Handler) (*http.Server, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("httpapi: server address is required")
	}
	if handler == nil {
		return nil, errors.New("httpapi: server handler is required")
	}
	return &http.Server{
		Addr:              addr,
		Handler:           LimitRequestBody(handler),
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
	}, nil
}
