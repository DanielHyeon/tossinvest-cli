package continuationlane

import (
	"reflect"
	"strings"
	"testing"
)

func TestImmutableEightFourTwoAllocationUsesFloorAndFinalRemainder(t *testing.T) {
	for _, tc := range []struct {
		q    uint64
		want [3]uint64
	}{
		{1, [3]uint64{0, 0, 1}},
		{2, [3]uint64{1, 0, 1}},
		{7, [3]uint64{4, 2, 1}},
		{14, [3]uint64{8, 4, 2}},
		{15, [3]uint64{8, 4, 3}},
		{^uint64(0), [3]uint64{10540996613548315208, 5270498306774157604, 2635249153387078803}},
	} {
		plan, err := BuildCampaignPlan(PlanRequest{LaneID: KRContinuationLaneID, LaneVersion: LaneVersionV1, Market: MarketKR,
			AccountRef: "acct", Symbol: "005930", CampaignID: "campaign", PositionGeneration: 1,
			RiskBudgetMinor: "100000000000000000000000000000", PerShareRiskMinor: "10", PlannedQuantity: tc.q,
			PolicyDigest: "policy", ConfigDigest: "config", AccountCurrency: "KRW", QuoteCurrency: "KRW"})
		if err != nil {
			t.Fatalf("Q=%d: %v", tc.q, err)
		}
		if plan.LegCeilings != tc.want || plan.LegCeilings[0]+plan.LegCeilings[1]+plan.LegCeilings[2] != tc.q {
			t.Fatalf("Q=%d got=%v want=%v", tc.q, plan.LegCeilings, tc.want)
		}
	}
}

func TestPartialCancelDoesNotReallocateAndA066CapStillBinds(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.Leg = LegProgress{Ordinal: 2, FilledQuantity: 0}
	request.Context.Cap = riskCapFor(plan, 3, 3, "20")
	got := EvaluateKR(request)
	if got.Kind != OutcomeDecision || got.Quantity != 3 || got.Lineage.PlannedCeiling != 4 {
		t.Fatalf("capped second leg=%+v", got)
	}

	request.Context.Cap = riskCapFor(plan, 10, 4, "20")
	got = EvaluateKR(request)
	if got.Quantity != 4 {
		t.Fatalf("unused first-leg quantity was reallocated: %+v", got)
	}
	request.Context.Leg = LegProgress{Ordinal: 2, FilledQuantity: 1}
	request.Context.Cap = riskCapFor(plan, 10, 3, "20")
	if got := EvaluateKR(request); got.Quantity != 3 {
		t.Fatalf("partial remaining=%+v", got)
	}
	request.Context.Leg.Cancelled = true
	if got := EvaluateKR(request); got.Kind != OutcomeRefusal || got.Code != RefusalLegTerminal || got.Quantity != 0 {
		t.Fatalf("cancelled leg admitted=%+v", got)
	}
	request.Context.Leg.Cancelled = false
	request.Context.Leg.Expired = true
	if got := EvaluateKR(request); got.Kind != OutcomeRefusal || got.Code != RefusalLegTerminal || got.Quantity != 0 {
		t.Fatalf("expired leg admitted=%+v", got)
	}
}

func TestScaleInStopNeverRetreatsAndInvalidStopRefuses(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.SavedEffectiveStopMinor = "100"
	request.Context.SavedStopProvenance = mintSavedStopProvenance(plan, evidence.Envelope, "100")
	request.Context.StopCandidate = mustStopCandidate(t, "90", "2026-08-04T00:00:02Z", "2026-08-04T00:01:00Z")
	got := EvaluateKR(request)
	if got.Kind != OutcomeDecision || got.EffectiveStopMinor != "100" {
		t.Fatalf("stop retreated=%+v", got)
	}
	request.Context.StopCandidate.Valid = false
	if got := EvaluateKR(request); got.Kind != OutcomeRefusal || got.Code != RefusalStopInvalid {
		t.Fatalf("invalid stop admitted=%+v", got)
	}
}

