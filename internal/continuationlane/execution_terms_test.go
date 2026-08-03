package continuationlane

import "testing"

func TestKRAndUSAcceptedOutcomesPreserveExplicitExecutionTerms(t *testing.T) {
	krPlan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	krEvidence, krConfig := validKRFixture()
	kr := EvaluateKR(validKREvaluation(t, krPlan, krEvidence, krConfig))

	usPlan := mustPlan(t, MarketUS, USContinuationLaneID, "USD", "USD", nil, 14, "1000")
	usEvidence, usConfig := validUSFixture()
	us := EvaluateUS(validUSEvaluation(t, usPlan, usEvidence, usConfig))

	for market, outcome := range map[Market]Outcome{MarketKR: kr, MarketUS: us} {
		if outcome.Kind != OutcomeDecision || outcome.Code != RefusalNone || outcome.EntryPriceMinor != "110" ||
			outcome.EffectiveStopMinor != "95" || outcome.TargetPriceMinor != "130" {
			t.Fatalf("%s exact execution terms lost: %+v", market, outcome)
		}
	}
}

func TestContinuationExecutionTermsMissingOrMutatedFailClosed(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.ExecutionTerms = ExecutionTermsPreimage{}
	if got := EvaluateKR(request); got.Code != RefusalExecutionTermsInvalid || got.Quantity != 0 {
		t.Fatalf("missing terms accepted: %+v", got)
	}

	request = validKREvaluation(t, plan, evidence, config)
	request.Context.ExecutionTerms.TargetPriceMinor = "131"
	if got := EvaluateKR(request); got.Code != RefusalExecutionTermsInvalid || got.Quantity != 0 {
		t.Fatalf("mutated terms accepted: %+v", got)
	}
}

func TestContinuationExecutionTermsConstructorRejectsEstimatedOrNonCanonicalValues(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	for _, prices := range [][2]string{{"", "130"}, {"110", ""}, {"0110", "130"}, {"110", "110"}, {"130", "120"}} {
		if value, err := NewExecutionTermsPreimage(plan, prices[0], prices[1]); err == nil || value.valid(plan) {
			t.Fatalf("invalid explicit prices accepted: %q/%q", prices[0], prices[1])
		}
	}
}
