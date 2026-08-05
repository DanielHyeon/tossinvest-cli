package official

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const defaultBaseURL = "https://openapi.tossinvest.com"
const defaultTimeout = 15 * time.Second

// Client is the official Toss Open API client.
// It manages OAuth2 token acquisition/refresh and provides authed HTTP helpers.
type Client struct {
	base string
	hc   *http.Client
	tm   *tokenManager
	// configMu protects construction-time options. New seals configuration before
	// returning so a custom Option that retained *Client cannot replay a standard
	// option between authority verification and the authoritative HTTP read.
	configMu            sync.RWMutex
	configurationSealed bool
	// authorityOrigin remains true only for the constructor-owned production
	// endpoint and HTTP transport. Options may customize ordinary API reads, but
	// an overridden transport can never attest official monetary authority.
	authorityOrigin    bool
	authorityTransport *http.Transport
	// mu serializes unresolved account discovery and validates later public
	// discovery against an implicit selection. accountsLocked requires it and
	// may perform the /accounts HTTP request while held.
	mu sync.Mutex
	// accountSeq is read atomically so an already selected account-scoped
	// request never waits behind unrelated public account-list I/O.
	accountSeq         atomic.Int64 // 0 = unresolved; positive = selected; negative = invalid
	accountSeqExplicit bool         // immutable after New returns
	// rates is the last rate-limit budget seen per request path (ratebudget.go).
	// Read-only from a caller's point of view and never consulted by this
	// package's own logic — it records, it does not decide.
	rates rateBudgets
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL (used in tests with httptest).
func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.configMu.Lock()
		defer c.configMu.Unlock()
		if c.configurationSealed {
			return
		}
		c.base = strings.TrimRight(u, "/")
		c.authorityOrigin = false
	}
}

// WithHTTPClient overrides the HTTP client (used in tests to share httptest transport).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.configMu.Lock()
		defer c.configMu.Unlock()
		if c.configurationSealed {
			return
		}
		c.hc = hc
		c.authorityOrigin = false
	}
}

// WithAccountSeq configures the account sequence used by account-scoped
// endpoints. A positive value is an explicit selection, zero remains unresolved
// and is discovered lazily, and a negative value is rejected before discovery
// or header emission.
func WithAccountSeq(seq int) Option {
	return func(c *Client) {
		c.configMu.Lock()
		defer c.configMu.Unlock()
		if c.configurationSealed {
			return
		}
		c.accountSeq.Store(int64(seq))
		c.accountSeqExplicit = seq > 0
	}
}

// New constructs a Client. cacheFile is the path for the on-disk token cache.
// Options are applied after defaults, so WithBaseURL/WithHTTPClient override them.
func New(creds Credentials, cacheFile string, opts ...Option) *Client {
	hc := newOfficialHTTPClient()
	transport := hc.Transport.(*http.Transport)
	c := &Client{
		base:               defaultBaseURL,
		hc:                 hc,
		authorityOrigin:    true,
		authorityTransport: transport,
	}
	for _, o := range opts {
		o(c)
	}
	// Seal the endpoint, transport and account option state in the same critical
	// section that binds token acquisition to those values. Replayed public Option
	// closures become harmless no-ops after this point.
	c.configMu.Lock()
	c.tm = newTokenManager(creds, c.base, cacheFile, c.hc)
	c.configurationSealed = true
	c.configMu.Unlock()
	return c
}

func newOfficialHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &http.Client{Timeout: defaultTimeout, Transport: transport}
}

// AuthorityOrigin is an opaque proof that a Client retained the constructor-
// owned production endpoint and transport. Its fields are private so callers
// cannot promote a configured test/proxy client into monetary authority.
type AuthorityOrigin struct {
	production bool
}

func (o AuthorityOrigin) Valid() bool { return o.production }

// AuthorityOrigin returns no capability after any endpoint or HTTP-client
// override, even when the configured value happens to resemble the default.
func (c *Client) AuthorityOrigin() (AuthorityOrigin, bool) {
	if c == nil {
		return AuthorityOrigin{}, false
	}
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	if !c.authorityOriginLocked() {
		return AuthorityOrigin{}, false
	}
	return AuthorityOrigin{production: true}, true
}

func (c *Client) authorityOriginLocked() bool {
	transport, transportOK := c.hcTransport()
	return c.configurationSealed && c.authorityOrigin && c.base == defaultBaseURL &&
		transportOK && c.authorityTransport != nil && transport == c.authorityTransport
}

func (c *Client) hcTransport() (*http.Transport, bool) {
	if c.hc == nil {
		return nil, false
	}
	transport, ok := c.hc.Transport.(*http.Transport)
	return transport, ok
}

