package journal

// exit_state_test.go is task 7.3: the pending lifecycle as the ledger sees it.
//
// The pure evaluation is internal/exitpolicy's and is tested there. What is
// tested here is everything that only exists because the state is persisted —
// that a proposal armed before a crash is still armed after it, that a refusal
// re-arms its level, that the level cannot be proposed twice, and that the
// cumulative taken fraction moves only inside the fill transaction.
//
// Every crash test is a real reopen of the same file, not a simulated one: the
// claim is about what survives a process ending, and a claim about durability
// checked against an in-memory object is a claim about nothing.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// exitFixture is a journal with the projection and the exit applier both bound,
// which is the wiring an engine runs with.
func exitFixture(t *testing.T) *Journal {
	t.Helper()
	j := openTestJournal(t)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	return j
}

// openedPosition places an entry, fills it, and opens the exit state — the
// sequence that produces a managed position.
func openedPosition(t *testing.T, j *Journal, quantity string) (order, ExitState) {
	t.Helper()
	ctx := context.Background()
	o := place(t, j, order{
		intentID: "i-entry", attemptID: "a-entry", orderID: "o-entry",
		decisionID: "d-entry", side: "BUY", quantity: quantity,
	})
	if _, err := j.RecordFill(ctx, terminalFill(o, quantity, "70000")); err != nil {
		t.Fatalf("RecordFill(entry): %v", err)
	}
	p := currentPosition(t, j, o)
	state, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000",
	})
	if err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	return o, state
}

func exitStateOf(t *testing.T, j *Journal, positionID string) ExitState {
	t.Helper()
	state, err := j.ExitState(context.Background(), positionID)
	if err != nil {
		t.Fatalf("ExitState(%s): %v", positionID, err)
	}
	return state
}

// --- opening ------------------------------------------------------------------

// TestOpeningWritesTheEntryStopAsTheBaseline is D5's first correction reaching
// the row: the position is protected from the instant it exists.
func TestOpeningWritesTheEntryStopAsTheBaseline(t *testing.T) {
	j := exitFixture(t)
	_, state := openedPosition(t, j, "10")

	if state.Baseline != "68000" {
		t.Errorf("baseline = %s, want the entry stop 68000", state.Baseline)
	}
	if state.HighWater != "70000" {
		t.Errorf("high water = %s, want the entry price", state.HighWater)
	}
	if state.InitialRisk != "2000" {
		t.Errorf("initial risk = %s, want entry − stop", state.InitialRisk)
	}
	if state.PolicyKind != ExitPolicyRatchet {
		t.Errorf("policy = %s, want the RATCHET default", state.PolicyKind)
	}
	if state.RatchetLevel != RatchetNone {
		t.Errorf("level = %s, want NONE", state.RatchetLevel)
	}
	if state.ActiveRung != exitpolicy.NoRung {
		t.Errorf("active rung = %d, want none under RATCHET", state.ActiveRung)
	}
	if state.TakenRatioTotal != "0" {
		t.Errorf("taken = %s, want nothing taken", state.TakenRatioTotal)
	}
	if state.Pending() {
		t.Error("a fresh state must hold no proposal")
	}
}

func TestTheLadderPolicyIsChosenExplicitly(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	p := currentPosition(t, j, o)

	state, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	})
	if err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	if state.PolicyKind != ExitPolicyLadder {
		t.Errorf("policy = %s, want LADDER", state.PolicyKind)
	}
	if state.ActiveRung != exitpolicy.NoRung {
		t.Errorf("active rung = %d, want none activated", state.ActiveRung)
	}
}

func TestAThirdPolicyKindIsRefused(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	_, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: "PARKER_VWAP",
		EntryPrice: "70000", InitialStop: "68000",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want a refusal — a position has exactly one of two policies", err)
	}
}

// TestAPositionWithNoEntryDecisionIsNotManaged is exit-policy's second scenario:
// an externally acquired position has no stop to build a baseline from.
func TestAPositionWithNoEntryDecisionIsNotManaged(t *testing.T) {
	j := exitFixture(t)
	insertPosition(t, j, "p-external", nil)

	_, err := j.OpenExitState(context.Background(), ExitStateSeed{
		PositionID: "p-external", EntryPrice: "70000", InitialStop: "68000",
	})
	if !errors.Is(err, ErrPositionNotExitEligible) {
		t.Fatalf("err = %v, want ErrPositionNotExitEligible", err)
	}
}

