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
	evaluation.entry.priceMinor, evaluation.stop.priceMinor, evaluation.target.priceMinor = "100", "95", "120"
	decision := routeDecision(descriptor, key, approved)

	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)},
		func(strategyrouter.RouteRequest) strategyrouter.RouteSetResult {
			return strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{decision}}
		},
		registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))

	if result.Code != RefusalNone || !result.ExecutionTerms.Valid() {
		t.Fatalf("accepted terms were not sealed: %+v", result)
	}
	if result.ExecutionTerms.Entry().PriceMinor() != "100" || result.ExecutionTerms.EffectiveStop().PriceMinor() != "95" ||
		result.ExecutionTerms.Target().PriceMinor() != "120" || result.ExecutionTerms.Quantity() != result.Quantity ||
		result.ExecutionTerms.LineageIdentity() != result.Lineage.Identity {
		t.Fatalf("execution terms changed: %+v", result.ExecutionTerms)
	}
}

func TestPriceProvenanceMajorDecimalPairedKRUS(t *testing.T) {
	tests := []struct {
		name, minor string
		scale       int
		want        string
	}{
		{name: "KR identity", minor: "50000", scale: 0, want: "50000"},
		{name: "US whole dollars", minor: "5000", scale: 2, want: "50"},
		{name: "US fractional dollars", minor: "5050", scale: 2, want: "50.5"},
		{name: "US sub-dollar", minor: "5", scale: 2, want: "0.05"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := (PriceProvenance{priceMinor: test.minor, minorScale: test.scale, unitVersion: "minor-v1"}).MajorDecimal()
			if !ok || got != test.want {
				t.Fatalf("MajorDecimal()=(%q,%t), want (%q,true)", got, ok, test.want)
			}
		})
	}
	for _, invalid := range []PriceProvenance{
		{priceMinor: "05", minorScale: 2, unitVersion: "minor-v1"},
		{priceMinor: "5", minorScale: -1, unitVersion: "minor-v1"},
		{priceMinor: "5", minorScale: 2, unitVersion: "forged"},
	} {
		if got, ok := invalid.MajorDecimal(); ok || got != "" {
			t.Fatalf("invalid provenance projected as (%q,%t)", got, ok)
		}
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
	evaluation.entry.priceMinor, evaluation.stop.priceMinor, evaluation.target.priceMinor = "100", "95", ""
	decision := routeDecision(descriptor, key, approved)

	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)},
		func(strategyrouter.RouteRequest) strategyrouter.RouteSetResult {
			return strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{decision}}
		},
		registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))

	if result.Code != RefusalExecutionTermsInvalid || result.Quantity != 0 || result.Lineage.Complete || result.ExecutionTerms.Valid() {
		t.Fatalf("missing target did not fail closed: %+v", result)
	}
}

func TestExecutionTermsDetectPostEvaluationMutation(t *testing.T) {
	lineage := sealLineage(Lineage{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930", CampaignID: "campaign", LegOrdinal: 1, Complete: true})
	evaluation := laneEvaluation{quantity: 2, entry: PriceProvenance{priceMinor: "100", source: "entry", version: "v1", digest: "e", asOf: "2026", currency: "KRW", unitVersion: "minor-v1"}, stop: PriceProvenance{priceMinor: "95", source: "stop", version: "v1", digest: "s", asOf: "2026", currency: "KRW", unitVersion: "minor-v1"}, target: PriceProvenance{priceMinor: "120", source: "target", version: "v1", digest: "t", asOf: "2026", currency: "KRW", unitVersion: "minor-v1"}, policy: ExecutionPolicy{identity: "p"}}
	terms, ok := sealExecutionTerms(lineage, evaluation)
	if !ok || !terms.Valid() {
		t.Fatalf("fixture terms invalid: %+v", terms)
	}

	mutations := []struct {
		name string
		edit func(*ExecutionTerms)
	}{
		{"account", func(v *ExecutionTerms) { v.accountRef = "other" }},
		{"market", func(v *ExecutionTerms) { v.market = strategyrouter.MarketUS }},
		{"symbol", func(v *ExecutionTerms) { v.symbol = "AAPL" }},
		{"campaign", func(v *ExecutionTerms) { v.campaignID = "other" }},
		{"leg", func(v *ExecutionTerms) { v.legOrdinal++ }},
		{"quantity", func(v *ExecutionTerms) { v.quantity++ }},
		{"entry", func(v *ExecutionTerms) { v.entry.priceMinor = "101" }},
		{"stop", func(v *ExecutionTerms) { v.stop.priceMinor = "96" }},
		{"target", func(v *ExecutionTerms) { v.target.priceMinor = "121" }},
		{"lineage", func(v *ExecutionTerms) { v.lineageIdentity = "forged" }},
		{"policy", func(v *ExecutionTerms) { v.policy.identity = "forged" }},
		{"identity", func(v *ExecutionTerms) { v.identity = "forged" }},
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
		evaluation := laneEvaluation{quantity: 1, entry: PriceProvenance{priceMinor: test.entry, source: "e", version: "v", digest: "d", asOf: "a", currency: "KRW", unitVersion: "u"}, stop: PriceProvenance{priceMinor: test.stop, source: "s", version: "v", digest: "d", asOf: "a", currency: "KRW", unitVersion: "u"}, target: PriceProvenance{priceMinor: test.target, source: "t", version: "v", digest: "d", asOf: "a", currency: "KRW", unitVersion: "u"}, policy: ExecutionPolicy{identity: "p"}}
		if terms, ok := sealExecutionTerms(lineage, evaluation); ok || terms.Valid() {
			t.Fatalf("invalid terms accepted: %+v", test)
		}
	}
}
