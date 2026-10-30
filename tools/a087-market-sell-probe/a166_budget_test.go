package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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

func TestA166ProbeRequiredNilFailsClosed(t *testing.T) {
	c, err := newBudgetedSendClient(time.Second, nil, func(error) {}, nil)
	if c != nil || !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("required nil: %v", err)
	}
}

func probeBudget(t *testing.T) *accountbudget.Budget {
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
func TestA166ProbeBudgetedTokenAndOnePost(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	b := probeBudget(t)
	var mu sync.Mutex
	var paths []string
	var body string
	var header http.Header
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, _ := io.ReadAll(r.Body)
		mu.Lock()
		paths = append(paths, r.URL.Path)
		if r.URL.Path == orderPath {
			body = string(v)
			header = r.Header.Clone()
		}
		mu.Unlock()
		switch r.URL.Path {
		case "/oauth2/token":
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
		case orderPath:
			w.Write([]byte(`{"result":{"orderId":"id1"}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	template := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}
	hc := &http.Client{Transport: template, Timeout: 4 * time.Second}
	client, e := official.NewWithAccountBudget(official.Credentials{APIKey: "test", SecretKey: "test"}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	d := h.deps(&out)
	d.budgetRequired = true
	d.budget = b
	d.budgetClient = client
	d.budgetCleanup = func(error) {}
	d.budgetTemplate = template
	code, e := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if e != nil || code != exitOK {
		t.Fatalf("code %d error %v out %s", code, e, out.String())
	}
	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 2 || paths[0] != "/oauth2/token" || paths[1] != orderPath || body == "" || header.Get(official.HeaderAuthorization) != "Bearer fixture" || header.Get(official.HeaderAccount) != "17" {
		t.Fatal("token and direct post mismatch")
	}
	if h.opened != 0 {
		t.Fatal("legacy opener used")
	}
}
func TestA166ProbeBudgetMismatchBeforeToken(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	b := probeBudget(t)
	other := probeBudget(t)
	var calls atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	template := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}
	client, e := official.NewWithAccountBudget(official.Credentials{}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(&http.Client{Transport: template}), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	d := h.deps(&out)
	d.budgetRequired = true
	d.budget = other
	d.budgetClient = client
	d.budgetCleanup = func(error) {}
	d.budgetTemplate = template
	code, e := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if code != exitError || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("code %d err %v", code, e)
	}
	if calls.Load() != 0 || h.mock.totalCalls() != 0 || h.opened != 0 || len(h.receiptFiles()) != 0 {
		t.Fatal("mismatch side effect")
	}
}
func TestA166ProbeRequiredNilRejectsBeforeOpenAndReceipt(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	var out bytes.Buffer
	d := h.deps(&out)
	d.budgetRequired = true
	code, e := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if code != exitError || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("code %d err %v", code, e)
	}
	if h.opened != 0 || h.mock.totalCalls() != 0 || len(h.receiptFiles()) != 0 {
		t.Fatal("nil budget produced side effect")
	}
}

func TestA166ProbeClosedBudgetRetainsUnknownReceipt(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	b := probeBudget(t)
	var tokenCalls, posts atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		posts.Add(1)
		w.WriteHeader(500)
	}))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	template := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}
	hc := &http.Client{Transport: template, Timeout: 4 * time.Second}
	client, e := official.NewWithAccountBudget(official.Credentials{APIKey: "test", SecretKey: "test"}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = client.AuthHeaders(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e = b.Close(); e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	d := h.deps(&out)
	d.budgetRequired = true
	d.budget = b
	d.budgetClient = client
	d.budgetCleanup = func(error) {}
	d.budgetTemplate = template
	code, e := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if e != nil || code != exitUnknown {
		t.Fatalf("code %d error %v", code, e)
	}
	var r receipt
	if e = json.Unmarshal(h.onlyReceipt(), &r); e != nil {
		t.Fatal(e)
	}
	if r.Outcome != outcomeUnknown || posts.Load() != 0 || tokenCalls.Load() != 1 {
		t.Fatalf("outcome=%s token=%d posts=%d", r.Outcome, tokenCalls.Load(), posts.Load())
	}
}

func TestA166ProbeDirectPostCannotBypassHeldBudget(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	b := probeBudget(t)
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
	template := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}
	hc := &http.Client{Transport: template, Timeout: 3 * time.Second}
	client, e := official.NewWithAccountBudget(official.Credentials{APIKey: "test", SecretKey: "test"}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = client.AuthHeaders(context.Background()); e != nil {
		t.Fatal(e)
	}
	held, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	d := h.deps(&out)
	d.budgetRequired = true
	d.budget = b
	d.budgetClient = client
	d.budgetCleanup = func(error) {}
	d.budgetTemplate = template
	d.sendTimeout = 120 * time.Millisecond
	code, e := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if releaseErr := held.Release(); releaseErr != nil {
		t.Fatal(releaseErr)
	}
	if code != exitUnknown || e != nil || posts.Load() != 0 {
		t.Fatalf("code=%d error=%v posts=%d", code, e, posts.Load())
	}
}

func TestA166ProbeUnsupportedTemplateRejectsBeforeToken(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	b := probeBudget(t)
	var calls atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
	defer s.Close()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	safe := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}
	client, e := official.NewWithAccountBudget(official.Credentials{}, "", b, func(error) {}, official.WithBaseURL(s.URL), official.WithHTTPClient(&http.Client{Transport: safe}), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	d := h.deps(&out)
	d.budgetRequired = true
	d.budget = b
	d.budgetClient = client
	d.budgetCleanup = func(error) {}
	d.budgetTemplate = &http.Transport{TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}}
	code, e := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if code != exitError || !errors.Is(e, accountbudget.ErrUnavailable) {
		t.Fatalf("code %d err %v", code, e)
	}
	if calls.Load() != 0 || len(h.receiptFiles()) != 0 {
		t.Fatal("unsupported template caused side effect")
	}
}