func TestA066CapBindsProposedQuantityAndFrozenFXSnapshot(t *testing.T) {
	fx := validUSFX()
	plan := mustPlan(t, MarketUS, USContinuationLaneID, "KRW", "USD", &fx, 14, "1000")
	evidence, config := validUSFixture()
	request := validUSEvaluation(t, plan, evidence, config)
	request.Context.Cap.ReservationQuantity++
	if got := EvaluateUS(request); got.Code != RefusalCapInvalid {
		t.Fatalf("reservation quantity mismatch accepted=%+v", got)
	}
	request = validUSEvaluation(t, plan, evidence, config)
	changed := fx
	changed.QuoteID = "other"
	request.Context.Cap.FX = &changed
	if got := EvaluateUS(request); got.Code != RefusalCapInvalid {
		t.Fatalf("mixed cap FX accepted=%+v", got)
	}
}

func TestA066CapPolicyDigestBindsExactPlanPolicy(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	input := riskCapInput{QFinal: 20, ReservationQuantity: plan.LegCeilings[0], ReservationMinor: "20", SnapshotID: "a066-snapshot",
		PolicyDigest: "other-policy", BucketSetDigest: "buckets", ObservedAt: "2026-08-04T00:00:00Z", FreshUntil: "2026-08-04T00:01:00Z"}
	if _, err := newRiskCap(plan, input); err == nil {
		t.Fatal("risk cap constructor accepted a policy digest outside the exact plan")
	}

	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.Cap.PolicyDigest = "other-policy"
	request.Context.Cap.seal = riskCapSeal(request.Context.Cap)
	if got := EvaluateKR(request); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("resealed mismatched cap policy authorized=%+v", got)
	}
}

func TestFrozenFXAndRiskCapRequireSealedConstructors(t *testing.T) {
	fx := validUSFX()
	plan := mustPlan(t, MarketUS, USContinuationLaneID, "KRW", "USD", &fx, 14, "1000")
	evidence, config := validUSFixture()
	request := validUSEvaluation(t, plan, evidence, config)
	request.Context.Cap.SnapshotID = "invented"
	if got := EvaluateUS(request); got.Code != RefusalCapInvalid {
		t.Fatalf("mutated sealed cap accepted=%+v", got)
	}

	forgedFX := fx
	forgedFX.QuoteID = "invented"
	if validFX(forgedFX) {
		t.Fatal("mutated sealed FX remained valid")
	}
	if _, err := BuildCampaignPlan(PlanRequest{LaneID: KRContinuationLaneID, LaneVersion: LaneVersionV1, Market: MarketKR,
		AccountRef: "acct", Symbol: "005930", CampaignID: "campaign", PositionGeneration: 1,
		RiskBudgetMinor: "100", PerShareRiskMinor: "10", PlannedQuantity: 14, PolicyDigest: "policy", ConfigDigest: "config",
		AccountCurrency: "KRW", QuoteCurrency: "KRW", FX: &fx}); err == nil {
		t.Fatal("same-currency campaign accepted an FX object")
	}
}

func TestRiskAdmissionUsesFilledHeldProposedCheckedBudget(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "100")
	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.Risk.FilledMinor = "40"
	request.Context.Risk.HeldMinor = "30"
	request.Context.Cap = riskCapFor(plan, 20, plan.LegCeilings[0], "30")
	if got := EvaluateKR(request); got.Kind != OutcomeDecision {
		t.Fatalf("exact budget boundary refused: %+v", got)
	}
	request.Context.Cap = riskCapFor(plan, 20, plan.LegCeilings[0], "31")
	if got := EvaluateKR(request); got.Kind != OutcomeRefusal || got.Code != RefusalRiskBudgetExceeded {
		t.Fatalf("budget excess=%+v", got)
	}
	request.Context.Risk.HeldMinor = "1" + string(make([]byte, 100))
	if got := EvaluateKR(request); got.Code != RefusalArithmeticOverflow {
		t.Fatalf("invalid/overflow usage=%+v", got)
	}
}

