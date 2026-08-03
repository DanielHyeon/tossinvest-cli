package reversallane

import (
	"testing"
	"time"
)

func TestStructuralConfirmationRequiresExactOrderWindowFreshnessAndScope(t *testing.T) {
	kr := mustKREvidence(t)
	valid := validStructure(kr.CommonEnvelope)
	if refusal := ValidateStructure(valid, kr.CommonEnvelope, time.Minute); refusal != "" {
		t.Fatalf("valid structure=%s", refusal)
	}
	equal := valid
	equal.Sweep.At = kr.EvaluatedAt
	equal.Break.At = kr.EvaluatedAt
	equal.Reclaim.At = kr.EvaluatedAt
	if refusal := ValidateStructure(equal, kr.CommonEnvelope, time.Minute); refusal != "" {
		t.Fatalf("inclusive structural equality=%s", refusal)
	}

	tests := []struct {
		name string
		edit func(*StructuralConfirmation)
		want RefusalCode
	}{
		{"missing reclaim", func(s *StructuralConfirmation) { s.Reclaim.RecordID = "" }, RefusalStructuralMissing},
		{"break before sweep", func(s *StructuralConfirmation) { s.Break.At = s.Sweep.At.Add(-time.Nanosecond) }, RefusalStructuralOrder},
		{"reclaim before break", func(s *StructuralConfirmation) { s.Reclaim.At = s.Break.At.Add(-time.Nanosecond) }, RefusalStructuralOrder},
		{"window one tick over", func(s *StructuralConfirmation) { s.Sweep.At = kr.EvaluatedAt.Add(-time.Minute - time.Nanosecond) }, RefusalStructuralStale},
		{"event stale", func(s *StructuralConfirmation) { s.Reclaim.FreshUntil = kr.EvaluatedAt.Add(-time.Nanosecond) }, RefusalStructuralStale},
		{"cross market reclaim", func(s *StructuralConfirmation) { s.Reclaim.Market = MarketUS }, RefusalScopeMismatch},
		{"generation mismatch", func(s *StructuralConfirmation) { s.Reclaim.PositionGeneration++ }, RefusalScopeMismatch},
		{"evidence version mismatch", func(s *StructuralConfirmation) { s.Reclaim.EvidenceVersion = "other" }, RefusalScopeMismatch},
		{"duplicate event record", func(s *StructuralConfirmation) {
			s.Reclaim.RecordID = s.Break.RecordID
			s.Reclaim.Digest = s.Break.Digest
		}, RefusalStructuralOrder},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := valid
			tt.edit(&candidate)
			if got := ValidateStructure(candidate, kr.CommonEnvelope, time.Minute); got != tt.want {
				t.Fatalf("got=%s want=%s", got, tt.want)
			}
		})
	}
}

func TestKRAndUSEvaluationReplayIsDeterministic(t *testing.T) {
	krPlan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	usPlan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "KRW", "USD", ptrFX(validFX()))
	krEvidence := mustKREvidence(t)
	usEvidence := mustUSEvidence(t)
	krRequest := KREvaluationRequest{Context: validContext(krPlan, 3), Evidence: krEvidence, Config: validKRConfig(), Structure: validStructure(krEvidence.CommonEnvelope)}
	usRequest := USEvaluationRequest{Context: validContext(usPlan, 3), Evidence: usEvidence, Config: validUSConfig(), Structure: validStructure(usEvidence.CommonEnvelope)}

	if first, replay := EvaluateKR(krRequest), EvaluateKR(krRequest); first != replay {
		t.Fatalf("KR replay diverged: first=%+v replay=%+v", first, replay)
	}
	if first, replay := EvaluateUS(usRequest), EvaluateUS(usRequest); first != replay {
		t.Fatalf("US replay diverged: first=%+v replay=%+v", first, replay)
	}
}

func TestEvaluationBindsCampaignScopeAndA066PolicyExactly(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	evidence := mustKREvidence(t)
	request := KREvaluationRequest{Context: validContext(plan, 2), Evidence: evidence, Config: validKRConfig(), Structure: validStructure(evidence.CommonEnvelope)}
	request.Evidence.AccountRef = "other-account"
	if got := EvaluateKR(request); got.Code != RefusalPlanInvalid || got.Quantity != 0 {
		t.Fatalf("cross-account evidence accepted=%+v", got)
	}
	request.Evidence = evidence
	request.Evidence.Symbol = "000660"
	if got := EvaluateKR(request); got.Code != RefusalPlanInvalid || got.Quantity != 0 {
		t.Fatalf("cross-symbol evidence accepted=%+v", got)
	}
	request.Evidence = evidence
	request.Context.Cap.PolicyDigest = "other-a066-policy"
	if got := EvaluateKR(request); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("mixed a066 policy accepted=%+v", got)
	}
}

