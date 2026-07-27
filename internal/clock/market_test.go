package clock_test

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parsing %q: %v", s, err)
	}
	return ts.UTC()
}

// TestMarketLocations pins the two timezones the execution path is allowed to
// reason in. A wrong zone here silently misjudges every session boundary.
func TestMarketLocations(t *testing.T) {
	kr, err := clock.MarketKR.Location()
	if err != nil {
		t.Fatalf("KR location: %v", err)
	}
	if kr.String() != "Asia/Seoul" {
		t.Fatalf("KR location: want Asia/Seoul, got %s", kr)
	}
	us, err := clock.MarketUS.Location()
	if err != nil {
		t.Fatalf("US location: %v", err)
	}
	if us.String() != "America/New_York" {
		t.Fatalf("US location: want America/New_York, got %s", us)
	}
	if _, err := clock.Market("jp").Location(); !errors.Is(err, clock.ErrUnknownMarket) {
		t.Fatalf("unknown market: want ErrUnknownMarket, got %v", err)
	}
}

// TestParseMarket verifies the case-insensitive parse against the lowercase
// market strings the existing intent types already use ("kr" / "us").
func TestParseMarket(t *testing.T) {
	cases := []struct {
		in      string
		want    clock.Market
		wantErr bool
	}{
		{in: "kr", want: clock.MarketKR},
		{in: "KR", want: clock.MarketKR},
		{in: " Kr ", want: clock.MarketKR},
		{in: "us", want: clock.MarketUS},
		{in: "US", want: clock.MarketUS},
		{in: "jp", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tc := range cases {
		got, err := clock.ParseMarket(tc.in)
		if tc.wantErr {
			if !errors.Is(err, clock.ErrUnknownMarket) {
				t.Errorf("ParseMarket(%q): want ErrUnknownMarket, got (%v, %v)", tc.in, got, err)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("ParseMarket(%q) = (%v, %v), want (%v, nil)", tc.in, got, err, tc.want)
		}
	}
}

// TestRegularSessionDSTTable is the DST transition table the order-execution
// spec requires: ET sessions must move with US daylight saving while KST never
// does. Offsets are asserted in UTC because that is the frame the broker
// timestamps and the journal use.
func TestRegularSessionDSTTable(t *testing.T) {
	cases := []struct {
		name        string
		market      clock.Market
		instant     string // any instant inside the target trading day
		wantDay     string
		wantOpen    string // UTC
		wantClose   string // UTC
		wantWeekend bool
	}{
		// --- US: last trading day before the spring-forward transition (EST, -5) ---
		{
			name: "US Fri before spring forward (EST)", market: clock.MarketUS,
			instant: "2026-03-06T15:00:00Z", wantDay: "2026-03-06",
			wantOpen: "2026-03-06T14:30:00Z", wantClose: "2026-03-06T21:00:00Z",
		},
		// --- US: spring-forward Sunday itself — market closed, but the window
		// must still be computed with the post-transition offset (EDT, -4) ---
		{
			name: "US spring-forward Sunday (EDT, weekend)", market: clock.MarketUS,
			instant: "2026-03-08T15:00:00Z", wantDay: "2026-03-08",
			wantOpen: "2026-03-08T13:30:00Z", wantClose: "2026-03-08T20:00:00Z",
			wantWeekend: true,
		},
		// --- US: first trading day after spring forward (EDT, -4) ---
		{
			name: "US Mon after spring forward (EDT)", market: clock.MarketUS,
			instant: "2026-03-09T15:00:00Z", wantDay: "2026-03-09",
			wantOpen: "2026-03-09T13:30:00Z", wantClose: "2026-03-09T20:00:00Z",
		},
		// --- US: last trading day before fall back (EDT, -4) ---
		{
			name: "US Fri before fall back (EDT)", market: clock.MarketUS,
			instant: "2026-10-30T15:00:00Z", wantDay: "2026-10-30",
			wantOpen: "2026-10-30T13:30:00Z", wantClose: "2026-10-30T20:00:00Z",
		},
		// --- US: first trading day after fall back (EST, -5) ---
		{
			name: "US Mon after fall back (EST)", market: clock.MarketUS,
			instant: "2026-11-02T15:00:00Z", wantDay: "2026-11-02",
			wantOpen: "2026-11-02T14:30:00Z", wantClose: "2026-11-02T21:00:00Z",
		},
		// --- KR: identical windows on both US transition weeks (KST, +9, no DST) ---
		{
			name: "KR Mon after US spring forward (KST)", market: clock.MarketKR,
			instant: "2026-03-09T03:00:00Z", wantDay: "2026-03-09",
			wantOpen: "2026-03-09T00:00:00Z", wantClose: "2026-03-09T06:30:00Z",
		},
		{
			name: "KR Mon after US fall back (KST)", market: clock.MarketKR,
			instant: "2026-11-02T03:00:00Z", wantDay: "2026-11-02",
			wantOpen: "2026-11-02T00:00:00Z", wantClose: "2026-11-02T06:30:00Z",
		},
		{
			name: "KR Saturday is not a session", market: clock.MarketKR,
			instant: "2026-03-07T03:00:00Z", wantDay: "2026-03-07",
			wantOpen: "2026-03-07T00:00:00Z", wantClose: "2026-03-07T06:30:00Z",
			wantWeekend: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := tc.market.RegularSession(mustParse(t, tc.instant))
			if err != nil {
				t.Fatalf("RegularSession: %v", err)
			}
			if s.Market != tc.market {
				t.Errorf("Market: want %v, got %v", tc.market, s.Market)
			}
			if s.Day != tc.wantDay {
				t.Errorf("Day: want %s, got %s", tc.wantDay, s.Day)
			}
			if want := mustParse(t, tc.wantOpen); !s.Open.Equal(want) {
				t.Errorf("Open: want %v, got %v (%v)", want, s.Open.UTC(), s.Open)
			}
			if want := mustParse(t, tc.wantClose); !s.Close.Equal(want) {
				t.Errorf("Close: want %v, got %v (%v)", want, s.Close.UTC(), s.Close)
			}
			if s.Weekend != tc.wantWeekend {
				t.Errorf("Weekend: want %v, got %v", tc.wantWeekend, s.Weekend)
			}
		})
	}
}

// TestUTCOffsetAcrossDST covers the ambiguous and skipped local hours directly:
// during the US fall-back hour the same local wall clock occurs twice, and our
// offset must follow the instant, not the wall clock.
func TestUTCOffsetAcrossDST(t *testing.T) {
	cases := []struct {
		name    string
		market  clock.Market
		instant string
		want    time.Duration
	}{
		{"US before spring forward", clock.MarketUS, "2026-03-08T06:30:00Z", -5 * time.Hour},
		{"US after spring forward", clock.MarketUS, "2026-03-08T07:30:00Z", -4 * time.Hour},
		{"US fall back first 01:30 (EDT)", clock.MarketUS, "2026-11-01T05:30:00Z", -4 * time.Hour},
		{"US fall back second 01:30 (EST)", clock.MarketUS, "2026-11-01T06:30:00Z", -5 * time.Hour},
		{"KR on US spring-forward day", clock.MarketKR, "2026-03-08T06:30:00Z", 9 * time.Hour},
		{"KR on US fall-back day", clock.MarketKR, "2026-11-01T06:30:00Z", 9 * time.Hour},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.market.UTCOffset(mustParse(t, tc.instant))
			if err != nil {
				t.Fatalf("UTCOffset: %v", err)
			}
			if got != tc.want {
				t.Fatalf("UTCOffset: want %v, got %v", tc.want, got)
			}
		})
	}
}

