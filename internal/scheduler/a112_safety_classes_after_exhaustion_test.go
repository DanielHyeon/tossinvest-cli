package scheduler

// a112 태스크 2.8 의 빈칸(2.x 대조 감사): 저우선 · 전략 예산 고갈 뒤 안전 등급 중 **비상 청산 하나만** 허용됨이 재어졌다. reconcile ·
// fill detection · protection 은 그 상태에서 확인된 적이 없다. 여기서는 안전 등급을 손으로 고르지 않고 isSafetyClass 로 전체 등급에서 걸러
// 하나씩 잰다 — 등급이 늘면 그 등급도 자동으로 들어온다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestEverySafetyClassStaysCallableAfterStrategyCapacityIsExhausted(t *testing.T) {
	at := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	c := a112Coordinator(t, at, 20) // 예비 10 · 재량 10
	key := budget(at, 20).Path
	scope := StrategyScope{Market: strategyrouter.MarketKR, Horizon: strategyrouter.HorizonShort, Family: strategyrouter.FamilyBreakoutRetest}
	for i := 0; ; i++ {
		grant := c.TryAcquireStrategy(key, scope, PollEntry, at)
		if !grant.Allowed {
			if grant.Refusal != StrategyBudgetDeferred || i != 10 {
				t.Fatalf("arrangement: exhaustion after %d grants with refusal %q, want 10 then %s", i, grant.Refusal, StrategyBudgetDeferred)
			}
			break
		}
	}
	all := []PollClass{PollCandidate, PollEntry, PollAnalytics, PollEmergencyExit, PollReconcile, PollFillDetection, PollProtection}
	// 위 목록이 이 패키지의 PollClass 상수 전부인지 AST 로 센다 — 등급이 늘었는데 목록이 그대로면 새 등급이 이 시험을 비켜 간다.
	if declared := a112PollClassConstants(t); declared != len(all) {
		t.Fatalf("PollClass constants=%d, this census lists %d — add the new class here", declared, len(all))
	}
	safety := 0
	for _, class := range all {
		if !isSafetyClass(class) {
			continue
		}
		safety++
		// 안전 등급은 예산 회계 밖이다(SAFETY_PRIORITY · commitment 없음) — 예비 수치는 아래 전략 거절이 보고한다.
		grant := c.TryAcquire(key, class, at)
		if !grant.Allowed || grant.Reason != BudgetSafetyPriority {
			t.Fatalf("safety class %s after strategy exhaustion: allowed=%v reason=%s, want allowed by safety priority", class, grant.Allowed, grant.Reason)
		}
	}
	if safety != 4 {
		t.Fatalf("safety classes=%d, want the four (emergency exit · reconcile · fill detection · protection)", safety)
	}
	// 안전 등급이 돈 뒤에도 전략 진입은 여전히 미뤄진다 — 예비는 전략 쪽으로 새지 않는다.
	if grant := c.TryAcquireStrategy(key, scope, PollEntry, at); grant.Allowed || grant.Reserve != SafetyReserve(20) {
		t.Fatalf("after the safety grants a strategy entry got allowed=%v reserve=%d, want deferred with the reserve %d intact",
			grant.Allowed, grant.Reserve, SafetyReserve(20))
	}
}

func a112PollClassConstants(t *testing.T) int {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "budget.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				if ident, ok := value.Type.(*ast.Ident); ok && ident.Name == "PollClass" {
					count += len(value.Names)
				}
			}
		}
	}
	return count
}
