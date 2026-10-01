package scheduler

// a112 7.1 · 7.2 — 전략 capability 의 `(market, horizon, family, poll_class)` subscope 결합(Manager 판정 Q1~Q4, 2026-10-01).
//
// 계약: 범위 있는 capability 는 같은 범위로만 완료된다(교차 family · 범위 없음과의 교차 replay 금지 — 거절은 `false`, 상태 불변). 모든 범위는 같은
// 물리 endpoint · reset 세대의 남은 수 · 안전 예비 · commitment 집합 · 발급 상한 · observation cycle 을 **공유**한다(family 가 용량을 복제하지 않음).
// 거절은 동결 골든 `refusal_enums.scheduler` 의 `BUDGET_DEFERRED` 하나(세부 사유는 Detail).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func a112Scope(family strategyrouter.Family) StrategyScope {
	return StrategyScope{Market: strategyrouter.MarketKR, Horizon: strategyrouter.HorizonShort, Family: family}
}

func a112Coordinator(t *testing.T, at time.Time, remaining int) *BudgetCoordinator {
	t.Helper()
	// uniqueEntropy — incrementingEntropy 는 256 바이트마다 되돌아 32 바이트 capability 가 여덟 번째부터 겹친다(발급 충돌 → TOKEN_UNAVAILABLE).
	c := newBudgetCoordinatorWithEntropyAndClock(&uniqueEntropy{}, func() time.Time { return at.Add(time.Second) })
	c.Observe(budget(at, remaining))
	return c
}

func a112CommitmentCount(c *BudgetCoordinator, key string) (total, completed int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, record := range c.endpoints[key].commitments {
		total++
		if record.completed {
			completed++
		}
	}
	return total, completed
}

// 시나리오 「continuation capability replay」: KR SHORT continuation 에 발급된 capability 로 KR SHORT breakout 완료 → 거부, 공유 commitment 불변.
func TestAContinuationCapabilityCannotCompleteABreakoutPoll(t *testing.T) {
	at := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	c := a112Coordinator(t, at, 40)
	key := budget(at, 40).Path
	grant := c.TryAcquireStrategy(key, a112Scope(strategyrouter.FamilyContinuation), PollEntry, at)
	if !grant.Allowed || grant.Refusal != "" {
		t.Fatalf("arrangement: continuation grant=%+v", grant)
	}
	before, _ := a112CommitmentCount(c, key)
	if c.CompleteStrategy(key, a112Scope(strategyrouter.FamilyBreakoutRetest), grant.Commitment) {
		t.Fatal("a continuation capability completed a breakout poll — family scope mismatch must be refused")
	}
	if after, completed := a112CommitmentCount(c, key); after != before || completed != 0 {
		t.Fatalf("the refused cross-family completion changed shared commitments: %d→%d completed=%d", before, after, completed)
	}
	if !c.CompleteStrategy(key, a112Scope(strategyrouter.FamilyContinuation), grant.Commitment) {
		t.Fatal("control: the same capability must complete under its own scope")
	}
	if c.CompleteStrategy(key, a112Scope(strategyrouter.FamilyContinuation), grant.Commitment) {
		t.Fatal("a capability completed twice")
	}
}

// 교차 replay 양방향: 범위 있는 토큰은 범위 없는 완료로, 범위 없는 토큰은 범위 있는 완료로 끝낼 수 없다.
func TestScopedAndUnscopedCapabilitiesCannotCompleteEachOther(t *testing.T) {
	at := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	c := a112Coordinator(t, at, 40)
	key := budget(at, 40).Path
	scope := a112Scope(strategyrouter.FamilyReversal)
	scoped := c.TryAcquireStrategy(key, scope, PollEntry, at)
	if !scoped.Allowed {
		t.Fatalf("arrangement: scoped grant=%+v", scoped)
	}
	if c.Complete(key, scoped.Commitment.token) {
		t.Fatal("a scoped capability completed through the unscoped API")
	}
	unscoped := c.TryAcquire(key, PollEntry, at)
	if !unscoped.Allowed {
		t.Fatalf("arrangement: unscoped grant=%+v", unscoped)
	}
	forged := StrategyCommitmentToken{token: unscoped.Commitment, scope: scope.digest()}
	if c.CompleteStrategy(key, scope, forged) {
		t.Fatal("an unscoped capability completed through the scoped API")
	}
	if _, completed := a112CommitmentCount(c, key); completed != 0 {
		t.Fatalf("refused cross completions marked %d commitment(s) completed", completed)
	}
	if !c.CompleteStrategy(key, scope, scoped.Commitment) || !c.Complete(key, unscoped.Commitment) {
		t.Fatal("control: each capability must complete through its own API")
	}
}

// 비복제: 범위 있는 발급과 범위 없는 발급은 같은 commitment 집합 · 같은 가용 수를 쓴다 — family 를 더해도 물리 용량이 늘지 않는다.
func TestEveryScopeSharesOnePhysicalCommitmentSet(t *testing.T) {
	at := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	c := a112Coordinator(t, at, 20) // 예비 10 · 재량 10
	key := budget(at, 20).Path
	families := []strategyrouter.Family{strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal,
		strategyrouter.FamilyWeeklyValue, strategyrouter.FamilyBreakoutRetest}
	granted := 0
	for i := 0; i < 20; i++ {
		grant := c.TryAcquireStrategy(key, a112Scope(families[i%len(families)]), PollEntry, at)
		if grant.Allowed {
			granted++
			continue
		}
		if grant.Refusal != StrategyBudgetDeferred {
			t.Fatalf("refusal=%q detail=%s, want %s", grant.Refusal, grant.Detail, StrategyBudgetDeferred)
		}
	}
	if granted != 10 {
		t.Fatalf("four families were granted %d slots from one endpoint with 10 discretionary — capacity must not multiply", granted)
	}
	if extra := c.TryAcquire(key, PollEntry, at); extra.Allowed {
		t.Fatal("the unscoped API found capacity the scoped grants had already used")
	}
	if total, _ := a112CommitmentCount(c, key); total != 10 {
		t.Fatalf("commitments=%d, want one shared set of 10", total)
	}
}

