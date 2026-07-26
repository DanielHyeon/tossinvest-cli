package exitpolicy_test

// ladder_test.go ports StockOS `tests/test_profit_ladder.py`. The file has 19
// tests; 13 are here and 6 are excluded, each for a reason that is a scope
// decision of this change rather than a shortcut.
//
// # Ported (13)
//
//	test_t1_alone_promotes_state_only_no_order        → TestTheFirstRungPromotesStateOnly
//	test_t2_alone_emits_partial_take_remaining_25_pct → TestTheSecondRungTakesAQuarterOfTheRemainder
//	test_final_rung_emits_take_full                   → TestTheFinalRungTakesEverything
//	test_protected_stop_breached_returns_stop_loss…   → TestABreachOfTheRungLockLiquidates
//	test_revisit_same_rung_no_state_change            → TestRevisitingARungChangesNothing
//	test_same_bar_t1_t2_t3_promotes_to_t3             → TestAWatermarkJumpPromotesToTheHighestRung
//	test_ladder_state_policy_id_mismatch_raises       → TestAStateFromAnotherPolicyIsRefused
//	test_rung_partial_ratio_must_be_zero_to_one       → TestARungRatioOutsideZeroToOneIsRefused
//	test_policy_targets_must_be_strictly_increasing   → TestTargetsMustStrictlyIncrease
//	test_policy_stops_must_be_monotone                → TestLocksMustNotDescend
//	test_pending_order_blocks_new_partial_action      → TestAPendingOrderBlocksTheNextRungPartial
//	test_evaluate_exit_xor_validation                 → TestAPolicyIsRequired
//	test_is_orderable_exit_action_filters_state_only  → TestOnlyOrderableActionsReachTheSubmissionPath
//
// The float fixtures are rewritten as decimals throughout: `entry * (1 +
// price_pct / 100)` is computed exactly here, and the assertions compare decimal
// strings rather than float equality (`protected_stop_pct == 1.0`).
//
// # Excluded (6)
//
//	test_idempotency_key_includes_policy_rung_action  — `ladder_order_idempotency_key`
//	test_idempotency_key_changes_with_action            is not ported. TossOS's broker
//	                                                    idempotency key is the journal's
//	                                                    f(decision_id, generation) from 2a
//	                                                    (internal/journal/idempotency.go); a
//	                                                    second scheme would be a second
//	                                                    authority on the same question.
//	test_backtest_same_bar_target_and_stop_stop_first — FillModel/STOP_FIRST needs OHLC input
//	test_backtest_same_bar_target_first_overrides       that only a backtest has. exit-policy
//	                                                    demotes it from SHALL to P3 explicitly.
//	test_evaluate_exit_legacy_path_unaffected_when…   — the legacy two-tier take-profit path is
//	                                                    signal-layer and not ported at all.
//	test_evaluate_exit_ladder_without_bar_range…      — an a039 regression guarding the case
//	                                                    where OHLC fields are absent. That is
//	                                                    the only input shape TossOS has, so
//	                                                    every case in this file is that test.
//
// # From tests/test_exit_strategy.py: 0 of 54 ported
//
// That file tests the pipeline around the ladder, not the ladder. Its cases fall
// into groups this change admits nothing from: ATR-armed breakeven and MFE
// trailing floors (9), trendline zones (2), trend adaptation and rapid profit
// locks (12), grace periods and stop confirmation (11), signal take-profits (6),
// time exits (3), legacy two-tier targets and their precedence (6), hard-percent
// stops (2), early-surge partials (2), the ladder/trailing interaction (1). Every
// one is either a signal input (P3, zero admitted here) or the legacy path.
//
// Two have structural replacements rather than ports:
// `test_breakeven_protection_moves_stop_after_small_profit` — a break-even
// promotion armed by ATR — is replaced by the R-triggered BREAKEVEN level
// (ratchet_test.go), and `test_non_a011_legacy_exit_ignores_bar_low_when_ladder
// _omitted` is vacuous here because no evaluation reads a bar low.

