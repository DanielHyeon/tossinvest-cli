package verifylive

// abort.go is how an operator ends a measurement that is still holding something.
//
// # The hole this fills
//
// Objects this tool creates are released by a *verdict*, not by a clock. A
// conditional order waits for conditional-cancel; a child order waits for the
// trigger observation to decide. That rule was chosen on purpose and the
// alternative was rejected on purpose: a wait that is long means the market has
// not come to the price yet, not that anything failed, and a lease that expired
// mid-measurement would cancel the subject of the thing being measured
// (verify-holds-what-it-awaits design.md D3).
//
// The cost of that rule is that a measurement which never reaches a terminal
// verdict — the operator walks away, the market closes, the process is killed
// between the trigger and the fill — leaves its objects held forever. The cleanup
// prologue will not offer them, because it is doing exactly what it was told.
//
// So there has to be a way to say "I am ending this", and it has to be a person
// saying it. That is this file.
//
// # Why it is not a shortcut
//
// It sends live cancels, so it goes through the same door everything else does:
// the targets are listed, the list becomes a plan, the plan is approved once by
// the same expiring confirmation `verify run` uses, and nothing off the list is
// sent. There is no new prompt, no typed phrase of its own and no extra approval
// step — the batch rail already exists and adding a second one would be friction
// pretending to be safety.
//
// Targets come from the record and nowhere else. What this tool created is a fact
// the evidence file holds; what else is on the account is none of its business.

import (
	"context"
	"fmt"
)

// StepAbort identifies the operator's ending on the plan and on the record.
//
// Like StepCleanup it is deliberately not in Steps(): it measures nothing. It is a
// separate identifier from StepCleanup because the two are different acts — the
// prologue removes what an earlier run *leaked*, and this removes what a
// measurement is still legitimately *holding*, which is why only this one may
// touch a held object.
const StepAbort StepID = "abort"

// AbortLabel is what it is called on screen.
const AbortLabel = "붙잡힌 측정 사슬을 운영자가 끝낸다"

func abortStep() Step {
	return Step{
		ID:      StepAbort,
		Title:   "End a measurement chain the operator is not going to finish",
		Mutates: true,
	}
}

// AbortTargets is everything the record says this tool still has live.
//
// Outstanding, not cleanupTargets: the whole point is to reach the objects the
// hold rule protects. Nothing else widens — a filled or cancelled object is not
// outstanding and is not here, and neither is anything the record never mentioned.
func AbortTargets(entries []Entry) []Artifact { return Outstanding(entries) }

// AbortResult is what an abort did.
type AbortResult struct {
	RunID string `json:"run_id"`
	// Targets are what the operator was asked to approve.
	Targets []Artifact `json:"targets,omitempty"`
	// Remaining are the ones still live afterwards — a cancel can fail, and the
	// honest answer is the one that says so.
	Remaining []Artifact `json:"remaining,omitempty"`
	// Approved reports that a person authorised the list.
	Approved bool `json:"approved"`
	// Reason is why nothing happened, when nothing did.
	Reason string `json:"reason,omitempty"`
}

