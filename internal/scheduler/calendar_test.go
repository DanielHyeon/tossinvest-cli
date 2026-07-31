package scheduler

import (
	"strings"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func at(t *testing.T, raw string) time.Time {
	t.Helper()
	got, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return got.UTC()
}

func window(t *testing.T, start, end string) *official.MarketCalendarSession {
	t.Helper()
	return &official.MarketCalendarSession{StartTime: at(t, start), EndTime: at(t, end)}
}

func krCalendar(t *testing.T, day, open, close string) official.MarketCalendarResponse {
	t.Helper()
	return official.MarketCalendarResponse{
		PreviousBusinessDay: official.MarketCalendarDay{Date: "2026-03-24", Integrated: &official.MarketCalendarSessions{
			RegularMarket: window(t, "2026-03-24T09:00:00+09:00", "2026-03-24T15:30:00+09:00"),
		}},
		Today: official.MarketCalendarDay{Date: day, Integrated: &official.MarketCalendarSessions{
			RegularMarket: window(t, open, close),
		}},
		NextBusinessDay: official.MarketCalendarDay{Date: "2026-03-26", Integrated: &official.MarketCalendarSessions{
			RegularMarket: window(t, "2026-03-26T09:00:00+09:00", "2026-03-26T15:30:00+09:00"),
		}},
	}
}

func usCalendar(t *testing.T, day, open, close string) official.MarketCalendarResponse {
	t.Helper()
	today, err := time.Parse("2006-01-02", day)
	if err != nil {
		t.Fatal(err)
	}
	previous := today.AddDate(0, 0, -1)
	for previous.Weekday() == time.Saturday || previous.Weekday() == time.Sunday {
		previous = previous.AddDate(0, 0, -1)
	}
	return official.MarketCalendarResponse{
		PreviousBusinessDay: official.MarketCalendarDay{Date: previous.Format("2006-01-02"), RegularMarket: &official.MarketCalendarSession{
			StartTime: time.Date(previous.Year(), previous.Month(), previous.Day(), 9, 30, 0, 0, time.FixedZone("ET", offsetForDay(previous))),
			EndTime:   time.Date(previous.Year(), previous.Month(), previous.Day(), 16, 0, 0, 0, time.FixedZone("ET", offsetForDay(previous))),
		}},
		Today: official.MarketCalendarDay{Date: day, RegularMarket: window(t, open, close)},
		NextBusinessDay: official.MarketCalendarDay{Date: "2026-03-10", RegularMarket: window(t,
			"2026-03-10T09:30:00-04:00", "2026-03-10T16:00:00-04:00")},
	}
}

func offsetForDay(day time.Time) int {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	_, offset := day.In(loc).Zone()
	return offset
}

func TestOfficialCalendarAdaptsRegularHolidayAndEarlyClose(t *testing.T) {
	fetched := at(t, "2026-03-25T08:30:00+09:00")
	regular, err := AdaptOfficialCalendar(marketclock.MarketKR, krCalendar(t, "2026-03-25",
		"2026-03-25T09:00:00+09:00", "2026-03-25T15:30:00+09:00"), fetched)
	if err != nil {
		t.Fatalf("regular: %v", err)
	}
	if regular.Today.Regular == nil || regular.Today.EarlyClose {
		t.Fatalf("regular day = %+v", regular.Today)
	}
	if regular.Version == "" || regular.Source != CalendarSourceOfficial {
		t.Fatalf("provenance = version:%q source:%q", regular.Version, regular.Source)
	}
	if !strings.HasPrefix(regular.Version, "sha256:") {
		t.Fatalf("calendar version = %q", regular.Version)
	}

	holidayPayload := krCalendar(t, "2026-03-25", "2026-03-25T09:00:00+09:00", "2026-03-25T15:30:00+09:00")
	holidayPayload.Today.Integrated = nil
	holiday, err := AdaptOfficialCalendar(marketclock.MarketKR, holidayPayload, fetched)
	if err != nil || holiday.Today.Regular != nil {
		t.Fatalf("holiday = %+v err=%v", holiday.Today, err)
	}

	early, err := AdaptOfficialCalendar(marketclock.MarketKR, krCalendar(t, "2026-03-25",
		"2026-03-25T09:00:00+09:00", "2026-03-25T12:30:00+09:00"), fetched)
	if err != nil || !early.Today.EarlyClose {
		t.Fatalf("early close = %+v err=%v", early.Today, err)
	}
}

func TestUSCalendarUsesExchangeTimezoneAcrossDST(t *testing.T) {
	cases := []struct {
		name, day, open, close, wantOpen, wantClose string
	}{
		{"EST", "2026-03-06", "2026-03-06T09:30:00-05:00", "2026-03-06T16:00:00-05:00", "2026-03-06T14:30:00Z", "2026-03-06T21:00:00Z"},
		{"EDT", "2026-03-09", "2026-03-09T09:30:00-04:00", "2026-03-09T16:00:00-04:00", "2026-03-09T13:30:00Z", "2026-03-09T20:00:00Z"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AdaptOfficialCalendar(marketclock.MarketUS,
				usCalendar(t, tc.day, tc.open, tc.close), at(t, tc.open).Add(-time.Hour))
			if err != nil {
				t.Fatal(err)
			}
			if !got.Today.Regular.Open.Equal(at(t, tc.wantOpen)) ||
				!got.Today.Regular.Close.Equal(at(t, tc.wantClose)) {
				t.Fatalf("session = %v..%v", got.Today.Regular.Open, got.Today.Regular.Close)
			}
		})
	}
}

