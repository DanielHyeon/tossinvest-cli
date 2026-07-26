package exitpolicy_test

// ratchet_test.go ports StockOS `tests/test_baseline_ratchet.py` in full — all
// six cases, the same fixture, the same numbers — and adds the cases the four
// D5 corrections created.
//
// # The ported six, and what changed about them
//
//	test_below_first_trigger_noops                   → TestBelowTheFirstTriggerNothingMoves
//	test_half_risk_trigger_raises_to_minus_half_r    → TestHalfRiskTriggerRaisesToMinusHalfR
//	test_breakeven_trigger_uses_real_breakeven       → TestBreakevenTriggerUsesTheRealBreakeven
//	test_one_r_suggests_partial_without_lowering_stop→ TestOneRSuggestsAPartialWithoutLoweringTheBaseline
//	test_partial_lock_moves_stop_above_entry         → TestPartialLockMovesTheBaselineAboveEntry
//	test_profit_lock_uses_high_since_entry           → TestProfitLockUsesTheWatermark
//
// Five carry over unchanged. The sixth gains an assertion rather than losing
// one: at a watermark of +2.0R the baseline locks +0.8R = 10160, and the
// observation that case makes (10100) is *below* it. The original module cannot
// see that — it only raises stops — so its test does not mention it. Here the
// same numbers also demonstrate the breach output, which is the honest reading
// of the fixture the original wrote.
//
// Nothing was excluded. The file has six tests and six are here.

import (
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// input is `tests/test_baseline_ratchet.py:12-25`: entry 10000, stop 9800,
// break-even 10030, previous stop 9800 unless overridden.
func input(current string, opts ...func(*exitpolicy.RatchetInput)) exitpolicy.RatchetInput {
	in := exitpolicy.RatchetInput{
		Entry:         "10000",
		InitialStop:   "9800",
		ObservedPrice: current,
		// The original's `high_since_entry` defaults to None and the evaluation
		// falls back to the current price. Here the watermark is a column, so its
		// default is the t0 seed — the entry price — which reproduces the same
		// probe for every case that does not set a high.
		HighWater:     "10000",
		Baseline:      "9800",
		RealBreakeven: "10030",
	}
	for _, o := range opts {
		o(&in)
	}
	return in
}

func high(v string) func(*exitpolicy.RatchetInput) {
	return func(in *exitpolicy.RatchetInput) { in.HighWater = v }
}

func baseline(v string) func(*exitpolicy.RatchetInput) {
	return func(in *exitpolicy.RatchetInput) { in.Baseline = v }
}

func evaluate(t *testing.T, in exitpolicy.RatchetInput) exitpolicy.RatchetDecision {
	t.Helper()
	got, err := exitpolicy.EvaluateRatchet(in)
	if err != nil {
		t.Fatalf("EvaluateRatchet(%+v): %v", in, err)
	}
	return got
}

func TestBelowTheFirstTriggerNothingMoves(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10050"))

	if got.Level != exitpolicy.LevelNone {
		t.Errorf("level = %s, want NONE at +0.25R", got.Level)
	}
	if got.Raised {
		t.Errorf("baseline was raised to %s; below the first trigger there is no candidate at all", got.Baseline)
	}
	if got.Baseline != "9800" {
		t.Errorf("baseline = %s, want the entry stop 9800 unchanged", got.Baseline)
	}
	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none", got.Proposal)
	}
}

func TestHalfRiskTriggerRaisesToMinusHalfR(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10080"))

	if got.Level != exitpolicy.LevelHalfRisk {
		t.Errorf("level = %s, want HALF_RISK at exactly +0.4R", got.Level)
	}
	if got.Baseline != "9900" {
		t.Errorf("baseline = %s, want entry + risk × −0.5 = 9900", got.Baseline)
	}
	if !got.Raised {
		t.Error("the baseline moved 9800 → 9900 and was not reported as raised")
	}
	if got.WinningReason != exitpolicy.CandidateBaselineRatchet {
		t.Errorf("winning reason = %s, want %s", got.WinningReason, exitpolicy.CandidateBaselineRatchet)
	}
}

