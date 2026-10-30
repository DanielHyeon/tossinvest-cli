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
	"sync/atomic"
	"testing"
	"time"
)

func TestA166VerifyReconcileFactoryBinding(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "verify_reconcile.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "buildVerifyReconcileReader" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if ok && id.Name == "newVerifyReconcileReaderForRoot" && len(call.Args) >= 4 {
				if v, ok := call.Args[1].(*ast.Ident); ok && v.Name == "root" {
					found = true
				}
			}
			return true
		})
	}
	if !found {
		t.Fatal("production reader must use root factory")
	}
}

func TestA166VerifyReconcileBudgetLocalTLS(t *testing.T) {
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
	bad := newRootBudgetOwner(func(k string) (string, bool) {
		if k == "TOSSOS_ACCOUNT_BUDGET_MODE" {
			return "required", true
		}
		return "", false
	}, nil)
	creds := official.Credentials{APIKey: "fake", SecretKey: "fake"}
	_, _, err := newVerifyReconcileReaderForRoot(context.Background(), &rootOptions{budgetOwner: bad}, creds, "", opts...)
	if !errors.Is(err, accountbudget.ErrUnavailable) || calls.Load() != 0 {
		t.Fatalf("required: err=%v calls=%d", err, calls.Load())
	}
	owner := a166LocalRequiredOwner(t)
	factory, err := owner.get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	b, err := factory.Budget()
	if err != nil {
		t.Fatal(err)
	}
	lease, err := b.Acquire(context.Background(), accountbudget.Critical)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	_, _, err = newVerifyReconcileReaderForRoot(ctx, &rootOptions{budgetOwner: owner}, creds, "", opts...)
	cancel()
	if err == nil || calls.Load() != 0 {
		t.Fatalf("first client bypassed budget: err=%v calls=%d", err, calls.Load())
	}
	if err := lease.Release(); err != nil {
		t.Fatal(err)
	}
	reader, account, err := newVerifyReconcileReaderForRoot(context.Background(), &rootOptions{budgetOwner: owner}, creds, "", opts...)
	if err != nil {
		t.Fatal(err)
	}
	if account.Seq != 17 {
		t.Fatalf("seq=%d", account.Seq)
	}
	budget, err := owner.factory.Budget()
	if err != nil {
		t.Fatal(err)
	}
	if !reader.(*reconcileReads).client.UsesAccountBudget(budget) {
		t.Fatal("bound client lost budget")
	}
	if calls.Load() != 2 {
		t.Fatalf("local calls=%d", calls.Load())
	}
}