import (
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// ladderInput is `_ctx` (test_profit_ladder.py:43-51) with the float arithmetic
// done exactly: entry 10000, observation at entry × (1 + pct/100).
//
// The two fields the original has no equivalent of are the watermark and the
// baseline. The watermark defaults to the observation (the original's `high`
// defaults to the same), and the baseline to the entry stop 9800, which is
// TossOS's t0 seed.
func ladderInput(price string, opts ...func(*exitpolicy.LadderInput)) exitpolicy.LadderInput {
	in := exitpolicy.LadderInput{
		EntryPrice:    "10000",
		ObservedPrice: price,
		HighWater:     price,
		Baseline:      "9800",
		Policy:        exitpolicy.DefaultLadderPolicy(),
		State: exitpolicy.LadderState{
			PolicyID:      "default_v1",
			ActivatedRung: exitpolicy.NoRung,
			PendingRung:   exitpolicy.NoRung,
		},
	}
	for _, o := range opts {
		o(&in)
	}
	return in
}

func atRung(index int, baseline string) func(*exitpolicy.LadderInput) {
	return func(in *exitpolicy.LadderInput) {
		in.State.ActivatedRung = index
		in.Baseline = baseline
	}
}

func evaluateLadder(t *testing.T, in exitpolicy.LadderInput) exitpolicy.LadderTransition {
	t.Helper()
	got, err := exitpolicy.EvaluateLadder(in)
	if err != nil {
		t.Fatalf("EvaluateLadder: %v", err)
	}
	return got
}

// --- the ported promotion cases -----------------------------------------------

func TestTheFirstRungPromotesStateOnly(t *testing.T) {
	t.Parallel()

	got := evaluateLadder(t, ladderInput("10160")) // +1.6 %

	if got.Action != exitpolicy.ActionLadderHoldStopPromoted {
		t.Errorf("action = %q, want the state-only promotion", got.Action)
	}
	if got.Reason != exitpolicy.ReasonStateOnly {
		t.Errorf("reason = %q, want %q", got.Reason, exitpolicy.ReasonStateOnly)
	}
	if got.NextState.ActivatedRung != 0 {
		t.Errorf("activated rung = %d, want 0", got.NextState.ActivatedRung)
	}
	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none — rung 0 takes nothing", got.Proposal)
	}
	// The percent the original stores becomes a price here: 0 % of entry.
	if got.Baseline != "10000" {
		t.Errorf("baseline = %s, want the breakeven lock 10000", got.Baseline)
	}
	if !got.Raised {
		t.Error("the baseline moved 9800 → 10000 and was not reported as raised")
	}
}

func TestTheSecondRungTakesAQuarterOfTheRemainder(t *testing.T) {
	t.Parallel()

	got := evaluateLadder(t, ladderInput("10260")) // +2.6 %

	if got.Action != exitpolicy.ActionLadderPartial {
		t.Fatalf("action = %q, want LADDER_PARTIAL", got.Action)
	}
	if got.Reason != exitpolicy.ReasonRungPartialTake {
		t.Errorf("reason = %q, want %q", got.Reason, exitpolicy.ReasonRungPartialTake)
	}
	if got.NextState.ActivatedRung != 1 {
		t.Errorf("activated rung = %d, want 1", got.NextState.ActivatedRung)
	}
	if got.Proposal.Ratio != "0.25" {
		t.Errorf("ratio = %s, want 0.25 of the remaining quantity", got.Proposal.Ratio)
	}
	if got.Baseline != "10100" {
		t.Errorf("baseline = %s, want the 1.0 %% lock 10100", got.Baseline)
	}
	if got.Proposal.Level != "1" {
		t.Errorf("proposal level = %q, want the rung index", got.Proposal.Level)
	}
}

func TestTheFinalRungTakesEverything(t *testing.T) {
	t.Parallel()

	got := evaluateLadder(t, ladderInput("10700")) // +7.0 %

	if got.Action != exitpolicy.ActionLadderTakeProfit {
		t.Fatalf("action = %q, want LADDER_TAKE_PROFIT", got.Action)
	}
	if got.Reason != exitpolicy.ReasonFinalRungTakeAll {
		t.Errorf("reason = %q, want %q", got.Reason, exitpolicy.ReasonFinalRungTakeAll)
	}
	if got.NextState.ActivatedRung != 3 {
		t.Errorf("activated rung = %d, want 3", got.NextState.ActivatedRung)
	}
	if got.Proposal.Ratio != "1" {
		t.Errorf("ratio = %s, want the whole remainder", got.Proposal.Ratio)
	}
}

