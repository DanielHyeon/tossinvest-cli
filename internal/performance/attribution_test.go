package performance

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

var attributionAt = time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

func TestDerivedAttributionUsesExactMarketCampaignLineageForSameTicker(t *testing.T) {
	krPosition, krEvents := completeFixture("KR", "005930", "kr-position")
	usPosition, usEvents := completeFixture("US", "005930", "us-position")
	store, err := BuildDerivedAttributionStore([]PositionEvidence{krPosition, usPosition}, append(krEvents, usEvents...))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.QueryAttribution(AttributionQuery{Ticker: "005930"}); !errors.Is(err, ErrMarketRequired) {
		t.Fatalf("ticker-only query error=%v", err)
	}
	rows, err := store.QueryAttribution(AttributionQuery{
		Market: "US", Ticker: "005930", LaneID: "lane-US", LaneVersion: "v3",
		CampaignID: "campaign-US", LegID: "leg-US",
	})
	if err != nil || len(rows) != 1 {
		t.Fatalf("US exact query rows=%+v err=%v", rows, err)
	}
	row := rows[0]
	if row.Key.Market != "US" || row.Key.LaneID != "lane-US" || row.Key.CampaignID != "campaign-US" ||
		row.Key.PositionID != "us-position" || row.LineageStatus != StatusComplete {
		t.Fatalf("cross-market attribution=%+v", row)
	}
	if row.Key.Ticker != "005930" {
		t.Fatalf("display ticker lost: %+v", row.Key)
	}
}

func TestMissingCompositeLineageStaysLinkMissingWithoutSymbolRepair(t *testing.T) {
	position, events := completeFixture("US", "SAME", "position-missing")
	events[1].Lineage.CampaignID = ""
	events[1].Lineage.DecisionID = ""
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := store.QueryAttribution(AttributionQuery{Market: "US", IncludeLinkMissing: true})
	if err != nil || len(rows) != 1 {
		t.Fatalf("missing rows=%+v err=%v", rows, err)
	}
	if rows[0].LineageStatus != StatusLinkMissing ||
		!reflect.DeepEqual(rows[0].MissingLineage, []string{"campaign_id", "decision_id"}) {
		t.Fatalf("missing lineage=%+v", rows[0])
	}
	complete, err := store.QueryAttribution(AttributionQuery{Market: "US"})
	if err != nil || len(complete) != 0 {
		t.Fatalf("link-missing row entered attributed sample: %+v err=%v", complete, err)
	}
}

func TestFillDeltaDeduplicatesExactEventAndFillIdentity(t *testing.T) {
	position, events := completeFixture("KR", "005930", "position-dedup")
	duplicateEvent := events[0]
	duplicateFill := events[0]
	duplicateFill.EventID = "delivery-retry"
	withDuplicates := append(append([]FillDelta{}, events...), duplicateEvent, duplicateFill)
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, withDuplicates)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "KR"})
	if len(rows) != 1 || rows[0].AcquiredQuantity != "10" || rows[0].ClosedQuantity != "4" || rows[0].ResidualQuantity != "6" {
		t.Fatalf("duplicate advanced totals: %+v", rows)
	}

	divergent := duplicateFill
	divergent.QuantityDelta = "11"
	if _, err := BuildDerivedAttributionStore([]PositionEvidence{position}, append(events, divergent)); !errors.Is(err, ErrDivergentFillReplay) {
		t.Fatalf("divergent fill replay error=%v", err)
	}
}

