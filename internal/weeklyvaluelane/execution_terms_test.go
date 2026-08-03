package weeklyvaluelane

import "testing"

func TestKRAndUSAcceptedOutcomesPreserveExactCappedTarget(t *testing.T) {
	krEvidence, usEvidence := mustKREvidence(t), mustUSEvidence(t)
	krPlan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	usPlan := mustPlan(t, MarketUS, USWeeklyLaneID, "USD", "USD", nil, 14, "1000")
	kr := EvaluateKR(validEvaluation(t, krPlan, krEvidence, validKRConfig()))
	us := EvaluateUS(validEvaluation(t, usPlan, usEvidence, validUSConfig()))

	for _, test := range []struct {
		market Market
		want   string
		got    Outcome
	}{{MarketKR, krEvidence.FairValueMinor, kr}, {MarketUS, usEvidence.FairValueMinor, us}} {
		if test.got.Kind != OutcomeDecision || test.got.EntryPriceMinor != "100" || test.got.EffectiveStopMinor != "90" ||
			test.got.TargetPriceMinor != test.want || test.got.TargetPriceMinor == "1300" {
			t.Fatalf("%s exact capped target lost: %+v", test.market, test.got)
		}
	}
}

func TestWeeklyMissingExplicitTargetFailsClosed(t *testing.T) {
	evidence := mustKREvidence(t)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	request := validEvaluation(t, plan, evidence, validKRConfig())
	request.StagedTargetMinor = ""
	if got := EvaluateKR(request); got.Code != RefusalExecutionTermsInvalid || got.Quantity != 0 || got.TargetPriceMinor != "" {
		t.Fatalf("missing target was estimated: %+v", got)
	}
}
