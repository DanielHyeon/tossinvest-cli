package verifylive

// redo_test.go pins the one thing RedoSet must never get wrong: a step that
// passed must not come back, because coming back means a live order for a
// measurement that has already been made.

import (
	"testing"
	"time"
)

func recordOf(t *testing.T, verdicts map[StepID][]Verdict) []Entry {
	t.Helper()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	var entries []Entry
	// Catalogue order, so a step with two verdicts gets them in the order given.
	for round := 0; round < 4; round++ {
		for _, step := range Steps() {
			vs := verdicts[step.ID]
			if round >= len(vs) {
				continue
			}
			entries = append(entries, Entry{
				Kind: KindStep, StepID: step.ID, Verdict: vs[round],
				StartedAt: now, FinishedAt: now,
			})
		}
	}
	return entries
}

func TestRedoSetTakesFailedAndSkippedStepsOnly(t *testing.T) {
	entries := recordOf(t, map[StepID][]Verdict{
		StepReadFixtures:        {VerdictPass},
		StepSellableBaseline:    {VerdictFail},
		StepIdempotency:         {VerdictFail},
		StepIdempotencyTTLEdge:  {VerdictSkipped},
		StepOrderCancel:         {VerdictFail},
		StepSellBoundary:        {VerdictSkipped},
		StepConditionalTrigger:  {VerdictDeferred},
		StepConditionalRegister: {VerdictRefused},
		StepConditionalPersist:  {VerdictAwaitingRestart},
	})

	got := RedoSet(entries)
	want := []StepID{
		StepSellableBaseline, StepIdempotency, StepIdempotencyTTLEdge,
		StepOrderCancel, StepSellBoundary,
	}
	if len(got) != len(want) {
		t.Fatalf("RedoSet = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("RedoSet = %v, want %v (catalogue order)", got, want)
		}
	}
}

// TestRedoSetNeverOffersAPassedStep is the safety half, stated on its own so a
// future change to the verdict list cannot quietly widen it.
func TestRedoSetNeverOffersAPassedStep(t *testing.T) {
	all := map[StepID][]Verdict{}
	for _, step := range Steps() {
		all[step.ID] = []Verdict{VerdictPass}
	}
	if got := RedoSet(recordOf(t, all)); len(got) != 0 {
		t.Fatalf("RedoSet on an all-pass record = %v, want nothing", got)
	}
}

// TestRedoSetReadsTheNewestVerdict: a step that failed and was then re-measured
// into a pass is done, and a step that passed and was later re-run into a failure
// is not.
func TestRedoSetReadsTheNewestVerdict(t *testing.T) {
	entries := recordOf(t, map[StepID][]Verdict{
		StepReadFixtures:     {VerdictFail, VerdictPass},
		StepSellableBaseline: {VerdictPass, VerdictFail},
	})
	got := RedoSet(entries)
	if len(got) != 1 || got[0] != StepSellableBaseline {
		t.Fatalf("RedoSet = %v, want only %s", got, StepSellableBaseline)
	}
}

func TestRedoSetOnAnEmptyRecordIsEmpty(t *testing.T) {
	if got := RedoSet(nil); len(got) != 0 {
		t.Fatalf("RedoSet(nil) = %v, want nothing", got)
	}
}

// TestRedoSetIgnoresTheApprovalLine. A declined batch writes one entry with no
// step id and a refused verdict; it must not turn into a redo target.
func TestRedoSetIgnoresTheApprovalLine(t *testing.T) {
	entries := []Entry{
		{Kind: KindApproval, Verdict: VerdictRefused, Title: "batch approval"},
		{Kind: KindApproval, Verdict: VerdictFail, Title: "batch approval"},
	}
	if got := RedoSet(entries); len(got) != 0 {
		t.Fatalf("RedoSet = %v, want nothing — an approval line names no step", got)
	}
}

func TestRedoableVerdictAgreesWithTheSet(t *testing.T) {
	cases := map[Verdict]bool{
		VerdictPass:            false,
		VerdictFail:            true,
		VerdictSkipped:         true,
		VerdictRefused:         false,
		VerdictDeferred:        false,
		VerdictAwaitingRestart: false,
	}
	for verdict, want := range cases {
		if got := RedoableVerdict(verdict); got != want {
			t.Errorf("RedoableVerdict(%s) = %v, want %v", verdict, got, want)
		}
	}
}

// --- reopening a chain whose subject is gone -----------------------------------
//
// A pass normally stays passed. The one exception is a pass whose established
// property is no longer true: conditional-register proves "a conditional order is
// registered and readable", and once that object is gone the property is false and
// every step that depends on it can only skip. See design.md D2.

