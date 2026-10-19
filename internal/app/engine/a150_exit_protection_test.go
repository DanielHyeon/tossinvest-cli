//go:build tossos_testseams

package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

type a150Floor struct{}

func (a150Floor) ConfirmedFloor(context.Context, string, string) (riskcalc.ConfirmedFloor, bool, error) {
	return riskcalc.ConfirmedFloor{Quantity: "2", Bound: riskcalc.FloorBoundHoldings}, true, nil
}

type a150Submitter struct {
	requests       []execgw.PlaceRequest
	unsettledCalls int
}

func (s *a150Submitter) Place(_ context.Context, req execgw.PlaceRequest) (execgw.Outcome, error) {
	s.requests = append(s.requests, req)
	return execgw.Outcome{State: journal.StateConfirmed, BrokerOrderID: "test-order"}, nil
}
func (*a150Submitter) Cancel(context.Context, execgw.CancelRequest) (execgw.Outcome, error) {
	return execgw.Outcome{}, nil
}
func (s *a150Submitter) UnsettledOnSymbol(context.Context, string, string) ([]journal.AttemptRecord, error) {
	s.unsettledCalls++
	return nil, nil
}

func a150ExitFixture(t *testing.T) (*journal.Journal, journal.Position, string, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
	proposal := proposals.forMarket(StrategyMarketKR).entries[0].authority.Proposal()
	if _, err := cycle.dispatch(ctx, deliverForTest(t, proposal)); err != nil {
		t.Fatal(err)
	}
	lease, err := j.LookupStrategyDispatchLease(ctx, spy.calls[0].Lease.LeaseID)
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", j.Path())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	p := journal.Position{ID: "a150-exit-position", AccountRef: lease.AccountRef, Market: "kr", Symbol: lease.Symbol, InstanceSeq: 1, EntryDecisionID: lease.GuardianDecisionID, State: "OPEN", Quantity: "1"}
	if _, err := db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price) VALUES(?,?,?,?,1,?,'OPEN','1','100')`, p.ID, p.AccountRef, p.Market, p.Symbol, p.EntryDecisionID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE position_campaigns SET actual_position_generation=1 WHERE id=?`, lease.CampaignID); err != nil {
		t.Fatal(err)
	}
	return j, p, lease.CampaignID, db
}

func TestA150ExitSeedsAdmittedPolicyAfterConfigurationChange(t *testing.T) {
	j, p, _, _ := a150ExitFixture(t)
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, CommonPolicy: exitpolicy.CommonLadderHybrid50}, clk: clock.NewFake(j.Now())}
	cycle := ExitCycle{}
	states, err := o.workingSet(context.Background(), &cycle)
	if err != nil || cycle.Err != nil || len(states) != 1 {
		t.Fatalf("states=%+v cycle=%+v err=%v", states, cycle, err)
	}
	if states[0].state.PolicyKind != journal.ExitPolicyRatchet {
		t.Fatalf("configuration replaced admitted policy: %+v", states[0].state)
	}
}

