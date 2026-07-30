package verifylive

// record_filled_test.go covers the second way an object this tool created can stop
// existing.
//
// Until the trigger measurement there was only one: this tool cancelled it. That
// made "not cancelled" and "live" the same sentence, and every consumer of
// Outstanding was written against it. A conditional order that fires produces a
// child order that is *meant* to fill, and a filled order is not cancelled — so
// without a second terminal state a successful measurement leaves an order the
// record insists is live forever: filling the exposure cap, printed by
// `verify status` as a leftover, and offered for cancellation by the next run's
// prologue (proposal.md G2).
//
// The cheaper fix — write the fill as a cancellation — is the one these tests
// exist to rule out. The record is the document every measurement in
// measurements.md is derived from, and a fill recorded as a cancel makes the
// trigger measurement's own conclusion read "we cancelled it".

import (
	"testing"
	"time"
)

func filledEntry(step StepID, a Artifact) Entry {
	return Entry{Kind: KindStep, StepID: step, Verdict: VerdictPass, Artifacts: []Artifact{a}}
}

// TestAFilledObjectIsNotOutstanding is the whole of G2, asserted through each of
// the four consumers that were wrong about it.
func TestAFilledObjectIsNotOutstanding(t *testing.T) {
	at := time.Date(2026, 7, 30, 4, 30, 0, 0, time.UTC)
	entries := []Entry{
		filledEntry(StepConditionalTrigger, Artifact{
			Kind: KindOrder, ID: "child-1", Symbol: "TSLA", CreatedAt: at,
			HeldUntil: StepConditionalTrigger, ChainID: "chain-1",
		}),
		filledEntry(StepConditionalTrigger, Artifact{
			Kind: KindOrder, ID: "child-1", Symbol: "TSLA", CreatedAt: at,
			Filled: true, FilledAt: at.Add(2 * time.Second),
			HeldUntil: StepConditionalTrigger, ChainID: "chain-1",
		}),
	}

	if got := Outstanding(entries); len(got) != 0 {
		t.Fatalf("Outstanding = %+v, want empty — a filled order does not exist any more", got)
	}
	if got := PendingCleanup(entries); len(got) != 0 {
		t.Errorf("PendingCleanup = %+v; the next run would put a cancellation of a filled order on the "+
			"list a person approves", got)
	}
	r := &Runner{prior: entries}
	if n := r.liveCount(KindOrder, nil); n != 0 {
		t.Errorf("liveCount = %d, want 0 — a filled order must not go on occupying the exposure cap", n)
	}

	// And the record says which of the two endings it was.
	last := entries[len(entries)-1].Artifacts[0]
	if last.Cancelled {
		t.Error("the fill was recorded as a cancellation; the measurement's own conclusion would then read " +
			"'we cancelled it'")
	}
	if !last.Filled || last.FilledAt.IsZero() {
		t.Errorf("artifact = %+v, want Filled with the time it was observed at", last)
	}
}

// TestFillIsMonotone. outstandingLines already refuses to let a later line
// resurrect a cancelled object (measurements.md M22 is what that looks like when
// it is missing). A fill has to be guarded the same way, and this is not
// hypothetical for the trigger step: it writes several lines about the same child
// order while it watches it, and the last one written wins.
func TestFillIsMonotone(t *testing.T) {
	at := time.Date(2026, 7, 30, 4, 30, 0, 0, time.UTC)
	live := Artifact{Kind: KindOrder, ID: "child-1", Symbol: "TSLA", CreatedAt: at}
	filled := live
	filled.Filled, filled.FilledAt = true, at.Add(time.Second)

	for _, c := range []struct {
		name string
		then Artifact
	}{
		{"a later line that does not know about the fill", live},
		{"a later line that cancels it instead", func() Artifact {
			a := live
			a.Cancelled, a.CancelledAt = true, at.Add(2*time.Second)
			return a
		}()},
	} {
		t.Run(c.name, func(t *testing.T) {
			entries := []Entry{
				filledEntry(StepConditionalTrigger, live),
				filledEntry(StepConditionalTrigger, filled),
				filledEntry(StepCosts, c.then),
			}
			if got := Outstanding(entries); len(got) != 0 {
				t.Fatalf("Outstanding = %+v, want empty; %s revived it", got, c.name)
			}
		})
	}
}

// TestRecordsWrittenBeforeFillsExistedKeepTheirJudgement is the toggle-off case
// (§0.3) and it is the reason the field is added rather than an existing one
// reinterpreted. Every line on both of the operator's real records was written
// without it, and none of their verdicts may move.
func TestRecordsWrittenBeforeFillsExistedKeepTheirJudgement(t *testing.T) {
	at := time.Date(2026, 7, 28, 4, 30, 0, 0, time.UTC)
	entries := []Entry{
		filledEntry(StepOrderCancel, Artifact{Kind: KindOrder, ID: "o-1", Symbol: "005930", CreatedAt: at}),
		filledEntry(StepConditionalRegister, Artifact{
			Kind: KindConditional, ID: "c-1", Symbol: "005930", CreatedAt: at,
			Deliberate: true,
		}),
		filledEntry(StepSellBoundary, Artifact{
			Kind: KindOrder, ID: "o-2", Symbol: "005930", CreatedAt: at,
			Cancelled: true, CancelledAt: at.Add(time.Minute),
		}),
	}

	out := Outstanding(entries)
	if len(out) != 2 {
		t.Fatalf("Outstanding = %+v, want the uncancelled order and the deliberate conditional", out)
	}
	if out[0].ID != "o-1" || out[1].ID != "c-1" {
		t.Errorf("Outstanding = %+v, want [o-1 c-1] in record order", out)
	}
	for _, a := range out {
		if a.Filled {
			t.Errorf("%s came back Filled although no line ever said so", a.ID)
		}
	}

	// The cleanup rule is the one that decides what a person is asked to approve,
	// so it is checked separately rather than inferred from Outstanding.
	pending := PendingCleanup(entries)
	if len(pending) != 1 || pending[0].ID != "o-1" {
		t.Errorf("PendingCleanup = %+v, want only the leftover order — the conditional is held until "+
			"conditional-cancel and no such verdict is on this record", pending)
	}
}

// TestAFilledFieldSurvivesTheRecordRoundTrip. The field has to be readable back
// out of the file, and a reader from before it existed has to keep working — which
// is why it is omitempty on a format version that does not move.
func TestAFilledFieldSurvivesTheRecordRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/" + FileName
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	at := time.Date(2026, 7, 30, 4, 30, 0, 0, time.UTC)
	if err := rec.Append(filledEntry(StepConditionalTrigger, Artifact{
		Kind: KindOrder, ID: "child-1", Symbol: "TSLA", CreatedAt: at,
		Filled: true, FilledAt: at.Add(time.Second),
	})); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	back, err := LoadEntries(path)
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	if len(back) != 1 {
		t.Fatalf("read %d entries, want 1", len(back))
	}
	if got := back[0].Artifacts[0]; !got.Filled || !got.FilledAt.Equal(at.Add(time.Second)) {
		t.Errorf("artifact = %+v, want the fill and its time to survive the round trip", got)
	}
	if back[0].FormatVersion != RecordFormatVersion {
		t.Errorf("format version = %d, want %d — adding an omitempty field does not break old readers "+
			"and must not claim to", back[0].FormatVersion, RecordFormatVersion)
	}
}
