package weeklyvaluelane

import (
	"testing"
	"time"
)

func TestRegistryShipsKRAndUSWeeklyValueTogetherDefaultOFF(t *testing.T) {
	descriptors := Descriptors()
	if err := ValidateRegistry(descriptors); err != nil || len(descriptors) != 2 {
		t.Fatalf("registry=%+v err=%v", descriptors, err)
	}
	seen := map[string]Descriptor{}
	for _, descriptor := range descriptors {
		seen[descriptor.ID] = descriptor
		if descriptor.Release != WeeklyValueRelease || descriptor.DesiredState != StateOff || descriptor.EffectiveState != StateOff {
			t.Fatalf("not same-release OFF=%+v", descriptor)
		}
	}
	if seen[KRWeeklyLaneID].Market != MarketKR || seen[USWeeklyLaneID].Market != MarketUS || ValidateRegistry(descriptors[:1]) == nil {
		t.Fatalf("pair conformance=%+v", seen)
	}
}

func TestKRAndUSWeeklyEvaluationAreSourceAndMarketIndependent(t *testing.T) {
	krEvidence := mustKREvidence(t)
	usEvidence := mustUSEvidence(t)
	krPlan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	fx := validFX()
	usPlan := mustPlan(t, MarketUS, USWeeklyLaneID, "KRW", "USD", &fx, 14, "1000")
	krRequest := validEvaluation(t, krPlan, krEvidence, validKRConfig())
	usRequest := validEvaluation(t, usPlan, usEvidence, validUSConfig())
	krRequest.Evidence.FreshUntil = krRequest.Evidence.EvaluatedAt.Add(-time.Nanosecond)
	krRequest.Evidence.seal = evidenceSnapshotSeal(krRequest.Evidence)
	krRequest.authorization = mintDormantEvaluationAuthorization(krPlan, krRequest.Evidence)
	kr := EvaluateKR(krRequest)
	us := EvaluateUS(usRequest)
	if kr.Kind != OutcomeRefusal || kr.Code != RefusalEvidenceStale || kr.Quantity != 0 {
		t.Fatalf("KR stale=%+v", kr)
	}
	if us.Kind != OutcomeDecision || us.Quantity == 0 || us.Lineage.Source != SourceEDGAR || kr.Lineage.Market != MarketKR || us.Lineage.Market != MarketUS {
		t.Fatalf("peer coupled: KR=%+v US=%+v", kr, us)
	}
	wrong := validEvaluation(t, krPlan, usEvidence, validUSConfig())
	if got := EvaluateKR(wrong); got.Code != RefusalMarketMismatch && got.Code != RefusalSourceMismatch {
		t.Fatalf("cross-source accepted=%+v", got)
	}
}

func TestStopCapInvalidationAndCommonExitAuthorityRemainSeparated(t *testing.T) {
	evidence := mustKREvidence(t)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	request := validEvaluation(t, plan, evidence, validKRConfig())
	request.SavedEffectiveStopMinor = "90"
	request.StopCandidate = mintStopCandidate(stopCandidateInput{PriceMinor: "80", Version: "stop-v1", Source: "structure", Policy: "stop-policy-v1",
		Digest: "stop-digest-wide", ObservedAt: evidence.EvaluatedAt, FreshUntil: evidence.EvaluatedAt.Add(time.Hour)})
	if got := EvaluateKR(request); got.Kind != OutcomeDecision || got.EffectiveStopMinor != "90" {
		t.Fatalf("stop retreated=%+v", got)
	}
	request.SavedEffectiveStopMinor = ""
	if got := EvaluateKR(request); got.Code != RefusalStructuralStopCap {
		t.Fatalf("wide stop narrowed/accepted=%+v", got)
	}
	request = validEvaluation(t, plan, evidence, validKRConfig())
	request.Invalidation = Invalidation{Structural: true, Code: "VALUE_THESIS_BROKEN"}
	got := EvaluateKR(request)
	if got.Kind != OutcomeInvalidation || got.Quantity != 0 || got.ExitDecisionCreated || !got.CommonExitIndependent || !(CommonExitProbe{Emergency: true}).CanProceed(got) {
		t.Fatalf("exit authority leak=%+v", got)
	}
}

func validEvaluation(t *testing.T, plan CampaignPlan, evidence DisclosureEvidence, config DisclosureConfig) EvaluationRequest {
	t.Helper()
	evaluated := evidence.EvaluatedAt
	week := validWeek(plan.Market(), map[Market]string{MarketKR: "KR-XKRX-2026-W32", MarketUS: "US-XNYS-2026-W32"}[plan.Market()], "generation-A", "calendar", "2026-08-03", evaluated)
	state, result := ApplyReservation(NewReservationState(), reserveCommand(0, plan.CampaignID(), week, "reservation-1", "reserve-1", 1))
	if !result.Applied {
		t.Fatal(result.Code)
	}
	quantity := PlannedLegQuantity(plan, LegProgress{Ordinal: 1}, 20)
	request := EvaluationRequest{CandidateID: "candidate", Plan: plan, Evidence: evidence, Config: config, MarketWeek: week, Reservations: state,
		ReservationID: "reservation-1", Leg: LegProgress{Ordinal: 1}, Cap: validCap(t, plan, 1, quantity, "20"), Risk: NewRiskState(plan),
		StopCandidate:   mintStopCandidate(stopCandidateInput{PriceMinor: "90", Version: "stop-v1", Source: "structure", Policy: "stop-policy-v1", Digest: "stop-digest", ObservedAt: evaluated, FreshUntil: evaluated.Add(time.Hour)}),
		EntryPriceMinor: "100", StagedTargetMinor: "1300", EntryCostsMinor: "1", EstimatedExitCostsLeviesMinor: "1", MinimumRRPPM: 1}
	request.authorization = mintDormantEvaluationAuthorization(plan, evidence)
	request.executionTerms = mintExecutionTermsPreimage(plan, evidence, request.EntryPriceMinor, request.StagedTargetMinor, request.EntryCostsMinor, request.EstimatedExitCostsLeviesMinor, request.MinimumRRPPM)
	return request
}
