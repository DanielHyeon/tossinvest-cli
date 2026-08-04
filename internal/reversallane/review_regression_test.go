package reversallane

import (
	"reflect"
	"testing"
	"time"
)

func TestCallerDeclaredRiskCapAndFXCannotAuthorizeDecision(t *testing.T) {
	evidence := mustUSEvidence(t)
	forgedFX := FrozenFX{QuoteID: "forged", AsOf: evidence.EvaluatedAt.Add(-time.Second).Format(time.RFC3339Nano), Direction: FXQuoteToAccount, RateQuoteToAccount: "2", Haircut: "1.1", Digest: "forged", Official: true, Frozen: true}
	if _, err := BuildCampaignPlan(validPlanRequest(MarketUS, USReversalLaneID, 14, "1000", "KRW", "USD", &forgedFX)); err == nil {
		t.Fatal("caller-declared FX authorized a plan")
	}

	fx := validFX()
	plan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "KRW", "USD", &fx)
	request := USEvaluationRequest{Context: validContext(plan, 1), Evidence: evidence, Config: validUSConfig()}
	request.Context.Cap = RiskCap{Market: MarketUS, QFinal: 20, ReservationMinor: "20", SnapshotID: "forged", PolicyDigest: "a066-policy", BucketSetDigest: "forged", Official: true, Frozen: true}
	if got := EvaluateUS(request); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("caller-declared cap authorized=%+v", got)
	}

	cap := mustRiskCap(t, plan, evidence.EvaluatedAt)
	cap.QFinal++
	request.Context.Cap = cap
	if got := EvaluateUS(request); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("post-mint cap mutation authorized=%+v", got)
	}
	staleCap := mustRiskCap(t, plan, evidence.EvaluatedAt)
	staleCap.FreshUntil = evidence.EvaluatedAt.Add(-time.Nanosecond)
	staleCap.seal = riskCapSeal(staleCap)
	request.Context.Cap = staleCap
	if got := EvaluateUS(request); got.Code != RefusalCapInvalid {
		t.Fatalf("stale cap authorized=%+v", got)
	}
}

