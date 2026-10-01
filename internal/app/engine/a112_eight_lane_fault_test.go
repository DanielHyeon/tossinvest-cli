//go:build tossos_testseams

package engine

// a112 태스크 5.6.2.2 — 5.6.1 이 두 시장 worker 위에서 잰 세 절을 **여덟 레인 런타임**(5.1.2 · 5.2 교체 뒤) 위에서 다시 잰다.
//
// 교체가 실제로 바꾼 모양(재유도 — 측정으로 확인, review 「5.6.2.2」 절):
//   - 감독자 worker 는 여전히 **시장 둘**이다(`NewRefreshingPairedStrategyEntrySupervisor` — KR · US). 여덟 레인은 worker 가 아니라 시장 주기
//     (`runProductionStrategyMarketCycle`) **안**에서 도는 레인 런타임(`strategyLaneRuntime.evaluate`)이고, 레인 고장은 레인 자기 잠금(strategyworker.Lane)
//     에서 끝난다 — 감독자의 `latchMarket` · fault 스트림에 닿지 않는다. 그래서 「fault 스트림 용량 = 잠길 수 있는 worker 수」 등식은 2 = 2 그대로이고
//     (레인 여덟은 그 수에 들지 않는다), handoff `default` 팔은 여전히 도달 불가다.
//   - 생산 감독자의 worker 서술자는 고정으로 `Effective=false · RefreshesAuthority=true` 라 refresh-only 삼킴(`runMarket` 의 refreshOnly 갈래)이 여전히
//     생산 구성이다 — 교체가 바꾸지 않았다.
//   - 시장 주기가 레인 쪽에서 오류를 받는 유일한 자리는 durable latch 기록 실패 · 빌드에 없는 레인 기록(5.3.3)이고, refresh-only 갈래가 그것을 기록하며
//     삼킨다(중앙 무결성이 아니다).
//
// 고장 주입: 생산 레인 step(`strategyFamilyLaneStep`)은 고장 입력이 없다(순수 평가 — 설계가 그렇게 막았다). 그래서 레인의 **자기 고장 입구**
// `Lane.Fail(reason, abnormal=true)` 로 여덟을 동시에 잠근다 — 유계 사이클의 panic · 마감 시한 정산이 부르는 것과 같은 입구다(strategyworker 시험이
// RunBounded → settle → failLocked 를 잰다).

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

type a112EightLaneRun struct {
	context    *Context
	supervisor *StrategyEntrySupervisor
	gate       *execgw.EntryGate
	lanes      *strategyLaneRuntime
	journal    *journal.Journal
	done       chan error
	cancel     context.CancelFunc
	safety     map[string]chan struct{}
}

// newA112EightLaneRun 은 생산 생성자로 레인 런타임 · 감독자를 세우고, 안전 loop 셋(fill detection · reconcile · exit observation 자리)과 함께
// Runtime 에 넣는다. 권한 새로 고침은 Context 의 1초 캐시에 fixture 조립을 넣어 주입한다(원격 0). ledger 가 nil 이 아니면 레인 원장을 바꾼다.
func newA112EightLaneRun(t *testing.T, ledger strategyLaneLedger) *a112EightLaneRun {
	t.Helper()
	cycle, proposals, j, _ := pairedStrategyDispatchCycleFixture(t)
	loader := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	clk := loader.clk
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	c := &Context{Journal: j, AccountRef: "acct-risk-loader", Entry: gate}
	c.strategyRefresh = &StrategyEntryProductionAssembly{dispatch: cycle, proposals: proposals, schedule: cycle.schedule}
	c.strategyRefreshAt = clk.Now()

	lanes, err := c.productionStrategyLanes(context.Background(), clk)
	if err != nil || lanes == nil {
		t.Fatalf("arrangement: production lane runtime: %v", err)
	}
	if ledger != nil {
		lanes.ledger = ledger
	}
	if got := len(lanes.lanes); got != 8 {
		t.Fatalf("arrangement: production lanes=%d, want 8", got)
	}
	// 여덟 레인을 **동시에** 잠근다(비정상 — 임계값을 기다리지 않음).
	var faults sync.WaitGroup
	for _, lane := range lanes.lanes {
		faults.Add(1)
		go func(lane *strategyworker.Lane) {
			defer faults.Done()
			lane.Fail("injected lane fault (5.6.2.2)", true)
		}(lane)
	}
	faults.Wait()
	for _, lane := range lanes.lanes {
		if !lane.Latched() {
			t.Fatalf("arrangement: lane %v not latched", lane.Key())
		}
	}

	supervisor, err := c.NewRefreshingPairedStrategyEntrySupervisor(clk)
	if err != nil {
		t.Fatal(err)
	}
	run := &a112EightLaneRun{context: c, supervisor: supervisor, gate: gate, lanes: lanes, journal: j, done: make(chan error, 1),
		safety: map[string]chan struct{}{}}
	loops := []SupervisedLoop{supervisor.SupervisedLoop()}
	for _, name := range []string{"fill-detection", "reconcile", "exit-observation"} {
		stopped := make(chan struct{})
		run.safety[name] = stopped
		loops = append(loops, SupervisedLoop{Name: name, Run: func(ctx context.Context) error {
			<-ctx.Done()
			close(stopped)
			return ctx.Err()
		}})
	}
	runtime, err := NewRuntime(RuntimeOptions{Loops: loops})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	run.cancel = cancel
	t.Cleanup(cancel)
	go func() { run.done <- runtime.Run(ctx) }()
	select {
	case <-supervisor.Ready():
	case <-time.After(3 * time.Second):
		t.Fatal("strategy supervisor readiness")
	}
	return run
}