func TestExplicitCorrectionReversesOnlyReferencedFillAndComposite(t *testing.T) {
	position, events := completeFixture("KR", "SAME", "position-correction")
	position.AcquiredQuantity = "8"
	position.ResidualQuantity = "5"
	position.TotalEntryBasis = "800"
	position.ResidualEntryBasis = "500"
	events[1].QuantityDelta = "3"
	events[1].AllocatedEntryBasis = amount("300", "journal-cost")
	events[1].ExitProceeds = amount("360", "broker-fill")
	correction := entryDelta(events[0].Lineage, "entry-correction", "entry-correction-fill", "-2")
	correction.CorrectionOfFillID = events[0].Lineage.FillID
	events = append(events, correction)
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "KR"})
	if len(rows) != 1 || rows[0].AcquiredQuantity != "8" || rows[0].ClosedQuantity != "3" || rows[0].ResidualQuantity != "5" {
		t.Fatalf("correction totals=%+v", rows)
	}

	wrongMarket := correction
	wrongMarket.EventID = "wrong-market-correction"
	wrongMarket.Lineage.FillID = "wrong-market-correction-fill"
	wrongMarket.Lineage.Market = "US"
	if _, err := BuildDerivedAttributionStore([]PositionEvidence{position}, append(events[:2], wrongMarket)); !errors.Is(err, ErrCorrectionLineage) {
		t.Fatalf("cross-market correction error=%v", err)
	}
}

func TestPartialEntriesStagedClosesAndCostPnLConserve(t *testing.T) {
	position, events := stagedFixture()
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "US"})
	if len(rows) != 1 {
		t.Fatalf("rows=%+v", rows)
	}
	row := rows[0]
	if row.AcquiredQuantity != "10" || row.ClosedQuantity != "7" || row.ResidualQuantity != "3" || row.FullyClosed {
		t.Fatalf("quantity conservation=%+v", row)
	}
	if row.TotalEntryBasis != "1000" || row.AllocatedEntryBasis != "700" || row.ResidualEntryBasis != "300" {
		t.Fatalf("basis conservation=%+v", row)
	}
	if len(row.CloseLegs) != 2 || row.CloseLegs[0].CloseLegID != "close-leg-1" || row.CloseLegs[1].CloseLegID != "close-leg-2" {
		t.Fatalf("staged closes=%+v", row.CloseLegs)
	}
	assertCompleteAmount(t, row.Source.GrossPnL, "110", "USD")
	assertCompleteAmount(t, row.Source.EntryFees, "7", "USD")
	assertCompleteAmount(t, row.Source.ExitFees, "9", "USD")
	assertCompleteAmount(t, row.Source.Taxes, "2", "USD")
	assertCompleteAmount(t, row.Source.FXCost, "3", "USD")
	assertCompleteAmount(t, row.Source.NetPnL, "89", "USD")
	if row.Reporting.NetPnL.Status != StatusComplete || row.Reporting.Currency != "KRW" ||
		row.Reporting.FXSource != "official-close-fx" || row.Reporting.FXAsOf.IsZero() ||
		row.Reporting.RoundingVersion != "round-v2" {
		t.Fatalf("reporting provenance=%+v", row.Reporting)
	}
	assertReportingEquation(t, row.Reporting)
}

func TestMissingFeeOrFXIsNotMeasuredAndNeverZeroFilled(t *testing.T) {
	for _, test := range []struct {
		name        string
		mutate      func([]FillDelta)
		wantSource  Status
		wantReport  Status
		wantMissing string
	}{
		{"missing fee", func(events []FillDelta) { events[4].ExitFees = nil }, StatusNotMeasured, StatusNotMeasured, "exit_fees"},
		{"missing FX", func(events []FillDelta) { events[4].FX = nil }, StatusComplete, StatusNotMeasured, "fx"},
	} {
		t.Run(test.name, func(t *testing.T) {
			position, events := stagedFixture()
			test.mutate(events)
			store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
			if err != nil {
				t.Fatal(err)
			}
			rows, _ := store.QueryAttribution(AttributionQuery{Market: "US"})
			row := rows[0]
			if row.Source.NetPnL.Status != test.wantSource || row.Reporting.NetPnL.Status != test.wantReport {
				t.Fatalf("missing evidence became measured: source=%+v reporting=%+v", row.Source, row.Reporting)
			}
			if row.Source.NetPnL.Status == StatusNotMeasured && row.Source.NetPnL.Value != "" ||
				row.Reporting.NetPnL.Status == StatusNotMeasured && row.Reporting.NetPnL.Value != "" {
				t.Fatalf("missing evidence became numeric zero: source=%+v reporting=%+v", row.Source.NetPnL, row.Reporting.NetPnL)
			}
			if !contains(row.MissingMeasurements, test.wantMissing) {
				t.Fatalf("missing provenance=%v want=%s", row.MissingMeasurements, test.wantMissing)
			}
			assertCompleteAmount(t, row.Source.GrossPnL, "110", "USD")
		})
	}
}

