package verifylive

// plan_test.go covers the approval model the whole tool now rests on: one typed
// string per run, for a list the operator has read.
//
// The properties are the ones that would make a batch approval worse than no
// approval at all:
//
//	the list is complete    every mutation the run will make is on it, derived from
//	                        the catalogue rather than from a second hand-kept list
//	the list is the limit   a request the list does not carry cannot be sent, and a
//	                        symbol substitution is a request the list does not carry
//	the answer is required  a wrong string, an expired one, or no terminal means
//	                        zero mutating requests
//	resume re-asks          a run that continues an earlier one approves what
//	                        remains, because the earlier approval covered the
//	                        earlier requests

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"
)

// --- the list is complete --------------------------------------------------------

// TestThePlanCarriesEveryMutationTheCatalogueDeclares.
//
// Structural, not textual: the expected set is computed from Steps() and the
// runner's own gates, and every declared StepMutation of every step that will run
// has to appear as a plan line for that step. A step that started sending something
// its catalogue entry does not declare would be caught by the guard at runtime; a
// step whose declaration never reached the summary is caught here.
func TestThePlanCarriesEveryMutationTheCatalogueDeclares(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 1)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	plan := r.Plan(context.Background())
	if len(plan.Mutations) == 0 {
		t.Fatal("the plan lists no mutation at all; the run would have nothing to approve")
	}

	// A one-share holding, so every declared quantity rule but the partial resolves.
	for _, step := range Steps() {
		if !step.Mutates {
			continue
		}
		if reason, skip := r.preflightStatic(step, func(StepID) bool { return true }); skip {
			t.Logf("%s is skipped for this account (%s), so it is not expected on the list", step.ID, reason)
			continue
		}
		for _, declared := range step.Mutations {
			if declared.Quantity == QuantityPartial {
				// Deliberately not planned on a one-share holding, and the plan says
				// so in its notes — asserted below.
				continue
			}
			if !planHasLine(plan, step.ID, declared.Kind) {
				t.Errorf("%s declares a %s mutation that never reached the approval list", step.ID, declared.Kind)
			}
		}
	}

	// And nothing appears on the list that no step declared.
	for _, line := range plan.Mutations {
		step, ok := StepByID(line.Step)
		if !ok {
			t.Fatalf("the plan lists a mutation for %s, which is not in the catalogue", line.Step)
		}
		declared := false
		for _, d := range step.Mutations {
			if d.Kind == line.Kind {
				declared = true
			}
		}
		if !declared {
			t.Errorf("the plan lists a %s mutation for %s that the catalogue does not declare", line.Kind, step.ID)
		}
	}
}

// TestEveryPlannedLineSaysWhatItIsAndHowItEnds. A line an operator cannot judge is
// a line they cannot approve.
func TestEveryPlannedLineSaysWhatItIsAndHowItEnds(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	plan := r.Plan(context.Background())
	ordinals := map[int]bool{}
	for i, m := range plan.Mutations {
		if m.Ordinal != i+1 {
			t.Errorf("line %d carries ordinal %d; the summary is numbered from one, in order", i+1, m.Ordinal)
		}
		if ordinals[m.Ordinal] {
			t.Errorf("ordinal %d appears twice", m.Ordinal)
		}
		ordinals[m.Ordinal] = true

		if m.Step == "" || m.Kind == "" {
			t.Errorf("line %d does not say which step sends what: %+v", m.Ordinal, m)
		}
		if strings.TrimSpace(m.Symbol) == "" {
			t.Errorf("line %d names no symbol: %+v", m.Ordinal, m)
		}
		if strings.TrimSpace(m.Ends) == "" {
			t.Errorf("line %d does not say how the exposure ends: %+v", m.Ordinal, m)
		}
		if m.MaxQuantity > 0 && strings.TrimSpace(m.Quantity) == "" {
			t.Errorf("line %d authorises a quantity it does not print: %+v", m.Ordinal, m)
		}
		if m.MaxQuantity > 0 && strings.TrimSpace(m.Pricing) == "" {
			t.Errorf("line %d places or moves an order without saying how it is priced: %+v", m.Ordinal, m)
		}
	}
}

