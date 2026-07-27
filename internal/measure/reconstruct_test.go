package measure_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/measure"
	"github.com/JungHoonGhae/tossinvest-cli/internal/measure/degrade"
)

// reconstruct_test.go is change add-net-rr-measurement task 2.5 (and 2.5c's race)
// against a real journal, because the properties under test are properties of the
// storage as much as of the job: what the preimage restores, what it cannot, and
// what the unique index does when a late write and a rebuild collide.

var reconstructInstant = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)

func openJournal(t *testing.T) (*journal.Journal, string) {
	t.Helper()
	dir := t.TempDir()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(dir, "journal.db"),
		Clock:    clock.NewFake(reconstructInstant),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j, dir
}

// issueEntry commits one EXPOSURE_RAISING decision with no observation, which is
// exactly the crash-window state the job exists to clean up after.
func issueEntry(t *testing.T, j *journal.Journal, id string) {
	t.Helper()
	ctx := context.Background()
	version, err := j.ReservationVersion(ctx, "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	issued := j.Now()
	if _, err := j.RecordDecisionAndReserve(ctx, journal.IssueRequest{
		Decision: journal.DecisionRequest{
			ID:          id,
			AccountRef:  "acct-1",
			SafetyClass: journal.SafetyClassExposureRaising,
			Kind:        journal.KindPlace,
			Preimage: journal.RiskIntent{
				AccountRef: "acct-1", Market: "kr", Symbol: "005930", Side: "BUY",
				Quantity: "10", EntryPrice: "70000", StopPrice: "68000",
				TargetPrice: "74000", PolicyVersion: "test-1",
			},
			LimitsJSON: `{"max_notional":"1000000"}`,
			Nonce:      "nonce-" + id,
			IssuedAt:   issued,
			ExpiresAt:  issued.Add(60 * time.Second),
		},
		Reserve: journal.ReserveRequest{
			SnapshotAsOf:    issued,
			ObservedVersion: version,
			SnapshotUsage: []journal.AggregateAmount{
				{Kind: journal.ReservationKindOpenExposure, Amount: "0", Currency: "KRW"},
			},
			Limits: []journal.AggregateAmount{
				{Kind: journal.ReservationKindOpenExposure, Amount: "1000000", Currency: "KRW"},
			},
			Reservations: []journal.ReservationRequest{
				{ID: "res-" + id, Kind: journal.ReservationKindOpenExposure,
					Amount: "100", Currency: "KRW"},
			},
		},
	}); err != nil {
		t.Fatalf("issuing %s: %v", id, err)
	}
}

func scanOptions() journal.GapScanOptions {
	return journal.GapScanOptions{
		WriteDeadline:  5 * time.Minute,
		Cycle:          time.Hour,
		PruningHorizon: 30 * 24 * time.Hour,
	}
}

// todaysRatios is the recompute a caller supplies. The fingerprint is deliberately
// *not* the one the original verdict would have used — that is the whole point of
// the reconstruction marker.
func todaysRatios(journal.MissingObservation) (measure.Ratios, error) {
	return measure.Ratios{
		BreakEvenPrice:  "70351.41",
		GrossRewardRisk: "2",
		NetRewardRisk:   "1.7031",
		Fingerprint:     "fp-today",
	}, nil
}

func idFor(m journal.MissingObservation) string { return "obs-rebuilt-" + m.DecisionID }

// TestReconstructionRestoresWhatThePreimageHasAndMarksWhatItDoesNot is the
// requirement's central distinction (SHALL NOT — 순 RR·실질본전·비용모델 지문은
// 복원이 아니라 재구성 시점 모델로 새로 산출하는 값이다).
func TestReconstructionRestoresWhatThePreimageHasAndMarksWhatItDoesNot(t *testing.T) {
	j, dir := openJournal(t)
	ctx := context.Background()
	counter := degrade.NewCounter(filepath.Join(dir, "counter.json"), nil)

	issueEntry(t, j, "decision-1")
	rebuiltAt := reconstructInstant.Add(time.Hour)

	report, err := measure.Run(ctx, j, rebuiltAt, scanOptions(), todaysRatios, idFor, counter)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.Found != 1 || report.Rebuilt != 1 || report.Failed != 0 {
		t.Fatalf("report = %+v, want one found and one rebuilt", report)
	}

	rows, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	got := rows[0]

	// Restored exactly, from the hashed preimage.
	if got.EntryPrice != "70000" || got.StopPrice != "68000" || got.TargetPrice != "74000" {
		t.Errorf("the geometry must come back exactly: %+v", got)
	}
	if got.Market != "kr" || got.Symbol != "005930" || got.AccountRef != "acct-1" {
		t.Errorf("the venue must come back exactly: %+v", got)
	}
	// Determined by the existence of a committed entry decision, not guessed.
	if got.Outcome != journal.OutcomeAllowedIssued {
		t.Errorf("outcome = %q, want ALLOWED_ISSUED", got.Outcome)
	}
	if got.DecisionID != "decision-1" {
		t.Errorf("decision reference = %q", got.DecisionID)
	}

	// Not restored: marked, dual-stamped, and carrying today's fingerprint.
	if !got.Reconstructed {
		t.Error("a rebuilt row must be distinguishable from one the verdict wrote")
	}
	if got.CostModelFingerprint != "fp-today" {
		t.Errorf("fingerprint = %q, want the one used to rebuild, not the original",
			got.CostModelFingerprint)
	}
	if !got.IssuedAt.Equal(reconstructInstant) {
		t.Errorf("issued_at = %s, want the decision's %s", got.IssuedAt, reconstructInstant)
	}
	if !got.ReconstructedAt.Equal(rebuiltAt) {
		t.Errorf("reconstructed_at = %s, want %s", got.ReconstructedAt, rebuiltAt)
	}
	if got.IssuedAt.Equal(got.ReconstructedAt) {
		t.Error("both instants are recorded precisely because they are different; " +
			"one row must not mix this run's fingerprint with that day's clock")
	}

	// A second run finds nothing: the row it wrote closed the gap.
	again, err := measure.Run(ctx, j, rebuiltAt.Add(time.Hour), scanOptions(), todaysRatios, idFor, counter)
	if err != nil {
		t.Fatal(err)
	}
	if again.Found != 0 {
		t.Errorf("a rebuilt gap must not be offered again: %+v", again)
	}
}

// TestLateWriteWinsOverAReconstruction is task 2.5c injected as a race: the real
// observation lands between the scan and the rebuild. One decision must keep one
// row, because a double-counted entry biases the distribution a live threshold
// will be drawn from.
func TestLateWriteWinsOverAReconstruction(t *testing.T) {
	j, dir := openJournal(t)
	ctx := context.Background()
	counter := degrade.NewCounter(filepath.Join(dir, "counter.json"), nil)

	issueEntry(t, j, "decision-1")
	rebuiltAt := reconstructInstant.Add(time.Hour)

	// racing writes the genuine observation during the recompute — after the scan
	// decided the row was missing and before the rebuild is inserted.
	racing := func(m journal.MissingObservation) (measure.Ratios, error) {
		if err := j.RecordEntryObservation(ctx, journal.EntryObservation{
			ID: "obs-live", AccountRef: "acct-1", Market: "kr", Symbol: "005930",
			EntryPrice: "70000", StopPrice: "68000", TargetPrice: "74000",
			BreakEvenPrice: "70351.41", GrossRewardRisk: "2", NetRewardRisk: "1.7031",
			CostScope: journal.CostScopeFeeTaxOnly, CostModelFingerprint: "fp-then",
			Outcome: journal.OutcomeAllowedIssued, DecisionID: m.DecisionID,
			ObservedAt: reconstructInstant, IssuedAt: reconstructInstant,
		}); err != nil {
			t.Fatalf("the racing live write: %v", err)
		}
		return todaysRatios(m)
	}

	report, err := measure.Run(ctx, j, rebuiltAt, scanOptions(), racing, idFor, counter)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.AlreadyPresent != 1 || report.Rebuilt != 0 {
		t.Fatalf("report = %+v, want the late write to have won", report)
	}
	if report.Failed != 0 {
		t.Error("losing to the real write is not a failure; it is the index working")
	}

	rows, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows for one decision = %d, want 1", len(rows))
	}
	if rows[0].ID != "obs-live" || rows[0].Reconstructed {
		t.Errorf("the surviving row must be the genuine one: %+v", rows[0])
	}
	if rows[0].CostModelFingerprint != "fp-then" {
		t.Errorf("the surviving row must keep the fingerprint from the verdict's own model, got %q",
			rows[0].CostModelFingerprint)
	}
	if counter.Snapshot().Total() != 0 {
		t.Error("a decision that ended up with exactly one row lost no measurement")
	}
}

