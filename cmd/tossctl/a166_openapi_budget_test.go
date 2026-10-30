package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestA166OpenAPIRootBinding(t *testing.T) {
	for _, target := range []struct{ file, fn string }{{"openapi.go", "newOpenAPICmd"}, {"init.go", "runInitOfficialFlow"}, {"doctor_openapi.go", "computeDoctorOpenAPISummary"}, {"console_openapi.go", "defaultConsoleOpenAPIDeps"}} {
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
				if !ok {
					return true
				}
				id, ok := c.Fun.(*ast.Ident)
				if ok && id.Name == "validateOpenAPICredentialsForRoot" {
					found = true
				}
				return true
			})
		}
		if !found {
			t.Errorf("%s must use root-bound validator", target.fn)
		}
	}
}

func TestA166OpenAPIBudgetLocalTLS(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"fake","accountSeq":17,"accountType":"BROKERAGE"}]}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected", 500)
		}
	}))
	defer server.Close()
	tr := server.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	opts := []official.Option{official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr})}
	creds := official.Credentials{APIKey: "fake", SecretKey: "fake"}
	bad := newRootBudgetOwner(func(k string) (string, bool) {
		if k == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	result, err := validateOpenAPICredentialsForRoot(context.Background(), &rootOptions{budgetOwner: bad, officialOptions: opts}, creds, "")
	if !errors.Is(err, accountbudget.ErrUnavailable) || result.OK || calls.Load() != 0 {
		t.Fatalf("required: result=%+v err=%v calls=%d", result, err, calls.Load())
	}
	owner := a166LocalRequiredOwner(t)
	factory, err := owner.get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	budget, err := factory.Budget()
	if err != nil {
		t.Fatal(err)
	}
	lease, err := budget.Acquire(context.Background(), accountbudget.Critical)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	result, err = defaultConsoleOpenAPIDeps(&rootOptions{budgetOwner: owner, officialOptions: opts}).validate(ctx, creds, "")
	cancel()
	if err == nil && result.OK || calls.Load() != 0 {
		t.Fatalf("console bypass: result=%+v err=%v calls=%d", result, err, calls.Load())
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	result, err = defaultConsoleOpenAPIDeps(&rootOptions{budgetOwner: owner, officialOptions: opts}).validate(context.Background(), creds, "")
	if err != nil || !result.OK || calls.Load() != 2 {
		t.Fatalf("local: result=%+v err=%v calls=%d", result, err, calls.Load())
	}
}

func TestA166ConsoleOnboardingBudgetFailureDoesNotSave(t *testing.T) {
	dir := isolate(t)
	owner := newRootBudgetOwner(func(k string) (string, bool) {
		if k == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	deps := defaultConsoleOpenAPIDeps(&rootOptions{configDir: dir, budgetOwner: owner})
	var saved bool
	deps.save = func(string, string, string) error { saved = true; return nil }
	var temporary string
	temp := deps.temp
	deps.temp = func(path string) (string, func() error, error) {
		p, cleanup, err := temp(path)
		temporary = p
		return p, cleanup, err
	}
	seam, err := newConsoleOpenAPISeam(&rootOptions{configDir: dir}, deps)
	if err != nil {
		t.Fatal(err)
	}
	result := seam.Save(context.Background(), "fake", "fake")
	if result.State == consoleOpenAPIReady || saved {
		t.Fatalf("factory failure accepted: state=%v saved=%v", result.State, saved)
	}
	if temporary == "" {
		t.Fatal("missing isolated token path")
	}
	if _, err := os.Stat(temporary); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary token not removed: %v", err)
	}
}

func TestA166ConsoleOnboardingRootWiring(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "console.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "runConsole" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok || len(c.Args) != 1 {
				return true
			}
			id, ok := c.Fun.(*ast.Ident)
			if !ok || id.Name != "defaultConsoleOpenAPIDeps" {
				return true
			}
			arg, ok := c.Args[0].(*ast.Ident)
			if ok && arg.Name == "root" {
				found = true
			}
			return true
		})
	}
	if !found {
		t.Fatal("console onboarding must pass its root owner")
	}
}