func TestABreachOfTheRungLockLiquidates(t *testing.T) {
	t.Parallel()

	// Rung 2 activated: its 2.0 % lock is 10200, and the observation is below it.
	got := evaluateLadder(t, ladderInput("10150", atRung(2, "10200")))

	if got.Action != exitpolicy.ActionLadderStop {
		t.Fatalf("action = %q, want STOP_LOSS_LADDER", got.Action)
	}
	if got.Reason != exitpolicy.ReasonStopBreached {
		t.Errorf("reason = %q, want %q", got.Reason, exitpolicy.ReasonStopBreached)
	}
	if got.Proposal.Ratio != "1" {
		t.Errorf("ratio = %s, want the whole position", got.Proposal.Ratio)
	}
}

func TestRevisitingARungChangesNothing(t *testing.T) {
	t.Parallel()

	got := evaluateLadder(t, ladderInput("10260", atRung(1, "10100")))

	if got.Action != exitpolicy.ActionNone {
		t.Errorf("action = %q, want none", got.Action)
	}
	if got.Reason != exitpolicy.ReasonNoRungPromotion {
		t.Errorf("reason = %q, want %q", got.Reason, exitpolicy.ReasonNoRungPromotion)
	}
	if got.NextState.ActivatedRung != 1 {
		t.Errorf("activated rung = %d, want 1", got.NextState.ActivatedRung)
	}
	if got.RungPromotedTo != exitpolicy.NoRung {
		t.Errorf("promoted to = %d, want none this evaluation", got.RungPromotedTo)
	}
	if got.Raised {
		t.Error("the baseline did not move and must not be reported as raised")
	}
}

// TestAWatermarkJumpPromotesToTheHighestRung is the original's same-bar case
// with the bar's high replaced by the watermark: three rungs cleared at once
// promotes to the third and proposes only its partial.
func TestAWatermarkJumpPromotesToTheHighestRung(t *testing.T) {
	t.Parallel()

	got := evaluateLadder(t, ladderInput("10400")) // +4.0 %: rungs 0, 1 and 2

	if got.NextState.ActivatedRung != 2 {
		t.Errorf("activated rung = %d, want 2", got.NextState.ActivatedRung)
	}
	if got.Action != exitpolicy.ActionLadderPartial {
		t.Errorf("action = %q, want the third rung's partial only", got.Action)
	}
	if got.Proposal.Ratio != "0.25" {
		t.Errorf("ratio = %s, want 0.25", got.Proposal.Ratio)
	}
}

// TestAWatermarkJumpFollowedByAPullbackLiquidates is what the original's
// same-bar tie-break approximated with `low`, expressed the way a tick stream
// actually delivers it: the high promoted the lock, the next observation is
// under it.
func TestAWatermarkJumpFollowedByAPullbackLiquidates(t *testing.T) {
	t.Parallel()

	in := ladderInput("10050")
	in.HighWater = "10400" // the watermark cleared rung 2 (2.0 % lock = 10200)

	got := evaluateLadder(t, in)

	if got.NextState.ActivatedRung != 2 {
		t.Errorf("activated rung = %d, want the promotion to have happened", got.NextState.ActivatedRung)
	}
	if got.Baseline != "10200" {
		t.Errorf("baseline = %s, want the promoted lock", got.Baseline)
	}
	if got.Action != exitpolicy.ActionLadderStop {
		t.Errorf("action = %q, want the liquidation — 10050 is under the lock it just set", got.Action)
	}
}

// --- the ported validation cases ----------------------------------------------

func TestAStateFromAnotherPolicyIsRefused(t *testing.T) {
	t.Parallel()

	in := ladderInput("10260")
	in.State.PolicyID = "other_v1"

	_, err := exitpolicy.EvaluateLadder(in)
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "state/policy mismatch") {
		t.Errorf("err = %v, want the original's message", err)
	}
}

func TestARungRatioOutsideZeroToOneIsRefused(t *testing.T) {
	t.Parallel()

	err := exitpolicy.Rung{TargetPct: "1.0", StopPct: "0", PartialRatio: "1.5"}.Validate()
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "partial_ratio") {
		t.Errorf("err = %v, want the original's message", err)
	}
}

func TestTargetsMustStrictlyIncrease(t *testing.T) {
	t.Parallel()

	err := exitpolicy.LadderPolicy{
		PolicyID: "bad",
		Rungs: []exitpolicy.Rung{
			{TargetPct: "1.0", StopPct: "0", PartialRatio: "0"},
			{TargetPct: "1.0", StopPct: "0.5", PartialRatio: "0.5"},
		},
	}.Validate()
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "strictly increasing") {
		t.Errorf("err = %v, want the original's message", err)
	}
}

