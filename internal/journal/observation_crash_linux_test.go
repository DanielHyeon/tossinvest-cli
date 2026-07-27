//go:build linux

package journal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// observation_crash_linux_test.go is change add-net-rr-measurement task 7.2's
// crash requirement, and it is the test that justifies design D6's whole shape.
//
// # The window this proves is real, and then proves is closeable
//
// The observation is written outside the issuance transaction. That is deliberate
// — inside, a full disk would roll the decision back and a measurement defect
// would refuse a trade the chain allowed — and the price is a genuine window: the
// decision commits, the process dies, and the observation never lands.
//
// Round 2 accepted that price on the condition that the resulting gap is
// *recoverable*. This test pays the price for real: a child process commits an
// entry decision and is SIGKILLed before it can write the observation, with no
// deferred close and no flush. Then the parent shows that the gap is detectable by
// anti-join and rebuildable from the preimage.
//
// A fake would not do. The claim is about what survives a process that had no
// chance to tidy up, and only a real kill produces that state.

const (
	crashModeObservationGap = "observation-gap"
	crashDecisionID         = "decision-crash-1"
)

// TestACrashBetweenCommitAndObservationIsRecoverable walks the whole window.
func TestACrashBetweenCommitAndObservationIsRecoverable(t *testing.T) {
	if os.Getenv(crashEnvMode) == crashModeObservationGap {
		crashChildObservationGap()
		return
	}

	path := filepath.Join(t.TempDir(), "journal.db")
	runCrashChild(t, "TestACrashBetweenCommitAndObservationIsRecoverable",
		crashModeObservationGap, path)

	ctx := context.Background()
	j := openTestJournalAt(t, path)

	// (a) The window is real: the decision committed, the observation did not.
	dec, err := j.LookupDecision(ctx, crashDecisionID)
	if err != nil {
		t.Fatalf("the decision committed before the crash is missing: %v", err)
	}
	if dec.SafetyClass != SafetyClassExposureRaising {
		t.Fatalf("decision came back as %s", dec.SafetyClass)
	}
	rows, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("the fixture did not reproduce the window: %d observations survived", len(rows))
	}

	// (b) The gap is detectable, by anti-join against the contract table.
	opts := GapScanOptions{
		WriteDeadline:  5 * time.Minute,
		Cycle:          time.Hour,
		PruningHorizon: 30 * 24 * time.Hour,
	}
	gap, err := j.DetectMissingEntryObservations(ctx, dec.IssuedAt.Add(time.Hour), opts)
	if err != nil {
		t.Fatalf("detecting the gap: %v", err)
	}
	if len(gap.Missing) != 1 {
		t.Fatalf("missing = %d, want the one decision whose observation was lost", len(gap.Missing))
	}
	missing := gap.Missing[0]
	if missing.DecisionID != crashDecisionID {
		t.Errorf("detected %q", missing.DecisionID)
	}

	// (c) What the preimage restores survived the kill byte-for-byte. This is what
	// makes the reconstruction deterministic rather than a guess.
	if missing.Preimage.EntryPrice != "70000" ||
		missing.Preimage.StopPrice != "68000" ||
		missing.Preimage.TargetPrice != "74000" {
		t.Errorf("the geometry did not survive: %+v", missing.Preimage)
	}
	if missing.Preimage.Market != "kr" || missing.Preimage.Symbol != "005930" ||
		missing.Preimage.Quantity != "10" || missing.Preimage.PolicyVersion != "crash-test" {
		t.Errorf("the venue, size or policy version did not survive: %+v", missing.Preimage)
	}
	if !missing.IssuedAt.Equal(dec.IssuedAt) {
		t.Errorf("issued_at = %s, want the decision's %s", missing.IssuedAt, dec.IssuedAt)
	}

	// (d) The rebuild lands, marked, with both instants and today's fingerprint.
	rebuiltAt := dec.IssuedAt.Add(time.Hour)
	if err := j.RecordEntryObservation(ctx, EntryObservation{
		ID:         "obs-rebuilt",
		AccountRef: missing.Preimage.AccountRef,
		Market:     missing.Preimage.Market,
		Symbol:     missing.Preimage.Symbol,
		EntryPrice: missing.Preimage.EntryPrice,
		StopPrice:  missing.Preimage.StopPrice,
		// Recomputed under today's model, not restored — the rates the original
		// verdict used are nowhere in the preimage.
		TargetPrice:          missing.Preimage.TargetPrice,
		BreakEvenPrice:       "70351.41",
		GrossRewardRisk:      "2",
		NetRewardRisk:        "1.7031",
		CostScope:            CostScopeFeeTaxOnly,
		CostModelFingerprint: "fp-at-reconstruction",
		Outcome:              OutcomeAllowedIssued,
		DecisionID:           missing.DecisionID,
		ObservedAt:           rebuiltAt,
		IssuedAt:             missing.IssuedAt,
		Reconstructed:        true,
		ReconstructedAt:      rebuiltAt,
	}); err != nil {
		t.Fatalf("rebuilding the lost observation: %v", err)
	}

	rows, err = j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows after the rebuild = %d, want 1", len(rows))
	}
	row := rows[0]
	if !row.Reconstructed {
		t.Error("a rebuilt row must be distinguishable from one a verdict wrote")
	}
	if row.IssuedAt.IsZero() || row.ReconstructedAt.IsZero() {
		t.Errorf("a rebuilt row carries both instants: %+v", row)
	}
	if row.IssuedAt.Equal(row.ReconstructedAt) {
		t.Error("the two instants are recorded because they differ")
	}

	// (e) The gap is closed: a second scan offers nothing, so a repeated run
	// cannot double-count this entry.
	gap, err = j.DetectMissingEntryObservations(ctx, rebuiltAt.Add(time.Hour), opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(gap.Missing) != 0 {
		t.Errorf("the rebuilt gap is still being offered: %+v", gap.Missing)
	}
	if gap.LapsedBeyondHorizon != 0 {
		t.Errorf("lapsed = %d, want 0", gap.LapsedBeyondHorizon)
	}

	if err := j.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after the crash and the rebuild: %v", err)
	}
}

