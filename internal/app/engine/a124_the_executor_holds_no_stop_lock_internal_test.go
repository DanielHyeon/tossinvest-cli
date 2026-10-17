package engine

// a124 tasks 2.6 (a) — 격리 단언(구조). 배달 실행자는 Notifier 를 들고 있지 않고 부르지도 않는다 —
// 그래서 손절 경로(exit 루프 · 비상 청산의 Notify)가 기다리는 n.mu 를 잡을 길이 없다(design D5 · D7, Q1).
// 행동 쪽 단언은 태그 파일 `a124_the_executor_holds_no_stop_lock_testseams_test.go`.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestTheExecutorCarriesNoNotifier(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "alertdelivery.go", nil, 0)
	if err != nil {
		t.Fatalf("parse alertdelivery.go: %v", err)
	}
	var seen []string
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if x.Sel.Name == "Notifier" {
				seen = append(seen, fset.Position(x.Pos()).String())
			}
		case *ast.Ident:
			if x.Name == "Notifier" {
				seen = append(seen, fset.Position(x.Pos()).String())
			}
		}
		return true
	})
	if len(seen) != 0 {
		t.Fatalf("alertdelivery.go names the Notifier at %v — the executor must not reach the mutex the stop path waits on", seen)
	}
}
