//go:build tossos_testseams

package engine

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
)

type strategyDispatchGatewaySpy struct {
	mu             sync.Mutex
	calls          []execgw.StrategyPlaceRequest
	observed       map[string]int
	failProtection map[string]error
	failEntryGate  map[string]error
}

func (spy *strategyDispatchGatewaySpy) ObserveStrategyProtection(_ context.Context, market string, _ uint64) (execgw.StrategyProtectionAuthority, error) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.observed["protection-"+market]++
	if err := spy.failProtection[market]; err != nil {
		return execgw.StrategyProtectionAuthority{}, err
	}
	return execgw.StrategyProtectionAuthorityForTest(strings.ToUpper(market), 9, strings.Repeat("a", 64)), nil
}

func (spy *strategyDispatchGatewaySpy) ObserveStrategyEntryGate(_ context.Context, market, _ string) (execgw.StrategyEntryGateAuthority, error) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.observed["gate-"+market]++
	if err := spy.failEntryGate[market]; err != nil {
		return execgw.StrategyEntryGateAuthority{}, err
	}
	return execgw.StrategyEntryGateAuthorityForTest(3, "sha256:"+strings.Repeat("b", 64)), nil
}

func (spy *strategyDispatchGatewaySpy) PlaceClaimedStrategy(_ context.Context, request execgw.StrategyPlaceRequest) (execgw.Outcome, error) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.calls = append(spy.calls, request)
	return execgw.Outcome{IntentID: request.IntentID, State: journal.StateConfirmed, BrokerOrderID: "fake-" + request.Intent.Market}, nil
}

func TestStrategyDispatchCyclePairsKRUSThroughDerivedLeaseAndGateway(t *testing.T) {
	seen := map[string]bool{}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
		result := proposals.forMarket(market).entries[0].authority.Proposal()
		out, err := cycle.dispatch(context.Background(), result)
		if err != nil || out.State != journal.StateConfirmed {
			t.Fatalf("%s outcome=%+v err=%v", market, out, err)
		}
		if len(spy.calls) != 1 {
			t.Fatalf("%s Gateway calls=%+v", market, spy.calls)
		}
		request := spy.calls[0]
		lease, err := j.LookupStrategyDispatchLease(context.Background(), request.Lease.LeaseID)
		if err != nil || lease.State != journal.StrategyDispatchLeaseClaimed || lease.Revision != 2 ||
			request.Lease.ExpectedRevision != 2 || request.Intent.Price != 100 || request.Intent.Quantity <= 0 {
			t.Fatalf("%s lease=%+v request=%+v err=%v", request.Intent.Market, lease, request, err)
		}
		for _, prefix := range []string{"protection-", "gate-"} {
			key := prefix + request.Intent.Market
			if spy.observed[key] != 1 {
				t.Fatalf("%s observations=%d", key, spy.observed[key])
			}
		}
		seen[request.Intent.Market] = true
	}
	if !seen["kr"] || !seen["us"] {
		t.Fatalf("unpaired dispatch markets=%v", seen)
	}
}

func TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner(t *testing.T) {
	cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
	type result struct {
		market StrategyMarket
		out    execgw.Outcome
		err    error
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	var runners sync.WaitGroup
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		runners.Add(1)
		go func() {
			defer runners.Done()
			<-start
			out, err := cycle.dispatch(context.Background(), proposals.forMarket(market).entries[0].authority.Proposal())
			results <- result{market: market, out: out, err: err}
		}()
	}
	close(start)
	runners.Wait()
	close(results)
	seen := map[StrategyMarket]bool{}
	for result := range results {
		if result.err != nil || result.out.State != journal.StateConfirmed {
			t.Fatalf("%s outcome=%+v err=%v", result.market, result.out, result.err)
		}
		seen[result.market] = true
	}
	if !seen[StrategyMarketKR] || !seen[StrategyMarketUS] {
		t.Fatalf("paired concurrent results=%v", seen)
	}
	spy.mu.Lock()
	calls := append([]execgw.StrategyPlaceRequest(nil), spy.calls...)
	spy.mu.Unlock()
	if len(calls) != 2 {
		t.Fatalf("Gateway calls=%+v", calls)
	}
	var ownerEpoch uint64
	var fencingToken string
	for _, call := range calls {
		lease, err := j.LookupStrategyDispatchLease(context.Background(), call.Lease.LeaseID)
		if err != nil {
			t.Fatal(err)
		}
		if ownerEpoch == 0 {
			ownerEpoch, fencingToken = lease.OwnerEpoch, lease.FencingToken
		}
		if lease.OwnerEpoch != ownerEpoch || lease.FencingToken != fencingToken {
			t.Fatalf("paired lease owners diverged first=%d/%s lease=%+v", ownerEpoch, fencingToken, lease)
		}
	}
}

func TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS(t *testing.T) {
	for _, failure := range []string{"protection", "entry-gate"} {
		for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
			t.Run(failure+"/"+string(market), func(t *testing.T) {
				cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
				result := proposals.forMarket(market).entries[0].authority.Proposal()
				refusal := errors.New(failure + " unavailable")
				if failure == "protection" {
					spy.failProtection = map[string]error{strings.ToLower(string(market)): refusal}
				} else {
					spy.failEntryGate = map[string]error{strings.ToLower(string(market)): refusal}
				}
				if _, err := cycle.dispatch(context.Background(), result); !errors.Is(err, refusal) {
					t.Fatalf("dispatch error=%v, want %v", err, refusal)
				}
				cas, err := j.CurrentPositionCampaignCAS(context.Background(), result.Lineage.AccountRef,
					string(result.Lineage.Market), result.Lineage.Symbol)
				if err != nil || cas.Claimed || cas.State != "FLAT" {
					t.Fatalf("post-refusal campaign CAS=%+v err=%v, want untouched FLAT", cas, err)
				}
				spy.mu.Lock()
				places := len(spy.calls)
				spy.mu.Unlock()
				if places != 0 {
					t.Fatalf("Gateway place calls=%d after pre-admission refusal", places)
				}
			})
		}
	}
}

func TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure(t *testing.T) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
	loader, ok := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	if !ok {
		t.Fatal("production first-leg authority loader unavailable")
	}
	now := loader.schedule.observedAt
	candidates := strategyCandidateAuthorityPair{observedAt: now,
		kr: readyCandidateAuthority(StrategyMarketKR), us: readyCandidateAuthority(StrategyMarketUS)}
	routes := strategyRouteAuthorityPair{observedAt: now,
		kr: readyRouteAuthority(StrategyMarketKR), us: readyRouteAuthority(StrategyMarketUS)}
	cycleFn := func(context.Context) error { return nil }
	build := func(market StrategyMarket, wiringReady bool) StrategyMarketWorker {
		return buildProductionStrategyMarketWorker(context.Background(), loader.clk, market, wiringReady, spy,
			loader.schedule, candidates, routes, loader.fx, proposals, loader.risk, loader.accounts, cycleFn)
	}

	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		worker := build(market, true)
		if !worker.Effective || worker.Cycle == nil || worker.PollInterval != DefaultStrategyCycleLimit || !worker.RefreshesAuthority {
			t.Fatalf("%s worker=%+v", market, worker)
		}
		if dormant := build(market, false); dormant.Effective || dormant.Cycle != nil || dormant.PollInterval != 0 {
			t.Fatalf("%s gate-OFF worker=%+v", market, dormant)
		}
	}

	spy.failProtection = map[string]error{"kr": errors.New("KR protection unavailable")}
	kr, us := build(StrategyMarketKR, true), build(StrategyMarketUS, true)
	if kr.Effective || !us.Effective {
		t.Fatalf("protection isolation KR=%+v US=%+v", kr, us)
	}
}

func readyCandidateAuthority(market StrategyMarket) strategyCandidateMarketAuthority {
	return strategyCandidateMarketAuthority{market: market, snapshot: StrategyCandidateMarketSnapshot{Market: market, Ready: true,
		Reason: StrategyCandidateReady, ThresholdSetDigest: "sha256:" + strings.Repeat("1", 64), EvidenceDigest: "sha256:" + strings.Repeat("2", 64)}}
}

func readyRouteAuthority(market StrategyMarket) strategyRouteMarketAuthority {
	return strategyRouteMarketAuthority{market: market, snapshot: StrategyRouteMarketSnapshot{Market: market, Ready: true,
		Reason: StrategyRouteReady, OwnerSetDigest: "sha256:" + strings.Repeat("3", 64)}}
}

