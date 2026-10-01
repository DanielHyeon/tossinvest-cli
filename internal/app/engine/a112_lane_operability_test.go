package engine

// a112 태스크 7.5 — 성능 · 운용성 시험(Manager 판정 2026-10-01: D1=(A), D2, C2, C3; C1 은 a112_lane_fanout_test.go).
//
//   - D1 레인 지연 독립(시험은 a112_lane_latency_testseam_test.go — seam 이 tossos_testseams 빌드에만 있다): 한 시장의 레인 넷은 시장 주기 **안에서** 동시에 돌고 join 된다. 멈춘 레인은 자기 마감 시한에 버려지고 이웃은
//     정상 지연으로 관측을 남긴다 — 시장 지연은 합이 아니라 최댓값(마감 시한 1 회). 스펙 시나리오 「KR breakout worker timeout」과 설계
//     「느린 worker 때문에 다른 worker 가 기다리지 않게」.
//   - D2 상태 행 일관: 투영의 레인 한 행은 한 잠금(Lane.Status)으로 읽힌다 — 찢긴 행(LATCHED 인데 revision 0)이 없다.
//   - C2 멈춘 원격 물결이 투영 Read 를 막지 않는다.
//   - C3 모든 진입 큐(시장 둘 · 레인 여덟)가 포화여도 안전 loop 셋은 제 cadence 로 돈다.
//
// 시간은 `testing/synctest` 거품의 가상 시계다 — 마감 시한(30 초)이 실제로 흐르지 않고, 모든 goroutine 이 멈춰 선 순간에만 앞으로 간다.
// 그래서 「경과 = 마감 시한 정확히 1 회」를 등식으로 잴 수 있다(순차였다면 정확히 2 회).

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// D2: 레인이 잠기는 동안 투영을 읽어도 한 행은 찢기지 않는다. 동시성이라 찢김을 결정적으로 만들 수 없으므로(판정: 비결정 CAUGHT 를 꾸미지
// 않음) 이 시험은 불변식을 반복해서 재고, 결정적 핀은 아래 AST 시험이 진다.
func TestAProjectedLaneRowIsNeverTorn(t *testing.T) {
	for round := 0; round < 200; round++ {
		runtime := newStrategyLaneRuntime(clock.System(), nil, "")
		store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)))
		if err != nil {
			t.Fatal(err)
		}
		c := &Context{strategyProjection: store, strategyLanes: runtime}
		var faults sync.WaitGroup
		for _, lane := range runtime.lanes {
			faults.Add(1)
			go func(lane *strategyworker.Lane) {
				defer faults.Done()
				lane.Fail("a112 torn-row probe", false)
			}(lane)
		}
		for read := 0; read < 4; read++ {
			snapshot, err := c.Read(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			for _, lane := range snapshot.Lanes {
				if lane.Health == nil {
					t.Fatalf("lane %s unobserved with a runtime", lane.LaneID)
				}
				switch *lane.Health {
				case strategyprojection.LaneLatched:
					if lane.LatchRevision < 1 || lane.FirstFailure == nil || lane.ConsecutiveFailures < 1 {
						t.Fatalf("torn LATCHED row %s: revision=%d first=%v failures=%d", lane.LaneID, lane.LatchRevision, lane.FirstFailure,
							lane.ConsecutiveFailures)
					}
				case strategyprojection.LaneDegraded:
					if lane.ConsecutiveFailures < 1 {
						t.Fatalf("torn DEGRADED row %s: failures=%d", lane.LaneID, lane.ConsecutiveFailures)
					}
				case strategyprojection.LaneHealthy:
					if lane.ConsecutiveFailures != 0 || lane.LatchRevision != 0 || lane.FirstFailure != nil {
						t.Fatalf("torn HEALTHY row %s: failures=%d revision=%d first=%v", lane.LaneID, lane.ConsecutiveFailures,
							lane.LatchRevision, lane.FirstFailure)
					}
				}
			}
		}
		faults.Wait()
	}
}

// D2 결정적 핀: 투영의 레인 행 함수는 상태를 Lane.Status() 한 번으로만 읽는다 — 상태 접근자를 따로 부르면(잠금이 여럿) 행이 찢길 수 있다.
// worker 값(Key · Policy · Horizon · Runtime)은 레인 수명 동안 불변이라 따로 읽어도 된다.
func TestTheLaneProjectionReadsALaneRowUnderOneLock(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "strategy_lane_projection.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	stateful := map[string]bool{"Health": true, "ConsecutiveFailures": true, "LatchRevision": true, "FirstFailure": true, "FirstAbnormal": true,
		"Latched": true, "Pending": true, "Dropped": true, "Abandoned": true, "NextDue": true, "RestartNotBefore": true}
	calls, status := []string{}, 0
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Name.Name != "strategyLaneProjection" {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if receiver, ok := selector.X.(*ast.Ident); ok && receiver.Name == "lane" {
				if selector.Sel.Name == "Status" {
					status++
				}
				if stateful[selector.Sel.Name] {
					calls = append(calls, selector.Sel.Name)
				}
			}
			return true
		})
	}
	if status != 1 || len(calls) != 0 {
		t.Fatalf("strategyLaneProjection reads lane.Status() %d time(s) and per-field state accessors %v — want exactly one Status() and none",
			status, calls)
	}
}

