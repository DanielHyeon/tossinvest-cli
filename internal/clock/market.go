package clock

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrUnknownMarket is returned for any market string outside the supported set.
// Every market-scoped judgement fails closed with this error rather than
// defaulting to one exchange's calendar: a mislabelled market must stop the
// caller, not silently borrow Seoul's or New York's hours.
var ErrUnknownMarket = errors.New("clock: unknown market")

// Market identifies an exchange calendar. The values match the lowercase market
// strings the existing order intents already carry ("kr" / "us"), so an intent
// field can be parsed straight into a Market.
type Market string

const (
	// MarketKR is the Korean market: Asia/Seoul, no daylight saving.
	MarketKR Market = "kr"
	// MarketUS is the US market: America/New_York, daylight saving applies.
	MarketUS Market = "us"
)

// ParseMarket converts a market string (any case, surrounding spaces allowed)
// into a Market.
func ParseMarket(s string) (Market, error) {
	switch m := Market(strings.ToLower(strings.TrimSpace(s))); m {
	case MarketKR, MarketUS:
		return m, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownMarket, s)
	}
}

// tzName is the IANA zone each market's local time is defined in. KST is fixed
// at +09:00 (Korea last observed daylight saving in 1988); ET alternates between
// EST (-05:00) and EDT (-04:00), which is why every US judgement must convert
// through the zone rather than through a fixed offset.
var tzName = map[Market]string{
	MarketKR: "Asia/Seoul",
	MarketUS: "America/New_York",
}

// regularHours is the regular (main) session window per market in local market
// time, as [open, close). Extended-hours sessions and exchange holidays are out
// of scope here on purpose: holidays come from the broker's calendar endpoint
// (internal/official.Client.Calendar), and this package must not carry a
// hand-maintained holiday table that would silently rot. Weekends are handled
// because they are a pure calendar rule with no external authority needed.
var regularHours = map[Market]struct{ openH, openM, closeH, closeM int }{
	MarketKR: {9, 0, 15, 30},
	MarketUS: {9, 30, 16, 0},
}

var (
	locOnce  sync.Once
	locCache map[Market]*time.Location
	locErr   error
)

func loadLocations() {
	locCache = make(map[Market]*time.Location, len(tzName))
	for market, name := range tzName {
		loc, err := time.LoadLocation(name)
		if err != nil {
			locErr = fmt.Errorf("clock: loading timezone %s for market %s: %w", name, market, err)
			return
		}
		locCache[market] = loc
	}
}

// Location returns the market's IANA timezone.
func (m Market) Location() (*time.Location, error) {
	if _, ok := tzName[m]; !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownMarket, string(m))
	}
	locOnce.Do(loadLocations)
	if locErr != nil {
		return nil, locErr
	}
	loc, ok := locCache[m]
	if !ok || loc == nil {
		return nil, fmt.Errorf("%w: %q", ErrUnknownMarket, string(m))
	}
	return loc, nil
}

// Session is one market's regular-session window for a single trading day.
//
// Open and Close are absolute instants carrying the market's local zone, so a
// comparison against a UTC instant is exact across DST transitions. The window is
// half-open: at the closing bell the market is already closed.
type Session struct {
	Market Market
	// Day is the market-local calendar date ("2006-01-02") the window belongs to.
	Day string
	// Open and Close bound the regular session in absolute time.
	Open, Close time.Time
	// Weekend is true when Day falls on a Saturday or Sunday. The window is
	// still computed (it is a well-defined wall-clock range) but Contains
	// always reports false.
	Weekend bool
}

// Contains reports whether t falls inside the session, i.e. [Open, Close) on a
// non-weekend day. The zero Session contains nothing.
func (s Session) Contains(t time.Time) bool {
	if s.Weekend || s.Open.IsZero() || s.Close.IsZero() {
		return false
	}
	return !t.Before(s.Open) && t.Before(s.Close)
}

// RegularSession returns the regular-session window of the market-local trading
// day that contains t.
func (m Market) RegularSession(t time.Time) (Session, error) {
	loc, err := m.Location()
	if err != nil {
		return Session{}, err
	}
	hours, ok := regularHours[m]
	if !ok {
		return Session{}, fmt.Errorf("%w: %q", ErrUnknownMarket, string(m))
	}

	local := t.In(loc)
	year, month, day := local.Date()
	weekday := local.Weekday()

	return Session{
		Market:  m,
		Day:     local.Format("2006-01-02"),
		Open:    time.Date(year, month, day, hours.openH, hours.openM, 0, 0, loc),
		Close:   time.Date(year, month, day, hours.closeH, hours.closeM, 0, 0, loc),
		Weekend: weekday == time.Saturday || weekday == time.Sunday,
	}, nil
}

// InRegularSession reports whether t is inside the market's regular session.
func (m Market) InRegularSession(t time.Time) (bool, error) {
	s, err := m.RegularSession(t)
	if err != nil {
		return false, err
	}
	return s.Contains(t), nil
}

// TradingDay returns the market-local calendar date ("2006-01-02") that t belongs
// to. This is the trading-day boundary: one UTC instant can sit on different
// trading days in Seoul and New York, and journal records label the day per
// market rather than per host timezone.
func (m Market) TradingDay(t time.Time) (string, error) {
	loc, err := m.Location()
	if err != nil {
		return "", err
	}
	return t.In(loc).Format("2006-01-02"), nil
}

// SameTradingDay reports whether a and b fall on the same market-local trading
// day.
func (m Market) SameTradingDay(a, b time.Time) (bool, error) {
	dayA, err := m.TradingDay(a)
	if err != nil {
		return false, err
	}
	dayB, err := m.TradingDay(b)
	if err != nil {
		return false, err
	}
	return dayA == dayB, nil
}

// UTCOffset returns the market's offset from UTC at instant t. For MarketUS this
// is -5h under EST and -4h under EDT; the offset follows the instant, so the
// ambiguous local hour of a fall-back day resolves correctly.
func (m Market) UTCOffset(t time.Time) (time.Duration, error) {
	loc, err := m.Location()
	if err != nil {
		return 0, err
	}
	_, offsetSeconds := t.In(loc).Zone()
	return time.Duration(offsetSeconds) * time.Second, nil
}