func mustRiskCap(t *testing.T, plan CampaignPlan, evaluatedAt time.Time) RiskCap {
	t.Helper()
	reservationQuantity := PlannedLegQuantity(plan, LegProgress{Ordinal: 1}, RiskCap{QFinal: 20})
	cap, err := mintRiskCap(plan, RiskCap{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: plan.Market(), QFinal: 20, ReservationQuantity: reservationQuantity, ReservationMinor: "20",
		SnapshotID: "a066-snapshot", PolicyDigest: "a066-policy", BucketSetDigest: "bucket-digest", Official: true, Frozen: true,
		ObservedAt: evaluatedAt.Add(-time.Second), FreshUntil: evaluatedAt.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	return cap
}

func TestRiskCapCannotReplayAcrossPlansWithSameMarketAndPolicy(t *testing.T) {
	evaluatedAt := time.Date(2026, 8, 4, 0, 0, 3, 0, time.UTC)
	first := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	request := validPlanRequest(MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	request.CampaignID = "campaign-KR-second"
	second, err := BuildCampaignPlan(request)
	if err != nil {
		t.Fatal(err)
	}
	cap := mustRiskCap(t, first, evaluatedAt)
	if refusal := AdmitRisk(second, NewRiskState(second), cap); refusal != RefusalCapInvalid {
		t.Fatalf("cross-plan cap replay refusal=%q, want %q", refusal, RefusalCapInvalid)
	}
	evidence := mustKREvidence(t)
	evaluation := KREvaluationRequest{Context: validContext(second, 1), Evidence: evidence, Config: validKRConfig()}
	evaluation.Context.Cap = cap
	if got := EvaluateKR(evaluation); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("cross-plan cap authorized evaluation=%+v", got)
	}
}

func TestFirstEntryEstablishesStopAndLaterScaleInCannotRetreat(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	evidence := mustKREvidence(t)
	request := KREvaluationRequest{Context: validContext(plan, 1), Evidence: evidence, Config: validKRConfig()}
	request.Context.SavedEffectiveStopMinor = ""
	if got := EvaluateKR(request); got.Kind != OutcomeDecision {
		t.Fatalf("first-entry stop establishment refused=%+v", got)
	}
	request.Context.StopCandidate.observedAt = evidence.EvaluatedAt.Add(time.Nanosecond)
	if got := EvaluateKR(request); got.Code != RefusalStopRetreat {
		t.Fatalf("future stop evidence accepted=%+v", got)
	}
	request.Context.StopCandidate.observedAt = evidence.EvaluatedAt
	request.Context.StopCandidate.policy = ""
	if got := EvaluateKR(request); got.Code != RefusalStopRetreat {
		t.Fatalf("unversioned stop accepted=%+v", got)
	}
}

func TestMissingFillAndInvalidCancelLatchWithoutDroppingEvidence(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "100", "KRW", "KRW", nil)
	state := NewRiskState(plan)
	state.HeldMinor = "20"
	missingID := validFillEvent(plan, "")
	next, result := ApplyFillRisk(state, plan, missingID)
	if !result.Applied || !next.Latches[LatchUnknownActualRisk] || len(next.Fills) != 1 || next.HeldMinor != state.HeldMinor || next.FilledMinor != state.FilledMinor {
		t.Fatalf("missing-ID fill silently lost=%+v/%+v", next, result)
	}
	retry, result := ApplyFillRisk(next, plan, missingID)
	if !result.Duplicate || len(retry.Fills) != 1 {
		t.Fatalf("same unidentified retry was not idempotent=%+v/%+v", retry, result)
	}
	different := missingID
	different.ObservedAt = different.ObservedAt.Add(time.Nanosecond)
	different.SourceDigest = "another-source-observation"
	twoFills, result := ApplyFillRisk(next, plan, different)
	if !result.Applied || result.Duplicate || len(twoFills.Fills) != 2 {
		t.Fatalf("different unidentified fills were coalesced=%+v/%+v", twoFills, result)
	}

	badCancelEvent := CancelRiskEvent{CancelID: "", CampaignID: plan.CampaignID(), LegOrdinal: 1, OrderRef: "order-1", ReleaseHeldMinor: "10", ObservedAt: missingID.ObservedAt, SourceDigest: "cancel-source-1"}
	badCancel, result := ApplyCancelRisk(state, badCancelEvent)
	if !result.Applied || !badCancel.Latches[LatchUnknownActualRisk] || badCancel.HeldMinor != "20" {
		t.Fatalf("invalid cancel silently ignored=%+v/%+v", badCancel, result)
	}
	badCancelRetry, result := ApplyCancelRisk(badCancel, badCancelEvent)
	if !result.Duplicate || result.Applied || len(badCancelRetry.Cancels) != 1 {
		t.Fatalf("same unidentified cancel retry was not idempotent=%+v/%+v", badCancelRetry, result)
	}
	differentCancel := badCancelEvent
	differentCancel.SourceDigest = "cancel-source-2"
	twoCancels, result := ApplyCancelRisk(badCancel, differentCancel)
	if !result.Applied || result.Duplicate || len(twoCancels.Cancels) != 2 || twoCancels.HeldMinor != "20" {
		t.Fatalf("different unidentified cancels were coalesced or released risk=%+v/%+v", twoCancels, result)
	}
	tooLarge, result := ApplyCancelRisk(state, CancelRiskEvent{CancelID: "cancel", ReleaseHeldMinor: "21"})
	if !result.Applied || !tooLarge.Latches[LatchUnknownActualRisk] || tooLarge.HeldMinor != "20" {
		t.Fatalf("invalid release silently ignored=%+v/%+v", tooLarge, result)
	}

	identified := missingID
	identified.FillID = "fill-conflict"
	first, result := ApplyFillRisk(state, plan, identified)
	if !result.Applied {
		t.Fatal("identified fill not applied")
	}
	identified.EntryPriceMinor = "101"
	conflicted, result := ApplyFillRisk(first, plan, identified)
	if !result.Applied || !result.Duplicate || !conflicted.Latches[LatchUnknownActualRisk] || len(conflicted.Fills) != 2 {
		t.Fatalf("inconsistent duplicate coalesced=%+v/%+v", conflicted, result)
	}
}

func TestZeroQuantityFillIsRecordedButNeverAppliedAsPositiveFill(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "100", "KRW", "KRW", nil)
	state := NewRiskState(plan)
	state.FilledMinor = "7"
	state.HeldMinor = "40"
	event := validFillEvent(plan, "zero-fill")
	event.Quantity = 0

	next, result := ApplyFillRisk(state, plan, event)
	if result.Applied || result.Duplicate {
		t.Fatalf("zero quantity reported as applied positive fill: %+v", result)
	}
	recorded, ok := next.Fills[event.FillID]
	if !ok || recorded.Applied {
		t.Fatalf("zero quantity evidence was not retained as non-applied: %+v", next.Fills)
	}
	if next.FilledMinor != "7" || next.HeldMinor != "40" || !next.Latches[LatchUnknownActualRisk] {
		t.Fatalf("zero quantity moved risk or failed closed: %+v", next)
	}
	retry, result := ApplyFillRisk(next, plan, event)
	if !result.Duplicate || result.Applied || !reflect.DeepEqual(retry, next) {
		t.Fatalf("zero quantity retry was not idempotent: %+v/%+v", retry, result)
	}

	positive := validFillEvent(plan, "shared-fill-id")
	positiveState, result := ApplyFillRisk(state, plan, positive)
	if !result.Applied {
		t.Fatal("positive baseline fill was not applied")
	}
	zeroConflict := positive
	zeroConflict.Quantity = 0
	conflicted, result := ApplyFillRisk(positiveState, plan, zeroConflict)
	if result.Applied || !result.Duplicate || !conflicted.Latches[LatchUnknownActualRisk] {
		t.Fatalf("zero quantity conflict was represented as a positive application: %+v/%+v", conflicted, result)
	}
	conflictKey := "conflict:" + zeroConflict.FillID + ":" + fillRiskFingerprint(zeroConflict)
	if recorded, ok := conflicted.Fills[conflictKey]; !ok || recorded.Applied {
		t.Fatalf("zero quantity conflict evidence was not retained as non-applied: %+v", conflicted.Fills)
	}
}

