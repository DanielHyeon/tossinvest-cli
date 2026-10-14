//go:build tossos_testseams

package engine

// a137 구현 리뷰 P1-3: 일일 손실 계좌 차단 · 측정 불가 결과는 **두 레인 종류 모두**(서명 활성화 effective · 권한 갱신 전용 refreshOnly)에서
// 주기 오류가 아님 — 잠금(latchMarket)도 삼킴 계수도 없음. 대조군: 표지 없는 결함은 오늘처럼 effective 는 잠기고 refreshOnly 는 삼켜 셈.

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
)

func a137Supervisor(t *testing.T, effective bool, cycle StrategyCycle) (*StrategyEntrySupervisor, *clock.Fake) {
	t.Helper()
	fake := clock.NewFake(time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC))
	worker := func(market StrategyMarket) StrategyMarketWorker {
		if !effective {
			return StrategyMarketWorker{Market: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: cycle}
		}
		return StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
			AuthorityExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			EvidenceDigest:     "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			LatchRevision:      1, PollInterval: DefaultStrategyCycleLimit}
	}
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: fake, CycleLimit: MaximumStrategyCycleLimit,
		Workers: []StrategyMarketWorker{worker(StrategyMarketKR), worker(StrategyMarketUS)}})
	if err != nil {
		t.Fatal(err)
	}
	return supervisor, fake
}

// a137RunCycles 는 KR 주기가 최소 n 번 돌 때까지 감독자를 돌리고 KR 스냅샷을 돌려줌.
func a137RunCycles(t *testing.T, effective bool, failure error, n int32) (StrategyWorkerSnapshot, int32) {
	t.Helper()
	var calls atomic.Int32
	cycle := func(ctx context.Context) error {
		calls.Add(1)
		return deliverEachStrategyHandoff(strategyhandoff.AdmitEachOwnerScope(true, a137Selected()), func(strategyhandoff.Delivered) error {
			return failure
		})
	}
	supervisor, fake := a137Supervisor(t, effective, cycle)
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
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if snapshot, ok := supervisor.Snapshot(StrategyMarketKR); ok && (snapshot.Latched || calls.Load() >= n) {
			break
		}
		fake.Advance(DefaultStrategyCycleLimit)
		time.Sleep(time.Millisecond)
	}
	// 마지막 주기의 결과가 기록될 시간을 줌.
	time.Sleep(20 * time.Millisecond)
	snapshot, _ := supervisor.Snapshot(StrategyMarketKR)
	return snapshot, calls.Load()
}

func TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind(t *testing.T) {
	quiet := map[string]error{
		"account block": &execgw.DailyLossEntryBlock{Cause: errors.New("limit")},
		"unmeasurable":  &execgw.DailyLossUnmeasurable{Cause: errors.New("closed today without a frozen outcome row")},
	}
	for name, failure := range quiet {
		for _, effective := range []bool{true, false} {
			snapshot, calls := a137RunCycles(t, effective, failure, 3)
			if calls < 3 || snapshot.Latched || snapshot.SwallowedCycleErrors != 0 {
				t.Errorf("%s effective=%v: cycles=%d snapshot=%+v — want ≥3 cycles, no latch and nothing swallowed", name, effective, calls, snapshot)
			}
		}
	}
}

func TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday(t *testing.T) {
	fault := errors.New("engine: first-leg admission AUTHORITY_MISMATCH: equity missing")
	if snapshot, _ := a137RunCycles(t, true, fault, 3); !snapshot.Latched {
		t.Fatalf("effective lane snapshot=%+v — an unmarked fault must latch as today", snapshot)
	}
	if snapshot, _ := a137RunCycles(t, false, fault, 3); snapshot.Latched || snapshot.SwallowedCycleErrors == 0 {
		t.Fatalf("refresh-only lane snapshot=%+v — an unmarked fault must be counted as swallowed as today", snapshot)
	}
}
