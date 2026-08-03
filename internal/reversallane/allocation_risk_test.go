package reversallane

import (
	"reflect"
	"testing"
	"time"
)

func TestImmutableTwoFourEightAllocationUsesFloorAndFinalRemainder(t *testing.T) {
	for _, tc := range []struct {
		q    uint64
		want [3]uint64
	}{
		{1, [3]uint64{0, 0, 1}},
		{7, [3]uint64{1, 2, 4}},
		{14, [3]uint64{2, 4, 8}},
		{15, [3]uint64{2, 4, 9}},
		{^uint64(0), [3]uint64{2635249153387078802, 5270498306774157604, 10540996613548315209}},
	} {
		plan := mustPlan(t, MarketKR, KRReversalLaneID, tc.q, "100000000000000000000000000000", "KRW", "KRW", nil)
		if got := plan.LegCeilings(); got != tc.want || got[0]+got[1]+got[2] != tc.q {
			t.Fatalf("Q=%d got=%v want=%v", tc.q, got, tc.want)
		}
		if plan.PlannedQuantity() != tc.q || plan.Digest() == "" {
			t.Fatalf("plan not immutable/bound: Q=%d", tc.q)
		}
	}
}

func TestPartialCancelRetryCannotReallocateAndA066CapBinds(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	if got := PlannedLegQuantity(plan, LegProgress{Ordinal: 2}, RiskCap{QFinal: 3, ReservationMinor: "20", Official: true, Frozen: true}); got != 3 {
		t.Fatalf("a066 cap=%d", got)
	}
	if got := PlannedLegQuantity(plan, LegProgress{Ordinal: 2}, RiskCap{QFinal: 99, ReservationMinor: "20", Official: true, Frozen: true}); got != 4 {
		t.Fatalf("unused first leg reallocated=%d", got)
	}
	if got := PlannedLegQuantity(plan, LegProgress{Ordinal: 2, FilledQuantity: 1}, RiskCap{QFinal: 99, ReservationMinor: "20", Official: true, Frozen: true}); got != 3 {
		t.Fatalf("partial remaining=%d", got)
	}
	for _, terminal := range []LegProgress{{Ordinal: 2, Cancelled: true}, {Ordinal: 2, Expired: true}} {
		if got := PlannedLegQuantity(plan, terminal, RiskCap{QFinal: 99, ReservationMinor: "20", Official: true, Frozen: true}); got != 0 {
			t.Fatalf("terminal leg reallocated=%d", got)
		}
	}
}

func TestRiskAdmissionUsesCheckedFilledHeldProposedBudget(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "100", "KRW", "KRW", nil)
	state := NewRiskState(plan)
	state.FilledMinor = "40"
	state.HeldMinor = "30"
	if refusal := AdmitRisk(plan, state, mustRiskCapWithReservation(t, plan, "30")); refusal != "" {
		t.Fatalf("exact budget=%s", refusal)
	}
	if refusal := AdmitRisk(plan, state, mustRiskCapWithReservation(t, plan, "31")); refusal != RefusalRiskBudgetExceeded {
		t.Fatalf("budget excess=%s", refusal)
	}
	state.HeldMinor = "not-a-number"
	if refusal := AdmitRisk(plan, state, mustRiskCapWithReservation(t, plan, "1")); refusal != RefusalArithmeticInvalid {
		t.Fatalf("invalid risk=%s", refusal)
	}
}

func TestActualRiskUsesTransferredFloorAndFrozenUSFXCeil(t *testing.T) {
	fx := validFX()
	plan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "KRW", "USD", &fx)
	event := validFillEvent(plan, "fill-1")
	event.TransferredReservationMinor = "60"
	event.FX = &fx
	if risk, known := CalculateActualRisk(plan, event); !known || risk != "60" {
		t.Fatalf("floor risk=%s known=%v", risk, known)
	}
	event.TransferredReservationMinor = "20"
	if risk, known := CalculateActualRisk(plan, event); !known || risk != "55" {
		t.Fatalf("ceil FX risk=%s known=%v", risk, known)
	}
	changed := fx
	changed.QuoteID = "new-quote"
	event.FX = &changed
	if risk, known := CalculateActualRisk(plan, event); known || risk != "" {
		t.Fatalf("mixed FX accepted=%s/%v", risk, known)
	}
}

