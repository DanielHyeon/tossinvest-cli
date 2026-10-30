package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA166L1CRequiredMissingManifestFailsBeforeWire(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", filepath.Join(t.TempDir(), "missing"))
	err := run("KR", "005930", t.TempDir())
	if !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("expected required failure, got %v", err)
	}
}

func TestA166L1CUsesLocalTLSAndStopsOnReadError(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "off")
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
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
	err := runWithOptions("KR", "005930", t.TempDir(), official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if err == nil || calls.Load() != 2 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestA166L1CAuxiliaryBusySendsNothing(t *testing.T) {
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
	fileID := func(path string) accountbudget.FileID {
		t.Helper()
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	identity := accountbudget.Identity{Root: fileID(root), Queue: fileID(filepath.Join(root, "queue.lock")), Wire: fileID(filepath.Join(root, "wire.lock"))}
	raw, err := json.Marshal(map[string]any{"version": 1, "root": root, "scope": "scope", "instance": "instance", "identity": identity})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(base, "manifest.json")
	if err = os.WriteFile(manifest, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", manifest)
	budget, err := accountbudget.OpenContext(context.Background(), accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	defer budget.Close()
	lease, err := budget.Acquire(context.Background(), accountbudget.Critical)
	if err != nil {
		t.Fatal(err)
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
	err = runWithOptions("KR", "005930", t.TempDir(), official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if !errors.Is(err, accountbudget.ErrUnavailable) || calls.Load() != 0 {
		t.Fatalf("busy err=%v calls=%d", err, calls.Load())
	}
}

func TestA166L1CCloseErrorRetainsOriginal(t *testing.T) {
	original := errors.New("fixture operation")
	closing := errors.New("fixture close")
	err := finishL1C(original, func() error { return closing })
	if !errors.Is(err, original) || !errors.Is(err, closing) {
		t.Fatalf("missing cause: %v", err)
	}
	if err = finishL1C(nil, func() error { return closing }); !errors.Is(err, closing) {
		t.Fatalf("close failure hidden: %v", err)
	}
}

type a166DeadlineTransport struct {
	base    http.RoundTripper
	checked atomic.Int32
}

func (d *a166DeadlineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	deadline, ok := req.Context().Deadline()
	if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > probeBudget {
		return nil, errors.New("missing probe deadline")
	}
	d.checked.Add(1)
	return d.base.RoundTrip(req)
}

func TestA166L1COffTwoReadsOutputAndDeadline(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "off")
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/orderbook":
			_, _ = w.Write([]byte(`{"result":{"timestamp":"2026-08-18T03:29:14.120+09:00","currency":"USD","asks":[{"price":"231.7000","volume":"160"}],"bids":[{"price":"231.6500","volume":"40"}]}}`))
		case "/api/v1/prices":
			_, _ = w.Write([]byte(`{"result":[{"symbol":"AAPL","lastPrice":"231.6800","currency":"USD","timestamp":"2026-08-18T03:29:14.100+09:00"}]}`))
		default:
			t.Errorf("unexpected local path: %s", r.URL.Path)
			w.WriteHeader(404)
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
	tracker := &a166DeadlineTransport{base: tr}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	prior := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = prior; _ = read.Close() }()
	err = runWithOptions("US", "AAPL", t.TempDir(), official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tracker}))
	_ = write.Close()
	os.Stdout = prior
	out, readErr := io.ReadAll(read)
	if err != nil || readErr != nil || calls.Load() != 3 || tracker.checked.Load() != 3 || !strings.Contains(string(out), "AAPL") {
		t.Fatalf("err=%v read=%v calls=%d deadlines=%d output=%q", err, readErr, calls.Load(), tracker.checked.Load(), out)
	}
}
