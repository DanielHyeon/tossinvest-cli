package official

// token_shared_cache_test.go covers change a082: the console, the engine and the
// API daemon share one token cache file, and before this change they spent their
// time invalidating each other's tokens.
//
// The measured symptom was a 24-hour token being re-exchanged seven times a
// minute in the running system. saveCache has exactly one caller — exchange —
// and a 24-hour token cannot expire that often, so every one of those writes was
// a forced refresh off a 401.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// lastTokenWinsServer is a broker that keeps exactly one token alive, which is
// what client_credentials grants normally do and what the production evidence
// implies: if superseded tokens stayed valid there would be no 401s, and with no
// 401s there is no path to exchange, and then the cache file could not be
// rewritten seven times a minute.
type lastTokenWinsServer struct {
	mu        sync.Mutex
	issued    int
	live      string
	rejectNth int // the request number to refuse regardless, kicking the loop off
	requests  int
}

func (s *lastTokenWinsServer) exchanges() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.issued
}

func (s *lastTokenWinsServer) handler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.URL.Path {
		case "/oauth2/token":
			s.issued++
			s.live = fmt.Sprintf("T%d", s.issued)
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"expires_in":86400,"token_type":"Bearer"}`, s.live)
		case "/api/v1/ping":
			s.requests++
			presented := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if s.requests == s.rejectNth || presented != s.live {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"result":{"ok":true}}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
}

// twoProcesses builds two clients over one cache file, which is what the
// container actually runs: console and engine in one image, httpapi in another,
// all three bound to the same config directory.
func twoProcesses(t *testing.T, srv *httptest.Server) (*Client, *Client, string) {
	t.Helper()
	cache := filepath.Join(t.TempDir(), "openapi-token.json")
	opts := []Option{WithBaseURL(srv.URL), WithHTTPClient(srv.Client())}
	creds := Credentials{APIKey: "k", SecretKey: "s"}
	return New(creds, cache, opts...), New(creds, cache, opts...), cache
}

func ping(t *testing.T, c *Client) error {
	t.Helper()
	var out struct {
		OK bool `json:"ok"`
	}
	return c.get(context.Background(), "/api/v1/ping", nil, &out)
}

// TestTwoProcessesSharingOneCacheFileStopBuyingTokensFromEachOther is the
// headline. One refusal is enough to start the loop: the refused process
// exchanges, which kills the token the other process is holding in memory, which
// makes that one exchange, and so on. Nothing in the old code ever let a process
// notice that a perfectly good token was already sitting in the shared file.
func TestTwoProcessesSharingOneCacheFileStopBuyingTokensFromEachOther(t *testing.T) {
	broker := &lastTokenWinsServer{rejectNth: 3}
	srv := httptest.NewServer(broker.handler(t))
	defer srv.Close()
	a, b, _ := twoProcesses(t, srv)

	const rounds = 12
	for i := 0; i < rounds; i++ {
		if err := ping(t, a); err != nil {
			t.Fatalf("round %d, process A: %v", i, err)
		}
		if err := ping(t, b); err != nil {
			t.Fatalf("round %d, process B: %v", i, err)
		}
	}

	// The token lives 24 hours and nothing here waits that long, so a converged
	// pair buys one token, plus one more for the single refusal that starts the
	// loop. Anything above that is the two of them taking turns.
	const want = 3
	if got := broker.exchanges(); got > want {
		t.Errorf("%d rounds bought %d tokens, want at most %d — the two processes are "+
			"invalidating each other instead of sharing the cache file", rounds, got, want)
	}
	if rounds*2 <= want {
		t.Fatalf("%d requests against a budget of %d tokens cannot tell a converged "+
			"pair from a ping-ponging one", rounds*2, want)
	}
}

// TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot states the rule on
// its own: a 401 says the token being held is stale, not that a new one must be
// bought. If the shared file already holds a different valid token, that is the
// answer, and buying another would invalidate whoever is using it.
func TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot(t *testing.T) {
	broker := &lastTokenWinsServer{}
	srv := httptest.NewServer(broker.handler(t))
	defer srv.Close()
	a, b, cache := twoProcesses(t, srv)

	if err := ping(t, a); err != nil { // A buys T1 and writes it to the file
		t.Fatal(err)
	}
	if err := ping(t, b); err != nil { // B reads T1 off the file
		t.Fatal(err)
	}
	// A now holds a token that the broker has forgotten, exactly as it would after
	// any other process rotated it.
	writeCachedToken(t, cache, "T-stale", time.Now().Add(24*time.Hour))
	a.tm.mu.Lock()
	a.tm.cache = &cachedToken{AccessToken: "T-stale", ExpiresAt: time.Now().Add(24 * time.Hour)}
	a.tm.mu.Unlock()
	writeCachedToken(t, cache, "T1", time.Now().Add(24*time.Hour))

	before := broker.exchanges()
	if err := ping(t, a); err != nil {
		t.Fatalf("the refused process could not recover: %v", err)
	}
	if got := broker.exchanges(); got != before {
		t.Errorf("the refused process bought %d token(s); the shared file already held "+
			"a valid one and buying kills the process using it", got-before)
	}
}

// TestARotationThatLandsMidRequestCostsNoToken is the case the file-changed
// check cannot reach.
//
// That check runs when a token is handed out. It cannot see a rotation that
// happens after the request is built and before the broker answers it — and that
// in-flight window is exactly where the ten production refusals landed, because
// it is the only way a refusal survives the retry at all. Buying a token here
// would kill the one the rotating process just started using, which is how the
// loop restarts after the file-changed check has already broken it once.
func TestARotationThatLandsMidRequestCostsNoToken(t *testing.T) {
	broker := &lastTokenWinsServer{}
	var cache string
	rotate := func() {
		// Stand in for the other process: it exchanged and wrote the shared file
		// while this request was on the wire. issued counts what the client under
		// test buys, so the stand-in must not touch it.
		broker.live = "T-from-the-other-process"
		writeCachedToken(t, cache, broker.live, time.Now().Add(24*time.Hour))
	}
	rotated := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		broker.mu.Lock()
		defer broker.mu.Unlock()
		switch r.URL.Path {
		case "/oauth2/token":
			broker.issued++
			broker.live = fmt.Sprintf("T%d", broker.issued)
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"expires_in":86400,"token_type":"Bearer"}`, broker.live)
		case "/api/v1/ping":
			presented := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !rotated && presented == broker.live {
				rotated = true
				rotate()
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if presented != broker.live {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"result":{"ok":true}}`))
		}
	}))
	defer srv.Close()

	cache = filepath.Join(t.TempDir(), "openapi-token.json")
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, cache,
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))

	if err := ping(t, c); err != nil { // buys T1, then the rotation refuses it
		t.Fatalf("the request did not recover from a mid-flight rotation: %v", err)
	}
	if !rotated {
		t.Fatal("the rotation never fired, so this test proves nothing")
	}
	// One exchange for the first token. The refusal must be answered by taking the
	// token the rotation left in the file, not by buying another.
	if got := broker.exchanges(); got != 1 {
		t.Errorf("a mid-flight rotation cost %d exchanges, want 1 — the refused process "+
			"bought a token while a valid one was sitting in the shared file, and that "+
			"purchase is what invalidates the process that just rotated", got)
	}
}

// TestARefusedProcessWithNothingToAdoptStillExchanges is the other half. Adopting
// is only right when there is something newer to adopt; a genuinely dead token
// must still be replaced, or a refusal becomes a permanent refusal.
func TestARefusedProcessWithNothingToAdoptStillExchanges(t *testing.T) {
	broker := &lastTokenWinsServer{}
	srv := httptest.NewServer(broker.handler(t))
	defer srv.Close()
	a, _, cache := twoProcesses(t, srv)

	if err := ping(t, a); err != nil {
		t.Fatal(err)
	}
	// Both the memory and the file hold the same token, and the broker has moved
	// on. There is nothing to adopt.
	writeCachedToken(t, cache, "T-dead", time.Now().Add(24*time.Hour))
	a.tm.mu.Lock()
	a.tm.cache = &cachedToken{AccessToken: "T-dead", ExpiresAt: time.Now().Add(24 * time.Hour)}
	a.tm.mu.Unlock()

	before := broker.exchanges()
	if err := ping(t, a); err != nil {
		t.Fatalf("a process with nothing to adopt did not recover: %v", err)
	}
	if got := broker.exchanges(); got != before+1 {
		t.Errorf("bought %d token(s), want exactly 1 — with nothing to adopt the only "+
			"way out of a refusal is a new token", got-before)
	}
}

// TestTokenPrefersTheCacheFileWhenAnotherProcessRewroteIt removes the wasted
// round trip. Adoption alone converges, but only after eating one refusal per
// rotation, and every refusal reopens the retry window this change exists to
// close.
func TestTokenPrefersTheCacheFileWhenAnotherProcessRewroteIt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":86400,"token_type":"Bearer"}`))
	}))
	defer srv.Close()
	cache := filepath.Join(t.TempDir(), "openapi-token.json")
	m := newTokenManager(Credentials{APIKey: "k", SecretKey: "s"}, srv.URL, cache, srv.Client())

	first, err := m.token(context.Background())
	if err != nil || first != "AT" {
		t.Fatalf("got %q, %v", first, err)
	}

	// Another process rotates the shared file. mtime resolution is coarse on some
	// filesystems, so move the timestamp explicitly rather than racing it.
	writeCachedToken(t, cache, "AT-from-another-process", time.Now().Add(24*time.Hour))
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(cache, future, future); err != nil {
		t.Fatal(err)
	}

	got, err := m.token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != "AT-from-another-process" {
		t.Errorf("token() returned %q; the cache file changed underneath it and it kept "+
			"serving its own copy until a refusal forced the issue", got)
	}
}

// writeCachedToken puts a token in the shared file the way another process would.
func writeCachedToken(t *testing.T, path, token string, expires time.Time) {
	t.Helper()
	data, err := json.Marshal(cachedToken{AccessToken: token, ExpiresAt: expires})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
