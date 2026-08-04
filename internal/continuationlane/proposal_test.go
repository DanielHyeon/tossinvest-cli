package continuationlane

import "testing"

func TestPairedContinuationProposalIsCapFreePlannedRemaining(t *testing.T) {
	krPlan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	krEvidence, krConfig := validKRFixture()
	kr := validKREvaluation(t, krPlan, krEvidence, krConfig)
	kr.Context.Cap = RiskCap{}

	usPlan := mustPlan(t, MarketUS, USContinuationLaneID, "USD", "USD", nil, 14, "1000")
	usEvidence, usConfig := validUSFixture()
	us := validUSEvaluation(t, usPlan, usEvidence, usConfig)
	us.Context.Cap = RiskCap{Market: MarketKR, QFinal: ^uint64(0)}

	for market, got := range map[Market]Outcome{MarketKR: ProposeKR(kr), MarketUS: ProposeUS(us)} {
		if got.Kind != OutcomeDecision || got.Code != RefusalNone || got.Quantity != 8 ||
			got.EntryPriceMinor != "110" || got.EffectiveStopMinor != "95" || got.TargetPriceMinor != "130" {
			t.Fatalf("%s cap-free proposal=%+v", market, got)
		}
	}
}

func TestPairedContinuationAdmissionReducesProposalWithoutChangingIt(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			plan := mustPlan(t, market, map[Market]string{MarketKR: KRContinuationLaneID, MarketUS: USContinuationLaneID}[market],
				map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market], map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market], nil, 14, "1000")
			if market == MarketKR {
				evidence, config := validKRFixture()
				request := validKREvaluation(t, plan, evidence, config)
				proposal := ProposeKR(request)
				request.Context.Cap = riskCapFor(plan, 3, 3, "20")
				admitted := EvaluateKR(request)
				assertContinuationProposalReduction(t, proposal, admitted)
				return
			}
			evidence, config := validUSFixture()
			request := validUSEvaluation(t, plan, evidence, config)
			proposal := ProposeUS(request)
			request.Context.Cap = riskCapFor(plan, 3, 3, "20")
			admitted := EvaluateUS(request)
			assertContinuationProposalReduction(t, proposal, admitted)
		})
	}
}

func assertContinuationProposalReduction(t *testing.T, proposal, admitted Outcome) {
	t.Helper()
	if proposal.Kind != OutcomeDecision || proposal.Quantity != 8 || admitted.Kind != OutcomeDecision || admitted.Quantity != 3 {
		t.Fatalf("proposal/admitted=%+v / %+v", proposal, admitted)
	}
	if proposal.EntryProvenance != admitted.EntryProvenance || proposal.StopProvenance != admitted.StopProvenance ||
		proposal.TargetProvenance != admitted.TargetProvenance || proposal.ExecutionPolicyDigest != admitted.ExecutionPolicyDigest {
		t.Fatalf("q_final changed non-quantity terms: proposal=%+v admitted=%+v", proposal, admitted)
	}
}
