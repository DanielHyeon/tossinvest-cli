package journal

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
)

// provenance_test.go is task 6.4: "이 포지션은 왜 존재하는가"를 단일 질의로
// 재구성할 수 있어야 하며(SHALL), 그 조인 경로는 스키마의 명시적 참조
// 컬럼이다(SHALL — 시간창 휴리스틱 매칭 금지).
//
// The two halves are tested apart: that the chain comes back complete and in
// time order, and that every link in it is a declared reference rather than a
// coincidence of timing.

// kindsOf lists the step kinds in the order the chain returned them.
func kindsOf(chain PositionProvenance) []string {
	out := make([]string, 0, len(chain.Steps))
	for _, step := range chain.Steps {
		out = append(out, step.Kind)
	}
	return out
}

func stepOf(t *testing.T, chain PositionProvenance, kind string) ProvenanceStep {
	t.Helper()
	for _, step := range chain.Steps {
		if step.Kind == kind {
			return step
		}
	}
	t.Fatalf("the chain has no %s step: %v", kind, kindsOf(chain))
	return ProvenanceStep{}
}

// TestWhyDoesThisPositionExist is the scenario: an OPEN position's provenance
// comes back as the decision, the order, the fill and the position it produced,
// in time order.
func TestWhyDoesThisPositionExist(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	position := currentPosition(t, j, o)

	chain, err := j.PositionProvenance(ctx, position.ID)
	if err != nil {
		t.Fatalf("PositionProvenance: %v", err)
	}
	if chain.PositionID != position.ID {
		t.Fatalf("chain is for %q, want %q", chain.PositionID, position.ID)
	}

	for _, kind := range []string{
		ProvenanceDecision, ProvenanceIntent, ProvenanceAttempt,
		ProvenanceFill, ProvenancePosition,
	} {
		stepOf(t, chain, kind)
	}

	// Time order, which is what makes it a reconstruction rather than a dump.
	if !sort.SliceIsSorted(chain.Steps, func(a, b int) bool {
		return chain.Steps[a].At < chain.Steps[b].At
	}) {
		t.Fatalf("the chain is not in time order: %+v", chain.Steps)
	}

	// The decision is the one the attempt named, and it carries the preimage the
	// gateway verified against.
	decision := stepOf(t, chain, ProvenanceDecision)
	if decision.Ref != "d-1" {
		t.Errorf("decision = %q, want d-1", decision.Ref)
	}
	if !strings.Contains(decision.Detail, SafetyClassExposureRaising) {
		t.Errorf("decision detail = %q, want the safety class", decision.Detail)
	}
	if fill := stepOf(t, chain, ProvenanceFill); fill.Ref != "o-1" {
		t.Errorf("fill = %q, want the broker order o-1", fill.Ref)
	}
}

