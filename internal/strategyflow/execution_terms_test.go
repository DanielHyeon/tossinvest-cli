package strategyflow

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestEvaluateSealsExactExecutionTerms(t *testing.T) {
	descriptor := descriptorByID(t, "kr_short_flow_continuation_v1")
	approved := approvedFixture(t, strategyrouter.MarketKR)
	key, err := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketKR, approved.Symbol(), 1)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := acceptedEvaluation(descriptor, key, approved)
	evaluation.entryPriceMinor = "100"
	evaluation.effectiveStopMinor = "95"
	evaluation.targetPriceMinor = "120"
	decision := routeDecision(descriptor, key, approved)

	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)},
		func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
			return strategyrouter.RouteResult{Decision: decision}
		},
		registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))

	if result.Code != RefusalNone || !result.ExecutionTerms.Valid() {
		t.Fatalf("accepted terms were not sealed: %+v", result)
	}
	if result.ExecutionTerms.EntryPriceMinor != "100" || result.ExecutionTerms.EffectiveStopMinor != "95" ||
		result.ExecutionTerms.TargetPriceMinor != "120" || result.ExecutionTerms.Quantity != result.Quantity ||
		result.ExecutionTerms.LineageIdentity != result.Lineage.Identity {
		t.Fatalf("execution terms changed: %+v", result.ExecutionTerms)
	}
}

func TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms(t *testing.T) {
	descriptor := descriptorByID(t, "us_short_participation_continuation_v1")
	approved := approvedFixture(t, strategyrouter.MarketUS)
	key, err := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketUS, approved.Symbol(), 1)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := acceptedEvaluation(descriptor, key, approved)
	evaluation.entryPriceMinor = "100"
	evaluation.effectiveStopMinor = "95"
	evaluation.targetPriceMinor = ""
	decision := routeDecision(descriptor, key, approved)

	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)},
		func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
			return strategyrouter.RouteResult{Decision: decision}
		},
		registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))

	if result.Code != RefusalExecutionTermsInvalid || result.Quantity != 0 || result.Lineage.Complete || result.ExecutionTerms.Valid() {
		t.Fatalf("missing target did not fail closed: %+v", result)
	}
}

func TestExecutionTermsDetectPostEvaluationMutation(t *testing.T) {
	lineage := sealLineage(Lineage{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930", CampaignID: "campaign", LegOrdinal: 1, Complete: true})
	terms, ok := sealExecutionTerms(lineage, 2, "100", "95", "120")
	if !ok || !terms.Valid() {
		t.Fatalf("fixture terms invalid: %+v", terms)
	}

	mutations := []struct {
		name string
		edit func(*ExecutionTerms)
	}{
		{"account", func(v *ExecutionTerms) { v.AccountRef = "other" }},
		{"market", func(v *ExecutionTerms) { v.Market = strategyrouter.MarketUS }},
		{"symbol", func(v *ExecutionTerms) { v.Symbol = "AAPL" }},
		{"campaign", func(v *ExecutionTerms) { v.CampaignID = "other" }},
		{"leg", func(v *ExecutionTerms) { v.LegOrdinal++ }},
		{"quantity", func(v *ExecutionTerms) { v.Quantity++ }},
		{"entry", func(v *ExecutionTerms) { v.EntryPriceMinor = "101" }},
		{"stop", func(v *ExecutionTerms) { v.EffectiveStopMinor = "96" }},
		{"target", func(v *ExecutionTerms) { v.TargetPriceMinor = "121" }},
		{"lineage", func(v *ExecutionTerms) { v.LineageIdentity = "forged" }},
		{"identity", func(v *ExecutionTerms) { v.Identity = "forged" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			changed := terms
			mutation.edit(&changed)
			if changed.Valid() {
				t.Fatalf("mutation remained valid: %+v", changed)
			}
		})
	}
}

func TestSealExecutionTermsRejectsNonCanonicalOrUnorderedPrices(t *testing.T) {
	lineage := sealLineage(Lineage{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930", CampaignID: "campaign", LegOrdinal: 1, Complete: true})
	for _, test := range []struct{ entry, stop, target string }{
		{"", "95", "120"}, {"0100", "95", "120"}, {"100", "95", ""}, {"100", "100", "120"}, {"100", "101", "120"}, {"100", "95", "100"},
	} {
		if terms, ok := sealExecutionTerms(lineage, 1, test.entry, test.stop, test.target); ok || terms.Valid() {
			t.Fatalf("invalid terms accepted: %+v", test)
		}
	}
}
