//go:build tossos_testseams

package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// a159Dispatch uses the admitted lease with a fake broker; fills still flow
// through the real journal hooks and preserve the admission's lineage.
type a159Dispatch struct {
	*strategyDispatchGatewaySpy
	j     *journal.Journal
	lease journal.StrategyDispatchLease
	fill  journal.FillObservation
}

func (g *a159Dispatch) PlaceClaimedStrategy(ctx context.Context, req execgw.StrategyPlaceRequest) (execgw.Outcome, error) {
	lease, err := g.j.LookupStrategyDispatchLease(ctx, req.Lease.LeaseID)
	if err != nil {
		return execgw.Outcome{}, err
	}
	g.lease = lease
	qty := fmt.Sprintf("%g", req.Intent.Quantity)
	price := fmt.Sprintf("%g", req.Intent.Price)
	at := g.j.Now().UTC().Format("2006-01-02T15:04:05Z")
	g.fill = journal.FillObservation{OrderID: "a159-buy", AccountRef: lease.AccountRef, Market: strings.ToLower(req.Intent.Market), TradingDay: "2026-08-04", Symbol: req.Intent.Symbol, Side: "BUY", State: "OPEN_PARTIALLY_FILLED", Quantity: qty, FilledQuantity: "3", AveragePrice: price, FilledAmount: "300", BrokerVisibleAt: at, ObservedAt: at}
	attempt, err := g.j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{ID: req.IntentID, Market: g.fill.Market, TradingDay: g.fill.TradingDay, AccountRef: lease.AccountRef, Symbol: req.Intent.Symbol, Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: qty, Price: price, Currency: "KRW", Source: "strategy/a159", Fingerprint: "a159-buy"},
		Kind:   journal.KindPlace, AttemptID: "a159-buy-attempt", AccountRef: lease.AccountRef, DecisionID: req.Decision.ID, SafetyClass: journal.SafetyClassExposureRaising, ClientOrderID: lease.OperationID,
	})
	if err != nil {
		return execgw.Outcome{}, err
	}
	result, err := attempt.DispatchStrategyVerified(ctx, req.Lease, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
		return journal.DispatchOutcome{Class: journal.DispatchAcked, BrokerOrderID: g.fill.OrderID}
	}, func(ctx context.Context, _ string) error {
		_, err := g.j.RecordFill(ctx, g.fill)
		return err
	})
	return execgw.Outcome{IntentID: req.IntentID, State: result.Final, BrokerOrderID: result.BrokerOrderID}, err
}

func TestA159ActualPartialFillFixture(t *testing.T) {
	ctx := context.Background()
	cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill, Exit: journal.ApplyExitFill}); err != nil {
		t.Fatal(err)
	}
	gateway := &a159Dispatch{strategyDispatchGatewaySpy: spy, j: j}
	cycle.gateway = gateway
	proposal := proposals.forMarket(StrategyMarketKR).entries[0].authority.Proposal()
	out, err := cycle.dispatch(ctx, deliverForTest(t, proposal))
	if err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("dispatch=%+v err=%v", out, err)
	}
	p, err := j.CurrentPosition(ctx, gateway.lease.AccountRef, "kr", gateway.lease.Symbol)
	if err != nil || p.Quantity != "3" || p.EntryDecisionID != gateway.lease.GuardianDecisionID {
		t.Fatalf("position=%+v err=%v", p, err)
	}
	wm, err := j.CampaignOrderWatermark(ctx, gateway.lease.CampaignID, 1, gateway.fill.OrderID)
	if err != nil || wm.CumulativeFilled != "3" || wm.RemainingQuantity != "5" {
		t.Fatalf("watermark=%+v err=%v", wm, err)
	}
}

