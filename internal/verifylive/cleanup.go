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
// Two rules, and the second one is the whole subtlety:
//
//	an order              is always a target. No step in the catalogue ever
//	                      cancels an order from an earlier run — each step cancels
//	                      what it placed itself — so if the prologue does not take
//	                      it, nothing will.
//	a conditional order   is a target only once the step that cancels it has a
//	                      terminal verdict. Before that it is the subject of a
//	                      measurement: conditional-persist observes that it
//	                      outlives the process that registered it, and a prologue
//	                      that cancelled it first would delete the evidence and
//	                      then report the step as unmeasurable.
//
// The list comes from the record, which is what makes "this tool created it" a
// fact rather than a guess. Nothing else on the account is ever a target.
func (r *Runner) cleanupTargets() []Artifact {
	var out []Artifact
	for _, a := range Outstanding(r.prior) {
		switch a.Kind {
		case KindOrder:
			out = append(out, a)
		case KindConditional:
			if settled, _ := r.settled(StepConditionalCancel); settled {
				out = append(out, a)
			}
		}
	}
	return out
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
