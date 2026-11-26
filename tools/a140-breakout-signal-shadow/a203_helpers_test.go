package main

// a203_helpers_test.go 는 a203 시험의 가짜 브로커 · 가짜 시계 — 네트워크 0, 종목 코드는 가짜(TST…).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// a203Clock 은 sleep 이 시계를 옮기는 가짜 벽시계.
type a203Clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *a203Clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *a203Clock) sleep(ctx context.Context, d time.Duration) error {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
	return ctx.Err()
}

func kst(y int, m time.Month, d, hh, mm int) time.Time {
	return time.Date(y, m, d, hh, mm, 0, 0, seoul)
}

// a203FakeLive 는 a203Live 가짜 — 평일 영업일 달력 · 일봉 · 랭킹 · 1분봉(정규장 390 + 애프터 30 + 프리 30 봉).
type a203FakeLive struct {
	mu       sync.Mutex
	clock    *a203Clock
	calls    []string
	sentAt   []time.Time
	failOn   map[int]error // n 번째 호출(1 부터)에 이 오류
	holidays map[string]bool
	ranking  []string
	readAt   time.Time
	dailyEnd string // 일봉 200 개의 마지막 날짜
	bars     func(symbol, date string) []ohlcv
	volume   func(symbol string, day int) float64
	closeAt  [2]int // 정규장 마감 시 · 분(기본 15:30 — 끝에서 끝 시험은 짧은 세션으로 적재 시간을 줄임)
}

func newA203FakeLive(clock *a203Clock) *a203FakeLive {
	return &a203FakeLive{clock: clock, failOn: map[int]error{}, holidays: map[string]bool{}, readAt: kst(2026, 11, 20, 3, 0),
		dailyEnd: "2026-11-26", bars: func(string, string) []ohlcv { return nil },
		volume: func(symbol string, day int) float64 { return 1000 }, closeAt: [2]int{15, 30}}
}

func (f *a203FakeLive) hit(kind string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, kind)
	if f.clock != nil {
		f.sentAt = append(f.sentAt, f.clock.now())
	}
	return f.failOn[len(f.calls)]
}

func (f *a203FakeLive) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *a203FakeLive) business(date string) bool {
	d, _ := time.ParseInLocation("2006-01-02", date, seoul)
	return d.Weekday() != time.Saturday && d.Weekday() != time.Sunday && !f.holidays[date]
}

func (f *a203FakeLive) step(date string, dir int) string {
	d, _ := time.ParseInLocation("2006-01-02", date, seoul)
	for {
		d = d.AddDate(0, 0, dir)
		if f.business(d.Format("2006-01-02")) {
			return d.Format("2006-01-02")
		}
	}
}

func (f *a203FakeLive) calendarDay(date string, regular bool) official.MarketCalendarDay {
	day := official.MarketCalendarDay{Date: date}
	if regular {
		d, _ := time.ParseInLocation("2006-01-02", date, seoul)
		y, m, dd := d.Date()
		day.Integrated = &official.MarketCalendarSessions{RegularMarket: &official.MarketCalendarSession{
			StartTime: time.Date(y, m, dd, 9, 0, 0, 0, seoul), EndTime: time.Date(y, m, dd, f.closeAt[0], f.closeAt[1], 0, 0, seoul)}}
	}
	return day
}

func (f *a203FakeLive) TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error) {
	if err := f.hit("calendar"); err != nil {
		return official.MarketCalendarResponse{}, err
	}
	return official.MarketCalendarResponse{PreviousBusinessDay: f.calendarDay(f.step(date, -1), true),
		Today: f.calendarDay(date, f.business(date)), NextBusinessDay: f.calendarDay(f.step(date, 1), true)}, nil
}

func (f *a203FakeLive) Rankings(ctx context.Context, typ, country, duration string, excludeCaution bool, count int) (domain.Ranking, error) {
	if err := f.hit("ranking"); err != nil {
		return domain.Ranking{}, err
	}
	out := domain.Ranking{Type: typ, MarketCountry: country, Duration: duration}
	for i, code := range f.ranking {
		out.Items = append(out.Items, domain.RankingItem{Rank: i + 1, Symbol: code, Currency: "KRW", LastPrice: 10000})
	}
	return out, nil
}