// TestThePromptPrintsEveryPlannedLine.
//
// The count is taken from the rendered text by matching the enumeration itself, so
// a line dropped between the plan and the page cannot hide behind a substring that
// happens to appear elsewhere.
func TestThePromptPrintsEveryPlannedLine(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	plan := r.Plan(context.Background())
	batch := NewBatch(plan, false, time.Now())
	prompt := batch.Prompt()

	numbered := regexp.MustCompile(`(?m)^ +(\d+)\. `)
	found := numbered.FindAllStringSubmatch(prompt, -1)
	if len(found) != len(plan.Mutations) {
		t.Fatalf("the prompt enumerates %d request(s), the plan has %d:\n%s",
			len(found), len(plan.Mutations), prompt)
	}
	// Compared with whitespace collapsed, because the prompt wraps long lines under
	// a hanging indent and a line break inside a headline is not a missing headline.
	flat := squashSpace(prompt)
	for _, m := range plan.Mutations {
		// The prompt is rendered in the display language (task 1.8 ③); the English
		// the KO fields shadow is what the plan digest is computed over and is not
		// what a person reads.
		if !strings.Contains(flat, squashSpace(m.HeadlineKO())) {
			t.Errorf("line %d (%s) is missing from the prompt:\n%s", m.Ordinal, m.HeadlineKO(), prompt)
		}
		if !strings.Contains(flat, squashSpace(m.EndsKO)) {
			t.Errorf("line %d does not say how its exposure ends in the prompt: %s", m.Ordinal, m.EndsKO)
		}
		if !strings.Contains(prompt, string(m.Step)) {
			t.Errorf("the prompt never names the step %s", m.Step)
		}
	}
	for _, want := range []string{batch.Nonce, "만료", "중단", plan.Account} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not contain %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "123-45-678901") {
		t.Error("the prompt printed the unmasked account number")
	}
}

// TestThePlanSaysWhichConditionalOutlivesTheRun. It is the one exposure that is
// still there when the run ends, so the summary has to be unmissable about it.
func TestThePlanSaysWhichConditionalOutlivesTheRun(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 1)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	plan := r.Plan(context.Background())
	var register PlannedMutation
	for _, m := range plan.Mutations {
		if m.Step == StepConditionalRegister && m.Kind == MutateRegisterConditional {
			register = m
		}
	}
	if register.Ordinal == 0 {
		t.Fatal("the conditional registration is not on the approval list")
	}
	ends := strings.ToLower(register.Ends)
	for _, want := range []string{"purpose", "outlive", "conditional-cancel"} {
		if !strings.Contains(ends, want) {
			t.Errorf("the conditional line does not explain that it is left alive on purpose (%q missing): %s",
				want, register.Ends)
		}
	}
	// It is the only line that leaves anything alive, and every line that creates
	// exposure says where that exposure goes.
	for _, m := range plan.Mutations {
		if m.Kind == MutateRegisterConditional {
			continue
		}
		text := strings.ToLower(m.Ends)
		for _, leak := range []string{"outlive", "on purpose", "left registered"} {
			if strings.Contains(text, leak) {
				t.Errorf("line %d (%s) also claims to leave something alive: %s", m.Ordinal, m.Kind, m.Ends)
			}
		}
		switch m.Kind {
		case MutatePlaceOrder, MutateReplayOrder, MutateConflictProbe, MutateAmendOrder, MutateReplayConditional:
			if !strings.Contains(text, "cancel") && !strings.Contains(text, "refus") {
				t.Errorf("line %d (%s) creates exposure without saying how it ends: %s", m.Ordinal, m.Kind, m.Ends)
			}
		}
	}
}

// TestThePlanExcludesWhatItCannotOfferAndSaysWhy.
func TestThePlanExcludesWhatItCannotOfferAndSaysWhy(t *testing.T) {
	broker := newFakeBroker() // no holdings at all
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{})

	plan := r.Plan(context.Background())
	for _, id := range []StepID{StepSellBoundary, StepConditionalRegister, StepConditionalModify, StepConditionalCancel} {
		if plan.Covers(id) {
			t.Errorf("%s is on the approval list although the account holds nothing", id)
		}
		if strings.TrimSpace(plan.ExclusionReason(id)) == "" {
			t.Errorf("%s is excluded with no reason", id)
		}
	}
	if !plan.Covers(StepOrderCancel) {
		t.Error("the buy-side steps must still be on the list on an account with no holdings")
	}
	if plan.Covers(StepIdempotencyTTLEdge) {
		t.Error("the opt-in validity-window step is on the list without --include-ttl-edge")
	}
}

