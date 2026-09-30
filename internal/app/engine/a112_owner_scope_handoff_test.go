//go:build tossos_testseams

package engine

// a112 태스크 5.2.2.1 — 활성화된 시장의 주문 경로는 소유자 범위마다 handoff 를 받는다(동결 골든 "at most one selected
// proposal per owner scope"). 활성화가 없는 시장(오늘 생산 — 배포 핀 0)은 시장 단위 상한 그대로다(토글 OFF = upstream).
//
// **오늘-동등성 핀**(Manager 조건 1)은 5.2.2.2 에서 의도대로 뒤집혔다: 5.2.2.1 판은 「두 범위면 하류 1차 레그 개수 관문이 거절해
// 주문 0」이었고, 개수 관문을 걷은 지금은 「범위마다 자기 권한이 있는 범위만 거래, 나머지는 범위 거절 타입」이다
// (TestTwoOwnerScopesTradeOnlyWhereEachHasItsOwnAuthority).
//
// **이 fixture 는 생산 모양이 아니다 — 의도적으로 만든 최악 조건이다**(2026-09-30 리뷰 보이스 A #1 · codex 정정). 위험 ·
// 계좌 권한은 범위 하나짜리 fixture 에서 온 것이라 둘째 범위(000660)의 범위별 권한이 없다. 생산 조립은 범위마다 권한을 다시
// 모으므로(5.2.2.2) 이 모양은 「한 범위의 위험 · 계좌 적재가 실패한 파도」와 같다 — 그 범위만 거절되고 나머지는 거래한다(J3).
// 두 순서(fixture 순서 · 조정자 사전순 000660 먼저)를 모두 못 박는다 — 거절된 범위가 앞이어도 뒤 범위가 굶지 않음.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112TwoScopeKR 는 KR 제안 권한에 소유자 범위를 하나 더 얹음(다른 종목 000660 — 같은 계좌·시장·세대, 봉인된 유효 제안).
// 원래 항목(005930)은 그대로 앞에 둠(fixture 순서). now 는 조립의 관측 시각(유효 창의 기준).
func a112TwoScopeKR(t *testing.T, authority strategyProposalMarketAuthority, now time.Time) strategyProposalMarketAuthority {
	t.Helper()
	return a112ExtraEntryKR(t, authority, now, "000660", false)
}

// a112ExtraEntryKR 는 KR 제안 권한에 봉인된 유효 제안 하나를 더 얹음. first 면 앞에(조정자 사전순을 흉내), 아니면 뒤에.
// symbol 이 원래 항목과 같으면 **같은 소유자 범위**의 둘째 제안이 된다(계좌 · 시장 · 세대가 같으므로).
func a112ExtraEntryKR(t *testing.T, authority strategyProposalMarketAuthority, now time.Time, symbol string, first bool,
) strategyProposalMarketAuthority {
	t.Helper()
	second, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, StrategyMarketKR), "acct-risk-loader", symbol,
		"campaign-risk-loader-kr-extra-"+symbol, 8, "100", "95", "120", now.Add(-time.Second), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:proposal-two-scope",
		map[string]strategyflow.Result{second.Lineage.Symbol: second})
	sealed, ok := batch.For(second.Lineage.Symbol)
	if !ok {
		t.Fatal("arrangement: the second-scope proposal could not be sealed")
	}
	extra := strategyProposalEntryAuthority{authority: sealed}
	if first {
		authority.entries = append([]strategyProposalEntryAuthority{extra}, authority.entries...)
	} else {
		authority.entries = append(append([]strategyProposalEntryAuthority(nil), authority.entries...), extra)
	}
	// 조립이 중재 때 적는 제안 집합 digest 를 같은 식으로 다시 적는다(6.2 A-lite 가독 계약 — 조립과 같은 모양).
	authority.snapshot.ProposalSetDigest = strategyProposalSetDigest(authority.entries)
	return authority
}

