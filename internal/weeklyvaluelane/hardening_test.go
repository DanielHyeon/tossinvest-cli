package weeklyvaluelane

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEvaluationRequiresSealedAuthorization(t *testing.T) {
	evidence := mustKREvidence(t)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	request := validEvaluation(t, plan, evidence, validKRConfig())
	request.authorization = evaluationAuthorization{}
	if got := EvaluateKR(request); got.Kind != OutcomeRefusal || got.Code != RefusalLaneOff || got.Quantity != 0 {
		t.Fatalf("unsealed caller activated dormant lane: %+v", got)
	}
}

func TestAuthorizedKRAndUSEvaluateIndependently(t *testing.T) {
	krEvidence, usEvidence := mustKREvidence(t), mustUSEvidence(t)
	krPlan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	fx := validFX()
	usPlan := mustPlan(t, MarketUS, USWeeklyLaneID, "KRW", "USD", &fx, 14, "1000")
	kr, us := EvaluateKR(validEvaluation(t, krPlan, krEvidence, validKRConfig())), EvaluateUS(validEvaluation(t, usPlan, usEvidence, validUSConfig()))
	if kr.Kind != OutcomeDecision || us.Kind != OutcomeDecision || kr.Lineage.Market != MarketKR || us.Lineage.Market != MarketUS {
		t.Fatalf("authorized peers not independent: KR=%+v US=%+v", kr, us)
	}
}

func TestEvaluationRejectsStaleOrQuantityMismatchedCap(t *testing.T) {
	evidence := mustKREvidence(t)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	request := validEvaluation(t, plan, evidence, validKRConfig())
	staleCap, err := mintRiskCap(plan, riskCapInput{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", QFinal: 20, ReservationQuantity: request.Cap.reservationQuantity,
		ReservationMinor: "20", MaxStopDistanceMinor: "15", SnapshotID: "stale-cap", PolicyDigest: "risk-policy", BucketSetDigest: "buckets",
		ObservedAt: request.Evidence.EvaluatedAt.Add(-time.Minute).Format(time.RFC3339Nano), FreshUntil: request.Evidence.EvaluatedAt.Format(time.RFC3339Nano)})
	if err != nil {
		t.Fatal(err)
	}
	request.Cap = staleCap
	request.Evidence.EvaluatedAt = request.Cap.freshUntil.Add(time.Nanosecond)
	request.Evidence.FreshUntil = request.Evidence.EvaluatedAt.Add(time.Hour)
	request.Evidence.seal = evidenceSnapshotSeal(request.Evidence)
	request.authorization = mintDormantEvaluationAuthorization(plan, request.Evidence)
	request.MarketWeek.FreshUntil = request.Evidence.EvaluatedAt.Add(time.Hour)
	if got := EvaluateKR(request); got.Code != RefusalCapInvalid {
		t.Fatalf("stale cap accepted: %+v", got)
	}

	request = validEvaluation(t, plan, evidence, validKRConfig())
	request.Cap = validCap(t, plan, 1, request.Cap.reservationQuantity+1, "20")
	if got := EvaluateKR(request); got.Code != RefusalCapInvalid {
		t.Fatalf("non-exact reservation quantity accepted: %+v", got)
	}
}

func TestMarketWeekDerivesStableIdentityExactlyAndBoundsIt(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	week := validWeek(MarketKR, "KR-XKRX-2026-W32", "generation-A", "calendar-A", "2026-08-03", evaluated)
	week.StableIdentity = "KR-XKRX-2099-W01"
	if code := ValidateMarketWeek(week, evaluated); code != RefusalCalendarInvalid {
		t.Fatalf("forged identity accepted: %s", code)
	}
	week = validWeek(MarketUS, "US-XNYS-2026-W11", "generation-A", "calendar-US", "2026-03-09", evaluated)
	if code := ValidateMarketWeek(week, evaluated); code != "" {
		t.Fatalf("derived DST week refused: %s", code)
	}
	week.CalendarGeneration = strings.Repeat("g", maxIdentityBytes+1)
	if code := ValidateMarketWeek(week, evaluated); code != RefusalCalendarInvalid {
		t.Fatalf("oversized calendar identity accepted: %s", code)
	}
}

