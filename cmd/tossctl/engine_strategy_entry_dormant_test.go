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
	supervisor, err := engine.NewPairedStrategyEntryProductionAssembly(engine.StrategyEntryProductionSnapshot{Clock: clock.NewFake(runtimeBranchNow)})
	if err != nil {
		t.Fatalf("NewPairedStrategyEntryProductionAssembly: %v", err)
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
	runtime, err := engineRuntime(context.Background(), runtimeBranchContext(), clock.NewFake(runtimeBranchNow), nil)
	if err != nil {
		t.Fatalf("engineRuntime: %v", err)
	}
	want := []string{"reconcile", "exit", "filldetect", engine.StrategyEntryLoopName}
	if got := runtime.LoopNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("production loop names=%v want=%v", got, want)
	}
}

func TestProductionRuntimeUsesNonBlockingContextOwnedPairedRefreshSupervisor(t *testing.T) {
	source := readSource(t, "engine.go")
	if !strings.Contains(source, "ectx.NewRefreshingPairedStrategyEntrySupervisor(clk)") {
		t.Fatal("engineRuntime does not delegate the same-wave KR/US refresh bootstrap to the engine context")
	}
	if strings.Contains(source, "ectx.NewPairedStrategyEntryProductionAssembly(ctx, clk)") {
		t.Fatal("engineRuntime still blocks safety-loop construction on paired authority collection")
	}
}

func TestProductionStrategyEntrySnapshotHasNoAuthorityOrMutationCapability(t *testing.T) {
	typ := reflect.TypeOf(engine.StrategyEntryProductionSnapshot{})
	if typ.NumField() != 5 {
		t.Fatalf("production snapshot fields=%d, want exact five read-only observations", typ.NumField())
	}
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		identity := field.Name + " " + field.Type.String()
		for _, forbidden := range []string{"Gateway", "Journal", "Activation", "Trigger", "Cycle", "Broker", "Trading", "Guardian", "Writer", "Order"} {
			if strings.Contains(identity, forbidden) {
				t.Fatalf("production snapshot gained forbidden capability %q in %s", forbidden, identity)
			}
		}
	}
}

func TestProductionStrategyEntryAssemblyResultHasNoAuthorityOrMutationCapability(t *testing.T) {
	typ := reflect.TypeOf(engine.StrategyEntryProductionAssembly{})
	routeField, ok := typ.FieldByName("Route")
	if !ok || routeField.Type != reflect.TypeOf(engine.PairedStrategyRouteSnapshot{}) {
		t.Fatalf("production assembly route field=%+v ok=%v", routeField, ok)
	}
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		identity := field.Type.String()
		for _, forbidden := range []string{"scheduler.Activation", "scheduler.CalendarSnapshot", "Gateway", "Journal", "Guardian", "Broker", "Order"} {
			if strings.Contains(identity, forbidden) {
				t.Fatalf("production assembly result gained forbidden authority %q in %s", forbidden, identity)
			}
		}
	}
}

func TestProductionStrategyEntryAssemblyCannotPromoteFromAutomationFacts(t *testing.T) {
	supervisor, err := engine.NewPairedStrategyEntryProductionAssembly(engine.StrategyEntryProductionSnapshot{
		Clock: clock.NewFake(runtimeBranchNow), AutomationVerified: true, EntryPermitted: true, ProtectionWired: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, market := range []engine.StrategyMarket{engine.StrategyMarketKR, engine.StrategyMarketUS} {
		snapshot, ok := supervisor.Snapshot(market)
		if !ok || snapshot.Effective || snapshot.AuthorityGeneration != 0 || snapshot.EvidenceDigest != "" || snapshot.LatchRevision != 0 {
			t.Fatalf("market=%s automation facts promoted dormant worker: %+v ok=%v", market, snapshot, ok)
		}
	}
}
