package risk

import (
	"strings"
	"testing"
)

// step_test.go is change add-net-rr-measurement task 3.4: the verdict reports
// which rung stopped it.
//
// # Why the verdict and not the reason code
//
// The observation table needs a "stopped step" column so an operator can tell "no
// setup matched" from "the threshold refused it". The obvious way to fill it is to
// map the reason code back to a rung — and it is wrong: ReasonInputUnavailable is
// raised at 42 sites across this package (chain.go 35, contract.go 4, reason.go 3),
// so the inverse is many-to-one and the column would be quietly false for exactly
// the refusals hardest to diagnose.
//
// So Evaluate stamps the rung it was running. The stamp is additive: Decision
// gained a field, no rung function changed, and no verdict moved.

// TestEveryRungReportsItsOwnName walks the whole chain, breaking one rung at a
// time, and checks the verdict names that rung.
//
// Golden against EntryChainSteps() rather than a second hand-written list: two
// lists would drift, and the failure mode of drift here is a column that says the
// wrong rung refused.
func TestEveryRungReportsItsOwnName(t *testing.T) {
	// break[step] makes exactly that rung refuse, with every earlier rung passing.
	breakers := map[string]func(*Input){
		"kill_switch":      func(in *Input) { in.Account.KillSwitchActive = true },
		"operating_mode":   func(in *Input) { in.Account.Mode = ModeHaltAll },
		"entry_gate_latch": func(in *Input) { in.Account.EntryBlockedLatch = true },
		"symbol_allowlist": func(in *Input) { in.Intent.Symbol = "NOTLISTED" },
		"stop_contract":    func(in *Input) { in.Intent.StopPrice = "" },
		"order_size":       func(in *Input) { in.Intent.Quantity = "0" },
		"min_reward_risk":  func(in *Input) { in.Intent.TargetPrice = "10100" },
		"cash":             func(in *Input) { in.Account.CashAvailable = krw("1") },
		"same_day_reentry": func(in *Input) { in.Account.SameDayEntryCount = 99 },
		"open_exposure":    func(in *Input) { in.Account.OpenExposure = krw("999999999") },
		"daily_loss":       func(in *Input) { in.Account.DailyRealizedLoss = krw("999999999") },
		"duplicate_order":  func(in *Input) { in.Account.DuplicateOrder = true },
	}

	steps := EntryChainSteps()
	if len(breakers) != len(steps) {
		t.Fatalf("the chain has %d rungs %v but this test breaks %d; a rung nobody breaks "+
			"is a rung whose reported name nobody checked", len(steps), steps, len(breakers))
	}
	for _, step := range steps {
		breaker, ok := breakers[step]
		if !ok {
			t.Fatalf("no breaker for rung %q", step)
		}
		t.Run(step, func(t *testing.T) {
			in := entryInput()
			breaker(&in)
			got := Evaluate(in)
			if got.Allowed {
				t.Fatalf("the fixture did not break %s; it was allowed", step)
			}
			if got.Step != step {
				t.Errorf("verdict reports step %q, want %q (reason %s: %s)",
					got.Step, step, got.Reason, got.Detail)
			}
		})
	}
}

// TestTheStepIsNotDerivedFromTheReason is the point of the whole field, shown
// with the code that made reason-to-rung inversion unusable: one reason arriving
// from two different rungs, each naming itself.
func TestTheStepIsNotDerivedFromTheReason(t *testing.T) {
	// An unreadable same-day entry count: the same_day_reentry rung raises
	// INPUT_UNAVAILABLE from inside the chain.
	viaRung := entryInput()
	viaRung.Account.SameDayEntryCount = -1
	rung := Evaluate(viaRung)

	// An unusable entry price reaches the same code from preflight, before any
	// rung has run.
	viaPreflight := entryInput()
	viaPreflight.Intent.LimitPrice = "not a number"
	preflight := Evaluate(viaPreflight)

	if rung.Reason != ReasonInputUnavailable || preflight.Reason != ReasonInputUnavailable {
		t.Fatalf("the fixture must produce one reason from two places, got %s and %s",
			rung.Reason, preflight.Reason)
	}
	if rung.Step == preflight.Step {
		t.Fatalf("both refusals report step %q; if the step were derivable from the reason "+
			"this column would be worthless", rung.Step)
	}
	if rung.Step != "same_day_reentry" {
		t.Errorf("an unreadable entry count is refused at same_day_reentry, got %q", rung.Step)
	}
	if preflight.Step != StepPreflight {
		t.Errorf("an unusable entry price is refused before the chain, got %q", preflight.Step)
	}
}

// TestPreflightAndReductionReportDistinctSteps: neither is a rung of the entry
// chain, and recording either as one would put a refusal in the wrong bucket.
// Preflight failures mean the caller handed over something unusable; a reduction
// is not measured by the entry chain at all.
func TestPreflightAndReductionReportDistinctSteps(t *testing.T) {
	chain := map[string]bool{}
	for _, s := range EntryChainSteps() {
		chain[s] = true
	}
	for _, step := range []string{StepPreflight, StepReduction} {
		if chain[step] {
			t.Errorf("%q collides with an entry chain rung name", step)
		}
	}

	noAccount := entryInput()
	noAccount.Intent.AccountRef = ""
	if got := Evaluate(noAccount); got.Step != StepPreflight {
		t.Errorf("a preflight refusal reports %q, want %q", got.Step, StepPreflight)
	}

	oversold := entryInput()
	oversold.Intent.Side = SideSell
	oversold.Intent.Quantity = "10"
	oversold.Account.HeldQuantity = "1"
	got := Evaluate(oversold)
	if got.Allowed {
		t.Fatal("the fixture must refuse")
	}
	if got.Step != StepReduction {
		t.Errorf("a reduction refusal reports %q, want %q", got.Step, StepReduction)
	}
}

// TestAnAllowedVerdictNamesNoStep: the column is "which rung stopped it", so an
// allowed verdict must leave it empty rather than naming the last rung it passed.
// The observation write refuses a row that claims both (journal validate).
func TestAnAllowedVerdictNamesNoStep(t *testing.T) {
	got := Evaluate(entryInput())
	if !got.Allowed {
		t.Fatalf("the fixture must be allowed: %s %s", got.Reason, got.Detail)
	}
	if got.Step != "" {
		t.Errorf("an allowed verdict names step %q, want none", got.Step)
	}

	reduction := entryInput()
	reduction.Intent.Side = SideSell
	reduction.Intent.Quantity = "1"
	reduction.Account.HeldQuantity = "10"
	if r := Evaluate(reduction); !r.Allowed || r.Step != "" {
		t.Errorf("an allowed reduction: allowed=%v step=%q", r.Allowed, r.Step)
	}
}

// TestTheStepIsAddativeToTheVerdict guards the change's own contract from the
// storage side: adding the field must not have altered what Allowed, Reason or
// Detail say. The chain suite as a whole is the real proof (task 3.3, task 7.3);
// this is the local statement of it.
func TestTheStepIsAdditiveToTheVerdict(t *testing.T) {
	in := entryInput()
	in.Intent.TargetPrice = "10100" // below the 2.0 minimum

	got := Evaluate(in)
	switch {
	case got.Allowed:
		t.Fatal("a sub-2.0 gross ratio is still refused; this change alters no verdict")
	case got.Reason != ReasonMinRRNotMet:
		t.Errorf("reason = %s, want MIN_RR_NOT_MET unchanged", got.Reason)
	case !strings.Contains(got.Detail, "below the minimum"):
		t.Errorf("detail = %q, want the unchanged comparison text", got.Detail)
	}
}