// BaseURL returns the base URL this client targets.
func (c *Client) BaseURL() string {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	return c.base
}

// apiEnvelope is the common response wrapper: {"result": <payload>}.
type apiEnvelope struct {
	Result json.RawMessage `json:"result"`
}

// doRequest executes req, returning (statusCode, body, error).
//
// It is also where the rate-limit headers are read (ratebudget.go). This is the
// only place they exist: the callers above get (status, body) and the status is
// mapped onto a sentinel error, so by the time anyone could ask about the budget
// the response is gone. Recording here is purely additive — the return values and
// every error are unchanged — and it covers the 429 case too, which is the single
// most informative response and the one the error mapping used to discard.
func (c *Client) doRequest(req *http.Request) (int, []byte, error) {
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %s", ErrTransport, err)
	}
	defer resp.Body.Close()
	c.rates.record(readRateBudget(req.URL.Path, resp.Header, time.Now()))
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("%w: reading body: %s", ErrTransport, err)
	}
	return resp.StatusCode, body, nil
}

// unwrapAndDecode extracts `result` from the response envelope and unmarshals
// it into out. If out is nil the body is discarded. Responses are expected to
// have the shape {"result": <payload>}; if the `result` key is absent and out
// is non-nil an error is returned.
func unwrapAndDecode(body []byte, out any) error {
	if out == nil {
		return nil
	}
	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("%w: decoding envelope: %s", ErrServer, err)
	}
	if env.Result == nil {
		return fmt.Errorf("%w: response has no 'result' key", ErrServer)
	}
	if err := json.Unmarshal(env.Result, out); err != nil {
		return fmt.Errorf("%w: decoding result payload: %s", ErrServer, err)
	}
	return nil
}

// get performs an authenticated GET request to path (relative to BaseURL).
// Query parameters q may be nil. On 401 the token is refreshed and the request
// is retried once. On non-2xx classifyStatus is returned. On 2xx the `result`
// envelope is unwrapped into out (out may be nil).
func (c *Client) get(ctx context.Context, path string, q url.Values, out any) error {
	return c.getWithHeaders(ctx, path, q, nil, out)
}