// TestHalfRiskExcludesTheRealBreakeven is the exclusion
// baseline_ratchet.py:96 encodes and exit-policy restates as "레벨 ≥ BREAKEVEN
// 일 때 실질 본전". At HALF_RISK the policy is deliberately still below entry.
func TestHalfRiskExcludesTheRealBreakeven(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10080"))

	if got.Baseline == "10030" {
		t.Fatal("the real break-even won at HALF_RISK; the first trigger would jump three steps")
	}
	if got.Baseline != "9900" {
		t.Errorf("baseline = %s, want 9900", got.Baseline)
	}
}

func TestBreakevenTriggerUsesTheRealBreakeven(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10160"))

	if got.Level != exitpolicy.LevelBreakeven {
		t.Errorf("level = %s, want BREAKEVEN at +0.8R", got.Level)
	}
	if got.Baseline != "10030" {
		t.Errorf("baseline = %s, want the real break-even 10030", got.Baseline)
	}
	if got.WinningReason != exitpolicy.CandidateRealBreakeven {
		t.Errorf("winning reason = %s, want %s (the tie-break is candidate order)",
			got.WinningReason, exitpolicy.CandidateRealBreakeven)
	}
}

func TestOneRSuggestsAPartialWithoutLoweringTheBaseline(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10200", baseline("10050")))

	if got.Proposal.Action != exitpolicy.ActionRatchetPartial {
		t.Fatalf("action = %q, want the 40%% partial at +1.0R", got.Proposal.Action)
	}
	if got.Proposal.Ratio != "0.4" {
		t.Errorf("ratio = %s, want 0.4 of the remaining quantity", got.Proposal.Ratio)
	}
	if got.Raised {
		t.Error("the baseline was reported as raised; the break-even 10030 is below the anchor 10050")
	}
	if got.Baseline != "10050" {
		t.Errorf("baseline = %s, want the R4 floor to hold it at 10050", got.Baseline)
	}
	if got.WinningReason != exitpolicy.CandidatePrevious {
		t.Errorf("winning reason = %s, want %s", got.WinningReason, exitpolicy.CandidatePrevious)
	}
}

func TestPartialLockMovesTheBaselineAboveEntry(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10240"))

	if got.Level != exitpolicy.LevelPartialLock {
		t.Errorf("level = %s, want PARTIAL_LOCK at +1.2R", got.Level)
	}
	if got.Baseline != "10060" {
		t.Errorf("baseline = %s, want entry + risk × 0.3 = 10060", got.Baseline)
	}
}

func TestProfitLockUsesTheWatermark(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10100", high("10400")))

	if got.Level != exitpolicy.LevelProfitLock {
		t.Errorf("level = %s, want PROFIT_LOCK — the watermark is +2.0R even though the price is +0.5R",
			got.Level)
	}
	if got.Baseline != "10160" {
		t.Errorf("baseline = %s, want entry + risk × 0.8 = 10160", got.Baseline)
	}
	if got.HighWater != "10400" {
		t.Errorf("high water = %s, want the watermark held at 10400", got.HighWater)
	}
	// The assertion the original cannot make: 10100 is below the baseline this
	// evaluation just locked, which is what a trailing stop-out is.
	if got.Proposal.Action != exitpolicy.ActionBaselineBreach {
		t.Errorf("action = %q, want the full liquidation — the observation is below the locked baseline",
			got.Proposal.Action)
	}
	if got.Proposal.Ratio != "1" {
		t.Errorf("ratio = %s, want the whole remaining position", got.Proposal.Ratio)
	}
}

// --- the watermark ------------------------------------------------------------