func TestRiskCapCannotReplayAcrossPlansAndRejectsZeroBasis(t *testing.T) {
	evaluatedAt := mustKREvidence(t).EvaluatedAt
	planA := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	requestB := validPlanRequest(MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	requestB.CampaignID = "campaign-KR-peer"
	planB, err := BuildCampaignPlan(requestB)
	if err != nil {
		t.Fatal(err)
	}
	if planA.Digest() == planB.Digest() {
		t.Fatal("distinct campaigns produced the same plan digest")
	}
	capA := mustRiskCap(t, planA, evaluatedAt)
	contextB := validContext(planB, 1)
	contextB.Cap = capA
	request := KREvaluationRequest{Context: contextB, Evidence: mustKREvidence(t), Config: validKRConfig()}
	if got := EvaluateKR(request); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("cap replayed across plans: %+v", got)
	}

	for _, candidate := range []RiskCap{
		{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: MarketKR, QFinal: 0, ReservationQuantity: 1, ReservationMinor: "20", SnapshotID: "zero-q", PolicyDigest: "a066-policy", BucketSetDigest: "bucket", Official: true, Frozen: true, ObservedAt: evaluatedAt, FreshUntil: evaluatedAt},
		{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: MarketKR, QFinal: 1, ReservationQuantity: 0, ReservationMinor: "20", SnapshotID: "zero-reservation-q", PolicyDigest: "a066-policy", BucketSetDigest: "bucket", Official: true, Frozen: true, ObservedAt: evaluatedAt, FreshUntil: evaluatedAt},
		{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: MarketKR, QFinal: 1, ReservationQuantity: 1, ReservationMinor: "0", SnapshotID: "zero-risk", PolicyDigest: "a066-policy", BucketSetDigest: "bucket", Official: true, Frozen: true, ObservedAt: evaluatedAt, FreshUntil: evaluatedAt},
	} {
		if _, err := mintRiskCap(planA, candidate); err == nil {
			t.Fatalf("zero cap basis accepted: %+v", candidate)
		}
	}
}