func TestASecondExitStateIsRefused(t *testing.T) {
	j := exitFixture(t)
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	_, err := j.OpenExitState(context.Background(), ExitStateSeed{
		PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000",
	})
	if !errors.Is(err, ErrExitStateExists) {
		t.Fatalf("err = %v, want ErrExitStateExists; two baselines for one position is the thing forbidden", err)
	}
}

func TestAStopAtOrAboveEntryIsRefused(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	p := currentPosition(t, j, o)

	_, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, EntryPrice: "70000", InitialStop: "70000",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want a refusal; the risk per unit does not exist", err)
	}
}

// --- judgements ---------------------------------------------------------------

func TestAJudgementAdvancesTheStateAndAppendsAnEvent(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "70800", HighWater: "70800",
		Baseline: "69000", RatchetLevel: RatchetHalfRisk, ActiveRung: exitpolicy.NoRung,
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}

	state := exitStateOf(t, j, p.ID)
	if state.Baseline != "69000" || state.HighWater != "70800" || state.RatchetLevel != RatchetHalfRisk {
		t.Errorf("state = %+v, want the judgement's values", state)
	}
	events, err := j.ExitEvents(ctx, p.ID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("%d events, want the open and the judgement", len(events))
	}
	if events[0].Action != ExitEventOpened {
		t.Errorf("first event = %s, want OPENED", events[0].Action)
	}
	last := events[len(events)-1]
	if last.ObservedPrice != "70800" || last.BaselineAfter != "69000" || last.LevelAfter != RatchetHalfRisk {
		t.Errorf("event = %+v, want the judgement recorded", last)
	}
}

// TestADescendingBaselineIsRefusedByTheLedgerToo is defence in depth. The policy
// package guarantees monotonicity, but the column is written by whatever exists
// in five years, not by whatever exists today.
func TestADescendingBaselineIsRefusedByTheLedgerToo(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, HighWater: "70800", Baseline: "69000",
		RatchetLevel: RatchetHalfRisk, ActiveRung: exitpolicy.NoRung,
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}

	err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, HighWater: "70800", Baseline: "68500",
		RatchetLevel: RatchetHalfRisk, ActiveRung: exitpolicy.NoRung,
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "monotone") {
		t.Errorf("err = %v, want it to name the invariant", err)
	}
	if got := exitStateOf(t, j, p.ID); got.Baseline != "69000" {
		t.Errorf("baseline = %s, want the refused write to have changed nothing", got.Baseline)
	}
}

func TestADescendingWatermarkIsRefused(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, HighWater: "69000", Baseline: "68000",
		RatchetLevel: RatchetNone, ActiveRung: exitpolicy.NoRung,
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want a refusal — the watermark is the maximum of the samples", err)
	}
}

// --- arming, once per level ---------------------------------------------------

func TestArmingRecordsTheProposalBeforeItIsSubmitted(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "72000", HighWater: "72000",
		Baseline: "70120", RatchetLevel: RatchetBreakeven, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionRatchetPartial), Level: RatchetBreakeven,
		},
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}

	state := exitStateOf(t, j, p.ID)
	if !state.Pending() {
		t.Fatal("the proposal was not armed")
	}
	if state.PendingAction != string(exitpolicy.ActionRatchetPartial) {
		t.Errorf("action = %s", state.PendingAction)
	}
	if state.PendingLevel != RatchetBreakeven {
		t.Errorf("level = %s, want the proposal's identity", state.PendingLevel)
	}
	// Armed with no intent yet: that is the ordering the crash contract needs.
	if state.PendingIntentID != "" {
		t.Errorf("intent = %s, want it attached later", state.PendingIntentID)
	}
}

func TestASecondProposalIsRefusedWhileOneIsOutstanding(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	armPartial(t, j, p.ID, "72000", RatchetBreakeven)

	err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "72400", HighWater: "72400",
		Baseline: "70600", RatchetLevel: RatchetPartialLock, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionRatchetPartial), Level: RatchetPartialLock,
		},
	})
	if !errors.Is(err, ErrProposalPending) {
		t.Fatalf("err = %v, want ErrProposalPending", err)
	}
	// The whole judgement rolled back, including the state advance: a proposal
	// and the state it was computed from are one write.
	if got := exitStateOf(t, j, p.ID); got.Baseline != "70120" {
		t.Errorf("baseline = %s, want the refused transaction to have changed nothing", got.Baseline)
	}
}

