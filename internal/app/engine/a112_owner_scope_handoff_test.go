//go:build tossos_testseams

package engine

// a112 태스크 5.2.2.1 — 활성화된 시장의 주문 경로는 소유자 범위마다 handoff 를 받는다(동결 골든 "at most one selected
// proposal per owner scope"). 활성화가 없는 시장(오늘 생산 — 배포 핀 0)은 시장 단위 상한 그대로다(토글 OFF = upstream).
//
// **오늘-동등성 핀**(Manager 조건 1): 소유자 범위가 둘인 활성화 시장이 handoff 를 통과해도 하류 1차 레그 권한이 거절해
// 주문은 0 이다. 이 핀이 5.2.2.2(6.2 봉인 뒤 하류 권한의 소유자 범위 전환) 착수 때 「무엇이 바뀌는가」의 기준선이다.
//
// **이 fixture 는 생산 모양이 아니다 — 의도적으로 만든 최악 조건이다**(2026-09-30 리뷰 보이스 A #1 · codex 정정). 위험 ·
// 계좌 권한은 범위 하나짜리 fixture 에서 온 것이라 범위가 둘이어도 Ready 로 남는다. 생산에서는 권한을 다시 모으므로 그 상태가
// 나올 수 없다 — 두 범위면 결과 권한(`ResultAuthority`)이 OverCapacity 로, 계좌 `collectMarket` B1 이 개수로 먼저 준비 안 됨이
// 되고, 1차 레그 B2 는 개수 · 위험 · 계좌 세 조건이 함께 거짓이다. 그래서 「B2 개수 조건이 유일한 방어」는 **이 fixture 의
// fixture 순서(원래 범위 먼저)에서만** 참이다. 조정자 순서(소유자 범위 사전순 — 000660 먼저)에서는 개수 조건을 지워도
// :221(identity) · :228(위험 범위)이 대신 막는다. 두 순서를 모두 못 박는다.

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

func TestTwoOwnerScopesStillPlaceNothingBecauseTheFirstLegGuardRefuses(t *testing.T) {
	for _, order := range []struct {
		name  string
		first bool
	}{{"fixture order (original scope first)", false}, {"coordinator order (000660 first)", true}} {
		t.Run(order.name, func(t *testing.T) { a112TwoScopePin(t, order.first) })
	}
}

func a112TwoScopePin(t *testing.T, coordinatorOrder bool) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
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
	var errs []string
	for _, handoff := range handoffs {
		if err := handoff.Deliver(func(delivered strategyhandoff.Delivered) error {
			_, err := cycle.dispatch(context.Background(), delivered)
			return err
		}); err != nil {
			errs = append(errs, err.Error())
		}
	}
	// 주문 0.
	if n := len(spy.calls); n != 0 {
		t.Fatalf("a two-scope market placed %d orders — the downstream per-market authority must still refuse (5.2.2.2 is what changes this)", n)
	}
	// 첫 범위를 거절한 것이 1차 레그 권한 B2 의 문구임을 고정. 이 fixture 의 fixture 순서에서는 B2 의 여섯 조건 중 거짓인
	// 것이 개수 하나뿐이다(아래 대조). 조정자 순서에서도 문구는 같고, 개수 조건이 없어도 :221 · :228 이 막는다(머리말).
	if len(errs) == 0 || !strings.Contains(errs[0], a112FirstLegCountGuard) {
		t.Fatalf("first scope error = %v, want the first-leg count guard %q", errs, a112FirstLegCountGuard)
	}
	// 대조: **같은 모양의 새 조립**(fixture 를 한 번 더 만든다 — 같은 고정 시각 · 같은 입력)에 같은 활성화로 원래 범위 **하나**만
	// 실으면 그 문구로 거절되지 않음. :217 의 문구는 여섯 조건(제안 개수 · 위험 · 환율 · 계좌 · 일정 준비 · 활성화)이 공유하고
	// 두 실행 사이에 달라진 것은 제안 개수뿐이므로, 위의 거절이 개수 조건(`len(proposal.entries) != 1`)에서 왔음이 이 대조로
	// 가려짐. 이 대조 절반만을 빨갛게 하는 변이(「B2 가 늘 거절」)가 변이 원장에 있다(F16).
	control, controlProposals, _, controlSpy := pairedStrategyDispatchCycleFixture(t)
	one := controlProposals.kr
	one.activation = two.activation
	control.proposals.kr = one
	control.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader).proposals.kr = one
	single := one.dispatchHandoffs()
	if len(single) != 1 {
		t.Fatalf("control: %d handoffs, want one scope", len(single))
	}
	controlErr := single[0].Deliver(func(delivered strategyhandoff.Delivered) error {
		_, err := control.dispatch(context.Background(), delivered)
		return err
	})
	if controlErr != nil && strings.Contains(controlErr.Error(), a112FirstLegCountGuard) {
		t.Fatalf("control: a single-scope market was refused by the same sentence (%v) — the refusal above is not attributable to the count", controlErr)
	}
	// 대조는 실제로 주문 경로 끝(Gateway 스파이 — 실주문 아님)까지 닿아야 함. 닿지 않으면 위 거절의 귀속이 성립하지 않음.
	controlSpy.mu.Lock()
	placed := len(controlSpy.calls)
	controlSpy.mu.Unlock()
	if controlErr != nil || placed == 0 {
		t.Fatalf("control: err=%v places=%d — a single-scope activated market must reach the gateway spy for the attribution to hold",
			controlErr, placed)
	}
}

// a112FirstLegCountGuard 는 strategy_account_first_leg_authority.go `collectStrategyFirstLegAuthority` 의 B2
// (`len(proposal.entries) != 1 || …`, :217–:219) 가 내는 문구임. 결정 (1) — L6 6.2 봉인 전 유일한 방어 — 의 다섯 줄 중 첫 줄.
const a112FirstLegCountGuard = "paired production authority is incomplete for market"

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
