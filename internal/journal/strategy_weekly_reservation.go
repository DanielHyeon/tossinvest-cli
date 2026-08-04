package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrWeeklyReservationConflict        = errors.New("journal: weekly market reservation conflict")
	ErrWeeklyReservationVersionConflict = errors.New("journal: weekly market reservation version conflict")
	ErrWeeklyReservationMissing         = errors.New("journal: weekly market reservation missing")
)

const (
	WeeklyReservationActive   = "ACTIVE"
	WeeklyReservationConsumed = "CONSUMED"
	WeeklyReservationReleased = "RELEASED"
)

type WeeklyMarketReservationRequest struct {
	ReservationID, CampaignID, Market           string
	StableWeek, Provider, TimeZone, SessionDate string
	CalendarGeneration, CalendarDigest          string
	IdempotencyKey                              string
	PlannedOrdinal                              int
	ExpectedVersion                             uint64
	ObservedAt, FreshUntil, EvaluatedAt         time.Time
}

type WeeklyMarketReservationSnapshot struct {
	ReservationID, CampaignID, Market           string
	StableWeek, Provider, TimeZone, SessionDate string
	CalendarGeneration, CalendarDigest          string
	Status                                      string
	PlannedOrdinal                              int
	ScopeVersion                                uint64
	PositiveLegCount                            int
	ObservedAt, FreshUntil, EvaluatedAt         time.Time
	RequestDigest, RecordDigest                 string
	Idempotent                                  bool
}

