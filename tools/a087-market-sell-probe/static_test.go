package main

// static_test.go 는 "단 1회 전송" 과 "주문 표면은 이 도구의 POST 하나" 를 도달 가능성으로 고정함
// (선례: tools/a128-jev-shadow-probe/static_test.go).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func nonTestSources(t *testing.T) map[string]*ast.File {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(thisFile), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]*ast.File{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		files[filepath.Base(path)] = parsed
	}
	if len(files) == 0 {
		t.Fatal("no source files found")
	}
	return files
}

func funcDecl(t *testing.T, name string) *ast.FuncDecl {
	t.Helper()
	for _, file := range nonTestSources(t) {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name && fn.Recv == nil {
				return fn
			}
		}
	}
	t.Fatalf("function %s not found", name)
	return nil
}

// TestOnlySendOnceCallsDoAndItHasNoLoop 는 http 전송(.Do / .Post / .Get / RoundTrip) 호출이 소스 전체에서
// sendOnce 안의 .Do 하나뿐이고, sendOnce 에 반복·goto·재귀가 없음을 봄.
func TestOnlySendOnceCallsDoAndItHasNoLoop(t *testing.T) {
	t.Parallel()
	transmit := map[string]bool{"Do": true, "Post": true, "PostForm": true, "Get": true, "Head": true, "RoundTrip": true}
	sites := map[string]int{}
	for _, file := range nonTestSources(t) {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && transmit[sel.Sel.Name] {
					sites[fn.Name.Name+"."+sel.Sel.Name]++
				}
				return true
			})
		}
	}
	if len(sites) != 1 || sites["sendOnce.Do"] != 1 {
		t.Fatalf("transmit call sites = %v, want exactly one sendOnce.Do", sites)
	}
	ast.Inspect(funcDecl(t, "sendOnce").Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			t.Errorf("sendOnce contains a loop at %v", n.Pos())
		case *ast.BranchStmt:
			if n.Tok == token.GOTO {
				t.Error("sendOnce contains goto")
			}
		case *ast.CallExpr:
			if ident, ok := n.Fun.(*ast.Ident); ok && ident.Name == "sendOnce" {
				t.Error("sendOnce calls itself")
			}
		}
		return true
	})
}

// TestSendOnceIsCalledOnceFromExecute 는 sendOnce 호출 자리가 execute 안의 하나뿐이고 반복문 밖임을 봄.
func TestSendOnceIsCalledOnceFromExecute(t *testing.T) {
	t.Parallel()
	callers := map[string]int{}
	for _, file := range nonTestSources(t) {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "sendOnce" {
						callers[fn.Name.Name]++
					}
				}
				return true
			})
		}
	}
	if len(callers) != 1 || callers["execute"] != 1 {
		t.Fatalf("sendOnce callers = %v, want exactly one call in execute", callers)
	}
	ast.Inspect(funcDecl(t, "execute").Body, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.ForStmt, *ast.RangeStmt, *ast.GoStmt:
			t.Errorf("execute contains a loop or goroutine (%T) — the single send must not be repeatable", node)
		}
		return true
	})
}

// TestSendClientRefusesRedirectsAndConnectionReuse 는 net/http 의 숨은 재전송 두 길이 닫혀 있음을 봄.
func TestSendClientRefusesRedirectsAndConnectionReuse(t *testing.T) {
	t.Parallel()
	client := newSendClient(time.Second)
	transport, ok := client.Transport.(*http.Transport)
	if !ok || !transport.DisableKeepAlives {
		t.Fatal("the send transport must disable keep-alives (no reused-connection replay)")
	}
	if client.CheckRedirect == nil || client.CheckRedirect(nil, nil) != http.ErrUseLastResponse {
		t.Fatal("the send client must not follow redirects")
	}
	if client.Timeout != time.Second {
		t.Fatalf("timeout = %s", client.Timeout)
	}
}

// TestNoProductionOrderSurfaceIsNamed 는 생산 주문·계좌 메서드를 부르지 않고 official 사용이 허용 목록 안임을 봄.
// AuthHeaders 는 headerSource 인터페이스 메서드로만 불림.
func TestNoProductionOrderSurfaceIsNamed(t *testing.T) {
	t.Parallel()
	forbidden := map[string]bool{
		"PlaceOrder": true, "CancelOrder": true, "ModifyOrder": true, "Place": true, "Cancel": true, "Amend": true,
		"CreateConditionalOrder": true, "CancelConditionalOrder": true, "ModifyConditionalOrder": true,
		"Accounts": true, "BuyingPower": true, "Holdings": true, "HoldingsRaw": true, "SellableQuantity": true,
		"SellableQuantityRaw": true, "Orders": true, "OrderByID": true, "OrdersRaw": true, "OrdersPageRaw": true,
		"OrderRawByID": true, "SaveCredentials": true, "DeleteCredentials": true, "WithBaseURL": true,
		"WithHTTPClient": true, "WithAccountSeq": true,
	}
	allowedOfficial := map[string]bool{
		"New": true, "LoadCredentials": true, "HeaderAuthorization": true, "HeaderAccount": true,
	}
	sawOfficial := false
	for name, file := range nonTestSources(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if forbidden[selector.Sel.Name] {
				t.Errorf("%s names %s; the probe has its own single POST and no other order or account surface", name, selector.Sel.Name)
			}
			if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == "official" {
				sawOfficial = true
				if !allowedOfficial[selector.Sel.Name] {
					t.Errorf("%s uses official.%s, outside the allowlist", name, selector.Sel.Name)
				}
			}
			return true
		})
	}
	if !sawOfficial {
		t.Fatal("the guard saw no official.* use at all — it is measuring nothing")
	}
}

// TestImportsStayInsideTheAllowlist 는 생산 패키지 import 를 official·app/paths·x/term 으로 묶음 —
// internal/trading·execgw 같은 주문 경로가 들어오지 않게 함.
func TestImportsStayInsideTheAllowlist(t *testing.T) {
	t.Parallel()
	allowedModule := map[string]bool{
		"github.com/JungHoonGhae/tossinvest-cli/internal/official":  true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/app/paths": true,
		"golang.org/x/term": true,
	}
	for name, file := range nonTestSources(t) {
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			if strings.Contains(path, ".") && !allowedModule[path] {
				t.Errorf("%s imports %q, outside the allowlist", name, path)
			}
		}
	}
}
