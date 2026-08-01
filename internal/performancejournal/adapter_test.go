package performancejournal

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

type fakeSource struct {
	rows                  []journal.ClosedStrategyTradeSource
	err                   error
	account               string
	closedAfter, closedTo time.Time
}

func (f *fakeSource) ClosedStrategyTradeSources(
	_ context.Context, account string, after, to time.Time,
) ([]journal.ClosedStrategyTradeSource, error) {
	f.account, f.closedAfter, f.closedTo = account, after, to
	return append([]journal.ClosedStrategyTradeSource(nil), f.rows...), f.err
}

func TestAdapterMapsEveryAuthorityIdentifierAndNullableCostExactly(t *testing.T) {
	after := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := after.Add(time.Hour)
	cost := "2.25"
	exact := &journal.ClosedStrategyLineage{
		CandidateLifeID: "candidate-1", ThresholdVersion: "threshold/v1",
		ThresholdSetDigest: "threshold-digest", EvidenceDigest: "evidence-digest",
		LaneID: "parker", LaneVersion: "lane/v1",
		StrategyDecisionIdentity: "strategy-decision-1", RiskIntentID: "risk-intent-1",
		StrategyAttemptID: "strategy-attempt-1", MutationAttemptID: "mutation-attempt-1",
		BrokerOrderID: "broker-order-1", FillID: "fill-1", PositionID: "position-1", CloseOutcomeID: "position-1",
	}
	source := &fakeSource{rows: []journal.ClosedStrategyTradeSource{{
		TradeID: "position-1", PositionID: "position-1", CloseID: "position-1", Market: "kr", Side: "BUY",
		DecisionAt: after.Add(-time.Second), DecisionPrice: "99", EntryAt: after, EntryPrice: "100", Quantity: "10",
		CostTotal: &cost, RealizedPnLAfterCosts: "80", RealizedR: "1.6", ClosedAt: after.Add(35 * time.Minute),
		PolicyID: "COMMON_LADDER_BALANCED", PolicyVersion: "policy/v1", Lineage: exact,
	}, {
		TradeID: "legacy", PositionID: "legacy", CloseID: "legacy", Market: "kr", Side: "BUY",
		DecisionAt: after, DecisionPrice: "90", EntryAt: after.Add(time.Second), EntryPrice: "91", Quantity: "1",
		CostTotal: nil, RealizedPnLAfterCosts: "1", RealizedR: "0.1", ClosedAt: to,
	}}}

	got, err := (&Reader{source: source}).ClosedStrategyTrades(context.Background(), performance.ClosedTradeWindow{
		AccountRef: "acct-1", ClosedAfter: after, ClosedAtOrBefore: to,
	})
	if err != nil {
		t.Fatal(err)
	}
	if source.account != "acct-1" || !source.closedAfter.Equal(after) || !source.closedTo.Equal(to) {
		t.Fatalf("source scope = %q %s %s", source.account, source.closedAfter, source.closedTo)
	}
	wantLineage := performance.Lineage{
		CandidateLifeID: "candidate-1", ThresholdVersion: "threshold/v1",
		ThresholdSetDigest: "threshold-digest", EvidenceDigest: "evidence-digest",
		LaneID: "parker", LaneVersion: "lane/v1", DecisionID: "strategy-decision-1", RiskIntentID: "risk-intent-1",
		AttemptID: "strategy-attempt-1", MutationAttemptID: "mutation-attempt-1",
		OrderID: "broker-order-1", FillID: "fill-1", PositionID: "position-1", CloseID: "position-1",
		PolicyID: "COMMON_LADDER_BALANCED", PolicyVersion: "policy/v1",
	}
	if len(got) != 2 || !reflect.DeepEqual(got[0].Lineage, wantLineage) || got[0].CostTotal != cost ||
		got[0].Lineage.Status() != performance.StatusComplete {
		t.Fatalf("exact mapped trades = %+v", got)
	}
	if got[1].CostTotal != "" || got[1].Lineage.Status() != performance.StatusLinkMissing ||
		got[1].Lineage.PositionID != "legacy" {
		t.Fatalf("legacy mapping invented authority: %+v", got[1])
	}
}

func TestAdapterRequiresReadOnlySourceAndPropagatesErrors(t *testing.T) {
	window := performance.ClosedTradeWindow{
		AccountRef: "acct-1", ClosedAfter: time.Now().Add(-time.Hour), ClosedAtOrBefore: time.Now(),
	}
	if _, err := (*Reader)(nil).ClosedStrategyTrades(context.Background(), window); err == nil {
		t.Fatal("nil adapter was accepted")
	}
	want := errors.New("synthetic source failure")
	if _, err := (&Reader{source: &fakeSource{err: want}}).ClosedStrategyTrades(context.Background(), window); !errors.Is(err, want) {
		t.Fatalf("source error = %v", err)
	}
}

func TestAdapterCapabilitySurfaceIsExactlyOneSelectMethod(t *testing.T) {
	typ := reflect.TypeOf(&Reader{})
	if typ.NumMethod() != 1 || typ.Method(0).Name != "ClosedStrategyTrades" {
		t.Fatalf("adapter methods = %v, want one SELECT-only handoff", typ.NumMethod())
	}
	value := typ.Elem()
	if value.NumField() != 1 || value.Field(0).Name != "source" || value.Field(0).PkgPath == "" {
		t.Fatalf("adapter fields expose authority: %+v", value)
	}
	sourceType := reflect.TypeOf((*source)(nil)).Elem()
	if sourceType.NumMethod() != 1 || sourceType.Method(0).Name != "ClosedStrategyTradeSources" {
		t.Fatalf("source methods = %d", sourceType.NumMethod())
	}
}