func TestConservationAndAuthoritativePolicyMismatchFailClosed(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*PositionEvidence, []FillDelta)
		want   error
	}{
		{"quantity loss", func(position *PositionEvidence, _ []FillDelta) { position.ResidualQuantity = "2" }, ErrQuantityConservation},
		{"basis loss", func(position *PositionEvidence, _ []FillDelta) { position.ResidualEntryBasis = "299" }, ErrBasisConservation},
		{"policy mismatch", func(_ *PositionEvidence, events []FillDelta) { events[3].Lineage.PolicyVersion = "other" }, ErrPolicyMismatch},
		{"positive correction", func(_ *PositionEvidence, events []FillDelta) {
			events[0].CorrectionOfFillID = "prior-fill"
		}, ErrCorrectionDelta},
	} {
		t.Run(test.name, func(t *testing.T) {
			position, events := stagedFixture()
			test.mutate(&position, events)
			if _, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events); !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
		})
	}
}

func TestDerivedStoreDeepCopiesInputsAndExposesNoAuthorityMethods(t *testing.T) {
	position, events := completeFixture("KR", "005930", "position-copy")
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	events[0].QuantityDelta = "999"
	position.ResidualQuantity = "999"
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "KR"})
	if rows[0].AcquiredQuantity != "10" || rows[0].ResidualQuantity != "6" {
		t.Fatalf("store retained caller aliases: %+v", rows[0])
	}
	typeOf := reflect.TypeOf(store)
	for _, forbidden := range []string{"Place", "Order", "Activate", "SetToggle", "WriteJournal", "Protect", "CollectQuotes"} {
		if _, ok := typeOf.MethodByName(forbidden); ok {
			t.Fatalf("derived store exposes authority method %s", forbidden)
		}
	}
}

func TestAdversarialCompositeLineageCannotLaunderAcrossLaneOrLeg(t *testing.T) {
	position, events := completeFixture("US", "SAME", "position-scope")
	events[1].Lineage.LaneID = "other-lane"
	if _, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events); !errors.Is(err, ErrLineageConflict) {
		t.Fatalf("cross-lane evidence error=%v", err)
	}

	firstPosition, firstEvents := completeFixture("US", "SAME", "shared-position")
	secondPosition, secondEvents := completeFixture("US", "SAME", "shared-position")
	secondPosition.LegID, secondEvents[0].Lineage.LegID, secondEvents[1].Lineage.LegID = "leg-US-2", "leg-US-2", "leg-US-2"
	secondPosition.AcquiredQuantity, secondPosition.ResidualQuantity = "5", "3"
	secondPosition.TotalEntryBasis, secondPosition.ResidualEntryBasis = "500", "300"
	secondEvents[0].QuantityDelta = "5"
	secondEvents[1].QuantityDelta = "2"
	secondEvents[1].AllocatedEntryBasis = amount("200", "journal-cost")
	secondEvents[1].ExitProceeds = amount("240", "broker-fill")
	store, err := BuildDerivedAttributionStore([]PositionEvidence{firstPosition, secondPosition}, append(firstEvents, secondEvents...))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := store.QueryAttribution(AttributionQuery{Market: "US", CampaignID: "campaign-US", LegID: "leg-US-2"})
	if err != nil || len(rows) != 1 || rows[0].Key.LegID != "leg-US-2" || rows[0].AcquiredQuantity != "5" {
		t.Fatalf("leg-scoped query rows=%+v err=%v", rows, err)
	}
}

