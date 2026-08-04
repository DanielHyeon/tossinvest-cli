package reversallane

import "testing"

func TestPairedReversalProposalIsCapFreePlannedRemaining(t *testing.T) {
	krPlan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	krEvidence := mustKREvidence(t)
	kr := KREvaluationRequest{Context: validContext(krPlan, 1), Evidence: krEvidence, Config: validKRConfig()}
	kr.Context.Cap = RiskCap{}

	usPlan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "USD", "USD", nil)
	usEvidence := mustUSEvidence(t)
	us := USEvaluationRequest{Context: validContext(usPlan, 1), Evidence: usEvidence, Config: validUSConfig()}
	us.Context.Cap = RiskCap{Market: MarketKR, QFinal: ^uint64(0)}

	for market, got := range map[Market]EvaluationResult{MarketKR: ProposeKR(kr), MarketUS: ProposeUS(us)} {
		if got.Kind != OutcomeDecision || got.Quantity != 2 || got.Lineage.CapSnapshotID != "" || got.Lineage.CapPolicyDigest != "" ||
			got.EntryPriceMinor != "110" || got.EffectiveStopMinor != "95" || got.TargetPriceMinor != "130" {
			t.Fatalf("%s cap-free proposal=%+v", market, got)
		}
	}
}

func TestPairedReversalAdmissionReducesProposalWithoutChangingTerms(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			plan := mustPlan(t, market, map[Market]string{MarketKR: KRReversalLaneID, MarketUS: USReversalLaneID}[market], 14, "1000",
				map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market], map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market], nil)
			evaluated := testEnvelopeForPlan(plan, mustKREvidence(t).EvaluatedAt).EvaluatedAt
			context := validContext(plan, 1)
			cap, err := mintRiskCap(plan, RiskCap{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: market, QFinal: 1,
				ReservationQuantity: 1, ReservationMinor: "20", SnapshotID: "a066-snapshot", PolicyDigest: "a066-policy", BucketSetDigest: "bucket-digest",
				Official: true, Frozen: true, ObservedAt: evaluated.Add(-1), FreshUntil: evaluated.Add(1)})
			if err != nil {
				t.Fatal(err)
			}
			context.Cap = cap
			if market == MarketKR {
				evidence := mustKREvidence(t)
				proposal := ProposeKR(KREvaluationRequest{Context: validContext(plan, 1), Evidence: evidence, Config: validKRConfig()})
				admitted := EvaluateKR(KREvaluationRequest{Context: context, Evidence: evidence, Config: validKRConfig()})
				assertReversalProposalReduction(t, proposal, admitted)
				return
			}
			evidence := mustUSEvidence(t)
			proposal := ProposeUS(USEvaluationRequest{Context: validContext(plan, 1), Evidence: evidence, Config: validUSConfig()})
			admitted := EvaluateUS(USEvaluationRequest{Context: context, Evidence: evidence, Config: validUSConfig()})
			assertReversalProposalReduction(t, proposal, admitted)
		})
	}
}

func assertReversalProposalReduction(t *testing.T, proposal, admitted EvaluationResult) {
	t.Helper()
	if proposal.Kind != OutcomeDecision || proposal.Quantity != 2 || admitted.Kind != OutcomeDecision || admitted.Quantity != 1 {
		t.Fatalf("proposal/admitted=%+v / %+v", proposal, admitted)
	}
	if proposal.EntryProvenance != admitted.EntryProvenance || proposal.StopProvenance != admitted.StopProvenance ||
		proposal.TargetProvenance != admitted.TargetProvenance || proposal.ExecutionPolicyDigest != admitted.ExecutionPolicyDigest {
		t.Fatalf("q_final changed non-quantity terms: proposal=%+v admitted=%+v", proposal, admitted)
	}
}
