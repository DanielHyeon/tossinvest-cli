package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// TestA110ChangingReconcileDisputesDoNotSuppressAnUncoveredAdoption joins the
// incident-shaped tracker sequence to the real reconciliation driver.  The
// ordinary blocks are for other symbols; once their credited rows are released,
// they must not be re-labelled as an account-wide permanent outage that stops a
// healthy candidate before its quote read and adoption transaction.
//
// The same chain then runs the separately scheduled exit observer once against
// that adopted position.  Adoption must first leave the canonical snapshot in
// the non-actionable SEED state; only the observer's persisted EVALUATED
// snapshot may make the operator view render actionable exit lines.
func TestA110ChangingReconcileDisputesDoNotSuppressAnUncoveredAdoption(t *testing.T) {
	ctx := context.Background()
	h := newDriverHarness(t, nil)
	h.holds("005930", "10", "55000", 70000)

	observe := func(diff reconcile.Diff) reconcile.Outcome {
		t.Helper()
		out, err := h.tracker.Observe(ctx, diff)
		if err != nil {
			t.Fatalf("Tracker.Observe(%+v): %v", diff, err)
		}
		return out
	}
	stamp := func() string { return h.clk.Now().UTC().Format(time.RFC3339) }
	advance := func() { h.clk.Advance(30 * time.Second) }

	first := reconcile.Diff{AsOf: stamp(), AccountRef: reconcileAccount,
		Quantities: []reconcile.QuantityMismatch{{Symbol: "AAPL", Local: "10", Broker: "4"}}}
	observe(first)
	h.tracker.AdjustmentApplied(first.AsOf, "AAPL")
	advance()

	second := reconcile.Diff{AsOf: stamp(), AccountRef: reconcileAccount,
		Quantities: []reconcile.QuantityMismatch{{Symbol: "MSFT", Local: "7", Broker: "2"}}}
	observe(second)
	h.tracker.AdjustmentApplied(second.AsOf, "MSFT")
	advance()

	third := reconcile.Diff{AsOf: stamp(), AccountRef: reconcileAccount,
		Quantities: []reconcile.QuantityMismatch{{Symbol: "TSLA", Local: "3", Broker: "0"}}}
	if out := observe(third); out.Permanent || h.tracker.Permanent() {
		t.Fatalf("changing disputes produced an account-wide permanent gate: %+v", out)
	}
	advance()
	observe(reconcile.Diff{AsOf: stamp(), AccountRef: reconcileAccount, Matched: 1})
	if h.tracker.Permanent() {
		t.Fatalf("false account permanent survived credited ordinary releases: %+v", h.tracker.Blocks())
	}

	cycle := h.cycle()
	if cycle.Err != nil {
		t.Fatalf("reconcile/adoption cycle: %v", cycle.Err)
	}
	if cycle.Adopted != 1 || h.prices.calls != 1 {
		t.Fatalf("changing unrelated disputes suppressed quote/adoption: cycle=%+v priceReads=%d", cycle, h.prices.calls)
	}
	if len(h.prices.asked) != 1 || len(h.prices.asked[0]) != 1 || h.prices.asked[0][0] != "005930" {
		t.Fatalf("adoption quote request = %+v, want exactly [005930]", h.prices.asked)
	}
	p := h.position("005930")
	if !p.Adopted() {
		t.Fatalf("candidate was not adopted after the unrelated ordinary blocks cleared: %+v", p)
	}
	resultFor := func() journal.ExitStateResult {
		t.Helper()
		results, err := h.journal.OpenExitStateResults(ctx, reconcileAccount)
		if err != nil {
			t.Fatalf("OpenExitStateResults: %v", err)
		}
		for _, result := range results {
			if result.State.PositionID == p.ID {
				return result
			}
		}
		t.Fatalf("no exit state result for adopted position %s: %+v", p.ID, results)
		return journal.ExitStateResult{}
	}

	before := resultFor()
	if before.State.SnapshotStatus != journal.SnapshotStatusSeed ||
		before.State.Snapshot.UnknownReason != "not_evaluated_yet" ||
		before.State.Snapshot.Snapshot != nil {
		t.Fatalf("adoption must persist a non-actionable SEED snapshot before exit observation: %+v", before.State)
	}
	beforeLine := operatorview.BuildExitLine(operatorview.Source{
		UnknownReason: before.State.Snapshot.UnknownReason,
	})
	if !beforeLine.Unknown() || beforeLine.InitialStop != "—" || beforeLine.NextTarget != "—" {
		t.Fatalf("SEED must not render actionable exit lines: %+v", beforeLine)
	}

	// Reuse the established exit-observer test components, but bind them to the
	// same driver journal and same adopted position.  70001 is strictly above
	// the seeded high-water, so it requires an EVALUATED snapshot to be recorded;
	// it is still far below the first 0.4R proposal threshold, so no order reaches
	// the submitter.
	exitPrices := &fakePrices{last: map[string]float64{"005930": 70001}}
	exitGate := execgw.NewEntryGate(h.clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryPrice: 15 * time.Second,
	})
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: h.journal, Clock: h.clk, AccountRef: reconcileAccount,
		Policy: exitPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a110-test/v1",
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	submit := &fakeSubmitter{}
	observer, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: h.journal, Prices: exitPrices,
		Retrier: &execgw.Retrier{Clock: h.clk, Gate: exitGate,
			Policy: execgw.RetryPolicy{MaxAttempts: 1, Budget: time.Second}},
		Issuer: guardian, Submit: submit, Alerts: &fakeAlerts{}, Costs: costs.DefaultModel(),
		Floor: &fakeFloor{}, SLO: &fakeSLO{}, Escalate: h.journal,
		AccountRef: reconcileAccount, Clock: h.clk,
		NewID: func() string { return "a110-exit-judgement" },
	})
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	observed := observer.ObserveOnce(ctx)
	if observed.Err != nil || observed.Observed != 1 || observed.Judged != 1 || observed.Proposed != 0 {
		t.Fatalf("exit observer did not evaluate the adopted position exactly once without proposing: %+v", observed)
	}
	if len(submit.places) != 0 {
		t.Fatalf("sub-threshold 70001 evaluation unexpectedly reached the order adapter: %+v", submit.places)
	}
	if exitPrices.calls != 1 || len(exitPrices.asked) != 1 || len(exitPrices.asked[0]) != 1 ||
		exitPrices.asked[0][0] != "005930" {
		t.Fatalf("exit observer quote request = calls:%d asked:%+v, want one [005930]", exitPrices.calls, exitPrices.asked)
	}

	after := resultFor()
	if after.State.SnapshotStatus != journal.SnapshotStatusEvaluated || after.State.Snapshot.Snapshot == nil {
		t.Fatalf("exit observer did not persist canonical evaluated snapshot: %+v", after.State)
	}
	afterLine := operatorview.BuildExitLine(operatorview.Source{
		Snapshot:          &after.State.Snapshot.Snapshot.Line,
		RemainingQuantity: p.Quantity,
		ObservationSource: after.State.Snapshot.Snapshot.ObservationSource,
		ObservedAt:        after.State.Snapshot.Snapshot.ObservedAt,
		EffectiveSource:   "persisted effective snapshot",
	})
	if !afterLine.Fresh() || afterLine.InitialStop == "—" || afterLine.CurrentProtection == "—" {
		t.Fatalf("only evaluated canonical snapshot may render actionable exit lines: %+v", afterLine)
	}
}
