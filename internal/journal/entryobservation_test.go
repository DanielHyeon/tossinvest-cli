package journal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// entryobservation_test.go is the observation table's contract (change
// add-net-rr-measurement tasks 2.1, 2.4–2.7).
//
// The theme running through it: this table exists so that a threshold can later
// be chosen from measurements rather than guesses, and every property here is
// about not corrupting that measurement — one row per decision, no row lost to a
// contract-table sweep, no row invented, and no write of it able to change a
// verdict.

func observationAt(id string, at time.Time) EntryObservation {
	return EntryObservation{
		ID:                   id,
		AccountRef:           "acct-1",
		Market:               "kr",
		Symbol:               "005930",
		EntryPrice:           "70000",
		StopPrice:            "68000",
		TargetPrice:          "74000",
		BreakEvenPrice:       "70351.41",
		GrossRewardRisk:      "2",
		NetRewardRisk:        "1.7031",
		CostScope:            CostScopeFeeTaxOnly,
		CostModelFingerprint: "fp-default",
		Outcome:              OutcomeRefusedChain,
		StoppedStep:          "min_reward_risk",
		ReasonCode:           "MIN_RR_NOT_MET",
		ObservedAt:           at,
	}
}

// TestObservationRoundTripsEveryColumn is task 2.1's enumeration: every item the
// requirement lists survives a write and a read. A column that stored but did not
// read back would make the analysis silently narrower than the spec.
func TestObservationRoundTripsEveryColumn(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	want := observationAt("obs-1", now)
	if err := j.RecordEntryObservation(ctx, want); err != nil {
		t.Fatalf("RecordEntryObservation: %v", err)
	}

	got, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("observations = %d, want 1", len(got))
	}
	if got[0] != want {
		t.Errorf("observation round trip:\n got %+v\nwant %+v", got[0], want)
	}
}

// TestObservationRecordsAllThreeOutcomes pins the enumeration that the null
// decision reference cannot express (D1 / R2-5): a chain refusal, an allowed and
// issued verdict, and an allowed verdict the ledger refused.
func TestObservationRecordsAllThreeOutcomes(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	refused := observationAt("obs-refused", now)

	issued := observationAt("obs-issued", now)
	issued.Outcome = OutcomeAllowedIssued
	issued.StoppedStep, issued.ReasonCode = "", ""
	issued.DecisionID = "decision-1"
	issued.IssuedAt = now

	issuanceRefused := observationAt("obs-issuance-refused", now)
	issuanceRefused.Outcome = OutcomeAllowedIssuanceRefused
	issuanceRefused.StoppedStep, issuanceRefused.ReasonCode = "", ""
	issuanceRefused.IssuanceReasonCode = "LIMIT_REACHED"

	for _, obs := range []EntryObservation{refused, issued, issuanceRefused} {
		if err := j.RecordEntryObservation(ctx, obs); err != nil {
			t.Fatalf("recording %s: %v", obs.ID, err)
		}
	}

	got, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]EntryObservation{}
	for _, o := range got {
		byID[o.ID] = o
	}
	if o := byID["obs-issuance-refused"]; o.Outcome != OutcomeAllowedIssuanceRefused ||
		o.IssuanceReasonCode != "LIMIT_REACHED" {
		t.Errorf("the issuance refusal reads back as %+v", o)
	}
	// The distinction the requirement is about: the two rows with no decision
	// reference are not the same fact, and nothing has to guess which is which.
	if byID["obs-refused"].Outcome == byID["obs-issuance-refused"].Outcome {
		t.Error("a chain refusal and a refused issuance must not share an outcome class")
	}
	if byID["obs-refused"].DecisionID != "" || byID["obs-issuance-refused"].DecisionID != "" {
		t.Error("neither refusal leaves a decision behind, so neither references one")
	}
}

