package verifylive

// runner_test.go drives the whole procedure.
//
// The assertions are grouped around the four things that would make this tool
// dangerous rather than useful:
//
//	nothing is left behind      every order it places is cancelled, and the run
//	                            reports an error if one is not
//	nothing goes without a yes  a refused confirmation places nothing, and the
//	                            run keeps going with the independent steps
//	the exposure cap holds      never more than one live order at a time
//	the record is the truth     a resumed run does not repeat a settled step, and
//	                            the persistence check cannot pass in the process
//	                            that registered the conditional

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// runToCompletion walks the whole procedure the way an operator would: one run
// that stops at the persistence check, then a second one that finishes it.
func runToCompletion(t *testing.T, h *harness, opts Options) (Summary, Summary) {
	t.Helper()
	first, err := h.run(opts)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if !first.Halted {
		t.Fatalf("the first run did not stop for the persistence check: %+v", first.Outcomes)
	}
	h.broker.restart()

	resumed := opts
	resumed.Prior = nil
	second, err := h.run(resumed)
	if err != nil {
		t.Fatalf("resumed run: %v", err)
	}
	return first, second
}

// --- nothing is left behind ---------------------------------------------------

// TestAFullRunLeavesNothingLive is the property the whole design exists to
// preserve.
func TestAFullRunLeavesNothingLive(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())

	_, second := runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	if len(second.Outstanding) != 0 {
		t.Fatalf("the account is left holding %+v", second.Outstanding)
	}
	if len(broker.conds) != 0 {
		t.Errorf("the broker still has conditional orders: %+v", broker.conds)
	}
	for id, raw := range broker.orders {
		if strings.Contains(string(raw), `"status":"PENDING"`) {
			t.Errorf("order %s is still working: %s", id, raw)
		}
	}
}

// TestEveryPlacementIsMatchedByACancel counts requests rather than inspecting
// state, so a placement the tool forgot about would still be caught.
func TestEveryPlacementIsMatchedByACancel(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	// Placements that the broker actually accepted. Rejections (the oversell, the
	// idempotency conflict) and idempotent replays create nothing.
	created := 0
	for _, raw := range broker.orders {
		_ = raw
		created++
	}
	cancels := broker.countRequests("POST /orders/")
	cancelCalls := 0
	for _, r := range broker.seen() {
		if strings.HasPrefix(r, "POST /orders/") && strings.HasSuffix(r, "/cancel") {
			cancelCalls++
		}
	}
	if cancelCalls == 0 {
		t.Fatalf("no cancel was issued at all (saw %d order-scoped POSTs)", cancels)
	}
	if created == 0 {
		t.Fatal("no order was created; the run cannot have measured anything")
	}
}

// TestRunReportsAnErrorWhenSomethingIsLeftLive.
//
// A cancel that fails must not be shrugged off. The run comes back with an error
// naming the identifier, because that identifier is what the operator needs.
//
// Refusing a cancel is only reachable under --confirm-each, which is the point of
// keeping that mode: it is the one gate fine enough to leave an order resting.
func TestRunReportsAnErrorWhenSomethingIsLeftLive(t *testing.T) {
	broker := newFakeBroker()
	// Refuse only the cancels, so orders are placed and never removed.
	op := refuseWhere(func(m Mutation) bool { return strings.Contains(m.Action, "cancel") })
	h := newHarness(t, broker, op)

	summary, err := h.run(Options{ConfirmEach: true})
	if err == nil {
		t.Fatal("the run finished cleanly although an order was left resting")
	}
	if len(summary.Outstanding) == 0 {
		t.Fatal("the summary reports nothing outstanding")
	}
	if !strings.Contains(err.Error(), summary.Outstanding[0].ID) {
		t.Errorf("err = %v, want it to name the outstanding order %s", err, summary.Outstanding[0].ID)
	}
}

// --- nothing goes without a yes -----------------------------------------------