func TestRiskCapReservationQuantityMustEqualEvaluationQuantity(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	evidence := mustKREvidence(t)
	cap, err := mintRiskCap(plan, RiskCap{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: MarketKR, QFinal: 20, ReservationQuantity: 3, ReservationMinor: "30",
		SnapshotID: "wrong-reservation-quantity", PolicyDigest: "a066-policy", BucketSetDigest: "bucket", Official: true, Frozen: true,
		ObservedAt: evidence.EvaluatedAt.Add(-time.Second), FreshUntil: evidence.EvaluatedAt.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	request := KREvaluationRequest{Context: validContext(plan, 2), Evidence: evidence, Config: validKRConfig()}
	request.Context.Cap = cap
	if got := EvaluateKR(request); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("mismatched reservation quantity authorized=%+v", got)
	}
}

func TestFillRiskStateCorruptionLatchesWithoutDestructiveAccounting(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "10000000000000000000000000000000000000000000000000000000000000000000000000000", "KRW", "KRW", nil)
	event := validFillEvent(plan, "fill-risk-corruption")

	wrongPlanState := NewRiskState(plan)
	wrongPlanState.PlanDigest = "other-plan"
	wrongPlanState.HeldMinor = "50"
	wrongPlanState.FilledMinor = "10"
	got, result := ApplyFillRisk(wrongPlanState, plan, event)
	if !result.Applied || !got.Latches[LatchUnknownActualRisk] || got.HeldMinor != "50" || got.FilledMinor != "10" || len(got.Fills) != 1 {
		t.Fatalf("plan mismatch corrupted accounting: %+v/%+v", got, result)
	}

	releaseOverHeld := NewRiskState(plan)
	releaseOverHeld.HeldMinor = "20"
	releaseOverHeld.FilledMinor = "10"
	got, result = ApplyFillRisk(releaseOverHeld, plan, event)
	if !result.Applied || !got.Latches[LatchUnknownActualRisk] || got.HeldMinor != "20" || len(got.Fills) != 1 {
		t.Fatalf("release over held was silently zeroed: %+v/%+v", got, result)
	}

	overflow := NewRiskState(plan)
	overflow.HeldMinor = "40"
	overflow.FilledMinor = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
	got, result = ApplyFillRisk(overflow, plan, event)
	if !result.Applied || !got.Latches[LatchUnknownActualRisk] || got.FilledMinor != overflow.FilledMinor || len(got.Fills) != 1 {
		t.Fatalf("filled overflow was committed: %+v/%+v", got, result)
	}
}

func TestZeroPlanAndOutOfRangeConfigFailClosed(t *testing.T) {
	for _, edit := range []func(*PlanRequest){
		func(request *PlanRequest) { request.PlannedQuantity = 0 },
		func(request *PlanRequest) { request.RiskBudgetMinor = "0" },
	} {
		request := validPlanRequest(MarketKR, KRReversalLaneID, 14, "100", "KRW", "KRW", nil)
		edit(&request)
		if _, err := BuildCampaignPlan(request); err == nil {
			t.Fatal("zero plan basis accepted")
		}
	}
	kr, config := mustKREvidence(t), validKRConfig()
	for _, edit := range []func(*KRConfig){
		func(config *KRConfig) { config.StructuralWindow = 0 },
		func(config *KRConfig) { config.StructuralWindow = maxStructuralWindow + time.Nanosecond },
		func(config *KRConfig) { config.MinimumAbsorptionPPM = 0 },
		func(config *KRConfig) { config.MinimumAbsorptionPPM = maxMetricPPM + 1 },
	} {
		candidate := config
		edit(&candidate)
		if got := EvaluateKRMetric(kr, candidate); got.Refusal != RefusalConfigMismatch {
			t.Fatalf("out-of-range config=%+v result=%+v", candidate, got)
		}
	}
}

func TestStructureIsRequiredForFinalLegOnlyAndPriceFieldCannotBypass(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	evidence := mustKREvidence(t)
	for _, declined := range []bool{false, true} {
		request := KREvaluationRequest{Context: validContext(plan, 3), Evidence: evidence, Config: validKRConfig()}
		request.Context.PriceDeclined = declined
		if got := EvaluateKR(request); got.Code != RefusalStructuralMissing || got.Quantity != 0 {
			t.Fatalf("declined=%v bypassed final structure: %+v", declined, got)
		}
	}
	earlier := KREvaluationRequest{Context: validContext(plan, 2), Evidence: evidence, Config: validKRConfig()}
	if got := EvaluateKR(earlier); got.Kind != OutcomeDecision || got.Quantity == 0 {
		t.Fatalf("structure incorrectly inferred for non-final leg=%+v", got)
	}
}