func TestLocksMustNotDescend(t *testing.T) {
	t.Parallel()

	err := exitpolicy.LadderPolicy{
		PolicyID: "bad",
		Rungs: []exitpolicy.Rung{
			{TargetPct: "1.0", StopPct: "1.0", PartialRatio: "0"},
			{TargetPct: "2.0", StopPct: "0.5", PartialRatio: "0"},
		},
	}.Validate()
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "monotonically non-decreasing") {
		t.Errorf("err = %v, want the original's message", err)
	}
}

func TestAPolicyIsRequired(t *testing.T) {
	t.Parallel()

	in := ladderInput("10260")
	in.Policy = exitpolicy.LadderPolicy{}

	_, err := exitpolicy.EvaluateLadder(in)
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal; a ladder state without its rung table judges nothing", err)
	}
}

func TestAnEmptyRungSetIsRefused(t *testing.T) {
	t.Parallel()

	err := exitpolicy.LadderPolicy{PolicyID: "empty"}.Validate()
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

func TestARungIndexOutsideThePolicyIsRefused(t *testing.T) {
	t.Parallel()

	in := ladderInput("10260")
	in.State.ActivatedRung = 9

	_, err := exitpolicy.EvaluateLadder(in)
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal; a rung the table does not have is a corrupted row", err)
	}
}

func TestOnlyOrderableActionsReachTheSubmissionPath(t *testing.T) {
	t.Parallel()

	orderable := map[exitpolicy.Action]bool{
		exitpolicy.ActionBaselineBreach:   true,
		exitpolicy.ActionRatchetPartial:   true,
		exitpolicy.ActionLadderPartial:    true,
		exitpolicy.ActionLadderTakeProfit: true,
		exitpolicy.ActionLadderStop:       true,

		exitpolicy.ActionNone:                   false,
		exitpolicy.ActionLadderHoldStopPromoted: false,
	}
	for action, want := range orderable {
		if got := action.Orderable(); got != want {
			t.Errorf("%q.Orderable() = %v, want %v", action, got, want)
		}
	}
}

// --- the pending lifecycle ----------------------------------------------------

func TestAPendingOrderBlocksTheNextRungPartial(t *testing.T) {
	t.Parallel()

	in := ladderInput("10260", atRung(0, "10000"))
	in.State.PendingAction = exitpolicy.ActionLadderPartial
	in.State.PendingRung = 1

	got := evaluateLadder(t, in)

	if got.Action != exitpolicy.ActionLadderHoldStopPromoted {
		t.Errorf("action = %q, want the state-only hold", got.Action)
	}
	if got.Reason != exitpolicy.ReasonBlockedByPending {
		t.Errorf("reason = %q, want the original's %q", got.Reason, exitpolicy.ReasonBlockedByPending)
	}
	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none", got.Proposal)
	}
	// The protection still moved: only the order is de-duplicated.
	if got.NextState.ActivatedRung != 1 || got.Baseline != "10100" {
		t.Errorf("state = rung %d baseline %s, want the promotion to have happened anyway",
			got.NextState.ActivatedRung, got.Baseline)
	}
}

// TestALadderBreachIsNotWithheldForAPendingPartial is §0 invariant 3 on the
// ladder side.
func TestALadderBreachIsNotWithheldForAPendingPartial(t *testing.T) {
	t.Parallel()

	in := ladderInput("10150", atRung(2, "10200"))
	in.State.PendingAction = exitpolicy.ActionLadderPartial
	in.State.PendingRung = 2

	got := evaluateLadder(t, in)

	if got.Action != exitpolicy.ActionLadderStop {
		t.Fatalf("action = %q, want the liquidation", got.Action)
	}
	if !got.CancelPendingFirst {
		t.Error("the loop was not told to cancel the working order first")
	}
}

func TestAPendingLadderStopIsNotProposedTwice(t *testing.T) {
	t.Parallel()

	in := ladderInput("10150", atRung(2, "10200"))
	in.State.PendingAction = exitpolicy.ActionLadderStop

	got := evaluateLadder(t, in)

	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none", got.Proposal)
	}
	if got.Suppressed != exitpolicy.SuppressedPending {
		t.Errorf("suppressed = %q, want %q", got.Suppressed, exitpolicy.SuppressedPending)
	}
}

// --- the decision-time / execution-time split ---------------------------------