// TestTheWholeHoldingSellIsLeftOffAboveTheCap — and the plan says so rather than
// silently shrinking the order.
func TestTheWholeHoldingSellIsLeftOffAboveTheCap(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	plan := r.Plan(context.Background())
	for _, m := range plan.Mutations {
		if m.Step == StepSellBoundary && m.Kind == MutatePlaceOrder && m.MaxQuantity == 3 {
			t.Errorf("a three-share whole-holding sell is on the list although the cap is %s",
				trim(DefaultMaxSellQuantity))
		}
	}
	notes := strings.Join(plan.Notes, "\n")
	if !strings.Contains(notes, "max-sell-quantity") {
		t.Errorf("the plan does not explain why the whole-holding boundary is missing:\n%s", notes)
	}
	// The oversell is still planned, at exactly one share more than the holding.
	oversell := 0.0
	for _, m := range plan.Mutations {
		if m.Step == StepSellBoundary && m.MaxQuantity > oversell {
			oversell = m.MaxQuantity
		}
	}
	if oversell != 4 {
		t.Errorf("the largest sell on the list is %s share(s), want the four-share oversell probe", trim(oversell))
	}
}

// --- the list is the limit --------------------------------------------------------

// TestApprovingTheBatchLetsEveryStepProceed, and asks exactly once.
func TestApprovingTheBatchLetsEveryStepProceed(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := alwaysConfirm()
	h := newHarness(t, broker, op)

	first, second := runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	if len(op.batches) != 2 {
		t.Fatalf("the operator was asked for %d approvals across two invocations, want one each", len(op.batches))
	}
	if len(op.seen) != 0 {
		t.Errorf("the batch model still showed %d per-mutation prompt(s): %v", len(op.seen), op.actions())
	}
	for _, id := range []StepID{StepIdempotency, StepOrderCancel, StepOrderAmend, StepSellBoundary, StepConditionalRegister} {
		if h.verdict(id) != VerdictPass {
			t.Errorf("%s verdict = %q under an approved batch, want pass", id, h.verdict(id))
		}
	}
	if h.verdict(StepConditionalCancel) != VerdictPass {
		t.Errorf("conditional-cancel verdict = %q after the resumed run", h.verdict(StepConditionalCancel))
	}
	_ = first
	if len(second.Outstanding) != 0 {
		t.Errorf("an approved run left something live: %+v", second.Outstanding)
	}
}

// TestEveryMutationTheRunMadeWasOnTheApprovedList.
//
// The guard is the enforcement, so this asserts the guard was actually exercised
// and never had to refuse: a run in which authorise fired would have halted, and a
// run in which it was never consulted would prove nothing.
func TestEveryMutationTheRunMadeWasOnTheApprovedList(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := alwaysConfirm()
	h := newHarness(t, broker, op)
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	approved := map[StepID]bool{}
	for _, b := range op.batches {
		for _, m := range b.Plan.Mutations {
			approved[m.Step] = true
		}
	}
	for _, e := range h.entries() {
		if len(e.Artifacts) == 0 {
			continue
		}
		if !approved[e.StepID] {
			t.Errorf("%s created or cancelled something although no batch approved it: %+v",
				e.StepID, e.Artifacts)
		}
	}
	mutated := 0
	for _, r := range broker.seen() {
		if strings.HasPrefix(r, "POST ") || strings.HasPrefix(r, "DELETE ") {
			mutated++
		}
	}
	if mutated == 0 {
		t.Fatal("the run sent no mutating request, so the approval proved nothing")
	}
}