func TestA150RestartAfterLatchSeedsAdmittedExit(t *testing.T) {
	ctx := context.Background()
	j, p, campaign, _ := a150ExitFixture(t)
	if err := j.ReconcileStrategyProtection(ctx, campaign, p); err != nil {
		t.Fatal(err)
	}
	path, now := j.Path(), j.Now()
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(now)
	reopened, err := journal.Open(ctx, journal.Options{Path: path, Clock: clk, FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	o := &ExitObserver{opts: ExitObserverOptions{Journal: reopened, AccountRef: p.AccountRef, CommonPolicy: exitpolicy.CommonLadderHybrid50}, clk: clk}
	cycle := ExitCycle{}
	states, err := o.workingSet(ctx, &cycle)
	if err != nil || cycle.Err != nil || len(states) != 1 || states[0].state.PolicyKind != journal.ExitPolicyRatchet {
		t.Fatalf("restart states=%+v cycle=%+v err=%v", states, cycle, err)
	}
	c, err := reopened.PositionCampaign(ctx, campaign)
	if err != nil || !c.EntryBlocked {
		t.Fatalf("restart lost entry latch: %+v err=%v", c, err)
	}
}

func TestA150ForeignGenerationDoesNotSeedButExistingExitSurvives(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			ctx := context.Background()
			j, p, campaign, db := a150ExitFixture(t)
			if _, err := db.Exec(`UPDATE positions SET instance_seq=2 WHERE id=?`, p.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE position_campaigns SET actual_position_generation=2 WHERE id=?`, campaign); err != nil {
				t.Fatal(err)
			}
			if existing {
				if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: "100", InitialStop: "90"}); err != nil {
					t.Fatal(err)
				}
			}
			o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef}, clk: clock.NewFake(j.Now())}
			cycle := ExitCycle{}
			states, err := o.workingSet(ctx, &cycle)
			if err != nil || cycle.Err == nil || len(states) != map[bool]int{false: 0, true: 1}[existing] {
				t.Fatalf("states=%+v cycle=%+v err=%v", states, cycle, err)
			}
			if existing && (states[0].identityErr != nil || states[0].state.InitialStop != "90") {
				t.Fatalf("existing protection changed: %+v", states[0])
			}
		})
	}
}

func TestA150ExistingExitRemainsManagedWhenLatchFails(t *testing.T) {
	ctx := context.Background()
	j, p, campaign, db := a150ExitFixture(t)
	decision, err := j.LookupDecision(ctx, p.EntryDecisionID)
	if err != nil {
		t.Fatal(err)
	}
	value, err := journal.ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		t.Fatal(err)
	}
	risk := value.(journal.RiskIntent)
	if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: risk.EntryPrice, InitialStop: risk.StopPrice}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER a150_fail_latch BEFORE UPDATE ON position_campaigns BEGIN SELECT RAISE(ABORT,'injected latch failure'); END`); err != nil {
		t.Fatal(err)
	}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef}, clk: clock.NewFake(j.Now())}
	cycle := ExitCycle{}
	states, err := o.workingSet(ctx, &cycle)
	if err != nil || len(states) != 1 || states[0].identityErr != nil || cycle.Err == nil {
		t.Fatalf("existing exit lost: states=%+v cycle=%+v err=%v campaign=%s", states, cycle, err, campaign)
	}
	o.opts.Costs = costs.DefaultModel()
	o.ladder = exitpolicy.DefaultLadderPolicy()
	o.refused = map[string]bool{}
	now := o.clk.Now()
	quote := observedQuote{Price: "96", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(o.clk)}
	if err := o.judge(ctx, states[0], quote, cycleObservation{at: now.UTC(), sequence: 1}, &cycle); err != nil {
		t.Fatal(err)
	}
	state, err := j.ExitState(ctx, p.ID)
	if err != nil || state.SnapshotStatus != journal.SnapshotStatusEvaluated || state.InitialStop != risk.StopPrice {
		t.Fatalf("protective evaluation missing: %+v err=%v", state, err)
	}
}