func TestA159PartialFillLatchRestartAndNextLeg(t *testing.T) {
	ctx := context.Background()
	for _, blocked := range []bool{false, true} {
		t.Run(fmt.Sprintf("blocked-%v", blocked), func(t *testing.T) {
			cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
			if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill, Exit: journal.ApplyExitFill}); err != nil {
				t.Fatal(err)
			}
			g := &a159Dispatch{strategyDispatchGatewaySpy: spy, j: j}
			cycle.gateway = g
			proposal := proposals.forMarket(StrategyMarketKR).entries[0].authority.Proposal()
			if _, err := cycle.dispatch(ctx, deliverForTest(t, proposal)); err != nil {
				t.Fatal(err)
			}
			p, err := j.CurrentPosition(ctx, g.lease.AccountRef, "kr", g.lease.Symbol)
			if err != nil {
				t.Fatal(err)
			}
			o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef}, clk: clock.NewFake(j.Now())}
			c := ExitCycle{}
			states, err := o.workingSet(ctx, &c)
			if err != nil || c.Err != nil || len(states) != 1 {
				t.Fatalf("open protection states=%+v cycle=%+v err=%v", states, c, err)
			}
			before := states[0].state
			if blocked {
				if err := j.ReconcileStrategyProtection(ctx, g.lease.CampaignID, p); err != nil {
					t.Fatal(err)
				}
			}
			camp, err := j.PositionCampaign(ctx, g.lease.CampaignID)
			if err != nil {
				t.Fatal(err)
			}
			path, now := j.Path(), j.Now()
			if err := j.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := journal.Open(ctx, journal.Options{Path: path, Clock: clock.NewFake(now), FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = reopened.Close() })
			if err := reopened.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill, Exit: journal.ApplyExitFill}); err != nil {
				t.Fatal(err)
			}
			replay, err := reopened.RecordFill(ctx, g.fill)
			if err != nil || replay.Changed {
				t.Fatalf("fill replay=%+v err=%v", replay, err)
			}
			if blocked {
				if err := reopened.ReconcileStrategyProtection(ctx, g.lease.CampaignID, p); err != nil {
					t.Fatal(err)
				}
			}
			wm, err := reopened.CampaignOrderWatermark(ctx, g.lease.CampaignID, 1, g.fill.OrderID)
			if err != nil || wm.CumulativeFilled != "3" || wm.RemainingQuantity != "5" {
				t.Fatalf("reopened watermark=%+v err=%v", wm, err)
			}
			after, err := reopened.PositionCampaign(ctx, g.lease.CampaignID)
			if err != nil || after.Version != camp.Version || after.EntryBlocked != blocked || after.ActualPositionGeneration != p.InstanceSeq {
				t.Fatalf("campaign=%+v before=%+v err=%v", after, camp, err)
			}
			pos, err := reopened.CurrentPosition(ctx, p.AccountRef, "kr", p.Symbol)
			if err != nil || pos.Quantity != "3" || pos.InstanceSeq != p.InstanceSeq {
				t.Fatalf("position=%+v err=%v", pos, err)
			}
			exit, err := reopened.ExitState(ctx, p.ID)
			if err != nil || exit.InitialStop != before.InitialStop || exit.PolicyIdentity != before.PolicyIdentity {
				t.Fatalf("exit=%+v before=%+v err=%v", exit, before, err)
			}
			db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			var events int
			if err := db.QueryRow(`SELECT count(*) FROM campaign_events WHERE campaign_id=? AND event_kind='PROTECTION_RECONCILE'`, g.lease.CampaignID).Scan(&events); err != nil || events != map[bool]int{true: 1, false: 0}[blocked] {
				t.Fatalf("events=%d err=%v", events, err)
			}
			req := journal.PlanCampaignLegRequest{CampaignID: g.lease.CampaignID, ExpectedVersion: after.Version, CommandKey: "a159-next", Sequence: 2, PlanID: "a159-next-plan", RequestedQuantity: "1"}
			leg, err := reopened.PlanCampaignLeg(ctx, req)
			if blocked {
				if !errors.Is(err, positioncampaign.ErrExposureBlocked) {
					t.Fatalf("expected exposure block: leg=%+v err=%v", leg, err)
				}
				for _, q := range []string{`SELECT count(*) FROM campaign_legs WHERE campaign_id=? AND sequence=2`, `SELECT count(*) FROM campaign_commands WHERE campaign_id=? AND command_key='a159-next'`} {
					var n int
					if err := db.QueryRow(q, g.lease.CampaignID).Scan(&n); err != nil || n != 0 {
						t.Fatalf("blocked command created rows=%d err=%v", n, err)
					}
				}
			} else if err != nil || leg.Sequence != 2 {
				t.Fatalf("positive control leg=%+v err=%v", leg, err)
			}
		})
	}
}