// TestASymbolSubstitutionIsNotAuthorised is requirement two at its sharpest: the
// approved list names the symbol, so a step that would send another one is refused
// rather than adapted to.
func TestASymbolSubstitutionIsNotAuthorised(t *testing.T) {
	plan := Plan{Mutations: []PlannedMutation{{
		Ordinal: 1, Step: StepOrderCancel, Kind: MutatePlaceOrder,
		Symbol: "005930", Side: "buy", MaxQuantity: 1,
	}}}

	if !plan.Authorises(StepOrderCancel, MutatePlaceOrder, "005930", "buy", 1) {
		t.Fatal("the approved request itself is not authorised")
	}
	for _, tc := range []struct {
		name                   string
		step                   StepID
		kind                   MutationKind
		symbol, side           string
		quantity               float64
		wantAuthorisedRegardle bool
	}{
		{name: "another symbol", step: StepOrderCancel, kind: MutatePlaceOrder, symbol: "000660", side: "buy", quantity: 1},
		{name: "the other side", step: StepOrderCancel, kind: MutatePlaceOrder, symbol: "005930", side: "sell", quantity: 1},
		{name: "another step", step: StepOrderAmend, kind: MutatePlaceOrder, symbol: "005930", side: "buy", quantity: 1},
		{name: "another class", step: StepOrderCancel, kind: MutateAmendOrder, symbol: "005930", side: "buy", quantity: 1},
		{name: "more shares", step: StepOrderCancel, kind: MutatePlaceOrder, symbol: "005930", side: "buy", quantity: 2},
	} {
		if plan.Authorises(tc.step, tc.kind, tc.symbol, tc.side, tc.quantity) {
			t.Errorf("%s: authorised by a plan that does not list it", tc.name)
		}
	}
}

// TestAMutationOutsideTheApprovedListSendsNothing, at the guard itself.
func TestAMutationOutsideTheApprovedListSendsNothing(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{})

	plan := r.Plan(context.Background())
	r.plan = &plan

	sr := &stepRun{step: mustStep(t, StepOrderCancel)}
	_, err := r.placeOrder(context.Background(), sr, orderSpec{
		Symbol: "000660", Market: MarketKR, Side: "buy", Quantity: 1, Price: 56000,
	}, "cancelled")
	if !errors.Is(err, ErrOutsidePlan) {
		t.Fatalf("err = %v, want ErrOutsidePlan for a symbol the approval never named", err)
	}
	for _, req := range broker.seen() {
		if strings.HasPrefix(req, "POST ") {
			t.Errorf("a request reached the broker anyway: %s", req)
		}
	}
	// And the same call for the approved symbol goes through, so the guard is not
	// simply refusing everything.
	if _, err := r.placeOrder(context.Background(), sr, orderSpec{
		Symbol: "005930", Market: MarketKR, Side: "buy", Quantity: 1, Price: 56000,
	}, "cancelled"); err != nil {
		t.Fatalf("the approved request was refused: %v", err)
	}
}

// TestARunThatWouldLeaveTheApprovedListStops.
//
// The oversell probe is dropped from the approved plan; the step still runs its
// earlier boundaries, and the moment it reaches the request nobody approved the
// whole run halts with nothing left behind.
func TestARunThatWouldLeaveTheApprovedListStops(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	plan := r.Plan(context.Background())
	var kept []PlannedMutation
	for _, m := range plan.Mutations {
		if m.Step == StepSellBoundary && m.MaxQuantity > DefaultMaxSellQuantity {
			continue // the oversell line
		}
		kept = append(kept, m)
	}
	if len(kept) == len(plan.Mutations) {
		t.Fatal("the oversell line was not on the plan, so this test proves nothing")
	}
	plan.Mutations = kept
	r.plan = &plan

	summary, err := r.Run(context.Background())
	if !errors.Is(err, ErrOutsidePlan) {
		t.Fatalf("err = %v, want the run to stop with ErrOutsidePlan", err)
	}
	if !summary.Halted {
		t.Error("the run did not report itself halted")
	}
	if h.verdict(StepSellBoundary) != VerdictFail {
		t.Errorf("sell-boundary verdict = %q, want fail", h.verdict(StepSellBoundary))
	}
	if v := h.verdict(StepConditionalRegister); v != "" {
		t.Errorf("conditional-register ran after the halt with verdict %q", v)
	}
	if out := Outstanding(h.entries()); len(out) != 0 {
		t.Errorf("the halted run left something live: %+v", out)
	}
	// The oversell never reached the broker.
	for _, req := range broker.seen() {
		if strings.HasPrefix(req, "POST /orders sell 005930 x4") {
			t.Errorf("the unapproved oversell was sent: %s", req)
		}
	}
}

