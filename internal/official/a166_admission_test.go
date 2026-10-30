//go:build linux

package official

import (
	"context"
	"crypto/x509"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func a166Budget(t *testing.T) *accountbudget.Budget {
	t.Helper()
	root := t.TempDir()
	if e := os.Chmod(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(root, "scope", "instance"); e != nil {
		t.Fatal(e)
	}
	b, e := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { b.Close() })
	return b
}
func TestA166RequiredBudgetMissing(t *testing.T) {
	c, err := NewWithAccountBudget(Credentials{}, "", nil, func(error) {})
	if c != nil || !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("missing budget accepted: %v", err)
	}
}
func TestA166AuthorityAndClientIsolation(t *testing.T) {
	b := a166Budget(t)
	c, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := c.AuthorityOrigin(); !ok {
		t.Fatal("default authority lost")
	}
	if c.tm.hc != c.hc {
		t.Fatal("token client differs")
	}
	WithBaseURL("https://invalid.example")(c)
	WithHTTPClient(&http.Client{})(c)
	if _, ok := c.AuthorityOrigin(); !ok || c.tm.hc != c.hc {
		t.Fatal("sealed default option replay changed authority or token client")
	}
	originalTransport := &http.Transport{}
	original := &http.Client{Transport: originalTransport}
	opt := WithHTTPClient(original)
	override, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {}, opt)
	if e != nil {
		t.Fatal(e)
	}
	if original.Transport == override.hc.Transport {
		t.Fatal("caller client mutated")
	}
	if originalTransport.TLSNextProto != nil || originalTransport.TLSClientConfig != nil {
		t.Fatal("caller transport mutated")
	}
	if _, ok := override.AuthorityOrigin(); ok {
		t.Fatal("override gained authority")
	}
	opt(override)
	if _, ok := override.AuthorityOrigin(); ok {
		t.Fatal("option replay gained authority")
	}
	if override.tm.hc != override.hc {
		t.Fatal("replay changed client")
	}
}
func TestA166ActualClientToken401Redirect(t *testing.T) {
	b := a166Budget(t)
	var tokens, orders, final atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokens.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token":"token","expires_in":3600,"token_type":"Bearer"}`))
		case "/order":
			if orders.Add(1) == 1 {
				w.WriteHeader(401)
				return
			}
			http.Redirect(w, r, "/final", http.StatusTemporaryRedirect)
		case "/final":
			final.Add(1)
			w.Write([]byte(`{"result":{}}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	hc := s.Client()
	tr := hc.Transport.(*http.Transport).Clone()
	tr.TLSClientConfig.RootCAs = roots
	tr.Proxy = nil
	original := &http.Client{Transport: tr}
	c, e := NewWithAccountBudget(Credentials{APIKey: "key", SecretKey: "secret"}, "", b, func(error) {}, WithBaseURL(s.URL), WithHTTPClient(original))
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := c.AuthorityOrigin(); ok {
		t.Fatal("test server authority")
	}
	e = c.send(context.Background(), func(tok string) (*http.Request, error) {
		return http.NewRequestWithContext(context.Background(), http.MethodGet, s.URL+"/order", nil)
	}, nil)
	if e != nil {
		t.Fatal(e)
	}
	if tokens.Load() < 2 || orders.Load() != 2 || final.Load() != 1 {
		t.Fatalf("token=%d order=%d final=%d", tokens.Load(), orders.Load(), final.Load())
	}
	if !strings.HasPrefix(c.base, "https://") {
		t.Fatal("wrong test server")
	}
}

func TestA166CanceledPrewriteIdentity(t *testing.T) {
	b := a166Budget(t)
	c, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.invalid/", nil)
	if e != nil {
		t.Fatal(e)
	}
	_, _, e = c.doRequest(req)
	if !errors.Is(e, accountbudget.ErrNotSent) || !errors.Is(e, ErrTransport) {
		t.Fatalf("lost prewrite evidence: %v", e)
	}
}
func TestA166RejectsNilClient(t *testing.T) {
	b := a166Budget(t)
	c, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {}, WithHTTPClient(nil))
	if c != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("nil client: %v", e)
	}
}

type a166FailTransport struct{ err error }

func (f a166FailTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, f.err }
func TestA166OffErrorContract(t *testing.T) {
	sentinel := errors.New("synthetic failure")
	c := New(Credentials{}, "", WithHTTPClient(&http.Client{Transport: a166FailTransport{sentinel}}))
	req, e := http.NewRequest(http.MethodGet, "https://example.invalid/path", nil)
	if e != nil {
		t.Fatal(e)
	}
	_, _, err := c.doRequest(req)
	if !errors.Is(err, ErrTransport) || errors.Is(err, sentinel) || !strings.Contains(err.Error(), "synthetic failure") {
		t.Fatalf("OFF contract changed: %v", err)
	}
}

func TestA166TokenPrewriteIdentity(t *testing.T) {
	b := a166Budget(t)
	c, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e = c.tm.exchange(ctx)
	if !errors.Is(e, accountbudget.ErrNotSent) || !errors.Is(e, ErrTransport) {
		t.Fatalf("lost token prewrite identity: %v", e)
	}
}

func TestA166Prior401IsNotErasedByLaterPrewriteError(t *testing.T) {
	b := a166Budget(t)
	var calls atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(http.StatusUnauthorized) }))
	defer s.Close()
	tr := s.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	c, e := NewWithAccountBudget(Credentials{}, "", b, func(error) {}, WithBaseURL(s.URL), WithHTTPClient(&http.Client{Transport: tr}))
	if e != nil {
		t.Fatal(e)
	}
	c.tm.cache = &cachedToken{AccessToken: "cached", ExpiresAt: time.Now().Add(time.Hour)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wrote atomic.Int32
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{WroteHeaders: func() { wrote.Add(1) }})
	ctx = WithAttemptObserver(ctx, func(a AttemptTrace) {
		if a.StatusCode == http.StatusUnauthorized {
			cancel()
		}
	})
	e = c.send(ctx, func(tok string) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodPost, s.URL+"/order", strings.NewReader("test"))
	}, nil)
	if !errors.Is(e, accountbudget.ErrNotSent) || wrote.Load() < 1 || calls.Load() != 1 {
		t.Fatalf("err=%v writes=%d calls=%d", e, wrote.Load(), calls.Load())
	}
}

func TestA166RequiredCleanupMissing(t *testing.T) {
	b := a166Budget(t)
	c, e := NewWithAccountBudget(Credentials{}, "", b, nil)
	if c != nil || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("missing cleanup accepted: %v", e)
	}
}
