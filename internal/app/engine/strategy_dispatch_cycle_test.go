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
		out, err := cycle.dispatch(context.Background(), deliverForTest(t, result))
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
	// a066 5.6.1 개정(2026-09-28, Manager 승인 교차 change 편집 — a112 파일):
	// 이 fixture 의 KR·US continuation 레인은 한 계좌에서 horizon SHORT bucket 을 공유함(horizon 은 lineage 의 레인
	// horizon 에서 오고 두 레인 다 SHORT 라 fixture 로 가를 수 없음). 5.6.1 부터 같은 snapshot wave 의 공유 bucket 두 번째
	// 진입은 원장보다 적은 사용량을 주장하므로 BUCKET_USAGE_STALE 로 거절되고 다음 wave 에 성립함. 이 fixture 의 risk
	// snapshot 원천은 고정 파일이라 "다음 wave" 를 만들 수 없음 — 다음 wave 성립은 journal 수준 시험
	// (TestA066KRUSConcurrentContract, TestFirstLegAtomicAdmissionSameAccountKRUSRecollectsTheSerializedLoser)이 잼.
	// 이 시험의 주제(두 dispatch 가 동시에 출발하고, 하나의 중앙 owner 아래에서 돌며, 번갈아 기다리지 않음)는 그대로 단언함:
	// 동시 출발 · 결과 둘 다 도착 · 성립한 쪽의 lease 와 거절된 쪽이 같은 중앙 owner(첫 epoch) · 거절된 쪽은 Gateway 에 닿지 않음.
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
		// 봉투는 경주 밖에서 미리 만든다. 고루틴 안에서 만들면 t.Fatalf 가
		// 시험 고루틴 밖에서 불릴 수 있고, 두 dispatch 가 동시에 출발한다는
		// 이 시험의 요점도 흐려진다.
		delivered := deliverForTest(t, proposals.forMarket(market).entries[0].authority.Proposal())
		runners.Add(1)
		go func() {
			defer runners.Done()
			<-start
			out, err := cycle.dispatch(context.Background(), delivered)
			results <- result{market: market, out: out, err: err}
		}()
	}
	close(start)
	runners.Wait()
	close(results)
	var admitted, refused []result
	for result := range results {
		switch {
		case result.err == nil && result.out.State == journal.StateConfirmed:
			admitted = append(admitted, result)
		case result.err != nil && strings.Contains(result.err.Error(), "engine: first-leg admission ATOMIC_ADMISSION_FAILED") &&
			strings.Contains(result.err.Error(), "BUCKET_USAGE_STALE"):
			// first-leg 다리가 오류를 ATOMIC_ADMISSION_FAILED 문자열로 눕히므로 타입이 아니라 거절 코드 문자열로 가름.
			// "first-leg admission" 접두는 dispatch 가 admission 에 닿았다는 뜻 — admission 바로 앞에서 dispatchOwner 를
			// 지나므로 거절된 쪽도 중앙 owner 를 얻은 뒤임.
			refused = append(refused, result)
		default:
			t.Fatalf("%s outcome=%+v err=%v", result.market, result.out, result.err)
		}
	}
	if len(admitted) != 1 || len(refused) != 1 || admitted[0].market == refused[0].market {
		t.Fatalf("one wave on shared buckets must admit exactly one market and refuse the other as stale: admitted=%+v refused=%+v", admitted, refused)
	}
	spy.mu.Lock()
	calls := append([]execgw.StrategyPlaceRequest(nil), spy.calls...)
	spy.mu.Unlock()
	if len(calls) != 1 || !strings.EqualFold(calls[0].Intent.Market, string(admitted[0].market)) {
		t.Fatalf("Gateway calls=%+v, want exactly the admitted %s market", calls, admitted[0].market)
	}
	lease, err := j.LookupStrategyDispatchLease(context.Background(), calls[0].Lease.LeaseID)
	if err != nil {
		t.Fatal(err)
	}
	// 하나의 중앙 owner: 두 dispatch 가 모두 dispatchOwner 를 거쳤고(거절된 쪽도 admission 전에 owner 를 얻음) owner 는
	// 한 번만 획득됨 — 첫 epoch 이고, 성립한 lease 가 그 owner 의 epoch·fencing 을 담음.
	central := cycle.owner.owner
	if central.Epoch != 1 || lease.OwnerEpoch != central.Epoch || lease.FencingToken != central.FencingToken {
		t.Fatalf("central owner epoch=%d/%s, admitted lease owner=%d/%s — the two dispatches did not share one owner",
			central.Epoch, central.FencingToken, lease.OwnerEpoch, lease.FencingToken)
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
				if _, err := cycle.dispatch(context.Background(), deliverForTest(t, result)); !errors.Is(err, refusal) {
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
	cycle := newStrategyDispatchCycle(j, spy, firstLeg, schedule, riskFixture.fx, riskPair, proposals, &strategyDispatchOwnerCoordinator{})
	cycle.revalidateSchedule = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority) error { return nil }
	// 생산 조립과 같은 모양으로 실시계를 넣는다 (태스크 8.7.2).
	cycle.now = fakeClock.Now
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