func TestReservationRejectsStaleCalendarAtTrustedCommandEvaluation(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	week := validWeek(MarketKR, "KR-XKRX-2026-W32", "generation-A", "calendar-A", "2026-08-03", evaluated)
	command := reserveCommand(0, "campaign", week, "reservation", "reserve", 1)
	command.EvaluatedAt = week.FreshUntil.Add(time.Nanosecond)
	command = authorizeReservationCommand(command)
	if next, got := ApplyReservation(NewReservationState(), command); got.Applied || got.Code != RefusalCalendarInvalid || !reflect.DeepEqual(next, NewReservationState()) {
		t.Fatalf("stale calendar reserved: %+v/%+v", next, got)
	}
}

func TestReservationScopesVersionCountAndOrdinals(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	kr := validWeek(MarketKR, "KR-XKRX-2026-W32", "generation-A", "kr", "2026-08-03", evaluated)
	us := validWeek(MarketUS, "US-XNYS-2026-W32", "generation-A", "us", "2026-08-03", evaluated)
	state, krResult := ApplyReservation(NewReservationState(), reserveCommand(0, "campaign", kr, "kr-res", "kr-idem", 1))
	state, usResult := ApplyReservation(state, reserveCommand(0, "campaign", us, "us-res", "us-idem", 1))
	state, peerResult := ApplyReservation(state, reserveCommand(0, "peer", kr, "peer-res", "peer-idem", 1))
	if !krResult.Applied || !usResult.Applied || !peerResult.Applied {
		t.Fatalf("independent scopes collided: kr=%+v us=%+v peer=%+v", krResult, usResult, peerResult)
	}
	if state.ScopeVersion("campaign", MarketKR) != 1 || state.ScopeVersion("campaign", MarketUS) != 1 || state.ScopeVersion("peer", MarketKR) != 1 {
		t.Fatalf("scope versions not independent: %+v", state)
	}
}