// TestAStepWithNoApprovedLinesIsSkippedNotAborted. The two directions are
// deliberately different: nothing on the list means the step never starts.
func TestAStepWithNoApprovedLinesIsSkippedNotAborted(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	plan := r.Plan(context.Background())
	var kept []PlannedMutation
	for _, m := range plan.Mutations {
		if m.Step == StepSellBoundary {
			continue
		}
		kept = append(kept, m)
	}
	plan.Mutations = kept
	r.plan = &plan

	if _, err := r.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if h.verdict(StepSellBoundary) != VerdictSkipped {
		t.Fatalf("sell-boundary verdict = %q, want skipped", h.verdict(StepSellBoundary))
	}
	e, _ := LastEntry(h.entries(), StepSellBoundary)
	if !strings.Contains(e.Reason, "승인된 배치에 없다") {
		t.Errorf("the skip reason does not say the step was not approved: %q", e.Reason)
	}
	if h.verdict(StepConditionalRegister) != VerdictPass {
		t.Error("dropping one step from the approval stopped the steps that were approved")
	}
}

// --- the answer is required --------------------------------------------------------

// TestAWrongNonceMutatesNothing.
func TestAWrongNonceMutatesNothing(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := refuseBatch(ErrRefused)
	h := newHarness(t, broker, op)

	summary, err := h.run(Options{HoldingSymbol: "005930"})
	if err != nil {
		t.Fatalf("declining the batch is a decision, not an error: %v", err)
	}
	if !summary.Halted {
		t.Error("the run did not report itself halted")
	}
	assertNothingWasSent(t, broker)
	if len(summary.Outcomes) != 0 {
		t.Errorf("steps ran after the batch was declined: %+v", summary.Outcomes)
	}
	// The decision itself is on the record, and it is not mistaken for a step.
	approval := approvalEntry(t, h.entries())
	if approval.Verdict != VerdictRefused {
		t.Errorf("the approval line says %q, want refused", approval.Verdict)
	}
	if BuildProgress("/nowhere", h.entries()).Steps != nil {
		t.Error("the declined approval was reported as a step")
	}
}

// TestDecliningTheBatchDoesNotLockTheOperatorOut.
//
// The refusal is on the record, but it is not a step, so the next attempt is a
// first attempt and not a resumption of nothing.
func TestDecliningTheBatchDoesNotLockTheOperatorOut(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := refuseBatch(ErrRefused)
	h := newHarness(t, broker, op)

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("first attempt: %v", err)
	}
	if n := StepCount(h.entries()); n != 0 {
		t.Fatalf("a declined run recorded %d step(s); it ran none", n)
	}

	op.answerBatch = func(Batch) error { return nil }
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("second attempt: %v", err)
	}
	if op.lastBatch(t).Resumed {
		t.Error("the second attempt introduced itself as a resumption although nothing had run")
	}
	if h.verdict(StepOrderCancel) != VerdictPass {
		t.Errorf("order-cancel verdict = %q on the approved attempt", h.verdict(StepOrderCancel))
	}
}

// TestAnExpiredBatchMutatesNothing. The expiry is real: the same nonce, typed late.
func TestAnExpiredBatchMutatesNothing(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := alwaysConfirm()
	op.answerBatch = func(b Batch) error {
		return b.Verify(b.Nonce, b.ExpiresAt.Add(time.Second))
	}
	h := newHarness(t, broker, op)

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("an expired approval is a refusal, not an error: %v", err)
	}
	assertNothingWasSent(t, broker)
	if v := approvalEntry(t, h.entries()).Verdict; v != VerdictRefused {
		t.Errorf("the approval line says %q, want refused", v)
	}
}

// TestTheBatchApprovalIsRecordedWithWhatWasApproved.
func TestTheBatchApprovalIsRecordedWithWhatWasApproved(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := alwaysConfirm()
	h := newHarness(t, broker, op)
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	entry := approvalEntry(t, h.entries())
	if entry.Verdict != VerdictPass {
		t.Errorf("the approval line says %q, want pass", entry.Verdict)
	}
	if entry.StepID != "" {
		t.Errorf("the approval line claims to be step %q", entry.StepID)
	}
	if entry.AccountRef == "123-45-678901" {
		t.Error("the approval line carries the unmasked account number")
	}
	values := map[string]string{}
	for _, o := range entry.Observations {
		values[o.Key] = o.Value
	}
	if values["approval.model"] != "batch" {
		t.Errorf("approval.model = %q, want batch", values["approval.model"])
	}
	if values["approval.plan_digest"] == "" {
		t.Error("the record does not fingerprint the list that was approved")
	}
	want := len(op.lastBatch(t).Plan.Mutations)
	if values["approval.requests_listed"] != trim(float64(want)) {
		t.Errorf("approval.requests_listed = %q, want %d", values["approval.requests_listed"], want)
	}
}

