package official

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newAccountSequenceTestClient(t *testing.T, srv *httptest.Server, opts ...Option) *Client {
	t.Helper()
	base := []Option{WithBaseURL(srv.URL), WithHTTPClient(srv.Client())}
	base = append(base, opts...)
	return New(
		Credentials{APIKey: "test-api-key-000000", SecretKey: "test-secret"},
		filepath.Join(t.TempDir(), "token.json"),
		base...,
	)
}

func writeAccountSequenceTestToken(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
}

func writeAccountSequenceTestList(w http.ResponseWriter, seq int) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"result":[{"accountNo":"123-45","accountSeq":`+
		strconv.Itoa(seq)+`,"accountType":"BROKERAGE"}]}`)
}

func writeAccountSequenceTestEcho(w http.ResponseWriter, header string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"result":{"account":"`+header+`"}}`)
}

func scopedAccountEcho(ctx context.Context, c *Client) (string, error) {
	var out struct {
		Account string `json:"account"`
	}
	err := c.getAcct(ctx, "/api/v1/account-echo", nil, &out)
	return out.Account, err
}

func waitForClientAccountSeqMutexWaiters(t *testing.T, minimum int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	buf := make([]byte, 1<<20)
	for time.Now().Before(deadline) {
		n := runtime.Stack(buf, true)
		waiters := 0
		for _, stack := range strings.Split(string(buf[:n]), "\n\n") {
			if strings.Contains(stack, "sync.(*Mutex).Lock") &&
				strings.Contains(stack, "(*Client).ensureAccountSeq") {
				waiters++
			}
		}
		if waiters >= minimum {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("account-sequence mutex waiters < %d before deadline", minimum)
}

func TestAccountsPrimesTheSequenceForTheNextScopedRead(t *testing.T) {
	var accountCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) != 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := newAccountSequenceTestClient(t, srv)
	if _, err := c.Accounts(context.Background()); err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	got, err := scopedAccountEcho(context.Background(), c)
	if err != nil {
		t.Fatalf("scoped read after Accounts: %v", err)
	}
	if got != "7" {
		t.Fatalf("X-Tossinvest-Account = %q, want 7", got)
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want 1", calls)
	}
}

func TestAccountsPreservesAnExplicitPositiveSequence(t *testing.T) {
	var accountCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			accountCalls.Add(1)
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := newAccountSequenceTestClient(t, srv, WithAccountSeq(99))
	if _, err := c.Accounts(context.Background()); err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	got, err := scopedAccountEcho(context.Background(), c)
	if err != nil {
		t.Fatalf("scoped read: %v", err)
	}
	if got != "99" {
		t.Fatalf("X-Tossinvest-Account = %q, want explicit 99", got)
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want public read only", calls)
	}
}

func TestExplicitNegativeSequenceIsNeverSent(t *testing.T) {
	var accountCalls atomic.Int32
	var scopedCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			accountCalls.Add(1)
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			scopedCalls.Add(1)
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := newAccountSequenceTestClient(t, srv, WithAccountSeq(-1))
	if _, err := scopedAccountEcho(context.Background(), c); err == nil {
		t.Fatal("scoped read accepted an explicit negative account sequence")
	}
	if calls := scopedCalls.Load(); calls != 0 {
		t.Fatalf("scoped endpoint calls = %d, want 0", calls)
	}
	if calls := accountCalls.Load(); calls != 0 {
		t.Fatalf("/api/v1/accounts calls = %d, want 0 for an invalid explicit sequence", calls)
	}
}

func TestDiscoveredNonpositiveSequenceIsNeverSent(t *testing.T) {
	for _, seq := range []int{0, -7} {
		t.Run(strconv.Itoa(seq), func(t *testing.T) {
			var accountCalls atomic.Int32
			var scopedCalls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					writeAccountSequenceTestToken(w)
				case "/api/v1/accounts":
					discovered := seq
					if accountCalls.Add(1) > 1 {
						discovered = 7
					}
					writeAccountSequenceTestList(w, discovered)
				case "/api/v1/account-echo":
					scopedCalls.Add(1)
					writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
				default:
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(srv.Close)

			c := newAccountSequenceTestClient(t, srv)
			if _, err := scopedAccountEcho(context.Background(), c); err == nil {
				t.Fatalf("scoped read accepted discovered account sequence %d", seq)
			}
			if calls := scopedCalls.Load(); calls != 0 {
				t.Fatalf("scoped endpoint calls = %d, want 0", calls)
			}
			if selected, usable := c.SelectedAccountSeq(); usable || selected != 0 {
				t.Fatalf("selection after invalid discovery = (%d, %t), want (0, false)",
					selected, usable)
			}
			got, err := scopedAccountEcho(context.Background(), c)
			if err != nil {
				t.Fatalf("scoped retry after invalid discovery: %v", err)
			}
			if got != "7" {
				t.Fatalf("retry X-Tossinvest-Account = %q, want 7", got)
			}
			if calls := accountCalls.Load(); calls != 2 {
				t.Fatalf("/api/v1/accounts calls = %d, want invalid + retry", calls)
			}
			if calls := scopedCalls.Load(); calls != 1 {
				t.Fatalf("scoped endpoint calls after retry = %d, want 1", calls)
			}
		})
	}
}

func TestEmptyDiscoveryCannotProduceAScopedHeader(t *testing.T) {
	var accountCalls atomic.Int32
	var scopedCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) > 1 {
				writeAccountSequenceTestList(w, 7)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"result":[]}`)
		case "/api/v1/account-echo":
			scopedCalls.Add(1)
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := newAccountSequenceTestClient(t, srv)
	if _, err := scopedAccountEcho(context.Background(), c); err == nil {
		t.Fatal("scoped read accepted an empty account discovery")
	}
	if calls := scopedCalls.Load(); calls != 0 {
		t.Fatalf("scoped endpoint calls = %d, want 0", calls)
	}
	if selected, usable := c.SelectedAccountSeq(); usable || selected != 0 {
		t.Fatalf("selection after empty discovery = (%d, %t), want (0, false)",
			selected, usable)
	}
	got, err := scopedAccountEcho(context.Background(), c)
	if err != nil {
		t.Fatalf("scoped retry after empty discovery: %v", err)
	}
	if got != "7" {
		t.Fatalf("retry X-Tossinvest-Account = %q, want 7", got)
	}
	if calls := accountCalls.Load(); calls != 2 {
		t.Fatalf("/api/v1/accounts calls = %d, want empty + retry", calls)
	}
	if calls := scopedCalls.Load(); calls != 1 {
		t.Fatalf("scoped endpoint calls after retry = %d, want 1", calls)
	}
}

func TestConcurrentScopedFirstUseSharesOneDiscovery(t *testing.T) {
	var accountCalls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	var startOnce sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			accountCalls.Add(1)
			startOnce.Do(func() { close(started) })
			<-release
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	releaseDiscovery := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseDiscovery)

	c := newAccountSequenceTestClient(t, srv)
	const callers = 8
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := scopedAccountEcho(context.Background(), c)
			if err == nil && got != "7" {
				err = errors.New("scoped request used a sequence other than 7")
			}
			errs <- err
		}()
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first scoped discovery did not reach /accounts")
	}
	waitForClientAccountSeqMutexWaiters(t, callers-1)
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls while first discovery is blocked = %d, want 1", calls)
	}
	releaseDiscovery()
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("concurrent scoped reads did not complete")
	}
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent scoped read: %v", err)
		}
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want 1", calls)
	}
}

func TestScopedReadWaitsForAnInflightPublicAccountDiscovery(t *testing.T) {
	var accountCalls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	var startOnce sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) != 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			startOnce.Do(func() { close(started) })
			<-release
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	releaseDiscovery := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseDiscovery)

	c := newAccountSequenceTestClient(t, srv)
	publicDone := make(chan error, 1)
	go func() {
		_, err := c.Accounts(context.Background())
		publicDone <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("public Accounts did not reach the server")
	}

	scopedDone := make(chan error, 1)
	go func() {
		got, err := scopedAccountEcho(context.Background(), c)
		if err == nil && got != "7" {
			err = errors.New("scoped request did not reuse sequence 7")
		}
		scopedDone <- err
	}()
	waitForClientAccountSeqMutexWaiters(t, 1)
	select {
	case err := <-scopedDone:
		t.Fatalf("scoped read completed before public discovery was released: %v", err)
	default:
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls while public discovery is blocked = %d, want 1", calls)
	}
	releaseDiscovery()

	for name, done := range map[string]<-chan error{
		"public Accounts": publicDone,
		"scoped read":     scopedDone,
	} {
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		case <-time.After(time.Second):
			t.Fatalf("%s deadlocked", name)
		}
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want 1", calls)
	}
}

func TestCachedScopedReadDoesNotWaitForPublicAccountListIO(t *testing.T) {
	var accountCalls atomic.Int32
	publicStarted := make(chan struct{})
	releasePublic := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) == 2 {
				close(publicStarted)
				<-releasePublic
			}
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	releaseDiscovery := sync.OnceFunc(func() { close(releasePublic) })
	t.Cleanup(releaseDiscovery)

	c := newAccountSequenceTestClient(t, srv)
	if _, err := c.Accounts(context.Background()); err != nil {
		t.Fatalf("priming Accounts: %v", err)
	}
	publicDone := make(chan error, 1)
	go func() {
		_, err := c.Accounts(context.Background())
		publicDone <- err
	}()
	select {
	case <-publicStarted:
	case <-time.After(time.Second):
		t.Fatal("second public Accounts did not reach the server")
	}

	scopedDone := make(chan error, 1)
	go func() {
		got, err := scopedAccountEcho(context.Background(), c)
		if err == nil && got != "7" {
			err = errors.New("cached scoped request did not use sequence 7")
		}
		scopedDone <- err
	}()
	select {
	case err := <-scopedDone:
		if err != nil {
			t.Fatalf("cached scoped read: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cached scoped read waited for unrelated public account-list I/O")
	}
	releaseDiscovery()
	select {
	case err := <-publicDone:
		if err != nil {
			t.Fatalf("second public Accounts: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("second public Accounts did not complete")
	}
}

func TestImplicitAccountSequenceDriftIsRejected(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "different positive", body: `[{"accountNo":"123-45","accountSeq":8}]`},
		{name: "zero", body: `[{"accountNo":"123-45","accountSeq":0}]`},
		{name: "negative", body: `[{"accountNo":"123-45","accountSeq":-8}]`},
		{name: "empty", body: `[]`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var accountCalls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					writeAccountSequenceTestToken(w)
				case "/api/v1/accounts":
					if accountCalls.Add(1) == 1 {
						writeAccountSequenceTestList(w, 7)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"result":`+tc.body+`}`)
				case "/api/v1/account-echo":
					writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
				default:
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(srv.Close)

			c := newAccountSequenceTestClient(t, srv)
			if _, err := c.Accounts(context.Background()); err != nil {
				t.Fatalf("priming Accounts: %v", err)
			}
			if _, err := c.Accounts(context.Background()); err == nil {
				t.Fatal("public Accounts accepted implicit account-sequence drift")
			}
			if seq, usable := c.SelectedAccountSeq(); !usable || seq != 7 {
				t.Fatalf("selected sequence after drift = (%d, %t), want (7, true)", seq, usable)
			}
			got, err := scopedAccountEcho(context.Background(), c)
			if err != nil {
				t.Fatalf("scoped read after rejected drift: %v", err)
			}
			if got != "7" {
				t.Fatalf("X-Tossinvest-Account after rejected drift = %q, want 7", got)
			}
			if calls := accountCalls.Load(); calls != 2 {
				t.Fatalf("/api/v1/accounts calls = %d, want prime + drift read", calls)
			}
		})
	}
}

