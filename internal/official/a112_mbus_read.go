package official

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode/utf8"
)

// A112MBUSResult is one opaque, bounded measurement response. It is only for
// the A112 source measurement tool; it is not production market evidence.
type A112MBUSResult struct {
	method         string
	path           string
	canonicalQuery string
	statusCode     int
	body           []byte
	cursorJSON     []byte
	cursorValue    []byte
	rateHeaders    map[string][]string
}

func (r A112MBUSResult) Method() string         { return r.method }
func (r A112MBUSResult) Path() string           { return r.path }
func (r A112MBUSResult) CanonicalQuery() string { return r.canonicalQuery }
func (r A112MBUSResult) StatusCode() int        { return r.statusCode }
func (r A112MBUSResult) Body() []byte           { return bytes.Clone(r.body) }
func (r A112MBUSResult) CursorJSON() []byte     { return bytes.Clone(r.cursorJSON) }
func (r A112MBUSResult) CursorValue() []byte    { return bytes.Clone(r.cursorValue) }

// RateHeader returns the unparsed values as received for an allowed header.
func (r A112MBUSResult) RateHeader(name string) []string {
	return append([]string(nil), r.rateHeaders[http.CanonicalHeaderKey(name)]...)
}

// RateHeaders returns a deep copy so callers cannot change retained evidence.
func (r A112MBUSResult) RateHeaders() map[string][]string {
	out := make(map[string][]string, len(r.rateHeaders))
	for key, values := range r.rateHeaders {
		out[key] = append([]string(nil), values...)
	}
	return out
}

// A112MBUSHoldError marks a fail-closed measurement result. Its message is
// diagnostic only and must never be promoted to official market evidence.
type A112MBUSHoldError struct{ reason string }

func (e *A112MBUSHoldError) Error() string { return "a112 m-b us measurement hold: " + e.reason }

func a112MBUSHoldf(format string, args ...any) error {
	return &A112MBUSHoldError{reason: fmt.Sprintf(format, args...)}
}

const a112MBUSBodyLimit = 2 << 20

type a112MBUSDescriptor struct {
	path          string
	query         url.Values
	requireCursor bool
}

// A112MBUSCandle performs one exact US/AAPL one-minute source measurement GET.
// before is an already-decoded cursor value; it is URL-escaped exactly once.
func A112MBUSCandle(ctx context.Context, client *Client, before []byte) (A112MBUSResult, error) {
	if !utf8.Valid(before) {
		return A112MBUSResult{}, a112MBUSHoldf("initial cursor is not valid UTF-8")
	}
	query := url.Values{"symbol": {"AAPL"}, "interval": {RawCandleIntervalOneMinute}, "count": {"200"}, "adjusted": {"false"}}
	if len(before) != 0 {
		query.Set("before", string(before))
	}
	return a112MBUSRead(ctx, client, a112MBUSDescriptor{path: "/api/v1/candles", query: query, requireCursor: true})
}

// A112MBUSOrderbook performs one exact raw AAPL orderbook measurement GET.
func A112MBUSOrderbook(ctx context.Context, client *Client) (A112MBUSResult, error) {
	return a112MBUSRead(ctx, client, a112MBUSDescriptor{path: "/api/v1/orderbook", query: url.Values{"symbol": {"AAPL"}}})
}

// A112MBUSCalendar performs one exact explicit-date US calendar measurement GET.
func A112MBUSCalendar(ctx context.Context, client *Client, date string) (A112MBUSResult, error) {
	if err := validateMarketCalendarDate(date); err != nil || date == "" {
		return A112MBUSResult{}, a112MBUSHoldf("calendar date must be exact YYYY-MM-DD")
	}
	return a112MBUSRead(ctx, client, a112MBUSDescriptor{path: "/api/v1/market-calendar/US", query: url.Values{"date": {date}}})
}

func a112MBUSRead(ctx context.Context, client *Client, descriptor a112MBUSDescriptor) (A112MBUSResult, error) {
	return a112MBUSReadAt(ctx, client, descriptor, time.Now)
}