func TestCorrectionIsOrderIndependentExactAndCannotOverrunOriginal(t *testing.T) {
	position, events := completeFixture("KR", "SAME", "position-order-independent")
	position.AcquiredQuantity, position.ResidualQuantity = "8", "4"
	position.TotalEntryBasis, position.ResidualEntryBasis = "800", "400"
	correction := entryDelta(events[0].Lineage, "entry-correction", "entry-correction-fill", "-2")
	correction.CorrectionOfFillID = events[0].Lineage.FillID
	ordered := []FillDelta{correction, events[1], events[0]}
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, ordered)
	if err != nil {
		t.Fatalf("correction-before-original error=%v", err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "KR"})
	if rows[0].AcquiredQuantity != "8" || rows[0].ClosedQuantity != "4" {
		t.Fatalf("order-independent correction totals=%+v", rows[0])
	}

	overrun := correction
	overrun.EventID, overrun.Lineage.FillID, overrun.QuantityDelta = "overrun", "overrun-fill", "-11"
	if _, err := BuildDerivedAttributionStore([]PositionEvidence{position}, append(events, overrun)); !errors.Is(err, ErrCorrectionOverrun) {
		t.Fatalf("correction overrun error=%v", err)
	}

	wrongDecision := correction
	wrongDecision.EventID, wrongDecision.Lineage.FillID, wrongDecision.Lineage.DecisionID = "wrong-decision", "wrong-decision-fill", "other"
	if _, err := BuildDerivedAttributionStore([]PositionEvidence{position}, append(events, wrongDecision)); !errors.Is(err, ErrCorrectionLineage) {
		t.Fatalf("cross-decision correction error=%v", err)
	}
}

func TestDecisionToCloseAndCostEvidenceRemainVisibleInDerivedRows(t *testing.T) {
	position, events := stagedFixture()
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "US"})
	row := rows[0]
	if row.CostPolicy.ID != "fifo" || row.CostPolicy.Version != "policy-v4" || len(row.EntryFills) != 3 {
		t.Fatalf("authority provenance=%+v entry=%+v", row.CostPolicy, row.EntryFills)
	}
	for _, leg := range row.CloseLegs {
		lineage := leg.Lineage
		if lineage.DecisionID == "" || lineage.AttemptID == "" || lineage.OrderID == "" || lineage.FillID == "" ||
			lineage.PositionID == "" || lineage.CloseID == "" || lineage.CloseLegID == "" ||
			leg.EntryFeeEvidence == nil || leg.ExitFeeEvidence == nil || leg.TaxEvidence == nil ||
			leg.FXCostEvidence == nil || leg.FXEvidence == nil || leg.FXEvidence.AsOf.IsZero() {
			t.Fatalf("decision-to-close evidence incomplete: %+v", leg)
		}
	}
}

func TestSameCurrencyReportingStillRequiresPersistedFXIdentity(t *testing.T) {
	position, events := completeFixture("KR", "005930", "same-currency")
	position.Policy = policy("KRW", "KRW")
	for index := range events {
		events[index].Lineage.PolicyID = position.Policy.ID
		events[index].Lineage.PolicyVersion = position.Policy.Version
	}
	events[1].FX = nil
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "KR"})
	if rows[0].Source.NetPnL.Status != StatusComplete || rows[0].Reporting.NetPnL.Status != StatusNotMeasured ||
		!contains(rows[0].MissingMeasurements, "fx") {
		t.Fatalf("identity conversion was silently assumed: source=%+v reporting=%+v missing=%v",
			rows[0].Source, rows[0].Reporting, rows[0].MissingMeasurements)
	}
}

func TestMultipleAnonymousFillLinksRemainLinkMissingNotDeduplicatedByGuess(t *testing.T) {
	position, events := completeFixture("US", "SAME", "missing-fills")
	events[0].Lineage.FillID = ""
	events[1].Lineage.FillID = ""
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "US", IncludeLinkMissing: true})
	if len(rows) != 1 || rows[0].LineageStatus != StatusLinkMissing || !contains(rows[0].MissingLineage, "fill_id") {
		t.Fatalf("anonymous fills were guessed/deduplicated: %+v", rows)
	}
}