// TestAnEvaluationNeverMovesTheExecutionTimeFields is profit_ladder.py:26-34.
// The two fields a fill owns must survive an evaluation untouched, because the
// evaluation runs every five seconds and the fill transaction is the only place
// that may move them.
func TestAnEvaluationNeverMovesTheExecutionTimeFields(t *testing.T) {
	t.Parallel()

	in := ladderInput("10700")
	in.State.TakenRatioTotal = "0.55"

	got := evaluateLadder(t, in)

	if got.NextState.TakenRatioTotal != "0.55" {
		t.Errorf("taken ratio = %s, want it untouched by a decision", got.NextState.TakenRatioTotal)
	}
	if got.NextState.Completed {
		t.Error("completed was set by a decision; it moves when a fill says so")
	}
}

func TestACompletedLadderJudgesNothingFurther(t *testing.T) {
	t.Parallel()

	in := ladderInput("10700", atRung(3, "10350"))
	in.State.Completed = true
	in.State.TakenRatioTotal = "1"

	got := evaluateLadder(t, in)

	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none from a finished ladder", got.Proposal)
	}
	if got.Reason != exitpolicy.ReasonCompleted {
		t.Errorf("reason = %q, want %q", got.Reason, exitpolicy.ReasonCompleted)
	}
}

// --- the t0 baseline on the ladder side ---------------------------------------

func TestOpenLadderStateStartsAtTheEntryStop(t *testing.T) {
	t.Parallel()

	open, err := exitpolicy.OpenLadderState("10000", "9800", exitpolicy.DefaultLadderPolicy())
	if err != nil {
		t.Fatalf("OpenLadderState: %v", err)
	}
	if open.Baseline != "9800" {
		t.Errorf("baseline = %s, want the entry stop", open.Baseline)
	}
	if open.State.ActivatedRung != exitpolicy.NoRung {
		t.Errorf("activated rung = %d, want none", open.State.ActivatedRung)
	}
	if open.State.PolicyID != "default_v1" {
		t.Errorf("policy id = %s, want the policy's", open.State.PolicyID)
	}
}

// TestALadderPositionIsProtectedBeforeItsFirstRung is the difference from the
// original, which checks nothing while `activated_rung_index < 0`.
func TestALadderPositionIsProtectedBeforeItsFirstRung(t *testing.T) {
	t.Parallel()

	got := evaluateLadder(t, ladderInput("9799"))

	if got.Action != exitpolicy.ActionLadderStop {
		t.Fatalf("action = %q, want the liquidation below the entry stop", got.Action)
	}
	if got.NextState.ActivatedRung != exitpolicy.NoRung {
		t.Errorf("activated rung = %d, want none — no rung was reached", got.NextState.ActivatedRung)
	}
}

// --- the default set ----------------------------------------------------------

func TestTheDefaultRungSetIsTheOriginals(t *testing.T) {
	t.Parallel()

	policy := exitpolicy.DefaultLadderPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("the shipped default does not validate: %v", err)
	}
	want := []exitpolicy.Rung{
		{TargetPct: "1.5", StopPct: "0", PartialRatio: "0"},
		{TargetPct: "2.5", StopPct: "1.0", PartialRatio: "0.25"},
		{TargetPct: "4.0", StopPct: "2.0", PartialRatio: "0.25"},
		{TargetPct: "6.0", StopPct: "3.5", PartialRatio: "1.0"},
	}
	if len(policy.Rungs) != len(want) {
		t.Fatalf("%d rungs, want %d", len(policy.Rungs), len(want))
	}
	for i := range want {
		if policy.Rungs[i] != want[i] {
			t.Errorf("rung %d = %+v, want profit_ladder.py:165-176 %+v", i, policy.Rungs[i], want[i])
		}
	}
	if !policy.FinalTakeFull {
		t.Error("final_take_full = false, want the original's true")
	}
}

func TestTheLockPriceIsExact(t *testing.T) {
	t.Parallel()

	// 3.5 % of 10000 is 10350 exactly; in binary floating point it is not.
	got, err := exitpolicy.LockPrice("10000", "3.5")
	if err != nil {
		t.Fatalf("LockPrice: %v", err)
	}
	if got != "10350" {
		t.Errorf("LockPrice = %s, want 10350", got)
	}
	// A price whose percent does not divide evenly still lands on the decimal.
	got, err = exitpolicy.LockPrice("73330", "1.0")
	if err != nil {
		t.Fatalf("LockPrice: %v", err)
	}
	if got != "74063.3" {
		t.Errorf("LockPrice = %s, want 74063.3", got)
	}
}

