package verifylive

// steps_trigger_test.go drives the one step that means to fill.
//
// Two things are being defended here and they pull against each other. The step
// has to actually work — a measurement that cannot observe a fire is worth
// nothing — and it has to be impossible to reach by accident, because reaching it
// sells a share and there is no undo. So the file opens with the second: what the
// tool does when nobody asked for this.
//
// The endings are enumerated rather than sampled. A step whose failure paths are
// undefined leaves live objects on somebody's account, and every one of these
// tests ends by asking the record what the account is holding.

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// triggerHarness is a KR account with a book tight enough to place a trigger in:
// bid 69800, last 70000, and a 100원 tick, so the first tick above the bid is
// 69900 — strictly inside the interval, which is what makes the basis observable.
func triggerHarness(t *testing.T, script func(*fakeBroker)) *harness {
	t.Helper()
	broker := newFakeBroker().withHolding("005930", 3).withBook("005930", 69800, 70100, 70000)
	if script != nil {
		script(broker)
	}
	return newHarness(t, broker, alwaysConfirm())
}

func seedM0TriggerPrerequisites(t *testing.T, h *harness) {
	t.Helper()
	// M0 is a resume/redo measurement, so it begins with every unrelated
	// verification step durably settled and no outstanding artifact.
	rec, err := OpenRecorder(h.record)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range Steps() {
		if step.ID == StepConditionalTrigger {
			continue
		}
		if err := rec.Append(Entry{Kind: KindStep, StepID: step.ID, Verdict: VerdictPass}); err != nil {
			_ = rec.Close()
			t.Fatal(err)
		}
	}
	if err := rec.Close(); err != nil {
		t.Fatal(err)
	}
}

func triggerOptions(t *testing.T, window time.Duration) Options {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	receipt, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = receipt.Close() })
	return Options{HoldingSymbol: "005930", IncludeTrigger: true, ConfirmEach: true, Resume: true,
		Redo: []StepID{StepConditionalTrigger}, Receipt: receipt, TriggerWindow: window}
}

// --- without the opt-in ----------------------------------------------------------

// TestWithoutTheOptInTheTriggerStepIsUnchanged is the toggle-off case, and it is
// the most important test in this file.
//
// Everything else here describes a step that sells a share. This one says that a
// run which did not ask for it behaves exactly as it did before the step could be
// driven at all: the same three unverified observations, the same deferred
// verdict, no line on the list a person approves, and — the assertion that would
// catch a wiring mistake nothing else would — the price of an order intended to
// fill is never computed, because the book is never even read.
func TestWithoutTheOptInTheTriggerStepIsUnchanged(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) { f.firesOnRead(1, 1, 1) })
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	if got := h.verdict(StepConditionalTrigger); got != VerdictDeferred {
		t.Fatalf("verdict = %q, want %q", got, VerdictDeferred)
	}
	entries := h.entries()
	want := map[string]string{
		"conditional.trigger_observed":           "false",
		"conditional.triggered_order_id_exposed": "unverified",
		"conditional.triggered_order_latency":    "unverified",
	}
	for key, value := range want {
		if !observationEquals(t, entries, StepConditionalTrigger, key, value) {
			t.Errorf("observation %s is not %q; task 2.6 reads these to decide where automatic entry stays "+
				"forbidden, and a run that stopped writing them would shrink that list silently", key, value)
		}
	}
	e, _ := LastEntry(entries, StepConditionalTrigger)
	if e.Mutating {
		t.Error("the record says the step mutated; it sent nothing")
	}
	if len(e.Artifacts) != 0 {
		t.Errorf("the step recorded artifacts %+v without being asked to run", e.Artifacts)
	}
	if n := h.broker.countRequests("GET /orderbook"); n != 0 {
		t.Errorf("the book was read %d time(s); nothing outside the opted-in trigger step needs a bid, and "+
			"reading one is the first step towards pricing an order that can fill", n)
	}
	if n := h.broker.countRequests("POST /conditional-orders 005930 key=TRIGGER-"); n != 0 {
		t.Errorf("%d trigger conditional(s) were registered without the opt-in", n)
	}

	// And the operator was never shown a line for it.
	for _, b := range h.op.batches {
		for _, m := range b.Plan.Mutations {
			if m.Step == StepConditionalTrigger {
				t.Errorf("the approval list carried %s for a step that will not run: %s", m.Kind, m.HeadlineKO())
			}
		}
	}
}