func TestAStateOnlyPromotionCannotBeArmed(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, HighWater: "70800", Baseline: "69000",
		RatchetLevel: RatchetHalfRisk, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionLadderHoldStopPromoted), Level: "0",
		},
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want a refusal; nothing would ever fill to resolve it", err)
	}
}

func TestTheIntentIsAttachedAfterItIsMinted(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	armPartial(t, j, p.ID, "72000", RatchetBreakeven)

	if err := j.AttachExitIntent(ctx, p.ID, "i-exit"); err != nil {
		t.Fatalf("AttachExitIntent: %v", err)
	}
	if got := exitStateOf(t, j, p.ID); got.PendingIntentID != "i-exit" {
		t.Errorf("intent = %s, want i-exit", got.PendingIntentID)
	}
	// Idempotent for the same intent, refused for a different one.
	if err := j.AttachExitIntent(ctx, p.ID, "i-exit"); err != nil {
		t.Errorf("re-attaching the same intent: %v", err)
	}
	if err := j.AttachExitIntent(ctx, p.ID, "i-other"); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("err = %v, want a refusal; two intents for one proposal is two orders", err)
	}
}

func TestAttachingWithNothingArmedIsRefused(t *testing.T) {
	j := exitFixture(t)
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	if err := j.AttachExitIntent(context.Background(), p.ID, "i-exit"); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

// --- refusal and cancellation re-arm ------------------------------------------

func TestARefusalReArmsTheLevel(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	armPartial(t, j, p.ID, "72000", RatchetBreakeven)

	if err := j.ResolveExitProposal(ctx, p.ID, ProposalRefused); err != nil {
		t.Fatalf("ResolveExitProposal: %v", err)
	}
	state := exitStateOf(t, j, p.ID)
	if state.Pending() {
		t.Fatal("the proposal is still armed after being refused")
	}
	if state.TakenRatioTotal != "0" {
		t.Errorf("taken = %s, want nothing taken by a refusal", state.TakenRatioTotal)
	}
	// The level is proposable again.
	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "72000", HighWater: "72000",
		Baseline: "70120", RatchetLevel: RatchetBreakeven, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionRatchetPartial), Level: RatchetBreakeven,
		},
	}); err != nil {
		t.Fatalf("re-proposing after a refusal: %v", err)
	}

	events, err := j.ExitEvents(ctx, p.ID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	var refused bool
	for _, e := range events {
		if e.Action == ExitEventProposalRefused {
			refused = true
		}
	}
	if !refused {
		t.Error("the refusal left no trace in the judgement history")
	}
}

func TestResolvingNothingIsNotAnError(t *testing.T) {
	j := exitFixture(t)
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	if err := j.ResolveExitProposal(context.Background(), p.ID, ProposalCancelled); err != nil {
		t.Fatalf("ResolveExitProposal on an empty state: %v; a retry must converge", err)
	}
}

// TestACancelledRungIsProposableAgain is the ladder half of the re-arm. The rung
// index is what that policy de-duplicates on, so it has to come back — while the
// baseline the rung set does not, because protection granted is not withdrawn.
func TestACancelledRungIsProposableAgain(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	p := currentPosition(t, j, o)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: p.ID, PolicyKind: ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}

	if err := j.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: p.ID, ObservedPrice: "71750", HighWater: "71750",
		Baseline: "70700", RatchetLevel: RatchetNone, ActiveRung: 1,
		Proposal: &ExitProposal{Action: string(exitpolicy.ActionLadderPartial), Level: "1"},
	}); err != nil {
		t.Fatalf("RecordExitJudgement: %v", err)
	}
	if got := exitStateOf(t, j, p.ID); got.ActiveRung != 1 {
		t.Fatalf("active rung = %d, want 1", got.ActiveRung)
	}

	if err := j.ResolveExitProposal(ctx, p.ID, ProposalCancelled); err != nil {
		t.Fatalf("ResolveExitProposal: %v", err)
	}
	state := exitStateOf(t, j, p.ID)
	if state.ActiveRung != 0 {
		t.Errorf("active rung = %d, want it rolled back to 0 so rung 1 can be proposed again",
			state.ActiveRung)
	}
	if state.Baseline != "70700" {
		t.Errorf("baseline = %s, want the lock the rung granted to stay", state.Baseline)
	}
}

// --- the fill path ------------------------------------------------------------