// --- the denominator rule on the ladder ---------------------------------------

// TestTheRungRatiosComposeAgainstTheRemainder walks the default set the way a
// position actually would: each rung takes its fraction of what is left, and the
// cumulative total is against the initial quantity (profit_ladder.py:57-66).
func TestTheRungRatiosComposeAgainstTheRemainder(t *testing.T) {
	t.Parallel()

	total := "0"
	for _, step := range []struct{ ratio, want string }{
		{"0", "0"},       // rung 0 takes nothing
		{"0.25", "0.25"}, // 25 % of the whole
		{"0.25", "0.4375"},
		{"1", "1"}, // the final rung takes the rest
	} {
		next, err := exitpolicy.CumulativeAfter(total, step.ratio)
		if err != nil {
			t.Fatalf("CumulativeAfter(%q, %q): %v", total, step.ratio, err)
		}
		if next != step.want {
			t.Fatalf("CumulativeAfter(%q, %q) = %s, want %s", total, step.ratio, next, step.want)
		}
		total = next
	}
}

// --- the monotone property, ladder side ---------------------------------------

// TestTheLadderIsMonotoneToo is the same triple as the ratchet's: baseline, rung
// and watermark all non-decreasing over arbitrary observation sequences. The
// ladder reaches its levels by a different route (a percent table rather than an
// R multiple), so the property has to be demonstrated on that route too.
func TestTheLadderIsMonotoneToo(t *testing.T) {
	t.Parallel()

	policy := exitpolicy.DefaultLadderPolicy()
	open, err := exitpolicy.OpenLadderState("10000", "9800", policy)
	if err != nil {
		t.Fatalf("OpenLadderState: %v", err)
	}

	for seed := int64(0); seed < 150; seed++ {
		rng := rand.New(rand.NewSource(seed))
		baseline, watermark, state := open.Baseline, open.HighWater, open.State
		var history []string

		for step := 0; step < 40; step++ {
			price := itoa(9000 + rng.Intn(2501))
			got, err := exitpolicy.EvaluateLadder(exitpolicy.LadderInput{
				EntryPrice: "10000", ObservedPrice: price,
				HighWater: watermark, Baseline: baseline,
				State: state, Policy: policy,
			})
			if err != nil {
				t.Fatalf("seed %d step %d: %v", seed, step, err)
			}
			history = append(history, price+"→"+got.Baseline+"/r"+itoa(got.NextState.ActivatedRung))

			if cmpDecimal(t, got.HighWater, watermark) < 0 {
				t.Fatalf("seed %d step %d: watermark fell %s → %s\n%s",
					seed, step, watermark, got.HighWater, strings.Join(history, "\n"))
			}
			if got.NextState.ActivatedRung < state.ActivatedRung {
				t.Fatalf("seed %d step %d: rung fell %d → %d\n%s",
					seed, step, state.ActivatedRung, got.NextState.ActivatedRung, strings.Join(history, "\n"))
			}
			if cmpDecimal(t, got.Baseline, baseline) < 0 {
				t.Fatalf("seed %d step %d: baseline fell %s → %s\n%s",
					seed, step, baseline, got.Baseline, strings.Join(history, "\n"))
			}
			if want := cmpDecimal(t, got.Baseline, baseline) > 0; got.Raised != want {
				t.Fatalf("seed %d step %d: Raised = %v for %s → %s",
					seed, step, got.Raised, baseline, got.Baseline)
			}

			baseline, watermark, state = got.Baseline, got.HighWater, got.NextState
		}
	}
}

func TestRungLabelsRoundTrip(t *testing.T) {
	t.Parallel()

	for _, index := range []int{0, 1, 2, 3, 17} {
		got, err := exitpolicy.RungIndex(itoa(index))
		if err != nil {
			t.Fatalf("RungIndex(%d): %v", index, err)
		}
		if got != index {
			t.Errorf("RungIndex(%d) = %d", index, got)
		}
	}
	for _, label := range []string{"", "BREAKEVEN", "-1", "1.5"} {
		if _, err := exitpolicy.RungIndex(label); !errors.Is(err, exitpolicy.ErrRefused) {
			t.Errorf("RungIndex(%q) was accepted; a level that is neither is a corrupted row", label)
		}
	}
}
