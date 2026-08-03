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

func TestWeeklyRRPolicyIdentityBindsFullPreimage(t *testing.T) {
	evidence := mustKREvidence(t)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	base := validEvaluation(t, plan, evidence, validKRConfig())
	first := EvaluateKR(base)
	changed := validEvaluation(t, plan, evidence, validKRConfig())
	changed.EntryCostsMinor = "2"
	changed.executionTerms = mintExecutionTermsPreimage(plan, evidence, changed.EntryPriceMinor, changed.StagedTargetMinor, changed.EntryCostsMinor, changed.EstimatedExitCostsLeviesMinor, changed.MinimumRRPPM)
	second := EvaluateKR(changed)
	if first.Kind != OutcomeDecision || second.Kind != OutcomeDecision || first.ExecutionPolicy.Identity == "" || first.ExecutionPolicy.Identity == second.ExecutionPolicy.Identity {
		t.Fatalf("RR policy preimage did not change identity: first=%+v second=%+v", first.ExecutionPolicy, second.ExecutionPolicy)
	}
	if second.ExecutionPolicy.EntryCostsMinor != "2" || second.ExecutionPolicy.DecisionDigest == "" || second.ExecutionPolicy.CalendarDigest == "" || second.ExecutionPolicy.CapSnapshotID == "" {
		t.Fatalf("full RR preimage missing: %+v", second.ExecutionPolicy)
	}
}

func TestWeeklySavedStopNeverInheritsCandidateProvenance(t *testing.T) {
	evidence := mustKREvidence(t)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	request := validEvaluation(t, plan, evidence, validKRConfig())
	request.SavedEffectiveStopMinor = "95"
	request.savedStopAuthority = savedStopAuthority{provenance: PriceProvenance{PriceMinor: "95", Source: "caller", Version: "forged", Digest: "forged"}}
	if forged := EvaluateKR(request); forged.Code != RefusalExecutionTermsInvalid || forged.Quantity != 0 {
		t.Fatalf("caller-forged saved stop authority accepted: %+v", forged)
	}
	request.savedStopAuthority = mintSavedStopAuthority(plan, evidence, "95")
	got := EvaluateKR(request)
	if got.Kind != OutcomeDecision || got.EffectiveStopMinor != "95" || got.StopProvenance.PriceMinor != "95" || got.StopProvenance.Source == request.StopCandidate.Source {
		t.Fatalf("saved stop inherited candidate provenance: %+v", got)
	}
}
