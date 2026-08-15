package verifylive

// cleanup.go lets a run remove what an earlier run left resting.
//
// # The hole this fills
//
// Three rules that are each right on their own combined, on 2026-07-27, into a
// state the tool could not leave:
//
//	a cancel that fails       leaves an order this tool created resting on the
//	                          account (measurements.md M16)
//	a terminal verdict        is skipped by a resume, so the step that would have
//	                          cancelled it never runs again
//	the exposure cap          counts that order off the *record*, so every later
//	                          mutating step is refused before it sends anything
//
// The account could not be cleared by hand either: cancelling in the broker's own
// app leaves the record still saying the order is live, and the cap reads the
// record. The tool had made a state only the tool could clear and had no way to
// clear it.
//
// # Why it is a plan line and not a special case
//
// The prologue sends live requests, so it goes through the same door everything
// else does: it is enumerated in the plan, a person approves that list, and
// Plan.Authorises refuses anything the list does not carry. A cleanup that
// bypassed the approval "because it is only a cancel" would be exactly the kind of
// unlisted request the batch model exists to make impossible — the fact that a
// cancel is the safe direction is an argument for putting it on the list, not for
// leaving it off.
//
// # Why it is not a step
//
// It measures nothing. The record marks it KindCleanup so every reader that walks
// the procedure — StepCount, RedoSet, the report, the console's step table —
// skips it rather than counting housekeeping as a measured capability.

import (
	"context"
	"errors"
	"fmt"
)

// StepCleanup identifies the prologue on the plan and on the record.
//
// It is deliberately not in Steps(): the catalogue is the list of capabilities
// this tool measures, and removing an order it should never have left behind is
// not one of them.
const StepCleanup StepID = "cleanup"

// CleanupLabel is what the prologue is called on screen. It is kept out of
// stepLabels because that map is asserted to hold the catalogue and nothing else.
const CleanupLabel = "이전 실행이 남긴 객체 정리"

// KindOrder and KindConditional are the artifact kinds the record uses. They were
// string literals in four files before the prologue had to switch on them.
const (
	KindOrder       = "order"
	KindConditional = "conditional-order"
)

// cleanupStep is the Step value the prologue runs as. Mutates is true because it
// sends live requests; Tasks is empty because it feeds no measurement.
func cleanupStep() Step {
	return Step{
		ID:      StepCleanup,
		Title:   "Cancel what an earlier run left resting on the account",
		Mutates: true,
	}
}

// cleanupTargets are the objects this run will offer to remove.
//
// One rule:
//
//	an object is a target once the step it names as its gate has a terminal
//	verdict *recorded after the line that named it*.
//
// An object that names no gate has nothing to wait for and is always a target;
// holdGate supplies the gate an old line meant but could not spell.
//
// The rule is what it is because of two failures a year apart in the same shape.
//
// On 2026-07-28 a conditional order registered on purpose — the persistence step
// reads it back from a later process — was cancelled by the next run's prologue on
// the authority of a `skipped` recorded on 07-26, when the account held nothing.
// It died 308 milliseconds before the measurement that existed to read it, and
// every run after that skipped in the same place; conditional-register was `pass`,
// so no redo could put the subject back (see redo.go). *A verdict recorded before
// an object existed is not a verdict about that object.*
//
// The second is the one this shape was generalised for. A conditional order that
// fires produces a child market sell, and that child has to be **allowed to fill**
// — not cancelling it is the measurement rather than a leak. Until 2026-07-30 the
// guard above protected conditional orders only, and every order this tool created
// was an unconditional target, so the next run would have cancelled the evidence.
// The fix is not a third special case: it is letting the object say what it is
// waiting for.
//
// Release is positional and never a clock (heldAfter). The list comes from the
// record, which is what makes "this tool created it" a fact rather than a guess.
// Nothing else on the account is ever a target.
func (r *Runner) cleanupTargets() []Artifact {
	return withoutM0ManualReconcile(r.prior, cleanupFrom(r.prior, func(id StepID) bool {
		settled, _ := r.settled(id)
		return settled
	}))
}

