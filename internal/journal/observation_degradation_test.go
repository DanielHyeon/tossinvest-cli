package journal

import (
	"bytes"
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/measure/degrade"
)

// observation_degradation_test.go is the round-3 P0 (change
// add-net-rr-measurement task 2.8) proved rather than described.
//
// Round 2 disposed of "observation failures must not be silent" by counting them
// in the observation table. Round 3 found that self-contradictory: an INSERT that
// fails because the storage is full or broken takes the degradation INSERT with
// it, since they are the same file. This test is the proof, and it is here rather
// than in internal/measure/degrade because this is the package with a fixture that can
// genuinely fill a database — and it is the file somebody would have to delete to
// move the count back where it does not work.
//
// internal/measure/degrade imports nothing from internal/journal, which is what
// makes the independence structural rather than a matter of care.

// TestDegradationCountSurvivesAFullObservationStore: the observation write fails
// for the canonical reason and the count of that failure is recorded anyway.
func TestDegradationCountSurvivesAFullObservationStore(t *testing.T) {
	dir := t.TempDir()
	j := openTestJournalAt(t, filepath.Join(dir, "journal.db"))
	ctx := context.Background()

	var logged bytes.Buffer
	counterPath := filepath.Join(dir, "observation-degradation.json")
	counter := degrade.NewCounter(counterPath, slog.New(slog.NewTextHandler(&logged, nil)))

	// A committed decision, so the fixture is the real shape: a verdict that
	// happened, whose measurement is about to be lost.
	if _, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "decision-1", "acct-1", "100", "0", "1000000",
			mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatal(err)
	}

	fillDatabase(t, j)

	obs := observationAt("obs-1", j.Now())
	obs.Outcome = OutcomeAllowedIssued
	obs.StoppedStep, obs.ReasonCode = "", ""
	obs.DecisionID = "decision-1"
	obs.IssuedAt = j.Now()
	obs.CostModelFingerprint = strings.Repeat("x", 2<<20)

	err := j.RecordEntryObservation(ctx, obs)
	if err == nil {
		t.Fatal("the fixture is wrong: the observation write must fail for this test to mean anything")
	}

	// The write into the *same* storage that just refused. This is the control:
	// if the count were a row here, this is where it would fail too.
	var stored int
	if scanErr := j.db.QueryRowContext(ctx,
		"SELECT count(*) FROM entry_decision_observations").Scan(&stored); scanErr != nil {
		t.Fatal(scanErr)
	}
	if stored != 0 {
		t.Fatalf("observation rows = %d, want 0 — the storage refused the write", stored)
	}

	// The independent store, which did not.
	counter.Record(degrade.LossObservationWrite, "decision_id", "decision-1", "error", err.Error())
	snap := counter.Snapshot()
	if snap.Counts[degrade.LossObservationWrite] != 1 {
		t.Fatalf("degradation count = %d, want 1", snap.Counts[degrade.LossObservationWrite])
	}
	if !snap.Durable {
		t.Error("the count must be durable: its file is not the database that filled up")
	}
	reopened := degrade.NewCounter(counterPath, nil)
	if got := reopened.Snapshot().Counts[degrade.LossObservationWrite]; got != 1 {
		t.Errorf("count read back from the independent store = %d, want 1", got)
	}

	// And the verdict it measured is untouched: the decision and its reservation
	// are still on disk, which is the whole reason the observation is outside the
	// transaction.
	if !decisionExists(t, j, "decision-1") {
		t.Error("a lost measurement must not remove the decision it measured")
	}
	if _, err := j.db.ExecContext(ctx, "PRAGMA max_page_count = 1073741823"); err != nil {
		t.Fatal(err)
	}
	if err := j.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after the failed observation: %v", err)
	}
}

// TestRefusalLossIsCountedAsUnrecoverable is design D6's asymmetry reaching the
// counter through the real API: a refusal has no decision and therefore no
// preimage, so its lost observation is gone for good and is counted under the kind
// that says so.
func TestRefusalLossIsCountedAsUnrecoverable(t *testing.T) {
	dir := t.TempDir()
	j := openTestJournalAt(t, filepath.Join(dir, "journal.db"))
	ctx := context.Background()
	counter := degrade.NewCounter(filepath.Join(dir, "counter.json"), nil)

	fillDatabase(t, j)
	refusal := observationAt("obs-refused", j.Now())
	refusal.CostModelFingerprint = strings.Repeat("x", 2<<20)

	err := j.RecordEntryObservation(ctx, refusal)
	if err == nil {
		t.Fatal("the observation write must fail for this test to mean anything")
	}
	counter.Record(degrade.LossRefusalUnrecoverable, "reason_code", refusal.ReasonCode)

	snap := counter.Snapshot()
	if snap.Counts[degrade.LossRefusalUnrecoverable] != 1 {
		t.Fatalf("unrecoverable count = %d, want 1", snap.Counts[degrade.LossRefusalUnrecoverable])
	}
	if snap.Counts[degrade.LossObservationWrite] != 0 {
		t.Error("a refusal loss must not also be counted as a rebuildable one; " +
			"the two numbers mean different things to whoever reads them")
	}

	// There is nothing to rebuild it from: the gap scan reads decisions, and a
	// refusal never wrote one.
	if _, err := j.db.ExecContext(ctx, "PRAGMA max_page_count = 1073741823"); err != nil {
		t.Fatal(err)
	}
	gap, err := j.DetectMissingEntryObservations(ctx, j.Now().Add(time.Hour), GapScanOptions{
		WriteDeadline: 5 * time.Minute, Cycle: time.Hour, PruningHorizon: 30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gap.Missing) != 0 {
		t.Errorf("a refusal must not appear as a rebuildable gap: %+v", gap.Missing)
	}
}