func TestPositiveFillsRequireDistinctSequentialOrdinals(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	state := NewReservationState()
	planRequest := validPlanRequest(MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	planRequest.CampaignID = "campaign"
	plan, err := BuildCampaignPlan(planRequest)
	if err != nil {
		t.Fatal(err)
	}
	risk := mustRiskState(t, plan, "0", "105")
	for ordinal := 1; ordinal <= 7; ordinal++ {
		week := validWeek(MarketKR, "KR-XKRX-2026-W"+string(rune('0'+ordinal)), "generation", "calendar", "2026-08-03", evaluated)
		// Keep the exact official identity but move campaigns only in this state-machine test.
		week.StableIdentity = "KR-XKRX-2026-W32"
		if ordinal > 1 {
			week.SessionDate = "2026-08-03"
			week.StableIdentity = "KR-XKRX-2026-W32"
			// A consumed canonical week cannot be reused; use subsequent exact ISO weeks.
			monday := time.Date(2026, 8, 3+7*(ordinal-1), 0, 0, 0, 0, time.UTC)
			week.SessionDate = monday.Format("2006-01-02")
			year, isoWeek := monday.ISOWeek()
			week.StableIdentity = stableMarketWeekIdentity(MarketKR, year, isoWeek)
		}
		reserve := reserveCommand(uint64((ordinal-1)*2), "campaign", week, "res-"+string(rune('0'+ordinal)), "reserve-"+string(rune('0'+ordinal)), ordinal)
		reserve.ExpectedVersion = state.ScopeVersion("campaign", MarketKR)
		reserve = authorizeReservationCommand(reserve)
		var got ReservationResult
		state, got = ApplyReservation(state, reserve)
		if !got.Applied {
			t.Fatalf("reserve ordinal %d: %+v", ordinal, got)
		}
		fill := ReservationCommand{Action: ReservationPositiveFill, ExpectedVersion: state.ScopeVersion("campaign", MarketKR), CampaignID: "campaign", MarketWeek: week,
			ReservationID: reserve.ReservationID, IdempotencyKey: "fill-" + string(rune('0'+ordinal)), PlannedOrdinal: ordinal, PositiveFillQuantity: 1, EvaluatedAt: evaluated}
		fill = authorizeReservationCommand(fill)
		event := validFill(plan, "fill-"+string(rune('0'+ordinal)))
		event.LegOrdinal, event.Quantity, event.TransferredReservationMinor = ordinal, 1, "15"
		combined, atomicResult := ApplyPositiveFillAtomic(PositiveFillState{Reservations: state, Risk: risk}, plan, fill, event)
		state, risk = combined.Reservations, combined.Risk
		if !atomicResult.Applied {
			t.Fatalf("fill ordinal %d: %+v", ordinal, atomicResult)
		}
	}
	if state.PositiveLegCount("campaign", MarketKR) != 7 || state.NextPlannedOrdinal("campaign", MarketKR) != 0 {
		t.Fatalf("seven distinct legs not consumed: %+v", state)
	}

	week := validWeek(MarketKR, "KR-XKRX-2026-W40", "generation", "calendar", "2026-09-28", evaluated)
	if _, got := ApplyReservation(state, reserveCommand(state.ScopeVersion("campaign", MarketKR), "campaign", week, "res-8", "reserve-8", 1)); got.Code != RefusalPlanExhausted {
		t.Fatalf("ordinal replay/exhaustion accepted: %+v", got)
	}
}

func TestPositiveFillCannotBypassAtomicRiskTransition(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	week := validWeek(MarketKR, "KR-XKRX-2026-W32", "generation", "calendar", "2026-08-03", evaluated)
	state, result := ApplyReservation(NewReservationState(), reserveCommand(0, "campaign", week, "reservation", "reserve", 1))
	if !result.Applied {
		t.Fatal(result.Code)
	}
	command := ReservationCommand{Action: ReservationPositiveFill, ExpectedVersion: 1, CampaignID: "campaign", MarketWeek: week, ReservationID: "reservation",
		IdempotencyKey: "fill", PlannedOrdinal: 1, PositiveFillQuantity: 1, EvaluatedAt: evaluated}
	command = authorizeReservationCommand(command)
	if next, got := ApplyReservation(state, command); got.Applied || got.Code != RefusalReservationConflict || !reflect.DeepEqual(next, state) {
		t.Fatalf("positive fill bypassed atomic transition: %+v/%+v", next, got)
	}
}

func TestDecodedEvidenceSealRejectsLiteralAndMutation(t *testing.T) {
	evidence := mustKREvidence(t)
	literal := evidence
	literal.seal = [32]byte{}
	if got := EvaluateKREvidence(literal, validKRConfig()); got.Accepted || got.Code != RefusalStrictSchema {
		t.Fatalf("caller literal accepted: %+v", got)
	}
	mutated := evidence
	mutated.FinancialInputs = append([]FinancialInput(nil), evidence.FinancialInputs...)
	mutated.FinancialInputs[0].ValueMinor = "1"
	if got := EvaluateKREvidence(mutated, validKRConfig()); got.Accepted || got.Code != RefusalStrictSchema {
		t.Fatalf("post-decode mutation accepted: %+v", got)
	}
}

func TestDecisionDigestCoversFullImmutableEvidence(t *testing.T) {
	body := mustFixture(t, "kr_opendart_v1.json")
	changed := bytes.Replace(body, []byte(`"value_minor":"110000"`), []byte(`"value_minor":"110001"`), 1)
	left, err := DecodeKREvidence(body)
	if err != nil {
		t.Fatal(err)
	}
	right, err := DecodeKREvidence(changed)
	if err != nil {
		t.Fatal(err)
	}
	leftResult, rightResult := EvaluateKREvidence(left, validKRConfig()), EvaluateKREvidence(right, validKRConfig())
	if !leftResult.Accepted || !rightResult.Accepted || leftResult.DecisionDigest == rightResult.DecisionDigest {
		t.Fatalf("financial vector absent from digest: left=%+v right=%+v", leftResult, rightResult)
	}
	for name, mutate := range map[string]func(*DisclosureEvidence){
		"revision predecessor": func(value *DisclosureEvidence) { value.SupersededRevisionID = "rev-0" },
		"point in time":        func(value *DisclosureEvidence) { value.ObservedAt = value.ObservedAt.Add(time.Second) },
		"dilution":             func(value *DisclosureEvidence) { value.DilutionFactsDigest = "changed-dilution" },
		"model":                func(value *DisclosureEvidence) { value.ModelID = "changed-model" },
	} {
		t.Run(name, func(t *testing.T) {
			changedEvidence, err := DecodeKREvidence(body)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&changedEvidence)
			changedEvidence.seal = evidenceSnapshotSeal(changedEvidence)
			changedResult := EvaluateKREvidence(changedEvidence, validKRConfig())
			if !changedResult.Accepted || changedResult.DecisionDigest == leftResult.DecisionDigest {
				t.Fatalf("%s absent from digest: base=%+v changed=%+v", name, leftResult, changedResult)
			}
		})
	}
}

