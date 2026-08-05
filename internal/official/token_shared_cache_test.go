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

// TestARotationThatLandsMidRequestCostsNoToken is the window the ten production
// refusals came out of.
//
// A rotation that lands after the token is handed out and before the broker
// answers is the only way a refusal survives send()'s retry at all; anywhere else
// the retry absorbs it. Buying a token here would kill the one the rotating
// holder just started using, which is how the loop restarts.
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

// TestAnAdoptedTokenThatIsAlsoRefusedStillEndsOnAMintedOne.
//
// An adopted token is not a verified one — its liveness is inferred from its
// expiry — and send() has one retry to spend. At a rotation boundary the holder
// that wrote the file last need not be the one that minted last, so the file can
// hold a token the broker has already dropped. Spending the only retry on it and
// surfacing the refusal is not a small matter: ErrAuth reaches
// execgw.ClassAuthFatal, which latches the entry gate and persists the latch so a
// restart cannot lift it, and the exit-loop cycle that raised it makes no
// stop-loss judgement.
func TestAnAdoptedTokenThatIsAlsoRefusedStillEndsOnAMintedOne(t *testing.T) {
	broker := &lastTokenWinsServer{}
	srv := httptest.NewServer(broker.handler(t))
	defer srv.Close()
	c, _, cache := twoProcesses(t, srv)

	if err := ping(t, c); err != nil {
		t.Fatal(err)
	}
	// The process holds a token the broker has dropped, and the file holds a
	// different one it has also dropped: the tempting thing to adopt is dead.
	writeCachedToken(t, cache, "T-written-last-but-dead", time.Now().Add(24*time.Hour))
	c.tm.mu.Lock()
	c.tm.cache = &cachedToken{AccessToken: "T-minted-earlier", ExpiresAt: time.Now().Add(24 * time.Hour)}
	c.tm.mu.Unlock()

	if err := ping(t, c); err != nil {
		t.Errorf("the request ended on a refusal: %v — the retry was spent on an adopted "+
			"token that was also dead, and nothing bought a live one", err)
	}
}

// TestASiblingGoroutineThatAlreadyReplacedTheTokenIsNotOutbought.
//
// The engine runs every supervised loop as its own goroutine against one client,
// so two loops refused on the same stale token arrive at refresh together. If the
// second one exchanges it invalidates the token the first one just obtained,
// which is the cross-process defect happening inside a single process.
func TestASiblingGoroutineThatAlreadyReplacedTheTokenIsNotOutbought(t *testing.T) {
	broker := &lastTokenWinsServer{}
	srv := httptest.NewServer(broker.handler(t))
	defer srv.Close()
	cache := filepath.Join(t.TempDir(), "openapi-token.json")
	m := newTokenManager(Credentials{APIKey: "k", SecretKey: "s"}, srv.URL, cache, srv.Client())

	stale, err := m.token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// The first sibling answers the refusal and replaces the shared token.
	if _, _, err := m.refresh(context.Background(), stale); err != nil {
		t.Fatal(err)
	}
	before := broker.exchanges()

	// Seven more siblings arrive holding the same stale token they were refused on.
	const siblings = 7
	var wg sync.WaitGroup
	for i := 0; i < siblings; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := m.refresh(context.Background(), stale); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if got := broker.exchanges() - before; got != 0 {
		t.Errorf("%d siblings refused on an already-replaced token bought %d more; the "+
			"replacement was sitting in front of them", siblings, got)
	}
}

// TestAReaderNeverSeesAHalfWrittenCacheFile.
//
// A plain write truncates first, and a reader landing in that window parses an
// empty file, concludes it has no token and buys one — invalidating the token the
// writer just obtained. That window matters more since a082, because the read
// that adopts a token happens exactly when another holder has just written one.
func TestAReaderNeverSeesAHalfWrittenCacheFile(t *testing.T) {
	dir := t.TempDir()
	cache := filepath.Join(dir, "openapi-token.json")
	m := newTokenManager(Credentials{APIKey: "k", SecretKey: "s"}, "", cache, nil)
	writeCachedToken(t, cache, "seed", time.Now().Add(24*time.Hour))

	const rounds = 400
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < rounds; i++ {
			ct := &cachedToken{
				AccessToken: fmt.Sprintf("T%d-%s", i, strings.Repeat("x", 512)),
				ExpiresAt:   time.Now().Add(24 * time.Hour),
			}
			if err := m.saveCache(ct); err != nil {
				t.Errorf("save: %v", err)
				return
			}
		}
	}()
	torn := 0
	go func() {
		defer wg.Done()
		for i := 0; i < rounds; i++ {
			ct, err := m.loadCache()
			if err != nil || ct == nil || ct.AccessToken == "" {
				torn++
			}
		}
	}()
	wg.Wait()

	if torn != 0 {
		t.Errorf("%d of %d reads saw a file that was missing or unparseable while it was "+
			"being replaced; a reader in that window buys a token and invalidates the "+
			"writer's", torn, rounds)
	}
}

// TestAdoptionRefusesAnExpiredFileToken.
//
// The file is shared, so what it holds can be stale in either direction. Adopting
// an expired token spends send()'s retry on something the broker will refuse, and
// the second pass has to buy one anyway — the expiry check is what keeps the first
// pass from being wasted.
func TestAdoptionRefusesAnExpiredFileToken(t *testing.T) {
	broker := &lastTokenWinsServer{}
	srv := httptest.NewServer(broker.handler(t))
	defer srv.Close()
	cache := filepath.Join(t.TempDir(), "openapi-token.json")
	m := newTokenManager(Credentials{APIKey: "k", SecretKey: "s"}, srv.URL, cache, srv.Client())

	writeCachedToken(t, cache, "T-expired", time.Now().Add(-time.Hour))
	tok, adopted, err := m.refresh(context.Background(), "T-refused")
	if err != nil {
		t.Fatal(err)
	}
	if adopted || tok == "T-expired" {
		t.Errorf("adopted=%v tok=%q; the file's token was already expired and taking it "+
			"spends the caller's only retry on a certain refusal", adopted, tok)
	}
}

// TestAdoptionKeysOnTheTokenItself.
//
// "Adopt only something different" is the whole narrowness argument, and different
// has to mean a different token — not a different expiry, not a newer file. A file
// holding the same token with a later expiry is the same dead token.
func TestAdoptionKeysOnTheTokenItself(t *testing.T) {
	broker := &lastTokenWinsServer{}
	srv := httptest.NewServer(broker.handler(t))
	defer srv.Close()
	cache := filepath.Join(t.TempDir(), "openapi-token.json")
	m := newTokenManager(Credentials{APIKey: "k", SecretKey: "s"}, srv.URL, cache, srv.Client())

	// Same token as the one refused, but stamped further into the future.
	writeCachedToken(t, cache, "T-refused", time.Now().Add(48*time.Hour))
	before := broker.exchanges()
	tok, adopted, err := m.refresh(context.Background(), "T-refused")
	if err != nil {
		t.Fatal(err)
	}
	if adopted || tok == "T-refused" {
		t.Errorf("adopted=%v tok=%q; a later expiry on the same token is still the token "+
			"the broker just refused", adopted, tok)
	}
	if got := broker.exchanges() - before; got != 1 {
		t.Errorf("bought %d token(s), want exactly 1", got)
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