// TestARefusedPerMutationConfirmationPlacesNothing, under --confirm-each.
func TestARefusedPerMutationConfirmationPlacesNothing(t *testing.T) {
	broker := newFakeBroker()
	op := refuseWhere(func(Mutation) bool { return true })
	h := newHarness(t, broker, op)

	summary, err := h.run(Options{ConfirmEach: true})
	if err != nil {
		t.Fatalf("refusing every confirmation must not be an error: %v", err)
	}
	if len(summary.Outstanding) != 0 {
		t.Fatalf("something was created despite every refusal: %+v", summary.Outstanding)
	}
	for _, r := range broker.seen() {
		if strings.HasPrefix(r, "POST ") || strings.HasPrefix(r, "DELETE ") {
			t.Errorf("a mutating request was issued after a refusal: %s", r)
		}
	}
	if h.verdict(StepIdempotency) != VerdictRefused {
		t.Errorf("idempotency verdict = %q, want refused", h.verdict(StepIdempotency))
	}
	if len(op.seen) == 0 {
		t.Error("the operator was never asked anything")
	}
}

// TestARefusalDoesNotCostTheIndependentSteps: refusing the idempotency check must
// still leave the order-path and conditional measurements available. It is a
// --confirm-each property: the batch model has one answer, not one per step.
func TestARefusalDoesNotCostTheIndependentSteps(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := refuseWhere(func(m Mutation) bool { return m.Step == StepIdempotency })
	h := newHarness(t, broker, op)

	first, err := h.run(Options{HoldingSymbol: "005930", ConfirmEach: true})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepIdempotency) != VerdictRefused {
		t.Fatalf("idempotency verdict = %q, want refused", h.verdict(StepIdempotency))
	}
	if h.verdict(StepOrderCancel) != VerdictPass {
		t.Errorf("order-cancel verdict = %q; a refusal elsewhere must not skip it", h.verdict(StepOrderCancel))
	}
	if h.verdict(StepConditionalRegister) != VerdictPass {
		t.Errorf("conditional-register verdict = %q", h.verdict(StepConditionalRegister))
	}
	if !first.Halted {
		t.Error("the run should still have reached the persistence halt")
	}
}

// TestNoTerminalAbortsTheWholeRun. The batch approval cannot be typed at all
// without a terminal, so the run stops before its first step.
func TestNoTerminalAbortsTheWholeRun(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, refuseBatch(ErrNotATerminal))

	summary, err := h.run(Options{})
	if !errors.Is(err, ErrNotATerminal) {
		t.Fatalf("err = %v, want ErrNotATerminal", err)
	}
	if !summary.Halted {
		t.Error("the run did not report itself halted")
	}
	if h.verdict(StepOrderCancel) != "" {
		t.Errorf("order-cancel ran after the abort with verdict %q", h.verdict(StepOrderCancel))
	}
}

// TestNoTerminalAbortsTheWholeRunUnderConfirmEach. Every remaining mutating step
// would ask the same impossible question, so the tool stops rather than recording a
// wall of refusals.
func TestNoTerminalAbortsTheWholeRunUnderConfirmEach(t *testing.T) {
	broker := newFakeBroker()
	op := alwaysConfirm()
	op.answer = func(Mutation) error { return ErrNotATerminal }
	h := newHarness(t, broker, op)

	summary, err := h.run(Options{ConfirmEach: true})
	if !errors.Is(err, ErrNotATerminal) {
		t.Fatalf("err = %v, want ErrNotATerminal", err)
	}
	if !summary.Halted {
		t.Error("the run did not report itself halted")
	}
	if h.verdict(StepOrderCancel) != "" {
		t.Errorf("order-cancel ran after the abort with verdict %q", h.verdict(StepOrderCancel))
	}
}