// C2: 원격 권한 물결이 멈춰 있어도(지도자가 아직 발표하지 않음) 투영 Read 는 마지막 발행 스냅숏으로 즉시 돌아온다 — Read 는 물결을 기다리지
// 않는다(5.2.1 의 구조 셈은 「잠금 안 원격 0」을, 이 시험은 그 운용 결과를 잰다).
func TestAStalledRemoteWaveNeverBlocksTheProjectionRead(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(time.Now().UTC()))
		if err != nil {
			t.Fatal(err)
		}
		c := &Context{strategyProjection: store}
		started := time.Now()
		_, wave, leader := c.joinStrategyRefreshWave(started)
		if !leader || wave == nil {
			t.Fatal("arrangement: no wave leader")
		}
		done := make(chan error, 1)
		go func() {
			_, err := c.Read(context.Background())
			done <- err
		}()
		synctest.Wait()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatal("the projection Read is blocked behind an unfinished remote authority wave")
		}
		c.publishStrategyRefreshWave(wave, started, StrategyEntryProductionAssembly{}, errors.New("a112 test wave released"))
	})
}

// C3: 진입 큐 전부 포화 — 두 시장 감독자 큐(사이클 하나가 돌며 멈춤 + 대기 칸 FULL)와 여덟 레인 칸(FULL) — 에서도 안전 loop 셋(fill ·
// reconcile · exit 관측 자리)은 1 초 cadence 로 정확히 돈다. 안전 loop 는 별도 goroutine · 별도 context 이고 진입 경로와 잠금 · 큐를 나누지
// 않는다는 것을 포화 아래에서 잰다(스펙 「entry worker 장애는 safety lifecycle을 지연하지 않는다」).
func TestSafetyLoopsKeepTheirCadenceWhileEveryEntryQueueIsSaturated(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		blocked := func(context.Context) error { <-release; return nil }
		worker := func(market StrategyMarket) StrategyMarketWorker {
			return StrategyMarketWorker{Market: market, Effective: true, Cycle: blocked, AuthorityGeneration: 7,
				AuthorityExpiresAt: time.Now().Add(24 * time.Hour), EvidenceDigest: "sha256:" + strings.Repeat("e", 64), LatchRevision: 1}
		}
		supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Workers: []StrategyMarketWorker{worker(StrategyMarketKR),
			worker(StrategyMarketUS)}, QueueDepth: 1, Clock: clock.System()})
		if err != nil {
			t.Fatal(err)
		}
		lanes := newStrategyLaneRuntime(clock.System(), nil, "")
		const cadence = time.Second
		var ticks [3]atomic.Int64
		loops := []SupervisedLoop{supervisor.SupervisedLoop()}
		for index, name := range []string{"fill-detection", "reconcile", "exit-observation"} {
			loops = append(loops, SupervisedLoop{Name: name, Run: func(ctx context.Context) error {
				ticker := time.NewTicker(cadence)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ticker.C:
						ticks[index].Add(1)
					}
				}
			}})
		}
		runtime, err := NewRuntime(RuntimeOptions{Loops: loops})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- runtime.Run(ctx) }()
		<-supervisor.Ready()
		synctest.Wait()
		// 포화: 시장마다 사이클 하나가 돌며 멈췄고(첫 poll) 대기 칸 하나를 채우면 다음은 FULL. 레인마다 칸 하나를 채우면 다음은 FULL.
		for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
			for supervisor.Trigger(market) == StrategyTriggerEnqueued {
			}
			if got := supervisor.Trigger(market); got != StrategyTriggerFull {
				t.Fatalf("%s market queue=%s, want FULL", market, got)
			}
		}
		for _, lane := range lanes.lanes {
			for lane.Offer() == strategyworker.TriggerEnqueued {
			}
			if got := lane.Offer(); got != strategyworker.TriggerFull {
				t.Fatalf("lane %v queue=%s, want FULL", lane.Key(), got)
			}
		}
		for index := range ticks {
			ticks[index].Store(0)
		}
		time.Sleep(10*cadence + cadence/2)
		for index, name := range []string{"fill-detection", "reconcile", "exit-observation"} {
			if got := ticks[index].Load(); got != 10 {
				t.Fatalf("%s ran %d cycles in 10.5 cadences under full entry saturation, want exactly 10", name, got)
			}
		}
		cancel()
		close(release)
		if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("runtime stop=%v", err)
		}
	})
}
