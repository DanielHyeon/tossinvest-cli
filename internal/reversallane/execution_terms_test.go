package reversallane

import "testing"

func TestKRAndUSAcceptedOutcomesPreserveExplicitExecutionTerms(t *testing.T) {
	krPlan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	krEvidence := mustKREvidence(t)
	kr := EvaluateKR(KREvaluationRequest{Context: validContext(krPlan, 1), Evidence: krEvidence, Config: validKRConfig()})

	usPlan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "USD", "USD", nil)
	usEvidence := mustUSEvidence(t)
	us := EvaluateUS(USEvaluationRequest{Context: validContext(usPlan, 1), Evidence: usEvidence, Config: validUSConfig()})

	for market, outcome := range map[Market]EvaluationResult{MarketKR: kr, MarketUS: us} {
		if outcome.Kind != OutcomeDecision || outcome.Code != "" || outcome.EntryPriceMinor != "110" ||
			outcome.EffectiveStopMinor != "95" || outcome.TargetPriceMinor != "130" {
			t.Fatalf("%s exact execution terms lost: %+v", market, outcome)
		}
	}
}

func TestReversalExecutionTermsMissingOrMutatedFailClosed(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	evidence := mustKREvidence(t)
	request := KREvaluationRequest{Context: validContext(plan, 1), Evidence: evidence, Config: validKRConfig()}
	request.Context.ExecutionTerms = ExecutionTermsPreimage{}
	if got := EvaluateKR(request); got.Code != RefusalExecutionTermsInvalid || got.Quantity != 0 {
		t.Fatalf("missing terms accepted: %+v", got)
	}

	request.Context = validContext(plan, 1)
	request.Context.ExecutionTerms.EntryPriceMinor = "111"
	if got := EvaluateKR(request); got.Code != RefusalExecutionTermsInvalid || got.Quantity != 0 {
		t.Fatalf("mutated terms accepted: %+v", got)
	}
}

func TestReversalExecutionTermsConstructorRejectsEstimatedOrNonCanonicalValues(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	for _, prices := range [][2]string{{"", "130"}, {"110", ""}, {"0110", "130"}, {"110", "110"}, {"130", "120"}} {
		if value, err := NewExecutionTermsPreimage(plan, prices[0], prices[1]); err == nil || value.valid(plan) {
			t.Fatalf("invalid explicit prices accepted: %q/%q", prices[0], prices[1])
		}
	}
}
