package reconcile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// external_test.go is the external-position branch of task 6.3
// (reconciliation delta: 외부 포지션의 투영 편입).
//
// Three things have to be true at once and they pull in different directions:
// the projection must stop lying about what the account holds (because the exit
// sizing reads it), the position must not become something the exit policy
// manages (because nobody set a stop on somebody else's shares), and a person
// has to be told (because the engine just discovered trading it did not do).

// recordingAlerter captures the alerts the ingest and the convergence raised.
type recordingAlerter struct {
	alerts []reconcile.ExternalPositionAlert
	closed []reconcile.ManagedCloseAlert
	err    error
}

func (a *recordingAlerter) ExternalPositionFound(_ context.Context, alert reconcile.ExternalPositionAlert) error {
	a.alerts = append(a.alerts, alert)
	return a.err
}

func (a *recordingAlerter) ManagedPositionClosedExternally(_ context.Context,
	alert reconcile.ManagedCloseAlert) error {
	a.closed = append(a.closed, alert)
	return a.err
}

// externalDiff is a comparison that found one holding the engine never bought.
func externalDiff(symbol, market, quantity string) reconcile.Diff {
	return reconcile.Diff{
		AsOf:       "2026-03-30T00:30:00Z",
		AccountRef: "acct-7",
		ExternalPos: []reconcile.ExternalPosition{
			{
				Holding:    reconcile.Holding{Symbol: symbol, Market: market, Quantity: quantity},
				Provenance: reconcile.ProvenanceExternal,
			},
		},
	}
}

func ingestorOn(j *journal.Journal, alerter reconcile.Alerter) *reconcile.Ingestor {
	return &reconcile.Ingestor{Journal: j, Alert: alerter, AccountRef: "acct-7"}
}