// --- the measurement ---------------------------------------------------------------

// TestTheTriggerStepObservesAFireAndItsChildFilling is the happy path, which is
// also the path that sells a share.
func TestTheTriggerStepObservesAFireAndItsChildFilling(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) {
		// A trade prints down to the trigger, then it fires, then the child appears
		// and fills: the ordinary last-trade shape.
		f.dropAfterReads = 2
		f.firesOnRead(3, 4, 2)
	})
	runToCompletion(t, h, triggerOptions(t, 2*time.Minute))

	if got := h.verdict(StepConditionalTrigger); got != VerdictPass {
		e, _ := LastEntry(h.entries(), StepConditionalTrigger)
		t.Fatalf("verdict = %q (%s), want pass", got, e.Reason)
	}
	entries := h.entries()

	for _, key := range []string{
		"conditional.trigger.condition_crossed_at",
		"conditional.trigger.trigger_first_observed_at",
		"conditional.trigger.triggered_order_id_first_seen_at",
		"conditional.trigger.child_order_filled_at",
	} {
		value, ok := h.observation(StepConditionalTrigger, key)
		if !ok || value == "unobserved" {
			t.Errorf("%s = %q; all four timestamps have to be observed on the happy path", key, value)
			continue
		}
		if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
			t.Errorf("%s = %q, which is not a timestamp: %v", key, value, err)
		}
		// The error bound is the point. A stamp with no interval beside it cannot
		// be read afterwards, because the broker gives no time of its own (M44).
		if detail := observationDetail(t, entries, StepConditionalTrigger, key); !strings.Contains(detail, "±") {
			t.Errorf("%s carries no error bound (%q)", key, detail)
		}
	}

	if !observationEquals(t, entries, StepConditionalTrigger, "conditional.trigger_observed", "true") {
		t.Error("the trigger was not recorded as observed")
	}
	if !observationEquals(t, entries, StepConditionalTrigger, "conditional.triggered_order_id_exposed", "true") {
		t.Error("triggeredOrderId exposure was not recorded")
	}
	if v, _ := h.observation(StepConditionalTrigger, "conditional.triggered_order_latency"); v == "unverified" {
		t.Error("the latency between the trigger and its child identifier is still unverified after a fire")
	}
	if v, _ := h.observation(StepConditionalTrigger, "conditional.trigger.book_at_trigger"); !strings.Contains(v, "bid") {
		t.Errorf("book at the trigger = %q, want the top of book that fired it — it is the only evidence "+
			"that narrows what the broker evaluated", v)
	}

	// The account is clean, and the record says how each object ended.
	if out := Outstanding(entries); len(out) != 0 {
		t.Errorf("still outstanding: %+v — a filled child and a fired conditional are both gone", out)
	}
	child := lastArtifactOfKind(t, entries, KindOrder)
	if !child.Filled || child.Cancelled {
		t.Errorf("child order = %+v, want Filled and not Cancelled: recording the fill as a cancellation "+
			"would make this measurement's own conclusion read 'we cancelled it'", child)
	}
	if child.ChainID == "" {
		t.Error("the child order carries no chain, so the record cannot say which conditional produced it")
	}
}

// TestTheTriggerBasisIsDecidedByTheOrderOfTwoObservations.
//
// The placement exists to answer one question — does the broker evaluate a
// conditional against the last trade or against the best bid — and the answer is
// which of two things was seen first, not how long either took. An ordering
// survives a coarse polling interval; a latency comparison would not.
func TestTheTriggerBasisIsDecidedByTheOrderOfTwoObservations(t *testing.T) {
	t.Run("fires before any trade prints down to it", func(t *testing.T) {
		h := triggerHarness(t, func(f *fakeBroker) {
			// No drop scripted at all: the last trade never reaches the trigger and
			// the thing fires anyway, which it can only do off the bid.
			f.firesOnRead(2, 3, 2)
		})
		runToCompletion(t, h, triggerOptions(t, 2*time.Minute))
		if got, _ := h.observation(StepConditionalTrigger, "conditional.trigger.basis"); got != "bid" {
			t.Errorf("basis = %q, want bid", got)
		}
	})

	t.Run("fires after a trade prints down to it", func(t *testing.T) {
		h := triggerHarness(t, func(f *fakeBroker) {
			f.dropAfterReads = 2
			f.firesOnRead(4, 5, 2)
		})
		runToCompletion(t, h, triggerOptions(t, 2*time.Minute))
		if got, _ := h.observation(StepConditionalTrigger, "conditional.trigger.basis"); got != "last-trade" {
			t.Errorf("basis = %q, want last-trade", got)
		}
	})
}

