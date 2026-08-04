package engine_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestStrategyOnlyFXReadIsNotAGlobalEngineStartupDependency(t *testing.T) {
	want := map[string]int{
		"GET /api/v1/accounts":      0,
		"GET /api/v1/exchange-rate": 0,
	}
	for _, endpoint := range engine.RequiredEndpoints() {
		if _, ok := want[endpoint]; ok {
			want[endpoint]++
		}
	}
	if want["GET /api/v1/accounts"] != 1 {
		t.Errorf("global RequiredEndpoints account count=%d, want 1: %v", want["GET /api/v1/accounts"], engine.RequiredEndpoints())
	}
	if want["GET /api/v1/exchange-rate"] != 0 {
		t.Errorf("strategy-only exchange rate escaped into the global startup interlock: %v", engine.RequiredEndpoints())
	}
}

func TestUSFXReadFailureDoesNotCancelKRIdentityOrSafetyBudgets(t *testing.T) {
	var krIdentityChecks atomic.Int32
	var usFXAttempts atomic.Int32
	krCycles := make(chan struct{}, 2)
	usReadFailure := errors.New("fake official FX transport unavailable")

	retrier := &execgw.Retrier{Policy: execgw.RetryPolicy{MaxAttempts: 3, Budget: 8 * time.Second}}
	supervisor := mustStrategySupervisor(t, engine.StrategyEntrySupervisorOptions{Workers: []engine.StrategyMarketWorker{
		activeStrategyWorker(engine.StrategyMarketKR, func(context.Context) error {
			krIdentityChecks.Add(1)
			krCycles <- struct{}{}
			return nil
		}),
		activeStrategyWorker(engine.StrategyMarketUS, func(ctx context.Context) error {
			return retrier.Query(ctx, execgw.QueryExchangeRate, func(context.Context) error {
				usFXAttempts.Add(1)
				return usReadFailure
			})
		}),
	}})

	safetyStopped := make(chan struct{})
	runtime, err := engine.NewRuntime(engine.RuntimeOptions{Loops: []engine.SupervisedLoop{
		supervisor.SupervisedLoop(),
		{Name: "paired-fx-safety-proof", Run: func(ctx context.Context) error {
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
	waitClosed(t, supervisor.Ready(), "paired FX supervisor readiness")
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued ||
		supervisor.Trigger(engine.StrategyMarketUS) != engine.StrategyTriggerEnqueued {
		t.Fatal("paired FX triggers were not accepted")
	}
	waitClosed(t, krCycles, "KR identity evaluation")
	fault := waitStrategyFault(t, supervisor)
	if fault.Market != engine.StrategyMarketUS || fault.FirstRefusal != engine.StrategyWorkerRefusalFailure || fault.Reason != usReadFailure.Error() {
		t.Fatalf("US FX fault=%+v", fault)
	}
	if got := usFXAttempts.Load(); got != 3 {
		t.Fatalf("US FX attempts=%d, want bounded 3", got)
	}
	if supervisor.Trigger(engine.StrategyMarketKR) != engine.StrategyTriggerEnqueued {
		t.Fatal("US FX failure canceled KR identity evaluation")
	}
	waitClosed(t, krCycles, "KR peer evaluation after US FX failure")
	if got := krIdentityChecks.Load(); got != 2 {
		t.Fatalf("KR identity checks=%d, want two independent evaluations", got)
	}
	select {
	case err := <-done:
		t.Fatalf("US FX failure canceled runtime: %v", err)
	case <-safetyStopped:
		t.Fatal("US FX failure canceled safety budget")
	default:
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("graceful stop=%v", err)
	}
}