func (f *a203FakeLive) Candles(ctx context.Context, symbol, interval string, count int, before string, adjusted bool) (domain.Chart, error) {
	if err := f.hit("daily"); err != nil {
		return domain.Chart{}, err
	}
	dates := []string{f.dailyEnd}
	for len(dates) < count {
		dates = append([]string{f.step(dates[0], -1)}, dates...)
	}
	chart := domain.Chart{Symbol: symbol, Interval: interval}
	for i, date := range dates {
		d, _ := time.ParseInLocation("2006-01-02", date, seoul)
		chart.Candles = append(chart.Candles, domain.Candle{Time: d, Open: 10000, High: 10050, Low: 9950, Close: 10000, Volume: f.volume(symbol, i)})
	}
	return chart, nil
}

// StrictMinuteCandles 는 before(닫는 시각 포함 상한) 이하 이름표의 봉을 최신이 앞으로 count 개까지 — 세션 날짜는 before 의 KST 날짜.
func (f *a203FakeLive) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	if err := f.hit("minute"); err != nil {
		return official.StrictMinutePage{}, err
	}
	var bound time.Time
	for _, layout := range []string{"2006-01-02T15:04:05.000-07:00", time.RFC3339} {
		if parsed, err := time.Parse(layout, before); err == nil {
			bound = parsed
			break
		}
	}
	date := bound.In(seoul).Format("2006-01-02")
	day, _ := time.ParseInLocation("2006-01-02", date, seoul)
	regular := f.bars(symbol, date)
	flat := ohlcv{10000, 10050, 9950, 10000, 100}
	type bar struct {
		open time.Time
		v    ohlcv
	}
	var all []bar
	regularMinutes := (f.closeAt[0]*60 + f.closeAt[1]) - 9*60
	for i := 0; i < 30+regularMinutes+30; i++ { // 08:30 부터 여는 봉(프리 30 · 정규 · 애프터 30)
		openAt := day.Add(8*time.Hour + 30*time.Minute + time.Duration(i)*time.Minute)
		v := flat
		if r := i - 30; r >= 0 && r < len(regular) {
			v = regular[r]
		}
		all = append(all, bar{openAt, v})
	}
	page := official.StrictMinutePage{Market: market, Symbol: symbol, Terminal: true, ReadAt: f.readAt, StatusCode: 200}
	for i := len(all) - 1; i >= 0; i-- {
		label := all[i].open.Add(time.Minute)
		if label.After(bound) {
			continue
		}
		if len(page.Candles) == count {
			page.Terminal = false
			break
		}
		b := all[i].v
		page.Candles = append(page.Candles, official.RawMinuteCandle{Timestamp: label.In(seoul).Format("2006-01-02T15:04:05.000-07:00"),
			Open: u(b.o), High: u(b.h), Low: u(b.l), Close: u(b.c), Volume: u(b.v), Currency: "KRW"})
	}
	if !page.Terminal {
		oldest, _ := time.Parse("2006-01-02T15:04:05.000-07:00", page.Candles[len(page.Candles)-1].Timestamp)
		page.NextBefore = oldest.Add(-time.Second).UTC().Format(time.RFC3339)
	}
	page.BodyDigest = digestFor("a203" + symbol + before)
	return page, nil
}

// a203WritePA 는 가짜 P(a) 파일을 상태 디렉터리에 두고 그 지문을 사전등록 값으로 바꿔 끼움(시험 끝에 되돌림).
func a203WritePA(t *testing.T, stateDir string, codes []string) {
	t.Helper()
	sum, raw := a203ListFingerprint(codes)
	path := filepath.Join(stateDir, a203PAFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	old := a203PAFingerprint
	a203PAFingerprint = sum
	t.Cleanup(func() { a203PAFingerprint = old })
}

func fakeCodes(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("TST%03d", i+1)
	}
	return out
}

func shaOf(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