// --- the digest is stable across builds -----------------------------------------------

// planDigest20260727 is the digest of the plan a 3-share KR holding produces, taken
// from the build that was in the operator's hands on 2026-07-27.
//
// It is pinned rather than recomputed because approval.plan_digest is what ties an
// evidence record to the list a person actually read: a resumed run whose digest no
// longer matches the record's cannot claim the same batch was approved across the
// restart (measurements.md M3), and a live verification that is halfway through
// would have to start over. Changing this constant is therefore a deliberate act
// with a live cost, not a test fixup — if an edit moves it, the edit changed what
// the operator is agreeing to.
const planDigest20260727 = "sha256:cef553cba1548ab3e918147119c0c587"

// TestThePlanDigestIsPinnedAcrossBuilds.
func TestThePlanDigestIsPinnedAcrossBuilds(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	r := h.runner(t, Options{HoldingSymbol: "005930"})

	if got := r.Plan(context.Background()).Digest(); got != planDigest20260727 {
		t.Errorf("plan digest = %q, want %q.\nThe approved-list fingerprint moved. Every record on disk "+
			"carries the old value, so a --resume can no longer prove the same batch was approved.", got,
			planDigest20260727)
	}
}

// TestThePlanDigestDoesNotDependOnWhatTheReadStepsSee.
//
// The digest is Digest(plan.Mutations), and the plan is derived from the step
// catalogue, the record and the operator's flags. The one account read it makes is
// the sellable quantity. Nothing a read-only step does — least of all how
// read-fixtures walks the order list — may reach it, because the operator resuming
// a half-finished verification is entitled to the same fingerprint from a build
// whose read steps were fixed underneath them.
func TestThePlanDigestDoesNotDependOnWhatTheReadStepsSee(t *testing.T) {
	empty := newFakeBroker().withHolding("005930", 3)

	busy := newFakeBroker().withHolding("005930", 3)
	busy.orderPages = [][]json.RawMessage{
		{mustOrderJSON("h-1", "005930", "BUY", "FILLED", 1, 70000, "")},
		{mustOrderJSON("h-2", "005930", "SELL", "CANCELED", 1, 70000, "2026-07-20T10:00:00+09:00")},
		{mustOrderJSON("h-3", "005930", "BUY", "PARTIAL_FILLED", 2, 70000, "")},
	}
	busy.withWorkingOrder("w-1", "005930", "PENDING")
	busy.withWorkingOrder("w-2", "005930", "PARTIAL_FILLED")

	digests := map[string]string{}
	for name, broker := range map[string]*fakeBroker{"empty history": empty, "busy account": busy} {
		h := newHarness(t, broker, alwaysConfirm())
		r := h.runner(t, Options{HoldingSymbol: "005930"})
		digests[name] = r.Plan(context.Background()).Digest()
	}
	if digests["empty history"] != digests["busy account"] {
		t.Errorf("the plan digest moved with the account's order history: %q vs %q — the approval "+
			"fingerprint must depend on the list, not on what the reads happen to find",
			digests["empty history"], digests["busy account"])
	}
}

// --- resume re-asks -----------------------------------------------------------------