// TestAPartialSellMovesTheTakenFractionAndResolvesTheProposal is the whole point
// of the apply hook: both moves land in the fill's own commit.
func TestAPartialSellMovesTheTakenFractionAndResolvesTheProposal(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	armPartial(t, j, p.ID, "72000", RatchetBreakeven)

	sell := place(t, j, order{
		intentID: "i-exit", attemptID: "a-exit", orderID: "o-exit",
		side: "SELL", quantity: "4",
	})
	if err := j.AttachExitIntent(ctx, p.ID, sell.intentID); err != nil {
		t.Fatalf("AttachExitIntent: %v", err)
	}

	if _, err := j.RecordFill(ctx, terminalFill(sell, "4", "72000")); err != nil {
		t.Fatalf("RecordFill(sell): %v", err)
	}

	state := exitStateOf(t, j, p.ID)
	if state.TakenRatioTotal != "0.4" {
		t.Errorf("taken = %s, want 4 of the initial 10", state.TakenRatioTotal)
	}
	if state.Pending() {
		t.Errorf("the proposal is still armed after its own fill: %+v", state)
	}
	events, err := j.ExitEvents(ctx, p.ID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	last := events[len(events)-1]
	if last.Action != ExitEventProposalFilled || last.ProposedIntentID != "i-exit" {
		t.Errorf("last event = %+v, want the fill recorded against its intent", last)
	}
}

// TestTheCumulativeFractionIsAgainstTheInitialQuantity walks two partials, which
// is the case where the two denominators differ: the second sells 3 of the
// remaining 6, and the cumulative total is 7 of the original 10.
func TestTheCumulativeFractionIsAgainstTheInitialQuantity(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	first := place(t, j, order{
		intentID: "i-e1", attemptID: "a-e1", orderID: "o-e1", side: "SELL", quantity: "4",
	})
	if _, err := j.RecordFill(ctx, terminalFill(first, "4", "72000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if got := exitStateOf(t, j, p.ID); got.TakenRatioTotal != "0.4" {
		t.Fatalf("taken = %s, want 0.4", got.TakenRatioTotal)
	}

	second := place(t, j, order{
		intentID: "i-e2", attemptID: "a-e2", orderID: "o-e2", side: "SELL", quantity: "3",
	})
	if _, err := j.RecordFill(ctx, terminalFill(second, "3", "73000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	if got := exitStateOf(t, j, p.ID); got.TakenRatioTotal != "0.7" {
		t.Errorf("taken = %s, want 0.7 — 3 of the remaining 6 is 30%% of the original",
			got.TakenRatioTotal)
	}
}

// TestABuyFillLeavesTheTakenFractionAlone is the documented limitation. A
// scale-in moves the quantity the fraction is measured against, and this change
// has no rule for that; leaving it alone is the visible choice.
func TestABuyFillLeavesTheTakenFractionAlone(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	sell := place(t, j, order{
		intentID: "i-e1", attemptID: "a-e1", orderID: "o-e1", side: "SELL", quantity: "4",
	})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "4", "72000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	add := place(t, j, order{
		intentID: "i-add", attemptID: "a-add", orderID: "o-add", side: "BUY", quantity: "5",
	})
	if _, err := j.RecordFill(ctx, terminalFill(add, "5", "71000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}

	if got := exitStateOf(t, j, p.ID); got.TakenRatioTotal != "0.4" {
		t.Errorf("taken = %s, want a buy to have moved nothing", got.TakenRatioTotal)
	}
}

// TestANonTerminalFillDoesNotResolveTheProposal — an order that is still working
// has not answered anything, and clearing it would let a second proposal be armed
// on top of a live order.
func TestANonTerminalFillDoesNotResolveTheProposal(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	armPartial(t, j, p.ID, "72000", RatchetBreakeven)

	sell := place(t, j, order{
		intentID: "i-exit", attemptID: "a-exit", orderID: "o-exit", side: "SELL", quantity: "4",
	})
	if err := j.AttachExitIntent(ctx, p.ID, sell.intentID); err != nil {
		t.Fatalf("AttachExitIntent: %v", err)
	}
	if _, err := j.RecordFill(ctx, fillOf(sell, "2", "72000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}

	state := exitStateOf(t, j, p.ID)
	if !state.Pending() {
		t.Error("a working order resolved its own proposal")
	}
	if state.TakenRatioTotal != "0.2" {
		t.Errorf("taken = %s, want the quantity that actually moved", state.TakenRatioTotal)
	}
}

// TestAFillFromAnotherIntentDoesNotResolveTheProposal — the resolution is keyed
// on the proposal's own intent, not on "a sell happened".
func TestAFillFromAnotherIntentDoesNotResolveTheProposal(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)
	armPartial(t, j, p.ID, "72000", RatchetBreakeven)
	if err := j.AttachExitIntent(ctx, p.ID, "i-exit"); err != nil {
		t.Fatalf("AttachExitIntent: %v", err)
	}

	other := place(t, j, order{
		intentID: "i-manual", attemptID: "a-manual", orderID: "o-manual",
		side: "SELL", quantity: "2",
	})
	if _, err := j.RecordFill(ctx, terminalFill(other, "2", "72000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}

	state := exitStateOf(t, j, p.ID)
	if !state.Pending() {
		t.Error("somebody else's fill resolved the proposal")
	}
	if state.TakenRatioTotal != "0.2" {
		t.Errorf("taken = %s, want a manual sale to count against the position", state.TakenRatioTotal)
	}
}

func TestClosingThePositionCompletesTheExitState(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o, _ := openedPosition(t, j, "10")
	p := currentPosition(t, j, o)

	sell := place(t, j, order{
		intentID: "i-exit", attemptID: "a-exit", orderID: "o-exit", side: "SELL", quantity: "10",
	})
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "72000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}

	state := exitStateOf(t, j, p.ID)
	if !state.Completed {
		t.Error("the position closed and its exit state is still running")
	}
	if state.TakenRatioTotal != "1" {
		t.Errorf("taken = %s, want the whole position", state.TakenRatioTotal)
	}
	open, err := j.OpenExitStates(ctx, "acct-1")
	if err != nil {
		t.Fatalf("OpenExitStates: %v", err)
	}
	if len(open) != 0 {
		t.Errorf("%d states still in the working set, want none", len(open))
	}
}

func TestAFillOnAPositionWithNoExitStateIsNotAnError(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	o := place(t, j, order{intentID: "i-1", attemptID: "a-1", orderID: "o-1", decisionID: "d-1"})

	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill: %v; a position the exit policy does not manage is not a fault", err)
	}
}

func TestAnExternalOrderIsIgnoredByTheExitApplier(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()

	if _, err := j.RecordFill(ctx, terminalFill(order{
		orderID: "o-external", symbol: "005930", market: "kr", quantity: "5",
	}, "5", "70000")); err != nil {
		t.Fatalf("RecordFill: %v; an order no local intent claims is a fact, not a fault", err)
	}
}

// --- crash restore, both directions -------------------------------------------

// TestAnArmedProposalSurvivesARestart is exit-policy's "제출 전 크래시"
// scenario, in the direction that prevents a duplicate.
func TestAnArmedProposalSurvivesARestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	var positionID string
	func() {
		j := openTestJournalAt(t, path)
		if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
			t.Fatalf("SetApplyHooks: %v", err)
		}
		o, _ := openedPosition(t, j, "10")
		positionID = currentPosition(t, j, o).ID
		armPartial(t, j, positionID, "72000", RatchetBreakeven)
	}()

	// A new process, the same file.
	restarted := openTestJournalAt(t, path)
	restored, err := restarted.OpenExitStates(ctx, "acct-1")
	if err != nil {
		t.Fatalf("OpenExitStates: %v", err)
	}
	if len(restored) != 1 {
		t.Fatalf("%d exit states restored, want 1", len(restored))
	}
	state := restored[0]
	if !state.Pending() {
		t.Fatal("the armed proposal did not survive the restart; the level would be proposed twice")
	}
	if state.PendingLevel != RatchetBreakeven || state.PendingAction != string(exitpolicy.ActionRatchetPartial) {
		t.Errorf("restored proposal = %+v, want the one that was armed", state)
	}
	// And it still refuses a second one, which is what "no duplicate" means.
	if err := restarted.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: positionID, ObservedPrice: "72000", HighWater: "72000",
		Baseline: "70120", RatchetLevel: RatchetBreakeven, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionRatchetPartial), Level: RatchetBreakeven,
		},
	}); !errors.Is(err, ErrProposalPending) {
		t.Fatalf("err = %v, want the restored proposal to still suppress a second", err)
	}
}

