package main

// a094 4.4 — 기동 순서: 따라잡기는 Recovery.Run 이 **성공한 뒤**, 준비 신호 앞(recoverThenReady 가 이 클로저를 부르고
// 그 반환 뒤에만 ready 를 부름 — TestEngineRuntime… 이 행동으로 고정). Run 이 실패하면 따라잡기를 하지 않는다.
// 좌표가 아니라 순서를 구조로 단언한다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA094TheBootCatchUpRunsAfterASuccessfulRecovery(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "engine.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var lit *ast.FuncLit
	ast.Inspect(f, func(n ast.Node) bool {
		vs, ok := n.(*ast.ValueSpec)
		if !ok || len(vs.Names) != 1 || vs.Names[0].Name != "engineRecoverySequence" || len(vs.Values) != 1 {
			return true
		}
		outer, _ := vs.Values[0].(*ast.FuncLit)
		if outer == nil || len(outer.Body.List) != 1 {
			return false
		}
		if ret, ok := outer.Body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
			lit, _ = ret.Results[0].(*ast.FuncLit)
		}
		return false
	})
	if lit == nil {
		t.Fatal("engineRecoverySequence no longer returns a closure — re-audit the boot order")
	}
	var order []string
	for _, stmt := range lit.Body.List {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			if c, ok := s.Rhs[0].(*ast.CallExpr); ok {
				if sel, ok := c.Fun.(*ast.SelectorExpr); ok {
					order = append(order, sel.Sel.Name)
				}
			}
		case *ast.IfStmt:
			cond, ok := s.Cond.(*ast.BinaryExpr)
			isErrCheck := ok && cond.Op == token.NEQ && identName(cond.X) == "err" && identName(cond.Y) == "nil"
			if len(s.Body.List) == 1 && isErrCheck {
				if _, ok := s.Body.List[0].(*ast.ReturnStmt); ok {
					order = append(order, "return-on-error")
				}
			}
		case *ast.ExprStmt:
			if c, ok := s.X.(*ast.CallExpr); ok {
				if sel, ok := c.Fun.(*ast.SelectorExpr); ok {
					order = append(order, sel.Sel.Name)
				}
			}
		case *ast.ReturnStmt:
			order = append(order, "return")
		}
	}
	want := []string{"Run", "return-on-error", "CatchUpExitProposals", "return"}
	if len(order) != len(want) {
		t.Fatalf("closure order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("closure order = %v, want %v — the catch-up must follow a successful Run and precede the return that lets ready fire", order, want)
		}
	}
}

func identName(e ast.Expr) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}