type a159ExitSubmitter struct {
	gateway  *execgw.Gateway
	j        *journal.Journal
	account  string
	requests []execgw.PlaceRequest
	cancels  int
}

func (s *a159ExitSubmitter) UnsettledOnSymbol(ctx context.Context, market, symbol string) ([]journal.AttemptRecord, error) {
	return s.gateway.UnsettledOnSymbol(ctx, market, symbol)
}
func (s *a159ExitSubmitter) Cancel(context.Context, execgw.CancelRequest) (execgw.Outcome, error) {
	s.cancels++
	return execgw.Outcome{}, errors.New("synthetic cancel unconfirmed")
}
func (s *a159ExitSubmitter) Place(ctx context.Context, r execgw.PlaceRequest) (execgw.Outcome, error) {
	s.requests = append(s.requests, r)
	attempt, err := s.j.Prepare(ctx, journal.PrepareRequest{Intent: journal.Intent{ID: r.IntentID, Market: "kr", TradingDay: "2026-08-04", AccountRef: s.account, Symbol: r.Intent.Symbol, Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: fmt.Sprintf("%g", r.Intent.Quantity), Price: fmt.Sprintf("%g", r.Intent.Price), Currency: "KRW", Source: "engine/a159", Fingerprint: "a159-sell"}, Kind: journal.KindPlace, AttemptID: "a159-sell-attempt", AccountRef: s.account, DecisionID: r.Decision.ID, SafetyClass: journal.SafetyClassRiskReducing, ClientOrderID: journal.DeriveClientOrderID(r.Decision.ID, 0)})
	if err != nil {
		return execgw.Outcome{}, err
	}
	if err = attempt.MarkDispatchStarted(ctx); err != nil {
		return execgw.Outcome{}, err
	}
	if err = attempt.MarkAcked(ctx, "a159-sell"); err != nil {
		return execgw.Outcome{}, err
	}
	if err = attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		return execgw.Outcome{}, err
	}
	return execgw.Outcome{IntentID: r.IntentID, AttemptID: "a159-sell-attempt", State: journal.StateConfirmed, BrokerOrderID: "a159-sell"}, nil
}

type a159Alerts struct{ events []obs.Event }

func (a *a159Alerts) Notify(_ context.Context, e obs.Event) error {
	a.events = append(a.events, e)
	return nil
}