// PendingCleanup is cleanupTargets for a caller holding only the record.
//
// It exists because "is there anything to do?" is asked in two places and must be
// answered the same way in both. The console refuses a resume that would walk the
// catalogue and settle nothing — the no-op that cost a market window on
// 2026-07-27 — and a record whose steps are all terminal *but* which still holds a
// leftover is not that no-op: the run has one live cancel to send. Answering that
// question from a second copy of the rule is how the deadlock this file exists to
// remove would grow back on the console side.
func PendingCleanup(entries []Entry) []Artifact {
	return withoutM0ManualReconcile(entries, cleanupFrom(entries, func(id StepID) bool { return Settled(entries, id) }))
}

func cleanupFrom(entries []Entry, settled func(StepID) bool) []Artifact {
	var out []Artifact
	for _, l := range outstandingLines(entries) {
		gate := holdGate(l.Artifact)
		if gate == "" {
			out = append(out, l.Artifact)
			continue
		}
		if settled(gate) && heldAfter(entries, gate, l.at) {
			out = append(out, l.Artifact)
		}
	}
	return out
}

// holdGate names the step whose verdict releases an object, or "" for an object
// nothing is waiting on.
//
// The two defaults are the rule this file carried before objects could say what
// they were waiting on, written out so the old records keep the judgement they
// have today:
//
//	conditional order   conditional-cancel. It was the only step that could hold
//	                    one, so `Deliberate` on an old line means exactly
//	                    "held until conditional-cancel".
//	order               nothing. Every order this tool left resting was a mistake
//	                    it had to be able to clear, because no step cancels an
//	                    order from an earlier run.
//
// The second default is the one a trigger measurement has to override: the child
// order a conditional produces when it fires must be *allowed to fill*, and
// "this run did not cancel it" is the measurement rather than a leak.
func holdGate(a Artifact) StepID {
	if a.HeldUntil != "" {
		return a.HeldUntil
	}
	if a.Kind == KindConditional {
		return StepConditionalCancel
	}
	return ""
}

// heldAfter reports whether the gate step's newest line came after the line at
// index at — the line that named the gate.
//
// Reading the gate and measuring the verdict against the same line is the whole
// definition. The alternative considered and rejected (design.md D2) measured
// from the most recent line that declared *any* hold, which let a cancel that
// failed after the hold be undone by any later mention of the object — rebuilding
// the leftover deadlock verify-clears-leftovers exists to remove.
//
// Position in the record, not a clock. The evidence file is append-only JSONL, so
// index order is monotone by construction; timestamps are not, and the record
// shows why an artifact's own CreatedAt cannot stand in for one — a cancellation
// line carries the zero time. A time-based release would also fire in the middle
// of a trigger measurement's wait, where a long wait means the market has not
// moved yet rather than that anything failed.
//
// Fail-closed: no line for the gate at all leaves decided at -1, which is not
// after anything, so the measurement keeps its subject.
func heldAfter(entries []Entry, gate StepID, at int) bool {
	decided := -1
	for i := range entries {
		if entries[i].StepID == gate {
			decided = i
		}
	}
	return decided > at
}

// planCleanup renders the targets as approval lines.
func (r *Runner) planCleanup() []PlannedMutation {
	var lines []PlannedMutation
	for _, a := range r.cleanupTargets() {
		kind := MutateCancelOrder
		whatKO := "미체결 주문"
		if a.Kind == KindConditional {
			kind = MutateCancelConditional
			whatKO = "조건주문"
		}
		lines = append(lines, PlannedMutation{
			Step:   StepCleanup,
			Kind:   kind,
			Symbol: a.Symbol,
			Ends: fmt.Sprintf("this request IS the ending: %s %s stops existing. Nothing is placed to "+
				"replace it", a.Kind, a.ID),
			Note: fmt.Sprintf("%s %s was created by an earlier run of this tool and never cancelled. It is "+
				"filling the live-exposure cap, so the steps below cannot send anything until it is gone",
				a.Kind, a.ID),
			EndsKO: fmt.Sprintf("이 요청 자체가 종료다 — %s %s가 사라진다. 대체로 무언가를 새로 내지 않는다",
				whatKO, a.ID),
			NoteKO: fmt.Sprintf("이 도구의 이전 실행이 만들고 취소하지 못한 %s이다(%s). 노출 상한을 채우고 있어 "+
				"이것이 사라지기 전에는 아래 단계들이 아무것도 보낼 수 없다", whatKO, a.ID),
		})
	}
	return lines
}

