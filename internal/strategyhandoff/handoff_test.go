package strategyhandoff

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

func selection(identity string) strategyflow.Result {
	return strategyflow.Result{Lineage: strategyflow.Lineage{Identity: identity, Symbol: identity}}
}

// 닫힌 시장에서는 항목이 남아 있어도 아무것도 건너가지 않는다.
func TestAClosedMarketHandsOffNothingEvenWhenASelectionIsStillAttached(t *testing.T) {
	handoff := Admit(false, []strategyflow.Result{selection("005930")})
	if _, ok := handoff.Single(); ok {
		t.Fatal("a closed market handed something off")
	}
	if handoff.Refusal() != MarketClosed {
		t.Fatalf("refusal=%q, want %q", handoff.Refusal(), MarketClosed)
	}
	if handoff.Pending() != 1 {
		t.Fatalf("pending=%d, want the attached selection counted", handoff.Pending())
	}
}

// 고른 것이 없는 시장은 "없음"이라고 말한다. 상한에 걸린 것과 같은 이름을
// 쓰면 후보가 없던 날과 상한이 막은 날이 한 글자로 뭉쳐진다.
func TestAMarketThatSelectedNothingSaysSoInsteadOfBorrowingTheCapacityName(t *testing.T) {
	handoff := Admit(true, nil)
	if _, ok := handoff.Single(); ok {
		t.Fatal("an empty market handed something off")
	}
	if handoff.Refusal() != NoSelection || handoff.Pending() != 0 {
		t.Fatalf("refusal=%q pending=%d, want an empty-selection refusal", handoff.Refusal(), handoff.Pending())
	}
}

// 조정자가 하나만 고른 시장은 그 하나가 그대로 건너간다.
func TestASingleSelectedScopeIsTheValueThatCrossesTheSeam(t *testing.T) {
	handoff := Admit(true, []strategyflow.Result{selection("005930")})
	result, ok := handoff.Single()
	if !ok {
		t.Fatalf("a single selection was refused: %q", handoff.Refusal())
	}
	if result.Lineage.Identity != "005930" {
		t.Fatalf("handed off %q, want 005930", result.Lineage.Identity)
	}
	if handoff.Refusal() != Admitted || handoff.Pending() != 1 {
		t.Fatalf("refusal=%q pending=%d, want an admitted single selection", handoff.Refusal(), handoff.Pending())
	}
}

// 종목이 둘인 시장은 지금도 아무것도 내보내지 않는다. 달라지는 것은 그 이유를
// 말한다는 점이다. 이 시장은 고장난 시장이 아니라 조정자가 소유자 범위마다
// 하나씩 골라 둘을 돌려준, 계약대로 동작한 결과다.
func TestTwoSelectedScopesAreRefusedByNameInsteadOfSilently(t *testing.T) {
	handoff := Admit(true, []strategyflow.Result{selection("005930"), selection("000660")})
	if _, ok := handoff.Single(); ok {
		t.Fatal("a two-scope market handed something off")
	}
	if handoff.Refusal() != OverCapacity || handoff.Pending() != 2 {
		t.Fatalf("refusal=%q pending=%d, want the capacity refusal naming both", handoff.Refusal(), handoff.Pending())
	}
}

// 이 시험이 이 파일에서 가장 중요하다.
//
// 첫 판본은 상한을 넘지 않으면 무조건 selected[0] 을 건넸다. 그래서 Capacity
// 를 2 로 올리는 것만으로 종목이 둘인 시장이 **승인되면서 두 번째 선택을
// 조용히 버렸다** — 거절 이름도 계수기도 없이. 없애겠다고 선언한 바로 그
// 결함을 seam 안에 새로 만든 것이었다.
//
// 지금은 건네줄 수 있는 것보다 많이 실린 handoff 는 값을 내주지 않는다.
// 상한을 올린 날 dispatch 는 하나를 내보내는 대신 아무것도 내보내지 않고,
// 그것이 다중 dispatch 를 구현하라는 신호다.
//
// 이 상태는 Admit 이 만들지 못한다(Capacity 가 1 이므로). 그래서 여기서
// 직접 만든다 — 만들 수 없다고 검사를 빼면, 상한을 올린 사람이 그 순간
// 아무 시험도 깨지지 않는 것을 보게 된다.
func TestARaisedCapacityHandsOffNothingRatherThanDroppingTheSecondSelection(t *testing.T) {
	overCarried := Handoff{selected: []strategyflow.Result{selection("005930"), selection("000660")}, pending: 2}
	if overCarried.Refusal() != Admitted {
		t.Fatalf("refusal=%q, want the admitted state this test rests on", overCarried.Refusal())
	}
	result, ok := overCarried.Single()
	if ok {
		t.Fatalf("Single() handed off %q while a second selection was still on board", result.Lineage.Identity)
	}
	if result.Lineage.Identity != "" {
		t.Fatalf("a refused Single() still carried %q", result.Lineage.Identity)
	}
}

// 거절된 handoff 는 값을 들고 있지 않다.
func TestARefusedHandoffCarriesNoResult(t *testing.T) {
	for _, handoff := range []Handoff{
		Admit(false, []strategyflow.Result{selection("005930")}),
		Admit(true, nil),
		Admit(true, []strategyflow.Result{selection("005930"), selection("000660")}),
	} {
		result, ok := handoff.Single()
		if ok || result.Lineage.Identity != "" {
			t.Fatalf("refusal %q still carried %q", handoff.Refusal(), result.Lineage.Identity)
		}
	}
}

// 부르는 쪽이 건네준 목록을 나중에 바꿔도 이미 만들어진 handoff 는 바뀌지 않는다.
func TestTheHandoffDoesNotAliasTheCallersSlice(t *testing.T) {
	selected := []strategyflow.Result{selection("005930")}
	handoff := Admit(true, selected)
	selected[0] = selection("000660")
	result, ok := handoff.Single()
	if !ok || result.Lineage.Identity != "005930" {
		t.Fatalf("handed off %q ok=%v, want the value as it was at admission", result.Lineage.Identity, ok)
	}
}