// TestObservationRefusesAMisdescribedRow is the write API refusing rows whose
// fields contradict their own outcome. A stored contradiction would be read later
// as a fact about the market rather than as the defect it is.
func TestObservationRefusesAMisdescribedRow(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	cases := map[string]func(*EntryObservation){
		"a chain refusal that references a decision": func(o *EntryObservation) {
			o.DecisionID = "decision-1"
		},
		"a chain refusal with no rung": func(o *EntryObservation) {
			o.StoppedStep = ""
		},
		"an issued row with no decision": func(o *EntryObservation) {
			o.Outcome = OutcomeAllowedIssued
			o.StoppedStep, o.ReasonCode = "", ""
		},
		"an issued row that also stopped at a rung": func(o *EntryObservation) {
			o.Outcome = OutcomeAllowedIssued
			o.DecisionID = "decision-1"
			o.IssuedAt = now
		},
		"a refused issuance with no ledger reason": func(o *EntryObservation) {
			o.Outcome = OutcomeAllowedIssuanceRefused
			o.StoppedStep, o.ReasonCode = "", ""
		},
		"an unfingerprinted ratio": func(o *EntryObservation) {
			o.CostModelFingerprint = ""
		},
		"a cost scope this build does not measure": func(o *EntryObservation) {
			o.CostScope = "FEE_TAX_AND_SLIPPAGE"
		},
		"a reconstructed row with no reconstruction instant": func(o *EntryObservation) {
			o.Outcome = OutcomeAllowedIssued
			o.StoppedStep, o.ReasonCode = "", ""
			o.DecisionID = "decision-1"
			o.IssuedAt = now
			o.Reconstructed = true
		},
		"a reconstructed refusal": func(o *EntryObservation) {
			o.Reconstructed = true
			o.ReconstructedAt = now
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			obs := observationAt("obs-bad", now)
			mutate(&obs)
			err := j.RecordEntryObservation(ctx, obs)
			if err == nil {
				t.Fatalf("%s must be refused", name)
			}
			if !errors.Is(err, ErrInvalidRequest) {
				t.Errorf("refusal must be ErrInvalidRequest: %v", err)
			}
		})
	}
}

// TestObservationSurvivesTheDecisionItReferences is task 2.4: the reference is
// not a foreign key, so a decision row can be deleted and the observation stays —
// and, in the other direction, the delete is not blocked by it.
//
// The precedent is spent_nonces.decision_id, whose comment says exactly this:
// "pruning expired decisions must not be blocked by (or cascade into) these rows".
func TestObservationSurvivesTheDecisionItReferences(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	if _, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "decision-1", "acct-1", "100", "0", "1000000", mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatalf("issuing the decision: %v", err)
	}
	obs := observationAt("obs-1", now)
	obs.Outcome = OutcomeAllowedIssued
	obs.StoppedStep, obs.ReasonCode = "", ""
	obs.DecisionID = "decision-1"
	obs.IssuedAt = now
	if err := j.RecordEntryObservation(ctx, obs); err != nil {
		t.Fatal(err)
	}

	// The reservation does carry a foreign key, so it goes first; that is the
	// contract half. The observation must not need the same treatment.
	if _, err := j.db.ExecContext(ctx, "DELETE FROM risk_reservations WHERE decision_id = ?", "decision-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, "DELETE FROM decisions WHERE id = ?", "decision-1"); err != nil {
		t.Fatalf("pruning a decision must not be blocked by an observation: %v", err)
	}

	got, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("observations after the decision was pruned = %d, want 1 (no cascade)", len(got))
	}
	if got[0].DecisionID != "decision-1" {
		t.Errorf("the dangling reference must be kept as written, got %q", got[0].DecisionID)
	}
	// Self-contained: everything the analysis needs is still here with the
	// contract row gone.
	if got[0].EntryPrice == "" || got[0].NetRewardRisk == "" || got[0].CostModelFingerprint == "" {
		t.Errorf("the observation must be readable without the decision: %+v", got[0])
	}
}