func TestOpenPositionWithoutCloseIsNotMeasuredNotZero(t *testing.T) {
	lineage := completeCompositeLineage("KR", "005930", "open-position")
	position := positionEvidence(lineage, "10", "10", "1000", "1000", policy("KRW", "KRW"))
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, []FillDelta{
		entryDelta(lineage, "open-entry", "open-entry-fill", "10"),
	})
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "KR"})
	if rows[0].Source.GrossPnL.Status != StatusNotMeasured || rows[0].Source.GrossPnL.Value != "" ||
		rows[0].Source.NetPnL.Status != StatusNotMeasured || !contains(rows[0].MissingMeasurements, "close_fill") {
		t.Fatalf("open position became zero-PnL sample: %+v", rows[0])
	}
}

func TestCloseCorrectionConservesQuantityBasisAndPnL(t *testing.T) {
	position, events := completeFixture("US", "AAPL", "close-correction")
	position.ResidualQuantity, position.ResidualEntryBasis = "7", "700"
	correction := closeDelta(events[1].Lineage, "close-correction-event", "close-correction-fill",
		"close-1", "close-leg-1", "-1", "-100", "-120")
	correction.CorrectionOfFillID = events[1].Lineage.FillID
	store, err := BuildDerivedAttributionStore([]PositionEvidence{position}, append(events, correction))
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := store.QueryAttribution(AttributionQuery{Market: "US"})
	if rows[0].ClosedQuantity != "3" || rows[0].AllocatedEntryBasis != "300" || rows[0].ResidualQuantity != "7" {
		t.Fatalf("close correction conservation=%+v", rows[0])
	}
	assertCompleteAmount(t, rows[0].Source.GrossPnL, "60", "USD")

	overrun := correction
	overrun.EventID, overrun.Lineage.FillID = "basis-overrun", "basis-overrun-fill"
	overrun.AllocatedEntryBasis = amount("-500", "journal-cost")
	if _, err := BuildDerivedAttributionStore([]PositionEvidence{position}, append(events, overrun)); !errors.Is(err, ErrCorrectionOverrun) {
		t.Fatalf("financial correction overrun error=%v", err)
	}
}

func TestAuthoritativeRoundingModesAreExactAndVersioned(t *testing.T) {
	value, _ := decimal("1.005")
	if got := ratText(roundRat(value, 2, RoundHalfEven)); got != "1" {
		t.Fatalf("half-even 1.005=%s", got)
	}
	value, _ = decimal("1.015")
	if got := ratText(roundRat(value, 2, RoundHalfEven)); got != "1.02" {
		t.Fatalf("half-even 1.015=%s", got)
	}
	value, _ = decimal("-1.005")
	if got := ratText(roundRat(value, 2, RoundHalfAwayFromZero)); got != "-1.01" {
		t.Fatalf("half-away -1.005=%s", got)
	}
}

func completeFixture(market, ticker, positionID string) (PositionEvidence, []FillDelta) {
	lineage := completeCompositeLineage(market, ticker, positionID)
	position := positionEvidence(lineage, "10", "6", "1000", "600", policy("USD", "KRW"))
	entry := entryDelta(lineage, "entry-event", "entry-fill", "10")
	close := closeDelta(lineage, "close-event", "close-fill", "close-1", "close-leg-1", "4", "400", "480")
	return position, []FillDelta{entry, close}
}

func stagedFixture() (PositionEvidence, []FillDelta) {
	lineage := completeCompositeLineage("US", "AAPL", "position-staged")
	position := positionEvidence(lineage, "10", "3", "1000", "300", policy("USD", "KRW"))
	events := []FillDelta{
		entryDelta(lineage, "entry-1", "entry-fill-1", "2"),
		entryDelta(lineage, "entry-2", "entry-fill-2", "3"),
		entryDelta(lineage, "entry-3", "entry-fill-3", "5"),
		closeDelta(lineage, "close-1", "close-fill-1", "close-1", "close-leg-1", "4", "400", "480"),
		closeDelta(lineage, "close-2", "close-fill-2", "close-2", "close-leg-2", "3", "300", "330"),
	}
	events[3].EntryFees, events[3].ExitFees, events[3].Taxes, events[3].FXCost =
		amountPtr("4", "broker-fee"), amountPtr("5", "exchange-fee"), amountPtr("1", "tax-ledger"), amountPtr("2", "fx-cost")
	events[4].EntryFees, events[4].ExitFees, events[4].Taxes, events[4].FXCost =
		amountPtr("3", "broker-fee"), amountPtr("4", "exchange-fee"), amountPtr("1", "tax-ledger"), amountPtr("1", "fx-cost")
	events[3].FX, events[4].FX = fx("1300"), fx("1310")
	return position, events
}