// ensureAccountSeq returns a positive selected sequence. Zero triggers the
// shared account-discovery path; a successful positive first record is cached
// and shared by concurrent scoped callers and by a public Accounts request
// already in flight. Failed, empty, or invalid discovery remains unresolved so
// a later caller may retry. A positive explicit value skips discovery, while a
// negative explicit value is rejected before any request or header emission.
func (c *Client) ensureAccountSeq(ctx context.Context) (int, error) {
	if seq := c.accountSeq.Load(); seq != 0 {
		if seq > 0 {
			return int(seq), nil
		}
		return 0, fmt.Errorf("lazy account-seq resolution: configured account sequence %d is not positive",
			seq)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if seq := c.accountSeq.Load(); seq != 0 {
		if seq > 0 {
			return int(seq), nil
		}
		return 0, fmt.Errorf("lazy account-seq resolution: configured account sequence %d is not positive",
			seq)
	}
	accts, err := c.accountsLocked(ctx)
	if err != nil {
		return 0, fmt.Errorf("lazy account-seq resolution: %w", err)
	}
	if len(accts) == 0 {
		return 0, fmt.Errorf("lazy account-seq resolution: no accounts found")
	}
	seq, err := strconv.Atoi(accts[0].ID)
	if err != nil {
		return 0, fmt.Errorf("lazy account-seq resolution: parsing account ID %q: %w", accts[0].ID, err)
	}
	if seq <= 0 {
		return 0, fmt.Errorf("lazy account-seq resolution: account ID %q is not positive", accts[0].ID)
	}
	selected := c.accountSeq.Load()
	if selected != int64(seq) {
		return 0, fmt.Errorf(
			"lazy account-seq resolution: discovered account sequence %d but selected sequence is %d",
			seq, selected,
		)
	}
	return seq, nil
}

// getAcct is like get but also resolves the account sequence (lazily if needed)
// and sets the X-Tossinvest-Account header. Used by account-scoped endpoints
// such as BuyingPower, Holdings, and Orders that require the header per the
// official API spec.
func (c *Client) getAcct(ctx context.Context, path string, q url.Values, out any) error {
	seq, err := c.ensureAccountSeq(ctx)
	if err != nil {
		return err
	}
	extra := map[string]string{"X-Tossinvest-Account": strconv.Itoa(seq)}
	return c.getWithHeaders(ctx, path, q, extra, out)
}

// postAcct is like post but injects the X-Tossinvest-Account header (lazy seq).
func (c *Client) postAcct(ctx context.Context, path string, body any, out any) error {
	seq, err := c.ensureAccountSeq(ctx)
	if err != nil {
		return err
	}
	extra := map[string]string{"X-Tossinvest-Account": strconv.Itoa(seq)}
	return c.postWithHeaders(ctx, path, body, extra, out)
}

// deleteAcct performs an authenticated DELETE with the account header, mirroring
// send runs the authenticated-request flow every verb shares: acquire a token,
// build the request, retry ONCE with a refreshed token on 401, classify non-2xx,
// and unwrap the `result` envelope into out (out may be nil).
//
// Callers supply makeReq because that is the only part that differs between
// verbs — method, body, query, and per-request headers. Keeping the retry here
// means the auth policy is decided in one place; it used to be hand-repeated in
// getWithHeaders, postWithHeaders and deleteAcct, where it could silently drift.
//
// makeReq must be callable twice: the retry rebuilds the request so the new
// token is applied (and so a consumed body reader is not reused).
func (c *Client) send(ctx context.Context, makeReq func(tok string) (*http.Request, error), out any) error {
	tok, err := c.tm.token(ctx)
	if err != nil {
		return err
	}
	req, err := makeReq(tok)
	if err != nil {
		return err
	}
	code, body, err := c.doRequest(req)
	if err != nil {
		return err
	}
	// Answer a 401 with a fresh token and retry. refresh may satisfy the first
	// attempt by taking a token another holder already obtained rather than buying
	// one — that is what stops the holders sharing this cache file from
	// invalidating each other (change a082) — but an adopted token is only
	// inferred to be live, from its expiry. If the broker refuses that one too, the
	// retry has been spent on a guess, and the loop below buys a token and tries
	// again so the request still ends on a minted one.
	//
	// The bound is two refreshes, and the second cannot adopt: refresh never
	// returns the token it was refused on, so the guess is excluded from its own
	// replacement and the second pass exchanges.
	for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++ {
		var adopted bool
		tok, adopted, err = c.tm.refresh(ctx, tok)
		if err != nil {
			return err
		}
		req, err = makeReq(tok)
		if err != nil {
			return err
		}
		code, body, err = c.doRequest(req)
		if err != nil {
			return err
		}
		if !adopted {
			break
		}
	}
	if code < 200 || code >= 300 {
		return classifyStatus(code, body)
	}
	return unwrapAndDecode(body, out)
}

// getWithHeaders' token/401-retry/unwrap flow.
func (c *Client) deleteAcct(ctx context.Context, path string, out any) error {
	seq, err := c.ensureAccountSeq(ctx)
	if err != nil {
		return err
	}
	rawURL := c.base + path
	makeReq := func(tok string) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrTransport, err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("X-Tossinvest-Account", strconv.Itoa(seq))
		return req, nil
	}
	return c.send(ctx, makeReq, out)
}

// getWithHeaders performs an authenticated GET request to path (relative to
// BaseURL), injecting any extraHeaders on top of the Authorization header.
// Query parameters q may be nil. On 401 the token is refreshed and the request
// is retried once. On non-2xx classifyStatus is returned. On 2xx the `result`
// envelope is unwrapped into out (out may be nil).
func (c *Client) getWithHeaders(ctx context.Context, path string, q url.Values, extraHeaders map[string]string, out any) error {
	rawURL := c.base + path
	if len(q) > 0 {
		rawURL += "?" + q.Encode()
	}

	makeReq := func(tok string) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrTransport, err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}
		return req, nil
	}

	return c.send(ctx, makeReq, out)
}

// post performs an authenticated POST request to path (relative to BaseURL).
// body is JSON-encoded and sent with Content-Type: application/json. On 401
// the token is refreshed and the request is retried once. On non-2xx
// classifyStatus is returned. On 2xx the `result` envelope is unwrapped into
// out (out may be nil).
func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	return c.postWithHeaders(ctx, path, body, nil, out)
}

// postWithHeaders is like post but also injects extraHeaders (e.g.
// X-Tossinvest-Account) on top of the Authorization header. extraHeaders may
// be nil. On 401 the token is refreshed and the request is retried once. On
// non-2xx classifyStatus is returned. On 2xx the `result` envelope is
// unwrapped into out (out may be nil).
func (c *Client) postWithHeaders(ctx context.Context, path string, body any, extraHeaders map[string]string, out any) error {
	rawURL := c.base + path

	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("%w: marshalling request body: %s", ErrTransport, err)
	}

	makeReq := func(tok string) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(encoded))
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrTransport, err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}
		return req, nil
	}

	return c.send(ctx, makeReq, out)
}
