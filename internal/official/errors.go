package official

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
)

// Sentinel errors returned by the official API client.
var (
	ErrTransport       = errors.New("official: transport error")
	ErrAuth            = errors.New("official: authentication failed")
	ErrIPNotAllowed    = errors.New("official: IP not allowed")
	ErrRateLimited     = errors.New("official: rate limited")
	ErrServer          = errors.New("official: server error")
	ErrAuthorityOrigin = errors.New("official: authoritative origin unavailable")
)

// APIError represents a non-retryable 4xx error from the official API.
// It is a passthrough and does NOT trigger fallback to the unofficial client.
type APIError struct {
	Code int
	Body string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("official: API error %d: %s", e.Code, e.Body)
}

// reIPWord matches the standalone word "ip" (case-insensitive after lowercasing).
// Compiled once to avoid repeated allocation on every error path.
var reIPWord = regexp.MustCompile(`\bip\b`)

// classifyStatus maps an HTTP status code (and optional response body) to a
// typed error. 401/403 become ErrIPNotAllowed when the body mentions "ip" as a
// whole word (case-insensitive), otherwise ErrAuth. 429 → ErrRateLimited. ≥500 →
// ErrServer. All other 4xx → *APIError (passthrough, no fallback).
func classifyStatus(code int, body []byte) error {
	switch {
	case code == 401 || code == 403:
		// Carry the code, never the body. 401 and 403 point at different causes —
		// a token another process rotated versus a credential or address the
		// broker will not accept — and collapsing them into one sentinel is how
		// change a082's token race read as "authentication failed" for three days
		// with nothing to tell it apart from a permissions problem. The body stays
		// out because it can carry account identifiers and this error is logged.
		//
		// The wrap is %w, so errors.Is keeps deciding what counts as a refusal.
		if reIPWord.Match(bytes.ToLower(body)) {
			return fmt.Errorf("%w (HTTP %d)", ErrIPNotAllowed, code)
		}
		return fmt.Errorf("%w (HTTP %d)", ErrAuth, code)
	case code == 429:
		return ErrRateLimited
	case code >= 500:
		return ErrServer
	default:
		return &APIError{Code: code, Body: string(body)}
	}
}

// ShouldFallback reports whether err indicates the official API is unavailable
// or the request should be retried via the unofficial (web-session) client.
// Returns true for sentinel errors; false for *APIError passthroughs.
func ShouldFallback(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return false
	}
	return errors.Is(err, ErrTransport) ||
		errors.Is(err, ErrAuth) ||
		errors.Is(err, ErrIPNotAllowed) ||
		errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrServer)
}