// 시나리오 「마지막 physical allowance 경쟁」: 레인 여덟이 같은 endpoint · 세대의 마지막 low-priority 자리를 동시에 acquire → 하나만 commit, 나머지는
// BUDGET_DEFERRED, 안전 예비 유지.
func TestEightLanesRacingForTheLastSlotCommitOnce(t *testing.T) {
	at := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	c := a112Coordinator(t, at, 20) // 예비 10 · 재량 10
	key := budget(at, 20).Path
	for i := 0; i < 9; i++ {
		if grant := c.TryAcquire(key, PollCandidate, at); !grant.Allowed {
			t.Fatalf("arrangement: fill slot %d refused: %+v", i, grant)
		}
	}
	scopes := []StrategyScope{}
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		for _, family := range []strategyrouter.Family{strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal,
			strategyrouter.FamilyWeeklyValue, strategyrouter.FamilyBreakoutRetest} {
			horizon := strategyrouter.HorizonShort
			if family == strategyrouter.FamilyWeeklyValue {
				horizon = strategyrouter.HorizonWeekly
			}
			scopes = append(scopes, StrategyScope{Market: market, Horizon: horizon, Family: family})
		}
	}
	start := make(chan struct{})
	grants := make([]StrategyBudgetGrant, len(scopes))
	var wg sync.WaitGroup
	for i, scope := range scopes {
		wg.Add(1)
		go func(i int, scope StrategyScope) {
			defer wg.Done()
			<-start
			grants[i] = c.TryAcquireStrategy(key, scope, PollEntry, at)
		}(i, scope)
	}
	close(start)
	wg.Wait()
	allowed := 0
	for _, grant := range grants {
		switch {
		case grant.Allowed:
			allowed++
		case grant.Refusal != StrategyBudgetDeferred:
			t.Fatalf("loser refusal=%q detail=%s, want %s", grant.Refusal, grant.Detail, StrategyBudgetDeferred)
		}
		if grant.Reserve != SafetyReserve(20) {
			t.Fatalf("reserve=%d, want the safety reserve %d kept", grant.Reserve, SafetyReserve(20))
		}
	}
	if allowed != 1 {
		t.Fatalf("eight lanes committed %d of the last slot, want exactly one", allowed)
	}
	if total, _ := a112CommitmentCount(c, key); total != 10 {
		t.Fatalf("commitments=%d, want the shared set full at 10", total)
	}
	if safety := c.TryAcquire(key, PollEmergencyExit, at); !safety.Allowed {
		t.Fatal("safety-class polling was blocked by the exhausted low-priority allowance")
	}
}

// 범위 검증: 모르는 family · 시장 · horizon 은 거절(BUDGET_DEFERRED, 세부 사유 INVALID), 안전 등급은 전략 API 의 대상이 아니다(commitment 없음).
func TestAnInvalidStrategyScopeOrSafetyClassIsDeferred(t *testing.T) {
	at := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	c := a112Coordinator(t, at, 40)
	key := budget(at, 40).Path
	for name, scope := range map[string]StrategyScope{
		"unknown family":  {Market: strategyrouter.MarketKR, Horizon: strategyrouter.HorizonShort, Family: "MOMENTUM"},
		"unknown market":  {Market: "JP", Horizon: strategyrouter.HorizonShort, Family: strategyrouter.FamilyContinuation},
		"unknown horizon": {Market: strategyrouter.MarketKR, Horizon: "LONG", Family: strategyrouter.FamilyContinuation},
	} {
		if grant := c.TryAcquireStrategy(key, scope, PollEntry, at); grant.Allowed || grant.Refusal != StrategyBudgetDeferred {
			t.Fatalf("%s: grant=%+v, want deferred", name, grant)
		}
	}
	if grant := c.TryAcquireStrategy(key, a112Scope(strategyrouter.FamilyContinuation), PollProtection, at); grant.Allowed ||
		grant.Refusal != StrategyBudgetDeferred {
		t.Fatalf("safety class through the strategy API: grant=%+v, want deferred", grant)
	}
	if total, _ := a112CommitmentCount(c, key); total != 0 {
		t.Fatalf("refused scopes left %d commitment(s)", total)
	}
}

// Q4 핀(K19 식): 새 두 메서드의 **생산 호출자는 0** 이다 — 레인 evidence polling 배선은 7.5 또는 활성화 로트가 이 시험을 의도적으로 뒤집으며 한다.
func TestTheStrategyBudgetAPIHasNoProductionCallerYet(t *testing.T) {
	root := filepath.Join("..", "..")
	var callers []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "openspec", ".sdd", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok && (selector.Sel.Name == "TryAcquireStrategy" || selector.Sel.Name == "CompleteStrategy") {
				callers = append(callers, path)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(callers) != 0 {
		t.Fatalf("production callers of the strategy budget API: %v — wiring belongs to a deliberate lot (7.5 or activation) that flips this pin", callers)
	}
}
