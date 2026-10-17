//go:build tossos_testseams

package engine

// a146 특성 시험(편집 전 HEAD 의 동작을 실제 경로로 못 박음 — design F1~F5).
//
// 실제 경로: a112 거래 fixture(생산 적재기 · 실제 Guardian · 실제 원장 · 실제 dispatch 주기 · 생산 전달 몸통
// `dispatchStrategyMarketHandoffs`, Gateway 만 스파이 — 실주문 아님)를 감독자 `runMarket` 의 주기 함수로 돌림.
//
//   - Guardian 상한 0(거래당 위험 예산 1 KRW → 한 주도 못 덮음): 서명 활성화(effective) 시장은 첫 주기에 latchMarket 으로 잠기고,
//     권한 갱신 전용(refreshOnly — 오늘 생산 감독자의 유일한 모양) 시장은 매 주기 삼킴 계수만 오름. 어느 쪽이든 그 주기의 뒤 범위는
//     평가되지 않음(precheck 1 회).
//   - 범위 거절(버킷 고갈): 같은 주기의 다른 범위는 발급되지만 건너뛴 거절이 주기 오류로 돌아와 effective 시장은 잠김 — a146 범위 밖
//     (design F6, 별도 발견). 이 둘은 a146 뒤에도 그대로여야 함.

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a146CountingGuardian 은 실제 Guardian 을 그대로 쓰되 precheck 횟수만 셈(주기가 몇 범위를 평가했는지 보는 계측기).
type a146CountingGuardian struct {
	*execgw.RiskGuardian
	prechecks atomic.Int32
}

func (g *a146CountingGuardian) PrecheckQFinalCampaignFirstLeg(request execgw.QFinalCampaignFirstLegIssuance) (execgw.QFinalCampaignFirstLegPrecheck, error) {
	g.prechecks.Add(1)
	return g.RiskGuardian.PrecheckQFinalCampaignFirstLeg(request)
}

// a146CapZeroFixture 는 거래당 위험 예산 1 KRW 의 거래 fixture — 두 KR 범위(005930 · 000660, 둘 다 손절 폭 5 KRW)의 Guardian 상한이 0.
func a146CapZeroFixture(t *testing.T) (a112TradingFixture, *a146CountingGuardian) {
	t.Helper()
	fixture := newA112TradingFixture(t, a112TradingOptions{riskBudget: "1"})
	counter := &a146CountingGuardian{RiskGuardian: fixture.guardian}
	fixture.cycle.firstLeg = newStrategyFirstLegAdmissionBridge(counter, fixture.loader)
	return fixture, counter
}

// a146BucketExhaustedFixture 는 둘째 범위(000660)를 reversal family 로 두고 그 family 버킷 한도를 0 으로 — 범위 거절 하나 + 발급 하나.
func a146BucketExhaustedFixture(t *testing.T) a112TradingFixture {
	t.Helper()
	reversal := a112KRDescriptorOfFamily(t, strategyrouter.FamilyReversal)
	return newA112TradingFixture(t, a112TradingOptions{secondLane: reversal,
		extraStrategies: []riskLoaderStrategy{{LaneID: reversal.LaneID, LaneVersion: reversal.LaneVersion, Horizon: riskbucket.HorizonShort,
			RiskID: "reversal", RiskVersion: "reversal-risk-v1", LimitMinor: "0"}}})
}

// a146RunMarket 은 KR 시장 하나를 fixture 의 실제 전달 몸통으로 돌리는 감독자를 세우고, 주기가 n 번 돌거나 잠길 때까지 돌린 뒤 KR 스냅숏을 돌려줌.
// US 는 아무것도 하지 않는 주기(같은 fixture 를 두 고루틴이 동시에 쓰지 않게).
func a146RunMarket(t *testing.T, fixture a112TradingFixture, effective bool, n int32) (StrategyWorkerSnapshot, int32, int) {
	t.Helper()
	var calls, finished atomic.Int32
	krCycle := func(ctx context.Context) error {
		calls.Add(1)
		defer finished.Add(1)
		return dispatchStrategyMarketHandoffs(ctx, fixture.journal, fixture.cycle, fixture.proposals.kr.dispatchHandoffs())
	}
	idle := func(context.Context) error { return nil }
	worker := func(market StrategyMarket, cycle StrategyCycle) StrategyMarketWorker {
		if !effective {
			return StrategyMarketWorker{Market: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: cycle}
		}
		return StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
			AuthorityExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			EvidenceDigest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			LatchRevision:      1, PollInterval: DefaultStrategyCycleLimit}
	}
	fake := clock.NewFake(time.Date(2026, 10, 17, 1, 0, 0, 0, time.UTC))
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: fake, CycleLimit: MaximumStrategyCycleLimit,
		Workers: []StrategyMarketWorker{worker(StrategyMarketKR, krCycle), worker(StrategyMarketUS, idle)}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	defer func() {
		cancel()
		<-done
	}()
	select {
	case <-supervisor.Ready():
	case <-time.After(3 * time.Second):
		t.Fatal("supervisor not ready")
	}
	// 가짜 시계는 주기가 돌지 않을 때만 한 폴 간격씩 민다 — 실제 원장 · 적재기를 도는 주기 도중에 밀면 감시견(30 초)이 그 주기를
	// 버려 결과가 deadline 으로 바뀜. 폴러는 기동 즉시 한 번 넣으므로 첫 주기는 밀기 전에 돎.
	settle := func() {
		until := time.Now().Add(3 * time.Second)
		for time.Now().Before(until) && finished.Load() < calls.Load() {
			time.Sleep(time.Millisecond)
		}
		time.Sleep(20 * time.Millisecond) // 감독자가 주기 결과를 반영할 시간
	}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		settle()
		if snapshot, ok := supervisor.Snapshot(StrategyMarketKR); ok && (snapshot.Latched || calls.Load() >= n) {
			break
		}
		before := calls.Load()
		fake.Advance(DefaultStrategyCycleLimit)
		for wait := time.Now().Add(time.Second); time.Now().Before(wait) && calls.Load() == before; {
			time.Sleep(time.Millisecond)
		}
	}
	settle()
	snapshot, _ := supervisor.Snapshot(StrategyMarketKR)
	return snapshot, calls.Load(), len(supervisor.Faults())
}