func TestA150ExistingExitRemainsManagedWhenAdmissionEvidenceUnreadable(t *testing.T) {
	ctx := context.Background()
	j, p, _, db := a150ExitFixture(t)
	decision, err := j.LookupDecision(ctx, p.EntryDecisionID)
	if err != nil {
		t.Fatal(err)
	}
	value, err := journal.ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		t.Fatal(err)
	}
	risk := value.(journal.RiskIntent)
	if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyRatchet, EntryPrice: risk.EntryPrice, InitialStop: risk.StopPrice}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP TRIGGER strategy_decision_lineage_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE strategy_decision_lineage SET decision_payload='{}'`); err != nil {
		t.Fatal(err)
	}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef}, clk: clock.NewFake(j.Now())}
	cycle := ExitCycle{}
	states, err := o.workingSet(ctx, &cycle)
	if err != nil || len(states) != 1 || states[0].identityErr != nil || cycle.Err == nil {
		t.Fatalf("existing exit lost on bad admission evidence: states=%+v cycle=%+v err=%v", states, cycle, err)
	}
}

func TestA150StopBreachUsesSellableFloorWhenLatchFails(t *testing.T) {
	ctx := context.Background()
	j, p, _, db := a150ExitFixture(t)
	if _, err := db.Exec(`UPDATE positions SET quantity='3' WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	decision, err := j.LookupDecision(ctx, p.EntryDecisionID)
	if err != nil {
		t.Fatal(err)
	}
	value, err := journal.ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		t.Fatal(err)
	}
	entry := value.(journal.RiskIntent)
	if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: entry.EntryPrice, InitialStop: entry.StopPrice}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER a150_fail_latch BEFORE UPDATE ON position_campaigns BEGIN SELECT RAISE(ABORT,'injected latch failure'); END`); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(j.Now())
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: clk, AccountRef: p.AccountRef, Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a150-test"})
	if err != nil {
		t.Fatal(err)
	}
	submit := &a150Submitter{}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, Costs: costs.DefaultModel(), Issuer: guardian, Submit: submit, Floor: a150Floor{}, NewID: func() string { return "a150-stop-intent" }}, clk: clk, ladder: exitpolicy.DefaultLadderPolicy(), refused: map[string]bool{}}
	cycle := ExitCycle{}
	states, err := o.workingSet(ctx, &cycle)
	if err != nil || cycle.Err == nil || len(states) != 1 {
		t.Fatalf("states=%+v cycle=%+v err=%v", states, cycle, err)
	}
	now := clk.Now()
	quote := observedQuote{Price: "1", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}
	if err := o.judge(ctx, states[0], quote, cycleObservation{at: now.UTC(), sequence: 1}, &cycle); err != nil {
		t.Fatal(err)
	}
	if len(submit.requests) != 1 || submit.requests[0].Intent.Quantity != 2 {
		t.Fatalf("sellable floor not respected: %+v", submit.requests)
	}
}

func TestA150MismatchedStopCannotWeakenAdmittedProtection(t *testing.T) {
	for _, tc := range []struct {
		oldStop   string
		failLatch bool
	}{{"90", false}, {"90", true}, {"97", false}} {
		t.Run(fmt.Sprintf("stop-%s-latch-failure-%t", tc.oldStop, tc.failLatch), func(t *testing.T) {
			ctx := context.Background()
			j, p, _, db := a150ExitFixture(t)
			if _, err := db.Exec(`UPDATE positions SET quantity='3' WHERE id=?`, p.ID); err != nil {
				t.Fatal(err)
			}
			decision, err := j.LookupDecision(ctx, p.EntryDecisionID)
			if err != nil {
				t.Fatal(err)
			}
			value, err := journal.ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
			if err != nil {
				t.Fatal(err)
			}
			entry := value.(journal.RiskIntent)
			if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: entry.EntryPrice, InitialStop: tc.oldStop}); err != nil {
				t.Fatal(err)
			}
			if tc.failLatch {
				if _, err := db.Exec(`CREATE TRIGGER a150_fail_latch BEFORE UPDATE ON position_campaigns BEGIN SELECT RAISE(ABORT,'injected latch failure'); END`); err != nil {
					t.Fatal(err)
				}
			}
			clk := clock.NewFake(j.Now())
			guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: clk, AccountRef: p.AccountRef, Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a150-test"})
			if err != nil {
				t.Fatal(err)
			}
			submit := &a150Submitter{}
			o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, Costs: costs.DefaultModel(), Issuer: guardian, Submit: submit, Floor: a150Floor{}, NewID: func() string { return "a150-mismatch-stop" }}, clk: clk, ladder: exitpolicy.DefaultLadderPolicy(), refused: map[string]bool{}}
			cycle := ExitCycle{}
			states, err := o.workingSet(ctx, &cycle)
			if err != nil || cycle.Err == nil || len(states) != 1 {
				t.Fatalf("states=%+v cycle=%+v err=%v", states, cycle, err)
			}
			now := clk.Now()
			quote := observedQuote{Price: "94", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}
			if err := o.judge(ctx, states[0], quote, cycleObservation{at: now.UTC(), sequence: 1}, &cycle); err != nil {
				t.Fatal(err)
			}
			if tc.oldStop == "90" {
				if len(submit.requests) != 0 || submit.unsettledCalls != 0 {
					t.Fatal("floor reconciliation submitted before commit")
				}
				events, err := j.ExitEvents(ctx, p.ID)
				if err != nil || events[len(events)-1].ArmSuppressedReason != journal.ArmSuppressedFloorReconciliation {
					t.Fatalf("missing truthful suppression: %+v err=%v", events, err)
				}
				states, err = o.workingSet(ctx, &cycle)
				if err != nil || len(states) != 1 {
					t.Fatalf("retry states=%+v err=%v", states, err)
				}
				clk.Advance(time.Second)
				now = clk.Now()
				quote = observedQuote{Price: "94", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}
				if err := o.judge(ctx, states[0], quote, cycleObservation{at: now.UTC(), sequence: 2}, &cycle); err != nil {
					t.Fatal(err)
				}
			}
			if len(submit.requests) != 1 {
				t.Fatalf("stop %s, admitted %s, observed 94, orders=%d", tc.oldStop, entry.StopPrice, len(submit.requests))
			}
			if submit.requests[0].Intent.Quantity != 2 {
				t.Fatalf("quantity=%v, want sellable floor 2", submit.requests[0].Intent.Quantity)
			}
		})
	}
}

func TestA150RaisedFloorPersistsAcrossRestartAndPendingSell(t *testing.T) {
	for _, price := range []string{"96", "94"} {
		t.Run(price, func(t *testing.T) { testA150FloorRestart(t, price) })
	}
}

func testA150FloorRestart(t *testing.T, price string) {
	t.Helper()
	ctx := context.Background()
	j, p, _, _ := a150ExitFixture(t)
	decision, err := j.LookupDecision(ctx, p.EntryDecisionID)
	if err != nil {
		t.Fatal(err)
	}
	value, err := journal.ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		t.Fatal(err)
	}
	entry := value.(journal.RiskIntent)
	if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: entry.EntryPrice, InitialStop: "90"}); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(j.Now())
	makeObserver := func(j *journal.Journal, submit *a150Submitter) *ExitObserver {
		guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: clk, AccountRef: p.AccountRef, Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a150-test"})
		if err != nil {
			t.Fatal(err)
		}
		return &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, Costs: costs.DefaultModel(), Issuer: guardian, Submit: submit, NewID: func() string { return "a150-pending-stop" }}, clk: clk, ladder: exitpolicy.DefaultLadderPolicy(), refused: map[string]bool{}}
	}
	submit := &a150Submitter{}
	o := makeObserver(j, submit)
	judge := func(o *ExitObserver, price string, seq uint64) {
		cycle := ExitCycle{}
		states, err := o.workingSet(ctx, &cycle)
		if err != nil || len(states) != 1 {
			t.Fatalf("states=%+v cycle=%+v err=%v", states, cycle, err)
		}
		now := clk.Now()
		if err := o.judge(ctx, states[0], observedQuote{Price: price, FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}, cycleObservation{at: now.UTC(), sequence: seq}, &cycle); err != nil {
			t.Fatal(err)
		}
	}
	judge(o, price, 1)
	state, err := j.ExitState(ctx, p.ID)
	if err != nil || state.Baseline != entry.StopPrice || state.InitialStop != "90" || state.SnapshotStatus != journal.SnapshotStatusEvaluated || len(submit.requests) != 0 {
		t.Fatalf("non-breach state=%+v requests=%d err=%v", state, len(submit.requests), err)
	}
	path := j.Path()
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := journal.Open(ctx, journal.Options{Path: path, Clock: clk, FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	clk.Advance(time.Second)
	o = makeObserver(reopened, submit)
	judge(o, "94", 2)
	state, err = reopened.ExitState(ctx, p.ID)
	if err != nil || state.Baseline != entry.StopPrice || state.InitialStop != "90" || !state.Pending() || len(submit.requests) != 1 {
		t.Fatalf("breach state=%+v requests=%d err=%v", state, len(submit.requests), err)
	}
	clk.Advance(time.Second)
	judge(o, "94", 3)
	if len(submit.requests) != 1 {
		t.Fatalf("duplicate pending sale: %d", len(submit.requests))
	}
}

func TestA150StaleFloorCandidateDoesNotQuarantineExistingExit(t *testing.T) {
	ctx := context.Background()
	j, p, _, _ := a150ExitFixture(t)
	if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: "100", InitialStop: "90"}); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(j.Now())
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: clk, AccountRef: p.AccountRef, Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a150-test"})
	if err != nil {
		t.Fatal(err)
	}
	submit := &a150Submitter{}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, Costs: costs.DefaultModel(), Issuer: guardian, Submit: submit, NewID: func() string { return "a150-stale" }}, clk: clk, ladder: exitpolicy.DefaultLadderPolicy(), refused: map[string]bool{}}
	cycle := ExitCycle{}
	states, err := o.workingSet(ctx, &cycle)
	if err != nil || len(states) != 1 || states[0].state.Baseline != "95" {
		t.Fatalf("states=%+v err=%v", states, err)
	}
	stale, newer := states[0], states[0]
	newer.state.Baseline = "90" // A concurrent observer still holding the pre-floor state.
	newer.floorExpectedState = nil
	now := clk.Now()
	if err := o.judge(ctx, newer, observedQuote{Price: "101", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}, cycleObservation{at: now.UTC(), sequence: 1}, &cycle); err != nil {
		t.Fatal(err)
	}
	clk.Advance(time.Second)
	now = clk.Now()
	err = o.judge(ctx, stale, observedQuote{Price: "94", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}, cycleObservation{at: now.UTC(), sequence: 2}, &cycle)
	if !errors.Is(err, journal.ErrExitObservationConflict) {
		t.Fatalf("stale candidate: %v", err)
	}
	if submit.unsettledCalls != 0 || len(submit.requests) != 0 {
		t.Fatal("stale candidate touched order transport")
	}
	if _, active, err := j.ActiveExitSnapshotQuarantine(ctx, p.ID, 1); err != nil || active {
		t.Fatalf("valid prior exit quarantined: active=%v err=%v", active, err)
	}
	states, err = o.workingSet(ctx, &cycle)
	if err != nil || len(states) != 1 {
		t.Fatalf("retry states=%+v err=%v", states, err)
	}
	clk.Advance(time.Second)
	now = clk.Now()
	if err := o.judge(ctx, states[0], observedQuote{Price: "94", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}, cycleObservation{at: now.UTC(), sequence: 3}, &cycle); err != nil {
		t.Fatal(err)
	}
	if len(submit.requests) != 0 {
		t.Fatal("floor retry submitted before floor commit")
	}
	states, err = o.workingSet(ctx, &cycle)
	if err != nil || len(states) != 1 {
		t.Fatalf("post-floor states=%+v err=%v", states, err)
	}
	clk.Advance(time.Second)
	now = clk.Now()
	if err := o.judge(ctx, states[0], observedQuote{Price: "94", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}, cycleObservation{at: now.UTC(), sequence: 4}, &cycle); err != nil {
		t.Fatal(err)
	}
	if len(submit.requests) != 1 {
		t.Fatalf("fresh protection was not submitted: %d", len(submit.requests))
	}
}

func TestA150FloorReconciliationPreservesExistingPendingSell(t *testing.T) {
	ctx := context.Background()
	j, p, _, _ := a150ExitFixture(t)
	if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: "100", InitialStop: "90"}); err != nil {
		t.Fatal(err)
	}
	if err := j.RecordExitJudgement(ctx, journal.ExitJudgement{PositionID: p.ID, ObservedPrice: "100", HighWater: "100", Baseline: "90", RatchetLevel: journal.RatchetNone, ActiveRung: exitpolicy.NoRung, Proposal: &journal.ExitProposal{Action: string(exitpolicy.ActionLadderPartial), Level: "0"}}); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(j.Now())
	submit := &a150Submitter{}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, Submit: submit, NewID: func() string { return "a150-existing-pending" }}, clk: clk, ladder: exitpolicy.DefaultLadderPolicy(), refused: map[string]bool{}}
	cycle := ExitCycle{}
	states, err := o.workingSet(ctx, &cycle)
	if err != nil || len(states) != 1 {
		t.Fatalf("states=%+v err=%v", states, err)
	}
	now := clk.Now()
	if err := o.judge(ctx, states[0], observedQuote{Price: "94", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}, cycleObservation{at: now.UTC(), sequence: 1}, &cycle); err != nil {
		t.Fatal(err)
	}
	state, err := j.ExitState(ctx, p.ID)
	if err != nil || !state.Pending() || state.PendingAction != string(exitpolicy.ActionLadderPartial) || state.Baseline != "95" || state.InitialStop != "90" || len(submit.requests) != 0 || submit.unsettledCalls != 0 {
		t.Fatalf("state=%+v requests=%d unsettled=%d err=%v", state, len(submit.requests), submit.unsettledCalls, err)
	}
}

func TestA150FloorSeedRejectsConcurrentLegacyProgress(t *testing.T) {
	for _, change := range []string{"watermark", "fill", "pending"} {
		t.Run(change, func(t *testing.T) { testA150FloorSeedConflict(t, change) })
	}
}

func testA150FloorSeedConflict(t *testing.T, change string) {
	t.Helper()
	ctx := context.Background()
	j, p, _, db := a150ExitFixture(t)
	if _, err := j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: exitpolicy.CommonLadderHybrid50, EntryPrice: "100", InitialStop: "90"}); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(j.Now())
	submit := &a150Submitter{}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, Submit: submit, NewID: func() string { return "a150-stale-seed" }}, clk: clk, ladder: exitpolicy.DefaultLadderPolicy(), refused: map[string]bool{}}
	cycle := ExitCycle{}
	states, err := o.workingSet(ctx, &cycle)
	if err != nil || len(states) != 1 {
		t.Fatalf("states=%+v err=%v", states, err)
	}
	switch change {
	case "watermark":
		if err := j.RecordExitJudgement(ctx, journal.ExitJudgement{PositionID: p.ID, ObservedPrice: "101", HighWater: "101", Baseline: "92", RatchetLevel: journal.RatchetNone, ActiveRung: exitpolicy.NoRung}); err != nil {
			t.Fatal(err)
		}
	case "fill":
		if _, err := db.Exec(`UPDATE positions SET quantity='0.5' WHERE id=?`, p.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE exit_states SET taken_ratio_total='0.5' WHERE position_id=?`, p.ID); err != nil {
			t.Fatal(err)
		}
	case "pending":
		if _, err := db.Exec(`UPDATE exit_states SET pending_action=?,pending_level='0' WHERE position_id=?`, string(exitpolicy.ActionLadderPartial), p.ID); err != nil {
			t.Fatal(err)
		}
	}
	now := clk.Now()
	err = o.judge(ctx, states[0], observedQuote{Price: "94", FetchedAt: now.UTC(), Source: "a150-test", leaseAnchor: clock.LeaseAnchor(clk)}, cycleObservation{at: now.UTC(), sequence: 1}, &cycle)
	if !errors.Is(err, journal.ErrExitObservationConflict) {
		t.Fatalf("stale legacy seed: %v", err)
	}
	state, err := j.ExitState(ctx, p.ID)
	if err != nil || (change == "watermark" && (state.HighWater != "101" || state.Baseline != "92")) || submit.unsettledCalls != 0 || len(submit.requests) != 0 {
		t.Fatalf("state=%+v err=%v", state, err)
	}
}

func TestA150UnrecoverablePositionIsCountedAndWarned(t *testing.T) {
	ctx := context.Background()
	j, p, campaign, db := a150ExitFixture(t)
	if _, err := db.Exec(`UPDATE positions SET instance_seq=2 WHERE id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE position_campaigns SET actual_position_generation=2 WHERE id=?`, campaign); err != nil {
		t.Fatal(err)
	}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, NewID: func() string { return "a150-unobserved" }}, clk: clock.NewFake(j.Now())}
	cycle := o.ObserveOnce(ctx)
	if cycle.Err == nil || cycle.Unobserved != 1 || cycle.Judged != 0 || len(o.protectionAlerts) != 1 {
		t.Fatalf("cycle=%+v alerts=%+v", cycle, o.protectionAlerts)
	}
}
