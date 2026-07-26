package position

import (
	"errors"
	"testing"
)

// position_test.go is the narrative half of task 6.1. table_test.go proves that
// every row of the transition table does what the table says; these tests
// exercise the sequences the spec names by name — 즉시 전량체결, OPENING 종료
// 판단(원주문 수량), SCALING 진입·종료, lineage 승계, 매도 체결 귀속, CLOSED
// 종결성 — and the cost basis the rows carry with them.

// buy is one BUY observation of an order for `ordered`, whose cumulative filled
// moved from prev to filled at an average of avg.
func buy(ordered, prevFilled, filled, prevAvg, avg, delta string) Event {
	return Event{
		Role: Entry, Delta: delta, OrderQuantity: ordered,
		PrevOrderFilled: prevFilled, OrderFilled: filled,
		PrevOrderAvgPrice: prevAvg, OrderAvgPrice: avg,
	}
}

func sell(ordered, prevFilled, filled, delta string) Event {
	return Event{
		Role: Exit, Delta: delta, OrderQuantity: ordered,
		PrevOrderFilled: prevFilled, OrderFilled: filled,
		PrevOrderAvgPrice: "", OrderAvgPrice: "70000",
	}
}

func apply(t *testing.T, inst Instance, ev Event) Outcome {
	t.Helper()
	out, err := Apply(inst, ev)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return out
}

// TestImmediateFullFillReachesOpenDirectly is the spec's own scenario: 진입
// 주문이 첫 관측에서 전량 체결로 나타나면 전이표에 따라 OPEN에 도달하며 오류가
// 아니다.
func TestImmediateFullFillReachesOpenDirectly(t *testing.T) {
	t.Parallel()

	out := apply(t, Instance{}, buy("10", "0", "10", "", "70000", "10"))
	if out.Row != "E20" {
		t.Fatalf("row = %s, want E20 (FLAT + ADDS + DONE)", out.Row)
	}
	if out.Next != Open {
		t.Errorf("state = %s, want OPEN without passing through OPENING", out.Next)
	}
	if !out.NewInstance {
		t.Error("the first fill on a flat symbol opens a new instance")
	}
	if out.Quantity != "10" || out.AvgPrice != "70000" {
		t.Errorf("instance = (%s, %s), want (10, 70000)", out.Quantity, out.AvgPrice)
	}
}

// TestOpeningCompletesAtTheOrderedQuantity pins 원주문 수량 as the judgement:
// the position stays OPENING while the entry order can still fill and reaches
// OPEN when it cannot.
func TestOpeningCompletesAtTheOrderedQuantity(t *testing.T) {
	t.Parallel()

	first := apply(t, Instance{}, buy("10", "0", "3", "", "70000", "3"))
	if first.Next != Opening || first.Row != "E19" {
		t.Fatalf("first partial = (%s, %s), want OPENING via E19", first.Row, first.Next)
	}
	if first.Quantity != "3" || first.AvgPrice != "70000" {
		t.Fatalf("after the first partial = (%s, %s), want (3, 70000)", first.Quantity, first.AvgPrice)
	}

	// 3 @ 70000 then 7 more @ 71000: the order's average moves to 70700, and the
	// position's basis is the same number because this is its only order.
	second := apply(t,
		Instance{State: first.Next, Quantity: first.Quantity, AvgPrice: first.AvgPrice},
		buy("10", "3", "10", "70000", "70700", "7"))
	if second.Row != "E23" || second.Next != Open {
		t.Fatalf("completion = (%s, %s), want OPEN via E23", second.Row, second.Next)
	}
	if second.Quantity != "10" || second.AvgPrice != "70700" {
		t.Errorf("after completion = (%s, %s), want (10, 70700)", second.Quantity, second.AvgPrice)
	}
}

