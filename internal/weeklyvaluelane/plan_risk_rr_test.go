package weeklyvaluelane

import (
	"reflect"
	"testing"
	"time"
)

func TestImmutableSevenLegAllocationAndNoUpwardReallocation(t *testing.T) {
	for _, tc := range []struct {
		q    uint64
		want [7]uint64
	}{{1, [7]uint64{0, 0, 0, 0, 0, 0, 1}}, {14, [7]uint64{2, 2, 2, 2, 2, 2, 2}}, {15, [7]uint64{2, 2, 2, 2, 2, 2, 3}}, {^uint64(0), [7]uint64{2635249153387078802, 2635249153387078802, 2635249153387078802, 2635249153387078802, 2635249153387078802, 2635249153387078802, 2635249153387078803}}} {
		plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, tc.q, "100000000000000000000000000000")
		if got := plan.LegCeilings(); got != tc.want || sumSeven(got) != tc.q {
			t.Fatalf("Q=%d got=%v want=%v", tc.q, got, tc.want)
		}
	}
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 15, "1000")
	if got := PlannedLegQuantity(plan, LegProgress{Ordinal: 7, FilledQuantity: 1}, 99); got != 2 {
		t.Fatalf("unused prior quantity reallocated=%d", got)
	}
	if got := PlannedLegQuantity(plan, LegProgress{Ordinal: 7, FilledQuantity: 1}, 1); got != 1 {
		t.Fatalf("a066 cap bypassed=%d", got)
	}
}

func TestPlanBoundPrivateCapFrozenFXAndCheckedRisk(t *testing.T) {
	fx := validFX()
	plan := mustPlan(t, MarketUS, USWeeklyLaneID, "KRW", "USD", &fx, 14, "100")
	cap := validCap(t, plan, 1, 2, "30")
	if !cap.validAt(plan, time.Date(2026, 8, 4, 0, 0, 3, 0, time.UTC), 2) {
		t.Fatal("valid cap refused")
	}
	peerRequest := validPlanRequest(MarketUS, USWeeklyLaneID, "KRW", "USD", &fx, 14, "100")
	peerRequest.CampaignID = "peer-campaign"
	peer, err := BuildCampaignPlan(peerRequest)
	if err != nil {
		t.Fatal(err)
	}
	if cap.validAt(peer, time.Date(2026, 8, 4, 0, 0, 3, 0, time.UTC), 2) {
		t.Fatal("cap replayed across plans")
	}

	state := mustRiskState(t, plan, "40", "30")
	if code := AdmitRisk(plan, state, cap); code != "" {
		t.Fatalf("exact budget=%s", code)
	}
	tooMuch := validCap(t, plan, 1, 2, "31")
	if code := AdmitRisk(plan, state, tooMuch); code != RefusalRiskBudgetExceeded {
		t.Fatalf("budget excess=%s", code)
	}
}

func TestFillRiskFloorFXDuplicateUnknownAndCorruptStatePreservation(t *testing.T) {
	fx := validFX()
	plan := mustPlan(t, MarketUS, USWeeklyLaneID, "KRW", "USD", &fx, 14, "100")
	state := mustRiskState(t, plan, "0", "60")
	event := validFill(plan, "fill-1")
	event.FX = &fx
	next, result := ApplyFillRisk(state, plan, event)
	if !result.Applied || next.FilledMinor() != "60" || next.HeldMinor() != "0" {
		t.Fatalf("fill=%+v/%+v", next, result)
	}
	retry, result := ApplyFillRisk(next, plan, event)
	if !result.Duplicate || !reflect.DeepEqual(retry, next) {
		t.Fatalf("duplicate moved risk=%+v/%+v", retry, result)
	}
	conflict := event
	conflict.EntryPriceMinor = "101"
	conflicted, result := ApplyFillRisk(next, plan, conflict)
	if !result.Duplicate || !conflicted.Latched(LatchUnknownActualRisk) {
		t.Fatalf("changed duplicate hidden=%+v/%+v", conflicted, result)
	}

	missing := validFill(plan, "")
	missing.FX = &fx
	unknown, result := ApplyFillRisk(state, plan, missing)
	if !result.Applied || !unknown.Latched(LatchUnknownActualRisk) || unknown.FilledMinor() != state.FilledMinor() || unknown.HeldMinor() != state.HeldMinor() {
		t.Fatalf("missing identity moved accounting=%+v/%+v", unknown, result)
	}
	corrupt := state
	corrupt.planDigest = "other"
	preserved, result := ApplyFillRisk(corrupt, plan, event)
	if result.Applied || result.Code != RefusalRiskLatched || !reflect.DeepEqual(preserved, corrupt) {
		t.Fatalf("corrupt state moved=%+v/%+v", preserved, result)
	}
}