func TestA159ProtectionExitAfterRestart(t *testing.T) {
	ctx := context.Background()
	cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill, Exit: journal.ApplyExitFill}); err != nil {
		t.Fatal(err)
	}
	g := &a159Dispatch{strategyDispatchGatewaySpy: spy, j: j}
	cycle.gateway = g
	proposal := proposals.forMarket(StrategyMarketKR).entries[0].authority.Proposal()
	if _, err := cycle.dispatch(ctx, deliverForTest(t, proposal)); err != nil {
		t.Fatal(err)
	}
	p, err := j.CurrentPosition(ctx, g.lease.AccountRef, "kr", g.lease.Symbol)
	if err != nil {
		t.Fatal(err)
	}
	foreign := p
	foreign.AccountRef = "different-account"
	if err := j.ReconcileStrategyProtection(ctx, g.lease.CampaignID, foreign); err == nil {
		t.Fatal("foreign account accepted")
	}
	if err := j.ReconcileStrategyProtection(ctx, g.lease.CampaignID, p); err != nil {
		t.Fatal(err)
	}
	campaign, err := j.PositionCampaign(ctx, g.lease.CampaignID)
	if err != nil {
		t.Fatal(err)
	}
	path, now := j.Path(), j.Now()
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(now)
	j, err = journal.Open(ctx, journal.Options{Path: path, Clock: clk, FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill, Exit: journal.ApplyExitFill}); err != nil {
		t.Fatal(err)
	}
	if replay, err := j.RecordFill(ctx, g.fill); err != nil || replay.Changed {
		t.Fatalf("duplicate fill=%+v err=%v", replay, err)
	}
	if err := j.ReconcileStrategyProtection(ctx, g.lease.CampaignID, p); err != nil {
		t.Fatal(err)
	}
	current, err := j.PositionCampaign(ctx, g.lease.CampaignID)
	if err != nil || !current.EntryBlocked || current.Version != campaign.Version {
		t.Fatalf("reopened campaign=%+v err=%v", current, err)
	}
	_, err = j.PlanCampaignLeg(ctx, journal.PlanCampaignLegRequest{CampaignID: g.lease.CampaignID, ExpectedVersion: current.Version, CommandKey: "a159-exit-next", Sequence: 2, PlanID: "a159-exit-next-plan", RequestedQuantity: "1"})
	if !errors.Is(err, positioncampaign.ErrExposureBlocked) {
		t.Fatalf("next leg not blocked: %v", err)
	}
	realGateway, err := execgw.New(execgw.Options{Journal: j, Trading: trading.NewService(config.Trading{}, nil), Clock: clk, AccountRef: p.AccountRef})
	if err != nil {
		t.Fatal(err)
	}
	submit := &a159ExitSubmitter{gateway: realGateway, j: j, account: p.AccountRef}
	alerts := &a159Alerts{}
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: clk, AccountRef: p.AccountRef, Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a159-test"})
	if err != nil {
		t.Fatal(err)
	}
	o := &ExitObserver{opts: ExitObserverOptions{Journal: j, AccountRef: p.AccountRef, Costs: costs.DefaultModel(), Issuer: guardian, Submit: submit, Floor: a150Floor{}, Alerts: alerts, NewID: func() string { return "a159-exit-intent" }}, clk: clk, ratchet: exitpolicy.DefaultRatchetConfig(), ladder: exitpolicy.DefaultLadderPolicy(), refused: map[string]bool{}, delayedSince: map[string]time.Time{}, delayAlerted: map[string]bool{}}
	observe := func(price string, seq uint64) {
		t.Helper()
		c := ExitCycle{}
		m, err := o.workingSet(ctx, &c)
		if err != nil || len(m) != 1 {
			t.Fatalf("working=%+v cycle=%+v err=%v", m, c, err)
		}
		at := clk.Now()
		q := observedQuote{Price: price, FetchedAt: at.UTC(), Source: "a159-synthetic", leaseAnchor: clock.LeaseAnchor(clk)}
		if err := o.judge(ctx, m[0], q, cycleObservation{at: at.UTC(), sequence: seq}, &c); err != nil {
			t.Fatalf("judge: %v", err)
		}
	}
	live, err := j.LiveOrdersForSymbol(ctx, p.AccountRef, "kr", p.Symbol)
	if err != nil || len(live) != 1 || live[0].OrderID != g.fill.OrderID {
		t.Fatalf("working BUY=%+v err=%v", live, err)
	}
	observe("100", 1)
	if len(submit.requests) != 0 {
		t.Fatal("non-breach submitted")
	}
	observe("94", 2)
	if len(submit.requests) != 0 || submit.cancels == 0 {
		t.Fatalf("working BUY was not withheld: %+v", submit)
	}
	clk.Advance(31 * time.Second)
	observe("94", 3)
	found := false
	for _, e := range alerts.events {
		if e.Type == obs.EventExitLiquidationDelayed {
			found = true
		}
	}
	if !found || len(submit.requests) != 0 {
		t.Fatalf("delay alert=%v requests=%v", found, submit.requests)
	}
	state, err := j.ExitState(ctx, p.ID)
	if err != nil || state.SnapshotStatus != journal.SnapshotStatusEvaluated {
		t.Fatalf("protection state=%+v err=%v", state, err)
	}
	terminal := g.fill
	terminal.State = "CLOSED_CANCELLED"
	terminal.Terminal = true
	terminal.BrokerVisibleAt = clk.Now().UTC().Format(time.RFC3339)
	terminal.ObservedAt = terminal.BrokerVisibleAt
	if _, err := j.RecordFill(ctx, terminal); err != nil {
		t.Fatal(err)
	}
	clk.Advance(time.Second)
	observe("94", 4)
	if len(submit.requests) != 1 || submit.requests[0].Intent.Quantity != 2 {
		state, _ := j.ExitState(ctx, p.ID)
		live, _ := j.LiveOrdersForSymbol(ctx, p.AccountRef, "kr", p.Symbol)
		t.Fatalf("expected capped sell=2: requests=%+v state=%+v live=%+v alerts=%+v", submit.requests, state, live, alerts.events)
	}
	attempt, err := j.LookupAttempt(ctx, "a159-sell-attempt")
	if err != nil || attempt.DecisionID != submit.requests[0].Decision.ID || attempt.SafetyClass != journal.SafetyClassRiskReducing || !strings.EqualFold(submit.requests[0].Intent.Side, "SELL") {
		t.Fatalf("sell authority=%+v err=%v", attempt, err)
	}
	pending, err := j.ExitState(ctx, p.ID)
	if err != nil || !pending.Pending() {
		t.Fatalf("pending exit=%+v err=%v", pending, err)
	}
	live, err = j.LiveOrdersForSymbol(ctx, p.AccountRef, "kr", p.Symbol)
	if err != nil || len(live) != 1 || live[0].OrderID != "a159-sell" {
		t.Fatalf("pending SELL=%+v err=%v", live, err)
	}
	observe("94", 5)
	if len(submit.requests) != 1 {
		t.Fatalf("duplicate sell: %+v", submit.requests)
	}
	live, err = j.LiveOrdersForSymbol(ctx, p.AccountRef, "kr", p.Symbol)
	if err != nil || len(live) != 1 || live[0].OrderID != "a159-sell" {
		t.Fatalf("pending SELL changed=%+v err=%v", live, err)
	}
}

