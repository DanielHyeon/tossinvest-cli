// Command a112-mb-us-source collects one bounded, read-only A112 US-source
// measurement receipt. It is intentionally not installed into tossctl.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	requestBudget = 15 * time.Second
	overallBudget = 120 * time.Second
)

// config is deliberately private: the only production entry point is main.
type config struct {
	tokenCache  string
	goBinary    string
	receiptRoot string
	sessionDate string
	before      []byte
}

// sourceResult is an opaque copy of a M-B0 response. The tool never decodes
// market values; it records the raw response for an independent reviewer.
type sourceResult struct {
	method      string
	path        string
	query       string
	statusCode  int
	body        []byte
	cursorJSON  []byte
	cursorValue []byte
	rateHeaders map[string][]string
}

type sourceReader interface {
	Candle(context.Context, []byte) (sourceResult, error)
	Orderbook(context.Context) (sourceResult, error)
	Calendar(context.Context, string) (sourceResult, error)
}

type dependencies struct {
	reader      sourceReader
	now         func() time.Time
	identity    identityProvider
	builder     buildRunner
	openReceipt func(string) (receiptStore, error)
}

// receiptStore is a capability rooted in an already verified directory FD.
// No method accepts an arbitrary destination path after preflight.
type receiptStore interface {
	writeJSON(string, any) error
	writeResult(string, sourceResult) error
	seal(receiptSeal) error
	verifySealed() error
	taint() error
	close() error
}

// run is kept separate from main so tests can use an in-memory reader and
// prove that every rejection performs zero network calls.
func run(ctx context.Context, cfg config, deps dependencies) (err error) {
	if err := validateConfig(cfg); err != nil {
		return holdf("invalid collector input: %v", err)
	}
	if ctx == nil || deps.reader == nil || deps.now == nil || deps.identity == nil || deps.builder == nil {
		return holdf("collector dependencies are incomplete")
	}
	started := deps.now()
	if started.IsZero() {
		return holdf("clock returned zero time")
	}
	collectorCtx, cancelCollector := context.WithDeadline(ctx, time.Now().Add(overallBudget))
	defer cancelCollector()
	live := func() error { return collectorLive(collectorCtx, deps.now, started) }
	if err := live(); err != nil {
		return holdf("collector unavailable: %v", err)
	}
	opener := deps.openReceipt
	if opener == nil {
		opener = openReceipt
	}
	store, err := opener(cfg.receiptRoot)
	if err != nil {
		return holdf("receipt root: %v", err)
	}
	defer func() { _ = store.close() }()
	defer func() {
		if err != nil {
			_ = store.taint()
		}
	}()
	if guarded, ok := store.(interface{ setLive(func() error) }); ok {
		guarded.setLive(live)
	}

	// Both the receipt capability and all pre-request identity inputs are
	// checked before a reader can make its first GET.
	if err := live(); err != nil {
		return holdf("before pre-request identity: %v", err)
	}
	pre, err := snapshotIdentity(collectorCtx, deps.identity, cfg)
	if err != nil {
		return holdf("pre-request identity: %v", err)
	}
	defer func() { _ = pre.close() }()
	if err := verifyPrescribedBinary(collectorCtx, pre, deps.builder); err != nil {
		return holdf("prescribed build identity: %v", err)
	}
	if err := live(); err != nil {
		return holdf("after pre-request identity: %v", err)
	}
	if err := store.writeJSON("preflight.json", receiptPreflight{Config: receiptConfigFrom(cfg), Identity: pre}); err != nil {
		return holdf("write preflight sentinel: %v", err)
	}
	if err := live(); err != nil {
		return holdf("after preflight sentinel: %v", err)
	}
	seen := map[string]struct{}{}
	before := bytes.Clone(cfg.before)
	if len(before) != 0 {
		seen[string(before)] = struct{}{}
	}
	terminal := false
	pages := 0
	for page := 0; page < 4; page++ {
		requestCtx, cancel, requestErr := boundedRequestContext(collectorCtx, deps.now, started)
		if requestErr != nil {
			return holdf("candle page %d: %v", page+1, requestErr)
		}
		result, readErr := deps.reader.Candle(requestCtx, before)
		cancel()
		if readErr != nil {
			return holdf("candle page %d: %v", page+1, readErr)
		}
		if err := live(); err != nil {
			return holdf("candle page %d unavailable: %v", page+1, err)
		}
		if err := validateSourceResult(result); err != nil {
			return holdf("candle page %d result: %v", page+1, err)
		}
		if err := store.writeResult(fmt.Sprintf("candle-%02d", page+1), result); err != nil {
			return holdf("candle page %d receipt: %v", page+1, err)
		}
		if err := live(); err != nil {
			return holdf("after candle page %d receipt: %v", page+1, err)
		}
		pages = page + 1
		cursor, done, err := nextCursor(result, seen)
		if err != nil {
			return holdf("candle page %d cursor: %v", page+1, err)
		}
		if done {
			terminal = true
			break
		}
		if err := requireRemaining(result.rateHeaders); err != nil {
			return holdf("candle page %d rate budget: %v", page+1, err)
		}
		seen[string(cursor)] = struct{}{}
		before = cursor
	}
	// 네 페이지를 다 써도 그것만으로는 HOLD가 아니다. 다섯 번째 캔들 요청은 하지 않고,
	// 어떻게 끝났는지(raw null인지 cap 소진인지)를 receipt에 그대로 적고 계속 간다.
	crawl := receiptCandleCrawl{Schema: candleCrawlSchema, Pages: pages, Terminal: candleTerminalNull}
	if !terminal {
		crawl.Terminal = candleTerminalCapExhausted
		crawl.LastCursorSHA256 = digestBytes(before)
	}
	if err := store.writeJSON("candle-crawl.json", crawl); err != nil {
		return holdf("candle crawl receipt: %v", err)
	}
	if err := live(); err != nil {
		return holdf("after candle crawl receipt: %v", err)
	}
	if err := collectSingle(collectorCtx, deps, started, "orderbook", store, func(requestCtx context.Context) (sourceResult, error) {
		return deps.reader.Orderbook(requestCtx)
	}); err != nil {
		return err
	}
	if err := collectSingle(collectorCtx, deps, started, "calendar", store, func(requestCtx context.Context) (sourceResult, error) {
		return deps.reader.Calendar(requestCtx, cfg.sessionDate)
	}); err != nil {
		return err
	}

	if err := live(); err != nil {
		return holdf("before post-request identity: %v", err)
	}
	post, err := snapshotIdentity(collectorCtx, deps.identity, cfg)
	if err != nil {
		return holdf("post-request identity: %v", err)
	}
	defer func() { _ = post.close() }()
	if !pre.equal(post) {
		return holdf("identity drift during measurement")
	}
	if err := live(); err != nil {
		return holdf("after post-request identity: %v", err)
	}
	if err := store.seal(receiptSeal{Identity: post}); err != nil {
		return holdf("seal receipt: %v", err)
	}
	if err := live(); err != nil {
		return holdf("after sealing: %v", err)
	}
	if err := store.verifySealed(); err != nil {
		return holdf("sealed receipt verification: %v", err)
	}
	if err := live(); err != nil {
		return holdf("before success: %v", err)
	}
	return nil
}