// TestTheWatermarkHoldsThroughAPullback is exit-policy's "폴링 사이 고점 후
// 되돌림" scenario: observation A at +1.3R then observation B at +0.6R.
func TestTheWatermarkHoldsThroughAPullback(t *testing.T) {
	t.Parallel()

	a := evaluate(t, input("10260")) // +1.3R
	if a.Level != exitpolicy.LevelPartialLock {
		t.Fatalf("A: level = %s, want PARTIAL_LOCK", a.Level)
	}

	b := evaluate(t, input("10120", high(a.HighWater), baseline(a.Baseline))) // +0.6R
	if b.HighWater != "10260" {
		t.Errorf("B: high water = %s, want 10260 held", b.HighWater)
	}
	if b.Level != exitpolicy.LevelPartialLock {
		t.Errorf("B: level = %s, want PARTIAL_LOCK held", b.Level)
	}
	if b.Baseline != a.Baseline {
		t.Errorf("B: baseline = %s, want it not to descend from %s", b.Baseline, a.Baseline)
	}
	if b.Raised {
		t.Error("B: a baseline that did not move must not be reported as raised")
	}
}

func TestTheWatermarkAdvancesToTheObservation(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("10500", high("10100")))

	if got.HighWater != "10500" {
		t.Errorf("high water = %s, want the observation to have advanced it", got.HighWater)
	}
}

// --- the t0 baseline ----------------------------------------------------------

// TestOpenRatchetStateStartsAtTheEntryStop is D5's first correction: the
// baseline is the entry decision's stop from the moment the position exists.
func TestOpenRatchetStateStartsAtTheEntryStop(t *testing.T) {
	t.Parallel()

	open, err := exitpolicy.OpenRatchetState("10000", "9800")
	if err != nil {
		t.Fatalf("OpenRatchetState: %v", err)
	}
	if open.Baseline != "9800" {
		t.Errorf("baseline = %s, want the entry stop 9800", open.Baseline)
	}
	if open.HighWater != "10000" {
		t.Errorf("high water = %s, want the entry price", open.HighWater)
	}
	if open.InitialRisk != "200" {
		t.Errorf("initial risk = %s, want entry − stop = 200", open.InitialRisk)
	}
	if open.Level != exitpolicy.LevelNone {
		t.Errorf("level = %s, want NONE", open.Level)
	}
}

// TestABreachBeforeTheFirstTriggerLiquidates is the exit-policy scenario "개시
// 직후 손절 하회": the t0 baseline is live before +0.4R.
func TestABreachBeforeTheFirstTriggerLiquidates(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("9799"))

	if got.Level != exitpolicy.LevelNone {
		t.Errorf("level = %s, want NONE — no ratchet step was taken", got.Level)
	}
	if got.Proposal.Action != exitpolicy.ActionBaselineBreach {
		t.Fatalf("action = %q, want the full liquidation below the entry stop", got.Proposal.Action)
	}
	if !got.Proposal.Action.RiskReducing() {
		t.Error("the liquidation is not classified risk-reducing")
	}
}

func TestTouchingTheBaselineIsNotABreach(t *testing.T) {
	t.Parallel()

	got := evaluate(t, input("9800"))

	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v; the baseline is a floor and touching it is not undercutting it",
			got.Proposal)
	}
}

// --- the pending lifecycle, decided ------------------------------------------

func TestAPendingProposalSuppressesTheNextPartial(t *testing.T) {
	t.Parallel()

	in := input("10200")
	in.PendingAction = exitpolicy.ActionRatchetPartial

	got := evaluate(t, in)

	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none while one is unresolved", got.Proposal)
	}
	if got.Suppressed != exitpolicy.SuppressedPending {
		t.Errorf("suppressed = %q, want %q", got.Suppressed, exitpolicy.SuppressedPending)
	}
}

// TestTheFortyPercentPartialIsProposedOncePerPosition is the correction to the
// original's memoryless suggestion (baseline_ratchet.py:113).
func TestTheFortyPercentPartialIsProposedOncePerPosition(t *testing.T) {
	t.Parallel()

	in := input("10500") // +2.5R: well past every trigger
	in.TakenRatioTotal = "0.4"

	got := evaluate(t, in)

	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none — 40%% has already been taken", got.Proposal)
	}
	if got.Suppressed != exitpolicy.SuppressedAlreadyTaken {
		t.Errorf("suppressed = %q, want %q", got.Suppressed, exitpolicy.SuppressedAlreadyTaken)
	}
	if got.Level != exitpolicy.LevelProfitLock {
		t.Errorf("level = %s: the suppression must not stop the baseline advancing", got.Level)
	}
}

