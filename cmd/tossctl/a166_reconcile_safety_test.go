package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
)

// Prevent a local-TLS fixture from silently losing the caller's test options.
func TestA166EngineReconcileOptionsGuard(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "engine_reconcile.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "buildReconcileResolveRuntime" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if !ok || id.Name != "newRootOfficialClient" {
				return true
			}
			if len(call.Args) == 5 && call.Ellipsis.IsValid() {
				if opt, ok := call.Args[4].(*ast.SelectorExpr); ok && opt.Sel.Name == "officialOptions" {
					if root, ok := opt.X.(*ast.Ident); ok && root.Name == "root" {
						found = true
					}
				}
			}
			return true
		})
	}
	if !found {
		t.Fatal("engine reconcile must pass fixture options to root client")
	}
}

func TestA166EngineReconcileLocalTLS(t *testing.T) {
	dir := isolate(t)
	writeConfigFile(t, dir, config.DefaultFile())
	writeCredentials(t, dir)
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"fake","accountSeq":1,"accountType":"BROKERAGE"}]}`)
		default:
			t.Errorf("unexpected local path %s", r.URL.Path)
			http.Error(w, "unexpected", 500)
		}
	}))
	defer server.Close()
	transport := server.Client().Transport.(*http.Transport).Clone()
	transport.Proxy = nil
	opts := []official.Option{official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: transport})}
	bad := newRootBudgetOwner(func(key string) (string, bool) {
		if key == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	_, err := buildReconcileResolveRuntime(context.Background(), &rootOptions{configDir: dir, budgetOwner: bad, officialOptions: opts})
	if !errors.Is(err, accountbudget.ErrUnavailable) || calls.Load() != 0 {
		t.Fatalf("required failure: err=%v local calls=%d", err, calls.Load())
	}
	rt, err := buildReconcileResolveRuntime(context.Background(), &rootOptions{configDir: dir, officialOptions: opts})
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("local calls=%d", calls.Load())
	}
	if err := rt.close(); err != nil {
		t.Fatal(err)
	}
	owner := a166LocalRequiredOwner(t)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	before := calls.Load()
	_, err = buildReconcileResolveRuntime(canceled, &rootOptions{configDir: dir, budgetOwner: owner, officialOptions: opts})
	if !errors.Is(err, context.Canceled) || calls.Load() != before {
		t.Fatalf("canceled startup: err=%v local calls=%d", err, calls.Load()-before)
	}
	rt, err = buildReconcileResolveRuntime(context.Background(), &rootOptions{configDir: dir, budgetOwner: owner, officialOptions: opts})
	if err != nil {
		t.Fatal(err)
	}
	defer rt.close()
	if owner.factory == nil {
		t.Fatal("required factory not created")
	}
	if calls.Load() < 3 {
		t.Fatalf("required did not reach local accounts: %d", calls.Load())
	}
}

func a166LocalRequiredOwner(t *testing.T) *rootBudgetOwner {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "group")
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
	id := func(path string) accountbudget.FileID {
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	manifest := struct {
		Version   int                    `json:"version"`
		Root      string                 `json:"root"`
		Scope     string                 `json:"scope"`
		Instance  string                 `json:"instance"`
		Identity  accountbudget.Identity `json:"identity"`
		EventLock accountbudget.FileID   `json:"event_lock"`
	}{EventLock: eventID, Version: 2, Root: root, Scope: "scope", Instance: "instance", Identity: accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(base, "manifest.json")
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	o := newRootBudgetOwner(func(key string) (string, bool) {
		switch key {
		case "TOSSOS_ACCOUNT_BUDGET_MODE":
			return "required", true
		case "TOSSOS_ACCOUNT_BUDGET_MANIFEST":
			return path, true
		}
		return "", false
	}, nil)
	t.Cleanup(func() {
		if err := o.Close(); err != nil {
			t.Error(err)
		}
	})
	return o
}
