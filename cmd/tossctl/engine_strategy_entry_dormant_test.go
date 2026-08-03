package main

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestProductionStrategyEntryAssemblyIsPairedDormantAndInert(t *testing.T) {
	supervisor, err := engineDormantStrategyEntry()
	if err != nil {
		t.Fatalf("engineDormantStrategyEntry: %v", err)
	}
	for _, market := range []engine.StrategyMarket{engine.StrategyMarketKR, engine.StrategyMarketUS} {
		snapshot, ok := supervisor.Snapshot(market)
		if !ok || snapshot.Effective || snapshot.Latched || snapshot.AuthorityGeneration != 0 ||
			snapshot.EvidenceDigest != "" || snapshot.LatchRevision != 0 || snapshot.QueueDepth != 0 {
			t.Fatalf("market=%s dormant snapshot=%+v ok=%v", market, snapshot, ok)
		}
		if got := supervisor.Trigger(market); got != engine.StrategyTriggerDisabled {
			t.Fatalf("market=%s pre-run trigger=%s", market, got)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- supervisor.Run(ctx) }()
	select {
	case <-supervisor.Ready():
	case <-time.After(time.Second):
		t.Fatal("dormant strategy-entry outer loop did not start")
	}
	for _, market := range []engine.StrategyMarket{engine.StrategyMarketKR, engine.StrategyMarketUS} {
		if got := supervisor.Trigger(market); got != engine.StrategyTriggerDisabled {
			t.Fatalf("market=%s running trigger=%s", market, got)
		}
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("dormant strategy-entry drain=%v", err)
	}
}

func TestProductionRuntimeIncludesOneDormantStrategyEntryOuterLoop(t *testing.T) {
	runtime, err := engineRuntime(runtimeBranchContext(), clock.NewFake(runtimeBranchNow), nil)
	if err != nil {
		t.Fatalf("engineRuntime: %v", err)
	}
	want := []string{"reconcile", "exit", "filldetect", engine.StrategyEntryLoopName}
	if got := runtime.LoopNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("production loop names=%v want=%v", got, want)
	}
}

func TestDormantProductionHelperHasNoAuthorityOrMutationInput(t *testing.T) {
	if typ := reflect.TypeOf(engineDormantStrategyEntry); typ.NumIn() != 0 {
		t.Fatalf("dormant production helper accepts %d inputs", typ.NumIn())
	}
	source := readSource(t, "engine.go")
	start := strings.Index(source, "func engineDormantStrategyEntry()")
	if start < 0 {
		t.Fatal("dormant strategy-entry helper is missing")
	}
	body := source[start:]
	if end := strings.Index(body, "\n}"); end >= 0 {
		body = body[:end+2]
	}
	if !strings.Contains(body, "engine.NewDormantStrategyEntrySupervisor()") {
		t.Fatalf("dormant helper body=%q", body)
	}
	for _, forbidden := range []string{"Gateway", "Journal", "Activation", "Trigger", "Cycle", "Live", "LIVE", "Enabled", "Automation"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("dormant helper gained forbidden authority %q: %s", forbidden, body)
		}
	}
}