// crashChildObservationGap commits an entry decision the way IssueEntry does, then
// dies in the instant the observation write would have occupied.
func crashChildObservationGap() {
	j := openCrashChildJournal()
	ctx := context.Background()

	version, err := j.ReservationVersion(ctx, "acct-1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "crash child: ReservationVersion: %v\n", err)
		os.Exit(3)
	}
	issued := j.Now()
	if _, err := j.RecordDecisionAndReserve(ctx, IssueRequest{
		Decision: DecisionRequest{
			ID:          crashDecisionID,
			AccountRef:  "acct-1",
			SafetyClass: SafetyClassExposureRaising,
			Kind:        KindPlace,
			Preimage: RiskIntent{
				AccountRef: "acct-1", Market: "kr", Symbol: "005930", Side: "BUY",
				Quantity: "10", EntryPrice: "70000", StopPrice: "68000",
				TargetPrice: "74000", PolicyVersion: "crash-test",
			},
			LimitsJSON: `{"max_notional":"1000000"}`,
			Nonce:      "nonce-crash-1",
			IssuedAt:   issued,
			ExpiresAt:  issued.Add(60 * time.Second),
		},
		Reserve: ReserveRequest{
			SnapshotAsOf:    issued,
			ObservedVersion: version,
			SnapshotUsage: []AggregateAmount{
				{Kind: ReservationKindOpenExposure, Amount: "0", Currency: "KRW"},
			},
			Limits: []AggregateAmount{
				{Kind: ReservationKindOpenExposure, Amount: "1000000", Currency: "KRW"},
			},
			Reservations: []ReservationRequest{
				{ID: "res-crash-1", Kind: ReservationKindOpenExposure,
					Amount: "100", Currency: "KRW"},
			},
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "crash child: RecordDecisionAndReserve: %v\n", err)
		os.Exit(3)
	}
	// The commit returned. This is precisely the instant IssueEntry hands the
	// observation to the observer — and the process dies instead.
	kill()
}