// goneSubjectRecord is the KR record after 2026-07-28: register passed, the
// prologue cancelled what it registered, and the three dependents skipped.
func goneSubjectRecord(dependents map[StepID]Verdict) []Entry {
	entries := []Entry{{
		Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass,
		Artifacts: []Artifact{{
			Kind: KindConditional, ID: "grLKqiGuCVS7mj", Symbol: "333430", Deliberate: true,
		}},
	}, {
		Kind: KindCleanup, StepID: StepCleanup, Verdict: VerdictPass,
		Artifacts: []Artifact{{
			Kind: KindConditional, ID: "grLKqiGuCVS7mj", Symbol: "333430", Cancelled: true,
		}},
	}}
	for _, step := range Steps() {
		if v, ok := dependents[step.ID]; ok {
			entries = append(entries, Entry{Kind: KindStep, StepID: step.ID, Verdict: v})
		}
	}
	return entries
}

func has(ids []StepID, want StepID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

// TestRedoSetReopensARegisterWhoseConditionalIsGone is the 2026-07-28 deadlock.
// Three runs after it, the console still could not put a conditional order back.
func TestRedoSetReopensARegisterWhoseConditionalIsGone(t *testing.T) {
	entries := goneSubjectRecord(map[StepID]Verdict{
		StepSellableReserved:   VerdictPass,
		StepConditionalPersist: VerdictSkipped,
		StepConditionalTrigger: VerdictDeferred,
		StepConditionalModify:  VerdictSkipped,
		StepConditionalCancel:  VerdictSkipped,
	})

	if !has(RedoSet(entries), StepConditionalRegister) {
		t.Fatalf("the chain cannot be measured again from the console: %v", RedoSet(entries))
	}
}

// TestRedoSetLeavesACompletedChainClosed is the narrowness half, taken from the
// real US record: every dependent passed, only the trigger is deferred.
func TestRedoSetLeavesACompletedChainClosed(t *testing.T) {
	entries := goneSubjectRecord(map[StepID]Verdict{
		StepSellableReserved:   VerdictPass,
		StepConditionalPersist: VerdictPass,
		StepConditionalTrigger: VerdictDeferred,
		StepConditionalModify:  VerdictPass,
		StepConditionalCancel:  VerdictPass,
	})

	if has(RedoSet(entries), StepConditionalRegister) {
		t.Fatalf("a chain that finished is being offered for re-measurement, which would place a live "+
			"conditional order for a property already established: %v", RedoSet(entries))
	}
}

// TestRedoSetDoesNotReopenWhileTheConditionalIsAlive: the ordinary resume is what
// continues a halted chain. Reopening would register a second conditional.
func TestRedoSetDoesNotReopenWhileTheConditionalIsAlive(t *testing.T) {
	entries := []Entry{{
		Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass,
		Artifacts: []Artifact{{
			Kind: KindConditional, ID: "grLKqiGuCVS7mj", Symbol: "333430", Deliberate: true,
		}},
	}, {
		Kind: KindStep, StepID: StepConditionalPersist, Verdict: VerdictAwaitingRestart,
	}}

	if has(RedoSet(entries), StepConditionalRegister) {
		t.Fatalf("the conditional is still live and the resume will read it; reopening would register a "+
			"second one: %v", RedoSet(entries))
	}
}

// TestRedoSetDoesNotReopenForADeferredDependentAlone: conditional-trigger can
// never pass — this tool does not place an order meant to fill. A step that says
// up front it cannot be driven is not a reason to send anything.
func TestRedoSetDoesNotReopenForADeferredDependentAlone(t *testing.T) {
	entries := goneSubjectRecord(map[StepID]Verdict{
		StepSellableReserved:   VerdictPass,
		StepConditionalPersist: VerdictPass,
		StepConditionalTrigger: VerdictDeferred,
		StepConditionalModify:  VerdictPass,
		StepConditionalCancel:  VerdictPass,
	})

	if has(RedoSet(entries), StepConditionalRegister) {
		t.Fatalf("a deferred dependent reopened the chain: %v", RedoSet(entries))
	}
}

// TestRedoSetDoesNotReopenAPassThatLeftNothingBehind: the rule is keyed to the one
// artifact designed to outlive its step. A step that merely passed is untouched.
func TestRedoSetDoesNotReopenAPassThatLeftNothingBehind(t *testing.T) {
	entries := []Entry{
		{Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictPass},
		{Kind: KindStep, StepID: StepSellBoundary, Verdict: VerdictSkipped},
	}
	if has(RedoSet(entries), StepOrderCancel) {
		t.Fatalf("a passed step with no deliberate conditional was reopened: %v", RedoSet(entries))
	}
}
