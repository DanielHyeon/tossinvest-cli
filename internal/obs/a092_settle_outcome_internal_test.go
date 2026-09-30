package obs

// a092 26라운드 codex P2 #6: 선점 판정은 한 곳 — 행 없음 · 모르는 결과는 선점이 아님. 동기 발송의 두 자리가 그 판정을 씀.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestA092OnlyNamedOutcomesArePreemptions(t *testing.T) {
	for outcome, want := range map[journal.SettleOutcome]bool{
		journal.SettleApplied: false, journal.SettleAlreadySettled: true, journal.SettleLeaseLost: true,
		journal.SettleNotFound: false, journal.SettleOutcome(99): false,
	} {
		if got := isPreemption(outcome); got != want {
			t.Errorf("isPreemption(%v) = %v, want %v", outcome, got, want)
		}
	}
}

func TestA092DeliverClassifiesThroughOneJudgement(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "notifier.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "deliver" {
			continue
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "isPreemption" {
					calls++
				}
			}
			return true
		})
	}
	if calls != 2 {
		t.Errorf("deliver judges preemption through isPreemption %d times, want 2 (attempt record and release)", calls)
	}
}