// TestObservationIsUniquePerDecision is task 2.5c: the unique index is what makes
// a reconstruction and a late-landing real write collapse into one row instead of
// double-counting the entry.
func TestObservationIsUniquePerDecision(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	issued := func(id string) EntryObservation {
		o := observationAt(id, now)
		o.Outcome = OutcomeAllowedIssued
		o.StoppedStep, o.ReasonCode = "", ""
		o.DecisionID = "decision-1"
		o.IssuedAt = now
		return o
	}
	if err := j.RecordEntryObservation(ctx, issued("obs-live")); err != nil {
		t.Fatal(err)
	}

	rebuilt := issued("obs-rebuilt")
	rebuilt.Reconstructed = true
	rebuilt.ReconstructedAt = now.Add(time.Hour)
	err := j.RecordEntryObservation(ctx, rebuilt)
	if !errors.Is(err, ErrObservationExists) {
		t.Fatalf("a second row for one decision must be refused as ErrObservationExists, got %v", err)
	}

	got, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("rows for one decision = %d, want 1", len(got))
	}
	if got[0].ID != "obs-live" {
		t.Errorf("the row that survived is %q, want the one written first", got[0].ID)
	}

	// Refusals share no decision, so many of them coexist: the uniqueness is on
	// the reference and not on the table.
	for _, id := range []string{"obs-r1", "obs-r2", "obs-r3"} {
		if err := j.RecordEntryObservation(ctx, observationAt(id, now)); err != nil {
			t.Fatalf("refusal %s must not collide: %v", id, err)
		}
	}
}

// TestGapScanFindsOnlyDecisionsPastTheWriteDeadline is task 2.5 + 2.5c: the
// anti-join finds the unobserved issued decisions, and a write still in flight is
// not one of them.
func TestGapScanFindsOnlyDecisionsPastTheWriteDeadline(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	issuedAt := j.Now()

	for _, id := range []string{"decision-observed", "decision-missing", "decision-inflight"} {
		if _, err := j.RecordDecisionAndReserve(ctx,
			issueRequest(j, id, "acct-1", "100", "0", "1000000", mustVersion(t, j, "acct-1"))); err != nil {
			t.Fatalf("issuing %s: %v", id, err)
		}
	}
	observed := observationAt("obs-1", issuedAt)
	observed.Outcome = OutcomeAllowedIssued
	observed.StoppedStep, observed.ReasonCode = "", ""
	observed.DecisionID = "decision-observed"
	observed.IssuedAt = issuedAt
	if err := j.RecordEntryObservation(ctx, observed); err != nil {
		t.Fatal(err)
	}

	opts := GapScanOptions{
		WriteDeadline:  5 * time.Minute,
		Cycle:          time.Hour,
		PruningHorizon: 30 * 24 * time.Hour,
	}
	// Every decision was issued at the same instant, so a `now` inside the write
	// deadline must find nothing at all — including the genuinely missing one.
	early, err := j.DetectMissingEntryObservations(ctx, issuedAt.Add(time.Minute), opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(early.Missing) != 0 {
		t.Errorf("a write still inside the deadline is not a gap, got %d", len(early.Missing))
	}

	gap, err := j.DetectMissingEntryObservations(ctx, issuedAt.Add(time.Hour), opts)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]MissingObservation{}
	for _, m := range gap.Missing {
		found[m.DecisionID] = m
	}
	if len(found) != 2 {
		t.Fatalf("missing = %v, want the two unobserved decisions", found)
	}
	if _, ok := found["decision-observed"]; ok {
		t.Error("a decision that has an observation is not missing one")
	}

	// What the preimage restores, and what it does not (task 2.5). The prices,
	// the venue, the size and the policy version come back exactly; the rate set
	// the original break-even used is nowhere in it.
	m := found["decision-missing"]
	if m.Preimage.EntryPrice != "70000" || m.Preimage.StopPrice != "68000" {
		t.Errorf("the preimage must restore the geometry, got %+v", m.Preimage)
	}
	if m.Preimage.Market != "kr" || m.Preimage.Symbol != "005930" ||
		m.Preimage.Quantity != "10" || m.Preimage.PolicyVersion != "test-1" {
		t.Errorf("the preimage must restore venue, size and policy version, got %+v", m.Preimage)
	}
	if !m.IssuedAt.Equal(issuedAt.Truncate(time.Second)) {
		t.Errorf("issued_at = %s, want the decision's %s", m.IssuedAt, issuedAt)
	}
}

