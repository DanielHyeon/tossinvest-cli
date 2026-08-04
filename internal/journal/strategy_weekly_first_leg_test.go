package journal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

func TestWeeklyFirstLegBindsPairedKRUSReservationsAtomically(t *testing.T) {
	j := openTestJournal(t)
	for _, market := range []string{"KR", "US"} {
		request := weeklyFirstLegFixture(t, j, "weekly-"+strings.ToLower(market), "acct-"+strings.ToLower(market), market, map[string]string{"KR": "005930", "US": "AAPL"}[market])
		receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
		if err != nil {
			t.Fatalf("%s first leg: %v", market, err)
		}
		var boundReservation, boundWeek string
		if err := j.db.QueryRow(`SELECT reservation_id,stable_week FROM strategy_weekly_first_leg_bindings WHERE decision_id=?`, receipt.DecisionID).Scan(&boundReservation, &boundWeek); err != nil {
			t.Fatal(err)
		}
		if boundReservation != request.Weekly.ReservationID || boundWeek != request.Weekly.StableWeek {
			t.Fatalf("%s weekly binding reservation=%s week=%s", market, boundReservation, boundWeek)
		}
		replay, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
		if err != nil || !replay.Idempotent {
			t.Fatalf("%s replay=%+v err=%v", market, replay, err)
		}
	}
}

func TestWeeklyFirstLegRequiresExactReservationAndNonWeeklyCannotSmuggleOne(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			j := openTestJournal(t)
			request := weeklyFirstLegFixture(t, j, "missing-"+strings.ToLower(market), "acct", market, map[string]string{"KR": "005930", "US": "AAPL"}[market])
			request.Weekly.RecordDigest = "different"
			if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); !errors.Is(err, ErrWeeklyReservationConflict) {
				t.Fatalf("divergent weekly reservation err=%v", err)
			}
			for _, table := range []string{"decisions", "strategy_first_leg_bindings", "strategy_weekly_first_leg_bindings"} {
				if got := countRiskBucketRows(t, j, table); got != 0 {
					t.Fatalf("%s rows=%d", table, got)
				}
			}
		})
	}

	j := openTestJournal(t)
	nonWeekly := firstLegAtomicFixture(t, j, "smuggle", "acct", "KR", "005930")
	nonWeekly.Weekly = &WeeklyFirstLegReservationBinding{ReservationID: "forged", StableWeek: "KR-XKRX-2026-W14"}
	if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), nonWeekly); !errors.Is(err, ErrWeeklyReservationConflict) {
		t.Fatalf("non-weekly reservation smuggling err=%v", err)
	}
}

func weeklyFirstLegFixture(t *testing.T, j *Journal, suffix, account, market, symbol string) QFinalCampaignFirstLegRequest {
	t.Helper()
	request := firstLegAtomicFixture(t, j, suffix, account, market, symbol)
	laneID := map[string]string{"KR": "kr_weekly_disclosure_value_v1", "US": "us_weekly_disclosure_value_v1"}[market]
	request.Issue.Admission.Owner.LaneID = laneID
	request.Strategy.Lineage.LaneID = laneID
	var record strategyengine.DecisionRecord
	if err := json.Unmarshal([]byte(request.Strategy.Lineage.DecisionPayload), &record); err != nil {
		t.Fatal(err)
	}
	record.LaneID, record.Identity = laneID, ""
	identityPayload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	identityHash := sha256.Sum256(identityPayload)
	record.Identity = "strategy-decision:v1:sha256:" + hex.EncodeToString(identityHash[:])
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	payloadHash := sha256.Sum256(payload)
	request.Strategy.Lineage.DecisionIdentity = record.Identity
	request.Strategy.Lineage.DecisionPayload = string(payload)
	request.Strategy.Lineage.DecisionPayloadDigest = "sha256:" + hex.EncodeToString(payloadHash[:])

	evaluated := request.Strategy.CreatedAt.UTC()
	provider, zone, prefix := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX"
	if market == "US" {
		provider, zone, prefix = "XNYS_OFFICIAL", "America/New_York", "US-XNYS"
	}
	sessionDate := "2026-03-30"
	session, err := time.Parse("2006-01-02", sessionDate)
	if err != nil {
		t.Fatal(err)
	}
	year, week := session.ISOWeek()
	stable := fmt.Sprintf("%s-%04d-W%02d", prefix, year, week)
	reservation, err := j.ReserveWeeklyMarket(context.Background(), WeeklyMarketReservationRequest{ReservationID: "weekly-reservation-" + suffix,
		CampaignID: request.Campaign.CampaignID, Market: market, StableWeek: stable, Provider: provider, TimeZone: zone, SessionDate: sessionDate,
		CalendarGeneration: "generation-A", CalendarDigest: "calendar-A", IdempotencyKey: "weekly-reserve-" + suffix,
		PlannedOrdinal: 1, ObservedAt: evaluated.Add(-time.Minute), FreshUntil: evaluated.Add(time.Hour), EvaluatedAt: evaluated})
	if err != nil {
		t.Fatal(err)
	}
	request.Weekly = &WeeklyFirstLegReservationBinding{ReservationID: reservation.ReservationID, StableWeek: reservation.StableWeek,
		PlannedOrdinal: reservation.PlannedOrdinal, ScopeVersion: reservation.ScopeVersion, RequestDigest: reservation.RequestDigest,
		RecordDigest: reservation.RecordDigest, CalendarGeneration: reservation.CalendarGeneration, CalendarDigest: reservation.CalendarDigest}
	return request
}