// TestAnUnfilledRemainderEndsTheOpeningWhenNothingCarriesIt is the other exit
// from OPENING: the order went terminal short of its quantity and no amendment
// took over, so what filled is what the position has.
func TestAnUnfilledRemainderEndsTheOpeningWhenNothingCarriesIt(t *testing.T) {
	t.Parallel()

	ev := buy("10", "3", "3", "70000", "70000", "0")
	ev.Terminal = true
	out := apply(t, Instance{State: Opening, Quantity: "3", AvgPrice: "70000"}, ev)
	if out.Row != "E05" || out.Next != Open {
		t.Fatalf("cancelled remainder = (%s, %s), want OPEN via E05", out.Row, out.Next)
	}
	if out.Quantity != "3" {
		t.Errorf("quantity = %s, want the 3 that filled", out.Quantity)
	}
}

// TestLineageSuccessionKeepsThePositionOpening is 정정 교체 주문의 lineage
// 승계. The parent went terminal at 3 of 10, but a replace edge leaves it, so
// the entry is not over — the child carries the remainder, and calling this OPEN
// would make the child's fills look like a scale-in.
func TestLineageSuccessionKeepsThePositionOpening(t *testing.T) {
	t.Parallel()

	parent := buy("10", "3", "3", "70000", "70000", "0")
	parent.Terminal, parent.HasSuccessor = true, true
	out := apply(t, Instance{State: Opening, Quantity: "3", AvgPrice: "70000"}, parent)
	if out.Row != "E06" || out.Next != Opening {
		t.Fatalf("amended parent = (%s, %s), want OPENING via E06", out.Row, out.Next)
	}

	// The child is a different order for the 7 that remained; completing it
	// completes the original request.
	child := apply(t,
		Instance{State: out.Next, Quantity: out.Quantity, AvgPrice: out.AvgPrice},
		buy("7", "0", "7", "", "71000", "7"))
	if child.Row != "E23" || child.Next != Open {
		t.Fatalf("child completion = (%s, %s), want OPEN via E23", child.Row, child.Next)
	}
	if child.Quantity != "10" {
		t.Errorf("quantity = %s, want the original request of 10", child.Quantity)
	}
	// 3 @ 70000 + 7 @ 71000 = 707000 over 10.
	if child.AvgPrice != "70700" {
		t.Errorf("average = %s, want 70700 across the whole chain", child.AvgPrice)
	}
}

// TestScalingEntersAndEnds is SCALING 진입·종료: a second entry order against an
// open position, and the return to OPEN when it is done.
func TestScalingEntersAndEnds(t *testing.T) {
	t.Parallel()

	scaleIn := apply(t,
		Instance{State: Open, Quantity: "10", AvgPrice: "70000"},
		buy("5", "0", "2", "", "72000", "2"))
	if scaleIn.Row != "E25" || scaleIn.Next != Scaling {
		t.Fatalf("scale-in = (%s, %s), want SCALING via E25", scaleIn.Row, scaleIn.Next)
	}
	// 10 @ 70000 + 2 @ 72000 = 844000 over 12.
	if scaleIn.Quantity != "12" || scaleIn.AvgPrice != "70333.333333333333" {
		t.Errorf("after the scale-in = (%s, %s), want (12, 70333.333333333333)",
			scaleIn.Quantity, scaleIn.AvgPrice)
	}

	done := apply(t,
		Instance{State: scaleIn.Next, Quantity: scaleIn.Quantity, AvgPrice: scaleIn.AvgPrice},
		buy("5", "2", "5", "72000", "72000", "3"))
	if done.Row != "E29" || done.Next != Open {
		t.Fatalf("scale-in completion = (%s, %s), want OPEN via E29", done.Row, done.Next)
	}
	if done.Quantity != "15" {
		t.Errorf("quantity = %s, want 15", done.Quantity)
	}
}