func completeCompositeLineage(market, ticker, positionID string) CompositeLineage {
	return CompositeLineage{Market: market, Ticker: ticker, CandidateID: "candidate-" + market,
		LaneID: "lane-" + market, LaneVersion: "v3", CampaignID: "campaign-" + market, LegID: "leg-" + market,
		DecisionID: "decision-" + market, AttemptID: "attempt-" + market, PositionID: positionID,
		PolicyID: "fifo", PolicyVersion: "policy-v4"}
}

func positionEvidence(lineage CompositeLineage, acquired, residual, totalBasis, residualBasis string, costPolicy CostPolicy) PositionEvidence {
	return PositionEvidence{Market: lineage.Market, PositionID: lineage.PositionID, CandidateID: lineage.CandidateID,
		LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion, CampaignID: lineage.CampaignID, LegID: lineage.LegID,
		AcquiredQuantity: acquired, ResidualQuantity: residual, TotalEntryBasis: totalBasis,
		ResidualEntryBasis: residualBasis, Policy: costPolicy}
}

func entryDelta(lineage CompositeLineage, eventID, fillID, quantity string) FillDelta {
	lineage.OrderID, lineage.FillID = "entry-order", fillID
	return FillDelta{EventID: eventID, Kind: FillKindEntry, Lineage: lineage, QuantityDelta: quantity}
}

func closeDelta(lineage CompositeLineage, eventID, fillID, closeID, closeLegID, quantity, basis, proceeds string) FillDelta {
	lineage.OrderID, lineage.FillID, lineage.CloseID, lineage.CloseLegID = "close-order", fillID, closeID, closeLegID
	return FillDelta{EventID: eventID, Kind: FillKindClose, Lineage: lineage, QuantityDelta: quantity,
		AllocatedEntryBasis: amount(basis, "journal-cost"), ExitProceeds: amount(proceeds, "broker-fill"),
		EntryFees: amountPtr("0", "broker-fee"), ExitFees: amountPtr("0", "exchange-fee"),
		Taxes: amountPtr("0", "tax-ledger"), FXCost: amountPtr("0", "fx-cost"), FX: fx("1300")}
}

func policy(sourceCurrency, reportingCurrency string) CostPolicy {
	return CostPolicy{ID: "fifo", Version: "policy-v4", SourceCurrency: sourceCurrency,
		ReportingCurrency: reportingCurrency, RoundingMode: RoundHalfEven, RoundingScale: 2,
		RoundingVersion: "round-v2"}
}

func amount(value, source string) AmountEvidence {
	return AmountEvidence{Value: value, Source: source, SourceVersion: "v1", ObservedAt: attributionAt}
}

func amountPtr(value, source string) *AmountEvidence {
	valueCopy := amount(value, source)
	return &valueCopy
}

func fx(rate string) *FXEvidence {
	return &FXEvidence{Source: "official-close-fx", SourceVersion: "v2", Rate: rate,
		AsOf: attributionAt, QuoteCurrency: "KRW", RoundingVersion: "round-v2"}
}

func assertCompleteAmount(t *testing.T, got AmountMetric, value, currency string) {
	t.Helper()
	if got.Status != StatusComplete || got.Value != value || got.Currency != currency {
		t.Fatalf("amount=%+v want %s %s", got, value, currency)
	}
}

func assertReportingEquation(t *testing.T, report PnLBreakdown) {
	t.Helper()
	want, ok := subtractDecimals(report.GrossPnL.Value, report.EntryFees.Value, report.ExitFees.Value,
		report.Taxes.Value, report.FXCost.Value)
	if !ok || report.NetPnL.Value != want {
		t.Fatalf("reporting equation gross=%+v costs=%+v/%+v/%+v/%+v net=%+v", report.GrossPnL,
			report.EntryFees, report.ExitFees, report.Taxes, report.FXCost, report.NetPnL)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
