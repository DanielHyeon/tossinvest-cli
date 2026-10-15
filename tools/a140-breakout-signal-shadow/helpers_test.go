package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// 시험 세션: 2026-08-14(금) KRX 정규장 09:00~15:30 KST.
var (
	seoul       = mustSeoul()
	testOpen    = time.Date(2026, 8, 14, 9, 0, 0, 0, seoul)
	testClose   = time.Date(2026, 8, 14, 15, 30, 0, 0, seoul)
	testSession = "KRX:2026-08-14"
)

func mustSeoul() *time.Location {
	loc, err := marketclock.MarketKR.Location()
	if err != nil {
		panic(err)
	}
	return loc
}

type ohlcv struct{ o, h, l, c, v uint64 }

// breakoutScenario 는 개장부터의 봉 목록 — 15 봉 opening range(저항 10,050 · 저점 9,950 · ATR 100), 16 번째 돌파(종가 10,080 ·
// RVOL 3× · 윗꼬리 16.7%), 17 번째 되돌림(10,060 — 저항과 10 차, 허용 25), 18 번째 되찾음(10,070 ≥ 10,050) = ARMED, 19 번째가 목표 10,378 을 넘음.
func breakoutScenario() []ohlcv {
	bars := make([]ohlcv, 0, 20)
	for i := 0; i < 15; i++ {
		bars = append(bars, ohlcv{10000, 10050, 9950, 10000, 100})
	}
	bars = append(bars,
		ohlcv{10040, 10090, 10030, 10080, 300}, // 15: 돌파
		ohlcv{10070, 10075, 10040, 10060, 100}, // 16: 되돌림
		ohlcv{10060, 10080, 10055, 10070, 100}, // 17: 되찾음 → ARMED
		ohlcv{10070, 10400, 10060, 10390, 100}, // 18: 목표 도달
	)
	return bars
}

// seriesOf 는 개장 분부터 이어진 봉 묶음(시험용 literal — 저장소 없이).
func seriesOf(bars []ohlcv) strategyevidence.BarSeries {
	records := make([]strategyevidence.ClosedBarRecord, len(bars))
	for i, b := range bars {
		openAt := testOpen.Add(time.Duration(i) * time.Minute)
		records[i] = strategyevidence.ClosedBarRecord{
			EvidenceID: "ev-" + strconv.Itoa(i), RevisionIdentity: "r1", PayloadDigest: "sha256:" + fmt.Sprintf("%064d", i),
			Payload: strategyevidence.ClosedBar1mPayload{
				Market: "KR", Symbol: "005930", SessionID: testSession, IntervalMS: 60_000, OpenAtMS: uint64(openAt.UnixMilli()),
				Closed: true, RegularSession: true, Currency: "KRW", OpenMinor: b.o, HighMinor: b.h, LowMinor: b.l,
				CloseMinor: b.c, Volume: b.v, Revision: 1,
			},
		}
	}
	return strategyevidence.BarSeries{Bars: records}
}

func testSessionInfo() sessionInfo {
	return sessionInfo{ID: testSession, CalendarVersion: "sha256:" + fmt.Sprintf("%064d", 7), Open: testOpen, Close: testClose}
}

func testParams() shadowParams {
	return shadowParams{RetestTolerancePPM: 250_000, MaxQuoteAgeMS: 5_000, MaxSpreadPPM: 3_000, MaxEntryDriftPPM: 5_000,
		RiskBudgetKRW: 33_333, NotionalCapKRW: 1_000_000, CostPPM: 5_000, MinRRPPM: 1_500_000}
}

// freshQuote 는 ARMED 분(18 번째 봉이 닫힌 09:18)의 호가 — bid 10,060 · ask 10,070.
func freshQuote(at time.Time) quoteSnap {
	return quoteSnap{Bid: 10060, Ask: 10070, Last: 10070, ObservedMS: uint64(at.Add(-time.Second).UnixMilli()),
		ReceivedMS: uint64(at.UnixMilli()), Currency: "KRW", Digest: "sha256:" + fmt.Sprintf("%064d", 9)}
}