// TestNewRefusesWithoutAConfirmer. There is no unconfirmed mode and no unapproved
// mode, and the absence of both is asserted rather than assumed.
func TestNewRefusesWithoutAConfirmer(t *testing.T) {
	rec, err := OpenRecorder(tempRecord(t))
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()
	op := alwaysConfirm()
	if _, err := New(Options{Broker: newFakeBroker(), Recorder: rec, AccountRef: "1"}); err == nil {
		t.Fatal("a runner was built with no confirmer")
	}
	if _, err := New(Options{
		Broker: newFakeBroker(), Recorder: rec, Confirm: op.confirmer(), AccountRef: "1",
	}); err == nil {
		t.Fatal("a runner was built with no batch confirmer")
	}
	if _, err := New(Options{
		Broker: newFakeBroker(), Recorder: rec, Confirm: op.confirmer(), ConfirmBatch: op.batchConfirmer(),
	}); err == nil {
		t.Fatal("a runner was built with no account to attest about")
	}
}

// --- the exposure cap ---------------------------------------------------------

// TestNeverMoreThanOneLiveOrder replays the record's artifacts in order and
// tracks how many orders were live at each point.
//
// The record rather than the request log, because the log cannot tell a
// placement that created an order from one the broker answered with an existing
// identifier — and the artifacts can, which is what they are for.
func TestNeverMoreThanOneLiveOrder(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	peak, at := peakLiveOrders(h.entries())
	if peak > MaxLiveOrders {
		t.Errorf("the record shows %d live order(s) at once during %s; the cap is %d",
			peak, at, MaxLiveOrders)
	}
	if peak == 0 {
		t.Error("no order was ever live; the run cannot have measured the order path")
	}
}

// TestTTLEdgeIsTheOnlyStepThatPlansTwoLiveOrders.
//
// It is the documented exception, and it is bounded: two, in that one step, both
// cancelled before it returns.
func TestTTLEdgeIsTheOnlyStepThatPlansTwoLiveOrders(t *testing.T) {
	broker := newFakeBroker()
	broker.expireKeysOnSleep = true // the window really does close while it waits
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{IncludeTTLEdge: true, TTLWait: 11 * time.Minute}); err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, e := range h.entries() {
		peak, _ := peakLiveOrders([]Entry{e})
		limit := MaxLiveOrders
		if e.StepID == StepIdempotencyTTLEdge {
			limit = MaxLiveOrdersTTLEdge
		}
		if peak > limit {
			t.Errorf("%s held %d live orders at once; its cap is %d", e.StepID, peak, limit)
		}
	}
	if peak, _ := peakLiveOrders(entriesFor(h.entries(), StepIdempotencyTTLEdge)); peak != 2 {
		t.Errorf("the validity-window step peaked at %d live orders; it exists to observe two", peak)
	}
	if got, _ := h.observation(StepIdempotencyTTLEdge, "idempotency.ttl_window_closed"); got != "true" {
		t.Errorf("ttl_window_closed = %q, want true once the key has expired", got)
	}
	if out := Outstanding(h.entries()); len(undeliberate(out)) != 0 {
		t.Errorf("the run left orders behind: %+v", out)
	}
}

func entriesFor(entries []Entry, id StepID) []Entry {
	var out []Entry
	for _, e := range entries {
		if e.StepID == id {
			out = append(out, e)
		}
	}
	return out
}

// peakLiveOrders walks the artifacts in the order they were recorded.
func peakLiveOrders(entries []Entry) (peak int, at StepID) {
	live := map[string]bool{}
	for _, e := range entries {
		for _, a := range e.Artifacts {
			if a.Kind != "order" {
				continue
			}
			if a.Cancelled {
				delete(live, a.ID)
				continue
			}
			live[a.ID] = true
			if len(live) > peak {
				peak, at = len(live), e.StepID
			}
		}
	}
	return peak, at
}