// 같은 소유자 범위의 봉인된 제안 둘(보이스 C #2) — 활성화된 시장에서도 범위당 상한 초과로 시장 전체 OverCapacity 하나.
// 앞 판본은 이것을 strategyhandoff 단위 시험으로만 잡았고, 엔진이 `AdmitEachOwnerScope` 를 원소마다 따로 부르는 편집
// (언급은 한 번 · 호출은 여럿 — 중복 검사 무력화)을 아무 시험도 못 봤다. 5.2.2.2 가 1차 레그 개수 관문을 걷어 내면 이 경로가
// 곧바로 주문으로 이어지므로 5.2.2.2 착수 조건에도 적혀 있다.
func TestTheSameOwnerScopeSealedTwiceRefusesTheActivatedMarket(t *testing.T) {
	_, proposals, _, _ := pairedStrategyDispatchCycleFixture(t)
	original := proposals.kr.entries[0].authority.Proposal().Lineage.Symbol
	twice := a112ExtraEntryKR(t, proposals.kr, proposals.observedAt, original, false)
	twice.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	handoffs := twice.dispatchHandoffs()
	if len(handoffs) != 1 || handoffs[0].Refusal() != strategyhandoff.OverCapacity || handoffs[0].Pending() != 2 {
		t.Fatalf("one owner scope sealed twice: %d handoffs, first refusal %q — want one OverCapacity for the whole market",
			len(handoffs), handoffs[0].Refusal())
	}
}

func TestOnlyAnActivatedMarketHandsEachOwnerScopeOff(t *testing.T) {
	_, proposals, _, _ := pairedStrategyDispatchCycleFixture(t)
	two := a112TwoScopeKR(t, proposals.kr, proposals.observedAt)

	// 활성화 없음 — 오늘의 시장 단위 상한(이름 붙은 거절).
	today := two.dispatchHandoffs()
	if len(today) != 1 || today[0].Refusal() != strategyhandoff.OverCapacity {
		t.Fatalf("without an activation: %d handoffs, first refusal %q — want the market-wide capacity refusal (toggle OFF = upstream)",
			len(today), today[0].Refusal())
	}

	// 활성화 있음 — 소유자 범위마다 하나.
	two.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	if !two.activation.Verified() {
		t.Fatal("arrangement: the test-seam activation is not verified")
	}
	scoped := two.dispatchHandoffs()
	if len(scoped) != 2 {
		t.Fatalf("with an activation: %d handoffs, want one per owner scope", len(scoped))
	}
	for i, handoff := range scoped {
		if _, ok := handoff.Single(); !ok {
			t.Errorf("scope %d refused %q", i, handoff.Refusal())
		}
	}

	// 활성화돼 있어도 준비되지 않은 시장은 범위로 쪼개지 않고 닫힘 하나 — 닫힌 시장에서 범위가 새지 않음.
	closed := two
	closed.snapshot.Ready = false
	if refused := closed.dispatchHandoffs(); len(refused) != 1 || refused[0].Refusal() != strategyhandoff.MarketClosed {
		t.Fatalf("activated but not ready: %d handoffs, first refusal %q — want one MarketClosed", len(refused), refused[0].Refusal())
	}
}

// 5.2.2.2 로 뒤집은 오늘-동등성 핀(5.2.2.1 판: 「두 범위면 주문 0 — 1차 레그 개수 관문」). 개수 관문을 걷은 뒤, 이 fixture 의
// 최악 조건(위험 · 계좌 권한이 원래 범위 하나에만 있음)에서 **권한을 가진 범위만** 거래하고 다른 범위는 그 범위만의 **타입 거절**로
// 건너뛰어진다(조용히가 아니라 반환 오류에 기록 — J3). 두 순서 모두: 거절된 범위가 앞이어도 뒤 범위가 굶지 않는다.
func TestTwoOwnerScopesTradeOnlyWhereEachHasItsOwnAuthority(t *testing.T) {
	for _, order := range []struct {
		name  string
		first bool
	}{{"fixture order (original scope first)", false}, {"coordinator order (000660 first)", true}} {
		t.Run(order.name, func(t *testing.T) { a112TwoScopePin(t, order.first) })
	}
}

