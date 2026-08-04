package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// authorizationOnlyGuardian deliberately does not implement
// engine.ReductionIssuer. It proves that engineRuntime stops at the exit-loop
// construction boundary and does not assemble a partial runtime.
type authorizationOnlyGuardian struct{}

var runtimeBranchNow = time.Date(2026, 8, 1, 9, 30, 0, 0, time.UTC)

func (authorizationOnlyGuardian) Authorize(context.Context, execgw.AuthorizationRequest) (execgw.GuardianDecision, error) {
	return execgw.GuardianDecision{}, errors.New("test guardian has no authority")
}

// runtimeBranchContext supplies only constructor-safe handles. No loop is run,
// no broker method is called, and the typed nil official client exists solely
// to satisfy the read-only interfaces checked by the constructors.
func runtimeBranchContext() *engine.Context {
	return &engine.Context{
		Official:   (*official.Client)(nil),
		AccountRef: "acct-runtime-branches",
		Journal:    &journal.Journal{},
		Gateway:    &execgw.Gateway{},
		Entry:      execgw.NewEntryGate(clock.NewFake(runtimeBranchNow), map[execgw.RequiredQuery]time.Duration{}),
		Resolver:   &execgw.Resolver{},
		Reconcile:  &reconcile.Tracker{},
		Retrier:    &execgw.Retrier{},
		Ingest:     &reconcile.Ingestor{},
		Converge:   &reconcile.Converger{},
		Automation: engine.AutomationStatus{Enabled: true, Verified: true},
		Guardian:   &execgw.RiskGuardian{},
	}
}

func TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess(t *testing.T) {
	clk := clock.NewFake(runtimeBranchNow)
	tests := []struct {
		name   string
		mutate func(*engine.Context)
		want   error
	}{
		{
			name: "B2 reconcile construction",
			mutate: func(value *engine.Context) {
				value.Reconcile = nil
			},
			want: engine.ErrReconcileDriverUnavailable,
		},
		{
			name: "B3 exit observer construction",
			mutate: func(value *engine.Context) {
				value.Guardian = authorizationOnlyGuardian{}
			},
			want: engine.ErrExitObserverUnavailable,
		},
		{
			name: "B4 recovery construction",
			mutate: func(value *engine.Context) {
				value.Resolver = nil
			},
			want: engine.ErrRecoveryUnavailable,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value := runtimeBranchContext()
			tc.mutate(value)
			runtime, err := engineRuntime(context.Background(), value, clk, nil)
			if runtime != nil || !errors.Is(err, tc.want) {
				t.Fatalf("runtime=%v err=%v want=%v", runtime, err, tc.want)
			}
		})
	}

	value := runtimeBranchContext()
	runtime, err := engineRuntime(context.Background(), value, clk, nil)
	if err != nil || runtime == nil {
		t.Fatalf("success runtime=%v err=%v", runtime, err)
	}
	// Construction arms recovery's entry latch but performs no broker read or
	// mutation. The loop set itself is pinned structurally by the existing
	// TestTheLoopSetIsTheSpecifiedThree regression.
	if refusal := value.Entry.CheckEntry(); refusal == nil || refusal.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("recovery did not fail closed at construction: %+v", refusal)
	}
}

func TestEngineRuntimeB1IsStructurallyUnreachableWithTheHardcodedNilHintPath(t *testing.T) {
	source := readSource(t, "engine.go")
	if !strings.Contains(source, "engineFillDetector(ectx, clk, nil)") {
		t.Fatal("engineRuntime no longer fixes the production hint input to nil; re-audit AST B1")
	}
	if _, err := engineFillDetector(runtimeBranchContext(), clock.NewFake(runtimeBranchNow), nil); err != nil {
		t.Fatalf("the hardcoded nil-hint detector path unexpectedly failed: %v", err)
	}
	// Keep imports bound to the same read-only surface used by production. This
	// compile-time assertion cannot invoke a network operation.
	var _ engine.OfficialReads = (*official.Client)(nil)
}