// TestPartialLiquidationThenFullFill is the spec's scenario: CLOSING 상태에서
// 청산 주문의 부분체결이 반영되면 잔여 수량이 감소하고 전량 체결 시 CLOSED로
// 전이한다.
func TestPartialLiquidationThenFullFill(t *testing.T) {
	t.Parallel()

	partial := apply(t,
		Instance{State: Open, Quantity: "10", AvgPrice: "70000"},
		sell("10", "0", "4", "4"))
	if partial.Row != "X22" || partial.Next != Closing {
		t.Fatalf("partial exit = (%s, %s), want CLOSING via X22", partial.Row, partial.Next)
	}
	if partial.Quantity != "6" || partial.AvgPrice != "70000" {
		t.Errorf("after the partial = (%s, %s); a sell realises P&L and does not move the unit cost",
			partial.Quantity, partial.AvgPrice)
	}

	rest := apply(t,
		Instance{State: partial.Next, Quantity: partial.Quantity, AvgPrice: partial.AvgPrice},
		sell("10", "4", "10", "6"))
	if rest.Row != "X41" || rest.Next != Closed {
		t.Fatalf("remainder = (%s, %s), want CLOSED via X41", rest.Row, rest.Next)
	}
	if rest.Quantity != "0" || !rest.Closed() {
		t.Errorf("after the full exit = %s, want a closed instance holding nothing", rest.Quantity)
	}
	if rest.AvgPrice != "70000" {
		t.Errorf("closed average = %s, want the acquisition cost kept for trade_outcomes", rest.AvgPrice)
	}
}

// TestACompletedPartialTakeReturnsToOpen is the ladder's shape: 40% is taken and
// the exit order is finished, so nothing is closing any more but the position is
// still held.
func TestACompletedPartialTakeReturnsToOpen(t *testing.T) {
	t.Parallel()

	out := apply(t,
		Instance{State: Closing, Quantity: "10", AvgPrice: "70000"},
		sell("4", "0", "4", "4"))
	if out.Row != "X29" || out.Next != Open {
		t.Fatalf("completed partial take = (%s, %s), want OPEN via X29", out.Row, out.Next)
	}
	if out.Quantity != "6" {
		t.Errorf("quantity = %s, want 6", out.Quantity)
	}
}

// TestReEntryOpensANewInstance is 청산 후 재진입: CLOSED된 심볼에 새 진입이
// 발생하면 새 position-instance 식별자와 새 진입 결정 참조로 시작한다.
func TestReEntryOpensANewInstance(t *testing.T) {
	t.Parallel()

	out := apply(t,
		Instance{State: Closed, Quantity: "0", AvgPrice: "70000"},
		buy("5", "0", "5", "", "80000", "5"))
	if out.Row != "E35" || out.Next != Open {
		t.Fatalf("re-entry = (%s, %s), want OPEN via E35", out.Row, out.Next)
	}
	if !out.NewInstance {
		t.Fatal("a re-entry starts a new instance; CLOSED is final")
	}
	if out.Quantity != "5" || out.AvgPrice != "80000" {
		t.Errorf("new instance = (%s, %s), want (5, 80000) — the closed instance's basis "+
			"must not price the new one", out.Quantity, out.AvgPrice)
	}
}

// TestClosedIsTerminalForSells is CLOSED 종결성 from the other side, and 매도
// 체결 귀속 when there is nothing to attribute a sell to.
func TestClosedIsTerminalForSells(t *testing.T) {
	t.Parallel()

	onClosed := apply(t, Instance{State: Closed, Quantity: "0", AvgPrice: "70000"}, sell("3", "0", "3", "3"))
	if onClosed.Refusal != RefusalSellOnClosed || onClosed.Row != "X59" { //nolint:staticcheck // X59 = CLOSED/OVERSHOOTS/DONE
		t.Fatalf("sell on a closed instance = (%s, %s), want X59/SELL_ON_CLOSED",
			onClosed.Row, onClosed.Refusal)
	}
	if onClosed.Next != Closed || onClosed.Quantity != "0" {
		t.Errorf("the refusal moved the instance to (%s, %s)", onClosed.Next, onClosed.Quantity)
	}

	onNothing := apply(t, Instance{}, sell("3", "0", "3", "3"))
	if onNothing.Refusal != RefusalUnattributedSell {
		t.Fatalf("sell against nothing = %s, want UNATTRIBUTED_SELL", onNothing.Refusal)
	}
}

