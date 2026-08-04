package weeklyvaluelane

import (
	"reflect"
	"testing"
	"time"
)

func TestOfficialIANAWeekIdentityIsStableAcrossCalendarCorrectionHolidayAndDST(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	krA := validWeek(MarketKR, "KR-XKRX-2026-W32", "generation-A", "calendar-A", "2026-08-03", evaluated)
	krB := krA
	krB.CalendarGeneration, krB.CalendarDigest = "generation-B", "calendar-B"
	if code := ValidateMarketWeek(krA, evaluated); code != "" {
		t.Fatalf("KR holiday week=%s", code)
	}
	if code := ValidateMarketWeek(krB, evaluated); code != "" || CanonicalReservationKey("campaign", krA) != CanonicalReservationKey("campaign", krB) {
		t.Fatalf("calendar correction changed identity: %s %q %q", code, CanonicalReservationKey("campaign", krA), CanonicalReservationKey("campaign", krB))
	}
	us := validWeek(MarketUS, "US-XNYS-2026-W11", "generation-A", "calendar-us", "2026-03-09", evaluated)
	if code := ValidateMarketWeek(us, evaluated); code != "" {
		t.Fatalf("US DST week=%s", code)
	}
	us.TimeZone = "Local"
	if code := ValidateMarketWeek(us, evaluated); code != RefusalCalendarInvalid {
		t.Fatalf("server-local fallback accepted=%s", code)
	}
}

func TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease(t *testing.T) {
	evaluated := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	weekA := validWeek(MarketKR, "KR-XKRX-2026-W32", "generation-A", "calendar-A", "2026-08-03", evaluated)
	weekB := weekA
	weekB.CalendarGeneration, weekB.CalendarDigest = "generation-B", "calendar-B"
	base := NewReservationState()
	firstCommand := reserveCommand(0, "campaign", weekA, "reservation-1", "idem-1", 1)
	first, result := ApplyReservation(base, firstCommand)
	if !result.Applied || first.ScopeVersion("campaign", MarketKR) != 1 || first.Len() != 1 {
		t.Fatalf("first reserve=%+v/%+v", first, result)
	}
	replay, replayResult := ApplyReservation(first, firstCommand)
	if !replayResult.Duplicate || !reflect.DeepEqual(replay, first) {
		t.Fatalf("replay changed state=%+v/%+v", replay, replayResult)
	}
	loser := reserveCommand(0, "campaign", weekB, "reservation-2", "idem-2", 1)
	if state, got := ApplyReservation(first, loser); got.Code != RefusalVersionConflict || !reflect.DeepEqual(state, first) {
		t.Fatalf("concurrent loser=%+v/%+v", state, got)
	}

	planRequest := validPlanRequest(MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "1000")
	planRequest.CampaignID = "campaign"
	plan, err := BuildCampaignPlan(planRequest)
	if err != nil {
		t.Fatal(err)
	}
	risk := mustRiskState(t, plan, "0", "15")
	fillCommand := ReservationCommand{Action: ReservationPositiveFill, ExpectedVersion: 1, CampaignID: "campaign", MarketWeek: weekA,
		ReservationID: "reservation-1", IdempotencyKey: "fill-1", PlannedOrdinal: 1, PositiveFillQuantity: 1, EvaluatedAt: evaluated}
	fillCommand = authorizeReservationCommand(fillCommand)
	fillEvent := validFill(plan, "fill-1")
	fillEvent.Quantity, fillEvent.TransferredReservationMinor = 1, "15"
	combined, atomicResult := ApplyPositiveFillAtomic(PositiveFillState{Reservations: first, Risk: risk}, plan, fillCommand, fillEvent)
	consumed := combined.Reservations
	consumedEntry, _ := consumed.Entry(CanonicalReservationKey("campaign", weekA))
	if !atomicResult.Applied || consumedEntry.Status != ReservationConsumed || consumed.PositiveLegCount("campaign", MarketKR) != 1 {
		t.Fatalf("positive partial fill=%+v/%+v", consumed, atomicResult)
	}
	releaseAfterFill := authorizeReservationCommand(ReservationCommand{Action: ReservationZeroRelease, ExpectedVersion: 2, CampaignID: "campaign", MarketWeek: weekA,
		ReservationID: "reservation-1", IdempotencyKey: "release-after-fill", PlannedOrdinal: 1, AuthoritativeZero: true, EvaluatedAt: evaluated})
	if state, got := ApplyReservation(consumed, releaseAfterFill); got.Code != RefusalReservationTerminal || !reflect.DeepEqual(state, consumed) {
		t.Fatalf("consumed week released=%+v/%+v", state, got)
	}

	weekUS := validWeek(MarketUS, "US-XNYS-2026-W32", "generation-A", "calendar-us", "2026-08-03", evaluated)
	state, got := ApplyReservation(NewReservationState(), reserveCommand(0, "campaign-us", weekUS, "reservation-us", "reserve-us", 1))
	if !got.Applied {
		t.Fatal(got.Code)
	}
	releaseUS := authorizeReservationCommand(ReservationCommand{Action: ReservationZeroRelease, ExpectedVersion: 1, CampaignID: "campaign-us", MarketWeek: weekUS,
		ReservationID: "reservation-us", IdempotencyKey: "release-us", PlannedOrdinal: 1, AuthoritativeZero: true, PendingAttempts: 0, EvaluatedAt: evaluated})
	released, got := ApplyReservation(state, releaseUS)
	releasedEntry, _ := released.Entry(CanonicalReservationKey("campaign-us", weekUS))
	if !got.Applied || released.PositiveLegCount("campaign-us", MarketUS) != 0 || releasedEntry.Status != ReservationReleased {
		t.Fatalf("zero release=%+v/%+v", released, got)
	}
	if replay, got := ApplyReservation(released, releaseUS); !got.Duplicate || !reflect.DeepEqual(replay, released) {
		t.Fatalf("release replay=%+v/%+v", replay, got)
	}
}

func validWeek(market Market, stable, generation, digest, sessionDate string, evaluated time.Time) MarketWeekEvidence {
	provider, zone := "XKRX_OFFICIAL", "Asia/Seoul"
	if market == MarketUS {
		provider, zone = "XNYS_OFFICIAL", "America/New_York"
	}
	return MarketWeekEvidence{Market: market, Provider: provider, Official: true, TimeZone: zone, SessionDate: sessionDate,
		StableIdentity: stable, CalendarGeneration: generation, CalendarDigest: digest, ObservedAt: evaluated.Add(-time.Minute), FreshUntil: evaluated.Add(time.Hour)}
}

func reserveCommand(version uint64, campaign string, week MarketWeekEvidence, reservationID, idempotency string, ordinal int) ReservationCommand {
	return authorizeReservationCommand(ReservationCommand{Action: ReservationReserve, ExpectedVersion: version, CampaignID: campaign, MarketWeek: week, ReservationID: reservationID,
		IdempotencyKey: idempotency, PlannedOrdinal: ordinal, EvaluatedAt: week.ObservedAt.Add(time.Minute)})
}
