package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
)

func TestA166OutcomeRequiredMissingManifest(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", filepath.Join(t.TempDir(), "missing"))
	owner := &candleOwner{ctx: context.Background(), mode: "required", modeSet: true, manifest: filepath.Join(t.TempDir(), "missing"), manifestSet: true}
	_, _, err := owner.openCandles(t.TempDir())
	if !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("expected required budget refusal, got %v", err)
	}
}

func TestA166OutcomeEmptyAndCloseFailure(t *testing.T) {
	var out bytes.Buffer
	sentinel := errors.New("close failed")
	calls := 0
	d := testDeps(t, &out)
	d.close = func() error { calls++; return sentinel }
	code, err := run(context.Background(), []string{"--journal", tempJournal(t)}, d)
	if code != 1 || !errors.Is(err, sentinel) || calls != 1 {
		t.Fatalf("code=%d err=%v close calls=%d", code, err, calls)
	}
}

type a166PriorityTransport struct{ base http.RoundTripper }

func (p a166PriorityTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if accountbudget.PriorityFromContext(req.Context()) != accountbudget.Auxiliary {
		return nil, errors.New("missing auxiliary priority")
	}
	return p.base.RoundTrip(req)
}

func TestA166OutcomeClosureLocalToken(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/candles":
			_, _ = w.Write([]byte(`{"result":{"candles":[{"timestamp":"2026-10-07T09:19:00+09:00","openPrice":"18540","highPrice":"18600","lowPrice":"18420","closePrice":"18430","volume":"275728","currency":"KRW"}],"nextBefore":"2026-10-07T09:18:00+09:00"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	tr := server.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	want := server.Listener.Addr().String()
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	owner := &candleOwner{ctx: context.Background(), opts: []official.Option{official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: a166PriorityTransport{base: tr}})}}
	read, _, err := owner.openCandles(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	label := time.Date(2026, 10, 7, 0, 19, 0, 0, time.UTC)
	price, bar, ok, err := read(context.Background(), "KR", "000001", label)
	if err != nil || !ok || price != "18430" || !bar.Equal(label) || calls.Load() != 2 {
		t.Fatalf("price=%s ok=%v err=%v calls=%d", price, ok, err, calls.Load())
	}
	if err = owner.close(); err != nil {
		t.Fatal(err)
	}
}

func TestA166OutcomeBusyAuxiliaryNoWire(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	base := t.TempDir()
	root := filepath.Join(base, "budget")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		t.Helper()
		s, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := s.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	ident := accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}
	raw, e := json.Marshal(map[string]any{"version": 1, "root": root, "scope": "scope", "instance": "instance", "identity": ident})
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(base, "manifest.json")
	if e = os.WriteFile(manifest, raw, 0600); e != nil {
		t.Fatal(e)
	}
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", manifest)
	b, e := accountbudget.OpenContext(context.Background(), accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	defer lease.Release()
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	tr := server.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	want := server.Listener.Addr().String()
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	owner := &candleOwner{ctx: context.Background(), mode: "required", modeSet: true, manifest: manifest, manifestSet: true, opts: []official.Option{official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr})}}
	read, _, err := owner.openCandles(base)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.close()
	_, _, _, err = read(context.Background(), "KR", "000001", time.Now())
	if !errors.Is(err, accountbudget.ErrUnavailable) || calls.Load() != 0 {
		t.Fatalf("busy err=%v calls=%d", err, calls.Load())
	}
}