// TestABreachIsNotWithheldForAPendingTakeProfit is §0 invariant 3. A partial
// that has not filled must not delay a liquidation.
func TestABreachIsNotWithheldForAPendingTakeProfit(t *testing.T) {
	t.Parallel()

	in := input("10100", high("10400"))
	in.PendingAction = exitpolicy.ActionRatchetPartial

	got := evaluate(t, in)

	if got.Proposal.Action != exitpolicy.ActionBaselineBreach {
		t.Fatalf("action = %q, want the liquidation; a pending take-profit does not delay a stop",
			got.Proposal.Action)
	}
	if !got.CancelPendingFirst {
		t.Error("the loop was not told to cancel the working order first; two sells can oversell")
	}
}

func TestAPendingBreachIsNotProposedTwice(t *testing.T) {
	t.Parallel()

	in := input("10100", high("10400"))
	in.PendingAction = exitpolicy.ActionBaselineBreach

	got := evaluate(t, in)

	if !got.Proposal.Zero() {
		t.Errorf("proposal = %+v, want none — the same liquidation is already outstanding", got.Proposal)
	}
	if got.Suppressed != exitpolicy.SuppressedPending {
		t.Errorf("suppressed = %q, want %q", got.Suppressed, exitpolicy.SuppressedPending)
	}
}

// --- input validation ---------------------------------------------------------

// TestUnusableInputsAreRefused ports `_validate_inputs`
// (baseline_ratchet.py:155-170) clause by clause.
func TestUnusableInputsAreRefused(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		mut   func(*exitpolicy.RatchetInput)
		field string
	}{
		{"entry must be positive", func(in *exitpolicy.RatchetInput) { in.Entry = "0" }, "entry price"},
		{"initial stop must be positive", func(in *exitpolicy.RatchetInput) { in.InitialStop = "0" }, "initial stop"},
		{"observed price must be positive", func(in *exitpolicy.RatchetInput) { in.ObservedPrice = "-1" }, "observed price"},
		{"real breakeven must be positive", func(in *exitpolicy.RatchetInput) { in.RealBreakeven = "0" }, "real breakeven"},
		{"stop must be below entry", func(in *exitpolicy.RatchetInput) { in.InitialStop = "10000" }, "initial stop"},
		{"baseline must be positive", func(in *exitpolicy.RatchetInput) { in.Baseline = "0" }, "baseline"},
		{"high water must be positive", func(in *exitpolicy.RatchetInput) { in.HighWater = "0" }, "high water"},
		// TossOS's two: both are columns the original does not have.
		{"taken ratio is a fraction", func(in *exitpolicy.RatchetInput) { in.TakenRatioTotal = "1.5" }, "taken ratio total"},
		{"prices must be readable", func(in *exitpolicy.RatchetInput) { in.ObservedPrice = "1e4" }, "observed price"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in := input("10100")
			tc.mut(&in)

			_, err := exitpolicy.EvaluateRatchet(in)
			if err == nil {
				t.Fatal("the evaluation was performed on an input that violates an invariant")
			}
			if !errors.Is(err, exitpolicy.ErrRefused) {
				t.Errorf("err = %v, want it to match ErrRefused so the loop can alert", err)
			}
			var refusal *exitpolicy.RefusalError
			if !errors.As(err, &refusal) {
				t.Fatalf("err = %v, want a *RefusalError naming the field", err)
			}
			if refusal.Field != tc.field {
				t.Errorf("field = %q, want %q", refusal.Field, tc.field)
			}
		})
	}
}

// TestARefusalIsNotAJudgement fixes the distinction exit-policy's "판정 거부 +
// 알림" depends on.
func TestARefusalIsNotAJudgement(t *testing.T) {
	t.Parallel()

	in := input("10100")
	in.InitialStop = "10500"

	got, err := exitpolicy.EvaluateRatchet(in)
	if err == nil {
		t.Fatal("want a refusal")
	}
	if got.HighWater != "" || got.Level != "" || got.Baseline != "" {
		t.Errorf("decision = %+v, want the zero value; a refused evaluation states nothing", got)
	}
}

// --- the configured table -----------------------------------------------------