// TestSessionContainsBoundaries pins the half-open [open, close) contract: at the
// closing bell the market is already closed.
func TestSessionContainsBoundaries(t *testing.T) {
	cases := []struct {
		name    string
		market  clock.Market
		instant string
		want    bool
	}{
		{"US one second before open", clock.MarketUS, "2026-03-09T13:29:59Z", false},
		{"US at open", clock.MarketUS, "2026-03-09T13:30:00Z", true},
		{"US one second before close", clock.MarketUS, "2026-03-09T19:59:59Z", true},
		{"US at close", clock.MarketUS, "2026-03-09T20:00:00Z", false},
		{"KR at open", clock.MarketKR, "2026-03-09T00:00:00Z", true},
		{"KR one second before close", clock.MarketKR, "2026-03-09T06:29:59Z", true},
		{"KR at close", clock.MarketKR, "2026-03-09T06:30:00Z", false},
		{"KR Saturday inside the hours", clock.MarketKR, "2026-03-07T03:00:00Z", false},
		{"US Sunday inside the hours", clock.MarketUS, "2026-03-08T15:00:00Z", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			at := mustParse(t, tc.instant)
			got, err := tc.market.InRegularSession(at)
			if err != nil {
				t.Fatalf("InRegularSession: %v", err)
			}
			if got != tc.want {
				t.Fatalf("InRegularSession(%s): want %v, got %v", tc.instant, tc.want, got)
			}
			// Session.Contains must agree with the convenience wrapper.
			s, err := tc.market.RegularSession(at)
			if err != nil {
				t.Fatalf("RegularSession: %v", err)
			}
			if s.Contains(at) != tc.want {
				t.Fatalf("Session.Contains disagrees with InRegularSession for %s", tc.instant)
			}
		})
	}
}