// --- the endings -------------------------------------------------------------------

// TestTheMarketNeverComingIsNotAFailure.
//
// The step cancels what it registered, inside the approval window that registered
// it — which is what the other eleven mutating steps do with their own objects and
// is a different thing from the clock-based lease the cleanup prologue was
// deliberately not given. The verdict is a skip, not a failure: a wait that ran out
// means the market did not move, and calling that a failed measurement would put a
// false negative into the evidence.
func TestTheMarketNeverComingIsNotAFailure(t *testing.T) {
	h := triggerHarness(t, nil) // nothing scripted: it never fires
	runToCompletion(t, h, triggerOptions(t, 20*time.Second))

	if got := h.verdict(StepConditionalTrigger); got != VerdictSkipped {
		t.Fatalf("verdict = %q, want skipped", got)
	}
	e, _ := LastEntry(h.entries(), StepConditionalTrigger)
	if !strings.Contains(e.Reason, "INCONCLUSIVE") {
		t.Errorf("reason = %q, want it to say the measurement was inconclusive", e.Reason)
	}
	if !observationEquals(t, h.entries(), StepConditionalTrigger,
		"conditional.trigger.cancel_race_recheck", "clean") {
		t.Error("the step did not re-read after its own cancel")
	}
	if out := Outstanding(h.entries()); len(out) != 0 {
		t.Errorf("still outstanding: %+v — the step must leave nothing behind when it gives up", out)
	}
}

// TestACancelThatRacesATriggerDoesNotReportThatNothingHappened.
//
// The worst possible ending: the broker has already sold a share and the tool
// records that the market never came. The step cancels on the way out and then
// reads once more precisely to catch this, and what it finds sends it back to
// watching rather than to a verdict.
func TestACancelThatRacesATriggerDoesNotReportThatNothingHappened(t *testing.T) {
	// The trigger wins at the instant of the cancel, which is the hard version: the
	// conditional is gone, so re-reading it says exactly what a clean cancel says.
	// Only the position can tell the two apart.
	h := triggerHarness(t, func(f *fakeBroker) { f.fireOnCancel = true })
	runToCompletion(t, h, triggerOptions(t, 15*time.Second))

	entries := h.entries()
	if !observationEquals(t, entries, StepConditionalTrigger,
		"conditional.trigger.cancel_race_recheck", "fired") {
		got, _ := h.observation(StepConditionalTrigger, "conditional.trigger.cancel_race_recheck")
		t.Fatalf("cancel_race_recheck = %q, want fired — the step ended before noticing the trigger", got)
	}
	if got := h.verdict(StepConditionalTrigger); got == VerdictSkipped {
		t.Error("the step ended INCONCLUSIVE although the conditional had fired")
	}
	if !observationEquals(t, entries, StepConditionalTrigger, "conditional.trigger_observed", "true") {
		t.Error("the trigger was not recorded as observed after the race")
	}
}

