package verifylive

// record_test.go covers the evidence file.
//
// The two properties that matter are the ones a crash exercises: a torn final
// line must not lose the steps before it, and an artifact cancelled in one entry
// must not come back as outstanding because a later entry mentions it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func tempRecord(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), FileName)
}

func TestRecorderRoundTrips(t *testing.T) {
	path := tempRecord(t)
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	now := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	for _, id := range []StepID{StepReadFixtures, StepIdempotency} {
		if err := rec.Append(Entry{
			StepID: id, Verdict: VerdictPass, StartedAt: now, FinishedAt: now,
			AccountRef:   "****8901",
			Observations: []Observation{{Key: "k", Value: "v"}},
		}); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	entries, err := LoadEntries(path)
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("loaded %d entries, want 2", len(entries))
	}
	if entries[0].FormatVersion != RecordFormatVersion || entries[0].Kind != KindStep {
		t.Errorf("defaults were not filled in: %+v", entries[0])
	}
	if !Settled(entries, StepIdempotency) || !Passed(entries, StepIdempotency) {
		t.Error("a recorded pass is not being read back as settled and passed")
	}
	if Settled(entries, StepOrderAmend) {
		t.Error("a step with no entry reported itself settled")
	}
}

// TestLoadEntriesToleratesATornFinalLine: a crash during an append must cost the
// step that was being written and nothing before it. Losing an earlier step would
// mean re-placing a live order for a measurement already made.
func TestLoadEntriesToleratesATornFinalLine(t *testing.T) {
	path := tempRecord(t)
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	if err := rec.Append(Entry{StepID: StepReadFixtures, Verdict: VerdictPass}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	rec.Close()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("reopening: %v", err)
	}
	if _, err := f.WriteString(`{"step_id":"idempotency","verd`); err != nil {
		t.Fatalf("writing the torn line: %v", err)
	}
	f.Close()

	entries, err := LoadEntries(path)
	if err != nil {
		t.Fatalf("LoadEntries must tolerate a torn final line: %v", err)
	}
	if len(entries) != 1 || entries[0].StepID != StepReadFixtures {
		t.Fatalf("entries = %+v, want only the intact first step", entries)
	}
}

func TestLoadEntriesRejectsANewerFormat(t *testing.T) {
	path := tempRecord(t)
	if err := os.WriteFile(path, []byte(`{"format_version":99,"step_id":"x"}`+"\n"), 0o600); err != nil {
		t.Fatalf("writing: %v", err)
	}
	if _, err := LoadEntries(path); err == nil {
		t.Fatal("a record from a newer build must be refused, not partially read")
	}
}

func TestLoadEntriesOnAMissingFile(t *testing.T) {
	entries, err := LoadEntries(filepath.Join(t.TempDir(), "nothing.jsonl"))
	if err != nil {
		t.Fatalf("a missing record is a normal state, not an error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %v, want none", entries)
	}
}

// TestOutstandingNetsCreationAgainstCancellation is the check the end of a run
// depends on: it decides whether the tool can say the account was left clean.
func TestOutstandingNetsCreationAgainstCancellation(t *testing.T) {
	now := time.Now().UTC()
	entries := []Entry{
		{StepID: StepIdempotency, Artifacts: []Artifact{
			{Kind: "order", ID: "ord-1", Symbol: "005930", CreatedAt: now},
			{Kind: "order", ID: "ord-1", Symbol: "005930", CancelledAt: now, Cancelled: true},
		}},
		{StepID: StepOrderCancel, Artifacts: []Artifact{
			{Kind: "order", ID: "ord-2", Symbol: "005930", CreatedAt: now},
		}},
		{StepID: StepConditionalRegister, Artifacts: []Artifact{
			{Kind: "conditional-order", ID: "co-1", Symbol: "005930", CreatedAt: now, Deliberate: true},
		}},
	}
	out := Outstanding(entries)
	if len(out) != 2 {
		t.Fatalf("Outstanding = %+v, want the uncancelled order and the deliberate conditional", out)
	}
	if out[0].ID != "ord-2" || out[1].ID != "co-1" {
		t.Errorf("Outstanding = %+v", out)
	}
	if leftovers := undeliberate(out); len(leftovers) != 1 || leftovers[0].ID != "ord-2" {
		t.Errorf("undeliberate = %+v, want only ord-2: the conditional is left on purpose", leftovers)
	}
}

// TestCancellationIsMonotone. A later entry that mentions an artifact without
// knowing it was cancelled must not resurrect it — otherwise a resumed run would
// report a phantom live order and refuse to finish.
func TestCancellationIsMonotone(t *testing.T) {
	now := time.Now().UTC()
	entries := []Entry{
		{Artifacts: []Artifact{{Kind: "order", ID: "ord-1", CreatedAt: now}}},
		{Artifacts: []Artifact{{Kind: "order", ID: "ord-1", Cancelled: true, CancelledAt: now}}},
		{Artifacts: []Artifact{{Kind: "order", ID: "ord-1", CreatedAt: now, Note: "mentioned again"}}},
	}
	if out := Outstanding(entries); len(out) != 0 {
		t.Errorf("Outstanding = %+v, want none: the cancel must win", out)
	}
}

// TestRecordCarriesNoUnmaskedAccountNumber. The record ends up in transcripts.
func TestRecordCarriesNoUnmaskedAccountNumber(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("run: %v", err)
	}
	data, err := os.ReadFile(h.record)
	if err != nil {
		t.Fatalf("reading the record: %v", err)
	}
	if strings.Contains(string(data), "123-45-678901") {
		t.Error("the record contains the unmasked account number")
	}
	if !strings.Contains(string(data), "8901") {
		t.Error("the record contains no masked account reference at all")
	}
}

func TestDigestIsStableAndTruncated(t *testing.T) {
	a := Digest(map[string]string{"symbol": "005930"})
	b := Digest(map[string]string{"symbol": "005930"})
	if a != b {
		t.Errorf("Digest is not deterministic: %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "sha256:") || len(a) != len("sha256:")+32 {
		t.Errorf("Digest = %q, want a sha256: prefix and 16 bytes of hex", a)
	}
	if Digest(map[string]string{"symbol": "000660"}) == a {
		t.Error("two different bodies produced the same digest")
	}
}

func TestProcessIdentityIsUniquePerProcess(t *testing.T) {
	now := time.Now().UTC()
	a, b := NewProcess(now), NewProcess(now)
	if a.InstanceID == b.InstanceID {
		t.Fatal("two process identities collided; conditional persistence could not be enforced")
	}
	if a.PID == 0 {
		t.Error("the process identity carries no PID")
	}
}