func TestPreOwnerAndNonConfirmedFillsAreNotPositionProvenance(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()
	if _, err := j.db.ExecContext(ctx, `INSERT INTO fill_events
		(order_id, account_ref, symbol, market, trading_day, side, delta_quantity,
		 cumulative_quantity, average_price, broker_visible_at, committed_at)
		VALUES ('o-real','acct-1','005930','kr','2026-03-30','BUY','7','7','60000','',
		        '2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	real := place(t, j, order{
		intentID: "i-real", attemptID: "a-real", orderID: "o-real", decisionID: "d-real",
	})
	if _, err := j.RecordFill(ctx, terminalFill(real, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	position := currentPosition(t, j, real)
	failed := insertNonConfirmedFillAttempt(t, j, "failed", "o-failed", "d-real")
	if _, err := j.RecordFill(ctx, terminalFill(failed, "4", "99000")); err != nil {
		t.Fatal(err)
	}

	chain, err := j.PositionProvenance(ctx, position.ID)
	if err != nil {
		t.Fatal(err)
	}
	var fills []ProvenanceStep
	for _, step := range chain.Steps {
		if step.Kind == ProvenanceFill {
			fills = append(fills, step)
		}
	}
	if len(fills) != 1 || fills[0].Ref != "o-real" {
		t.Fatalf("fill provenance=%+v, want only the post-owner confirmed fill", fills)
	}
}

func insertNonConfirmedFillAttempt(t *testing.T, j *Journal, id, orderID, decisionID string) order {
	t.Helper()
	o := (order{
		intentID: "i-" + id, attemptID: "a-" + id, orderID: orderID,
		decisionID: decisionID, side: "SELL", quantity: "4",
	}).withDefaults()
	if _, err := j.db.Exec(`INSERT INTO intents
		(id, created_at, market, trading_day, account_ref, symbol, side, order_type,
		 quantity, price, currency, source, fingerprint)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, o.intentID, "2026-03-30T00:30:00Z", o.market,
		o.tradingDay, o.account, o.symbol, o.side, "LIMIT", o.quantity, "70000", "KRW",
		"engine/test", "fp-"+o.intentID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO mutation_attempts
		(id, intent_id, kind, state, attempt_no, broker_order_id, fingerprint,
		 recorded_at, decision_id)
		VALUES (?,?,?,?,?,?,?,?,NULLIF(?,''))`, o.attemptID, o.intentID, "PLACE",
		string(StateFailedConfirmed), 1, o.orderID, "fp-"+o.intentID,
		"2026-03-30T00:30:00Z", decisionID); err != nil {
		t.Fatal(err)
	}
	return o
}

// TestTheProvenanceChainReachesTheClose walks the whole thing: entry, an
// adjustment, an exit judgement, the liquidation it proposed, and the close.
//
// The exit_events row is inserted directly because the producer is task 7.4 and
// does not exist yet. That is exactly what has to be tested now: the join path
// must work before there is anything on it, or 7.4 discovers at the end that the
// chain it was supposed to complete cannot be walked.
//
// Everything here shares one instant, because the test clock is frozen and
// journal stamps are second-resolution anyway. That makes it the test of the
// causal tie-break as well: the order below comes from the rank column, not from
// the timestamps.
func TestTheProvenanceChainReachesTheClose(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	entry := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(entry, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	position := currentPosition(t, j, entry)

	// An adjustment: the ledger disagreeing with the account is part of why the
	// position is the size it is.
	watermark, err := j.FillWatermark(ctx, entry.symbol)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, adjustmentFor(entry, "10", "9", watermark)); err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}

	// The exit judgement, and the liquidation it proposed.
	exit := place(t, j, order{
		intentID: "i-exit", attemptID: "a-exit", orderID: "o-exit", side: "SELL", quantity: "9",
	})
	if _, err := j.db.ExecContext(ctx, `
		INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after,
		   action, proposed_intent_id, created_at)
		VALUES (?,?,?,?,?,?,?,?)`,
		position.ID, "62000", "75000", "69000", RatchetBreakeven,
		"FLATTEN", exit.intentID, "2026-03-30T00:30:00Z"); err != nil {
		t.Fatalf("recording the exit judgement: %v", err)
	}
	if _, err := j.RecordFill(ctx, terminalFill(exit, "9", "62000")); err != nil {
		t.Fatal(err)
	}

	closed := currentPosition(t, j, entry)
	if closed.State != PositionClosed {
		t.Fatalf("precondition: state = %s, want CLOSED", closed.State)
	}

	chain, err := j.PositionProvenance(ctx, position.ID)
	if err != nil {
		t.Fatalf("PositionProvenance: %v", err)
	}
	for _, kind := range []string{
		ProvenanceDecision, ProvenanceIntent, ProvenanceAttempt, ProvenanceFill,
		ProvenancePosition, ProvenanceAdjustment, ProvenanceExitEvent,
		ProvenanceExitIntent, ProvenanceExitAttempt, ProvenanceExitFill, ProvenanceClose,
	} {
		stepOf(t, chain, kind)
	}
	if !sort.SliceIsSorted(chain.Steps, func(a, b int) bool {
		return chain.Steps[a].At < chain.Steps[b].At
	}) {
		t.Fatalf("the chain is not in time order: %+v", chain.Steps)
	}
	if last := chain.Steps[len(chain.Steps)-1]; last.Kind != ProvenanceClose {
		t.Fatalf("the chain ends with %s, want the close", last.Kind)
	}
	if adjustment := stepOf(t, chain, ProvenanceAdjustment); !strings.Contains(adjustment.Detail, "9") {
		t.Errorf("adjustment detail = %q, want the value it converged to", adjustment.Detail)
	}
	if event := stepOf(t, chain, ProvenanceExitEvent); !strings.Contains(event.Detail, "FLATTEN") {
		t.Errorf("exit event detail = %q, want the action it proposed", event.Detail)
	}
}

// TestAnEmptyExitHistoryStillWalks: exit_events has no producer until task 7.4,
// and a join that only works once rows exist is a join nobody has tested.
func TestAnEmptyExitHistoryStillWalks(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	position := currentPosition(t, j, o)

	chain, err := j.PositionProvenance(ctx, position.ID)
	if err != nil {
		t.Fatalf("PositionProvenance: %v", err)
	}
	for _, step := range chain.Steps {
		switch step.Kind {
		case ProvenanceExitEvent, ProvenanceExitIntent, ProvenanceExitAttempt,
			ProvenanceExitFill, ProvenanceClose:
			t.Fatalf("an open position with no exit history produced a %s step: %+v", step.Kind, step)
		}
	}
	if len(chain.Steps) == 0 {
		t.Fatal("the entry half of the chain must still be there")
	}
}

// TestAnExternalPositionHasNoEntryProvenance: `entry_decision_id` NULL is the
// fact that no decision justifies the position, and the chain says so by
// carrying no decision rather than by attaching the nearest one in time. That
// substitution is precisely the 시간창 휴리스틱 the spec forbids.
func TestAnExternalPositionHasNoEntryProvenance(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	// A decision, an order and a fill on one symbol — the near-in-time decoy.
	decoy := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(decoy, "10", "70000")); err != nil {
		t.Fatal(err)
	}

	// An external holding on another symbol, folded in at the same instant.
	watermark, err := j.FillWatermark(ctx, "000660")
	if err != nil {
		t.Fatal(err)
	}
	result, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: decoy.account, Market: decoy.market, Symbol: "000660",
		Kind: AdjustmentExternal, ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark,
		NewQuantity: "5", BrokerAsOf: "2026-03-30T00:30:00Z",
		Evidence: "the account holds it and no local instance explains it",
	})
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}

	chain, err := j.PositionProvenance(ctx, result.Position.ID)
	if err != nil {
		t.Fatalf("PositionProvenance: %v", err)
	}
	for _, step := range chain.Steps {
		switch step.Kind {
		case ProvenanceDecision, ProvenanceIntent, ProvenanceAttempt, ProvenanceFill:
			t.Fatalf("an external position inherited a %s (%s) it has no reference to",
				step.Kind, step.Ref)
		}
	}
	// What it does have is the adjustment that folded it in, which is the honest
	// answer to why it exists.
	stepOf(t, chain, ProvenanceAdjustment)
	stepOf(t, chain, ProvenancePosition)
}

// TestAnotherDecisionsFillsAreNotThisPositionsProvenance: the fill link is
// decision → attempt → broker order → fill_events, all declared columns. A
// second decision filling the same symbol in the same second must not appear.
func TestAnotherDecisionsFillsAreNotThisPositionsProvenance(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	mine := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(mine, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	position := currentPosition(t, j, mine)

	// Somebody else's order on another symbol, same instant, its own decision.
	theirs := place(t, j, order{
		intentID: "i-2", attemptID: "a-2", orderID: "o-2", decisionID: "d-2", symbol: "000660",
	})
	if _, err := j.RecordFill(ctx, terminalFill(theirs, "7", "120000")); err != nil {
		t.Fatal(err)
	}

	chain, err := j.PositionProvenance(ctx, position.ID)
	if err != nil {
		t.Fatalf("PositionProvenance: %v", err)
	}
	for _, step := range chain.Steps {
		if step.Ref == "d-2" || step.Ref == "i-2" || step.Ref == "a-2" || step.Ref == "o-2" {
			t.Fatalf("the chain picked up %s %q from another decision", step.Kind, step.Ref)
		}
	}
}

func TestReusedOrderIDDoesNotAttachPriorDayFillToNewPositionProvenance(t *testing.T) {
	j := projectingJournal(t)
	ctx := context.Background()

	first := place(t, j, order{
		intentID: "i-day-1", attemptID: "a-day-1", orderID: "reused-order",
		decisionID: "d-day-1", tradingDay: "2026-03-30", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(first, "10", "70000")); err != nil {
		t.Fatal(err)
	}
	exit := place(t, j, order{
		intentID: "i-exit-1", attemptID: "a-exit-1", orderID: "exit-day-1",
		decisionID: "d-exit-1", tradingDay: "2026-03-30", side: "SELL", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(exit, "10", "71000")); err != nil {
		t.Fatal(err)
	}

	second := place(t, j, order{
		intentID: "i-day-2", attemptID: "a-day-2", orderID: "reused-order",
		decisionID: "d-day-2", tradingDay: "2026-03-31", quantity: "2",
	})
	if _, err := j.RecordFill(ctx, terminalFill(second, "2", "72000")); err != nil {
		t.Fatal(err)
	}
	position := currentPosition(t, j, second)
	chain, err := j.PositionProvenance(ctx, position.ID)
	if err != nil {
		t.Fatal(err)
	}

	var fills []ProvenanceStep
	for _, step := range chain.Steps {
		if step.Kind == ProvenanceFill {
			fills = append(fills, step)
		}
	}
	if len(fills) != 1 || !strings.Contains(fills[0].Detail, "filled 2 ") {
		t.Fatalf("new position fills=%+v, want only the reused id's later-day fill", fills)
	}
}

// TestProvenanceOfNothingIsNotFound: a chain with no rows is a position that
// does not exist, and saying so is different from returning an empty answer a
// caller reads as "no reason".
func TestProvenanceOfNothingIsNotFound(t *testing.T) {
	j := projectingJournal(t)
	if _, err := j.PositionProvenance(context.Background(), "pos-nope"); !errors.Is(err, ErrPositionNotFound) {
		t.Fatalf("err = %v, want ErrPositionNotFound", err)
	}
}