// TestAResolvedProposalDoesNotComeBackAfterARestart is the other direction: the
// state must not restore a proposal that was already answered, and the taken
// fraction it left behind must still suppress the once-per-position partial.
func TestAResolvedProposalDoesNotComeBackAfterARestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	var positionID string
	func() {
		j := openTestJournalAt(t, path)
		if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
			t.Fatalf("SetApplyHooks: %v", err)
		}
		o, _ := openedPosition(t, j, "10")
		positionID = currentPosition(t, j, o).ID
		armPartial(t, j, positionID, "72000", RatchetBreakeven)

		sell := place(t, j, order{
			intentID: "i-exit", attemptID: "a-exit", orderID: "o-exit", side: "SELL", quantity: "4",
		})
		if err := j.AttachExitIntent(ctx, positionID, sell.intentID); err != nil {
			t.Fatalf("AttachExitIntent: %v", err)
		}
		if _, err := j.RecordFill(ctx, terminalFill(sell, "4", "72000")); err != nil {
			t.Fatalf("RecordFill: %v", err)
		}
	}()

	restarted := openTestJournalAt(t, path)
	restored, err := restarted.ExitState(ctx, positionID)
	if err != nil {
		t.Fatalf("ExitState: %v", err)
	}
	if restored.Pending() {
		t.Errorf("a resolved proposal came back as pending: %+v", restored)
	}
	if restored.TakenRatioTotal != "0.4" {
		t.Fatalf("taken = %s, want the fill's fraction to have survived", restored.TakenRatioTotal)
	}

	// The restored state is what the evaluator reads, and it suppresses the
	// once-per-position partial — which is the "미재발의" half of the contract
	// being satisfied by the row rather than by memory.
	decision, err := exitpolicy.EvaluateRatchet(exitpolicy.RatchetInput{
		Entry: restored.EntryPrice, InitialStop: restored.InitialStop,
		ObservedPrice: "73000", HighWater: restored.HighWater, Baseline: restored.Baseline,
		RealBreakeven: "70120", TakenRatioTotal: restored.TakenRatioTotal,
		PendingAction: exitpolicy.Action(restored.PendingAction),
	})
	if err != nil {
		t.Fatalf("EvaluateRatchet: %v", err)
	}
	if !decision.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none — 40%% was already taken", decision.Proposal)
	}
	if decision.Suppressed != exitpolicy.SuppressedAlreadyTaken {
		t.Errorf("suppressed = %q, want the once-per-position rule", decision.Suppressed)
	}
}

