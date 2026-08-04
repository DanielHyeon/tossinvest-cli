package engine_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

const strategyEvidenceDigest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestStrategyEntrySupervisorDefaultsPairedMarketsDormant(t *testing.T) {
	supervisor, err := engine.NewDormantStrategyEntrySupervisor()
	if err != nil {
		t.Fatalf("NewDormantStrategyEntrySupervisor: %v", err)
	}
	if loop := supervisor.SupervisedLoop(); loop.Name != engine.StrategyEntryLoopName || loop.Run == nil || loop.Health != nil {
		t.Fatalf("outer loop = %+v", loop)
	}
	for _, market := range []engine.StrategyMarket{engine.StrategyMarketKR, engine.StrategyMarketUS} {
		snapshot, ok := supervisor.Snapshot(market)
		if !ok || snapshot.Market != market || snapshot.Effective || snapshot.Latched || snapshot.LatchID != "" || snapshot.AuthorityGeneration != 0 {
			t.Fatalf("market=%s dormant snapshot=%+v ok=%v", market, snapshot, ok)
		}
		if got := supervisor.Trigger(market); got != engine.StrategyTriggerDisabled {
			t.Fatalf("market=%s trigger=%s, want DISABLED", market, got)
		}
	}
}

func TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently(t *testing.T) {
	started := make(chan engine.StrategyMarket, 2)
	release := make(chan struct{})
	cycle := func(market engine.StrategyMarket) engine.StrategyCycle {
		return func(ctx context.Context) error {
			started <- market
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, cycle(engine.StrategyMarketKR)),
		activeStrategyWorker(engine.StrategyMarketUS, cycle(engine.StrategyMarketUS)),
	}})
	if got := supervisor.Trigger(engine.StrategyMarketKR); got != engine.StrategyTriggerDisabled {
		t.Fatalf("pre-run trigger=%s, want DISABLED", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued || supervisor.Trigger(engine.StrategyMarketUS) != engine.StrategyTriggerEnqueued {
		t.Fatal("paired triggers were not accepted")
	}
	seen := map[engine.StrategyMarket]bool{}
	for len(seen) < 2 {
		select {
		case market := <-started:
			seen[market] = true
		case <-time.After(time.Second):
			t.Fatalf("both markets did not enter concurrently: %v", seen)
		}
	}
	close(release)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v, want cancellation", err)
	}
}

func TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive(t *testing.T) {
	krFailure := errors.New("KR evidence unavailable")
	usCompleted := make(chan struct{})
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { return krFailure }),
		activeStrategyWorker(engine.StrategyMarketUS, func(context.Context) error { close(usCompleted); return nil }),
	}})
	safetyStopped := make(chan struct{})
	runtime, err := engine.NewRuntime(engine.RuntimeOptions{Loops: []engine.SupervisedLoop{
		supervisor.SupervisedLoop(),
		{Name: "safety-proof", Run: func(ctx context.Context) error {
			<-ctx.Done()
			close(safetyStopped)
			return ctx.Err()
		}},
	}})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runtime.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued || supervisor.Trigger(engine.StrategyMarketUS) != engine.StrategyTriggerEnqueued {
		t.Fatal("initial triggers were not accepted")
	}
	waitClosed(t, usCompleted, "US completion")
	fault := waitStrategyFault(t, supervisor)
	if fault.Market != engine.StrategyMarketKR || fault.LatchID != "strategy-latch:KR:7:12" ||
		fault.ExpectedRevision != 11 || fault.NextRevision != 12 || fault.AuthorityGeneration != 7 || fault.AuthorityExpiresAt.IsZero() ||
		fault.EvidenceDigest != strategyEvidenceDigest || fault.Reason != krFailure.Error() || fault.Abnormal || fault.ObservedAt.IsZero() {
		t.Fatalf("fault=%+v", fault)
	}
	kr, _ := supervisor.Snapshot(engine.StrategyMarketKR)
	us, _ := supervisor.Snapshot(engine.StrategyMarketUS)
	if !kr.Latched || kr.Effective || kr.LatchID != fault.LatchID || kr.LatchRevision != 12 {
		t.Fatalf("KR snapshot=%+v", kr)
	}
	if us.Latched || !us.Effective {
		t.Fatalf("US changed with KR failure: %+v", us)
	}
	if got := supervisor.Trigger(engine.StrategyMarketKR); got != engine.StrategyTriggerDisabled {
		t.Fatalf("latched KR trigger=%s", got)
	}
	select {
	case err := <-done:
		t.Fatalf("market failure stopped runtime: %v", err)
	case <-safetyStopped:
		t.Fatal("market failure stopped safety")
	case <-time.After(30 * time.Millisecond):
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("graceful stop=%v", err)
	}
}

func TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority(t *testing.T) {
	var krCalls atomic.Int32
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { krCalls.Add(1); return nil }),
		activeStrategyWorker(engine.StrategyMarketUS, func(context.Context) error { panic("US panic") }),
	}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketUS) != engine.StrategyTriggerEnqueued {
		t.Fatal("US trigger refused")
	}
	fault := waitStrategyFault(t, supervisor)
	if fault.Market != engine.StrategyMarketUS || !fault.Abnormal {
		t.Fatalf("panic fault=%+v", fault)
	}
	if supervisor.Trigger(engine.StrategyMarketUS) != engine.StrategyTriggerDisabled {
		t.Fatal("US recovered without a durable release receipt")
	}
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("KR disabled by US panic")
	}
	eventually(t, func() bool { return krCalls.Load() == 1 }, "KR after US panic")
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

func TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive(t *testing.T) {
	for _, failedMarket := range []engine.StrategyMarket{engine.StrategyMarketKR, engine.StrategyMarketUS} {
		failedMarket := failedMarket
		for _, abnormal := range []bool{false, true} {
			abnormal := abnormal
			name := string(failedMarket) + "/cycle-error"
			expectedRefusal := engine.StrategyWorkerRefusalFailure
			if abnormal {
				name = string(failedMarket) + "/panic"
				expectedRefusal = engine.StrategyWorkerRefusalAbnormal
			}
			t.Run(name, func(t *testing.T) {
				base := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
				fakeClock := clock.NewFake(base)
				peerMarket := engine.StrategyMarketUS
				if failedMarket == engine.StrategyMarketUS {
					peerMarket = engine.StrategyMarketKR
				}
				peerCycles := make(chan struct{}, 2)
				cycle := func(market engine.StrategyMarket) engine.StrategyCycle {
					return func(context.Context) error {
						if market == failedMarket {
							if abnormal {
								panic(string(market) + " worker returned abnormally")
							}
							return errors.New(string(market) + " worker returned abnormally")
						}
						peerCycles <- struct{}{}
						return nil
					}
				}
				supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{
					Clock: fakeClock,
					Workers: []engine.StrategyMarketWorker{
						activeStrategyWorker(engine.StrategyMarketKR, cycle(engine.StrategyMarketKR)),
						activeStrategyWorker(engine.StrategyMarketUS, cycle(engine.StrategyMarketUS)),
					},
				})

				safetyNames := []string{"fill-detection", "reconcile", "protection", "exit-observer", "emergency-reduction"}
				safetyStarted := make(chan string, len(safetyNames))
				safetyStopped := make(chan string, len(safetyNames))
				loops := []engine.SupervisedLoop{supervisor.SupervisedLoop()}
				for _, name := range safetyNames {
					name := name
					loops = append(loops, engine.SupervisedLoop{Name: name, Run: func(ctx context.Context) error {
						safetyStarted <- name
						<-ctx.Done()
						safetyStopped <- name
						return ctx.Err()
					}})
				}
				runtime, err := engine.NewRuntime(engine.RuntimeOptions{Loops: loops})
				if err != nil {
					t.Fatalf("NewRuntime: %v", err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				done := make(chan error, 1)
				go func() { done <- runtime.Run(ctx) }()
				waitClosed(t, supervisor.Ready(), "paired supervisor readiness")
				started := map[string]bool{}
				for len(started) < len(safetyNames) {
					select {
					case name := <-safetyStarted:
						started[name] = true
					case <-time.After(time.Second):
						t.Fatalf("safety loops did not all start: %v", started)
					}
				}
				if supervisor.Trigger(failedMarket) != engine.StrategyTriggerEnqueued || supervisor.Trigger(peerMarket) != engine.StrategyTriggerEnqueued {
					t.Fatal("paired market triggers were not accepted")
				}
				select {
				case <-peerCycles:
				case <-time.After(time.Second):
					t.Fatalf("%s peer cycle stopped by %s abnormal return", peerMarket, failedMarket)
				}
				fault := waitStrategyFault(t, supervisor)
				if fault.Market != failedMarket || fault.FirstRefusal != expectedRefusal || fault.Abnormal != abnormal || fault.RestartAttempt != 1 ||
					!fault.RestartNotBefore.Equal(base.Add(engine.DefaultStrategyRestartStep)) {
					t.Fatalf("market-local fault/restart=%+v", fault)
				}
				failed, _ := supervisor.Snapshot(failedMarket)
				peer, _ := supervisor.Snapshot(peerMarket)
				if !failed.Latched || failed.Effective || failed.FirstRefusal != expectedRefusal || failed.FirstAbnormal != abnormal || failed.RestartAttempt != 1 ||
					!failed.RestartNotBefore.Equal(base.Add(engine.DefaultStrategyRestartStep)) {
					t.Fatalf("failed market snapshot=%+v", failed)
				}
				if peer.Latched || !peer.Effective || peer.FirstRefusal != engine.StrategyWorkerRefusalNone || peer.RestartAttempt != 0 || !peer.RestartNotBefore.IsZero() {
					t.Fatalf("peer market contaminated=%+v", peer)
				}
				select {
				case err := <-done:
					t.Fatalf("market-local failure stopped runtime: %v", err)
				case stopped := <-safetyStopped:
					t.Fatalf("market-local failure stopped safety loop %s", stopped)
				default:
				}
				if supervisor.Trigger(failedMarket) != engine.StrategyTriggerDisabled {
					t.Fatal("bounded child restart restored entry without durable authority")
				}
				if supervisor.Trigger(peerMarket) != engine.StrategyTriggerEnqueued {
					t.Fatal("peer market stopped during local restart wait")
				}
				select {
				case <-peerCycles:
				case <-time.After(time.Second):
					t.Fatal("peer market did not continue during local restart wait")
				}
				cancel()
				if err := <-done; err != nil {
					t.Fatalf("graceful stop=%v", err)
				}
			})
		}
	}
}

func TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace(t *testing.T) {
	for _, failedMarket := range []engine.StrategyMarket{engine.StrategyMarketKR, engine.StrategyMarketUS} {
		failedMarket := failedMarket
		for _, abnormal := range []bool{false, true} {
			abnormal := abnormal
			name := string(failedMarket) + "/cycle-error"
			if abnormal {
				name = string(failedMarket) + "/panic"
			}
			t.Run(name, func(t *testing.T) {
				base := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
				raceClock := newRestartHandoffRaceClock(base, engine.DefaultStrategyRestartStep)
				cycle := func(context.Context) error {
					if abnormal {
						panic(string(failedMarket) + " restart handoff race")
					}
					return errors.New(string(failedMarket) + " restart handoff race")
				}
				workers := []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR}, {Market: engine.StrategyMarketUS}}
				if failedMarket == engine.StrategyMarketKR {
					workers[0] = activeStrategyWorker(failedMarket, cycle)
				} else {
					workers[1] = activeStrategyWorker(failedMarket, cycle)
				}
				supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Clock: raceClock, Workers: workers})
				ctx, cancel := context.WithCancel(context.Background())
				done := make(chan error, 1)
				go func() { done <- supervisor.Run(ctx) }()
				waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
				if supervisor.Trigger(failedMarket) != engine.StrategyTriggerEnqueued {
					t.Fatalf("%s trigger refused", failedMarket)
				}
				fault := waitStrategyFault(t, supervisor)
				if !fault.RestartNotBefore.Equal(base.Add(engine.DefaultStrategyRestartStep)) {
					t.Fatalf("published restart deadline=%s", fault.RestartNotBefore)
				}
				waitClosed(t, raceClock.advanced, "clock advance between fault handoff and restart wait")
				eventually(t, func() bool { return raceClock.fake.Sleepers() == 0 }, "absolute restart deadline completion")
				if supervisor.Trigger(failedMarket) != engine.StrategyTriggerDisabled {
					t.Fatal("elapsed restart deadline restored entry without durable authority")
				}
				select {
				case err := <-done:
					t.Fatalf("market-local restart race stopped supervisor: %v", err)
				default:
				}
				cancel()
				if err := <-done; !errors.Is(err, context.Canceled) {
					t.Fatalf("Run=%v", err)
				}
			})
		}
	}
}

func TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal(t *testing.T) {
	base := time.Date(9999, 12, 31, 23, 59, 58, 500000000, time.UTC)
	fakeClock := clock.NewFake(base)
	kr := activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { panic("later abnormal return") })
	kr.AuthorityExpiresAt = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	kr.FirstRefusal = engine.StrategyWorkerRefusalFailure
	kr.RestartAttempt = ^uint64(0)
	kr.RestartNotBefore = base
	us := activeStrategyWorker(engine.StrategyMarketUS, func(context.Context) error { return nil })
	us.AuthorityExpiresAt = kr.AuthorityExpiresAt
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Clock: fakeClock, Workers: []engine.StrategyMarketWorker{
		kr, us,
	}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("KR trigger refused")
	}
	fault := waitStrategyFault(t, supervisor)
	if fault.FirstRefusal != engine.StrategyWorkerRefusalFailure || fault.RestartAttempt != ^uint64(0) ||
		fault.RestartNotBefore.Before(base) || fault.RestartNotBefore.Sub(base) > engine.MaximumStrategyRestartBackoff {
		t.Fatalf("saturated fault=%+v", fault)
	}
	krSnapshot, _ := supervisor.Snapshot(engine.StrategyMarketKR)
	usSnapshot, _ := supervisor.Snapshot(engine.StrategyMarketUS)
	if krSnapshot.FirstRefusal != engine.StrategyWorkerRefusalFailure || krSnapshot.RestartAttempt != ^uint64(0) ||
		krSnapshot.RestartNotBefore.Before(base) || krSnapshot.RestartNotBefore.Sub(base) > engine.MaximumStrategyRestartBackoff {
		t.Fatalf("saturated snapshot=%+v", krSnapshot)
	}
	if usSnapshot.Latched || !usSnapshot.Effective || usSnapshot.RestartAttempt != 0 {
		t.Fatalf("saturated KR restart contaminated US=%+v", usSnapshot)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

func TestMarketQueueSaturationDoesNotConsumePeerQueue(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { close(started); <-release; return nil }),
		activeStrategyWorker(engine.StrategyMarketUS, func(context.Context) error { return nil }),
	}, QueueDepth: 1})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("first KR trigger refused")
	}
	waitClosed(t, started, "first KR callback")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("bounded KR queue did not accept its one waiting cycle")
	}
	if got := supervisor.Trigger(engine.StrategyMarketKR); got != engine.StrategyTriggerFull {
		t.Fatalf("saturated KR trigger=%s, want FULL", got)
	}
	if got := supervisor.Trigger(engine.StrategyMarketUS); got != engine.StrategyTriggerEnqueued {
		t.Fatalf("US trigger after KR saturation=%s", got)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
	close(release)
}

