package main

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
)

type consoleMarketScheduleReader struct {
	store    *scheduler.DesiredStore
	calendar typedMarketCalendarReader
	now      func() time.Time
	err      error
}

type typedMarketCalendarReader interface {
	TypedMarketCalendar(context.Context, string, string) (official.MarketCalendarResponse, error)
}

type consoleMarketScheduleStatus struct {
	SchedulerDesired   bool
	AutoStartDesired   bool
	SchedulerEffective string
	AutoStartEffective bool
	Market             string
	Session            string
	ApplyTiming        string
	CalendarSource     string
	CalendarVersion    string
	CalendarFetchedAt  time.Time
	DecisionReason     string
	NextTransition     time.Time
}

func consoleMarketScheduleSeam(root *rootOptions, calendars ...typedMarketCalendarReader) *consoleMarketScheduleReader {
	configPath, err := configFilePath(root)
	if err != nil {
		return &consoleMarketScheduleReader{err: err}
	}
	var calendar typedMarketCalendarReader
	if len(calendars) > 0 {
		calendar = calendars[0]
	}
	return &consoleMarketScheduleReader{store: scheduler.NewDesiredStore(
		filepath.Join(filepath.Dir(configPath), scheduler.DesiredFileName),
	), calendar: calendar, now: func() time.Time { return time.Now().UTC() }}
}

func (r *consoleMarketScheduleReader) Read(ctx context.Context) (consoleMarketScheduleStatus, error) {
	if r.err != nil {
		return consoleMarketScheduleStatus{}, r.err
	}
	desired, err := r.store.Load(ctx)
	if err != nil {
		return consoleMarketScheduleStatus{}, err
	}
	status := consoleMarketScheduleStatus{
		SchedulerDesired: desired.Enabled,
		AutoStartDesired: desired.AutoStart,
		// a047 has no production activation-manifest verifier yet. Showing an
		// approved desired state as effective here would invent that approval.
		SchedulerEffective: "DISABLED",
		AutoStartEffective: false,
		Market:             string(desired.Market),
		Session:            string(desired.Session),
		ApplyTiming:        "다음 엔진 기동",
		DecisionReason:     "NOT_ACTIVATED",
	}
	if desired.Market == scheduler.MarketScopeNone {
		return status, nil
	}
	if r.calendar == nil {
		return consoleMarketScheduleStatus{}, errors.New("scheduler calendar reader is unavailable")
	}
	market, country, ok := scheduleCalendarMarket(desired.Market)
	if !ok {
		return consoleMarketScheduleStatus{}, errors.New("scheduler market calendar provenance is unavailable")
	}
	requestAt := time.Now().UTC()
	if r.now != nil {
		requestAt = r.now().UTC()
	}
	location, err := market.Location()
	if err != nil {
		return consoleMarketScheduleStatus{}, err
	}
	payload, err := r.calendar.TypedMarketCalendar(ctx, country, requestAt.In(location).Format("2006-01-02"))
	if err != nil {
		return consoleMarketScheduleStatus{}, err
	}
	fetchedAt := time.Now().UTC()
	if r.now != nil {
		fetchedAt = r.now().UTC()
	}
	snapshot, err := scheduler.AdaptOfficialCalendar(market, payload, fetchedAt)
	if err != nil {
		return consoleMarketScheduleStatus{}, err
	}
	status.CalendarSource = snapshot.Source
	status.CalendarVersion = snapshot.Version
	status.CalendarFetchedAt = snapshot.FetchedAt
	return status, nil
}

func scheduleCalendarMarket(scope scheduler.MarketScope) (marketclock.Market, string, bool) {
	switch scope {
	case scheduler.MarketScopeKR:
		return marketclock.MarketKR, "KR", true
	case scheduler.MarketScopeUS:
		return marketclock.MarketUS, "US", true
	default:
		return "", "", false
	}
}