func TestCancelledPublicDiscoveryUnlocksTheNextScopedDiscovery(t *testing.T) {
	var accountCalls atomic.Int32
	started := make(chan struct{})
	var startOnce sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) == 1 {
				startOnce.Do(func() { close(started) })
				<-r.Context().Done()
				return
			}
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := newAccountSequenceTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancelled := make(chan error, 1)
	go func() {
		_, err := c.Accounts(ctx)
		cancelled <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("cancel test did not reach /accounts")
	}
	cancel()
	select {
	case err := <-cancelled:
		if err == nil {
			t.Fatal("cancelled Accounts returned nil")
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled Accounts did not return")
	}

	done := make(chan error, 1)
	go func() {
		got, err := scopedAccountEcho(context.Background(), c)
		if err == nil && got != "7" {
			err = errors.New("retry used a sequence other than 7")
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("scoped discovery after cancellation: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("scoped discovery deadlocked after cancellation")
	}
	if calls := accountCalls.Load(); calls != 2 {
		t.Fatalf("/api/v1/accounts calls = %d, want cancelled + retry", calls)
	}
}

func TestFailedPublicDiscoveryDoesNotPrimeTheSequence(t *testing.T) {
	var accountCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := newAccountSequenceTestClient(t, srv)
	if _, err := c.Accounts(context.Background()); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("Accounts error = %v, want ErrRateLimited", err)
	}
	got, err := scopedAccountEcho(context.Background(), c)
	if err != nil {
		t.Fatalf("scoped discovery after 429: %v", err)
	}
	if got != "7" {
		t.Fatalf("X-Tossinvest-Account = %q, want 7", got)
	}
	if calls := accountCalls.Load(); calls != 2 {
		t.Fatalf("/api/v1/accounts calls = %d, want failed + retry", calls)
	}
}

func TestMalformedPublicDiscoveryDoesNotPrimeTheSequence(t *testing.T) {
	var accountCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeAccountSequenceTestToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) == 1 {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w,
					`{"result":[{"accountNo":"123-45","accountSeq":7},`)
				return
			}
			writeAccountSequenceTestList(w, 7)
		case "/api/v1/account-echo":
			writeAccountSequenceTestEcho(w, r.Header.Get("X-Tossinvest-Account"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := newAccountSequenceTestClient(t, srv)
	if _, err := c.Accounts(context.Background()); err == nil {
		t.Fatal("Accounts accepted malformed JSON")
	}
	if seq, usable := c.SelectedAccountSeq(); usable || seq != 0 {
		t.Fatalf("malformed discovery primed sequence (%d, %t), want (0, false)", seq, usable)
	}
	got, err := scopedAccountEcho(context.Background(), c)
	if err != nil {
		t.Fatalf("scoped discovery after malformed response: %v", err)
	}
	if got != "7" {
		t.Fatalf("X-Tossinvest-Account = %q, want 7", got)
	}
	if calls := accountCalls.Load(); calls != 2 {
		t.Fatalf("/api/v1/accounts calls = %d, want malformed + retry", calls)
	}
}