func TestActualRiskRequiresPlanScopedFillProvenanceAndFreshFrozenFX(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "1000", "KRW", "KRW", nil)
	for name, edit := range map[string]func(*FillRiskEvent){
		"campaign": func(event *FillRiskEvent) { event.CampaignID = "other-campaign" },
		"leg":      func(event *FillRiskEvent) { event.LegOrdinal = 4 },
		"order":    func(event *FillRiskEvent) { event.OrderRef = "" },
		"observed": func(event *FillRiskEvent) { event.ObservedAt = time.Time{} },
		"source":   func(event *FillRiskEvent) { event.SourceDigest = "" },
	} {
		t.Run(name, func(t *testing.T) {
			event := validFillEvent(plan, "fill-1")
			edit(&event)
			if risk, known := CalculateActualRisk(plan, event); known || risk != "" {
				t.Fatalf("unscoped fill accepted: risk=%q known=%v event=%+v", risk, known, event)
			}
		})
	}

	staleFX, err := mintFrozenFX(FrozenFX{Authority: a066FXAuthority, Version: "a066-fx-v1", QuoteID: "fx-stale", AsOf: "2026-08-04T00:00:00Z", Direction: FXQuoteToAccount,
		RateQuoteToAccount: "2", Haircut: "1.1", Digest: "fx-stale-digest", Official: true, Frozen: true, FreshUntil: "2026-08-04T00:00:03Z"})
	if err != nil {
		t.Fatal(err)
	}
	usPlan := mustPlan(t, MarketUS, USReversalLaneID, 14, "1000", "KRW", "USD", &staleFX)
	event := validFillEvent(usPlan, "fill-us")
	event.FX = &staleFX
	if risk, known := CalculateActualRisk(usPlan, event); known || risk != "" {
		t.Fatalf("stale fill FX accepted: risk=%q known=%v", risk, known)
	}
}

func TestFillDuplicateCancelOverageUnknownPreserveFillAndLatch(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "100", "KRW", "KRW", nil)
	state := NewRiskState(plan)
	state.HeldMinor = "60"
	event := validFillEvent(plan, "fill-1")
	next, result := ApplyFillRisk(state, plan, event)
	if !result.Applied || result.Duplicate || next.FilledMinor != "40" || next.HeldMinor != "20" {
		t.Fatalf("fill=%+v result=%+v", next, result)
	}
	retry, result := ApplyFillRisk(next, plan, event)
	if !result.Duplicate || !reflect.DeepEqual(retry, next) {
		t.Fatalf("duplicate moved risk=%+v/%+v", retry, result)
	}
	cancelled, result := ApplyCancelRisk(next, CancelRiskEvent{CancelID: "cancel-1", ReleaseHeldMinor: "20"})
	if !result.Applied || cancelled.HeldMinor != "0" || cancelled.FilledMinor != "40" {
		t.Fatalf("cancel=%+v/%+v", cancelled, result)
	}
	if duplicate, result := ApplyCancelRisk(cancelled, CancelRiskEvent{CancelID: "cancel-1", ReleaseHeldMinor: "20"}); !result.Duplicate || !reflect.DeepEqual(duplicate, cancelled) {
		t.Fatal("duplicate cancel moved state")
	}

	overage := event
	overage.FillID = "fill-overage"
	overage.TransferredReservationMinor = "90"
	overage.EntryPriceMinor = "200"
	overage.EffectiveStopMinor = "100"
	overageBase := next
	overageBase.HeldMinor = "110"
	overState, result := ApplyFillRisk(overageBase, plan, overage)
	if !result.Applied || !overState.Latches[LatchCampaignRiskOverage] || !overState.Fills[overage.FillID].Applied {
		t.Fatalf("overage dropped=%+v/%+v", overState, result)
	}

	unknown := event
	unknown.FillID = "fill-unknown"
	unknown.EntryFeesMinor = ""
	unknownState, result := ApplyFillRisk(state, plan, unknown)
	if !result.Applied || !unknownState.Latches[LatchUnknownActualRisk] || !unknownState.Fills[unknown.FillID].Applied {
		t.Fatalf("unknown fill dropped=%+v/%+v", unknownState, result)
	}
}

