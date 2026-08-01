package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

type stubScheduleCalendar struct {
	mu       sync.Mutex
	response official.MarketCalendarResponse
	err      error
	country  string
	date     string
	calls    int
}

func (s *stubScheduleCalendar) TypedMarketCalendar(_ context.Context, country, date string) (official.MarketCalendarResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.country, s.date = country, date
	s.calls++
	return s.response, s.err
}

func (s *stubScheduleCalendar) observation() (country, date string, calls int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.country, s.date, s.calls
}

type scheduleCalendarBroker struct {
	*official.Client
	calendar *stubScheduleCalendar
}

func (b scheduleCalendarBroker) TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error) {
	return b.calendar.TypedMarketCalendar(ctx, country, date)
}

// brokerWithoutCalendar is intentionally only the verification surface. It
// proves the calendar adapter cannot widen a broker that lacks the reviewed
// typed official read.
type brokerWithoutCalendar struct{ verifylive.Broker }

func TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef(t *testing.T) {
	calendar := &stubScheduleCalendar{response: productionKRCalendar()}
	built := 0
	previous := verifyBrokerFactory
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		built++
		return scheduleCalendarBroker{calendar: calendar}, "  123-45-678901  ", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = previous })

	shared := newConsoleBroker(&rootOptions{})
	for i := 0; i < 2; i++ {
		if _, err := shared.TypedMarketCalendar(context.Background(), "KR", "2026-08-01"); err != nil {
			t.Fatalf("TypedMarketCalendar call %d: %v", i+1, err)
		}
	}
	if built != 1 {
		t.Fatalf("calendar adapter built the shared broker %d times, want 1", built)
	}
	_, accountRef, err := shared.resolve()
	if err != nil {
		t.Fatalf("resolve cached broker: %v", err)
	}
	if accountRef != "123-45-678901" {
		t.Errorf("cached accountRef = %q, want the trimmed exact broker reference", accountRef)
	}
	country, date, calls := calendar.observation()
	if country != "KR" || date != "2026-08-01" || calls != 2 {
		t.Errorf("calendar delegation = country %q date %q calls %d", country, date, calls)
	}
}

func TestConsoleBrokerTypedMarketCalendarFailsClosed(t *testing.T) {
	t.Run("resolver error", func(t *testing.T) {
		want := errors.New("account resolution unavailable")
		previous := verifyBrokerFactory
		verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
			return nil, "", want
		}
		t.Cleanup(func() { verifyBrokerFactory = previous })

		_, err := newConsoleBroker(&rootOptions{}).TypedMarketCalendar(
			context.Background(), "KR", "2026-08-01",
		)
		if !errors.Is(err, want) {
			t.Fatalf("TypedMarketCalendar error = %v, want resolver error", err)
		}
	})

	t.Run("broker lacks typed calendar", func(t *testing.T) {
		previous := verifyBrokerFactory
		verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
			return brokerWithoutCalendar{}, "123-45-678901", nil
		}
		t.Cleanup(func() { verifyBrokerFactory = previous })

		_, err := newConsoleBroker(&rootOptions{}).TypedMarketCalendar(
			context.Background(), "KR", "2026-08-01",
		)
		if err == nil || !strings.Contains(err.Error(), "has no typed official calendar read") {
			t.Fatalf("TypedMarketCalendar error = %v, want fail-closed capability error", err)
		}
	})
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
	country, date, _ := calendar.observation()
	if country != "KR" || date != "2026-08-01" {
		t.Fatalf("calendar request = %s %s", country, date)
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
