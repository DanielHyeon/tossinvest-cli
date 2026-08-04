package engine_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
)

func TestStrategyEntrySupervisorPollsKRAndUSImmediatelyInTheSameWave(t *testing.T) {
	started := make(chan engine.StrategyMarket, 2)
	release := make(chan struct{})
	worker := func(market engine.StrategyMarket) engine.StrategyMarketWorker {
		descriptor := activeStrategyWorker(market, func(ctx context.Context) error {
			started <- market
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		descriptor.PollInterval = engine.MinimumStrategyPollInterval
		descriptor.RefreshesAuthority = true
		return descriptor
	}
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		worker(engine.StrategyMarketKR), worker(engine.StrategyMarketUS),
	}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "paired polling readiness")

	seen := map[engine.StrategyMarket]bool{}
	for len(seen) < 2 {
		select {
		case market := <-started:
			seen[market] = true
		case <-time.After(time.Second):
			t.Fatalf("paired pollers did not start in the same wave: %v", seen)
		}
	}
	close(release)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v, want cancellation", err)
	}
}

func TestStrategyEntrySupervisorZeroPollIntervalKeepsExplicitTriggerSemantics(t *testing.T) {
	var calls atomic.Int32
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { calls.Add(1); return nil }),
		{Market: engine.StrategyMarketUS},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "explicit-trigger readiness")
	time.Sleep(30 * time.Millisecond)
	if calls.Load() != 0 {
		t.Fatalf("zero-interval worker auto-polled %d times", calls.Load())
	}
	if got := supervisor.Trigger(engine.StrategyMarketKR); got != engine.StrategyTriggerEnqueued {
		t.Fatalf("explicit trigger=%s, want ENQUEUED", got)
	}
	eventually(t, func() bool { return calls.Load() == 1 }, "explicit strategy cycle")
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v, want cancellation", err)
	}
}

func TestStrategyEntrySupervisorPollerKeepsCompletePeerRunningWhenOtherMarketDormant(t *testing.T) {
	var usCalls atomic.Int32
	us := activeStrategyWorker(engine.StrategyMarketUS, func(context.Context) error { usCalls.Add(1); return nil })
	us.PollInterval = engine.MinimumStrategyPollInterval
	us.RefreshesAuthority = true
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		{Market: engine.StrategyMarketKR}, us,
	}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "isolated polling readiness")
	eventually(t, func() bool { return usCalls.Load() == 1 }, "US polling with dormant KR")
	if got := supervisor.Trigger(engine.StrategyMarketKR); got != engine.StrategyTriggerDisabled {
		t.Fatalf("dormant KR trigger=%s, want DISABLED", got)
	}
	kr, _ := supervisor.Snapshot(engine.StrategyMarketKR)
	usSnapshot, _ := supervisor.Snapshot(engine.StrategyMarketUS)
	if kr.Effective || !usSnapshot.Effective || usSnapshot.Latched {
		t.Fatalf("market isolation KR=%+v US=%+v", kr, usSnapshot)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v, want cancellation", err)
	}
}

func TestStrategyEntrySupervisorPollsDormantRefreshWorkerWithoutOpeningPublicTrigger(t *testing.T) {
	called := make(chan struct{}, 1)
	kr := engine.StrategyMarketWorker{Market: engine.StrategyMarketKR, PollInterval: engine.MinimumStrategyPollInterval,
		RefreshesAuthority: true, Cycle: func(context.Context) error { called <- struct{}{}; return nil }}
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		kr, {Market: engine.StrategyMarketUS},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	waitClosed(t, supervisor.Ready(), "dormant refresh readiness")
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("dormant authority worker was never refreshed")
	}
	if got := supervisor.Trigger(engine.StrategyMarketKR); got != engine.StrategyTriggerDisabled {
		t.Fatalf("public dormant trigger=%s, want DISABLED", got)
	}
	snapshot, _ := supervisor.Snapshot(engine.StrategyMarketKR)
	if snapshot.Effective || snapshot.AuthorityGeneration != 0 {
		t.Fatalf("refresh opened dormant authority: %+v", snapshot)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v", err)
	}
}

func TestStrategyEntrySupervisorRejectsUnsafeProductionPollIntervals(t *testing.T) {
	for _, interval := range []time.Duration{-time.Second, time.Millisecond, engine.MaximumStrategyCycleLimit + time.Second} {
		kr := activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error { return nil })
		kr.PollInterval = interval
		kr.RefreshesAuthority = true
		if supervisor, err := engine.NewStrategyEntrySupervisor(engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
			kr, {Market: engine.StrategyMarketUS},
		}}); err == nil || supervisor != nil {
			t.Fatalf("interval=%s supervisor=%v err=%v, want refusal", interval, supervisor, err)
		}
	}
}