func TestADescendingTriggerTableIsRefused(t *testing.T) {
	t.Parallel()

	cfg := exitpolicy.DefaultRatchetConfig()
	cfg.PartialLockTriggerR = "0.5" // below the break-even trigger
	in := input("10100")
	in.Config = &cfg

	_, err := exitpolicy.EvaluateRatchet(in)
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "not above the trigger before it") {
		t.Errorf("err = %v, want it to name the ordering", err)
	}
}

func TestADescendingLockTableIsRefused(t *testing.T) {
	t.Parallel()

	cfg := exitpolicy.DefaultRatchetConfig()
	cfg.ProfitLockStopR = "0.1" // below the partial lock's 0.3
	in := input("10100")
	in.Config = &cfg

	_, err := exitpolicy.EvaluateRatchet(in)
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if !strings.Contains(err.Error(), "does not descend") {
		t.Errorf("err = %v, want it to name the ratchet contract", err)
	}
}

func TestTheDefaultTableIsTheOriginals(t *testing.T) {
	t.Parallel()

	cfg := exitpolicy.DefaultRatchetConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("the shipped default does not validate: %v", err)
	}
	want := exitpolicy.RatchetConfig{
		HalfRiskTriggerR: "0.4", BreakevenTriggerR: "0.8", PartialTriggerR: "1.0",
		PartialLockTriggerR: "1.2", ProfitLockTriggerR: "2.0",
		HalfRiskStopR: "-0.5", PartialLockStopR: "0.3", ProfitLockStopR: "0.8",
		PartialRatio: "0.4",
	}
	if cfg != want {
		t.Errorf("config = %+v, want baseline_ratchet.py:32-41 %+v", cfg, want)
	}
}

// --- the denominator rule -----------------------------------------------------

func TestCumulativeAfterConvertsRemainingToInitial(t *testing.T) {
	t.Parallel()

	cases := []struct{ taken, ratio, want string }{
		{"0", "0.4", "0.4"},
		{"0.4", "0.25", "0.55"}, // 25 % of the remaining 60 % is 15 %
		{"0.55", "1", "1"},      // the rest of it
		{"", "0.4", "0.4"},      // the empty spelling of zero
		{"0", "1", "1"},         // a full take
		{"0.4", "0.5", "0.7"},   // half of the remaining 60 %
		{"0.9", "0.5", "0.95"},
		{"0", "0.333", "0.333"},
		{"0.5", "0.5", "0.75"},
		{"0.999", "0.5", "0.9995"},
	}
	for _, tc := range cases {
		got, err := exitpolicy.CumulativeAfter(tc.taken, tc.ratio)
		if err != nil {
			t.Fatalf("CumulativeAfter(%q, %q): %v", tc.taken, tc.ratio, err)
		}
		if got != tc.want {
			t.Errorf("CumulativeAfter(%q, %q) = %s, want %s", tc.taken, tc.ratio, got, tc.want)
		}
	}
}

func TestTakenAfterFillReconstructsTheInitialQuantity(t *testing.T) {
	t.Parallel()

	// 100 held, none taken, 40 sold → 60 remain and 40 % of the initial is taken.
	first, err := exitpolicy.TakenAfterFill("0", "40", "60")
	if err != nil {
		t.Fatalf("TakenAfterFill: %v", err)
	}
	if first != "0.4" {
		t.Fatalf("taken = %s, want 0.4", first)
	}
	// 30 of the remaining 60 → 70 of the original 100.
	second, err := exitpolicy.TakenAfterFill(first, "30", "30")
	if err != nil {
		t.Fatalf("TakenAfterFill: %v", err)
	}
	if second != "0.7" {
		t.Fatalf("taken = %s, want 0.7", second)
	}
	// The rest.
	third, err := exitpolicy.TakenAfterFill(second, "30", "0")
	if err != nil {
		t.Fatalf("TakenAfterFill: %v", err)
	}
	if third != "1" {
		t.Fatalf("taken = %s, want the whole position", third)
	}
}