// TestAnOversellIsRefusedAndNotClamped is 산식 보정 금지 in its clearest form:
// the projection would have to invent four shares to make the arithmetic work,
// so it refuses and leaves the account to be the authority.
func TestAnOversellIsRefusedAndNotClamped(t *testing.T) {
	t.Parallel()

	out := apply(t, Instance{State: Open, Quantity: "10", AvgPrice: "70000"}, sell("14", "0", "14", "14"))
	if out.Refusal != RefusalOversell || out.Row != "X50" {
		t.Fatalf("oversell = (%s, %s), want X50/OVERSELL", out.Row, out.Refusal)
	}
	if out.Quantity != "10" {
		t.Errorf("quantity = %s; a refused transition must not clamp to zero", out.Quantity)
	}
	if !out.Reconcile() {
		t.Error("an oversell must raise RECONCILE")
	}
}

// TestABuyFillingWhileAnExitIsWorkingIsRefused documents the one refusal that is
// a judgement rather than an arithmetic impossibility.
func TestABuyFillingWhileAnExitIsWorkingIsRefused(t *testing.T) {
	t.Parallel()

	out := apply(t, Instance{State: Closing, Quantity: "6", AvgPrice: "70000"},
		buy("10", "0", "2", "", "71000", "2"))
	if out.Refusal != RefusalEntryWhileClosing || out.Row != "E31" {
		t.Fatalf("buy while closing = (%s, %s), want E31/ENTRY_WHILE_CLOSING", out.Row, out.Refusal)
	}
	if out.Quantity != "6" || out.Next != Closing {
		t.Errorf("the refusal moved the instance to (%s, %s)", out.Next, out.Quantity)
	}
}

// TestACorrectionMovesTheCostBasisAndNotTheQuantity is the delta-0 contract the
// apply hook hands this package (issues.md, task 0.3: 정정과 terminal 전이에서는
// delta 0으로도 hook이 호출된다).
func TestACorrectionMovesTheCostBasisAndNotTheQuantity(t *testing.T) {
	t.Parallel()

	// The broker restates the order's average from 70000 to 70500 at an
	// unchanged cumulative 10.
	out := apply(t,
		Instance{State: Open, Quantity: "10", AvgPrice: "70000"},
		buy("10", "10", "10", "70000", "70500", "0"))
	if out.Row != "E08" || out.Next != Open {
		t.Fatalf("correction = (%s, %s), want OPEN via E08", out.Row, out.Next)
	}
	if out.Quantity != "10" {
		t.Errorf("quantity = %s; a correction restates a price, not a quantity", out.Quantity)
	}
	if out.AvgPrice != "70500" {
		t.Errorf("average = %s, want 70500 — the restatement is the position's cost basis moving",
			out.AvgPrice)
	}
}

// TestACorrectionOnASellDoesNotTouchTheBasis is the same event on the other
// side. Under average-cost accounting a sell never moves the unit cost, so
// restating a sell's price restates realised P&L (task 8.1) and nothing here.
func TestACorrectionOnASellDoesNotTouchTheBasis(t *testing.T) {
	t.Parallel()

	ev := sell("4", "4", "4", "0")
	ev.PrevOrderAvgPrice, ev.OrderAvgPrice = "72000", "72500"
	out := apply(t, Instance{State: Open, Quantity: "6", AvgPrice: "70000"}, ev)
	if out.Quantity != "6" || out.AvgPrice != "70000" {
		t.Errorf("after a sell-side correction = (%s, %s), want (6, 70000)", out.Quantity, out.AvgPrice)
	}
}