func TestA159OtherCampaignRemainsIndependent(t *testing.T) {
	ctx := context.Background()
	f := newA112TradingFixture(t, a112TradingOptions{})
	if err := f.journal.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Campaign: journal.ApplyPositionCampaignFill, Exit: journal.ApplyExitFill}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.cycle.dispatch(ctx, deliverForTest(t, f.second.Proposal())); err != nil {
		t.Fatalf("other scope first leg: %v", err)
	}
	if len(f.spy.calls) != 1 {
		t.Fatalf("other requests=%d", len(f.spy.calls))
	}
	other, err := f.journal.LookupStrategyDispatchLease(ctx, f.spy.calls[0].Lease.LeaseID)
	if err != nil {
		t.Fatal(err)
	}
	f.wave(t)
	g := &a159Dispatch{strategyDispatchGatewaySpy: f.spy, j: f.journal}
	f.cycle.gateway = g
	if _, err := f.cycle.dispatch(ctx, deliverForTest(t, f.winner.Proposal())); err != nil {
		t.Fatalf("filled scope first leg: %v", err)
	}
	p, err := f.journal.CurrentPosition(ctx, g.lease.AccountRef, "kr", g.lease.Symbol)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.journal.ReconcileStrategyProtection(ctx, g.lease.CampaignID, p); err != nil {
		t.Fatal(err)
	}
	campaign, err := f.journal.PositionCampaign(ctx, other.CampaignID)
	if err != nil || campaign.EntryBlocked || campaign.Symbol == p.Symbol {
		t.Fatalf("other campaign=%+v err=%v", campaign, err)
	}
	leg, err := f.journal.PlanCampaignLeg(ctx, journal.PlanCampaignLegRequest{CampaignID: other.CampaignID, ExpectedVersion: campaign.Version, CommandKey: "a159-other-next", Sequence: 2, PlanID: "a159-other-plan", RequestedQuantity: "1"})
	if err != nil || leg.Sequence != 2 {
		t.Fatalf("other scope next leg=%+v err=%v", leg, err)
	}
}