// a112MBUSReadAt keeps time injectable only for token expiry and deadline
// enforcement. It is private so no caller can serialize or supply a clock as
// source/finality evidence.
func a112MBUSReadAt(ctx context.Context, client *Client, descriptor a112MBUSDescriptor, now func() time.Time) (A112MBUSResult, error) {
	if now == nil {
		return A112MBUSResult{}, a112MBUSHoldf("clock is required")
	}
	nowValue := now()
	deadline, ok := ctx.Deadline()
	if !ok || !a112MBUSDeadlineValid(deadline, nowValue) {
		return A112MBUSResult{}, a112MBUSHoldf("caller deadline must be positive and no later than 15 seconds")
	}
	if client == nil {
		return A112MBUSResult{}, a112MBUSHoldf("client is required")
	}

	client.configMu.RLock()
	defer client.configMu.RUnlock()
	if !client.authorityOriginLocked() || client.tm == nil || client.authorityTransport == nil || client.tm.base != client.base || client.tm.hc != client.hc {
		return A112MBUSResult{}, a112MBUSHoldf("sealed same-client authority origin is required")
	}
	transport := client.authorityTransport.Clone()
	base := client.base
	cacheFile := client.tm.cacheFile
	transport.Proxy = nil
	transport.DisableCompression = true

	token, err := a112MBUSCachedToken(cacheFile, nowValue)
	if err != nil {
		return A112MBUSResult{}, a112MBUSHoldf("cached token: %v", err)
	}

	endpoint, err := url.Parse(base + descriptor.path)
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" {
		return A112MBUSResult{}, a112MBUSHoldf("invalid authority endpoint")
	}
	endpoint.RawQuery = descriptor.query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return A112MBUSResult{}, a112MBUSHoldf("building GET: %v", err)
	}
	req.Header.Set(HeaderAuthorization, "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "identity")
	// An empty slice suppresses net/http's default Go user agent on the wire.
	req.Header["User-Agent"] = nil

	hc := &http.Client{
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("a112 measurement redirects are refused")
		},
	}
	resp, err := hc.Do(req)
	if err != nil {
		return A112MBUSResult{}, a112MBUSHoldf("one-shot GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if resp.StatusCode == http.StatusTooManyRequests {
			if err := a112MBUSValidate429(resp.Header); err != nil {
				return A112MBUSResult{}, a112MBUSHoldf("429 diagnostic: %v", err)
			}
		}
		return A112MBUSResult{}, a112MBUSHoldf("response status %d", resp.StatusCode)
	}
	if encoding := resp.Header.Values("Content-Encoding"); len(encoding) != 0 && !(len(encoding) == 1 && encoding[0] == "identity") {
		return A112MBUSResult{}, a112MBUSHoldf("non-identity content encoding")
	}
	if resp.ContentLength > a112MBUSBodyLimit {
		return A112MBUSResult{}, a112MBUSHoldf("content-length exceeds 2 MiB")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, a112MBUSBodyLimit+1))
	if err != nil {
		return A112MBUSResult{}, a112MBUSHoldf("reading response: %v", err)
	}
	if len(body) > a112MBUSBodyLimit {
		return A112MBUSResult{}, a112MBUSHoldf("response exceeds 2 MiB")
	}
	rateHeaders, err := a112MBUSValidateSuccessRateHeaders(resp.Header)
	if err != nil {
		return A112MBUSResult{}, a112MBUSHoldf("rate headers: %v", err)
	}
	cursorJSON, cursorValue, err := a112MBUSCursor(body, descriptor.requireCursor)
	if err != nil {
		return A112MBUSResult{}, a112MBUSHoldf("response cursor: %v", err)
	}
	return A112MBUSResult{
		method: http.MethodGet, path: descriptor.path, canonicalQuery: endpoint.RawQuery,
		statusCode: resp.StatusCode, body: bytes.Clone(body), cursorJSON: cursorJSON,
		cursorValue: cursorValue, rateHeaders: rateHeaders,
	}, nil
}

func a112MBUSDeadlineValid(deadline, now time.Time) bool {
	remaining := deadline.Sub(now)
	return remaining > 0 && remaining <= defaultTimeout
}