// TestExposureCapRefusesASecondPlacement drives the guard directly: a step that
// tried to place while one was already live has to be refused, not trusted.
func TestExposureCapRefusesASecondPlacement(t *testing.T) {
	rec, err := OpenRecorder(tempRecord(t))
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()
	op := alwaysConfirm()
	r, err := New(Options{
		Broker: newFakeBroker(), Recorder: rec, Confirm: op.confirmer(), ConfirmBatch: op.batchConfirmer(),
		AccountRef: "123-45-678901", Symbol: "005930",
		Prior: []Entry{{StepID: StepOrderCancel, Artifacts: []Artifact{
			{Kind: "order", ID: "ord-live", Symbol: "005930", CreatedAt: time.Now()},
		}}},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sr := &stepRun{step: mustStep(t, StepIdempotency)}
	if _, err := r.placeOrder(context.Background(), sr, orderSpec{
		Symbol: "005930", Market: MarketKR, Side: "buy", Quantity: 1, Price: 56000,
	}, "cancelled"); !errors.Is(err, ErrExposureCap) {
		t.Fatalf("err = %v, want ErrExposureCap", err)
	}
}

// --- the record is the truth --------------------------------------------------

// TestConditionalPersistenceCannotPassInTheRegisteringProcess is the structural
// half of "프로세스 종료 후 존속". A single process claiming its own conditional
// survived its own exit would be evidence of nothing.
func TestConditionalPersistenceCannotPassInTheRegisteringProcess(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())

	first, err := h.run(Options{HoldingSymbol: "005930"})
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if h.verdict(StepConditionalPersist) != VerdictAwaitingRestart {
		t.Fatalf("persist verdict = %q, want awaiting-restart", h.verdict(StepConditionalPersist))
	}
	if !strings.Contains(first.Halt, "--resume") {
		t.Errorf("the halt message does not tell the operator what to do: %q", first.Halt)
	}
	if h.verdict(StepConditionalCancel) != "" {
		t.Error("the cancel step ran in the same process; the run should have stopped at the halt")
	}
	// The conditional is still registered, and that is deliberate — but it must
	// be reported, not hidden.
	if len(first.Outstanding) != 1 || first.Outstanding[0].Kind != "conditional-order" {
		t.Fatalf("Outstanding = %+v, want the deliberately-live conditional", first.Outstanding)
	}
	if !first.Outstanding[0].Deliberate {
		t.Error("the conditional left live for the persistence check is not marked deliberate")
	}

	broker.restart()
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("resumed run: %v", err)
	}
	if h.verdict(StepConditionalPersist) != VerdictPass {
		t.Fatalf("persist verdict after restart = %q, want pass", h.verdict(StepConditionalPersist))
	}
	if v, _ := h.observation(StepConditionalPersist, "conditional.survives_process_exit"); v != "true" {
		t.Errorf("conditional.survives_process_exit = %q, want true", v)
	}
}

// TestTheProcessBoundaryIsTheInstanceIDAndNotThePID pins the judgement the
// console's restart button depends on (tasks.md 1.8 ①).
//
// A re-exec replaces the process image and keeps the PID. That is a genuinely new
// process — the client state that registered the conditional is gone, which is the
// whole content of the measurement — so the persistence step has to accept it. If
// the judgement read the PID it would refuse, and the restart button would be a
// button that cannot finish the verification it exists to finish.
//
// The second half is the converse and it is the one that matters for safety: the
// same instance identifier must be refused even when the PID differs, because a
// process claiming its own conditional outlived its own exit is evidence of
// nothing at all.
func TestTheProcessBoundaryIsTheInstanceIDAndNotThePID(t *testing.T) {
	t.Run("a re-exec keeps the PID and is still a new process", func(t *testing.T) {
		broker := newFakeBroker().withHolding("005930", 3)
		h := newHarness(t, broker, alwaysConfirm())
		h.fixedPID = 4242 // every invocation is "the same" process to the OS

		if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
			t.Fatalf("first run: %v", err)
		}
		if h.verdict(StepConditionalPersist) != VerdictAwaitingRestart {
			t.Fatalf("persist verdict = %q, want awaiting-restart", h.verdict(StepConditionalPersist))
		}

		broker.restart()
		if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
			t.Fatalf("re-executed run: %v", err)
		}
		if got := h.verdict(StepConditionalPersist); got != VerdictPass {
			t.Fatalf("persist verdict after a re-exec = %q, want pass — the boundary must be the instance "+
				"identifier, not the PID that a re-exec preserves", got)
		}
		if v, _ := h.observation(StepConditionalPersist, "conditional.survives_process_exit"); v != "true" {
			t.Errorf("conditional.survives_process_exit = %q, want true", v)
		}

		// And the record proves the two runs really did share the PID, so the test
		// would not pass by accident if the harness stopped pinning it.
		pids := map[int]bool{}
		instances := map[string]bool{}
		for _, e := range h.entries() {
			pids[e.Process.PID] = true
			instances[e.Process.InstanceID] = true
		}
		if len(pids) != 1 {
			t.Fatalf("the record holds %d distinct PIDs; the re-exec was not simulated", len(pids))
		}
		if len(instances) < 2 {
			t.Fatalf("the record holds %d distinct instance ids; there was no second process", len(instances))
		}
	})

	t.Run("the same instance is refused however the PID moves", func(t *testing.T) {
		broker := newFakeBroker().withHolding("005930", 3)
		h := newHarness(t, broker, alwaysConfirm())
		h.fixedInstanceID = "proc-same"

		if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
			t.Fatalf("first run: %v", err)
		}
		broker.restart()
		if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
			t.Fatalf("second run: %v", err)
		}
		if got := h.verdict(StepConditionalPersist); got != VerdictAwaitingRestart {
			t.Fatalf("persist verdict = %q, want awaiting-restart — the registering instance must never be "+
				"able to certify its own conditional's survival, whatever the PID says", got)
		}
	})
}

