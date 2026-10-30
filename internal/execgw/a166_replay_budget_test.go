package execgw

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func a166ReplayBudget(t *testing.T) *accountbudget.Budget {
	t.Helper()
	root := t.TempDir()
	if e := os.Chmod(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(root, "scope", "instance"); e != nil {
		t.Fatal(e)
	}
	b, e := accountbudget.OpenContext(context.Background(), accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := b.Close(); e != nil {
			t.Error(e)
		}
	})
	return b
}
func TestA166ReplayRequiredNilFailsClosed(t *testing.T) {
	transport, err := NewBudgetedHTTPReplay("https://example.invalid", &http.Client{}, nil, nil, func(error) {})
	if transport != nil || !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("required nil: %v", err)
	}
}
func TestA166ReplayTokenAndStoredBodyShareBudget(t *testing.T) {
	b := a166ReplayBudget(t)
	var mu sync.Mutex
	var paths []string
	var orderBody string
	var orderHeader http.Header
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		mu.Lock()
		paths = append(paths, r.URL.Path)
		if r.URL.Path == replayPlacePath {
			orderBody = string(raw)
			orderHeader = r.Header.Clone()
		}
		mu.Unlock()
		switch r.URL.Path {
		case "/oauth2/token":
			w.Write([]byte(`{"access_token":"fixture-token","expires_in":3600,"token_type":"Bearer"}`))
		case replayPlacePath:
			w.Write([]byte(`{"result":{"orderId":"id1","clientOrderId":"k1"}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	template := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}
	hc := &http.Client{Transport: template, Timeout: 5 * time.Second}
	auth, e := official.NewWithAccountBudget(official.Credentials{APIKey: "test", SecretKey: "test"}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	replay, e := NewBudgetedHTTPReplay(s.URL, hc, auth, b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	body := `{"clientOrderId":"k1", "sentinel":123}`
	result, e := replay.ReplayPlace(accountbudget.WithPriority(context.Background(), accountbudget.Critical), ReplayBody{stored: body})
	if e != nil {
		t.Fatal(e)
	}
	if result.Status != 200 || result.OrderID != "id1" {
		t.Fatalf("response %+v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 2 || paths[0] != "/oauth2/token" || paths[1] != replayPlacePath {
		t.Fatalf("paths %v", paths)
	}
	if orderBody != body || orderHeader.Get(official.HeaderAuthorization) != "Bearer fixture-token" || orderHeader.Get(official.HeaderAccount) != "17" {
		t.Fatalf("wire body/header mismatch")
	}
}
func TestA166ReplayRejectsForeignBudgetAndEndpointBeforeHeaders(t *testing.T) {
	b := a166ReplayBudget(t)
	other := a166ReplayBudget(t)
	auth, e := official.NewWithAccountBudget(official.Credentials{}, "", b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	hc := &http.Client{Transport: &http.Transport{}, Timeout: time.Second}
	for _, tc := range []struct {
		base   string
		budget *accountbudget.Budget
	}{{auth.BaseURL(), other}, {"https://unrelated.invalid", b}} {
		r, e := NewBudgetedHTTPReplay(tc.base, hc, auth, tc.budget, func(error) {})
		if r != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
			t.Fatalf("mismatch accepted %v", e)
		}
	}
}

func TestA166ReplayEnforcingRefusesRedirectAndPreservesCallerClient(t *testing.T) {
	b := a166ReplayBudget(t)
	var tokenCalls, posts, redirects atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokenCalls.Add(1)
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
		case replayPlacePath:
			posts.Add(1)
			http.Redirect(w, r, "/redirected", http.StatusTemporaryRedirect)
		case "/redirected":
			redirects.Add(1)
		}
	}))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	base := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}
	original := &http.Client{Transport: base, Timeout: 4 * time.Second}
	auth, e := official.NewWithAccountBudget(official.Credentials{APIKey: "test", SecretKey: "test"}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(original), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	replay, e := NewBudgetedHTTPReplay(s.URL, original, auth, b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	if original.Transport != base || original.CheckRedirect != nil || base.TLSNextProto != nil || len(base.TLSClientConfig.NextProtos) != 0 {
		t.Fatal("caller mutated")
	}
	result, e := replay.ReplayPlace(context.Background(), ReplayBody{stored: `{"clientOrderId":"k1"}`})
	if e != nil {
		t.Fatal(e)
	}
	if result.Status != 307 || posts.Load() != 1 || redirects.Load() != 0 || tokenCalls.Load() != 1 {
		t.Fatalf("status=%d token=%d post=%d redirected=%d", result.Status, tokenCalls.Load(), posts.Load(), redirects.Load())
	}
}

func TestA166ReplayDirectPostCannotBypassHeldBudget(t *testing.T) {
	b := a166ReplayBudget(t)
	var posts atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		posts.Add(1)
		w.Write([]byte(`{"result":{"orderId":"id"}}`))
	}))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	hc := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}, Timeout: 3 * time.Second}
	auth, e := official.NewWithAccountBudget(official.Credentials{APIKey: "test", SecretKey: "test"}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = auth.AuthHeaders(context.Background()); e != nil {
		t.Fatal(e)
	}
	replay, e := NewBudgetedHTTPReplay(s.URL, hc, auth, b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	held, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(accountbudget.WithPriority(context.Background(), accountbudget.Critical), 100*time.Millisecond)
	defer cancel()
	_, e = replay.ReplayPlace(ctx, ReplayBody{stored: `{"clientOrderId":"key"}`})
	if releaseErr := held.Release(); releaseErr != nil {
		t.Fatal(releaseErr)
	}
	if !errors.Is(e, accountbudget.ErrNotSent) || posts.Load() != 0 {
		t.Fatalf("error=%v posts=%d", e, posts.Load())
	}
}