// TestAResumedRunApprovesItsRemainingBatch.
//
// The second invocation must ask again, must not re-list what the first one already
// settled, and must list what is left — including the cancel of the conditional the
// first run deliberately left alive.
func TestAResumedRunApprovesItsRemainingBatch(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := alwaysConfirm()
	h := newHarness(t, broker, op)

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	// What the first run actually settled — the halt at the persistence check
	// leaves the later conditional steps unsettled, and those must be re-listed.
	settled := map[StepID]bool{}
	for _, e := range h.entries() {
		if isStepEntry(e) && e.Verdict.Terminal() {
			settled[e.StepID] = true
		}
	}
	broker.restart()
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("resumed run: %v", err)
	}

	if len(op.batches) != 2 {
		t.Fatalf("the run was approved %d time(s) across two invocations, want once each", len(op.batches))
	}
	first, second := op.batches[0], op.batches[1]
	if first.Resumed {
		t.Error("the first batch claims to be a resumption")
	}
	if !second.Resumed {
		t.Error("the resumed batch does not tell the operator it is the remaining part")
	}
	if !strings.Contains(second.Prompt(), "남은 부분") {
		t.Errorf("the resumed prompt does not say it covers what is left:\n%s", second.Prompt())
	}
	if !first.Plan.Covers(StepOrderCancel) || !first.Plan.Covers(StepConditionalRegister) {
		t.Fatal("the first batch did not list the steps the first run went on to settle")
	}

	for _, m := range second.Plan.Mutations {
		if settled[m.Step] {
			t.Errorf("the resumed batch re-lists %s, which the first run already settled — approving it "+
				"again would place a second live order for a measurement already made", m.Step)
		}
	}
	if !second.Plan.Covers(StepConditionalCancel) {
		t.Error("the resumed batch does not list the cancel of the conditional left alive")
	}
	if second.Nonce == first.Nonce {
		t.Error("the resumed run re-used the first run's approval string")
	}
}

// TestAResumedRunThatIsDeclinedCancelsNothing — the conditional stays registered
// and the tool says so, rather than removing it without an approval.
func TestAResumedRunThatIsDeclinedCancelsNothing(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := alwaysConfirm()
	h := newHarness(t, broker, op)

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("first run: %v", err)
	}
	broker.restart()

	op.answerBatch = func(Batch) error { return ErrRefused }
	summary, err := h.run(Options{HoldingSymbol: "005930"})
	if err != nil {
		t.Fatalf("declining the resumed batch is a decision, not an error: %v", err)
	}
	if len(summary.Outstanding) != 1 || summary.Outstanding[0].Kind != "conditional-order" {
		t.Fatalf("Outstanding = %+v, want the conditional the first run left alive", summary.Outstanding)
	}
	if len(broker.conds) != 1 {
		t.Errorf("the broker has %d conditional order(s); a declined approval must not cancel one", len(broker.conds))
	}
}

// --- --confirm-each still works ------------------------------------------------------

// TestConfirmEachAsksPerMutationAndNeverForABatch.
func TestConfirmEachAsksPerMutationAndNeverForABatch(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	op := alwaysConfirm()
	h := newHarness(t, broker, op)

	first, err := h.run(Options{HoldingSymbol: "005930", ConfirmEach: true})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(op.batches) != 0 {
		t.Errorf("--confirm-each asked for a batch approval %d time(s)", len(op.batches))
	}
	if len(op.seen) < 5 {
		t.Fatalf("--confirm-each showed only %d prompt(s): %v", len(op.seen), op.actions())
	}
	for _, m := range op.seen {
		if m.Nonce == "" || m.Step == "" || strings.TrimSpace(m.Action) == "" {
			t.Errorf("a per-mutation prompt is incomplete: %+v", m)
		}
	}
	if !first.Halted {
		t.Error("--confirm-each did not reach the persistence halt")
	}
	if h.verdict(StepOrderCancel) != VerdictPass {
		t.Errorf("order-cancel verdict = %q under --confirm-each", h.verdict(StepOrderCancel))
	}
}

// --- helpers ---------------------------------------------------------------------------

// squashSpace collapses every run of whitespace to one space, so an assertion
// about what the prompt says survives the prompt being wrapped.
func squashSpace(s string) string { return strings.Join(strings.Fields(s), " ") }

func planHasLine(p Plan, step StepID, kind MutationKind) bool {
	for _, m := range p.Mutations {
		if m.Step == step && m.Kind == kind {
			return true
		}
	}
	return false
}

func approvalEntry(t *testing.T, entries []Entry) Entry {
	t.Helper()
	for _, e := range entries {
		if e.Kind == KindApproval {
			return e
		}
	}
	t.Fatal("the record carries no batch-approval line")
	return Entry{}
}

func assertNothingWasSent(t *testing.T, broker *fakeBroker) {
	t.Helper()
	for _, req := range broker.seen() {
		if strings.HasPrefix(req, "POST ") || strings.HasPrefix(req, "DELETE ") {
			t.Errorf("a mutating request was issued without an approved batch: %s", req)
		}
	}
}