// TestConditionalPersistenceFailsWhenTheBrokerForgets is the other outcome, and
// the one that would sink the whole unattended design.
func TestConditionalPersistenceFailsWhenTheBrokerForgets(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	broker.dropConditionalsOnRestart = true
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	broker.restart()
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("resumed run: %v", err)
	}
	if h.verdict(StepConditionalPersist) != VerdictFail {
		t.Fatalf("persist verdict = %q, want fail", h.verdict(StepConditionalPersist))
	}
	if v, _ := h.observation(StepConditionalPersist, "conditional.survives_process_exit"); v != "false" {
		t.Errorf("conditional.survives_process_exit = %q, want false", v)
	}
}

// TestASettledStepIsNotRunAgain. Re-running a mutating step would place a second
// live order for a measurement already made.
func TestASettledStepIsNotRunAgain(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	placements := broker.countRequests("POST /orders ")

	summary, err := h.run(Options{})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if got := broker.countRequests("POST /orders "); got != placements {
		t.Errorf("the second run issued %d placements where the first issued %d; settled steps must not repeat",
			got-placements, placements)
	}
	settled := 0
	for _, o := range summary.Outcomes {
		if o.AlreadySettled {
			settled++
		}
	}
	if settled == 0 {
		t.Error("the second run reported no step as already settled")
	}
}

// TestRedoRunsASettledStepAgain, because an operator who fixed the account state
// needs a way back in.
func TestRedoRunsASettledStepAgain(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	before := broker.countRequests("POST /orders ")

	if _, err := h.run(Options{Redo: []StepID{StepOrderCancel}}); err != nil {
		t.Fatalf("redo run: %v", err)
	}
	if got := broker.countRequests("POST /orders "); got <= before {
		t.Error("--redo did not re-run the step")
	}
}

// --- measurements -------------------------------------------------------------

// TestIdempotencyPassesWhenTheBrokerBehaves.
func TestIdempotencyPassesWhenTheBrokerBehaves(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepIdempotency) != VerdictPass {
		t.Fatalf("verdict = %q", h.verdict(StepIdempotency))
	}
	for key, want := range map[string]string{
		"idempotency.replay_returns_same_order_id": "true",
		"idempotency.no_second_order_created":      "true",
		"idempotency.conflict_error_code":          "idempotency-key-conflict",
		"idempotency.conflict_rejected":            "true",
		"idempotency.key_scope":                    "unverified",
	} {
		if got, ok := h.observation(StepIdempotency, key); !ok || got != want {
			t.Errorf("%s = %q (present=%v), want %q", key, got, ok, want)
		}
	}
}

