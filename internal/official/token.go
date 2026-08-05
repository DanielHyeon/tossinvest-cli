package official

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// tokenResponse is the JSON payload returned by POST /oauth2/token.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// cachedToken is persisted to cacheFile (0600).
type cachedToken struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// tokenManager handles OAuth2 client_credentials token acquisition and
// caching. It is safe for concurrent use.
type tokenManager struct {
	creds     Credentials
	base      string
	cacheFile string
	hc        *http.Client

	mu    sync.Mutex
	cache *cachedToken
	// stamp is the cache file's modification time as of the last read or write
	// this manager made. It is how a process notices that another one rotated the
	// shared token: console, engine and the API daemon all run against a single
	// config directory, so the file changes underneath a perfectly valid
	// in-memory copy (change a082).
	stamp time.Time
}

// newTokenManager constructs a tokenManager. base is the API base URL
// (e.g. "https://openapi.tossinvest.com"), cacheFile is the path where the
// token will be persisted on disk, and hc is the HTTP client to use.
func newTokenManager(creds Credentials, base, cacheFile string, hc *http.Client) *tokenManager {
	if hc == nil {
		hc = http.DefaultClient
	}
	return &tokenManager{
		creds:     creds,
		base:      base,
		cacheFile: cacheFile,
		hc:        hc,
	}
}

// token returns a valid access token, reusing the cached one when it has not
// yet expired (using a 60-second skew). If no valid in-memory token exists it
// tries to load from cacheFile before issuing a new exchange.
func (m *tokenManager) token(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// 1. In-memory cache, unless another process has rewritten the shared file
	//    since this manager last read or wrote it. Trusting memory alone means
	//    presenting a token the broker has already replaced, and finding out only
	//    when a request is refused (change a082).
	if isStillValid(m.cache, now) && !m.cacheFileChanged() {
		return m.cache.AccessToken, nil
	}

	// 2. Disk cache.
	if ct, err := m.loadCache(); err == nil && isStillValid(ct, now) {
		m.cache = ct
		m.stampCacheFile()
		return ct.AccessToken, nil
	}

	// 3. Exchange.
	return m.exchange(ctx)
}

// isStillValid reports whether ct is non-nil and not expired, applying a
// 60-second skew so a token is refreshed slightly before it actually expires.
func isStillValid(ct *cachedToken, now time.Time) bool {
	return ct != nil && ct.ExpiresAt.Add(-60*time.Second).After(now)
}

// refresh answers a rejected token: it adopts a newer one from the shared cache
// file when there is one, and otherwise exchanges.
//
// A 401 says the token this process is holding is stale. It does not say a new
// token has to be bought, and the difference matters because the broker keeps one
// token alive per credential. Buying invalidates whatever token another process
// is using, that process is refused in turn and buys its own, and the three
// processes sharing this file spend the day taking the token away from each other
// — measured at seven exchanges a minute for a token that lives a day (a082).
//
// Adopting only when the file holds a *different* token is what keeps this
// narrow. If the file holds the same token that was just refused, or holds
// nothing usable, there is nothing to adopt and a genuinely dead token still gets
// replaced.
func (m *tokenManager) refresh(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ct, err := m.loadCache(); err == nil && isStillValid(ct, time.Now()) &&
		m.cache != nil && ct.AccessToken != m.cache.AccessToken {
		m.cache = ct
		m.stampCacheFile()
		return ct.AccessToken, nil
	}
	return m.exchange(ctx)
}

// cacheFileChanged reports that the cache file has been written since this
// manager last read or wrote it.
//
// A stat is a local filesystem call and this is on every authenticated request,
// which is why it is a stat and not a read: comparing the token itself would mean
// reading and parsing the file every time. A stat failure is reported as changed,
// which costs one extra disk read and never serves a token that has been
// superseded.
//
// Deliberately not a lock. This path carries every broker read the engine makes,
// including the ones the exit loop judges stops on, so making one process wait on
// another here would put that wait inside the interval between stop-loss
// judgements (design D3).
func (m *tokenManager) cacheFileChanged() bool {
	info, err := os.Stat(m.cacheFile)
	if err != nil {
		return true
	}
	return !info.ModTime().Equal(m.stamp)
}

// stampCacheFile records the cache file's modification time as current.
// Must be called with m.mu held.
func (m *tokenManager) stampCacheFile() {
	if info, err := os.Stat(m.cacheFile); err == nil {
		m.stamp = info.ModTime()
		return
	}
	m.stamp = time.Time{}
}

// exchange performs the actual POST /oauth2/token request.
// Must be called with m.mu held.
func (m *tokenManager) exchange(ctx context.Context) (string, error) {
	endpoint := strings.TrimRight(m.base, "/") + "/oauth2/token"

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", m.creds.APIKey)
	form.Set("client_secret", m.creds.SecretKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrTransport, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrTransport, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrTransport, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", classifyStatus(resp.StatusCode, body)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("%w: parsing token response: %s", ErrServer, err)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("%w: empty access_token", ErrServer)
	}

	ct := &cachedToken{
		AccessToken: tr.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
	}
	m.cache = ct
	_ = m.saveCache(ct) // best-effort; do not fail the call if disk write fails
	m.stampCacheFile()
	return ct.AccessToken, nil
}

// loadCache reads and unmarshals the cache file. Returns (nil, nil) if the
// file does not exist.
func (m *tokenManager) loadCache() (*cachedToken, error) {
	data, err := os.ReadFile(m.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		// Other read errors (e.g. permission, corruption) are returned but
		// treated as best-effort by the caller: a failed cache read should
		// never block token acquisition, it just forces a fresh exchange.
		return nil, err
	}
	var ct cachedToken
	if err := json.Unmarshal(data, &ct); err != nil {
		return nil, err
	}
	return &ct, nil
}

// saveCache persists ct to cacheFile with 0600 permissions.
func (m *tokenManager) saveCache(ct *cachedToken) error {
	dir := filepath.Dir(m.cacheFile)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(ct)
	if err != nil {
		return err
	}
	return os.WriteFile(m.cacheFile, data, 0o600)
}
