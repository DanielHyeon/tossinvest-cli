package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA166ProbeRequiredUsesNormalPriority(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "execute" {
			return true
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "WithPriority" {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "accountbudget" {
				return true
			}
			arg, ok := call.Args[1].(*ast.SelectorExpr)
			if ok && arg.Sel.Name == "Normal" {
				found = true
			}
			return true
		})
		return false
	})
	if !found {
		t.Fatal("required probe does not assign Normal priority")
	}
}

func TestA166ProbeRequiredMissingAfterLocalGuards(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	var out bytes.Buffer
	d := h.deps(&out)
	calls := 0
	d.prepareBudget = func(ctx context.Context, dir string) (*accountbudgetruntime.Factory, *accountbudget.Budget, *official.Client, error) {
		calls++
		return prepareProbeBudget(ctx, dir, "required", true, filepath.Join(t.TempDir(), "missing"), true)
	}
	code, err := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", "invalid", "--out", h.out}, d)
	if code != exitError || err == nil || calls != 0 {
		t.Fatalf("local guard code=%d err=%v calls=%d", code, err, calls)
	}
	code, err = run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if code != exitError || !errors.Is(err, accountbudget.ErrUnavailable) || calls != 1 {
		t.Fatalf("required code=%d err=%v calls=%d", code, err, calls)
	}
}

func TestA166ProbeFactoryLocalTokenAndPost(t *testing.T) {
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	base := t.TempDir()
	root := filepath.Join(base, "budget")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	eventID := func() accountbudget.FileID {
		b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
		if err != nil {
			t.Fatal(err)
		}
		id, err := accountbudget.ProvisionObservation(b)
		err = errors.Join(err, b.Close())
		if err != nil {
			t.Fatal(err)
		}
		return id
	}()
	fileID := func(path string) accountbudget.FileID {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		st := info.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(st.Dev), Inode: st.Ino}
	}
	identity := accountbudget.Identity{Root: fileID(root), Queue: fileID(filepath.Join(root, "queue.lock")), Wire: fileID(filepath.Join(root, "wire.lock"))}
	data, err := json.Marshal(map[string]any{"event_lock": eventID, "version": 2, "root": root, "scope": "scope", "instance": "instance", "identity": identity})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(base, "manifest.json")
	if err := os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	var tokenCalls, posts atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokenCalls.Add(1)
			_, _ = w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
		case orderPath:
			posts.Add(1)
			_, _ = w.Write([]byte(`{"result":{"orderId":"id1"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != want {
			return nil, errors.New("fixture refused address")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	var out bytes.Buffer
	d := h.deps(&out)
	d.budgetTemplate = tr
	d.prepareBudget = func(ctx context.Context, dir string) (*accountbudgetruntime.Factory, *accountbudget.Budget, *official.Client, error) {
		f, err := accountbudgetruntime.Open(ctx, accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
		if err != nil {
			return nil, nil, nil, err
		}
		b, err := f.Budget()
		if err != nil {
			return nil, nil, nil, errors.Join(err, f.Close())
		}
		c, err := f.NewClient(official.Credentials{APIKey: "fixture", SecretKey: "fixture"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(17))
		if err != nil {
			return nil, nil, nil, errors.Join(err, f.Close())
		}
		return f, b, c, nil
	}
	code, err := run(context.Background(), []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--out", h.out}, d)
	if err != nil || code != exitOK || tokenCalls.Load() != 1 || posts.Load() != 1 {
		t.Fatalf("code=%d err=%v tokens=%d posts=%d", code, err, tokenCalls.Load(), posts.Load())
	}
}