func TestShutdownAndTriggerShareBarrierAndDrainBothQueues(t *testing.T) {
	block := make(chan struct{})
	started := make(chan engine.StrategyMarket, 2)
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { started <- engine.StrategyMarketKR; <-block; return nil }),
		activeStrategyWorker(engine.StrategyMarketUS, func(context.Context) error { started <- engine.StrategyMarketUS; <-block; return nil }),
	}, QueueDepth: 8})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued || supervisor.Trigger(engine.StrategyMarketUS) != engine.StrategyTriggerEnqueued {
		t.Fatal("initial context-ignoring callbacks were not scheduled")
	}
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("context-ignoring callbacks did not start")
		}
	}

	start := make(chan struct{})
	var submitters sync.WaitGroup
	for i := 0; i < 64; i++ {
		submitters.Add(1)
		go func(i int) {
			defer submitters.Done()
			<-start
			market := engine.StrategyMarketKR
			if i%2 == 1 {
				market = engine.StrategyMarketUS
			}
			_ = supervisor.Trigger(market)
		}(i)
	}
	close(start)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("context-ignoring evaluations blocked shutdown")
	}
	submitters.Wait()
	for _, market := range []engine.StrategyMarket{engine.StrategyMarketKR, engine.StrategyMarketUS} {
		if got := supervisor.Trigger(market); got != engine.StrategyTriggerDisabled {
			t.Fatalf("post-shutdown market=%s trigger=%s", market, got)
		}
		snapshot, _ := supervisor.Snapshot(market)
		if snapshot.QueueDepth != 0 {
			t.Fatalf("post-shutdown market=%s depth=%d", market, snapshot.QueueDepth)
		}
		if !snapshot.AbandonedEvaluation {
			t.Fatalf("context-ignoring market=%s callback was not recorded as bounded abandoned work", market)
		}
	}
	close(block) // let abandoned evaluation-only callbacks exit; they hold no writer.
}

func TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction(t *testing.T) {
	fakeClock := clock.NewFake(time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC))
	release := make(chan struct{})
	var starts atomic.Int32
	var returns atomic.Int32
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error {
			starts.Add(1)
			<-release
			returns.Add(1)
			return engine.StrategyCentralIntegrityFailure(errors.New("late central result"))
		}),
		{Market: engine.StrategyMarketUS},
	}, Clock: fakeClock, CycleLimit: time.Second})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("KR trigger refused")
	}
	if !fakeClock.WaitForSleepers(1, time.Second) {
		t.Fatal("watchdog did not arm")
	}
	fakeClock.Advance(time.Second)
	fault := waitStrategyFault(t, supervisor)
	if fault.Market != engine.StrategyMarketKR || fault.Reason != engine.ErrStrategyCycleDeadline.Error() || !fault.Abnormal {
		t.Fatalf("deadline fault=%+v", fault)
	}
	if got := supervisor.Trigger(engine.StrategyMarketKR); got != engine.StrategyTriggerDisabled {
		t.Fatalf("watchdog-abandoned market accepted another cycle: %s", got)
	}
	close(release)
	eventually(t, func() bool { return returns.Load() == 1 }, "late callback return")
	select {
	case extra := <-supervisor.Faults():
		t.Fatalf("late result created a second action: %+v", extra)
	case err := <-done:
		t.Fatalf("late central result escaped watchdog: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	kr, _ := supervisor.Snapshot(engine.StrategyMarketKR)
	if !kr.Latched || kr.Effective || kr.LatchRevision != 12 || !kr.AbandonedEvaluation {
		t.Fatalf("late result changed latch: %+v", kr)
	}
	if starts.Load() != 1 {
		t.Fatalf("detached evaluations=%d, want structural maximum one per market", starts.Load())
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

func TestExpiredAuthorityLatchesBeforeEvaluation(t *testing.T) {
	base := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	fakeClock := clock.NewFake(base)
	var calls atomic.Int32
	kr := activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { calls.Add(1); return nil })
	kr.AuthorityExpiresAt = base.Add(time.Second)
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		kr, {Market: engine.StrategyMarketUS},
	}, Clock: fakeClock})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	fakeClock.Advance(time.Second)
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("expired authority trigger was not handed to the child for fail-closed latching")
	}
	fault := waitStrategyFault(t, supervisor)
	if fault.Reason != engine.ErrStrategyAuthorityExpired.Error() || fault.AuthorityExpiresAt != base.Add(time.Second) || calls.Load() != 0 {
		t.Fatalf("expired authority fault=%+v calls=%d", fault, calls.Load())
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

func TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety(t *testing.T) {
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error {
			return engine.StrategyCentralIntegrityFailure(errors.New("owner fence CAS failed"))
		}),
		{Market: engine.StrategyMarketUS},
	}})
	safetyStopped := make(chan struct{})
	runtime, err := engine.NewRuntime(engine.RuntimeOptions{Loops: []engine.SupervisedLoop{
		supervisor.SupervisedLoop(),
		{Name: "safety-proof", Run: func(ctx context.Context) error { <-ctx.Done(); close(safetyStopped); return ctx.Err() }},
	}})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- runtime.Run(context.Background()) }()
	waitClosed(t, supervisor.Ready(), "strategy supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("KR trigger refused")
	}
	if err := <-done; !errors.Is(err, engine.ErrLoopFailed) {
		t.Fatalf("runtime=%v, want ErrLoopFailed", err)
	}
	waitClosed(t, safetyStopped, "safety drain")
}