// ---- 가짜 브로커 ----

type fakeSource struct {
	mu       sync.Mutex
	bars     []ohlcv // 개장부터의 봉. 지금 시각 이전에 닫힌 봉 + 자라는 봉 하나를 내줌.
	now      func() time.Time
	ranking  []domain.RankingItem
	calls    map[string]int
	rateOnce map[string]bool // 이 메서드는 첫 호출에 rate limit 을 한 번 냄
	bid, ask string
	pageSize int // 0 = 한 페이지에 전부(마지막 페이지)
	// tick 은 브로커 호출마다 시계를 조금 옮김 — 실제처럼 호가의 received_at 이 그 앞에 잡은 시각보다 늦고,
	// 적재 시각이 폴 전 시각보다 늦게 함(같은 순간에 멈춘 시계는 평가 시각 결함을 가림 — 코드 리뷰 P1-3).
	tick func()
	// failAlways 는 이 메서드가 언제나 rate limit 을 냄(재시도를 다 쓰는 경우).
	failAlways map[string]bool
	// missing 은 브로커가 내주지 않는 봉 번호(VI 단일가처럼 빈 분).
	missing map[int]bool
}

func newFakeSource(bars []ohlcv, now func() time.Time) *fakeSource {
	return &fakeSource{bars: bars, now: now, calls: map[string]int{}, rateOnce: map[string]bool{}, failAlways: map[string]bool{}, bid: "10060", ask: "10070",
		ranking: []domain.RankingItem{{Rank: 1, Symbol: "005930", Currency: "KRW", LastPrice: 10000}}}
}

func (f *fakeSource) hit(name string) error {
	if f.tick != nil {
		f.tick()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[name]++
	if f.failAlways[name] {
		return fmt.Errorf("fake: %w", official.ErrRateLimited)
	}
	if f.rateOnce[name] {
		f.rateOnce[name] = false
		return fmt.Errorf("fake: %w", official.ErrRateLimited)
	}
	return nil
}

func (f *fakeSource) count(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[name]
}

func digestFor(label string) string {
	sum := sha256.Sum256([]byte(label))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (f *fakeSource) Rankings(ctx context.Context, typ, marketCountry, duration string, excludeCaution bool, count int) (domain.Ranking, error) {
	if err := f.hit("Rankings"); err != nil {
		return domain.Ranking{}, err
	}
	return domain.Ranking{Type: typ, MarketCountry: marketCountry, Items: f.ranking}, nil
}

func (f *fakeSource) TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error) {
	if err := f.hit("TypedMarketCalendar"); err != nil {
		return official.MarketCalendarResponse{}, err
	}
	day := func(date string) official.MarketCalendarDay {
		parsed, _ := time.ParseInLocation("2006-01-02", date, seoul)
		y, m, d := parsed.Date()
		return official.MarketCalendarDay{Date: date, Integrated: &official.MarketCalendarSessions{RegularMarket: &official.MarketCalendarSession{
			StartTime: time.Date(y, m, d, 9, 0, 0, 0, seoul), EndTime: time.Date(y, m, d, 15, 30, 0, 0, seoul)}}}
	}
	return official.MarketCalendarResponse{PreviousBusinessDay: day("2026-08-13"), Today: day("2026-08-14"), NextBusinessDay: day("2026-08-17")}, nil
}

func (f *fakeSource) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	if err := f.hit("StrictMinuteCandles"); err != nil {
		return official.StrictMinutePage{}, err
	}
	now := f.now()
	bound := now
	if before != "" {
		for _, layout := range []string{"2006-01-02T15:04:05.000-07:00", time.RFC3339} {
			if parsed, err := time.Parse(layout, before); err == nil {
				bound = parsed
				break
			}
		}
	}
	var candles []official.RawMinuteCandle
	terminal := true
	// 이름표(닫는 시각) ≤ before 인 봉만, 최신이 앞(엄격 리더 계약 「모든 봉 ≤ before」 — officialbars checkOverlap 주석).
	// 그래서 자라는 봉은 오지 않고, 생산자는 받은 것 중 가장 새 봉도 적재하지 않으므로 봉 k 는 k+2 분에 적재됨.
	for i := len(f.bars) - 1; i >= 0; i-- {
		openAt := testOpen.Add(time.Duration(i) * time.Minute)
		label := openAt.Add(time.Minute)
		if openAt.After(now) || label.After(bound) || f.missing[i] {
			continue
		}
		if f.pageSize > 0 && len(candles) == f.pageSize {
			terminal = false
			break
		}
		b := f.bars[i]
		candles = append(candles, official.RawMinuteCandle{Timestamp: label.In(seoul).Format("2006-01-02T15:04:05.000-07:00"),
			Open: u(b.o), High: u(b.h), Low: u(b.l), Close: u(b.c), Volume: u(b.v), Currency: "KRW"})
	}
	page := official.StrictMinutePage{Market: market, Symbol: symbol, Candles: candles, Terminal: terminal, ReadAt: now,
		StatusCode: 200, BodyDigest: digestFor("page" + now.String() + before)}
	if !terminal {
		oldest, _ := time.Parse("2006-01-02T15:04:05.000-07:00", candles[len(candles)-1].Timestamp)
		page.NextBefore = oldest.Add(-time.Second).UTC().Format(time.RFC3339)
	}
	return page, nil
}