// waitBothMarketsObserved 는 두 시장 주기가 레인 여덟 전부를 한 번 돌았을 때까지 기다린다(감독자의 첫 poll 은 즉시다).
func (run *a112EightLaneRun) waitBothMarketsObserved(t *testing.T) []strategyLaneObservation {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		observed := 0
		values := run.lanes.observations()
		for _, observation := range values {
			if observation.Trigger != "" {
				observed++
			}
		}
		if observed == 8 {
			return values
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("lane observations did not cover all eight lanes: %+v", run.lanes.observations())
	return nil
}

// 세 절(Done): 여덟 레인이 동시에 고장 나도 ① 안전 loop 셋은 계속 돈다 ② 엔진(Runtime)은 서지 않는다 ③ 레인 고장은 감독자의 fault 스트림 ·
// 시장 잠금 · 진입 게이트에 닿지 않는다(국소). 그리고 두 시장 주기는 잠긴 레인 여덟을 실제로 돌았다(관측 DISABLED · LATCHED).
func TestEightSimultaneousLaneFaultsLeaveTheSafetyLoopsRunning(t *testing.T) {
	run := newA112EightLaneRun(t, nil)
	observations := run.waitBothMarketsObserved(t)
	for _, observation := range observations {
		if observation.Trigger != strategyworker.TriggerDisabled || observation.Health != strategyworker.LaneLatched {
			t.Fatalf("lane %v observation trigger=%s health=%s, want DISABLED · LATCHED", observation.Key, observation.Trigger, observation.Health)
		}
	}
	// 국소: 레인 고장은 감독자의 장부를 건드리지 않는다.
	if got := len(run.supervisor.Faults()); got != 0 {
		t.Fatalf("lane faults reached the supervisor's fault stream (%d) — a lane fault must stay in its lane", got)
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		// 잠긴 레인은 시장 주기의 오류도 아니다(삼킨 오류 0) — 레인 고장이 시장 주기 오류로 올라오면 그것이 이미 경계 넘기다.
		if snapshot, _ := run.supervisor.Snapshot(market); snapshot.Latched || snapshot.SwallowedCycleErrors != 0 {
			t.Fatalf("%s market worker touched by lane faults: %+v", market, snapshot)
		}
	}
	if blocks := run.gate.Blocks(); len(blocks) != 0 {
		t.Fatalf("lane faults latched the entry gate: %v", blocks)
	}
	// 생존: 엔진 · 안전 loop.
	select {
	case err := <-run.done:
		t.Fatalf("eight lane faults stopped the runtime (and every safety loop): %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	for name, stopped := range run.safety {
		select {
		case <-stopped:
			t.Fatalf("safety loop %s stopped after eight lane faults", name)
		default:
		}
	}
	// 레인 잠금은 durable 하다(5.3.3) — 여덟 전부 원장에 남는다. 관측 기록(record)이 원장 기록(persist)보다 앞이므로 원장 쪽을 기다린다.
	var open []journal.StrategyLaneLatch
	var err error
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		if open, err = run.journal.OpenStrategyLaneLatches(context.Background(), "acct-risk-loader"); err == nil && len(open) == 8 {
			break
		}
	}
	if err != nil || len(open) != 8 {
		t.Fatalf("durable lane latches=%d (err=%v), want 8", len(open), err)
	}
	run.cancel()
	if err := <-run.done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("graceful stop=%v", err)
	}
}