func a112TwoScopePin(t *testing.T, coordinatorOrder bool) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
	original := proposals.kr.entries[0].authority.Proposal().Lineage.Symbol
	two := a112ExtraEntryKR(t, proposals.kr, proposals.observedAt, "000660", coordinatorOrder)
	two.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	// dispatch 와 1차 레그 권한이 같은 제안 쌍을 본다(위험 · 계좌 권한은 범위 하나짜리 fixture 그대로 — 머리말의 최악 조건).
	cycle.proposals.kr = two
	loader, ok := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	if !ok {
		t.Fatalf("arrangement: first-leg loader is %T", cycle.firstLeg.loader)
	}
	loader.proposals.kr = two

	handoffs := two.dispatchHandoffs()
	if len(handoffs) != 2 {
		t.Fatalf("arrangement: %d handoffs, want both owner scopes across the seam", len(handoffs))
	}
	err := deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {
		_, err := cycle.dispatch(context.Background(), delivered)
		return err
	})
	spy.mu.Lock()
	placed := make([]string, 0, len(spy.calls))
	for _, call := range spy.calls {
		placed = append(placed, call.Intent.Symbol)
	}
	spy.mu.Unlock()
	if strings.Join(placed, ",") != original {
		t.Fatalf("placed=%v err=%v — want only the scope that holds its own risk and account authority (%s)", placed, err, original)
	}
	var refusal *strategyScopeRefusal
	if !errors.As(err, &refusal) || refusal.scope.Symbol != "000660" || !strings.Contains(refusal.detail, "risk authority") {
		t.Fatalf("err=%v — want the other scope's typed refusal (its missing risk authority) returned, not swallowed", err)
	}
}

// 주문 경로의 반복: 승인된 소유자 범위는 조정자 순서대로 **전부** 몸통에 건너감.
func TestEveryAdmittedOwnerScopeReachesTheDeliveryBody(t *testing.T) {
	selected := []strategyflow.Result{
		{Lineage: strategyflow.Lineage{Identity: "a", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930", PositionGeneration: 1}},
		{Lineage: strategyflow.Lineage{Identity: "b", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "000660", PositionGeneration: 1}},
	}
	var seen []string
	err := deliverEachStrategyHandoff(strategyhandoff.AdmitEachOwnerScope(true, selected), func(delivered strategyhandoff.Delivered) error {
		seen = append(seen, delivered.Result().Lineage.Identity)
		return nil
	})
	if err != nil || strings.Join(seen, ",") != "a,b" {
		t.Fatalf("delivered=%v err=%v, want a,b in coordinator order", seen, err)
	}
}

// 한 범위의 몸통 오류는 같은 주기의 뒤 범위를 내보내지 않고 그대로 올라감(보수 방향 — 예상 밖 오류 뒤 추가 주문 없음).
// 그리고 handoff 가 하나인 오늘 경로의 오류는 감싸지 않은 **같은 값**임(오류 분류 보존).
func TestAScopeFaultStopsTheCycleBeforeTheNextScope(t *testing.T) {
	selected := []strategyflow.Result{
		{Lineage: strategyflow.Lineage{Identity: "a", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930", PositionGeneration: 1}},
		{Lineage: strategyflow.Lineage{Identity: "b", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "000660", PositionGeneration: 1}},
	}
	fault := errors.New("journal read failed")
	var seen []string
	err := deliverEachStrategyHandoff(strategyhandoff.AdmitEachOwnerScope(true, selected), func(delivered strategyhandoff.Delivered) error {
		seen = append(seen, delivered.Result().Lineage.Identity)
		return fault
	})
	if err != fault || strings.Join(seen, ",") != "a" {
		t.Fatalf("delivered=%v err=%v, want only a and the unwrapped fault", seen, err)
	}
	if err := deliverEachStrategyHandoff([]strategyhandoff.Handoff{strategyhandoff.Admit(true, selected[:1])},
		func(strategyhandoff.Delivered) error { return fault }); err != fault {
		t.Fatalf("single handoff err=%v, want the body's own error value", err)
	}
}