func (f *fakeSource) StrictOrderbookTop(ctx context.Context, market, symbol string) (official.StrictTopOfBook, error) {
	if err := f.hit("StrictOrderbookTop"); err != nil {
		return official.StrictTopOfBook{}, err
	}
	now := f.now()
	return official.StrictTopOfBook{Market: market, Symbol: symbol, Currency: "KRW", Ask: official.StrictQuoteLevel{Price: f.ask, Volume: "10"},
		Bid: official.StrictQuoteLevel{Price: f.bid, Volume: "10"}, SourceTimestamp: now.Add(-time.Second).Format(time.RFC3339),
		SourceInstant: now.Add(-time.Second), ReadAt: now, StatusCode: 200, BodyDigest: digestFor("book" + now.String())}, nil
}

func (f *fakeSource) StrictLastPrice(ctx context.Context, market, symbol string) (official.StrictLastPrice, error) {
	if err := f.hit("StrictLastPrice"); err != nil {
		return official.StrictLastPrice{}, err
	}
	now := f.now()
	return official.StrictLastPrice{Market: market, Symbol: symbol, Currency: "KRW", Last: f.ask,
		SourceTimestamp: now.Add(-time.Second).Format(time.RFC3339), SourceInstant: now.Add(-time.Second), ReadAt: now,
		StatusCode: 200, BodyDigest: digestFor("last" + now.String())}, nil
}

var errUnscripted = errors.New("unscripted")

// ---- 저장소 · 원장 ----

func openTestStore(t *testing.T, now func() time.Time) *strategyevidence.Store {
	t.Helper()
	store, err := strategyevidence.Open(context.Background(), strategyevidence.Options{
		Path: filepath.Join(t.TempDir(), "shadow-evidence.sqlite"), Clock: clockFunc(now)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

// clockFunc 는 시험이 움직이는 시계를 strategyevidence 의 Clock 으로 씀.
type clockFunc func() time.Time

func (c clockFunc) Now() time.Time                                   { return c().UTC() }
func (c clockFunc) Since(t time.Time) time.Duration                  { return c().Sub(t) }
func (c clockFunc) Sleep(ctx context.Context, d time.Duration) error { return ctx.Err() }

func testCalendar(t *testing.T) scheduler.CalendarSnapshot {
	t.Helper()
	src := newFakeSource(nil, func() time.Time { return testOpen })
	payload, _ := src.TypedMarketCalendar(context.Background(), "KR", "2026-08-14")
	cal, err := scheduler.AdaptOfficialCalendar(marketclock.MarketKR, payload, testOpen.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return cal
}

func u(v uint64) string { return strconv.FormatUint(v, 10) }
