package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

type httpAPIHoldingsFixture struct {
	calls int
	rows  []domain.Position
}

func (f *httpAPIHoldingsFixture) Holdings(context.Context, string) ([]domain.Position, error) {
	f.calls++
	return append([]domain.Position(nil), f.rows...), nil
}

type httpAPIOrdersFixture struct {
	calls   int
	reading console.OrdersReading
}

type httpAPIManagementRuntimeFixture struct {
	runtime positionpolicy.ManagementRuntime
	err     error
}

func (f httpAPIManagementRuntimeFixture) Runtime(context.Context) (positionpolicy.ManagementRuntime, error) {
	return f.runtime, f.err
}

func (f *httpAPIOrdersFixture) Orders(context.Context) (console.OrdersReading, error) {
	f.calls++
	return f.reading, nil
}

func TestHTTPAPIProductionReaderMasksAccountAndCachesBrokerResources(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	holdings := &httpAPIHoldingsFixture{rows: []domain.Position{{
		Symbol: "005930", Name: "삼성전자", MarketCode: "KR", Quantity: 3, AveragePrice: 70000, CurrentPrice: 71000,
	}}}
	orders := &httpAPIOrdersFixture{reading: console.OrdersReading{
		AccountRef: "raw-account-1234", Open: []console.OrderRecord{{
			ID: "order-1", Symbol: "005930", Market: "KR", Side: "BUY", Status: "OPEN", Quantity: "1",
		}},
	}}
	reader := &httpAPIReader{
		now: func() time.Time { return now }, holdings: holdings, orders: orders,
		accountRef: func() (string, error) { return "raw-account-1234", nil },
	}

	for range 2 {
		positions, err := reader.Positions(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(positions.Items) != 1 || positions.Items[0].Symbol != "005930" ||
			positions.Items[0].ManagementStatus != "unknown" || strings.Contains(positions.Items[0].AccountLabel, "raw-account") {
			t.Fatalf("position projection=%+v", positions.Items)
		}
		projectedOrders, err := reader.Orders(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(projectedOrders.Items) != 1 || projectedOrders.Items[0].ID != "order-1" ||
			strings.Contains(projectedOrders.Items[0].AccountLabel, "raw-account") {
			t.Fatalf("order projection=%+v", projectedOrders.Items)
		}
	}
	if holdings.calls != 1 || orders.calls != 1 {
		t.Fatalf("broker calls holdings/orders=%d/%d want=1/1 within cache interval", holdings.calls, orders.calls)
	}
}

func TestStoredJournalPricesNeverMasqueradeAsAnEffectiveExitLine(t *testing.T) {
	projected := httpapi.Position{}
	applyStoredPosition(&projected, journal.PositionExit{
		HasExit: true,
		Exit:    journal.ExitState{EntryPrice: "70000", InitialStop: "66500", Baseline: "68000", HighWater: "73000"},
	}, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))

	if projected.ExitLine.Status != "unknown" || projected.ExitLine.CurrentProtection != "—" ||
		projected.ExitLine.NextTarget != "—" {
		t.Fatalf("raw journal prices became actionable exit line: %+v", projected.ExitLine)
	}
	if projected.StoredExitEvidence == nil || projected.StoredExitEvidence.EntryPrice != "70000" ||
		projected.StoredExitEvidence.InitialStop != "66500" || projected.StoredExitEvidence.Baseline != "68000" ||
		projected.StoredExitEvidence.HighWater != "73000" || projected.StoredExitEvidence.EffectiveKnown ||
		projected.StoredExitEvidence.Label != "원장 기록 · 실효 미확인" {
		t.Fatalf("stored evidence projection=%+v", projected.StoredExitEvidence)
	}
}

func TestHTTPAPIPositionProjectionUsesTheSharedManagementPriority(t *testing.T) {
	runtime := positionpolicy.ManagementRuntime{EffectiveKnown: true,
		Effective: positionpolicy.NewAdoptionSettings(false, 0.05, []string{"AAPL"}, []string{"AAPL"}, ""),
		Blocks: []positionpolicy.ReconcileBlock{positionpolicy.NewReconcileBlock(positionpolicy.ScopeAccount,
			"", "", "QUANTITY_MISMATCH", "raw detail is not public", time.Date(2026, 8, 1, 1, 0, 0, 0, time.UTC), true)}}

	managed := httpapi.Position{Market: "US", Symbol: "AAPL", InJournal: true, Eligible: true}
	applyManagementProjection(&managed, true, true, false, runtime)
	if managed.AdoptionStatus != "MANAGED" || !managed.StatusKnown || !managed.Excluded || managed.Candidate ||
		managed.CoveringBlock != nil {
		t.Fatalf("managed priority drift: %+v", managed)
	}

	blockedRuntime := runtime
	blockedRuntime.Effective = positionpolicy.NewAdoptionSettings(false, 0.05, []string{"AAPL"}, nil, "")
	blocked := httpapi.Position{Market: "US", Symbol: "AAPL"}
	applyManagementProjection(&blocked, true, false, false, blockedRuntime)
	if blocked.AdoptionStatus != "RECONCILE_BLOCKED" || blocked.AdoptionReason != "RECONCILE_BLOCK" ||
		!blocked.Candidate || blocked.Eligible || blocked.CoveringBlock == nil ||
		blocked.CoveringBlock.Scope != "ACCOUNT" || blocked.CoveringBlock.Reason != "QUANTITY_MISMATCH" {
		t.Fatalf("blocked projection drift: %+v", blocked)
	}
}

func TestHTTPAPIRuntimeFailureRemainsUnknownData(t *testing.T) {
	reader := &httpAPIReader{managementRuntime: httpAPIManagementRuntimeFixture{err: errors.New("engine unavailable")}}
	runtime := reader.readManagementRuntime(context.Background())
	if runtime.EffectiveKnown {
		t.Fatalf("runtime failure became known: %+v", runtime)
	}
	position := httpapi.Position{Market: "US", Symbol: "AAPL"}
	applyManagementProjection(&position, true, false, false, runtime)
	if position.AdoptionStatus != "UNKNOWN" || position.StatusKnown || position.DesignationKnown ||
		position.Included || position.Excluded || position.Candidate || position.AdoptionReason != "RUNTIME_UNAVAILABLE" {
		t.Fatalf("runtime failure projection=%+v", position)
	}
}

func TestHTTPAPIReleasedLifecycleSuppressesCanonicalExitAndWinsPriority(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	stored := journal.PositionExit{
		Position: journal.Position{ID: "pos-released", Market: "us", Symbol: "AAPL", AdoptionID: "adopt-1", Quantity: "1"},
		HasExit:  true,
		Exit: journal.ExitState{EntryPrice: "200", InitialStop: "190", Baseline: "190", HighWater: "210",
			Snapshot: journal.ExitSnapshotView{Snapshot: &journal.StoredExitSnapshot{
				Line: exitpolicy.ExitLineSnapshot{SnapshotID: "snapshot-released", DecisionID: "decision-released",
					EntryPrice: "200", InitialStop: "190", CurrentProtection: "195", NextTarget: "220"},
				ObservationSource: "official.quote", ObservedAt: now.Format(time.RFC3339Nano),
			}}},
	}
	out := httpapi.Position{Market: "US", Symbol: "AAPL"}
	applyStoredPosition(&out, stored, now)
	if out.ExitLine.CurrentProtection != "195" {
		t.Fatalf("precondition canonical exit=%+v", out.ExitLine)
	}
	state := positionpolicy.State{Status: positionpolicy.StatusReleased, Version: 1, ExitEligible: true}
	managed, released := lifecycleFlags(state, true)
	applyReleasedExitTruth(&out, stored)
	runtime := positionpolicy.ManagementRuntime{EffectiveKnown: true,
		Effective: positionpolicy.NewAdoptionSettings(true, .03, []string{"AAPL"}, nil, ""),
		Blocks:    []positionpolicy.ReconcileBlock{{Scope: positionpolicy.ScopeAccount, Reason: "QUANTITY_MISMATCH"}}}
	applyManagementProjection(&out, true, managed, released, runtime)
	if out.AdoptionStatus != "UNMANAGED" || out.AdoptionReason != "OPERATOR_RELEASED" || out.Eligible ||
		out.CoveringBlock != nil || out.ExitLine.CurrentProtection != "—" || out.ExitLine.NextTarget != "—" ||
		out.StoredExitEvidence == nil || out.StoredExitEvidence.EffectiveKnown {
		t.Fatalf("released projection=%+v", out)
	}
}

func TestHTTPAPIVirtualReleasedDefaultIsNotOperatorRelease(t *testing.T) {
	state := positionpolicy.State{Status: positionpolicy.StatusReleased, Version: 0}
	managed, released := lifecycleFlags(state, true)
	if managed || released {
		t.Fatalf("virtual lifecycle became durable release: managed=%v released=%v", managed, released)
	}
	out := httpapi.Position{Market: "US", Symbol: "AAPL"}
	runtime := positionpolicy.ManagementRuntime{EffectiveKnown: true,
		Effective: positionpolicy.NewAdoptionSettings(false, .05, nil, nil, "")}
	applyManagementProjection(&out, true, managed, released, runtime)
	if out.AdoptionStatus != "UNMANAGED" || out.AdoptionReason != "NOT_SELECTED" {
		t.Fatalf("virtual lifecycle projection=%+v", out)
	}
}