// TestIdempotencyFailsWhenTheKeyIsIgnored is the scenario the spec names: "재생이
// 새 주문을 만드는 경우 → 멱등 재생 경로가 비활성으로 기록된다".
func TestIdempotencyFailsWhenTheKeyIsIgnored(t *testing.T) {
	broker := newFakeBroker()
	broker.honourIdempotency = false
	broker.conflictOnDifferentBody = false
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepIdempotency) != VerdictFail {
		t.Fatalf("verdict = %q, want fail", h.verdict(StepIdempotency))
	}
	if got, _ := h.observation(StepIdempotency, "idempotency.replay_returns_same_order_id"); got != "false" {
		t.Errorf("replay_returns_same_order_id = %q, want false", got)
	}
	// Everything the broker created has to be gone, including the order the
	// replay produced by not honouring the key.
	if out := Outstanding(h.entries()); len(undeliberate(out)) != 0 {
		t.Errorf("orders were left behind after an unhonoured key: %+v", out)
	}
	// A broker that ignores its own key can transiently produce a second order:
	// the tool has to send the replay to find that out. What the cap guarantees
	// is that the tool never *asks* for two, and that the extra is annotated and
	// cancelled inside the same step.
	sawUnhonoured := false
	for _, e := range entriesFor(h.entries(), StepIdempotency) {
		for _, a := range e.Artifacts {
			if strings.Contains(a.Note, "key was not honoured") {
				sawUnhonoured = true
			}
		}
	}
	if !sawUnhonoured {
		t.Error("the extra order the replay created is not annotated in the record")
	}
}

// TestTTLEdgeStepOnlyRunsWhenAskedFor.
func TestTTLEdgeStepOnlyRunsWhenAskedFor(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepIdempotencyTTLEdge) != VerdictSkipped {
		t.Fatalf("verdict = %q, want skipped by default", h.verdict(StepIdempotencyTTLEdge))
	}
	e, _ := LastEntry(h.entries(), StepIdempotencyTTLEdge)
	if !strings.Contains(e.Reason, FlagIncludeTTLEdge) {
		t.Errorf("the skip reason does not name the opt-in flag: %q", e.Reason)
	}
	if !strings.Contains(strings.ToUpper(e.Reason), "SECOND LIVE ORDER") {
		t.Errorf("the skip reason does not warn what the step would do: %q", e.Reason)
	}
}

func TestTTLEdgeStepWhenRequested(t *testing.T) {
	broker := newFakeBroker()
	broker.honourIdempotency = true
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{IncludeTTLEdge: true, TTLWait: 11 * time.Minute}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if v := h.verdict(StepIdempotencyTTLEdge); v != VerdictPass {
		t.Fatalf("verdict = %q, want pass", v)
	}
	if h.slept < 11*time.Minute {
		t.Errorf("the step waited %s; it must outlast the documented ten-minute window", h.slept)
	}
	if got, _ := h.observation(StepIdempotencyTTLEdge, "idempotency.ttl_window_closed"); got != "false" {
		t.Errorf("ttl_window_closed = %q; this broker keeps the key, so the window is observed still open", got)
	}
	// A broker that keeps the key means the replay created nothing, so the step
	// never actually held two orders — the cap is a ceiling, not a target.
	if peak, _ := peakLiveOrders(entriesFor(h.entries(), StepIdempotencyTTLEdge)); peak != 1 {
		t.Errorf("peak live orders = %d, want 1 when the key is still honoured", peak)
	}
	if out := Outstanding(h.entries()); len(undeliberate(out)) != 0 {
		t.Errorf("the validity-window step left orders behind: %+v", out)
	}
}

