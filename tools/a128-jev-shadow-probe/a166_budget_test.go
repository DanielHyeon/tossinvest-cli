package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type a166AuxTransport struct{ base http.RoundTripper }

func (a a166AuxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if accountbudget.PriorityFromContext(req.Context()) != accountbudget.Auxiliary {
		return nil, errors.New("missing auxiliary priority")
	}
	return a.base.RoundTrip(req)
}

func TestA166JEVRequiredMissingManifest(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "required")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", filepath.Join(t.TempDir(), "missing"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := runCommand(ctx, []string{"--market", "KR", "--out", filepath.Join(t.TempDir(), "ledger"), "--session", "09:00-15:30"}, func(string) string { return "fixture" })
	if !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("expected required budget refusal, got %v", err)
	}
}

func TestA166JEVOffLocalTokenFailure(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fixture-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fixture-secret")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MODE", "off")
	t.Setenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST", "")
	if err := os.Unsetenv("TOSSOS_ACCOUNT_BUDGET_MANIFEST"); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	base := t.TempDir()
	err := runCommandWithOptions(ctx, []string{"--market", "KR", "--out", filepath.Join(base, "ledger")}, func(string) string { return "fixture" }, official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: a166AuxTransport{base: tr}}))
	if err == nil || calls.Load() != 1 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestA166JEVBusyAuxiliaryNoWire(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := runCommandWithOptions(ctx, []string{"--market", "KR", "--out", filepath.Join(base, "ledger")}, func(string) string { return "fixture" }, official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}))
	if !errors.Is(err, context.DeadlineExceeded) || calls.Load() != 0 {
		t.Fatalf("busy err=%v calls=%d", err, calls.Load())
	}
}

func TestA166JEVCleanupError(t *testing.T) {
	original := errors.New("fixture operation")
	closing := errors.New("fixture close")
	err := finishTool(original, func() error { return closing })
	if !errors.Is(err, original) || !errors.Is(err, closing) {
		t.Fatalf("lost cause: %v", err)
	}
}