func pairedStrategyDispatchCycleFixture(t *testing.T) (*strategyDispatchCycle, strategyProposalAuthorityPair, *journal.Journal, *strategyDispatchGatewaySpy) {
	t.Helper()
	riskFixture := newStrategyRiskLoaderFixture(t)
	now := riskFixture.results.observedAt
	riskPair := riskFixture.loader.collect(context.Background(), riskFixture.results, riskFixture.fx)
	proposals := strategyProposalAuthorityPair{observedAt: now}
	accounts := strategyAccountAuthorityPair{observedAt: now}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		result := riskFixture.results.forMarket(market).result
		batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:proposal-"+string(market), map[string]strategyflow.Result{result.Lineage.Symbol: result})
		authority, ok := batch.For(result.Lineage.Symbol)
		if !ok {
			t.Fatal("missing proposal test authority")
		}
		proposal := strategyProposalMarketAuthority{market: market, entries: []strategyProposalEntryAuthority{{authority: authority}},
			snapshot: StrategyProposalMarketSnapshot{Market: market, Ready: true, Reason: StrategyProposalReady}}
		quote, cash, accountMarket := "KRW", "5000000", strategyaccount.MarketKR
		if market == StrategyMarketUS {
			quote, cash, accountMarket = "USD", "1000", strategyaccount.MarketUS
		}
		state := risk.AccountState{Mode: risk.ModeNormal, AllowedSymbols: []string{result.Lineage.Symbol}, HeldQuantity: "0",
			CashAvailable: riskcalc.Money{Amount: cash, Currency: quote}, OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
			DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"}, AccountEquity: riskcalc.Money{Amount: "10000000", Currency: "KRW"}}
		account := strategyAccountMarketAuthority{market: market,
			authority: strategyaccount.AuthorityForTest(accountMarket, quote, state, now.Add(-time.Second), now.Add(time.Minute), 1,
				"sha256:"+strings.Repeat("c", 64)), snapshot: StrategyAccountMarketSnapshot{Market: market, Ready: true, Reason: StrategyAccountReady}}
		if market == StrategyMarketKR {
			proposals.kr, accounts.kr = proposal, account
		} else {
			proposals.us, accounts.us = proposal, account
		}
	}
	schedule := pairedDispatchSchedule(now)
	fakeClock := clock.NewFake(now)
	j, err := journal.Open(context.Background(), journal.Options{Path: filepath.Join(t.TempDir(), journal.DBFileName), Clock: fakeClock,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = j.Close() })
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: fakeClock, AccountRef: "acct-risk-loader",
		Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "engine.automation_gate/risk-policy-v1"})
	if err != nil {
		t.Fatal(err)
	}
	loader := newProductionStrategyFirstLegAuthorityLoader(fakeClock, j, guardian, schedule, proposals, riskPair, riskFixture.fx, accounts)
	firstLeg := newStrategyFirstLegAdmissionBridge(guardian, loader)
	spy := &strategyDispatchGatewaySpy{observed: map[string]int{}}
	cycle := newStrategyDispatchCycle(j, spy, firstLeg, schedule, riskFixture.fx, riskPair, &strategyDispatchOwnerCoordinator{})
	cycle.revalidateSchedule = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority) error { return nil }
	return cycle, proposals, j, spy
}

func pairedDispatchSchedule(now time.Time) strategyScheduleAuthorityPair {
	makeMarket := func(market StrategyMarket, fill string) strategyScheduleMarketAuthority {
		digest := "sha256:" + strings.Repeat(fill, 64)
		desired := scheduler.DesiredState{Revision: 7, Version: scheduler.SchedulerVersion, Enabled: true, AutoStart: true,
			Market: strategySchedulerMarket(market), Session: scheduler.SessionRegular, Actor: "human", ApprovedAt: now.Add(-time.Minute),
			CalendarVersion: "sha256:" + strings.Repeat("d", 64), ConfigVersion: strategyRuntimeConfigDigest()}
		return strategyScheduleMarketAuthority{market: market, desired: desired,
			calendar: scheduler.CalendarSnapshot{Version: desired.CalendarVersion},
			restore: scheduler.RestoreResult{Restored: true, Reason: scheduler.ResumeExactManifest,
				Activation: scheduler.ActivationForTest(desired.ActivationBinding(strategyRuntimeBuildDigest()))},
			snapshot: StrategyScheduleMarketSnapshot{Market: market, Ready: true, Reason: scheduler.ResumeExactManifest,
				CalendarVersion: desired.CalendarVersion, ActivationManifestDigest: digest}}
	}
	return strategyScheduleAuthorityPair{observedAt: now, kr: makeMarket(StrategyMarketKR, "e"), us: makeMarket(StrategyMarketUS, "f")}
}

var _ strategyDispatchGateway = (*strategyDispatchGatewaySpy)(nil)