// TestARefusedProposalIsReArmedAfterARestart closes the third crash case: a
// refusal that committed before the process died must not leave the level
// permanently unproposable.
func TestARefusedProposalIsReArmedAfterARestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	var positionID string
	func() {
		j := openTestJournalAt(t, path)
		if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
			t.Fatalf("SetApplyHooks: %v", err)
		}
		o, _ := openedPosition(t, j, "10")
		positionID = currentPosition(t, j, o).ID
		armPartial(t, j, positionID, "72000", RatchetBreakeven)
		if err := j.ResolveExitProposal(ctx, positionID, ProposalRefused); err != nil {
			t.Fatalf("ResolveExitProposal: %v", err)
		}
	}()

	restarted := openTestJournalAt(t, path)
	if err := restarted.RecordExitJudgement(ctx, ExitJudgement{
		PositionID: positionID, ObservedPrice: "72000", HighWater: "72000",
		Baseline: "70120", RatchetLevel: RatchetBreakeven, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionRatchetPartial), Level: RatchetBreakeven,
		},
	}); err != nil {
		t.Fatalf("re-proposing after a refused-then-restarted proposal: %v", err)
	}
}

// --- the working set ----------------------------------------------------------

func TestTheWorkingSetIsScopedToTheAccount(t *testing.T) {
	j := exitFixture(t)
	ctx := context.Background()
	openedPosition(t, j, "10")

	states, err := j.OpenExitStates(ctx, "acct-1")
	if err != nil {
		t.Fatalf("OpenExitStates: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("%d states, want 1", len(states))
	}
	other, err := j.OpenExitStates(ctx, "acct-2")
	if err != nil {
		t.Fatalf("OpenExitStates: %v", err)
	}
	if len(other) != 0 {
		t.Errorf("%d states on another account, want none of this account's", len(other))
	}
}

// armPartial records a judgement that arms a 40 % partial, which several tests
// need as a starting state.
func armPartial(t *testing.T, j *Journal, positionID, price, level string) {
	t.Helper()
	if err := j.RecordExitJudgement(context.Background(), ExitJudgement{
		PositionID: positionID, ObservedPrice: price, HighWater: price,
		Baseline: "70120", RatchetLevel: level, ActiveRung: exitpolicy.NoRung,
		Proposal: &ExitProposal{
			Action: string(exitpolicy.ActionRatchetPartial), Level: level,
		},
	}); err != nil {
		t.Fatalf("arming a proposal: %v", err)
	}
}