// TestUnmeasurableIntentIsStillRecorded: a rebuild that cannot be priced today
// still writes the row. "This entry was issued and we cannot say what its net
// ratio was" is a fact the analysis needs — dropping it would make the rebuilt
// population silently exclude exactly the intents the model struggles with.
func TestUnmeasurableIntentIsStillRecorded(t *testing.T) {
	j, dir := openJournal(t)
	ctx := context.Background()
	counter := degrade.NewCounter(filepath.Join(dir, "counter.json"), nil)

	issueEntry(t, j, "decision-1")
	unpriceable := func(journal.MissingObservation) (measure.Ratios, error) {
		return measure.Ratios{Fingerprint: "fp-today"}, errors.New("no rate for this market")
	}

	report, err := measure.Run(ctx, j, reconstructInstant.Add(time.Hour), scanOptions(),
		unpriceable, idFor, counter)
	if err != nil {
		t.Fatal(err)
	}
	if report.Rebuilt != 1 || report.Unmeasurable != 1 {
		t.Fatalf("report = %+v, want one row written and counted unmeasurable", report)
	}

	rows, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].NetRewardRisk != "" || rows[0].BreakEvenPrice != "" {
		t.Errorf("an unpriceable rebuild must leave the ratios empty, not zero: %+v", rows[0])
	}
	if rows[0].EntryPrice == "" || rows[0].StopPrice == "" || rows[0].TargetPrice == "" {
		t.Errorf("the geometry is still restorable and must still be written: %+v", rows[0])
	}
}