// F1 · F3: 전달 몸통 한 번 — Guardian 상한 0 은 AUTHORITY_MISMATCH(EXISTING_GUARDIAN_CAP) 오류로 주기를 멈추고 둘째 범위는 평가되지 않음.
func TestA146CharacterizationACapZeroEntryStopsTheCycleAtItsScope(t *testing.T) {
	fixture, counter := a146CapZeroFixture(t)
	err := fixture.deliverKR(t)
	if err == nil || a112ScopeRefusalOf(err) != nil || !strings.Contains(err.Error(), string(StrategyFirstLegAuthorityMismatch)) ||
		!strings.Contains(err.Error(), string(riskbucket.RefusalExistingGuardianCap)) {
		t.Fatalf("err=%v — want today's untyped AUTHORITY_MISMATCH carrying EXISTING_GUARDIAN_CAP", err)
	}
	if got := counter.prechecks.Load(); got != 1 {
		t.Fatalf("prechecks=%d — want 1: today the cycle stops at the first scope and the second is never evaluated", got)
	}
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v — a cap-0 entry must never reach the broker", placed)
	}
}

// F4: 서명 활성화(effective) 시장 — 그 오류가 첫 주기에 latchMarket 으로 시장을 잠금(감독자 메모리 잠금, fault 하나).
func TestA146CharacterizationACapZeroEntryLatchesAnEffectiveMarket(t *testing.T) {
	fixture, _ := a146CapZeroFixture(t)
	snapshot, calls, faults := a146RunMarket(t, fixture, true, 3)
	if !snapshot.Latched || snapshot.FirstRefusal != StrategyWorkerRefusalFailure || faults != 1 ||
		!strings.Contains(snapshot.FirstFailure, string(riskbucket.RefusalExistingGuardianCap)) {
		t.Fatalf("cycles=%d faults=%d snapshot=%+v — want the effective market latched by the cap-0 refusal", calls, faults, snapshot)
	}
	if calls != 1 {
		t.Fatalf("cycles=%d — a latched market must not evaluate again", calls)
	}
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v", placed)
	}
}

// F5: 권한 갱신 전용(refreshOnly — 오늘 생산 감독자) 시장 — 잠기지 않고 매 주기 삼킨 오류로 셈.
func TestA146CharacterizationACapZeroEntryIsASwallowedFaultOnARefreshOnlyMarket(t *testing.T) {
	fixture, counter := a146CapZeroFixture(t)
	snapshot, calls, _ := a146RunMarket(t, fixture, false, 3)
	if snapshot.Latched || calls < 3 || snapshot.SwallowedCycleErrors < 2 ||
		!strings.Contains(snapshot.FirstSwallowedFailure, string(riskbucket.RefusalExistingGuardianCap)) {
		t.Fatalf("cycles=%d snapshot=%+v — want not latched and every cycle counted as a swallowed fault", calls, snapshot)
	}
	// 매 주기 첫 범위에서 멈춤 — precheck 수 = 주기 수(둘째 범위는 한 번도 평가되지 않음).
	if got := counter.prechecks.Load(); got != calls {
		t.Fatalf("prechecks=%d cycles=%d — want one precheck per cycle (the second scope starves)", got, calls)
	}
}

// F6(a146 범위 밖 — 별도 발견): 버킷 고갈 범위 거절은 같은 주기의 다른 범위 발급을 막지 않지만, 건너뛴 거절이 주기 오류로 돌아와
// effective 시장을 잠금(refreshOnly 는 삼킴). a146 은 이것을 바꾸지 않음 — 편집 뒤에도 이 시험은 그대로 통과해야 함.
func TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket(t *testing.T) {
	fixture := a146BucketExhaustedFixture(t)
	snapshot, _, faults := a146RunMarket(t, fixture, true, 3)
	if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {
		t.Fatalf("placed=%s — want the continuation scope issued in the same cycle", got)
	}
	if !snapshot.Latched || faults != 1 || !strings.Contains(snapshot.FirstFailure, string(riskbucket.RefusalBucketCapExhausted)) {
		t.Fatalf("faults=%d snapshot=%+v — want the effective market latched by the reported scope refusal", faults, snapshot)
	}

	refresh := a146BucketExhaustedFixture(t)
	snapshot, _, _ = a146RunMarket(t, refresh, false, 3)
	if snapshot.Latched || snapshot.SwallowedCycleErrors == 0 || !strings.Contains(snapshot.FirstSwallowedFailure, string(riskbucket.RefusalBucketCapExhausted)) {
		t.Fatalf("snapshot=%+v — want the refresh-only market to count the scope refusal as swallowed", snapshot)
	}
}