// TestATriggerWithNoObservableFillLeavesTheChildHeld.
//
// The child order a conditional produces has to be allowed to fill — not
// cancelling it is the measurement rather than a leak — and this is the path where
// that matters most, because it is the one where the step failed. sweepStep
// cancels what a step left resting; it must not reach this.
func TestATriggerWithNoObservableFillLeavesTheChildHeld(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) {
		f.firesOnRead(2, 3, 0) // fires and links, never reports filled
	})
	opts := triggerOptions(t, 2*time.Minute)
	seedM0TriggerPrerequisites(t, h)
	// The run must NOT report a clean finish. The conditional looks like it fired
	// and the step could not prove it, so it stays on the record as live — and a
	// stop that can fire, sitting on a real holding with nobody watching, is
	// exactly what the end-of-run check is for.
	second, err := h.run(opts)
	if err == nil {
		t.Error("the run reported a clean finish although a fire-capable conditional was still live")
	}
	if len(second.Outstanding) == 0 {
		t.Error("the summary told the operator nothing was left")
	}

	if got := h.verdict(StepConditionalTrigger); got != VerdictFail {
		t.Fatalf("verdict = %q, want fail — an unobserved fill is a measured negative", got)
	}
	if entry, ok := LastEntry(h.entries(), StepConditionalTrigger); !ok || strings.Contains(entry.Reason, "verify abort") || !strings.Contains(entry.Reason, "수동") {
		t.Fatalf("unobserved-fill guidance = %+v, want manual reconciliation without verify abort", entry)
	}
	entries := h.entries()
	var out []Artifact
	for _, a := range Outstanding(entries) {
		if a.Kind == KindOrder {
			out = append(out, a)
		}
	}
	if len(out) != 1 {
		t.Fatalf("outstanding orders = %+v, want exactly the child order, still held", out)
	}
	child := out[0]
	if !child.Deliberate {
		t.Error("the child order is outstanding but not marked deliberate; every screen would call it a leak")
	}
	if child.HeldUntil != StepConditionalTrigger {
		t.Errorf("HeldUntil = %q, want %q — without it the next run's prologue offers to cancel the evidence",
			child.HeldUntil, StepConditionalTrigger)
	}
	// M0 checkpoints keep both members of the observed pair visible for manual
	// reconciliation, but out of every automatic cancellation target list.
	if targets := PendingCleanup(entries); len(targets) != 0 {
		t.Fatalf("PendingCleanup = %+v, want no automatic M0 cancellation", targets)
	}
	if targets := AbortTargets(entries); len(targets) != 0 {
		t.Fatalf("AbortTargets = %+v, want M0 manual reconciliation only", targets)
	}
	runner := h.runner(t, Options{})
	result, abortErr := runner.Abort(context.Background(), "")
	if abortErr != nil || len(result.Targets) != 0 {
		t.Fatalf("Abort = %+v err=%v, want zero M0 mutations", result, abortErr)
	}
	if n := h.broker.countRequests("POST /orders/" + child.ID + "/cancel"); n != 0 {
		t.Errorf("the child order was cancelled %d time(s). Letting it fill IS the measurement", n)
	}
	if n := h.broker.countRequests("DELETE /conditional-orders/"); n != 0 {
		t.Errorf("the observed parent conditional was cancelled %d time(s); M0 requires manual reconciliation", n)
	}
}

// TestAStopWhoseConditionWasMetAndDidNotFireIsAFailure.
//
// The most consequential negative this step can produce, and the one an earlier
// draft folded into "the market never came". They are not the same result: one
// says the price did not move, the other says the price moved and the broker's
// protective order did nothing. If 2c's protection ledger is going to rest on
// conditional orders, this is the observation that would say it cannot.
func TestAStopWhoseConditionWasMetAndDidNotFireIsAFailure(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) {
		f.dropAfterReads = 2 // the price reaches the trigger
		// and nothing is scripted to fire.
	})
	runToCompletion(t, h, triggerOptions(t, 5*time.Minute))

	if got := h.verdict(StepConditionalTrigger); got != VerdictFail {
		e, _ := LastEntry(h.entries(), StepConditionalTrigger)
		t.Fatalf("verdict = %q (%s), want fail — a stop that ignored its own condition is not an "+
			"inconclusive measurement", got, e.Reason)
	}
	if !observationEquals(t, h.entries(), StepConditionalTrigger,
		"conditional.fires_when_its_condition_is_met", "false") {
		t.Error("the observation 2c would have to read before trusting a conditional order was not recorded")
	}
	// It still cleans up after itself: a measured negative is not a licence to
	// leave a live stop behind.
	if out := Outstanding(h.entries()); len(out) != 0 {
		t.Errorf("still outstanding: %+v", out)
	}
}

// TestTheTriggerStepSkipsWhenThereIsNothingLeftToSell.
//
// A one-share holding with the register step's stop already reserving it. The
// answer has to come before anything is registered, which is also where the
// fractional-remainder question (issues.md J3) surfaces: before a sale, not after.
func TestTheTriggerStepSkipsWhenThereIsNothingLeftToSell(t *testing.T) {
	// One share, which the register step's stop reserves. There is nothing left for
	// a second stop to sell, and the step has to notice before it registers one.
	broker := newFakeBroker().withHolding("005930", 1).withBook("005930", 69800, 70100, 70000)
	broker.sellable["005930"] = 0 // an independently held reservation; M0 must not create another stop.
	h := newHarness(t, broker, alwaysConfirm())
	runToCompletion(t, h, triggerOptions(t, 20*time.Second))

	if got := h.verdict(StepConditionalTrigger); got != VerdictSkipped {
		t.Fatalf("verdict = %q, want skipped", got)
	}
	if n := h.broker.countRequests("POST /conditional-orders 005930 key=TRIGGER-"); n != 0 {
		t.Errorf("%d conditional(s) were registered on an account with nothing to sell", n)
	}
}