func TestExactCappedTargetRRInclusiveBoundaryAndStructuralStopCap(t *testing.T) {
	result := CalculateRR(RRInput{EntryPriceMinor: "100", StagedTargetMinor: "130", FairValueMinor: "120", EffectiveStopMinor: "90", Quantity: 10,
		EntryCostsMinor: "20", EstimatedExitCostsLeviesMinor: "30", MinimumRRPPM: 1_000_000, AccountCurrency: "KRW", QuoteCurrency: "KRW"})
	if !result.Accepted || result.TargetMinor != "120" || result.RewardMinor != "150" || result.RiskMinor != "150" || result.RRPPM != 1_000_000 {
		t.Fatalf("RR boundary=%+v", result)
	}
	below := CalculateRR(RRInput{EntryPriceMinor: "100", StagedTargetMinor: "130", FairValueMinor: "119", EffectiveStopMinor: "90", Quantity: 10,
		EntryCostsMinor: "20", EstimatedExitCostsLeviesMinor: "30", MinimumRRPPM: 1_000_000, AccountCurrency: "KRW", QuoteCurrency: "KRW"})
	if below.Code != RefusalRRThreshold {
		t.Fatalf("capped target inflated=%+v", below)
	}
	invalid := CalculateRR(RRInput{EntryPriceMinor: "100", StagedTargetMinor: "120", FairValueMinor: "120", EffectiveStopMinor: "100", Quantity: 1,
		EntryCostsMinor: "0", EstimatedExitCostsLeviesMinor: "0", MinimumRRPPM: 1, AccountCurrency: "KRW", QuoteCurrency: "KRW"})
	if invalid.Code != RefusalRRInvalid {
		t.Fatalf("zero risk accepted=%+v", invalid)
	}
}

func sumSeven(values [7]uint64) uint64 {
	var sum uint64
	for _, value := range values {
		sum += value
	}
	return sum
}

func validPlanRequest(market Market, laneID, accountCurrency, quoteCurrency string, fx *FrozenFX, q uint64, budget string) PlanRequest {
	return PlanRequest{LaneID: laneID, LaneVersion: LaneVersionV1, Market: market, AccountRef: "acct", Symbol: map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market],
		CampaignID: "campaign-" + string(market), PositionGeneration: 1, RiskBudgetMinor: budget, PerShareRiskMinor: "10", PlannedQuantity: q,
		PolicyDigest: "risk-policy", ConfigDigest: map[Market]string{MarketKR: "model-config-kr", MarketUS: "model-config-us"}[market], AccountCurrency: accountCurrency, QuoteCurrency: quoteCurrency, FX: fx}
}

func mustPlan(t *testing.T, market Market, laneID, accountCurrency, quoteCurrency string, fx *FrozenFX, q uint64, budget string) CampaignPlan {
	t.Helper()
	plan, err := BuildCampaignPlan(validPlanRequest(market, laneID, accountCurrency, quoteCurrency, fx, q, budget))
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func validFX() FrozenFX {
	fx, err := mintFrozenFX(frozenFXInput{Authority: a066FXAuthority, Version: "a066-fx-v1", QuoteID: "fx-1", AsOf: "2026-08-03T00:00:00Z", FreshUntil: "2026-08-11T00:00:00Z", Direction: FXQuoteToAccount, RateQuoteToAccount: "2", Haircut: "1.1", Digest: "fx-digest"})
	if err != nil {
		panic(err)
	}
	return fx
}

func validCap(t *testing.T, plan CampaignPlan, ordinal int, reservationQuantity uint64, reservationMinor string) RiskCap {
	t.Helper()
	cap, err := mintRiskCap(plan, riskCapInput{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", QFinal: 20, ReservationQuantity: reservationQuantity,
		ReservationMinor: reservationMinor, MaxStopDistanceMinor: "15", SnapshotID: "cap-1", PolicyDigest: "risk-policy", BucketSetDigest: "buckets",
		ObservedAt: "2026-08-03T00:00:00Z", FreshUntil: "2026-08-11T00:00:00Z", FX: plan.FX()})
	if err != nil {
		t.Fatal(err)
	}
	return cap
}

func validFill(plan CampaignPlan, id string) FillRiskEvent {
	return FillRiskEvent{FillID: id, CampaignID: plan.CampaignID(), LegOrdinal: 1, OrderRef: "order-1", Quantity: 2, TransferredReservationMinor: "60",
		EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "3", EstimatedExitFeesLeviesMinor: "2", ObservedAt: "2026-08-04T00:00:03Z", SourceDigest: "fill-source"}
}

func mustRiskState(t *testing.T, plan CampaignPlan, filledMinor, heldMinor string) RiskState {
	t.Helper()
	state, err := mintRiskState(plan, filledMinor, heldMinor)
	if err != nil {
		t.Fatal(err)
	}
	return state
}