func (j *Journal) ReserveWeeklyMarket(ctx context.Context, request WeeklyMarketReservationRequest) (WeeklyMarketReservationSnapshot, error) {
	if j == nil || j.db == nil {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationMissing
	}
	request = canonicalWeeklyReservationRequest(request)
	if err := validateWeeklyReservationRequest(request); err != nil {
		return WeeklyMarketReservationSnapshot{}, err
	}
	requestDigest := weeklyReservationRequestDigest(request)
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return WeeklyMarketReservationSnapshot{}, err
	}
	defer tx.Rollback()

	var priorDigest, priorID string
	err = tx.QueryRowContext(ctx, `SELECT request_digest,reservation_id FROM strategy_weekly_reservation_receipts WHERE campaign_id=? AND market=? AND idempotency_key=?`,
		request.CampaignID, request.Market, request.IdempotencyKey).Scan(&priorDigest, &priorID)
	if err == nil {
		if priorDigest != requestDigest || priorID != request.ReservationID {
			return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationConflict
		}
		snapshot, err := scanWeeklyReservationTx(ctx, tx, request.CampaignID, request.Market, request.StableWeek)
		if err != nil {
			return WeeklyMarketReservationSnapshot{}, err
		}
		snapshot.Idempotent = true
		if err := tx.Commit(); err != nil {
			return WeeklyMarketReservationSnapshot{}, err
		}
		return snapshot, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return WeeklyMarketReservationSnapshot{}, err
	}

	var version uint64
	var positive int
	err = tx.QueryRowContext(ctx, `SELECT version,positive_leg_count FROM strategy_weekly_reservation_scopes WHERE campaign_id=? AND market=?`, request.CampaignID, request.Market).Scan(&version, &positive)
	if errors.Is(err, sql.ErrNoRows) {
		version, positive = 0, 0
	} else if err != nil {
		return WeeklyMarketReservationSnapshot{}, err
	}
	if version != request.ExpectedVersion {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationVersionConflict
	}
	if positive >= 7 || request.PlannedOrdinal != positive+1 {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationConflict
	}
	var conflicts int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM strategy_weekly_market_reservations WHERE campaign_id=? AND market=? AND (stable_week=? OR status='ACTIVE')`,
		request.CampaignID, request.Market, request.StableWeek).Scan(&conflicts); err != nil {
		return WeeklyMarketReservationSnapshot{}, err
	}
	if conflicts != 0 {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationConflict
	}
	nextVersion := version + 1
	stamp := request.EvaluatedAt.UTC().Format(time.RFC3339Nano)
	if version == 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO strategy_weekly_reservation_scopes(campaign_id,market,version,positive_leg_count,updated_at) VALUES(?,?,?,?,?)`, request.CampaignID, request.Market, nextVersion, positive, stamp); err != nil {
			return WeeklyMarketReservationSnapshot{}, fmt.Errorf("%w: %v", ErrWeeklyReservationConflict, err)
		}
	} else if result, err := tx.ExecContext(ctx, `UPDATE strategy_weekly_reservation_scopes SET version=?,updated_at=? WHERE campaign_id=? AND market=? AND version=?`, nextVersion, stamp, request.CampaignID, request.Market, version); err != nil {
		return WeeklyMarketReservationSnapshot{}, err
	} else if changed, _ := result.RowsAffected(); changed != 1 {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationVersionConflict
	}
	recordDigest := weeklyReservationRecordDigest(request, nextVersion, requestDigest)
	if _, err := tx.ExecContext(ctx, `INSERT INTO strategy_weekly_market_reservations(reservation_id,campaign_id,market,stable_week,provider,time_zone,session_date,calendar_generation,calendar_digest,planned_ordinal,status,scope_version,observed_at,fresh_until,evaluated_at,request_digest,record_digest,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,'ACTIVE',?,?,?,?,?,?,?,?)`,
		request.ReservationID, request.CampaignID, request.Market, request.StableWeek, request.Provider, request.TimeZone, request.SessionDate,
		request.CalendarGeneration, request.CalendarDigest, request.PlannedOrdinal, nextVersion, canonicalWeeklyTime(request.ObservedAt),
		canonicalWeeklyTime(request.FreshUntil), canonicalWeeklyTime(request.EvaluatedAt), requestDigest, recordDigest, stamp, stamp); err != nil {
		return WeeklyMarketReservationSnapshot{}, fmt.Errorf("%w: %v", ErrWeeklyReservationConflict, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO strategy_weekly_reservation_receipts(campaign_id,market,idempotency_key,request_digest,reservation_id,created_at) VALUES(?,?,?,?,?,?)`,
		request.CampaignID, request.Market, request.IdempotencyKey, requestDigest, request.ReservationID, stamp); err != nil {
		return WeeklyMarketReservationSnapshot{}, fmt.Errorf("%w: %v", ErrWeeklyReservationConflict, err)
	}
	snapshot, err := scanWeeklyReservationTx(ctx, tx, request.CampaignID, request.Market, request.StableWeek)
	if err != nil {
		return WeeklyMarketReservationSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return WeeklyMarketReservationSnapshot{}, err
	}
	return snapshot, nil
}

func (j *Journal) WeeklyMarketReservation(ctx context.Context, campaignID, market, stableWeek string) (WeeklyMarketReservationSnapshot, error) {
	if j == nil || j.db == nil {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationMissing
	}
	return scanWeeklyReservation(ctx, j.db, campaignID, market, stableWeek)
}

func (r *ReadOnly) WeeklyMarketReservation(ctx context.Context, campaignID, market, stableWeek string) (WeeklyMarketReservationSnapshot, error) {
	if r == nil || r.db == nil || r.version < 27 {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationMissing
	}
	return scanWeeklyReservation(ctx, r.db, campaignID, market, stableWeek)
}

type weeklyReservationQuery interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func scanWeeklyReservation(ctx context.Context, query weeklyReservationQuery, campaignID, market, stableWeek string) (WeeklyMarketReservationSnapshot, error) {
	return scanWeeklyReservationRow(query.QueryRowContext(ctx, `SELECT r.reservation_id,r.campaign_id,r.market,r.stable_week,r.provider,r.time_zone,r.session_date,r.calendar_generation,r.calendar_digest,r.status,r.planned_ordinal,r.scope_version,s.positive_leg_count,r.observed_at,r.fresh_until,r.evaluated_at,r.request_digest,r.record_digest FROM strategy_weekly_market_reservations r JOIN strategy_weekly_reservation_scopes s ON s.campaign_id=r.campaign_id AND s.market=r.market WHERE r.campaign_id=? AND r.market=? AND r.stable_week=?`, strings.TrimSpace(campaignID), strings.ToUpper(strings.TrimSpace(market)), strings.TrimSpace(stableWeek)))
}

func scanWeeklyReservationTx(ctx context.Context, tx *sql.Tx, campaignID, market, stableWeek string) (WeeklyMarketReservationSnapshot, error) {
	return scanWeeklyReservation(ctx, tx, campaignID, market, stableWeek)
}

func scanWeeklyReservationRow(row *sql.Row) (WeeklyMarketReservationSnapshot, error) {
	var result WeeklyMarketReservationSnapshot
	var observed, fresh, evaluated string
	if err := row.Scan(&result.ReservationID, &result.CampaignID, &result.Market, &result.StableWeek, &result.Provider, &result.TimeZone,
		&result.SessionDate, &result.CalendarGeneration, &result.CalendarDigest, &result.Status, &result.PlannedOrdinal, &result.ScopeVersion,
		&result.PositiveLegCount, &observed, &fresh, &evaluated, &result.RequestDigest, &result.RecordDigest); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationMissing
		}
		return WeeklyMarketReservationSnapshot{}, err
	}
	var err error
	if result.ObservedAt, err = time.Parse(time.RFC3339Nano, observed); err != nil {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationConflict
	}
	if result.FreshUntil, err = time.Parse(time.RFC3339Nano, fresh); err != nil {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationConflict
	}
	if result.EvaluatedAt, err = time.Parse(time.RFC3339Nano, evaluated); err != nil {
		return WeeklyMarketReservationSnapshot{}, ErrWeeklyReservationConflict
	}
	return result, nil
}

func canonicalWeeklyReservationRequest(request WeeklyMarketReservationRequest) WeeklyMarketReservationRequest {
	request.ReservationID = strings.TrimSpace(request.ReservationID)
	request.CampaignID = strings.TrimSpace(request.CampaignID)
	request.Market = strings.ToUpper(strings.TrimSpace(request.Market))
	request.StableWeek = strings.TrimSpace(request.StableWeek)
	request.Provider = strings.TrimSpace(request.Provider)
	request.TimeZone = strings.TrimSpace(request.TimeZone)
	request.SessionDate = strings.TrimSpace(request.SessionDate)
	request.CalendarGeneration = strings.TrimSpace(request.CalendarGeneration)
	request.CalendarDigest = strings.TrimSpace(request.CalendarDigest)
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	return request
}

func validateWeeklyReservationRequest(request WeeklyMarketReservationRequest) error {
	for _, value := range []string{request.ReservationID, request.CampaignID, request.StableWeek, request.Provider, request.TimeZone, request.SessionDate,
		request.CalendarGeneration, request.CalendarDigest, request.IdempotencyKey} {
		if value == "" || len(value) > 256 || strings.IndexFunc(value, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0 {
			return ErrWeeklyReservationConflict
		}
	}
	provider, zone, prefix := "", "", ""
	switch request.Market {
	case "KR":
		provider, zone, prefix = "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-"
	case "US":
		provider, zone, prefix = "XNYS_OFFICIAL", "America/New_York", "US-XNYS-"
	default:
		return ErrWeeklyReservationConflict
	}
	if request.Provider != provider || request.TimeZone != zone || !strings.HasPrefix(request.StableWeek, prefix) || request.PlannedOrdinal < 1 || request.PlannedOrdinal > 7 ||
		request.ObservedAt.IsZero() || request.FreshUntil.IsZero() || request.EvaluatedAt.IsZero() || request.EvaluatedAt.Before(request.ObservedAt) || request.EvaluatedAt.After(request.FreshUntil) {
		return ErrWeeklyReservationConflict
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		return ErrWeeklyReservationConflict
	}
	session, err := time.ParseInLocation("2006-01-02", request.SessionDate, location)
	if err != nil || session.Weekday() != time.Monday {
		return ErrWeeklyReservationConflict
	}
	year, week := session.ISOWeek()
	if request.StableWeek != fmt.Sprintf("%s%04d-W%02d", prefix, year, week) {
		return ErrWeeklyReservationConflict
	}
	return nil
}

func weeklyReservationRequestDigest(request WeeklyMarketReservationRequest) string {
	parts := []string{"journal-weekly-reservation-request:v1", request.ReservationID, request.CampaignID, request.Market, request.StableWeek,
		request.Provider, request.TimeZone, request.SessionDate, request.CalendarGeneration, request.CalendarDigest, request.IdempotencyKey,
		strconv.Itoa(request.PlannedOrdinal), strconv.FormatUint(request.ExpectedVersion, 10), canonicalWeeklyTime(request.ObservedAt),
		canonicalWeeklyTime(request.FreshUntil), canonicalWeeklyTime(request.EvaluatedAt)}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

func weeklyReservationRecordDigest(request WeeklyMarketReservationRequest, scopeVersion uint64, requestDigest string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{"journal-weekly-reservation-record:v1", requestDigest, strconv.FormatUint(scopeVersion, 10), WeeklyReservationActive}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func canonicalWeeklyTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
