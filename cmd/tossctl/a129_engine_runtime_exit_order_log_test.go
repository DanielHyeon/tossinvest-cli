package main

// a129 생산 조립 쪽 — engineRuntime 이 exit 관측자에게 청산 주문 전용 로거(ExitOrderLog)를 엔진 로거로 넘기고 관측자 전체 Log 는
// 넘기지 않음. a090 의 구조 핀과 같은 방법: 이 함수의 engine.ExitObserverOptions 합성 리터럴을 go/parser 로 읽어 키를 셈.
// 행동 절반 — Context.ExitObserver 가 그 필드를 통과시키고 그 로거에 줄이 나옴 — 은 internal/app/engine 의 a129 시험이 잼.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA129EngineRuntimeWiresTheDedicatedExitOrderLogger(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "engine.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing engine.go: %v", err)
	}
	var literals []*ast.CompositeLit
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "engineRuntime" || fn.Recv != nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			if sel, ok := lit.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "ExitObserverOptions" {
				literals = append(literals, lit)
			}
			return true
		})
	}
	if len(literals) != 1 {
		t.Fatalf("engineRuntime builds %d ExitObserverOptions literals, want exactly one", len(literals))
	}
	keys := map[string]ast.Expr{}
	for _, elt := range literals[0].Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			t.Fatal("an unkeyed ExitObserverOptions field — the census cannot read it")
		}
		keys[kv.Key.(*ast.Ident).Name] = kv.Value
	}
	value, ok := keys["ExitOrderLog"]
	if !ok {
		t.Fatal("engineRuntime does not wire ExitOrderLog — an exit order leaves no line in engine.log (2026-10-07 010170)")
	}
	if id, ok := value.(*ast.Ident); !ok || id.Name != "logger" {
		t.Fatalf("ExitOrderLog = %T, want the engine logger parameter", value)
	}
	if _, ok := keys["Log"]; ok {
		t.Fatal("engineRuntime wires the observer's whole Log — its existing lines carry the account in the clear")
	}
}