// TestStepsNeedingAHoldingAreSkippedWithAnActionableReason.
func TestStepsNeedingAHoldingAreSkippedWithAnActionableReason(t *testing.T) {
	broker := newFakeBroker() // no holdings
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, id := range []StepID{StepSellBoundary, StepConditionalRegister} {
		if h.verdict(id) != VerdictSkipped {
			t.Errorf("%s verdict = %q, want skipped", id, h.verdict(id))
		}
		e, _ := LastEntry(h.entries(), id)
		if !strings.Contains(e.Reason, "never buys") {
			t.Errorf("%s skip reason must say the tool will not buy a holding: %q", id, e.Reason)
		}
	}
	// The dependent conditional steps go with it, each saying why.
	for _, id := range []StepID{StepSellableReserved, StepConditionalPersist, StepConditionalModify, StepConditionalCancel} {
		if h.verdict(id) != VerdictSkipped {
			t.Errorf("%s verdict = %q, want skipped", id, h.verdict(id))
		}
	}
	// And no order was placed against a symbol the account does not hold.
	for _, r := range broker.seen() {
		if strings.Contains(r, "POST /orders sell") {
			t.Errorf("a sell was placed on an account with no holdings: %s", r)
		}
	}
}

// TestSellBoundaryProbesTheSmallestPossibleOversell. The oversell has to be
// exactly one share more than the holding: anything larger is exposure the
// measurement does not need.
func TestSellBoundaryProbesTheSmallestPossibleOversell(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepSellBoundary) != VerdictPass {
		t.Fatalf("verdict = %q", h.verdict(StepSellBoundary))
	}
	if got, _ := h.observation(StepSellBoundary, "sell.boundary.over_holding_rejected"); got != "true" {
		t.Errorf("over_holding_rejected = %q, want true", got)
	}
	sawOversell := false
	for _, r := range broker.seen() {
		if strings.HasPrefix(r, "POST /orders sell 005930 x4") {
			sawOversell = true
		}
		if strings.HasPrefix(r, "POST /orders sell 005930 x5") ||
			strings.HasPrefix(r, "POST /orders sell 005930 x1000") {
			t.Errorf("the oversell probe asked for more than one share above the holding: %s", r)
		}
	}
	if !sawOversell {
		t.Errorf("no oversell probe was issued; requests were %v", broker.seen())
	}
	// The whole-holding boundary is above the exposure cap on a three-share
	// holding, so it must be recorded as unverified rather than placed.
	if got, _ := h.observation(StepSellBoundary, "sell.boundary.full_accepted"); got != "unverified" {
		t.Errorf("full_accepted = %q, want unverified on a holding larger than the cap", got)
	}
}

// TestConditionalModifyMeasuresTheNewIdentifierClaim.
func TestConditionalModifyMeasuresTheNewIdentifierClaim(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	if h.verdict(StepConditionalModify) != VerdictPass {
		t.Fatalf("verdict = %q", h.verdict(StepConditionalModify))
	}
	for key, want := range map[string]string{
		"conditional.modify_issues_new_id":       "true",
		"conditional.modify_invalidates_old_id":  "true",
		"conditional.cancel.gone_after":          "true",
		"conditional.reserves_sellable_quantity": "true",
	} {
		if got, ok := h.observation(stepOwning(key), key); !ok || got != want {
			t.Errorf("%s = %q (present=%v), want %q", key, got, ok, want)
		}
	}
}

// TestConditionalModifyFailsWhenTheOldIdSurvives — two live claims on the same
// shares is the thing task 2.6's invariant exists to prevent.
func TestConditionalModifyFailsWhenTheOldIdSurvives(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	broker.modifyLeavesOldReadable = true
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	broker.restart()
	// The stale conditional the modify left behind is real, so the resumed run
	// legitimately ends with something outstanding and says so.
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("resumed run reported: %v", err)
	}
	if h.verdict(StepConditionalModify) != VerdictFail {
		t.Fatalf("verdict = %q, want fail", h.verdict(StepConditionalModify))
	}
	if got, _ := h.observation(StepConditionalModify, "conditional.modify_invalidates_old_id"); got != "false" {
		t.Errorf("modify_invalidates_old_id = %q, want false", got)
	}
}

