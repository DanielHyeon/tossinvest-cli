//go:build tossos_testseams

package engine

// a146 실제 경로 시험 — 편집 전 HEAD 의 특성 시험(design F5 · F6 · F8, 영수증 `analysis/red/characterization-head.txt`)을 a146 동작으로 뒤집은 것과,
// a146 이 바꾸지 않는 범위 거절 특성(F8) 하나.
//
// 실제 경로: a112 거래 fixture(생산 적재기 · 실제 Guardian · 실제 원장 · 실제 dispatch 주기 · 생산 전달 몸통
// `dispatchStrategyMarketHandoffs`, Gateway 만 스파이 — 실주문 아님)를 감독자 `runMarket` 의 주기 함수로 돌림.
//
//   - Guardian 상한 0(거래당 위험 예산 1 KRW → 한 주도 못 덮음): 편집 전에는 effective 시장이 첫 주기에 latchMarket 으로 잠기고 refreshOnly
//     (오늘 생산 감독자의 유일한 모양) 시장은 매 주기 삼킴 계수만 오르며 그 주기의 뒤 범위가 평가되지 않았음. a146 뒤: 둘 다 잠금 · 삼킴 0,
//     매 주기 두 범위 모두 평가 · 스킵으로 셈.
//   - 범위 거절(버킷 고갈): 같은 주기의 다른 범위는 발급되지만 건너뛴 거절이 주기 오류로 돌아와 effective 시장은 잠김 — a146 범위 밖
//     (design F8 · D7, 별도 발견). a146 뒤에도 그대로여야 함.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a146CountingGuardian 은 실제 Guardian 을 그대로 쓰되 precheck 횟수를 세고 첫 거절을 보관함(주기가 몇 범위를 평가했는지 보는 계측기 ·
// 실제 Guardian 이 낸 거절 모양의 표본).
type a146CountingGuardian struct {
	*execgw.RiskGuardian
	prechecks atomic.Int32
	mu        sync.Mutex
	firstErr  error
}

func (g *a146CountingGuardian) PrecheckQFinalCampaignFirstLeg(request execgw.QFinalCampaignFirstLegIssuance) (execgw.QFinalCampaignFirstLegPrecheck, error) {
	g.prechecks.Add(1)
	precheck, err := g.RiskGuardian.PrecheckQFinalCampaignFirstLeg(request)
	g.mu.Lock()
	if g.firstErr == nil {
		g.firstErr = err
	}
	g.mu.Unlock()
	return precheck, err
}

// a146RealCapZeroRefusal 은 실제 Guardian 이 예산 1 KRW 에서 낸 precheck 거절 — 원인 사슬 안에 sentinel 이 든 진짜 모양(엔진 패키지는
// QFinalRefusal 의 원인 칸을 채울 수 없음).
func a146RealCapZeroRefusal(t *testing.T) error {
	t.Helper()
	fixture, counter := a146CapZeroFixture(t)
	_ = fixture.deliverKR(t)
	counter.mu.Lock()
	defer counter.mu.Unlock()
	var qFinal *execgw.QFinalRefusal
	if !errors.As(counter.firstErr, &qFinal) || qFinal.Code != riskbucket.RefusalExistingGuardianCap || !errors.Is(qFinal, risk.ErrStrategyQuantityZero) {
		t.Fatalf("arrangement: the real Guardian refusal is %v, not the cap-zero shape", counter.firstErr)
	}
	return counter.firstErr
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

// F4 뒤집힘(a146): 서명 활성화(effective) 시장 — 상한 0 스킵만 있는 주기는 잠그지 않음. 매 주기 두 범위가 다 스킵으로 세어지고 브로커 0.
// (편집 전 HEAD 에서는 첫 주기에 잠겼음 — `analysis/red/characterization-head.txt` 의 `…LatchesAnEffectiveMarket`.)
func TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket(t *testing.T) {
	fixture, counter := a146CapZeroFixture(t)
	snapshot, calls, faults := a146RunMarket(t, fixture, true, 3)
	if snapshot.Latched || faults != 0 || calls < 3 || snapshot.FirstFailure != "" {
		t.Fatalf("cycles=%d faults=%d snapshot=%+v — want the effective market never latched by cap-0 skips", calls, faults, snapshot)
	}
	if got := counter.prechecks.Load(); got != 2*calls || snapshot.GuardianCapZeroSkips != uint64(2*calls) {
		t.Fatalf("prechecks=%d skips=%d cycles=%d — want both scopes evaluated and skipped every cycle", got, snapshot.GuardianCapZeroSkips, calls)
	}
	if placed := fixture.placedSymbols(); len(placed) != 0 {
		t.Fatalf("placed=%v", placed)
	}
}

// F5 뒤집힘(a146): 권한 갱신 전용(refreshOnly — 오늘 생산 감독자) 시장 — 삼킴 계수가 오르지 않고, 뒤 범위가 굶지 않음.
// (편집 전 HEAD 에서는 매 주기 삼킴 + precheck 1 회였음.)
func TestA146ACapZeroEntryIsNotASwallowedFaultOnARefreshOnlyMarket(t *testing.T) {
	fixture, counter := a146CapZeroFixture(t)
	snapshot, calls, _ := a146RunMarket(t, fixture, false, 3)
	if snapshot.Latched || calls < 3 || snapshot.SwallowedCycleErrors != 0 || snapshot.FirstSwallowedFailure != "" {
		t.Fatalf("cycles=%d snapshot=%+v — want nothing swallowed for cap-0 skips", calls, snapshot)
	}
	if got := counter.prechecks.Load(); got != 2*calls || snapshot.GuardianCapZeroSkips != uint64(2*calls) {
		t.Fatalf("prechecks=%d skips=%d cycles=%d — want the second scope evaluated every cycle", got, snapshot.GuardianCapZeroSkips, calls)
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
