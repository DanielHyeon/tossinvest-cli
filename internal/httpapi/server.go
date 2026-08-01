package httpapi

import (
	"errors"
	"net/http"
	"strconv"
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
		// HTTP/2 may represent a bodyless END_STREAM request with a non-NoBody
		// EOF reader. ContentLength is the protocol-neutral inbound signal:
		// zero is known empty; unknown (-1) and declared (>0) stay bounded.
		if requestHasBody(r) {
			r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// requestHasBody classifies a server-parsed request as declared or unknown-length.
// A positive or malformed Content-Length preserved by an HTTP/2 parser remains
// fail-closed even when net/http normalizes the ContentLength field to zero.
func requestHasBody(request *http.Request) bool {
	if request == nil {
		return false
	}
	if request.ContentLength != 0 {
		return true
	}
	for _, raw := range request.Header.Values("Content-Length") {
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			for _, digit := range value {
				if digit < '0' || digit > '9' {
					return true
				}
			}
			length, err := strconv.ParseUint(value, 10, 64)
			if err != nil || length != 0 {
				return true
			}
		}
	}
	return false
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