// TestGapScanRefusesASchedulePastThePruningHorizon is task 2.5d: a scan whose
// cycle is not shorter than the horizon loses rows by construction, so it is
// refused rather than run — the shape PruneSpentNonces uses for the same class of
// mistake.
func TestGapScanRefusesASchedulePastThePruningHorizon(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	bad := map[string]GapScanOptions{
		"a cycle as long as the horizon": {
			WriteDeadline: time.Minute, Cycle: 24 * time.Hour, PruningHorizon: 24 * time.Hour,
		},
		"a cycle longer than the horizon": {
			WriteDeadline: time.Minute, Cycle: 48 * time.Hour, PruningHorizon: 24 * time.Hour,
		},
		"a write deadline past the horizon": {
			WriteDeadline: 48 * time.Hour, Cycle: time.Hour, PruningHorizon: 24 * time.Hour,
		},
		"no write deadline": {
			Cycle: time.Hour, PruningHorizon: 24 * time.Hour,
		},
		"no horizon": {
			WriteDeadline: time.Minute, Cycle: time.Hour,
		},
	}
	for name, opts := range bad {
		t.Run(name, func(t *testing.T) {
			_, err := j.DetectMissingEntryObservations(ctx, j.Now(), opts)
			if !errors.Is(err, ErrGapScanUnusable) {
				t.Fatalf("%s must be refused as ErrGapScanUnusable, got %v", name, err)
			}
		})
	}
}