// Abort cancels what this tool is holding, under one approval.
//
// It never consults a clock to decide what is eligible. Elapsed time is not an
// input: the list is the record's outstanding set at the moment the operator
// asked, and an object that has been held for a month is treated exactly like one
// held for a minute.
func (r *Runner) Abort(ctx context.Context, why string) (AbortResult, error) {
	result := AbortResult{RunID: r.runID}
	targets := AbortTargets(r.prior)
	if len(targets) == 0 {
		result.Reason = "이 도구의 기록에 살아 있는 객체가 없다 — 끝낼 사슬이 없다"
		fmt.Fprintln(r.out, result.Reason)
		return result, nil
	}
	result.Targets = targets

	plan := Plan{RunID: r.runID, Account: maskedAccount(r.accountRef)}
	for i, a := range targets {
		line := abortLine(a, why)
		line.Ordinal = i + 1
		plan.Mutations = append(plan.Mutations, line)
	}
	r.plan = &plan

	startedAt := r.now()
	fmt.Fprintf(r.out, "이 요청들이 나간다 — 그 외에는 아무것도 나가지 않는다.\n\n")
	plan.WriteLines(r.out)

	batch := NewBatch(plan, StepCount(r.prior) > 0, r.now())
	err := r.confirmBatch(batch)
	verdict, reason := VerdictPass, ""
	if err != nil {
		verdict, reason = VerdictRefused, err.Error()
		r.plan = nil
	}
	if recErr := r.recordApproval(plan, batch, verdict, reason, startedAt); recErr != nil {
		return result, recErr
	}
	if err != nil {
		result.Reason = "승인되지 않았다. 아무것도 전송되지 않았다"
		fmt.Fprintf(r.out, "\n%s\n", result.Reason)
		result.Remaining = targets
		return result, nil
	}
	result.Approved = true

	sr := &stepRun{step: abortStep(), startedAt: startedAt}
	fmt.Fprintf(r.out, "\n▸ %-22s %s\n", StepAbort, AbortLabel)
	r.runCleanup(ctx, sr, targets)
	r.closeChains(sr, targets, why)

	entry := r.entryFor(sr)
	entry.Kind = KindCleanup
	entry.Reason = abortReason(why, sr.reason)
	if appendErr := r.recorder.Append(entry); appendErr != nil {
		return result, appendErr
	}
	r.written = append(r.written, entry)
	fmt.Fprintf(r.out, "  %s%s\n", sr.verdict, reasonSuffix(entry.Reason))

	result.Remaining = Outstanding(r.allEntries())
	if len(result.Remaining) > 0 {
		return result, fmt.Errorf("verify: %d건이 아직 계좌에 살아 있다: %s — 브로커 앱에서 직접 취소한 뒤 "+
			"`tossctl verify status`로 다시 확인하라",
			len(result.Remaining), describeArtifacts(result.Remaining))
	}
	return result, nil
}

// closeChains records that the measurements these objects belonged to are over.
//
// A chain identifier is how the record says "these two were one protection"
// (record.go ChainID). Cancelling the objects without saying why the chain ended
// would leave the evidence file describing a measurement that simply stops, which
// is the shape of M37 — a run whose objects disappeared with no line explaining it.
func (r *Runner) closeChains(sr *stepRun, targets []Artifact, why string) {
	seen := map[string]bool{}
	for _, a := range targets {
		chain := a.ChainID
		if chain == "" {
			chain = ChainOf(r.prior, a.Kind, a.ID)
		}
		if chain == "" || seen[chain] {
			continue
		}
		seen[chain] = true
		sr.observe("chain.closed."+chain, "operator-abort", why)
	}
	sr.observe("abort.targets", fmt.Sprint(len(targets)), describeArtifacts(targets))
}

func abortReason(why, failure string) string {
	if failure != "" {
		return failure
	}
	if why == "" {
		return "운영자가 사슬을 끝냈다"
	}
	return "운영자가 사슬을 끝냈다 — " + why
}

// abortLine renders one target as the line a person approves.
func abortLine(a Artifact, why string) PlannedMutation {
	kind, whatKO := MutateCancelOrder, "미체결 주문"
	if a.Kind == KindConditional {
		kind, whatKO = MutateCancelConditional, "조건주문"
	}
	held := ""
	heldKO := ""
	if a.HeldUntil != "" {
		held = fmt.Sprintf(" It is being held for %s, which has not decided; ending the chain is what "+
			"releases it.", a.HeldUntil)
		heldKO = fmt.Sprintf(" %s의 판정을 기다리며 붙잡혀 있다 — 사슬을 끝내는 것이 이것을 놓아주는 "+
			"유일한 방법이다.", a.HeldUntil)
	}
	return PlannedMutation{
		Step:   StepAbort,
		Kind:   kind,
		Symbol: a.Symbol,
		Ends: fmt.Sprintf("this request IS the ending: %s %s stops existing. Nothing is placed to replace it",
			a.Kind, a.ID),
		Note: fmt.Sprintf("%s %s was created by this tool and is still live.%s %s",
			a.Kind, a.ID, held, why),
		EndsKO: fmt.Sprintf("이 요청 자체가 종료다 — %s %s가 사라진다. 대체로 무언가를 새로 내지 않는다",
			whatKO, a.ID),
		NoteKO: fmt.Sprintf("이 도구가 만들었고 아직 살아 있는 %s이다(%s).%s %s",
			whatKO, a.ID, heldKO, why),
	}
}
