package engine_test

// 리뷰 P2-1: a090 의 경과는 단조 앵커로만 잼(spec 「단조 시계」 · design D3). 주입 시계는 lease 확장을 구현하지 않아
// clock.LeaseAnchor 가 Now 로 떨어지므로, 앵커를 o.clk.Now() 로 바꾼 변이는 행동 시험을 통과함 — 생산 systemClock.Now 는 UTC 로
// 단조 성분을 벗김. 그래서 exit_unobserved.go 의 시계 호출을 구조로 못 박음(a111 의 lease 헬퍼 핀과 같은 방식).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestA090ElapsedUsesOnlyTheLeaseHelpers(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "exit_unobserved.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing exit_unobserved.go: %v", err)
	}
	calls := map[string]int{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch x := sel.X.(type) {
		case *ast.Ident:
			calls[x.Name+"."+sel.Sel.Name]++
		case *ast.SelectorExpr: // o.clk.Now()
			if id, ok := x.X.(*ast.Ident); ok {
				calls[id.Name+"."+x.Sel.Name+"."+sel.Sel.Name]++
			}
		}
		return true
	})
	if calls["clock.LeaseAnchor"] != 1 {
		t.Errorf("clock.LeaseAnchor calls = %d, want exactly the one in unobservedMoment", calls["clock.LeaseAnchor"])
	}
	if calls["clock.LeaseElapsed"] < 2 {
		t.Errorf("clock.LeaseElapsed calls = %d, want the threshold and the release measurements", calls["clock.LeaseElapsed"])
	}
	for _, forbidden := range []string{"o.clk.Now", "o.clk.Since", "time.Now", "time.Since"} {
		if calls[forbidden] != 0 {
			t.Errorf("%s is called %d times — elapsed must come from the monotonic lease helpers", forbidden, calls[forbidden])
		}
	}
}
