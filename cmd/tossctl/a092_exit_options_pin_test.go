package main

// a092 22.3 C1 구조 핀: 생산 조립(engineRuntime)은 exit 관측 루프에 알림 경로 부품을 직접 넘기지 않음.
// 기록 전용 인스턴스는 Context.ExitObserver 가 주입 지점별로 만든다 — 여기서 동기 알림기(ectx.Notifier)를 넘기면
// 그 기본값이 서지 않고 exit goroutine 이 다시 원격 전송을 기다림(21라운드 C1 의 형태).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA092EngineRuntimeDoesNotHandTheExitLoopASyncAlertPath(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "engine.go", nil, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var fn *ast.FuncDecl
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == "engineRuntime" {
			fn = fd
		}
	}
	if fn == nil {
		t.Fatal("engineRuntime not found in engine.go")
	}
	literals := 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "ExitObserverOptions" {
			return true
		}
		literals++
		for _, el := range lit.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				t.Errorf("positional ExitObserverOptions literal at %s — the pin reads keys", fset.Position(el.Pos()))
				continue
			}
			switch key := kv.Key.(*ast.Ident); key.Name {
			case "Alerts", "Announcer", "Retrier", "Floor":
				t.Errorf("engineRuntime sets ExitObserverOptions.%s at %s — the exit loop's alert paths are "+
					"bound record-only by Context.ExitObserver and must not be overridden here",
					key.Name, fset.Position(kv.Pos()))
			}
		}
		return true
	})
	if literals != 1 {
		t.Errorf("ExitObserverOptions literals in engineRuntime = %d, want 1", literals)
	}
}