func TestActualRiskUsesTransferredFloorAndFrozenUSFX(t *testing.T) {
	fx := validUSFX()
	plan := mustPlan(t, MarketUS, USContinuationLaneID, "KRW", "USD", &fx, 14, "1000")
	event := scopedFillRiskEvent(plan, FillRiskEvent{FillID: "fill-1", Quantity: 2, TransferredReservationMinor: "60", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "3", EstimatedExitFeesLeviesMinor: "2", FX: &fx})
	risk, known := CalculateActualRisk(plan, event)
	if !known || risk != "60" {
		t.Fatalf("actual risk=%s known=%v, want transferred floor 60", risk, known)
	}
	event.TransferredReservationMinor = "20"
	risk, known = CalculateActualRisk(plan, event)
	if !known || risk != "55" {
		t.Fatalf("FX-converted risk=%s known=%v, want ceil(25*2*1.1)=55", risk, known)
	}
	changed := fx
	changed.QuoteID = "new-favorable-quote"
	changed.RateQuoteToAccount = "1"
	event.FX = &changed
	if risk, known := CalculateActualRisk(plan, event); known || risk != "" {
		t.Fatalf("mixed FX snapshot accepted: %s/%v", risk, known)
	}
}

func TestFillRiskDuplicateCancelOverageAndUnknownNeverDropFill(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "100")
	state := NewRiskState(plan)
	state.HeldMinor = "60"
	event := scopedFillRiskEvent(plan, FillRiskEvent{FillID: "fill-1", Quantity: 2, TransferredReservationMinor: "40", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "3", EstimatedExitFeesLeviesMinor: "2"})
	next, result := ApplyFillRisk(state, plan, event)
	if !result.Applied || result.Duplicate || next.FilledMinor != "40" || next.HeldMinor != "20" {
		t.Fatalf("first fill state=%+v result=%+v", next, result)
	}
	retry, result := ApplyFillRisk(next, plan, event)
	if !result.Duplicate || !reflect.DeepEqual(retry, next) {
		t.Fatalf("duplicate moved risk: %+v %+v", retry, result)
	}
	cancel := scopedCancelRiskEvent(plan, CancelRiskEvent{CancelID: "cancel-1", ReleaseHeldMinor: "20"})
	cancelled, result := ApplyCancelRisk(next, plan, cancel)
	if !result.Applied || cancelled.HeldMinor != "0" || cancelled.FilledMinor != "40" {
		t.Fatalf("cancel did not release only unfilled held: %+v %+v", cancelled, result)
	}
	if retryCancel, result := ApplyCancelRisk(cancelled, plan, cancel); !result.Duplicate || !reflect.DeepEqual(retryCancel, cancelled) {
		t.Fatal("cancel retry was not idempotent")
	}

	overage := event
	overage.FillID = "fill-overage"
	overage.TransferredReservationMinor = "20"
	overage.EntryPriceMinor = "200"
	overage.EffectiveStopMinor = "100"
	overState, result := ApplyFillRisk(next, plan, overage)
	if !result.Applied || !overState.Latches[LatchCampaignRiskOverage] || overState.Fills[overage.FillID].RiskMinor == "" {
		t.Fatalf("overage fill was dropped or unlatchd: %+v %+v", overState, result)
	}

	unknown := event
	unknown.FillID = "fill-unknown"
	unknown.EntryFeesMinor = ""
	unknownState, result := ApplyFillRisk(state, plan, unknown)
	if !result.Applied || !unknownState.Latches[LatchUnknownActualRisk] || !unknownState.Fills[unknown.FillID].Applied {
		t.Fatalf("unknown-risk fill was dropped: %+v %+v", unknownState, result)
	}
}

