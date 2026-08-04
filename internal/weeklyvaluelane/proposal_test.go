package weeklyvaluelane

import (
	"testing"
	"time"
)

func TestPairedWeeklyProposalIsCapFreeButKeepsDurableReservation(t *testing.T) {
	krEvidence, usEvidence := mustKREvidence(t), mustUSEvidence(t)
	krPlan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	usPlan := mustPlan(t, MarketUS, USWeeklyLaneID, "USD", "USD", nil, 14, "1000")
	kr := validEvaluation(t, krPlan, krEvidence, validKRConfig())
	us := validEvaluation(t, usPlan, usEvidence, validUSConfig())
	kr.Cap = RiskCap{}
	us.Cap = RiskCap{qFinal: ^uint64(0)}

	for market, got := range map[Market]Outcome{MarketKR: ProposeKR(kr), MarketUS: ProposeUS(us)} {
		if got.Kind != OutcomeDecision || got.Quantity != 2 || got.Lineage.ReservationID != "reservation-1" ||
			got.Lineage.CapSnapshotID != "" || got.Lineage.CapReservationMinor != "" || got.Lineage.CapReservationQuantity != 0 ||
			got.ExecutionPolicy.CapSnapshotID != "" || got.EntryPriceMinor != "100" || got.EffectiveStopMinor != "90" {
			t.Fatalf("%s cap-free weekly proposal=%+v", market, got)
		}
	}

	kr.ReservationID = "missing"
	if got := ProposeKR(kr); got.Code != RefusalReservationMissing || got.Quantity != 0 {
		t.Fatalf("weekly proposal synthesized reservation: %+v", got)
	}
}

func TestPairedWeeklyAdmissionReducesProposalWithoutChangingPrices(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			evidence := map[Market]DisclosureEvidence{MarketKR: mustKREvidence(t), MarketUS: mustUSEvidence(t)}[market]
			plan := mustPlan(t, market, map[Market]string{MarketKR: KRWeeklyLaneID, MarketUS: USWeeklyLaneID}[market],
				map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market], map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market], nil, 14, "1000")
			config := map[Market]DisclosureConfig{MarketKR: validKRConfig(), MarketUS: validUSConfig()}[market]
			request := validEvaluation(t, plan, evidence, config)
			proposal := map[Market]Outcome{MarketKR: ProposeKR(request), MarketUS: ProposeUS(request)}[market]
			cap, err := mintRiskCap(plan, riskCapInput{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", QFinal: 1, ReservationQuantity: 1,
				ReservationMinor: "20", MaxStopDistanceMinor: "15", SnapshotID: "cap-1", PolicyDigest: "risk-policy", BucketSetDigest: "buckets",
				ObservedAt: evidence.EvaluatedAt.Add(-time.Minute).Format(time.RFC3339Nano), FreshUntil: evidence.EvaluatedAt.Add(time.Hour).Format(time.RFC3339Nano)})
			if err != nil {
				t.Fatal(err)
			}
			request.Cap = cap
			admitted := map[Market]Outcome{MarketKR: EvaluateKR(request), MarketUS: EvaluateUS(request)}[market]
			if proposal.Kind != OutcomeDecision || proposal.Quantity != 2 || admitted.Kind != OutcomeDecision || admitted.Quantity != 1 {
				t.Fatalf("proposal/admitted=%+v / %+v", proposal, admitted)
			}
			if proposal.EntryProvenance != admitted.EntryProvenance || proposal.StopProvenance != admitted.StopProvenance ||
				proposal.TargetProvenance.PriceMinor != admitted.TargetProvenance.PriceMinor || proposal.TargetProvenance.Source != admitted.TargetProvenance.Source ||
				proposal.TargetProvenance.Version != admitted.TargetProvenance.Version || proposal.TargetProvenance.AsOf != admitted.TargetProvenance.AsOf ||
				proposal.TargetProvenance.Currency != admitted.TargetProvenance.Currency || proposal.TargetProvenance.MinorScale != admitted.TargetProvenance.MinorScale ||
				proposal.TargetProvenance.UnitVersion != admitted.TargetProvenance.UnitVersion {
				t.Fatalf("q_final changed weekly prices: proposal=%+v admitted=%+v", proposal, admitted)
			}
		})
	}
}
