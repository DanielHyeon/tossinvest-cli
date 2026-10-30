package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestA166CandidateSoakCanceledBeforeConstruction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := buildCandidatePanel(ctx, &rootOptions{}, "KR"); !errors.Is(err, context.Canceled) {
		t.Fatalf("candidate cancellation: %v", err)
	}
	if _, err := buildSoakWiring(ctx, &rootOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("soak cancellation: %v", err)
	}
}

func TestA166CandidateSoakLocalOptionsGuard(t *testing.T) {
	for _, target := range []struct{ file, fn string }{{"candidatepanel.go", "buildCandidatePanel"}, {"soak.go", "buildSoakWiring"}} {
		f, err := parser.ParseFile(token.NewFileSet(), target.file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Name.Name != target.fn {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				c, ok := n.(*ast.CallExpr)
				if !ok || len(c.Args) != 5 || !c.Ellipsis.IsValid() {
					return true
				}
				id, ok := c.Fun.(*ast.Ident)
				if !ok || id.Name != "newRootOfficialClient" {
					return true
				}
				v, ok := c.Args[4].(*ast.SelectorExpr)
				if ok && v.Sel.Name == "officialOptions" {
					found = true
				}
				return true
			})
		}
		if !found {
			t.Errorf("%s missing fixture options", target.fn)
		}
	}
}

func TestA166CandidateSoakBudgetLocalTLS(t *testing.T) {
	for _, kind := range []string{"candidate", "soak"} {
		t.Run(kind, func(t *testing.T) {
			dir := isolate(t)
			writeCredentials(t, dir)
			var calls atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.URL.Path == "/oauth2/token" {
					fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
					return
				}
				http.Error(w, "fixture failure", 500)
			}))
			defer server.Close()
			tr := server.Client().Transport.(*http.Transport).Clone()
			tr.Proxy = nil
			want := server.Listener.Addr().String()
			tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
				if network != "tcp" || address != want {
					return nil, errors.New("fixture refused destination")
				}
				return (&net.Dialer{}).DialContext(ctx, network, want)
			}
			if conn, err := tr.DialContext(context.Background(), "tcp", "127.0.0.1:1"); err == nil {
				conn.Close()
				t.Fatal("fixture accepted a different destination")
			}
			opts := []official.Option{official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr})}
			owner := a166LocalRequiredOwner(t)
			factory, err := owner.get(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			budget, err := factory.Budget()
			if err != nil {
				t.Fatal(err)
			}
			root := &rootOptions{configDir: dir, budgetOwner: owner, officialOptions: opts}
			var read func(context.Context) error
			switch kind {
			case "candidate":
				panel, err := buildCandidatePanel(context.Background(), root, "KR")
				if err != nil || len(panel) == 0 {
					t.Fatalf("panel %v %v", panel, err)
				}
				read = func(ctx context.Context) error { _, err := panel[0].Read(ctx, "KR"); return err }
			case "soak":
				wiring, err := buildSoakWiring(context.Background(), root)
				if err != nil {
					t.Fatal(err)
				}
				read = func(ctx context.Context) error { _, err := wiring.reads.Accounts(ctx); return err }
			}
			lease, err := budget.Acquire(context.Background(), accountbudget.Critical)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
			err = read(ctx)
			cancel()
			if err == nil || calls.Load() != 0 {
				t.Fatalf("budget bypass err=%v calls=%d", err, calls.Load())
			}
			if err = lease.Release(); err != nil {
				t.Fatal(err)
			}
			err = read(context.Background())
			if err == nil || calls.Load() < 2 {
				t.Fatalf("local read err=%v calls=%d", err, calls.Load())
			}
		})
	}
}

func TestA166CandidateSoakRequiredFailure(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir)
	owner := newRootBudgetOwner(func(k string) (string, bool) {
		if k == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	root := &rootOptions{configDir: dir, budgetOwner: owner}
	if panel, err := buildCandidatePanel(context.Background(), root, "KR"); !errors.Is(err, accountbudget.ErrUnavailable) || panel != nil {
		t.Fatalf("candidate required: %v %v", panel, err)
	}
	if _, err := buildSoakWiring(context.Background(), root); !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("soak required: %v", err)
	}
}

func TestA166CandidateWiringReleasesStoreOnFactoryFailure(t *testing.T) {
	previousStore, previousPanel := candidateStoreFactory, candidatePanelFactory
	t.Cleanup(func() { candidateStoreFactory, candidatePanelFactory = previousStore, previousPanel })
	released := 0
	sentinel := errors.New("fixture factory error")
	candidateStoreFactory = func(context.Context, *rootOptions) (*candidate.Store, func(), error) {
		return nil, func() { released++ }, nil
	}
	ctx := context.WithValue(context.Background(), a166CandidateContextKey{}, true)
	candidatePanelFactory = func(got context.Context, _ *rootOptions, _ string) ([]candidate.Source, error) {
		if got != ctx {
			t.Error("candidate context changed")
		}
		return nil, sentinel
	}
	_, _, _, err := candidateWiring(ctx, &rootOptions{}, "KR")
	if !errors.Is(err, sentinel) || released != 1 {
		t.Fatalf("err=%v releases=%d", err, released)
	}
}

type a166CandidateContextKey struct{}