// TestTheTriggerStepSkipsWhenTheGridHasNoRoom. A last trade less than one tick
// above the bid is a real shape in US, where a trade can print inside the quoted
// spread (M49). There is nowhere to put a trigger, and the step must not round its
// way to one.
func TestTheTriggerStepSkipsWhenTheGridHasNoRoom(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) { f.withBook("005930", 70000, 70100, 70000) })
	runToCompletion(t, h, triggerOptions(t, 20*time.Second))

	if got := h.verdict(StepConditionalTrigger); got != VerdictSkipped {
		t.Fatalf("verdict = %q, want skipped", got)
	}
	if n := h.broker.countRequests("POST /conditional-orders 005930 key=TRIGGER-"); n != 0 {
		t.Error("a conditional was registered although there was no valid trigger price")
	}
}

// --- the exposure cap ----------------------------------------------------------------

// TestTheTriggerStepMayHoldASecondConditionalAndNothingElseMay.
//
// The trigger step registers its own stop rather than moving the one the register
// step left alive, so on a full run there are briefly two. That is the only
// exception, and it is scoped to the step by the same mechanism the validity-window
// step's order cap uses.
func TestTheTriggerStepMayHoldASecondConditionalAndNothingElseMay(t *testing.T) {
	h := triggerHarness(t, nil)
	seedM0TriggerPrerequisites(t, h)
	r := h.runner(t, triggerOptions(t, time.Minute))
	r.prior = []Entry{{StepID: StepConditionalRegister, Artifacts: []Artifact{
		{Kind: KindConditional, ID: "co-live", Symbol: "005930", CreatedAt: time.Now()},
	}}}

	if err := r.checkConditionalCap(&stepRun{step: mustStep(t, StepConditionalTrigger)}); err != nil {
		t.Errorf("the trigger step was refused its own conditional: %v", err)
	}
	if err := r.checkConditionalCap(&stepRun{step: mustStep(t, StepConditionalRegister)}); err == nil {
		t.Error("the register step was allowed a second conditional; the cap must stay at one everywhere else")
	}

	// And two is the ceiling even for the trigger step.
	r.prior = append(r.prior, Entry{StepID: StepConditionalTrigger, Artifacts: []Artifact{
		{Kind: KindConditional, ID: "co-trigger", Symbol: "005930", CreatedAt: time.Now()},
	}})
	if err := r.checkConditionalCap(&stepRun{step: mustStep(t, StepConditionalTrigger)}); err == nil {
		t.Error("a third conditional was allowed")
	}
}

// TestTheOrderCapIsUntouchedByTheTriggerStep. The child order never passes
// checkOrderCap — this tool does not place it, the broker creates it and the step
// discovers it — so raising the order cap here would have widened an exposure
// nothing was asking for. Once it fills it stops being outstanding at all.
func TestTheOrderCapIsUntouchedByTheTriggerStep(t *testing.T) {
	if MaxLiveOrders != 1 {
		t.Fatalf("MaxLiveOrders = %d, want 1", MaxLiveOrders)
	}
	h := triggerHarness(t, nil)
	seedM0TriggerPrerequisites(t, h)
	r := h.runner(t, triggerOptions(t, time.Minute))
	r.prior = []Entry{{StepID: StepOrderCancel, Artifacts: []Artifact{
		{Kind: KindOrder, ID: "ord-live", Symbol: "005930", CreatedAt: time.Now()},
	}}}
	if err := r.checkOrderCap(&stepRun{step: mustStep(t, StepConditionalTrigger)}); err == nil {
		t.Error("the trigger step was granted a raised order cap; nothing in it places an order")
	}
}

// --- helpers -------------------------------------------------------------------------

// lastArtifactOfKind returns the newest line written about any artifact of a kind.
func lastArtifactOfKind(t *testing.T, entries []Entry, kind string) Artifact {
	t.Helper()
	var out Artifact
	found := false
	for _, e := range entries {
		for _, a := range e.Artifacts {
			if a.Kind == kind {
				out, found = a, true
			}
		}
	}
	if !found {
		t.Fatalf("the record holds no %s artifact at all", kind)
	}
	return out
}
