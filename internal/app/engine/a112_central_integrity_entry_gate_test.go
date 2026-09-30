package engine_test

// a112 태스크 5.6.2.1 — 사람 결정 (6)(2026-09-30, HANDOFF 「결정 (1)(5)(6) 기록」): fail-closed 의 수단은 프로세스 정지가
// 아니라 `execgw.EntryGate.Block` 이다.
//
// 오늘 생산이 도는 유일한 구성(권한 갱신 전용 worker)에서는 `runMarket` 의 판정 순서가 refreshOnly → 중앙 무결성이라
// 중앙 무결성 오류조차 삼켜졌다. 순서는 루프 생존을 위해 그대로 두고, 삼키지 않고 신규 진입을 닫는다:
//   ① 게이트에 전략 중앙 무결성 사유가 선다(신규 진입 fail-closed — design.md:198 고장표).
//   ② 감독자 Run 이 반환하지 않는다(엔진 정지 = 손절 없음 — spec 「lane worker 가 safety loop 를 취소해서는 안 된다」).
//   ③ 다른 시장은 계속 돈다(시장 국소성).
// 대조: 보통 사이클 오류는 게이트를 잠그지 않는다 — 이 fail-closed 가 거부하는 정상 입력이 없음을 값으로 둔다.

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

type a112CentralRun struct {
	gate     *execgw.EntryGate
	kr, us   *atomic.Int32
	done     chan error
	cancel   context.CancelFunc
	fake     *clock.Fake
	superv   *engine.StrategyEntrySupervisor
	withGate bool
}

// a112RunRefreshOnlyPair 는 두 시장을 생산과 같은 권한 갱신 전용 worker 로 세우고, KR 사이클은 krErr 를 돌려줌.
func a112RunRefreshOnlyPair(t *testing.T, krErr error, withGate bool) *a112CentralRun {
	t.Helper()
	base := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	fake := clock.NewFake(base)
	var kr, us atomic.Int32
	worker := func(market engine.StrategyMarket, cycle engine.StrategyCycle) engine.StrategyMarketWorker {
		return engine.StrategyMarketWorker{Market: market, Cycle: cycle,
			PollInterval: engine.DefaultStrategyCycleLimit, RefreshesAuthority: true}
	}
	opts := engine.StrategyEntrySupervisorOptions{
		Clock: fake, CycleLimit: engine.MaximumStrategyCycleLimit,
		Workers: []engine.StrategyMarketWorker{
			worker(engine.StrategyMarketKR, func(context.Context) error { kr.Add(1); return krErr }),
			worker(engine.StrategyMarketUS, func(context.Context) error { us.Add(1); return nil }),
		},
	}
	gate := execgw.NewEntryGate(fake, map[execgw.RequiredQuery]time.Duration{})
	if withGate {
		opts.EntryGate = gate
	}
	supervisor := mustStrategySupervisor(t, opts)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	return &a112CentralRun{gate: gate, kr: &kr, us: &us, done: done, cancel: cancel, fake: fake, superv: supervisor, withGate: withGate}
}

// driveCycles 는 두 시장이 각각 n 번 이상 돌 때까지 가짜 시계를 민다(또는 감독자가 반환할 때까지).
func (r *a112CentralRun) driveCycles(n int32) {
	deadline := time.Now().Add(3 * time.Second)
	for (r.kr.Load() < n || r.us.Load() < n) && time.Now().Before(deadline) {
		select {
		case err := <-r.done:
			r.done <- err
			return
		default:
		}
		r.fake.Advance(engine.DefaultStrategyCycleLimit)
		time.Sleep(time.Millisecond)
	}
}

func TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine(t *testing.T) {
	run := a112RunRefreshOnlyPair(t, engine.StrategyCentralIntegrityFailure(errors.New("owner fence CAS failed")), true)
	run.driveCycles(2)

	// ① 신규 진입 fail-closed.
	detail, blocked := run.gate.Blocks()[execgw.ReasonStrategyCentralIntegrity]
	if !blocked {
		t.Fatalf("a central integrity fault on the refresh-only worker left new entries open — it was swallowed "+
			"(결정 (6): fail-closed 의 수단은 EntryGate). blocks=%v", run.gate.Blocks())
	}
	if !strings.Contains(detail, string(engine.StrategyMarketKR)) {
		t.Errorf("latch detail does not name the market: %q", detail)
	}
	if rejected := run.gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonStrategyCentralIntegrity {
		t.Errorf("CheckEntry = %v, want the strategy central integrity refusal", rejected)
	}
	// ② 엔진(감독자 Run)은 서지 않는다.
	select {
	case err := <-run.done:
		t.Fatalf("the central fault stopped the supervisor (and with it every safety loop): %v", err)
	default:
	}
	// ③ 다른 시장은 계속 돈다 — 그리고 KR 자신도 다음 poll 을 계속 받는다(잠그는 것은 진입이지 루프가 아님).
	if run.us.Load() < 2 || run.kr.Load() < 2 {
		t.Fatalf("cycles kr=%d us=%d, want both ≥2 after the fault", run.kr.Load(), run.us.Load())
	}
	run.cancel()
	if err := <-run.done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

// 대조: 보통 사이클 오류(원장 읽기 실패 · Gateway 거절 등 — 오늘 생산 사이클이 실제로 내는 것)는 게이트를 잠그지 않는다.
// 이 fail-closed 가 닫는 입력은 중앙 무결성 표지를 단 오류뿐이고, 그 표지를 만드는 생산 호출자는 0 이다
// (`a112_central_integrity_census_test.go`).
func TestAnOrdinaryRefreshOnlyCycleErrorDoesNotBlockEntry(t *testing.T) {
	run := a112RunRefreshOnlyPair(t, errors.New("journal read failed"), true)
	run.driveCycles(2)
	if blocks := run.gate.Blocks(); len(blocks) != 0 {
		t.Fatalf("an ordinary cycle error latched the entry gate: %v", blocks)
	}
	run.cancel()
	if err := <-run.done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

// 게이트가 없는 조립(생산에는 없다 — 생산 생성자가 게이트를 요구함)에서는 중앙 무결성 오류를 삼키지 않고 기존의 프로세스
// 전체 fail-closed 로 올린다. 조용한 삼킴은 어느 조립에서도 남지 않는다.
func TestWithoutAnEntryGateACentralFaultIsNotSwallowed(t *testing.T) {
	run := a112RunRefreshOnlyPair(t, engine.StrategyCentralIntegrityFailure(errors.New("owner fence CAS failed")), false)
	run.driveCycles(2)
	select {
	case err := <-run.done:
		if !errors.Is(err, engine.ErrStrategyCentralIntegrity) {
			t.Fatalf("Run=%v, want the central integrity failure", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("with no entry gate the central fault was swallowed")
	}
}