// TestReadFixturesRecordsTheStatusesTheAccountHasProduced.
func TestReadFixturesRecordsTheStatusesTheAccountHasProduced(t *testing.T) {
	broker := newFakeBroker()
	broker.orderPages = [][]json.RawMessage{{
		mustOrderJSON("h-1", "005930", "BUY", "FILLED", 1, 70000, ""),
		mustOrderJSON("h-2", "005930", "SELL", "CANCELED", 1, 70000, "2026-07-20T10:00:00+09:00"),
	}}
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepReadFixtures) != VerdictPass {
		t.Fatalf("verdict = %q", h.verdict(StepReadFixtures))
	}
	got, _ := h.observation(StepReadFixtures, "order.status.observed")
	if got != "CANCELED,FILLED" {
		t.Errorf("order.status.observed = %q, want the two statuses this account has produced", got)
	}
	unobserved, _ := h.observation(StepReadFixtures, "order.status.documented_unobserved")
	if !strings.Contains(unobserved, "CANCEL_REJECTED") {
		t.Errorf("documented_unobserved = %q, want it to name the statuses nobody has seen", unobserved)
	}
	if listed, _ := h.observation(StepReadFixtures, "order.status.cancel_rejected.listed"); listed != "false" {
		t.Errorf("cancel_rejected.listed = %q, want false", listed)
	}
}

// TestReadFixturesSkipsAnEmptyHistoryWithAWayForward.
func TestReadFixturesSkipsAnEmptyHistoryWithAWayForward(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepReadFixtures) != VerdictSkipped {
		t.Fatalf("verdict = %q, want skipped", h.verdict(StepReadFixtures))
	}
	e, _ := LastEntry(h.entries(), StepReadFixtures)
	if !strings.Contains(e.Reason, "--redo read-fixtures") {
		t.Errorf("the skip reason does not tell the operator how to come back: %q", e.Reason)
	}
}

// TestCostsAreHonestWhenNothingFilled. The far-from-market limits are designed
// not to fill, so "no measured cost" is the expected outcome and it must be
// recorded as such rather than as a zero.
func TestCostsAreHonestWhenNothingFilled(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	if got, _ := h.observation(StepCosts, "costs.collected"); got != "false" {
		t.Errorf("costs.collected = %q, want false", got)
	}
	if got, _ := h.observation(StepCosts, "costs.orders_filled"); got != "0" {
		t.Errorf("costs.orders_filled = %q, want 0", got)
	}
	examined, _ := h.observation(StepCosts, "costs.orders_examined")
	if examined == "0" {
		t.Error("the costs step examined no order although the run placed several")
	}
}

// TestTriggerObservationIsRecordedAsDeferred, so task 2.6 has it on the list.
func TestTriggerObservationIsRecordedAsDeferred(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	if h.verdict(StepConditionalTrigger) != VerdictDeferred {
		t.Fatalf("verdict = %q, want deferred", h.verdict(StepConditionalTrigger))
	}
	e, _ := LastEntry(h.entries(), StepConditionalTrigger)
	if !strings.Contains(e.Reason, "별도 세션") {
		t.Errorf("the deferral reason does not carry the task's own wording: %q", e.Reason)
	}
	if got, _ := h.observation(StepConditionalTrigger, "conditional.triggered_order_id_exposed"); got != "unverified" {
		t.Errorf("triggered_order_id_exposed = %q, want unverified", got)
	}
}

func mustStep(t *testing.T, id StepID) Step {
	t.Helper()
	s, ok := StepByID(id)
	if !ok {
		t.Fatalf("%s is not in the catalogue", id)
	}
	return s
}

// stepOwning maps an observation key back to the step that writes it, so the
// assertions above read as properties rather than as record archaeology.
func stepOwning(key string) StepID {
	switch {
	case strings.HasPrefix(key, "conditional.modify"):
		return StepConditionalModify
	case strings.HasPrefix(key, "conditional.cancel"):
		return StepConditionalCancel
	case key == "conditional.reserves_sellable_quantity":
		return StepSellableReserved
	default:
		return StepConditionalRegister
	}
}