func TestEvaluationRejectsStaleOrUnsealedStop(t *testing.T) {
	evidence := mustKREvidence(t)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	request := validEvaluation(t, plan, evidence, validKRConfig())
	request.StopCandidate.FreshUntil = request.Evidence.EvaluatedAt.Add(-time.Nanosecond)
	if got := EvaluateKR(request); got.Code != RefusalStopInvalid {
		t.Fatalf("stale stop accepted: %+v", got)
	}
	request = validEvaluation(t, plan, evidence, validKRConfig())
	request.StopCandidate.seal = [32]byte{}
	if got := EvaluateKR(request); got.Code != RefusalStopInvalid {
		t.Fatalf("unsealed stop accepted: %+v", got)
	}
}

func TestEffectiveStopUsesFreshTighterCandidate(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	candidate := mintStopCandidate(stopCandidateInput{PriceMinor: "95", Version: "stop-v1", Source: "structure", Policy: "policy", Digest: "digest",
		ObservedAt: evaluated.Add(-time.Minute), FreshUntil: evaluated.Add(time.Minute)})
	if got, code := effectiveStop("90", candidate, evaluated); code != "" || got != "95" {
		t.Fatalf("fresh tighter stop not selected: stop=%s code=%s", got, code)
	}
}

func TestApplyPositiveFillAtomicCommitsReservationAndRiskTogether(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	week := validWeek(MarketKR, "KR-XKRX-2026-W32", "generation", "calendar", "2026-08-03", evaluated)
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	reservations, reserve := ApplyReservation(NewReservationState(), reserveCommand(0, plan.CampaignID(), week, "reservation", "reserve", 1))
	if !reserve.Applied {
		t.Fatal(reserve.Code)
	}
	risk := mustRiskState(t, plan, "0", "60")
	command := ReservationCommand{Action: ReservationPositiveFill, ExpectedVersion: reservations.ScopeVersion(plan.CampaignID(), MarketKR), CampaignID: plan.CampaignID(), MarketWeek: week,
		ReservationID: "reservation", IdempotencyKey: "fill", PlannedOrdinal: 1, PositiveFillQuantity: 2, EvaluatedAt: evaluated}
	command = authorizeReservationCommand(command)
	event := validFill(plan, "fill")
	next, result := ApplyPositiveFillAtomic(PositiveFillState{Reservations: reservations, Risk: risk}, plan, command, event)
	if !result.Applied || next.Reservations.PositiveLegCount(plan.CampaignID(), MarketKR) != 1 || next.Risk.FilledMinor() != "60" || next.Risk.HeldMinor() != "0" {
		t.Fatalf("atomic fill not committed: state=%+v result=%+v", next, result)
	}

	badCommand := command
	badCommand.IdempotencyKey = "bad-fill"
	badCommand.ReservationID = "missing"
	badCommand = authorizeReservationCommand(badCommand)
	badEvent := event
	badEvent.FillID = "bad-fill"
	unchanged, failed := ApplyPositiveFillAtomic(next, plan, badCommand, badEvent)
	if failed.Applied || !reflect.DeepEqual(unchanged, next) {
		t.Fatalf("failed atomic transition partially committed: %+v/%+v", unchanged, failed)
	}
}

func TestRiskStatePostMutationAndCopiedMapTamperFailClosed(t *testing.T) {
	plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "100")
	cap := validCap(t, plan, 1, 1, "10")
	state := mustRiskState(t, plan, "90", "0")
	if code := AdmitRisk(plan, state, cap); code != "" {
		t.Fatalf("baseline risk state invalid: %s", code)
	}
	tampered := state
	tampered.filledMinor = "0"
	if code := AdmitRisk(plan, tampered, cap); code != RefusalRiskLatched {
		t.Fatalf("scalar risk tamper admitted: %s", code)
	}
	if next, result := ApplyFillRisk(tampered, plan, validFill(plan, "tampered-fill")); result.Applied || result.Code != RefusalRiskLatched || !reflect.DeepEqual(next, tampered) {
		t.Fatalf("scalar risk tamper applied: %+v/%+v", next, result)
	}

	state = mustRiskState(t, plan, "0", "20")
	event := validFill(plan, "")
	latched, result := ApplyFillRisk(state, plan, event)
	if !result.Applied || !latched.Latched(LatchUnknownActualRisk) {
		t.Fatalf("failed to create authoritative latch: %+v/%+v", latched, result)
	}
	mapTampered := latched
	delete(mapTampered.latches, LatchUnknownActualRisk)
	if code := AdmitRisk(plan, mapTampered, cap); code != RefusalRiskLatched {
		t.Fatalf("copied latch-map tamper admitted: %s", code)
	}
}
