package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

type stubScheduleCalendar struct {
	response official.MarketCalendarResponse
	country  string
	date     string
}

func (s *stubScheduleCalendar) TypedMarketCalendar(_ context.Context, country, date string) (official.MarketCalendarResponse, error) {
	s.country, s.date = country, date
	return s.response, nil
}

type scheduleCalendarBroker struct {
	verifylive.Broker
	calendar *stubScheduleCalendar
}

func (b scheduleCalendarBroker) TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error) {
	return b.calendar.TypedMarketCalendar(ctx, country, date)
}

func TestConsoleMarketScheduleSeamReadsClosedDefaults(t *testing.T) {
	root := &rootOptions{configDir: t.TempDir()}
	reading, err := consoleMarketScheduleSeam(root).Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if reading.SchedulerDesired || reading.AutoStartDesired || reading.SchedulerEffective != "DISABLED" || reading.Market != "none" || reading.Session != "regular" {
		t.Fatalf("reading = %+v", reading)
	}
}

func TestConsoleMarketScheduleSeamDoesNotActivateApprovedDesiredState(t *testing.T) {
	dir := t.TempDir()
	desired := scheduler.DesiredState{
		Version: scheduler.SchedulerVersion, Enabled: true, AutoStart: true,
		Market: scheduler.MarketScopeKR, Session: scheduler.SessionRegular,
		Actor: "operator", ApprovedAt: time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC),
		CalendarVersion: "calendar-v1", ConfigVersion: "config-v1",
	}
	if err := scheduler.NewDesiredStore(filepath.Join(dir, scheduler.DesiredFileName)).Save(context.Background(), desired); err != nil {
		t.Fatal(err)
	}
	calendar := &stubScheduleCalendar{response: productionKRCalendar()}
	shared := &consoleBroker{client: scheduleCalendarBroker{calendar: calendar}}
	reader := consoleMarketScheduleSeam(&rootOptions{configDir: dir}, shared)
	reader.now = func() time.Time { return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC) }
	reading, err := reader.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reading.SchedulerDesired || !reading.AutoStartDesired || reading.SchedulerEffective != "DISABLED" || reading.AutoStartEffective || reading.DecisionReason != "NOT_ACTIVATED" {
		t.Fatalf("dormant reading = %+v", reading)
	}
	if reading.CalendarSource != scheduler.CalendarSourceOfficial || reading.CalendarVersion == "" || reading.CalendarFetchedAt.IsZero() {
		t.Fatalf("authoritative calendar provenance missing: %+v", reading)
	}
	if calendar.country != "KR" || calendar.date != "2026-08-01" {
		t.Fatalf("calendar request = %s %s", calendar.country, calendar.date)
	}
}

func TestConsoleMarketScheduleFetchedAtIsResponseCompletionTime(t *testing.T) {
	dir := t.TempDir()
	desired := scheduler.DesiredState{
		Version: scheduler.SchedulerVersion, Market: scheduler.MarketScopeKR, Session: scheduler.SessionRegular,
	}
	if err := scheduler.NewDesiredStore(filepath.Join(dir, scheduler.DesiredFileName)).Save(context.Background(), desired); err != nil {
		t.Fatal(err)
	}
	calendar := &stubScheduleCalendar{response: productionKRCalendar()}
	reader := consoleMarketScheduleSeam(&rootOptions{configDir: dir}, calendar)
	requestAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	completedAt := requestAt.Add(3 * time.Second)
	times := []time.Time{requestAt, completedAt}
	reader.now = func() time.Time {
		got := times[0]
		times = times[1:]
		return got
	}
	reading, err := reader.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reading.CalendarFetchedAt.Equal(completedAt) {
		t.Fatalf("calendar fetchedAt = %s, want response completion %s", reading.CalendarFetchedAt, completedAt)
	}
}

func productionKRCalendar() official.MarketCalendarResponse {
	day := func(date, open, close string) official.MarketCalendarDay {
		start, _ := time.Parse(time.RFC3339, open)
		end, _ := time.Parse(time.RFC3339, close)
		return official.MarketCalendarDay{Date: date, Integrated: &official.MarketCalendarSessions{
			RegularMarket: &official.MarketCalendarSession{StartTime: start, EndTime: end},
		}}
	}
	return official.MarketCalendarResponse{
		PreviousBusinessDay: day("2026-07-31", "2026-07-31T09:00:00+09:00", "2026-07-31T15:30:00+09:00"),
		Today:               day("2026-08-01", "2026-08-01T09:00:00+09:00", "2026-08-01T15:30:00+09:00"),
		NextBusinessDay:     day("2026-08-03", "2026-08-03T09:00:00+09:00", "2026-08-03T15:30:00+09:00"),
	}
}
