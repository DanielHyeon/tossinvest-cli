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