func TestMissingFillIdentityUsesPreimageDigestAndCorruptUsageNeverRetreats(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	event := FillRiskEvent{CampaignID: plan.CampaignID, LegOrdinal: 1, OrderRef: "order-1", Quantity: 1, TransferredReservationMinor: "10", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0", ObservedAt: "2026-08-04T00:00:03Z", SourceDigest: "fill-observation-1"}
	state := NewRiskState(plan)
	state.HeldMinor = "100"
	first, firstResult := ApplyFillRisk(state, plan, event)
	retry, retryResult := ApplyFillRisk(first, plan, event)
	if !firstResult.Applied || firstResult.Duplicate || retryResult.Applied || !retryResult.Duplicate || !reflect.DeepEqual(retry, first) || len(retry.Fills) != 1 || !retry.Latches[LatchUnknownActualRisk] {
		t.Fatalf("unidentified fill retry was not digest-idempotent: first=%+v retry=%+v results=%+v/%+v", first, retry, firstResult, retryResult)
	}
	different := event
	different.SourceDigest = "fill-observation-2"
	second, secondResult := ApplyFillRisk(first, plan, different)
	if !secondResult.Applied || secondResult.Duplicate || len(second.Fills) != 2 {
		t.Fatalf("distinct unidentified fill preimages coalesced: second=%+v result=%+v", second, secondResult)
	}

	corrupt := NewRiskState(plan)
	corrupt.FilledMinor = "corrupt-prior-usage"
	corrupt.HeldMinor = strings.Repeat("9", maxMinorDecimalDigits+1)
	next, result := ApplyFillRisk(corrupt, plan, scopedFillRiskEvent(plan, FillRiskEvent{FillID: "fill-with-corrupt-state", Quantity: 1, TransferredReservationMinor: "10", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0"}))
	if !result.Applied || next.FilledMinor != corrupt.FilledMinor || next.HeldMinor != corrupt.HeldMinor || !next.Latches[LatchUnknownActualRisk] {
		t.Fatalf("corrupt prior exposure retreated: before=%+v after=%+v result=%+v", corrupt, next, result)
	}
}

func TestFillAccountingValidationIsAtomic(t *testing.T) {
	const maxUint256 = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, maxUint256)
	base := scopedFillRiskEvent(plan, FillRiskEvent{FillID: "atomic-fill", Quantity: 1, TransferredReservationMinor: "10", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0"})
	tests := []struct {
		name  string
		state RiskState
		event FillRiskEvent
	}{
		{name: "transferred exceeds held", state: func() RiskState { s := NewRiskState(plan); s.FilledMinor, s.HeldMinor = "7", "9"; return s }(), event: base},
		{name: "corrupt held", state: func() RiskState { s := NewRiskState(plan); s.FilledMinor, s.HeldMinor = "7", "corrupt"; return s }(), event: base},
		{name: "corrupt filled", state: func() RiskState { s := NewRiskState(plan); s.FilledMinor, s.HeldMinor = "corrupt", "20"; return s }(), event: base},
		{name: "filled plus risk overflow", state: func() RiskState { s := NewRiskState(plan); s.FilledMinor, s.HeldMinor = maxUint256, "20"; return s }(), event: base},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next, result := ApplyFillRisk(test.state, plan, test.event)
			if !result.Applied || result.Duplicate || next.HeldMinor != test.state.HeldMinor || next.FilledMinor != test.state.FilledMinor || !next.Latches[LatchUnknownActualRisk] || len(next.Fills) != 1 || !next.Fills[test.event.FillID].Applied {
				t.Fatalf("invalid accounting partially committed: before=%+v after=%+v result=%+v", test.state, next, result)
			}
			retry, retryResult := ApplyFillRisk(next, plan, test.event)
			if retryResult.Applied || !retryResult.Duplicate || !reflect.DeepEqual(retry, next) {
				t.Fatalf("atomic failure retry was not idempotent: retry=%+v result=%+v", retry, retryResult)
			}
		})
	}
}

func TestMissingCancelIdentityUsesPreimageDigestWithoutReleasingHeldRisk(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	state := NewRiskState(plan)
	state.HeldMinor = "100"
	event := CancelRiskEvent{CampaignID: plan.CampaignID, LegOrdinal: 1, OrderRef: "order-1", ReleaseHeldMinor: "20", ObservedAt: "2026-08-04T00:00:03Z", SourceDigest: "cancel-observation-1"}
	first, firstResult := ApplyCancelRisk(state, plan, event)
	retry, retryResult := ApplyCancelRisk(first, plan, event)
	if !firstResult.Applied || firstResult.Duplicate || retryResult.Applied || !retryResult.Duplicate || !reflect.DeepEqual(retry, first) || first.HeldMinor != "100" || len(retry.Cancels) != 1 || !retry.Latches[LatchUnknownActualRisk] {
		t.Fatalf("unidentified cancel retry moved/coalesced risk incorrectly: first=%+v retry=%+v results=%+v/%+v", first, retry, firstResult, retryResult)
	}
	different := event
	different.SourceDigest = "cancel-observation-2"
	second, secondResult := ApplyCancelRisk(first, plan, different)
	if !secondResult.Applied || secondResult.Duplicate || second.HeldMinor != "100" || len(second.Cancels) != 2 {
		t.Fatalf("distinct unidentified cancel preimages coalesced/released risk: second=%+v result=%+v", second, secondResult)
	}
}

func TestZeroQuantityFillIsNonAppliedEvidenceAndNeverMovesRisk(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	state := NewRiskState(plan)
	state.FilledMinor = "7"
	state.HeldMinor = "100"
	event := FillRiskEvent{FillID: "zero-fill", CampaignID: plan.CampaignID, LegOrdinal: 1, OrderRef: "order-1", Quantity: 0, TransferredReservationMinor: "20", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0", ObservedAt: "2026-08-04T00:00:03Z", SourceDigest: "zero-fill-observation"}

	next, result := ApplyFillRisk(state, plan, event)
	if result.Applied || result.Duplicate || next.Fills[event.FillID].Applied || next.FilledMinor != "7" || next.HeldMinor != "100" || !next.Latches[LatchUnknownActualRisk] {
		t.Fatalf("zero fill applied or moved risk: next=%+v result=%+v", next, result)
	}
	retry, retryResult := ApplyFillRisk(next, plan, event)
	if retryResult.Applied || !retryResult.Duplicate || !reflect.DeepEqual(retry, next) {
		t.Fatalf("zero fill retry was not idempotent: retry=%+v result=%+v", retry, retryResult)
	}

	positive := scopedFillRiskEvent(plan, FillRiskEvent{FillID: "shared-fill-id", Quantity: 1, TransferredReservationMinor: "20", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0"})
	positiveState, positiveResult := ApplyFillRisk(state, plan, positive)
	if !positiveResult.Applied {
		t.Fatal("positive baseline fill was not applied")
	}
	zeroConflict := positive
	zeroConflict.Quantity = 0
	conflicted, conflictResult := ApplyFillRisk(positiveState, plan, zeroConflict)
	if conflictResult.Applied || conflictResult.Duplicate || conflicted.HeldMinor != positiveState.HeldMinor || conflicted.FilledMinor != positiveState.FilledMinor || !conflicted.Latches[LatchUnknownActualRisk] || len(conflicted.Fills) != 2 {
		t.Fatalf("zero FillID conflict synthesized positive accounting: conflicted=%+v result=%+v", conflicted, conflictResult)
	}
}

func TestRiskCapCannotReplayAcrossExactPlanIdentity(t *testing.T) {
	first := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	request := PlanRequest{LaneID: KRContinuationLaneID, LaneVersion: LaneVersionV1, Market: MarketKR, AccountRef: first.AccountRef, Symbol: first.Symbol,
		CampaignID: "campaign-KR-peer", PositionGeneration: first.PositionGeneration, RiskBudgetMinor: first.RiskBudgetMinor, PerShareRiskMinor: first.PerShareRiskMinor,
		PlannedQuantity: first.PlannedQuantity, PolicyDigest: first.PolicyDigest, ConfigDigest: first.ConfigDigest, AccountCurrency: first.AccountCurrency, QuoteCurrency: first.QuoteCurrency}
	second, err := BuildCampaignPlan(request)
	if err != nil {
		t.Fatal(err)
	}
	evidence, config := validKRFixture()
	evaluation := validKREvaluation(t, second, evidence, config)
	evaluation.Context.Cap = validRiskCap(first)
	if got := EvaluateKR(evaluation); got.Code != RefusalCapInvalid || got.Quantity != 0 {
		t.Fatalf("cap replayed across exact plan identity: %+v", got)
	}
}

func TestForeignOrIncompleteRiskEventsPreserveAccountingAndLatch(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	state := NewRiskState(plan)
	state.FilledMinor = "7"
	state.HeldMinor = "100"
	baseFill := scopedFillRiskEvent(plan, FillRiskEvent{FillID: "fill-scope", Quantity: 1, TransferredReservationMinor: "20", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0"})
	fillCases := map[string]func(*FillRiskEvent){
		"foreign campaign": func(event *FillRiskEvent) { event.CampaignID = "other-campaign" },
		"invalid leg":      func(event *FillRiskEvent) { event.LegOrdinal = 99 },
		"empty order":      func(event *FillRiskEvent) { event.OrderRef = "" },
		"empty source":     func(event *FillRiskEvent) { event.SourceDigest = "" },
		"invalid time":     func(event *FillRiskEvent) { event.ObservedAt = "not-a-time" },
		"oversized order":  func(event *FillRiskEvent) { event.OrderRef = strings.Repeat("o", 257) },
		"oversized source": func(event *FillRiskEvent) { event.SourceDigest = strings.Repeat("s", 257) },
	}
	for name, edit := range fillCases {
		t.Run("fill "+name, func(t *testing.T) {
			event := baseFill
			edit(&event)
			next, result := ApplyFillRisk(state, plan, event)
			if result.Applied || result.Duplicate || next.FilledMinor != state.FilledMinor || next.HeldMinor != state.HeldMinor || !next.Latches[LatchUnknownActualRisk] || len(next.Fills) != 1 {
				t.Fatalf("unscoped fill changed accounting: next=%+v result=%+v", next, result)
			}
			retry, retryResult := ApplyFillRisk(next, plan, event)
			if retryResult.Applied || !retryResult.Duplicate || !reflect.DeepEqual(retry, next) {
				t.Fatalf("unscoped fill retry was not idempotent: retry=%+v result=%+v", retry, retryResult)
			}
		})
	}

	baseCancel := scopedCancelRiskEvent(plan, CancelRiskEvent{CancelID: "cancel-scope", ReleaseHeldMinor: "20"})
	cancelCases := map[string]func(*CancelRiskEvent){
		"foreign campaign": func(event *CancelRiskEvent) { event.CampaignID = "other-campaign" },
		"invalid leg":      func(event *CancelRiskEvent) { event.LegOrdinal = 99 },
		"empty order":      func(event *CancelRiskEvent) { event.OrderRef = "" },
		"empty source":     func(event *CancelRiskEvent) { event.SourceDigest = "" },
		"invalid time":     func(event *CancelRiskEvent) { event.ObservedAt = "not-a-time" },
		"oversized order":  func(event *CancelRiskEvent) { event.OrderRef = strings.Repeat("o", 257) },
		"oversized source": func(event *CancelRiskEvent) { event.SourceDigest = strings.Repeat("s", 257) },
	}
	for name, edit := range cancelCases {
		t.Run("cancel "+name, func(t *testing.T) {
			event := baseCancel
			edit(&event)
			next, result := ApplyCancelRisk(state, plan, event)
			if result.Applied || result.Duplicate || next.FilledMinor != state.FilledMinor || next.HeldMinor != state.HeldMinor || !next.Latches[LatchUnknownActualRisk] || len(next.Cancels) != 1 {
				t.Fatalf("unscoped cancel changed accounting: next=%+v result=%+v", next, result)
			}
			retry, retryResult := ApplyCancelRisk(next, plan, event)
			if retryResult.Applied || !retryResult.Duplicate || !reflect.DeepEqual(retry, next) {
				t.Fatalf("unscoped cancel retry was not idempotent: retry=%+v result=%+v", retry, retryResult)
			}
		})
	}

	longRequest := PlanRequest{LaneID: KRContinuationLaneID, LaneVersion: LaneVersionV1, Market: MarketKR, AccountRef: "acct", Symbol: "005930",
		CampaignID: strings.Repeat("c", 257), PositionGeneration: 1, RiskBudgetMinor: "1000", PerShareRiskMinor: "10", PlannedQuantity: 14,
		PolicyDigest: "risk-policy", ConfigDigest: "kr-config", AccountCurrency: "KRW", QuoteCurrency: "KRW"}
	longPlan, err := BuildCampaignPlan(longRequest)
	if err != nil {
		t.Fatal(err)
	}
	longState := NewRiskState(longPlan)
	longState.HeldMinor = "100"
	longFill := scopedFillRiskEvent(longPlan, FillRiskEvent{FillID: "long-campaign-fill", Quantity: 1, TransferredReservationMinor: "10", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0"})
	if next, result := ApplyFillRisk(longState, longPlan, longFill); result.Applied || next.HeldMinor != longState.HeldMinor || !next.Latches[LatchUnknownActualRisk] {
		t.Fatalf("oversized matching campaign scope accepted: next=%+v result=%+v", next, result)
	}
	longCancel := scopedCancelRiskEvent(longPlan, CancelRiskEvent{CancelID: "long-campaign-cancel", ReleaseHeldMinor: "10"})
	if next, result := ApplyCancelRisk(longState, longPlan, longCancel); result.Applied || next.HeldMinor != longState.HeldMinor || !next.Latches[LatchUnknownActualRisk] {
		t.Fatalf("oversized matching cancel campaign scope accepted: next=%+v result=%+v", next, result)
	}
}

func TestZeroQuantityPlanAndUnboundedRiskInputFailClosed(t *testing.T) {
	if _, err := BuildCampaignPlan(PlanRequest{LaneID: KRContinuationLaneID, LaneVersion: LaneVersionV1, Market: MarketKR,
		AccountRef: "acct", Symbol: "005930", CampaignID: "campaign", PositionGeneration: 1,
		RiskBudgetMinor: "100", PerShareRiskMinor: "10", PlannedQuantity: 0, PolicyDigest: "policy", ConfigDigest: "config",
		AccountCurrency: "KRW", QuoteCurrency: "KRW"}); err == nil {
		t.Fatal("zero planned quantity accepted")
	}
	if _, err := parseUnboundedUnsigned(strings.Repeat("9", maxMinorDecimalDigits+1)); err == nil {
		t.Fatal("oversized persisted integer accepted")
	}
}

func TestStopCandidateTimeMustBeValidAndNotAfterEvaluation(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.StopCandidate.ObservedAt = "not-a-time"
	if got := EvaluateKR(request); got.Code != RefusalStopInvalid {
		t.Fatalf("invalid stop time accepted=%+v", got)
	}
	request.Context.StopCandidate.ObservedAt = "2026-08-04T00:00:04Z"
	if got := EvaluateKR(request); got.Code != RefusalStopInvalid {
		t.Fatalf("future stop observation accepted=%+v", got)
	}
	request = validKREvaluation(t, plan, evidence, config)
	request.Context.StopCandidate = mustStopCandidate(t, "95", "2025-08-04T00:00:00Z", "2025-08-04T00:01:00Z")
	if got := EvaluateKR(request); got.Code != RefusalStopInvalid {
		t.Fatalf("stale but well-formed stop provenance accepted=%+v", got)
	}
	request.Context.StopCandidate = mustStopCandidate(t, "95", "2026-08-04T00:00:02Z", evidence.Envelope.EvaluatedAt)
	if got := EvaluateKR(request); got.Kind != OutcomeDecision {
		t.Fatalf("inclusive stop freshness boundary refused=%+v", got)
	}
	request.Context.StopCandidate.Digest = "mutated"
	if got := EvaluateKR(request); got.Code != RefusalStopInvalid {
		t.Fatalf("post-seal stop mutation accepted=%+v", got)
	}
}

func TestInvalidationSuppressesAddAndLeavesCommonExitIndependent(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.Invalidation = Invalidation{Structural: true, Code: "STRUCTURE_BROKEN"}
	got := EvaluateKR(request)
	if got.Kind != OutcomeInvalidation || got.Quantity != 0 || !got.CommonExitIndependent || got.ExitDecisionCreated {
		t.Fatalf("invalidation authority leak=%+v", got)
	}
}

func TestInvalidationWithoutTypedCodeIsRefused(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	evidence, config := validKRFixture()
	request := validKREvaluation(t, plan, evidence, config)
	request.Context.Invalidation = Invalidation{Structural: true}
	got := EvaluateKR(request)
	if got.Kind != OutcomeRefusal || got.Code != RefusalInvalidationInvalid || got.Quantity != 0 || got.ExitDecisionCreated {
		t.Fatalf("untyped invalidation accepted=%+v", got)
	}
}

func mustPlan(t *testing.T, market Market, laneID, accountCurrency, quoteCurrency string, fx *FrozenFX, q uint64, budget string) CampaignPlan {
	t.Helper()
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: laneID, LaneVersion: LaneVersionV1, Market: market, AccountRef: "acct", Symbol: map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market], CampaignID: "campaign-" + string(market), PositionGeneration: 1, RiskBudgetMinor: budget, PerShareRiskMinor: "10", PlannedQuantity: q, PolicyDigest: "risk-policy", ConfigDigest: map[Market]string{MarketKR: "kr-config", MarketUS: "us-config"}[market], AccountCurrency: accountCurrency, QuoteCurrency: quoteCurrency, FX: fx})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func validRiskCap(plan CampaignPlan) RiskCap {
	ceiling := plan.LegCeilings[0]
	return riskCapFor(plan, 20, ceiling, "20")
}

func riskCapFor(plan CampaignPlan, qFinal, reservationQuantity uint64, reservationMinor string) RiskCap {
	input := riskCapInput{QFinal: qFinal, ReservationQuantity: reservationQuantity, ReservationMinor: reservationMinor, SnapshotID: "a066-snapshot", PolicyDigest: plan.PolicyDigest, BucketSetDigest: "buckets", ObservedAt: "2026-08-04T00:00:00Z", FreshUntil: "2026-08-04T00:01:00Z"}
	if plan.FX != nil {
		fx := *plan.FX
		input.FX = &fx
	}
	cap, err := newRiskCap(plan, input)
	if err != nil {
		panic(err)
	}
	return cap
}

func scopedFillRiskEvent(plan CampaignPlan, event FillRiskEvent) FillRiskEvent {
	event.CampaignID = plan.CampaignID
	event.LegOrdinal = 1
	event.OrderRef = "order-1"
	event.ObservedAt = "2026-08-04T00:00:03Z"
	event.SourceDigest = "fill-source"
	return event
}

func scopedCancelRiskEvent(plan CampaignPlan, event CancelRiskEvent) CancelRiskEvent {
	event.CampaignID = plan.CampaignID
	event.LegOrdinal = 1
	event.OrderRef = "order-1"
	event.ObservedAt = "2026-08-04T00:00:03Z"
	event.SourceDigest = "cancel-source"
	return event
}

func validContext(plan CampaignPlan) EvaluationContext {
	stop, err := newStopCandidate(stopCandidateInput{PriceMinor: "95", Source: "risk", Policy: "stop-v1", Version: "v1", Digest: "stop-digest", ObservedAt: "2026-08-04T00:00:02Z", FreshUntil: "2026-08-04T00:01:00Z"})
	if err != nil {
		panic(err)
	}
	return EvaluationContext{Enabled: true, CandidateID: "candidate", Plan: plan, Leg: LegProgress{Ordinal: 1}, Cap: validRiskCap(plan), Risk: NewRiskState(plan),
		SavedEffectiveStopMinor: "90", StopCandidate: stop}
}

func mustStopCandidate(t *testing.T, price, observedAt, freshUntil string) StopCandidate {
	t.Helper()
	stop, err := newStopCandidate(stopCandidateInput{PriceMinor: price, Source: "risk", Policy: "stop-v1", Version: "v1", Digest: "stop-digest", ObservedAt: observedAt, FreshUntil: freshUntil})
	if err != nil {
		t.Fatal(err)
	}
	return stop
}

func validKREvaluation(t *testing.T, plan CampaignPlan, evidence KREvidence, config KRFlowConfig) KREvaluationRequest {
	t.Helper()
	context := validContext(plan)
	terms, err := mintExecutionTermsPreimage(plan, evidence.Envelope, "110", "130")
	if err != nil {
		t.Fatal(err)
	}
	context.ExecutionTerms, context.SavedStopProvenance = terms, mintSavedStopProvenance(plan, evidence.Envelope, context.SavedEffectiveStopMinor)
	return KREvaluationRequest{Context: context, Evidence: evidence, Config: config}
}

func validUSEvaluation(t *testing.T, plan CampaignPlan, evidence USEvidence, config USParticipationConfig) USEvaluationRequest {
	t.Helper()
	context := validContext(plan)
	terms, err := mintExecutionTermsPreimage(plan, evidence.Envelope, "110", "130")
	if err != nil {
		t.Fatal(err)
	}
	context.ExecutionTerms, context.SavedStopProvenance = terms, mintSavedStopProvenance(plan, evidence.Envelope, context.SavedEffectiveStopMinor)
	return USEvaluationRequest{Context: context, Evidence: evidence, Config: config}
}

func validUSFX() FrozenFX {
	fx, err := newFrozenFX(frozenFXInput{QuoteID: "fx-1", AsOf: "2026-08-04T00:00:00Z", Direction: FXQuoteToAccount, RateQuoteToAccount: "2", Haircut: "1.1", Digest: "fx-digest"})
	if err != nil {
		panic(err)
	}
	return fx
}
