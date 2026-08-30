package strategyhandoff

import (
	"errors"
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
	// Admit 이 승인으로 만든 상태다(내부 refusal 은 빈 값). 밖으로 나가는
	// 이름은 OverCarried 이고, 그 두 답이 갈라지지 않는 것을
	// TestAnAdmittedButUndeliverableHandoffHasAName 이 지킨다.
	if overCarried.refusal != Admitted {
		t.Fatalf("내부 refusal=%q, want the admitted state this test rests on", overCarried.refusal)
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

// Deliver 는 승인됐을 때만 몸통을 부른다. 이 서명이 있는 이유는 하나다:
// 부르는 쪽에 **무시할 boolean 을 주지 않기 위해서**.
//
// 적대 리뷰가 `Single()` 판본에서 이것을 뚫었다. 답을 받아 놓고
// `if !handedOff { }` 처럼 아무것도 안 하는 조건에 넣으면, 답을 "썼다"는
// 검사는 통과하고 관문은 사라진다. 두 스위트가 모두 초록이었다.
// Deliver 에는 그런 편집이 없다 — 몸통이 도는지 마는지를 부르는 쪽이 정하지 않는다.
func TestDeliverRunsTheBodyOnlyWhenSomethingCrossedTheSeam(t *testing.T) {
	for _, tc := range []struct {
		name    string
		handoff Handoff
		wantRun bool
	}{
		{"닫힌 시장", Admit(false, []strategyflow.Result{selection("005930")}), false},
		{"고른 것 없음", Admit(true, nil), false},
		{"상한 초과", Admit(true, []strategyflow.Result{selection("005930"), selection("000660")}), false},
		{"승인", Admit(true, []strategyflow.Result{selection("005930")}), true},
		{"승인됐으나 과적재", Handoff{selected: []strategyflow.Result{selection("005930"), selection("000660")}, pending: 2}, false},
	} {
		ran := 0
		got := tc.handoff.Deliver(func(result strategyflow.Result) error {
			ran++
			if result.Lineage.Identity != "005930" {
				t.Fatalf("%s: 몸통이 %q 를 받았다", tc.name, result.Lineage.Identity)
			}
			return nil
		})
		if got != nil {
			t.Fatalf("%s: Deliver=%v, want nil", tc.name, got)
		}
		if want := map[bool]int{true: 1, false: 0}[tc.wantRun]; ran != want {
			t.Fatalf("%s: 몸통이 %d 번 돌았다, want %d", tc.name, ran, want)
		}
	}
}

// 거절은 오류가 아니다. 몸통의 오류만 밖으로 나간다.
func TestDeliverReturnsOnlyTheBodysError(t *testing.T) {
	boom := errors.New("dispatch failed")
	if got := Admit(true, []strategyflow.Result{selection("005930")}).Deliver(
		func(strategyflow.Result) error { return boom }); !errors.Is(got, boom) {
		t.Fatalf("Deliver=%v, want the body's error", got)
	}
	if got := Admit(true, nil).Deliver(func(strategyflow.Result) error { return boom }); got != nil {
		t.Fatalf("거절이 오류로 새어 나왔다: %v", got)
	}
}

// nil 몸통으로 부르는 것은 프로그래밍 오류다. 조용히 성공시키면 "건넸다"와
// "건넬 곳이 없었다"가 같은 답이 된다.
func TestDeliverRefusesANilBody(t *testing.T) {
	if got := Admit(true, []strategyflow.Result{selection("005930")}).Deliver(nil); !errors.Is(got, ErrNoDelivery) {
		t.Fatalf("Deliver(nil)=%v, want %v", got, ErrNoDelivery)
	}
}

// Refusal() 과 Single() 은 같은 술어를 써야 한다.
//
// 첫 수정 판본은 그러지 않았다: 상한을 올리면 Refusal() 은 Admitted 라고
// 하면서 Single() 은 값을 안 줬다. 주문 경로는 fail-closed 였지만 **이름과
// 계수기는 "승인"이라고 보고**했다 — 이 change 가 없애려던 바로 그 무명 상태다.
func TestAnAdmittedButUndeliverableHandoffHasAName(t *testing.T) {
	overCarried := Handoff{selected: []strategyflow.Result{selection("005930"), selection("000660")}, pending: 2}
	if got := overCarried.Refusal(); got != OverCarried {
		t.Fatalf("Refusal()=%q, want %q — 건네주지 못하는 상태에 이름이 없으면 운영자는 그것을 볼 수 없다", got, OverCarried)
	}
	if _, ok := overCarried.Single(); ok {
		t.Fatal("Single() 이 과적재 handoff 에서 값을 내줬다")
	}
	// 두 답이 갈라지지 않는다는 것을 모든 경우에서 확인한다.
	for _, handoff := range []Handoff{
		Admit(false, []strategyflow.Result{selection("005930")}),
		Admit(true, nil),
		Admit(true, []strategyflow.Result{selection("005930")}),
		Admit(true, []strategyflow.Result{selection("005930"), selection("000660")}),
		overCarried,
		{},
	} {
		_, ok := handoff.Single()
		if ok != (handoff.Refusal() == Admitted) {
			t.Fatalf("Refusal()=%q 인데 Single() ok=%v — 두 답이 갈라졌다", handoff.Refusal(), ok)
		}
	}
}

// 거절된 Single 은 영값을 돌려준다. 이 성질이 세 소비자(worker 승격, 결과
// 권한, 읽기 전용 projection)의 두 번째 방벽이다 — 그 셋은 값을 쓰기 전에
// ValidProposal 을 다시 보고, 영값 Result 는 그것을 통과하지 못한다.
//
// 앞선 판본에서 그 방벽은 **우연**이었다("거절이면 값이 비어 있더라"). 여기서
// 값으로 못 박아 우연을 사실로 바꾼다.
func TestARefusedSingleReturnsTheZeroResult(t *testing.T) {
	var zero strategyflow.Result
	for _, handoff := range []Handoff{
		Admit(false, []strategyflow.Result{selection("005930")}),
		Admit(true, nil),
		Admit(true, []strategyflow.Result{selection("005930"), selection("000660")}),
		{selected: []strategyflow.Result{selection("005930"), selection("000660")}, pending: 2},
		{},
	} {
		result, ok := handoff.Single()
		if ok {
			t.Fatalf("%q 가 값을 내줬다", handoff.Refusal())
		}
		if result != zero {
			t.Fatalf("%q 의 거절이 영값이 아닌 %+v 를 돌려줬다", handoff.Refusal(), result)
		}
		if result.ValidProposal() {
			t.Fatalf("%q 의 영값이 ValidProposal 을 통과했다 — 세 소비자의 두 번째 방벽이 사라진다", handoff.Refusal())
		}
	}
}