func TestFinalLegRejectsPriceDeclineWithoutCompleteStructure(t *testing.T) {
	plan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "KRW", "USD", ptrFX(validFX()))
	request := USEvaluationRequest{Context: validContext(plan, 3), Evidence: mustUSEvidence(t), Config: validUSConfig()}
	request.Context.PriceDeclined = true
	result := EvaluateUS(request)
	if result.Kind != OutcomeRefusal || result.Code != RefusalStructuralMissing || result.Quantity != 0 || result.ExitDecisionCreated {
		t.Fatalf("price decline opened final leg: %+v", result)
	}
}

func TestKRAndUSFinalLegEvaluationIsIndependentInSameCycle(t *testing.T) {
	krPlan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	usPlan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "KRW", "USD", ptrFX(validFX()))
	krEvidence := mustKREvidence(t)
	usEvidence := mustUSEvidence(t)
	krRequest := KREvaluationRequest{Context: validContext(krPlan, 3), Evidence: krEvidence, Config: validKRConfig(), Structure: validStructure(krEvidence.CommonEnvelope)}
	usRequest := USEvaluationRequest{Context: validContext(usPlan, 3), Evidence: usEvidence, Config: validUSConfig(), Structure: validStructure(usEvidence.CommonEnvelope)}
	usRequest.Structure.Reclaim.Market = MarketKR

	kr := EvaluateKR(krRequest)
	us := EvaluateUS(usRequest)
	if kr.Kind != OutcomeDecision || kr.Quantity == 0 {
		t.Fatalf("valid KR blocked by US: %+v", kr)
	}
	if us.Kind != OutcomeRefusal || us.Code != RefusalScopeMismatch || us.Quantity != 0 {
		t.Fatalf("invalid US=%+v", us)
	}
	if kr.Lineage.Market != MarketKR || us.Lineage.Market != MarketUS || kr.Lineage.LaneID == us.Lineage.LaneID {
		t.Fatalf("lineage mixed: KR=%+v US=%+v", kr.Lineage, us.Lineage)
	}
}

func TestInvalidationSuppressesAddWithoutTakingCommonExitAuthority(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	evidence := mustKREvidence(t)
	request := KREvaluationRequest{Context: validContext(plan, 2), Evidence: evidence, Config: validKRConfig(), Structure: validStructure(evidence.CommonEnvelope)}
	request.Context.Invalidation = Invalidation{Structural: true, Code: "STRUCTURE_BROKEN"}
	result := EvaluateKR(request)
	if result.Kind != OutcomeInvalidation || result.Quantity != 0 || result.Code != RefusalInvalidated || result.ExitDecisionCreated || !result.CommonExitIndependent {
		t.Fatalf("invalidation authority leak=%+v", result)
	}
	if !(CommonExitProbe{StopTriggered: true}).CanProceed(result) {
		t.Fatal("common exit was coupled to lane invalidation")
	}
}

func TestStopNeverRetreatsOnScaleIn(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	evidence := mustKREvidence(t)
	request := KREvaluationRequest{Context: validContext(plan, 2), Evidence: evidence, Config: validKRConfig(), Structure: validStructure(evidence.CommonEnvelope)}
	request.Context.SavedEffectiveStopMinor = "100"
	request.Context.StopCandidate = StopCandidate{PriceMinor: "99", Valid: true, Source: "structure", Policy: "stop-v1", Digest: "stop-digest", ObservedAt: evidence.EvaluatedAt}
	if got := EvaluateKR(request); got.Kind != OutcomeRefusal || got.Code != RefusalStopRetreat {
		t.Fatalf("retreating stop accepted=%+v", got)
	}
	request.Context.StopCandidate.PriceMinor = "100"
	if got := EvaluateKR(request); got.Kind != OutcomeDecision {
		t.Fatalf("equal stop boundary refused=%+v", got)
	}
}

func validStructure(envelope CommonEnvelope) StructuralConfirmation {
	makeEvent := func(kind StructuralEventKind, at time.Time) StructuralEvent {
		return StructuralEvent{Kind: kind, AccountRef: envelope.AccountRef, Market: envelope.Market, Symbol: envelope.Symbol, PositionGeneration: envelope.PositionGeneration, EvidenceVersion: envelope.SchemaVersion, RecordID: string(kind) + "-record", Digest: string(kind) + "-digest", At: at, FreshUntil: envelope.FreshUntil}
	}
	return StructuralConfirmation{
		Sweep:   makeEvent(EventSweep, envelope.EvaluatedAt.Add(-30*time.Second)),
		Break:   makeEvent(EventBreak, envelope.EvaluatedAt.Add(-20*time.Second)),
		Reclaim: makeEvent(EventReclaim, envelope.EvaluatedAt.Add(-10*time.Second)),
	}
}