// TestGapScanCountsWhatTheScheduleIsLosing is the other half of task 2.5d: a
// decision older than the pruning horizon is past rebuilding, and the count of
// those is the measurement this schedule is giving up.
func TestGapScanCountsWhatTheScheduleIsLosing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	issuedAt := j.Now()

	if _, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "decision-old", "acct-1", "100", "0", "1000000", mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatal(err)
	}

	opts := GapScanOptions{
		WriteDeadline:  5 * time.Minute,
		Cycle:          time.Hour,
		PruningHorizon: 24 * time.Hour,
	}
	// Inside the horizon: rebuildable, and not yet lost.
	gap, err := j.DetectMissingEntryObservations(ctx, issuedAt.Add(time.Hour), opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(gap.Missing) != 1 || gap.LapsedBeyondHorizon != 0 {
		t.Fatalf("inside the horizon: missing %d, lapsed %d; want 1 and 0",
			len(gap.Missing), gap.LapsedBeyondHorizon)
	}

	// Past it: no longer offered for reconstruction, and counted as lost.
	gap, err = j.DetectMissingEntryObservations(ctx, issuedAt.Add(48*time.Hour), opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(gap.Missing) != 0 {
		t.Errorf("a decision past the horizon must not be offered for reconstruction, got %d",
			len(gap.Missing))
	}
	if gap.LapsedBeyondHorizon != 1 {
		t.Errorf("lapsed = %d, want 1", gap.LapsedBeyondHorizon)
	}
}

// TestGapScanIgnoresReductions is the entry-only scope (task 4.3 seen from the
// storage side): a RISK_REDUCING decision has no stop and no target, so an
// observation of it would be a row of nulls, and the scan must not manufacture
// one.
func TestGapScanIgnoresReductions(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	issued := j.Now()

	if _, err := j.RecordDecision(ctx, DecisionRequest{
		ID:          "decision-exit",
		AccountRef:  "acct-1",
		SafetyClass: SafetyClassRiskReducing,
		Kind:        KindPlace,
		Preimage: ReductionIntent{
			AccountRef: "acct-1", Market: "kr", Symbol: "005930", Side: "SELL",
			MaxQuantity: "10", Reason: "stop breach",
		},
		Nonce:     "nonce-exit",
		IssuedAt:  issued,
		ExpiresAt: issued.Add(60 * time.Second),
	}); err != nil {
		t.Fatal(err)
	}

	gap, err := j.DetectMissingEntryObservations(ctx, issued.Add(time.Hour), GapScanOptions{
		WriteDeadline: 5 * time.Minute, Cycle: time.Hour, PruningHorizon: 30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gap.Missing) != 0 || gap.LapsedBeyondHorizon != 0 {
		t.Errorf("an exit decision is not a missing entry observation: %+v", gap)
	}
}

// TestPruneEntryObservationsKeepsTheWindow is task 2.6: the table has a retention
// policy, it is the same 180 days the outcome table uses, and the sweep is a
// plain delete that the order path never calls.
func TestPruneEntryObservationsKeepsTheWindow(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	if EntryObservationRetention != TradeOutcomeRetention {
		t.Errorf("observation retention %s must match the outcome retention %s the analysis joins against",
			EntryObservationRetention, TradeOutcomeRetention)
	}

	old := observationAt("obs-old", now.Add(-EntryObservationRetention-time.Hour))
	fresh := observationAt("obs-fresh", now.Add(-time.Hour))
	for _, o := range []EntryObservation{old, fresh} {
		if err := j.RecordEntryObservation(ctx, o); err != nil {
			t.Fatal(err)
		}
	}

	n, err := j.PruneEntryObservations(ctx, now.Add(-EntryObservationRetention))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("pruned %d rows, want 1", n)
	}
	got, err := j.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "obs-fresh" {
		t.Errorf("after the sweep: %+v, want only obs-fresh", got)
	}
}

// TestPruneEntryObservationsDoesNotBlockTheOrderPath is the other half of 2.6:
// the sweep takes no lock the issuance transaction needs. A retention job that
// serialised against the order path would make the measurement slow the trading
// down, which is the failure mode the whole "observation is outside the
// transaction" decision exists to avoid.
func TestPruneEntryObservationsDoesNotBlockTheOrderPath(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	for i := 0; i < 20; i++ {
		obs := observationAt("obs-"+strings.Repeat("x", i)+"1", now.Add(-EntryObservationRetention-time.Hour))
		if err := j.RecordEntryObservation(ctx, obs); err != nil {
			t.Fatal(err)
		}
	}

	done := make(chan error, 1)
	go func() {
		_, err := j.PruneEntryObservations(ctx, now.Add(-EntryObservationRetention))
		done <- err
	}()
	if _, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "decision-1", "acct-1", "100", "0", "1000000", mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatalf("an issuance must not be refused because a sweep is running: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("the sweep: %v", err)
	}
	if !decisionExists(t, j, "decision-1") {
		t.Error("the issuance must have committed")
	}
}

// TestObservationWriteFailureLeavesTheLedgerAlone is task 2.7: the record call is
// its own statement, so a failure of it changes nothing else. The caller's verdict
// is not this function's to alter, and the storage layer must not make it so by
// taking anything else down with it.
func TestObservationWriteFailureLeavesTheLedgerAlone(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	if _, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "decision-1", "acct-1", "100", "0", "1000000", mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatal(err)
	}

	// A malformed observation: refused by the API before SQLite sees it.
	bad := observationAt("obs-bad", now)
	bad.CostModelFingerprint = ""
	if err := j.RecordEntryObservation(ctx, bad); err == nil {
		t.Fatal("a malformed observation must be refused")
	}

	// A storage failure: the page cap makes the insert fail the way a full disk
	// would. The row is padded so it needs a page the cap will not give it —
	// TestPrepareFailureBlocksSubmission pads an intent note for the same reason.
	fillDatabase(t, j)
	full := observationAt("obs-full", now)
	full.CostModelFingerprint = strings.Repeat("x", 2<<20)
	if err := j.RecordEntryObservation(ctx, full); err == nil {
		t.Fatal("an observation must fail when the database cannot grow")
	}

	if _, err := j.db.ExecContext(ctx, "PRAGMA max_page_count = 1073741823"); err != nil {
		t.Fatal(err)
	}
	// Neither failure touched the decision, the reservation, or the journal's
	// integrity — the verdict stands exactly as it was recorded.
	if !decisionExists(t, j, "decision-1") {
		t.Error("a failed observation must not remove the decision")
	}
	var reservations int
	if err := j.db.QueryRowContext(ctx,
		"SELECT count(*) FROM risk_reservations WHERE decision_id = 'decision-1'").Scan(&reservations); err != nil {
		t.Fatal(err)
	}
	if reservations != 1 {
		t.Errorf("reservations after a failed observation = %d, want 1", reservations)
	}
	if err := j.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after a failed observation: %v", err)
	}
}