func TestStrategySupervisorRejectsInvalidAssemblies(t *testing.T) {
	cycle := func(context.Context) error { return nil }
	cases := map[string]engine.StrategyEntrySupervisorOptions{
		"no workers":                  {},
		"duplicate KR":                {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR}, {Market: engine.StrategyMarketKR}}},
		"combined":                    {Workers: []engine.StrategyMarketWorker{{Market: "KR+US"}, {Market: engine.StrategyMarketUS}}},
		"active nil cycle":            {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR, Effective: true, AuthorityGeneration: 7, AuthorityExpiresAt: time.Now().Add(time.Hour), EvidenceDigest: strategyEvidenceDigest, LatchRevision: 11}, {Market: engine.StrategyMarketUS}}},
		"active incomplete authority": {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR, Effective: true, Cycle: cycle}, {Market: engine.StrategyMarketUS}}},
		"active expired authority":    {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR, Effective: true, Cycle: cycle, AuthorityGeneration: 7, AuthorityExpiresAt: time.Unix(1, 0), EvidenceDigest: strategyEvidenceDigest, LatchRevision: 11}, {Market: engine.StrategyMarketUS}}},
		"active malformed restart":    {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR, Effective: true, Cycle: cycle, AuthorityGeneration: 7, AuthorityExpiresAt: time.Now().Add(time.Hour), EvidenceDigest: strategyEvidenceDigest, LatchRevision: 11, RestartAttempt: 1}, {Market: engine.StrategyMarketUS}}},
		"dormant carries authority":   {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR, AuthorityGeneration: 7}, {Market: engine.StrategyMarketUS}}},
		"dormant carries restart":     {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR, RestartAttempt: 1, FirstRefusal: engine.StrategyWorkerRefusalFailure, RestartNotBefore: time.Now()}, {Market: engine.StrategyMarketUS}}},
		"queue too large":             {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR}, {Market: engine.StrategyMarketUS}}, QueueDepth: engine.MaximumStrategyQueueDepth + 1},
		"cycle limit too large":       {Workers: []engine.StrategyMarketWorker{{Market: engine.StrategyMarketKR}, {Market: engine.StrategyMarketUS}}, CycleLimit: engine.MaximumStrategyCycleLimit + time.Nanosecond},
	}
	for name, opts := range cases {
		t.Run(name, func(t *testing.T) {
			if supervisor, err := engine.NewStrategyEntrySupervisor(opts); err == nil || supervisor != nil {
				t.Fatalf("supervisor=%v err=%v", supervisor, err)
			}
		})
	}
}

func TestSupervisorHasNoBooleanRecoveryOrDurableMutationCallbackSurface(t *testing.T) {
	options := reflect.TypeOf(engine.StrategyEntrySupervisorOptions{})
	for _, forbidden := range []string{"Recover", "Latch", "Release"} {
		if field, ok := options.FieldByName(forbidden); ok {
			t.Fatalf("options expose forbidden durable mutation callback %s (%s)", forbidden, field.Type)
		}
	}
	typeOfSupervisor := reflect.TypeOf((*engine.StrategyEntrySupervisor)(nil))
	for _, forbidden := range []string{"Recover", "Release", "Activate", "Enable"} {
		if method, ok := typeOfSupervisor.MethodByName(forbidden); ok {
			t.Fatalf("supervisor exposes in-memory authority transition %s (%s)", forbidden, method.Type)
		}
	}
}

func activeStrategyWorker(market engine.StrategyMarket, cycle engine.StrategyCycle) engine.StrategyMarketWorker {
	return engine.StrategyMarketWorker{Market: market, Effective: true, Cycle: cycle, AuthorityGeneration: 7,
		AuthorityExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), EvidenceDigest: strategyEvidenceDigest, LatchRevision: 11}
}

// restartHandoffRaceClock deterministically advances at the wait-side Now call:
// constructor, evaluation and latch observation consume the first three calls.
// An implementation that sleeps the originally calculated duration never makes
// the fourth call and therefore fails the regression instead of oversleeping.
type restartHandoffRaceClock struct {
	fake      *clock.Fake
	step      time.Duration
	mu        sync.Mutex
	nowCalls  int
	advanced  chan struct{}
	advanceAt int
}

func newRestartHandoffRaceClock(now time.Time, step time.Duration) *restartHandoffRaceClock {
	return &restartHandoffRaceClock{fake: clock.NewFake(now), step: step, advanced: make(chan struct{}), advanceAt: 4}
}

func (c *restartHandoffRaceClock) Now() time.Time {
	c.mu.Lock()
	c.nowCalls++
	advance := c.nowCalls == c.advanceAt
	c.mu.Unlock()
	if advance {
		c.fake.Advance(c.step)
		close(c.advanced)
	}
	return c.fake.Now()
}

func (c *restartHandoffRaceClock) Since(t time.Time) time.Duration { return c.fake.Since(t) }

func (c *restartHandoffRaceClock) Sleep(ctx context.Context, d time.Duration) error {
	return c.fake.Sleep(ctx, d)
}

func mustStrategySupervisor(t *testing.T, opts engine.StrategyEntrySupervisorOptions) *engine.StrategyEntrySupervisor {
	t.Helper()
	supervisor, err := engine.NewStrategyEntrySupervisor(opts)
	if err != nil {
		t.Fatalf("NewStrategyEntrySupervisor: %v", err)
	}
	return supervisor
}

func waitStrategyFault(t *testing.T, supervisor *engine.StrategyEntrySupervisor) engine.StrategyWorkerFault {
	t.Helper()
	select {
	case fault := <-supervisor.Faults():
		return fault
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for strategy fault")
		return engine.StrategyWorkerFault{}
	}
}

func waitClosed(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func eventually(t *testing.T, condition func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}