// TestAnExternalHoldingIsFoldedIntoTheProjection is the scenario: the account
// holds something with no local instance, so an EXTERNAL adjustment folds it in
// and the local belief stops being wrong.
func TestAnExternalHoldingIsFoldedIntoTheProjection(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	alerter := &recordingAlerter{}

	report, err := ingestorOn(j, alerter).IngestExternalPositions(ctx, externalDiff("TSLA", "us", "3"))
	if err != nil {
		t.Fatalf("IngestExternalPositions: %v", err)
	}
	if len(report.Folded) != 1 {
		t.Fatalf("folded = %+v, want the one external holding", report.Folded)
	}
	folded := report.Folded[0]
	if folded.Symbol != "TSLA" || folded.Quantity != "3" || !folded.Applied {
		t.Fatalf("folded = %+v, want TSLA 3 newly applied", folded)
	}

	// The projection now agrees with the account, which is what the exit sizing
	// reads (SHALL — 청산 수량 판정이 실제 보유를 알아야 한다).
	local, err := reconcile.LocalStateFromJournal(ctx, j, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if local.Positions["TSLA"] != "3" {
		t.Fatalf("local belief = %q, want the account's 3", local.Positions["TSLA"])
	}

	// The adjustment records what it was and where it came from.
	adjustments, err := j.PositionAdjustments(ctx, folded.PositionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(adjustments) != 1 {
		t.Fatalf("adjustments = %+v, want the one that folded it in", adjustments)
	}
	if adjustments[0].Kind != journal.AdjustmentExternal {
		t.Errorf("kind = %q, want EXTERNAL", adjustments[0].Kind)
	}
	if adjustments[0].BrokerAsOf != "2026-03-30T00:30:00Z" {
		t.Errorf("broker as-of = %q, want the snapshot's", adjustments[0].BrokerAsOf)
	}
}

// TestAnExternalPositionIsNotAnExitPolicyTarget: no decision justifies it, so it
// has no entry stop, so there is no t0 baseline and the ratchet has nothing to
// protect it with (design D4/D5). NULL entry_decision_id is the marking.
func TestAnExternalPositionIsNotAnExitPolicyTarget(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)

	report, err := ingestorOn(j, &recordingAlerter{}).
		IngestExternalPositions(ctx, externalDiff("TSLA", "us", "3"))
	if err != nil {
		t.Fatalf("IngestExternalPositions: %v", err)
	}
	if report.Folded[0].ExitEligible {
		t.Fatal("an externally acquired position must not be an exit-policy target")
	}

	stored, err := j.LookupPosition(ctx, report.Folded[0].PositionID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.EntryDecisionID != "" {
		t.Fatalf("entry decision = %q, want NULL — no decision justifies it", stored.EntryDecisionID)
	}
	if stored.ExitEligible() {
		t.Fatal("the stored row must report itself exempt from the exit policy")
	}
	if stored.Quantity != "3" || stored.State != journal.PositionOpen {
		t.Fatalf("stored = (%s, %s), want (OPEN, 3)", stored.State, stored.Quantity)
	}
}

// TestFoldingInAnExternalHoldingAlerts: the engine just discovered trading it
// did not do, and that is never something to record silently.
func TestFoldingInAnExternalHoldingAlerts(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	alerter := &recordingAlerter{}

	if _, err := ingestorOn(j, alerter).
		IngestExternalPositions(ctx, externalDiff("TSLA", "us", "3")); err != nil {
		t.Fatalf("IngestExternalPositions: %v", err)
	}
	if len(alerter.alerts) != 1 {
		t.Fatalf("alerts = %+v, want one", alerter.alerts)
	}
	alert := alerter.alerts[0]
	if alert.Symbol != "TSLA" || alert.Quantity != "3" || alert.AccountRef != "acct-7" {
		t.Fatalf("alert = %+v, want the holding it folded in", alert)
	}
	if alert.ExitEligible {
		t.Error("the alert must tell the operator the position is unprotected by the exit policy")
	}
	if alert.PositionID == "" || alert.BrokerAsOf == "" {
		t.Errorf("alert = %+v, want the instance and the as-of an operator can look up", alert)
	}
}

// TestReIngestingTheSameHoldingAlertsOnce: the reconciliation loop runs every 30
// seconds, and an alert that repeats on every pass is an alert nobody reads.
// The adjustment id is derived from what the adjustment is, so the second pass
// recognises it as already applied.
func TestReIngestingTheSameHoldingAlertsOnce(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	alerter := &recordingAlerter{}
	in := ingestorOn(j, alerter)
	diff := externalDiff("TSLA", "us", "3")

	if _, err := in.IngestExternalPositions(ctx, diff); err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	second, err := in.IngestExternalPositions(ctx, diff)
	if err != nil {
		t.Fatalf("second ingest: %v", err)
	}
	if len(second.Folded) != 1 || second.Folded[0].Applied {
		t.Fatalf("second ingest = %+v, want it recognised as already applied", second.Folded)
	}
	if len(alerter.alerts) != 1 {
		t.Fatalf("alerts = %d, want one — an alert per poll is an alert nobody reads", len(alerter.alerts))
	}

	// And it is one instance, not one per pass.
	positions, err := j.Positions(ctx, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 1 {
		t.Fatalf("positions = %+v, want the one folded instance", positions)
	}
}

// staleStore is a journal whose world moved between the read and the commit.
type staleStore struct {
	reconcile.AdjustmentStore
	calls int
}

func (s *staleStore) ApplyPositionAdjustment(context.Context, journal.AdjustmentRequest) (journal.AdjustmentResult, error) {
	s.calls++
	return journal.AdjustmentResult{}, &journal.StaleAdjustmentError{
		AccountRef: "acct-7", Symbol: "TSLA", Invariant: "fill watermark",
		Expected: "0", Actual: "1",
	}
}

// TestAStaleIngestIsDiscardedForRecollection: a fill landed between reading the
// account and committing the adjustment, so the adjustment was computed for a
// position that is not the one it would be applied to. Discard and re-collect
// (SHALL — 어긋나면 조정을 폐기하고 재수집한다), rather than double-counting.
func TestAStaleIngestIsDiscardedForRecollection(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	alerter := &recordingAlerter{}
	in := &reconcile.Ingestor{
		Journal:    &staleStore{AdjustmentStore: j},
		Alert:      alerter,
		AccountRef: "acct-7",
	}

	report, err := in.IngestExternalPositions(ctx, externalDiff("TSLA", "us", "3"))
	if !errors.Is(err, journal.ErrAdjustmentStale) {
		t.Fatalf("err = %v, want the stale view surfaced for re-collection", err)
	}
	if len(report.Folded) != 0 {
		t.Fatalf("folded = %+v, want nothing folded from a stale view", report.Folded)
	}
	if len(alerter.alerts) != 0 {
		t.Fatalf("alerts = %+v, want none for an adjustment that was discarded", alerter.alerts)
	}
}

// TestAHoldingWithNoMarketIsRefusedRatherThanGuessed: the holdings snapshot's
// market dimension is `[미측정]`, so a holding may arrive without one. The
// projection is keyed by market, and inventing one would put the position on a
// venue the operator would then go looking for it on. The caller declares the
// account's venue instead.
func TestAHoldingWithNoMarketIsRefusedRatherThanGuessed(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)

	if _, err := ingestorOn(j, &recordingAlerter{}).
		IngestExternalPositions(ctx, externalDiff("TSLA", "", "3")); err == nil {
		t.Fatal("a holding with no market and no declared default must be refused")
	}

	// With the account's venue declared, the same holding folds in under it.
	in := ingestorOn(j, &recordingAlerter{})
	in.DefaultMarket = "kr"
	report, err := in.IngestExternalPositions(ctx, externalDiff("TSLA", "", "3"))
	if err != nil {
		t.Fatalf("IngestExternalPositions: %v", err)
	}
	if report.Folded[0].Market != "kr" {
		t.Fatalf("market = %q, want the declared default", report.Folded[0].Market)
	}
}

// TestNothingExternalFoldsNothing: the ordinary pass, on which the ingest must
// write nothing and alert nobody.
func TestNothingExternalFoldsNothing(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	alerter := &recordingAlerter{}

	report, err := ingestorOn(j, alerter).IngestExternalPositions(ctx, reconcile.Diff{
		AsOf: "2026-03-30T00:30:00Z", AccountRef: "acct-7", Matched: 2,
	})
	if err != nil {
		t.Fatalf("IngestExternalPositions: %v", err)
	}
	if len(report.Folded) != 0 || len(alerter.alerts) != 0 {
		t.Fatalf("report = %+v alerts = %+v, want an untouched account", report, alerter.alerts)
	}
	positions, err := j.Positions(ctx, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 0 {
		t.Fatalf("positions = %+v, want none written", positions)
	}
}

// TestAnIngestWithoutAnAsOfIsRefused: an adjustment with no as-of cannot be
// ordered against the fills it competes with, and the journal refuses one. The
// ingest refuses first, naming the snapshot rather than the row.
func TestAnIngestWithoutAnAsOfIsRefused(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	diff := externalDiff("TSLA", "us", "3")
	diff.AsOf = ""

	if _, err := ingestorOn(j, &recordingAlerter{}).IngestExternalPositions(ctx, diff); err == nil {
		t.Fatal("an ingest from a snapshot with no as-of must be refused")
	}
}