// TestLapsedGapsAreCounted is task 2.5d reaching the counter: a decision past the
// pruning horizon is not offered for rebuilding and the loss is tallied.
func TestLapsedGapsAreCounted(t *testing.T) {
	j, dir := openJournal(t)
	ctx := context.Background()
	counter := degrade.NewCounter(filepath.Join(dir, "counter.json"), nil)

	issueEntry(t, j, "decision-old")
	opts := journal.GapScanOptions{
		WriteDeadline: 5 * time.Minute, Cycle: time.Hour, PruningHorizon: 24 * time.Hour,
	}

	report, err := measure.Run(ctx, j, reconstructInstant.Add(48*time.Hour), opts,
		todaysRatios, idFor, counter)
	if err != nil {
		t.Fatal(err)
	}
	if report.Found != 0 || report.LapsedBeyondHorizon != 1 {
		t.Fatalf("report = %+v, want nothing rebuildable and one lapsed", report)
	}
	if got := counter.Snapshot().Counts[degrade.LossLapsedBeyondHorizon]; got != 1 {
		t.Errorf("lapsed count = %d, want 1", got)
	}
}

// TestRunRefusesAnUnusableSchedule: the schedule guard reaches the caller as an
// error from Run, so a misconfigured job fails loudly at its own boundary instead
// of quietly measuring less than it claims.
func TestRunRefusesAnUnusableSchedule(t *testing.T) {
	j, dir := openJournal(t)
	counter := degrade.NewCounter(filepath.Join(dir, "counter.json"), nil)

	_, err := measure.Run(context.Background(), j, reconstructInstant, journal.GapScanOptions{
		WriteDeadline: time.Minute, Cycle: 48 * time.Hour, PruningHorizon: 24 * time.Hour,
	}, todaysRatios, idFor, counter)
	if !errors.Is(err, journal.ErrGapScanUnusable) {
		t.Fatalf("a cycle longer than the horizon must be refused, got %v", err)
	}
}
