package performance

import (
	"testing"
	"time"
)

func completeLineage() Lineage {
	return Lineage{
		CandidateLifeID:    "candidate-life-1",
		ThresholdVersion:   "threshold/v1",
		ThresholdSetDigest: "set-digest-1",
		EvidenceDigest:     "evidence-digest-1",
		LaneID:             "krx_parker_vwap_conservative_v1", LaneVersion: "lane/v1",
		DecisionID: "decision-1", AttemptID: "attempt-1", OrderID: "opaque-order-1",
		FillID: "fill-1", PositionID: "position-1", CloseID: "position-1",
		PolicyID: "COMMON_LADDER_BALANCED", PolicyVersion: "policy/v1",
	}
}

func measuredTrade(at time.Time) Trade {
	return Trade{
		ID: "trade-1", Lineage: completeLineage(), Market: "kr", Side: SideBuy,
		DecisionAt: at.Add(-time.Second), DecisionPrice: "99",
		EntryAt: at, EntryPrice: "100", Quantity: "10", CostTotal: "2",
		RealizedPnLAfterCosts: "80", RealizedR: "1.6", ClosedAt: at.Add(35 * time.Minute),
	}
}

func TestMeasureReusesInclusiveMarkoutToleranceAndKeepsMissingStates(t *testing.T) {
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	observations := []Observation{
		{ID: "before", PositionID: "position-1", At: at.Add(4*time.Minute + 59*time.Second), Price: "104", Source: "existing-position", SourceVersion: "v1"},
		{ID: "five", PositionID: "position-1", At: at.Add(5 * time.Minute), Price: "105", Source: "existing-position", SourceVersion: "v1"},
		{ID: "fifteen-boundary", PositionID: "position-1", At: at.Add(16 * time.Minute), Price: "110", Source: "existing-position", SourceVersion: "v1"},
		{ID: "thirty-too-late", PositionID: "position-1", At: at.Add(31*time.Minute + time.Nanosecond), Price: "95", Source: "existing-position", SourceVersion: "v1"},
		{ID: "lifetime-low", PositionID: "position-1", At: at.Add(20 * time.Minute), Price: "95", Source: "existing-position", SourceVersion: "v1"},
	}

	got := Measure(trade, observations, at.Add(time.Hour))
	if got.LineageStatus != StatusComplete {
		t.Fatalf("lineage status = %q, want complete", got.LineageStatus)
	}
	five := got.Markout(5)
	if five.Status != StatusComplete || five.GrossPct != "5" || five.CostAdjustedPct != "4.8" ||
		five.ObservationID != "five" || five.Source != "existing-position" {
		t.Fatalf("5m markout = %+v", five)
	}
	fifteen := got.Markout(15)
	if fifteen.Status != StatusComplete || fifteen.GrossPct != "10" || fifteen.ObservationID != "fifteen-boundary" {
		t.Fatalf("15m inclusive boundary = %+v", fifteen)
	}
	thirty := got.Markout(30)
	if thirty.Status != StatusNotMeasured || thirty.GrossPct != "" || thirty.CostAdjustedPct != "" {
		t.Fatalf("30m missing markout = %+v, want not_measured without synthetic zero", thirty)
	}
	if got.Slippage.Status != StatusComplete || got.Slippage.Value != "1.010101010101" {
		t.Fatalf("slippage = %+v", got.Slippage)
	}
	if got.Slippage.Source != "journal-decision-entry" || got.Slippage.SourceVersion != SemanticsVersion {
		t.Fatalf("slippage provenance = %+v", got.Slippage)
	}
	if got.MFE.Status != StatusComplete || got.MFE.Value != "10" ||
		got.MAE.Status != StatusComplete || got.MAE.Value != "-5" {
		t.Fatalf("MFE/MAE = %+v/%+v", got.MFE, got.MAE)
	}
	if got.MFE.Source != "existing-position" || got.MFE.ObservationID != "fifteen-boundary" ||
		got.MAE.Source != "existing-position" || got.MAE.ObservationID != "lifetime-low" {
		t.Fatalf("MFE/MAE provenance = %+v/%+v", got.MFE, got.MAE)
	}
}

func TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance(t *testing.T) {
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	trade.Side = SideSell
	trade.DecisionPrice = "101"
	observations := []Observation{
		{ID: "sell-5", PositionID: trade.Lineage.PositionID, At: at.Add(5 * time.Minute), Price: "95", Source: "position-cache", SourceVersion: "v7"},
		{ID: "sell-15", PositionID: trade.Lineage.PositionID, At: at.Add(15 * time.Minute), Price: "110", Source: "position-cache", SourceVersion: "v7"},
		{ID: "sell-30", PositionID: trade.Lineage.PositionID, At: at.Add(30 * time.Minute), Price: "90", Source: "position-cache", SourceVersion: "v7"},
	}

	got := Measure(trade, observations, at.Add(time.Hour))
	wants := map[int]struct{ gross, after, id string }{
		5:  {gross: "5", after: "4.8", id: "sell-5"},
		15: {gross: "-10", after: "-10.2", id: "sell-15"},
		30: {gross: "10", after: "9.8", id: "sell-30"},
	}
	for minutes, want := range wants {
		metric := got.Markout(minutes)
		if metric.Status != StatusComplete || metric.GrossPct != want.gross ||
			metric.CostAdjustedPct != want.after || metric.ObservationID != want.id ||
			metric.Source != "position-cache" || metric.SourceVersion != "v7" {
			t.Errorf("SELL %dm markout = %+v, want %+v", minutes, metric, want)
		}
	}
	if got.Slippage.Value != "0.990099009901" || got.Slippage.Source != "journal-decision-entry" {
		t.Fatalf("SELL slippage = %+v", got.Slippage)
	}
	if got.MFE.Value != "10" || got.MFE.ObservationID != "sell-30" || got.MFE.Source != "position-cache" ||
		got.MAE.Value != "-10" || got.MAE.ObservationID != "sell-15" || got.MAE.SourceVersion != "v7" {
		t.Fatalf("SELL MFE/MAE = %+v/%+v", got.MFE, got.MAE)
	}
}

func TestMeasureNeverUsesAnotherPositionOrInventsLineage(t *testing.T) {
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	trade.Lineage.FillID = ""
	got := Measure(trade, []Observation{{
		ID: "wrong-position", PositionID: "position-2", At: at.Add(5 * time.Minute),
		Price: "110", Source: "existing-position", SourceVersion: "v1",
	}}, at.Add(time.Hour))
	if got.LineageStatus != StatusLinkMissing {
		t.Fatalf("lineage status = %q, want link_missing", got.LineageStatus)
	}
	if got.Markout(5).Status != StatusNotMeasured || got.MFE.Status != StatusNotMeasured {
		t.Fatalf("foreign position observation was consumed: %+v", got)
	}
}

func TestTradeValidationRejectsInvalidStoredAmountsAndIdentity(t *testing.T) {
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*Trade)
	}{
		{name: "market missing", mutate: func(trade *Trade) { trade.Market = " " }},
		{name: "entry price zero", mutate: func(trade *Trade) { trade.EntryPrice = "0" }},
		{name: "quantity negative", mutate: func(trade *Trade) { trade.Quantity = "-1" }},
		{name: "cost negative", mutate: func(trade *Trade) { trade.CostTotal = "-0.01" }},
		{name: "decision price zero", mutate: func(trade *Trade) { trade.DecisionPrice = "0" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trade := measuredTrade(at)
			test.mutate(&trade)
			if err := trade.validate(); err == nil {
				t.Fatal("invalid trade passed validation")
			}
		})
	}
}

func TestExcursionsAreBoundedByTheZeroReturnAtEntry(t *testing.T) {
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	onlyFavorable := Measure(trade, []Observation{{
		ID: "up", PositionID: trade.Lineage.PositionID, At: at.Add(time.Minute),
		Price: "105", Source: "existing-position", SourceVersion: "v1",
	}}, at.Add(time.Hour))
	if onlyFavorable.MFE.Value != "5" || onlyFavorable.MAE.Value != "0" {
		t.Fatalf("only favorable MFE/MAE = %+v/%+v", onlyFavorable.MFE, onlyFavorable.MAE)
	}
	if onlyFavorable.MFE.ObservationID != "up" || onlyFavorable.MAE.Source != "journal-entry" {
		t.Fatalf("only favorable provenance = %+v/%+v", onlyFavorable.MFE, onlyFavorable.MAE)
	}

	onlyAdverse := Measure(trade, []Observation{{
		ID: "down", PositionID: trade.Lineage.PositionID, At: at.Add(time.Minute),
		Price: "95", Source: "existing-position", SourceVersion: "v1",
	}}, at.Add(time.Hour))
	if onlyAdverse.MFE.Value != "0" || onlyAdverse.MAE.Value != "-5" {
		t.Fatalf("only adverse MFE/MAE = %+v/%+v", onlyAdverse.MFE, onlyAdverse.MAE)
	}
}
