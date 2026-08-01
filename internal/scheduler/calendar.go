// Package scheduler contains fail-closed market scheduling primitives. It is
// deliberately not connected to the strategy runtime until the activation
// manifest contract is implemented by change a047.
package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	CalendarSourceOfficial = "official-openapi"
	calendarMaxAge         = 6 * time.Hour
	regularSessionDuration = 6*time.Hour + 30*time.Minute
)

type CalendarValidity string

const (
	CalendarValid          CalendarValidity = "VALID"
	CalendarStale          CalendarValidity = "STALE"
	CalendarRefreshTooLate CalendarValidity = "REFRESH_TOO_LATE"
	CalendarClockSkew      CalendarValidity = "CLOCK_SKEW"
)

type SessionWindow struct {
	Open  time.Time `json:"open"`
	Close time.Time `json:"close"`
}

type CalendarDay struct {
	Date       string         `json:"date"`
	Regular    *SessionWindow `json:"regular,omitempty"`
	EarlyClose bool           `json:"earlyClose"`
}

type CalendarSnapshot struct {
	Market              marketclock.Market `json:"market"`
	Version             string             `json:"version"`
	Source              string             `json:"source"`
	FetchedAt           time.Time          `json:"fetchedAt"`
	PreviousBusinessDay CalendarDay        `json:"previousBusinessDay"`
	Today               CalendarDay        `json:"today"`
	NextBusinessDay     CalendarDay        `json:"nextBusinessDay"`
}

// AdaptOfficialCalendar validates official sessions in the exchange's IANA
// timezone and derives a canonical version that is independent of fetch time.
func AdaptOfficialCalendar(market marketclock.Market, payload official.MarketCalendarResponse, fetchedAt time.Time) (CalendarSnapshot, error) {
	if fetchedAt.IsZero() {
		return CalendarSnapshot{}, fmt.Errorf("calendar fetched-at is required")
	}
	loc, err := market.Location()
	if err != nil {
		return CalendarSnapshot{}, err
	}
	previous, err := adaptCalendarDay(market, loc, payload.PreviousBusinessDay)
	if err != nil {
		return CalendarSnapshot{}, fmt.Errorf("previous business day: %w", err)
	}
	today, err := adaptCalendarDay(market, loc, payload.Today)
	if err != nil {
		return CalendarSnapshot{}, fmt.Errorf("today: %w", err)
	}
	next, err := adaptCalendarDay(market, loc, payload.NextBusinessDay)
	if err != nil {
		return CalendarSnapshot{}, fmt.Errorf("next business day: %w", err)
	}
	previousDate, previousErr := time.Parse("2006-01-02", previous.Date)
	todayDate, todayErr := time.Parse("2006-01-02", today.Date)
	nextDate, nextErr := time.Parse("2006-01-02", next.Date)
	if previousErr != nil || todayErr != nil || nextErr != nil || !previousDate.Before(todayDate) || !todayDate.Before(nextDate) {
		return CalendarSnapshot{}, errors.New("calendar requires previous < today < next business-day evidence")
	}
	if previous.Regular == nil || next.Regular == nil {
		return CalendarSnapshot{}, errors.New("previous and next business days require regular-session evidence")
	}
	canonical := struct {
		Market              marketclock.Market `json:"market"`
		Source              string             `json:"source"`
		PreviousBusinessDay CalendarDay        `json:"previousBusinessDay"`
		Today               CalendarDay        `json:"today"`
		NextBusinessDay     CalendarDay        `json:"nextBusinessDay"`
	}{market, CalendarSourceOfficial, previous, today, next}
	raw, err := json.Marshal(canonical)
	if err != nil {
		return CalendarSnapshot{}, fmt.Errorf("canonical calendar: %w", err)
	}
	digest := sha256.Sum256(raw)
	return CalendarSnapshot{
		Market: market, Version: "sha256:" + hex.EncodeToString(digest[:]), Source: CalendarSourceOfficial,
		FetchedAt: fetchedAt, PreviousBusinessDay: previous, Today: today, NextBusinessDay: next,
	}, nil
}

func adaptCalendarDay(market marketclock.Market, loc *time.Location, day official.MarketCalendarDay) (CalendarDay, error) {
	if day.Date == "" {
		return CalendarDay{}, nil
	}
	if _, err := time.Parse("2006-01-02", day.Date); err != nil {
		return CalendarDay{}, fmt.Errorf("invalid date %q: %w", day.Date, err)
	}
	var session *official.MarketCalendarSession
	switch market {
	case marketclock.MarketKR:
		if day.Integrated != nil {
			session = day.Integrated.RegularMarket
		}
	case marketclock.MarketUS:
		session = day.RegularMarket
	default:
		return CalendarDay{}, fmt.Errorf("unsupported market %q", market)
	}
	out := CalendarDay{Date: day.Date}
	if session == nil {
		return out, nil
	}
	if session.StartTime.IsZero() || session.EndTime.IsZero() || !session.StartTime.Before(session.EndTime) {
		return CalendarDay{}, fmt.Errorf("invalid regular session")
	}
	if session.StartTime.In(loc).Format("2006-01-02") != day.Date || session.EndTime.In(loc).Format("2006-01-02") != day.Date {
		return CalendarDay{}, fmt.Errorf("regular session is outside exchange-local date %s", day.Date)
	}
	out.Regular = &SessionWindow{Open: session.StartTime, Close: session.EndTime}
	out.EarlyClose = session.EndTime.Sub(session.StartTime) < regularSessionDuration
	return out, nil
}

// ValidityAt applies the six-hour freshness window and requires the calendar
// used during a live session to have been fetched before that session opened.
func (c CalendarSnapshot) ValidityAt(now time.Time) CalendarValidity {
	if c.FetchedAt.IsZero() || now.Before(c.FetchedAt) {
		return CalendarClockSkew
	}
	if now.Sub(c.FetchedAt) >= calendarMaxAge {
		return CalendarStale
	}
	if c.Today.Regular != nil && !now.Before(c.Today.Regular.Open) && now.Before(c.Today.Regular.Close) && !c.FetchedAt.Before(c.Today.Regular.Open) {
		return CalendarRefreshTooLate
	}
	return CalendarValid
}