func mustPlan(t *testing.T, market Market, laneID string, q uint64, budget, accountCurrency, quoteCurrency string, fx *FrozenFX) CampaignPlan {
	t.Helper()
	plan, err := BuildCampaignPlan(validPlanRequest(market, laneID, q, budget, accountCurrency, quoteCurrency, fx))
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func validPlanRequest(market Market, laneID string, q uint64, budget, accountCurrency, quoteCurrency string, fx *FrozenFX) PlanRequest {
	return PlanRequest{LaneID: laneID, LaneVersion: LaneVersionV1, Market: market,
		AccountRef: "acct", Symbol: map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market], CampaignID: "campaign-" + string(market), PositionGeneration: 1,
		RiskBudgetMinor: budget, PerShareRiskMinor: "10", PlannedQuantity: q, PolicyDigest: "a066-policy", ConfigDigest: map[Market]string{MarketKR: "kr-config-digest", MarketUS: "us-config-digest"}[market],
		AccountCurrency: accountCurrency, QuoteCurrency: quoteCurrency, FX: fx}
}

func validFillEvent(plan CampaignPlan, fillID string) FillRiskEvent {
	return FillRiskEvent{
		FillID:                       fillID,
		CampaignID:                   plan.CampaignID(),
		LegOrdinal:                   1,
		OrderRef:                     "order-1",
		Quantity:                     2,
		TransferredReservationMinor:  "40",
		EntryPriceMinor:              "100",
		EffectiveStopMinor:           "90",
		EntryFeesMinor:               "3",
		EstimatedExitFeesLeviesMinor: "2",
		ObservedAt:                   time.Date(2026, 8, 4, 0, 0, 4, 0, time.UTC),
		SourceDigest:                 "fill-source-digest",
	}
}

func mustRiskCapWithReservation(t *testing.T, plan CampaignPlan, reservation string) RiskCap {
	t.Helper()
	evaluated := time.Date(2026, 8, 4, 0, 0, 3, 0, time.UTC)
	cap, err := mintRiskCap(plan, RiskCap{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: plan.Market(), QFinal: 20, ReservationQuantity: 1, ReservationMinor: reservation,
		SnapshotID: "a066-snapshot", PolicyDigest: "a066-policy", BucketSetDigest: "bucket-digest", Official: true, Frozen: true,
		ObservedAt: evaluated.Add(-time.Second), FreshUntil: evaluated.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	return cap
}

func validFX() FrozenFX {
	fx, err := mintFrozenFX(FrozenFX{Authority: a066FXAuthority, Version: "a066-fx-v1", QuoteID: "fx-1", AsOf: "2026-08-04T00:00:00Z", Direction: FXQuoteToAccount,
		RateQuoteToAccount: "2", Haircut: "1.1", Digest: "fx-digest", Official: true, Frozen: true, FreshUntil: "2026-08-04T00:01:00Z"})
	if err != nil {
		panic(err)
	}
	return fx
}

func ptrFX(value FrozenFX) *FrozenFX { return &value }

func validContext(plan CampaignPlan, ordinal int) EvaluationContext {
	evaluated := time.Date(2026, 8, 4, 0, 0, 3, 0, time.UTC)
	terms, err := NewExecutionTermsPreimage(plan, "110", "130")
	if err != nil {
		panic(err)
	}
	return EvaluationContext{Enabled: true, CandidateID: "candidate", Plan: plan, Leg: LegProgress{Ordinal: ordinal},
		Cap:  mustRiskCapNoTest(plan, ordinal, evaluated),
		Risk: NewRiskState(plan), SavedEffectiveStopMinor: "90", StopCandidate: StopCandidate{PriceMinor: "95", Valid: true, Source: "risk", Policy: "stop-v1", Digest: "stop-digest", ObservedAt: evaluated},
		ExecutionTerms: terms}
}

func mustRiskCapNoTest(plan CampaignPlan, ordinal int, evaluated time.Time) RiskCap {
	reservationQuantity := PlannedLegQuantity(plan, LegProgress{Ordinal: ordinal}, RiskCap{QFinal: 20})
	cap, err := mintRiskCap(plan, RiskCap{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: plan.Market(), QFinal: 20, ReservationQuantity: reservationQuantity, ReservationMinor: "20",
		SnapshotID: "a066-snapshot", PolicyDigest: "a066-policy", BucketSetDigest: "bucket-digest", Official: true, Frozen: true,
		ObservedAt: evaluated.Add(-time.Second), FreshUntil: evaluated.Add(time.Minute)})
	if err != nil {
		panic(err)
	}
	return cap
}