// TestTradingDayBoundary is the cross-market boundary case: one UTC instant can
// belong to different trading days in Seoul and New York, and the journal must
// label it per market.
func TestTradingDayBoundary(t *testing.T) {
	// 2026-03-30 is a Monday. 13:30Z = 22:30 KST (same KR day) = 09:30 EDT.
	// 20:00Z     = 05:00 KST next day  = 16:00 EDT (same US day).
	early := mustParse(t, "2026-03-30T13:30:00Z")
	late := mustParse(t, "2026-03-30T20:00:00Z")

	krEarly, err := clock.MarketKR.TradingDay(early)
	if err != nil {
		t.Fatalf("KR TradingDay: %v", err)
	}
	krLate, err := clock.MarketKR.TradingDay(late)
	if err != nil {
		t.Fatalf("KR TradingDay: %v", err)
	}
	if krEarly != "2026-03-30" || krLate != "2026-03-31" {
		t.Fatalf("KR trading days: want 2026-03-30 / 2026-03-31, got %s / %s", krEarly, krLate)
	}

	usEarly, err := clock.MarketUS.TradingDay(early)
	if err != nil {
		t.Fatalf("US TradingDay: %v", err)
	}
	usLate, err := clock.MarketUS.TradingDay(late)
	if err != nil {
		t.Fatalf("US TradingDay: %v", err)
	}
	if usEarly != "2026-03-30" || usLate != "2026-03-30" {
		t.Fatalf("US trading days: want both 2026-03-30, got %s / %s", usEarly, usLate)
	}

	krSame, err := clock.MarketKR.SameTradingDay(early, late)
	if err != nil {
		t.Fatalf("KR SameTradingDay: %v", err)
	}
	usSame, err := clock.MarketUS.SameTradingDay(early, late)
	if err != nil {
		t.Fatalf("US SameTradingDay: %v", err)
	}
	if krSame {
		t.Error("KR SameTradingDay: want false (KST rolled over), got true")
	}
	if !usSame {
		t.Error("US SameTradingDay: want true (still the same ET session day), got false")
	}
}

// TestUnknownMarketFailsClosed keeps every market-scoped judgement erroring out
// rather than defaulting to one exchange's calendar.
func TestUnknownMarketFailsClosed(t *testing.T) {
	at := mustParse(t, "2026-03-09T15:00:00Z")
	m := clock.Market("nasdaq")

	if _, err := m.RegularSession(at); !errors.Is(err, clock.ErrUnknownMarket) {
		t.Errorf("RegularSession: want ErrUnknownMarket, got %v", err)
	}
	if _, err := m.InRegularSession(at); !errors.Is(err, clock.ErrUnknownMarket) {
		t.Errorf("InRegularSession: want ErrUnknownMarket, got %v", err)
	}
	if _, err := m.TradingDay(at); !errors.Is(err, clock.ErrUnknownMarket) {
		t.Errorf("TradingDay: want ErrUnknownMarket, got %v", err)
	}
	if _, err := m.UTCOffset(at); !errors.Is(err, clock.ErrUnknownMarket) {
		t.Errorf("UTCOffset: want ErrUnknownMarket, got %v", err)
	}
	if _, err := m.SameTradingDay(at, at); !errors.Is(err, clock.ErrUnknownMarket) {
		t.Errorf("SameTradingDay: want ErrUnknownMarket, got %v", err)
	}
}

// TestSessionZeroValueContains guards against a zero Session accidentally
// reporting an open market.
func TestSessionZeroValueContains(t *testing.T) {
	var s clock.Session
	if s.Contains(mustParse(t, "2026-03-09T15:00:00Z")) {
		t.Fatal("zero Session must never contain an instant")
	}
}