func a112MBUSCursor(body []byte, required bool) ([]byte, []byte, error) {
	envelope, err := a112MBUSStrictJSONObject(body)
	if err != nil {
		return nil, nil, fmt.Errorf("malformed result envelope: %w", err)
	}
	resultRaw, found := envelope["result"]
	if !found {
		return nil, nil, errors.New("result is absent")
	}
	result, err := a112MBUSStrictJSONObject(resultRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("malformed result payload: %w", err)
	}
	if !required {
		return nil, nil, nil
	}
	raw, found := result["nextBefore"]
	if !found || raw == nil {
		return nil, nil, errors.New("nextBefore is absent")
	}
	raw = bytes.Clone(raw)
	if bytes.Equal(raw, []byte("null")) {
		return raw, nil, nil
	}
	if err := a112MBUSValidateJSONStringSurrogates(raw); err != nil {
		return nil, nil, err
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || value == "" {
		return nil, nil, errors.New("nextBefore must be non-empty string or null")
	}
	return raw, []byte(value), nil
}

// a112MBUSValidateJSONStringSurrogates rejects escapes that encoding/json
// would otherwise turn into U+FFFD. A literal U+FFFD remains valid evidence;
// only malformed UTF-16 surrogate escape sequences are refused here.
func a112MBUSValidateJSONStringSurrogates(raw []byte) error {
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return errors.New("nextBefore is not a JSON string")
	}
	for index := 1; index < len(raw)-1; index++ {
		if raw[index] != '\\' {
			continue
		}
		index++
		if index >= len(raw)-1 {
			return errors.New("incomplete JSON escape")
		}
		if raw[index] != 'u' {
			continue
		}
		if index+4 >= len(raw)-1 {
			return errors.New("incomplete Unicode escape")
		}
		codepoint, ok := a112MBUSHex16(raw[index+1 : index+5])
		if !ok {
			return errors.New("invalid Unicode escape")
		}
		index += 4
		if codepoint < 0xD800 || codepoint > 0xDFFF {
			continue
		}
		if codepoint >= 0xDC00 {
			return errors.New("unpaired low surrogate")
		}
		if index+6 >= len(raw)-1 || raw[index+1] != '\\' || raw[index+2] != 'u' {
			return errors.New("unpaired high surrogate")
		}
		low, ok := a112MBUSHex16(raw[index+3 : index+7])
		if !ok || low < 0xDC00 || low > 0xDFFF {
			return errors.New("high surrogate is not followed by low surrogate")
		}
		index += 6
	}
	return nil
}

func a112MBUSHex16(raw []byte) (uint16, bool) {
	if len(raw) != 4 {
		return 0, false
	}
	var value uint16
	for _, byteValue := range raw {
		value <<= 4
		switch {
		case byteValue >= '0' && byteValue <= '9':
			value += uint16(byteValue - '0')
		case byteValue >= 'a' && byteValue <= 'f':
			value += uint16(byteValue-'a') + 10
		case byteValue >= 'A' && byteValue <= 'F':
			value += uint16(byteValue-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

// a112MBUSStrictJSONObject preserves each raw field value while rejecting
// duplicate keys and invalid source bytes. encoding/json's map decode is not
// suitable here because it silently applies last-wins replacement semantics.
func a112MBUSStrictJSONObject(source []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(source) {
		return nil, errors.New("source is not valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(source))
	opening, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := opening.(json.Delim); !ok || delimiter != '{' {
		return nil, errors.New("value is not an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("object key is not a string")
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate object key %q", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields[key] = bytes.Clone(value)
	}
	closing, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := closing.(json.Delim); !ok || delimiter != '}' {
		return nil, errors.New("object did not close")
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return nil, errors.New("trailing JSON value")
		}
		return nil, err
	}
	return fields, nil
}

func a112MBUSValidateSuccessRateHeaders(header http.Header) (map[string][]string, error) {
	allowed := []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"}
	out := make(map[string][]string, len(allowed))
	for _, name := range allowed {
		out[http.CanonicalHeaderKey(name)] = append([]string(nil), header.Values(name)...)
	}
	for _, name := range allowed[:3] {
		values := out[http.CanonicalHeaderKey(name)]
		if len(values) != 1 {
			return nil, fmt.Errorf("%s cardinality %d", name, len(values))
		}
		parsed, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || parsed < 0 || (name == "X-RateLimit-Limit" && parsed == 0) {
			return nil, fmt.Errorf("%s is not valid non-negative integer", name)
		}
	}
	if len(out[http.CanonicalHeaderKey("Retry-After")]) != 0 {
		return nil, errors.New("Retry-After must be absent on success")
	}
	return out, nil
}

func a112MBUSValidate429(header http.Header) error {
	values := header.Values("Retry-After")
	if len(values) != 1 {
		return fmt.Errorf("Retry-After cardinality %d", len(values))
	}
	parsed, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || parsed < 0 {
		return errors.New("Retry-After is not valid non-negative integer")
	}
	return nil
}