func TestTakenAfterFillOnAnEmptyPositionStandsStill(t *testing.T) {
	t.Parallel()

	got, err := exitpolicy.TakenAfterFill("0.4", "0", "0")
	if err != nil {
		t.Fatalf("TakenAfterFill: %v", err)
	}
	if got != "0.4" {
		t.Errorf("taken = %s, want it unchanged; there is no fraction of nothing", got)
	}
}

// --- the triple monotone property ---------------------------------------------

// TestTripleMonotoneProperty is exit-policy's "기준선 단조 property" scenario:
// for an arbitrary sequence of observations, the baseline, the level and the
// watermark are all non-decreasing.
//
// All three, not one. The watermark is monotone by construction, the level is a
// function of the watermark, and the baseline is a max whose first term is its
// own previous value — so a failure in any one of them would be a failure of a
// different link, and asserting only the baseline would let a level regression
// through whenever the baseline happened to be pinned by the R4 floor.
//
// The sequence is generated rather than enumerated because the property is about
// the ordering of arbitrary observations, and the interesting sequences are the
// ones nobody would think to write down: a spike then a collapse then a slow
// grind back through the same triggers.
func TestTripleMonotoneProperty(t *testing.T) {
	t.Parallel()

	const (
		entry       = "10000"
		initialStop = "9800"
	)
	open, err := exitpolicy.OpenRatchetState(entry, initialStop)
	if err != nil {
		t.Fatalf("OpenRatchetState: %v", err)
	}

	for seed := int64(0); seed < 200; seed++ {
		rng := rand.New(rand.NewSource(seed))
		state := open
		var history []string

		for step := 0; step < 40; step++ {
			// Prices from 9000 to 11500 in whole KRW: below the entry stop, across
			// every trigger, and past the last one.
			price := itoa(9000 + rng.Intn(2501))
			in := exitpolicy.RatchetInput{
				Entry: entry, InitialStop: initialStop,
				ObservedPrice: price,
				HighWater:     state.HighWater,
				Baseline:      state.Baseline,
				RealBreakeven: "10030",
			}
			got, err := exitpolicy.EvaluateRatchet(in)
			if err != nil {
				t.Fatalf("seed %d step %d: EvaluateRatchet(%s): %v", seed, step, price, err)
			}
			history = append(history, price+"→"+got.Baseline+"/"+string(got.Level)+"/"+got.HighWater)

			if cmpDecimal(t, got.HighWater, state.HighWater) < 0 {
				t.Fatalf("seed %d step %d: watermark fell %s → %s\n%s",
					seed, step, state.HighWater, got.HighWater, strings.Join(history, "\n"))
			}
			if got.Level.Rank() < state.Level.Rank() {
				t.Fatalf("seed %d step %d: level fell %s → %s\n%s",
					seed, step, state.Level, got.Level, strings.Join(history, "\n"))
			}
			if cmpDecimal(t, got.Baseline, state.Baseline) < 0 {
				t.Fatalf("seed %d step %d: baseline fell %s → %s\n%s",
					seed, step, state.Baseline, got.Baseline, strings.Join(history, "\n"))
			}
			// Raised is the strict `>` the spec records a rise on, so it must agree
			// with the comparison rather than be a second opinion about it.
			if want := cmpDecimal(t, got.Baseline, state.Baseline) > 0; got.Raised != want {
				t.Fatalf("seed %d step %d: Raised = %v for %s → %s",
					seed, step, got.Raised, state.Baseline, got.Baseline)
			}

			state.HighWater, state.Baseline, state.Level = got.HighWater, got.Baseline, got.Level
		}
	}
}

// TestTheLevelIsAFunctionOfTheWatermarkAlone is the property the monotone
// argument rests on: two evaluations with the same watermark produce the same
// level whatever the observation was.
func TestTheLevelIsAFunctionOfTheWatermarkAlone(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 200; i++ {
		watermark := itoa(10000 + rng.Intn(1500))
		a := evaluate(t, input(itoa(9500+rng.Intn(500)), high(watermark)))
		b := evaluate(t, input(watermark, high(watermark)))
		if a.Level != b.Level {
			t.Fatalf("watermark %s: level %s from a low observation but %s from a high one",
				watermark, a.Level, b.Level)
		}
	}
}
