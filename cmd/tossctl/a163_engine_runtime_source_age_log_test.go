package main

// a163 tasks 2.11(a) — engineRuntime 이 청산 관측자에게 체결 나이 보고 전용 로거(SourceAgeLog)를 엔진 로거로 넘김(freeze 리뷰 P1-3).
// 관측자 전체 Log 는 생산에서 nil 이고 알림 경로(RecordOnly)는 필드를 지우므로, 이 배선이 없으면 시험만 초록이고 생산 로그는 0 줄임.
// a129 의 구조 핀과 같은 방법: 이 함수의 engine.ExitObserverOptions 리터럴 키를 go/parser 로 셈. 행동 절반은 internal/app/engine 의 a163 시험.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA163EngineRuntimeWiresTheSourceAgeLogger(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "engine.go", nil, 0)
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
			if lit, ok := n.(*ast.CompositeLit); ok {
				if sel, ok := lit.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "ExitObserverOptions" {
					literals = append(literals, lit)
				}
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
	value, ok := keys["SourceAgeLog"]
	if !ok {
		t.Fatal("engineRuntime does not wire SourceAgeLog — the source-age lines would reach no sink in production")
	}
	if id, ok := value.(*ast.Ident); !ok || id.Name != "logger" {
		t.Fatalf("SourceAgeLog = %T, want the engine logger parameter", value)
	}
	if _, ok := keys["SourceAgeReport"]; ok {
		t.Fatal("engineRuntime sets SourceAgeReport — the toggle comes from the loaded config in Context.ExitObserver, never from a caller")
	}
}