func TestCalendarVersionIsCanonicalAndSemantic(t *testing.T) {
	payload := krCalendar(t, "2026-03-25", "2026-03-25T09:00:00+09:00", "2026-03-25T15:30:00+09:00")
	a, err := AdaptOfficialCalendar(marketclock.MarketKR, payload, at(t, "2026-03-25T08:00:00+09:00"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := AdaptOfficialCalendar(marketclock.MarketKR, payload, at(t, "2026-03-25T08:30:00+09:00"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Version != b.Version {
		t.Fatalf("fetched-at changed response digest: %q != %q", a.Version, b.Version)
	}
	payload.Today.Integrated.RegularMarket.EndTime = at(t, "2026-03-25T12:30:00+09:00")
	c, err := AdaptOfficialCalendar(marketclock.MarketKR, payload, a.FetchedAt)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version == a.Version {
		t.Fatal("changed early close did not change canonical calendar digest")
	}
}

func TestCalendarFreshnessAndSessionPreparationFailClosed(t *testing.T) {
	open := at(t, "2026-03-25T09:00:00+09:00")
	payload := krCalendar(t, "2026-03-25", "2026-03-25T09:00:00+09:00", "2026-03-25T15:30:00+09:00")
	cases := []struct {
		name    string
		fetched time.Time
		now     time.Time
		want    CalendarValidity
	}{
		{"fresh before session", open.Add(-time.Hour), open.Add(time.Minute), CalendarValid},
		{"six hours is stale", open.Add(-time.Hour), open.Add(5 * time.Hour), CalendarStale},
		{"refresh after open cannot prepare session", open.Add(time.Minute), open.Add(2 * time.Minute), CalendarRefreshTooLate},
		{"backward clock jump", open.Add(-time.Hour), open.Add(-2 * time.Hour), CalendarClockSkew},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snap, err := AdaptOfficialCalendar(marketclock.MarketKR, payload, tc.fetched)
			if err != nil {
				t.Fatal(err)
			}
			if got := snap.ValidityAt(tc.now); got != tc.want {
				t.Fatalf("ValidityAt = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCalendarRejectsMalformedExchangeDay(t *testing.T) {
	payload := usCalendar(t, "2026-03-09", "2026-03-10T09:30:00-04:00", "2026-03-10T16:00:00-04:00")
	if _, err := AdaptOfficialCalendar(marketclock.MarketUS, payload, at(t, "2026-03-09T08:00:00-04:00")); err == nil {
		t.Fatal("session whose exchange-local day disagrees with date was accepted")
	}
}

func TestCalendarRejectsMissingOrNonChronologicalBusinessDayEvidence(t *testing.T) {
	fetched := at(t, "2026-03-25T08:00:00+09:00")
	for _, mutate := range []func(*official.MarketCalendarResponse){
		func(p *official.MarketCalendarResponse) { p.Today.Date = "" },
		func(p *official.MarketCalendarResponse) { p.PreviousBusinessDay.Date = p.Today.Date },
		func(p *official.MarketCalendarResponse) { p.NextBusinessDay.Date = "2026-03-24" },
		func(p *official.MarketCalendarResponse) { p.PreviousBusinessDay.Integrated = nil },
		func(p *official.MarketCalendarResponse) { p.NextBusinessDay.Integrated = nil },
	} {
		payload := krCalendar(t, "2026-03-25", "2026-03-25T09:00:00+09:00", "2026-03-25T15:30:00+09:00")
		mutate(&payload)
		if _, err := AdaptOfficialCalendar(marketclock.MarketKR, payload, fetched); err == nil {
			t.Fatalf("malformed chronology accepted: %+v", payload)
		}
	}
}