func snapshotIdentity(ctx context.Context, identity identityProvider, cfg config) (identitySnapshot, error) {
	if contextual, ok := identity.(contextualIdentityProvider); ok {
		return contextual.snapshotContext(ctx, cfg)
	}
	return identity.snapshot(cfg)
}

func collectSingle(ctx context.Context, deps dependencies, started time.Time, name string, store receiptStore, read func(context.Context) (sourceResult, error)) error {
	requestCtx, cancel, err := boundedRequestContext(ctx, deps.now, started)
	if err != nil {
		return holdf("%s: %v", name, err)
	}
	result, readErr := read(requestCtx)
	cancel()
	if readErr != nil {
		return holdf("%s: %v", name, readErr)
	}
	if err := collectorLive(ctx, deps.now, started); err != nil {
		return holdf("%s unavailable: %v", name, err)
	}
	if err := validateSourceResult(result); err != nil {
		return holdf("%s result: %v", name, err)
	}
	if err := store.writeResult(name, result); err != nil {
		return holdf("%s receipt: %v", name, err)
	}
	if err := collectorLive(ctx, deps.now, started); err != nil {
		return holdf("after %s receipt: %v", name, err)
	}
	return nil
}

func validateConfig(cfg config) error {
	if cfg.tokenCache == "" || cfg.goBinary == "" || cfg.receiptRoot == "" {
		return errors.New("--token-cache, --go-binary and --receipt-root are required")
	}
	if !isAbsolutePath(cfg.tokenCache) || !isAbsolutePath(cfg.goBinary) || !isAbsolutePath(cfg.receiptRoot) {
		return errors.New("token cache, go binary and receipt root must be absolute paths")
	}
	if len(cfg.sessionDate) != len("2006-01-02") {
		return errors.New("session date must be exact YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", cfg.sessionDate); err != nil {
		return errors.New("session date must be exact YYYY-MM-DD")
	}
	if !utf8.Valid(cfg.before) {
		return errors.New("before must be valid UTF-8")
	}
	return nil
}

func collectorLive(ctx context.Context, now func() time.Time, started time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	current := now()
	if current.Before(started) || current.Sub(started) >= overallBudget {
		return errors.New("absolute 120-second collector deadline expired")
	}
	return nil
}

func isAbsolutePath(path string) bool { return len(path) > 0 && path[0] == '/' }

func boundedRequestContext(parent context.Context, now func() time.Time, started time.Time) (context.Context, context.CancelFunc, error) {
	if err := parent.Err(); err != nil {
		return nil, nil, err
	}
	current := now()
	if current.Before(started) || current.Sub(started) > overallBudget || overallBudget-current.Sub(started) < requestBudget {
		return nil, nil, errors.New("less than 15 seconds remain in the 120-second collector budget")
	}
	requestCtx, cancel := context.WithTimeout(parent, requestBudget)
	return requestCtx, cancel, nil
}

func nextCursor(result sourceResult, seen map[string]struct{}) ([]byte, bool, error) {
	if bytes.Equal(result.cursorJSON, []byte("null")) {
		if len(result.cursorValue) != 0 {
			return nil, false, errors.New("raw null cursor has decoded value")
		}
		return nil, true, nil
	}
	if len(result.cursorJSON) == 0 || len(result.cursorValue) == 0 || !utf8.Valid(result.cursorValue) {
		return nil, false, errors.New("cursor must be raw null or a non-empty UTF-8 value")
	}
	var decoded string
	if err := json.Unmarshal(result.cursorJSON, &decoded); err != nil || decoded == "" || !bytes.Equal([]byte(decoded), result.cursorValue) {
		return nil, false, errors.New("raw cursor JSON is not the exact decoded cursor value")
	}
	if _, duplicate := seen[string(result.cursorValue)]; duplicate {
		return nil, false, errors.New("cursor loop")
	}
	return bytes.Clone(result.cursorValue), false, nil
}

func requireRemaining(headers map[string][]string) error {
	if headers == nil {
		return errors.New("rate headers absent")
	}
	values := headers["X-Ratelimit-Remaining"]
	if len(values) != 1 {
		return errors.New("remaining header must have exactly one value")
	}
	remaining, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || remaining < 1 {
		return errors.New("remaining header is invalid or below one")
	}
	return nil
}

func validateSourceResult(result sourceResult) error {
	if len(result.body) == 0 {
		return errors.New("raw body is absent")
	}
	if result.statusCode != 0 && (result.statusCode < 200 || result.statusCode >= 300) {
		return errors.New("response status is not successful")
	}
	for _, name := range []string{"X-Ratelimit-Limit", "X-Ratelimit-Remaining", "X-Ratelimit-Reset"} {
		values := result.rateHeaders[name]
		if len(values) != 1 {
			return fmt.Errorf("%s must have exactly one value", name)
		}
		parsed, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || parsed < 0 || (name == "X-Ratelimit-Limit" && parsed == 0) {
			return fmt.Errorf("%s must be a valid non-negative integer", name)
		}
	}
	if values := result.rateHeaders["Retry-After"]; len(values) != 0 {
		return errors.New("Retry-After must be absent on success")
	}
	return nil
}

type holdError struct{ reason string }

func (e *holdError) Error() string           { return "a112 m-b1 measurement HOLD: " + e.reason }
func holdf(format string, args ...any) error { return &holdError{reason: fmt.Sprintf(format, args...)} }

type receiptConfig struct {
	SessionDate  string `json:"session_date"`
	BeforeSHA256 string `json:"before_sha256"`
}

func receiptConfigFrom(cfg config) receiptConfig {
	return receiptConfig{SessionDate: cfg.sessionDate, BeforeSHA256: digestBytes(cfg.before)}
}

type receiptPreflight struct {
	Config   receiptConfig    `json:"config"`
	Identity identitySnapshot `json:"identity"`
}

type receiptSeal struct {
	Identity identitySnapshot `json:"identity"`
}

const (
	candleCrawlSchema          = "a112-mb-us-candle-crawl:v1"
	candleTerminalNull         = "null"
	candleTerminalCapExhausted = "cap_exhausted"
)

// receiptCandleCrawl은 캔들 crawl이 어떻게 끝났는지만 적는다. cursor 원문은 이미
// candle-NN.cursor.raw.json에 있으므로 여기에는 해시만 남기고 시각은 남기지 않는다.
type receiptCandleCrawl struct {
	Schema           string `json:"schema"`
	Pages            int    `json:"pages"`
	Terminal         string `json:"terminal"`
	LastCursorSHA256 string `json:"last_cursor_sha256,omitempty"`
}

func marshalReceiptMeta(result sourceResult) ([]byte, error) {
	meta := struct {
		Method            string            `json:"method"`
		Path              string            `json:"path"`
		CanonicalQuery    string            `json:"canonical_query"`
		StatusCode        int               `json:"status_code"`
		BodySHA256        string            `json:"body_sha256"`
		CursorJSONSHA256  string            `json:"cursor_json_sha256,omitempty"`
		CursorValueSHA256 string            `json:"cursor_value_sha256,omitempty"`
		RateHeaderSHA256  map[string]string `json:"rate_header_sha256"`
	}{
		Method: result.method, Path: result.path, CanonicalQuery: result.query, StatusCode: result.statusCode,
		BodySHA256: digestBytes(result.body), CursorJSONSHA256: digestBytes(result.cursorJSON),
		CursorValueSHA256: digestBytes(result.cursorValue), RateHeaderSHA256: headerDigests(result.rateHeaders),
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(meta); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(encoded.Bytes(), []byte{'\n'}), nil
}

func headerDigests(input map[string][]string) map[string]string {
	out := make(map[string]string, len(input))
	for key, values := range input {
		out[key] = digestBytes([]byte(strings.Join(values, "\x00")))
	}
	return out
}