// failingLaneLedger 는 레인 잠금을 원장에 남기지 못하는 원장이다(기록 실패 — 레인 쪽이 시장 주기에 돌려주는 유일한 오류).
type failingLaneLedger struct{}

func (failingLaneLedger) OpenStrategyLaneLatches(context.Context, string) ([]journal.StrategyLaneLatch, error) {
	return nil, nil
}

func (failingLaneLedger) RecordStrategyLaneLatch(context.Context, journal.StrategyLaneLatch) (journal.StrategyLaneLatch, error) {
	return journal.StrategyLaneLatch{}, errors.New("lane latch ledger write failed")
}

func (failingLaneLedger) RecoverStrategyLaneLatch(context.Context, int64, uint64) error { return nil }

// 레인 잠금을 원장에 **남기지 못하면** 시장 주기가 오류를 돌려준다(5.3.3 — 조용히 넘기면 재시작이 잠긴 레인을 연다). 생산 구성(refresh-only)에서
// 그 오류는 기록되며 삼켜지고(포화 계수), 중앙 무결성이 아니므로 진입 게이트 · 엔진 · 안전 loop 를 건드리지 않는다.
func TestALaneLatchThatCannotBeRecordedIsCountedNotEscalated(t *testing.T) {
	run := newA112EightLaneRun(t, failingLaneLedger{})
	run.waitBothMarketsObserved(t)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		kr, _ := run.supervisor.Snapshot(StrategyMarketKR)
		us, _ := run.supervisor.Snapshot(StrategyMarketUS)
		if kr.SwallowedCycleErrors > 0 && us.SwallowedCycleErrors > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		snapshot, _ := run.supervisor.Snapshot(market)
		if snapshot.SwallowedCycleErrors == 0 || snapshot.Latched {
			t.Fatalf("%s snapshot=%+v — want the unrecorded lane latch counted as a swallowed cycle error, not a market latch", market, snapshot)
		}
	}
	if blocks := run.gate.Blocks(); len(blocks) != 0 || len(run.supervisor.Faults()) != 0 {
		t.Fatalf("an unrecorded lane latch escalated: gate=%v faults=%d", blocks, len(run.supervisor.Faults()))
	}
	select {
	case err := <-run.done:
		t.Fatalf("an unrecorded lane latch stopped the runtime: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	for name, stopped := range run.safety {
		select {
		case <-stopped:
			t.Fatalf("safety loop %s stopped", name)
		default:
		}
	}
}

// 재유도 ①(구조): 생산 감독자의 worker 는 시장 둘이고 둘 다 refresh-only 다 — 여덟 레인은 worker 가 아니므로 fault 스트림 용량 등식(2 = 2)과 handoff
// `default` 팔의 도달 불가가 교체 뒤에도 그대로다. 누군가 레인을 감독자 worker 로 올리면(교체의 다른 모양) 이 시험이 그 순간을 알린다.
func TestTheProductionSupervisorStillHasTwoRefreshOnlyMarketWorkers(t *testing.T) {
	run := newA112EightLaneRun(t, nil)
	if got := len(run.supervisor.workers); got != 2 {
		t.Fatalf("production supervisor workers=%d, want 2 market workers — if lanes became supervisor workers, re-derive the fault "+
			"stream capacity (5.6.1 등식) and the refresh-only swallow", got)
	}
	if got, want := cap(run.supervisor.Faults()), len(run.supervisor.workers); got != want {
		t.Fatalf("fault stream capacity=%d, workers=%d", got, want)
	}
	for market, worker := range run.supervisor.workers {
		if worker.effective || !worker.descriptor.RefreshesAuthority {
			t.Fatalf("%s worker effective=%v refreshes=%v — the production configuration is refresh-only", market, worker.effective,
				worker.descriptor.RefreshesAuthority)
		}
	}
}