// runCleanup cancels the targets, through the same gate every other live request
// goes through.
//
// A failure here is recorded and the run continues. Stopping would be worse than
// the state it is trying to fix: the read-only steps still measure something, and
// the mutating ones are refused by the exposure cap with the reason they have
// always been refused with. What must never happen is the opposite — the leftover
// quietly dropping off the record — and it cannot, because Outstanding is computed
// from the cancellations that actually happened.
func (r *Runner) runCleanup(ctx context.Context, sr *stepRun, targets []Artifact) {
	var first error
	for _, a := range targets {
		var err error
		switch a.Kind {
		case KindOrder:
			err = r.cancelOrder(ctx, sr, a.ID, a.Symbol, "이전 실행이 남긴 미체결 주문 — 노출 상한을 채우고 있다")
		case KindConditional:
			err = r.cancelConditional(ctx, sr, a.ID, a.Symbol, "이전 실행이 남긴 조건주문 — 존속 측정은 이미 끝났다")
		default:
			continue
		}
		if err == nil {
			continue
		}
		if first == nil {
			first = err
		}
		if errors.Is(err, ErrOutsidePlan) || errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			break
		}
	}
	sr.resolve(first)
}

// sweepStep cancels what a step placed and did not clear.
//
// The success paths already do this; the failure paths walked away. On 2026-07-28
// that cost two measurements at once: order-amend was refused, returned on the
// error, and left the order it had just placed resting — filling the one-order
// exposure cap that sell-boundary then needed (measurements.md M24). The step's
// verdict is untouched by the sweep: a swept step is still a failed measurement,
// and saying otherwise would be the record flattering the run.
//
// Scope is exactly the artifacts this step created and did not cancel. Not the
// record's leftovers — those belong to the prologue, which puts each one on a list
// a person approves.
//
// Two conditions skip it, and they are one rule said twice: when the run is not
// allowed to send anything, it does not send this either.
//
//	sr.abort set     the step tried to send something outside the approved plan,
//	                 or stdin was not a terminal, or the context died. The run is
//	                 stopping; the artifact is reported and the next run's prologue
//	                 offers its cancel on a list.
//	ctx cancelled    the operator interrupted.
func (r *Runner) sweepStep(ctx context.Context, sr *stepRun) {
	if sr.abort != nil || ctx.Err() != nil {
		return
	}
	mine := []Entry{{StepID: sr.step.ID, Artifacts: sr.artifacts}}
	for _, a := range Outstanding(mine) {
		if a.Kind != KindOrder {
			// A conditional order outliving its step is the persistence design,
			// not a leak: conditional-persist has to read it from a later process.
			continue
		}
		if a.Deliberate {
			// An order the step declared it is holding. The trigger observation is
			// the case: the child order a conditional produces has to be allowed to
			// fill, and cancelling it here would sweep away the measurement on the
			// path where the step failed to see the fill in time — the one path
			// where the evidence matters most. Nothing before 2026-07-30 marked an
			// order deliberate, so this changes no existing step.
			continue
		}
		if err := r.cancelOrder(ctx, sr, a.ID, a.Symbol, "이 단계가 실패해 남긴 주문 — 다음 단계의 노출 상한을 비운다"); err != nil {
			// Honest and non-fatal: the step already has its verdict, and the
			// artifact stays outstanding for the screen, the record and the next
			// run's prologue to deal with.
			sr.observe("step.sweep.failed", "true", truncateError(err))
			return
		}
	}
}

// cleanup is the prologue as the run sees it: it writes its own line to the record
// and reports whether the run has to stop.
func (r *Runner) cleanup(ctx context.Context) (Outcome, error, bool) {
	targets := r.cleanupTargets()
	if len(targets) == 0 {
		return Outcome{}, nil, false
	}

	sr := &stepRun{step: cleanupStep(), startedAt: r.now()}
	fmt.Fprintf(r.out, "▸ %-22s %s\n", StepCleanup, CleanupLabel)
	r.runCleanup(ctx, sr, targets)

	entry := r.entryFor(sr)
	entry.Kind = KindCleanup
	if err := r.recorder.Append(entry); err != nil {
		return Outcome{}, err, true
	}
	r.written = append(r.written, entry)
	fmt.Fprintf(r.out, "  %s%s\n\n", sr.verdict, reasonSuffix(sr.reason))

	outcome := Outcome{Step: StepCleanup, Title: sr.step.Title, Verdict: sr.verdict, Reason: sr.reason}
	return outcome, sr.abort, sr.abort != nil
}
