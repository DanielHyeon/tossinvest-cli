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
	campaignLineage       journal.PositionCampaignLineageRead
	campaign              journal.PositionCampaignRecord
}

func (f *fakeSource) ClosedStrategyTradeSources(
	_ context.Context, account string, after, to time.Time,
) ([]journal.ClosedStrategyTradeSource, error) {
	f.account, f.closedAfter, f.closedTo = account, after, to
	return append([]journal.ClosedStrategyTradeSource(nil), f.rows...), f.err
}

func (f *fakeSource) PositionCampaignLineage(_ context.Context, _ string) (journal.PositionCampaignLineageRead, error) {
	return f.campaignLineage, f.err
}

func (f *fakeSource) PositionCampaign(_ context.Context, _ string) (journal.PositionCampaignRecord, error) {
	return f.campaign, f.err
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
	if typ.NumMethod() != 2 || typ.Method(0).Name != "AttributionRebuild" || typ.Method(1).Name != "ClosedStrategyTrades" {
		t.Fatalf("adapter methods = %v, want two SELECT-only handoffs", typ.NumMethod())
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

func TestAttributionAdapterEnrichesOnlyExactCampaignFactsAndPreservesMissingEvidence(t *testing.T) {
	now := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	exact := &journal.ClosedStrategyLineage{
		CandidateLifeID: "candidate-1", LaneID: "lane", LaneVersion: "lane/v1",
		StrategyDecisionIdentity: "decision-1", StrategyAttemptID: "attempt-1",
		BrokerOrderID: "order-1", FillID: "fill-1", PositionID: "position-1", CloseOutcomeID: "position-1",
	}
	source := &fakeSource{
		rows: []journal.ClosedStrategyTradeSource{{TradeID: "position-1", PositionID: "position-1", CloseID: "position-1",
			Market: "us", Quantity: "2", RealizedPnLAfterCosts: "10", ClosedAt: now, PolicyID: "policy", PolicyVersion: "policy/v1", Lineage: exact}},
		campaignLineage: journal.PositionCampaignLineageRead{PositionID: "position-1", AccountRef: "acct-1", Market: "US", Symbol: "AAPL", PositionGeneration: 7, CampaignID: "campaign-1",
			Status: journal.PositionCampaignLineageKnown},
		campaign: journal.PositionCampaignRecord{ID: "campaign-1", AccountRef: "acct-1", Market: "US", Symbol: "AAPL", LaneID: "lane",
			LaneVersion: "lane/v1", DecisionID: "decision-1", ActualPositionGeneration: 7},
	}
	got, err := (&Reader{source: source}).AttributionRebuild(context.Background(), performance.AttributionEvidenceWindow{
		AccountRef: "acct-1", ClosedAfter: now.Add(-time.Hour), ClosedAtOrBefore: now,
	}, "build-1", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Positions) != 0 || len(got.FillDeltas) != 0 || len(got.Unavailable) != 1 {
		t.Fatalf("adapter invented reconstructable fill evidence: %+v", got)
	}
	row := got.Unavailable[0]
	if row.Key.Market != "US" || row.Key.Ticker != "AAPL" || row.Key.CampaignID != "campaign-1" ||
		row.Key.LegID != "" || row.LineageStatus != performance.StatusLinkMissing {
		t.Fatalf("available lineage mapping=%+v", row)
	}
	if row.Source.NetPnL.Value != "" || row.Reporting.NetPnL.Value != "" ||
		row.Source.NetPnL.Status != performance.StatusNotMeasured || row.Reporting.NetPnL.Status != performance.StatusNotMeasured {
		t.Fatalf("missing cost/FX was invented: source=%+v reporting=%+v", row.Source, row.Reporting)
	}
	for _, want := range []string{"leg_id", "close_leg_id"} {
		if !containsString(row.MissingLineage, want) {
			t.Fatalf("missing lineage=%v, want %s", row.MissingLineage, want)
		}
	}
	for _, want := range []string{"entry_fill_deltas", "close_fill_deltas", "fees", "taxes", "fx"} {
		if !containsString(row.MissingMeasurements, want) {
			t.Fatalf("missing measurements=%v, want %s", row.MissingMeasurements, want)
		}
	}
}

func TestAttributionAdapterRejectsKnownCampaignGenerationMismatch(t *testing.T) {
	now := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	source := &fakeSource{
		rows: []journal.ClosedStrategyTradeSource{{PositionID: "position-1", CloseID: "position-1", Market: "US", ClosedAt: now}},
		campaignLineage: journal.PositionCampaignLineageRead{PositionID: "position-1", AccountRef: "acct-1", Market: "US", Symbol: "AAPL",
			PositionGeneration: 8, CampaignID: "campaign-1", Status: journal.PositionCampaignLineageKnown},
		campaign: journal.PositionCampaignRecord{ID: "campaign-1", AccountRef: "acct-1", Market: "US", Symbol: "AAPL",
			ActualPositionGeneration: 7},
	}
	_, err := (&Reader{source: source}).AttributionRebuild(context.Background(), performance.AttributionEvidenceWindow{
		AccountRef: "acct-1", ClosedAfter: now.Add(-time.Hour), ClosedAtOrBefore: now,
	}, "build-1", now.Add(time.Minute))
	if !errors.Is(err, ErrCampaignLineageConflict) {
		t.Fatalf("generation mismatch err=%v", err)
	}
}

func TestAttributionAdapterNeverInventsCampaignForLegacyPosition(t *testing.T) {
	now := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	source := &fakeSource{rows: []journal.ClosedStrategyTradeSource{{TradeID: "legacy", PositionID: "legacy", CloseID: "legacy",
		Market: "kr", ClosedAt: now}}, campaignLineage: journal.PositionCampaignLineageRead{PositionID: "legacy",
		Status: journal.PositionCampaignLineageLegacyUnknown}}
	got, err := (&Reader{source: source}).AttributionRebuild(context.Background(), performance.AttributionEvidenceWindow{
		AccountRef: "acct-1", ClosedAfter: now.Add(-time.Hour), ClosedAtOrBefore: now,
	}, "build-1", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Unavailable) != 1 || got.Unavailable[0].Key.CampaignID != "" ||
		!containsString(got.Unavailable[0].MissingLineage, "campaign_id") {
		t.Fatalf("legacy campaign was inferred: %+v", got.Unavailable)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