// TestAnUnpricedFillPoisonsTheCostBasis is the fail-closed direction for a
// missing average price. Calling it 0 would understate the break-even, which is
// the direction that sells at a loss believing it is flat.
func TestAnUnpricedFillPoisonsTheCostBasis(t *testing.T) {
	t.Parallel()

	unpriced := apply(t, Instance{}, buy("10", "0", "10", "", "", "10"))
	if unpriced.AvgPrice != Unknown {
		t.Fatalf("average = %q, want the unknown marker rather than a made-up zero", unpriced.AvgPrice)
	}
	if unpriced.Quantity != "10" {
		t.Errorf("quantity = %s; the quantity is still known", unpriced.Quantity)
	}

	// And it stays unknown: the missing piece is not recoverable by averaging in
	// the ones that came after it.
	later := apply(t,
		Instance{State: Open, Quantity: "10", AvgPrice: unpriced.AvgPrice},
		buy("5", "0", "5", "", "71000", "5"))
	if later.AvgPrice != Unknown {
		t.Errorf("average = %q, want it to stay unknown for the life of the instance", later.AvgPrice)
	}
	if later.Quantity != "15" {
		t.Errorf("quantity = %s, want 15", later.Quantity)
	}
}

// TestNoOrderedQuantityLeavesTheOrderWorking pins the fallback: completion
// cannot be judged without 원주문 수량, so the order counts as working until the
// broker calls it terminal. The alternative — assuming completion — would close
// an OPENING position that is still filling.
func TestNoOrderedQuantityLeavesTheOrderWorking(t *testing.T) {
	t.Parallel()

	out := apply(t, Instance{}, buy("", "0", "3", "", "70000", "3"))
	if out.Disposition != Working || out.Next != Opening {
		t.Fatalf("unknown ordered quantity = (%s, %s), want WORKING/OPENING", out.Disposition, out.Next)
	}

	terminal := buy("", "3", "3", "70000", "70000", "0")
	terminal.Terminal = true
	ended := apply(t, Instance{State: Opening, Quantity: "3", AvgPrice: "70000"}, terminal)
	if ended.Next != Open {
		t.Errorf("terminal with no ordered quantity = %s, want OPEN", ended.Next)
	}
}

// TestApplyRefusesInputItCannotRead separates a malformed event from a refusal.
// A refusal is a judgement about a real event; this is the absence of one, and
// it must not be recorded as a RECONCILE cause.
func TestApplyRefusesInputItCannotRead(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		inst Instance
		ev   Event
	}{
		"a negative delta": {Instance{State: Open, Quantity: "10"}, buy("10", "0", "10", "", "1", "-1")},
		"a negative held quantity": {
			Instance{State: Open, Quantity: "-1"}, buy("10", "0", "1", "", "1", "1")},
		"a delta that is not a decimal": {
			Instance{State: Open, Quantity: "10"}, buy("10", "0", "1", "", "1", "1e3")},
		"a role that is not a role": {Instance{State: Open, Quantity: "10"}, Event{Role: "HOLD", Delta: "1"}},
		"a state that is not a state": {Instance{State: "HALF_OPEN", Quantity: "10"},
			buy("10", "0", "1", "", "1", "1")},
	}
	for name, tc := range cases {
		out, err := Apply(tc.inst, tc.ev)
		if !errors.Is(err, ErrInvalidEvent) {
			t.Errorf("%s: err = %v, want ErrInvalidEvent", name, err)
		}
		if out.Refusal != RefusalNone {
			t.Errorf("%s: unreadable input became the refusal %q", name, out.Refusal)
		}
	}
}

// TestRoleForSideRefusesAnythingElse pins the direction source. A fill carries
// no side, so a side this build does not issue must stop the projection rather
// than default to one.
func TestRoleForSideRefusesAnythingElse(t *testing.T) {
	t.Parallel()

	for side, want := range map[string]Role{"BUY": Entry, "buy": Entry, "SELL": Exit, "sell": Exit} {
		got, err := RoleForSide(side)
		if err != nil || got != want {
			t.Errorf("RoleForSide(%q) = (%s, %v), want %s", side, got, err, want)
		}
	}
	for _, side := range []string{"", "SHORT", "BUY_TO_COVER", "b u y"} {
		if _, err := RoleForSide(side); !errors.Is(err, ErrInvalidEvent) {
			t.Errorf("RoleForSide(%q) must refuse", side)
		}
	}
}
