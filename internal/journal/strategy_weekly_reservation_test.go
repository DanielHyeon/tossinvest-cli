package journal

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestWeeklyMarketReservationPersistsPairedKRUSAndReplays(t *testing.T) {
	j := openTestJournal(t)
	for _, market := range []string{"KR", "US"} {
		request := weeklyReservationFixture(market)
		first, err := j.ReserveWeeklyMarket(context.Background(), request)
		if err != nil {
			t.Fatalf("%s reserve: %v", market, err)
		}
		replay, err := j.ReserveWeeklyMarket(context.Background(), request)
		if err != nil || !replay.Idempotent || replay != firstWithReplay(first) {
			t.Fatalf("%s replay=%+v first=%+v err=%v", market, replay, first, err)
		}
		got, err := j.WeeklyMarketReservation(context.Background(), request.CampaignID, market, request.StableWeek)
		if err != nil || got.ReservationID != request.ReservationID || got.Status != WeeklyReservationActive || got.ScopeVersion != 1 {
			t.Fatalf("%s projection=%+v err=%v", market, got, err)
		}
	}
}

func TestWeeklyMarketReservationRejectsCalendarCorrectionDuplicate(t *testing.T) {
	j := openTestJournal(t)
	request := weeklyReservationFixture("KR")
	if _, err := j.ReserveWeeklyMarket(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	corrected := request
	corrected.ReservationID = "reservation-KR-corrected"
	corrected.IdempotencyKey = "idem-KR-corrected"
	corrected.CalendarGeneration = "generation-B"
	corrected.CalendarDigest = "calendar-B"
	corrected.ExpectedVersion = 1
	if _, err := j.ReserveWeeklyMarket(context.Background(), corrected); !errors.Is(err, ErrWeeklyReservationConflict) {
		t.Fatalf("calendar correction duplicate err=%v", err)
	}
}

func TestWeeklyMarketReservationConcurrentSameScopeHasOneWinner(t *testing.T) {
	j := openTestJournal(t)
	requests := []WeeklyMarketReservationRequest{weeklyReservationFixture("US"), weeklyReservationFixture("US")}
	requests[1].ReservationID = "reservation-US-loser"
	requests[1].IdempotencyKey = "idem-US-loser"
	var wg sync.WaitGroup
	errs := make(chan error, len(requests))
	for _, request := range requests {
		request := request
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := j.ReserveWeeklyMarket(context.Background(), request)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	success, refused := 0, 0
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, ErrWeeklyReservationConflict) || errors.Is(err, ErrWeeklyReservationVersionConflict) {
			refused++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || refused != 1 {
		t.Fatalf("success=%d refused=%d", success, refused)
	}
	var rows int
	if err := j.db.QueryRow(`SELECT count(*) FROM strategy_weekly_market_reservations WHERE campaign_id=? AND market='US'`, requests[0].CampaignID).Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("rows=%d err=%v", rows, err)
	}
}

func weeklyReservationFixture(market string) WeeklyMarketReservationRequest {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	provider, zone, stable := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-2026-W32"
	if market == "US" {
		provider, zone, stable = "XNYS_OFFICIAL", "America/New_York", "US-XNYS-2026-W32"
	}
	return WeeklyMarketReservationRequest{ReservationID: "reservation-" + market, CampaignID: "campaign-" + market, Market: market,
		StableWeek: stable, Provider: provider, TimeZone: zone, SessionDate: "2026-08-03", CalendarGeneration: "generation-A",
		CalendarDigest: "calendar-A", IdempotencyKey: "idem-" + market, PlannedOrdinal: 1, ExpectedVersion: 0,
		ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Hour), EvaluatedAt: now}
}

func firstWithReplay(first WeeklyMarketReservationSnapshot) WeeklyMarketReservationSnapshot {
	first.Idempotent = true
	return first
}
