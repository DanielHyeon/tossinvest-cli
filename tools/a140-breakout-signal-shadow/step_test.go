package main

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// movingClock 은 시험이 분 단위로 옮기는 시계.
type movingClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *movingClock) Now() time.Time  { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *movingClock) set(t time.Time) { c.mu.Lock(); c.now = t; c.mu.Unlock() }

func newTestProbe(t *testing.T, src *fakeSource, clk *movingClock, ledgerDir string) *probe {
	t.Helper()
	store := openTestStore(t, clk.Now)
	led, err := openLedger(ledgerDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = led.Close() })
	cfg := defaultProbeConfig()
	cfg.Params = testParams()
	cfg.ReadGap, cfg.RetryDelays = 0, []time.Duration{0, 0}
	p := newProbe(src, store, led, cfg, clk.Now)
	return p
}

func runMinutes(t *testing.T, p *probe, clk *movingClock, from, to int) {
	t.Helper()
	for minute := from; minute <= to; minute++ {
		clk.set(testOpen.Add(time.Duration(minute)*time.Minute + 5*time.Second))
		if err := p.step(context.Background()); err != nil {
			t.Fatalf("step at minute %d: %v", minute, err)
		}
	}
}

func rowsOf(t *testing.T, dir string) []ledgerRow {
	t.Helper()
	rows, err := readLedger(filepath.Join(dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestASessionProposesOnceReadsQuotesOnlyAtStartAndArmedAndLabels(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	bars := append(breakoutScenario(), ohlcv{10390, 10395, 10380, 10390, 100}, ohlcv{10390, 10395, 10380, 10390, 100})
	src := newFakeSource(bars, clk.Now)
	dir := t.TempDir()
	p := newTestProbe(t, src, clk, dir)
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	if got := src.count("StrictOrderbookTop"); got != 1 {
		t.Fatalf("session start read the book %d times; want 1 (the structural seal)", got)
	}
	runMinutes(t, p, clk, 1, 20)

	var proposed, labels []ledgerRow
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindSignal && row.Phase == "PROPOSED" {
			proposed = append(proposed, row)
		}
		if row.Kind == kindLabel {
			labels = append(labels, row)
		}
	}
	if len(proposed) != 1 || proposed[0].Quantity != 99 || proposed[0].Entry != 10070 || proposed[0].Target != 10378 || proposed[0].Attempt != 1 {
		t.Fatalf("PROPOSED rows = %+v; want one, 99 shares, entry 10,070, target 10,378, first attempt", proposed)
	}
	if len(labels) != 1 || labels[0].Outcome != "target" || labels[0].SetupID != proposed[0].SetupID {
		t.Fatalf("labels = %+v", labels)
	}
	// 호가: 시작 1 + ARMED 시도 1 — 그 뒤에는 다시 읽지 않는다(setup 이 끝남).
	if got := src.count("StrictOrderbookTop"); got != 2 {
		t.Fatalf("book reads = %d; want 2", got)
	}
	// 분마다 정확히 1 페이지(폴 하한). 리더는 이름표(닫는 시각) ≤ before 인 봉만 주고 생산자는 가장 새 봉을 적재하지 않으므로
	// 봉 k 는 k+2 분에 적재됨 — ARMED(봉 17)는 09:19, 라벨(봉 18)은 09:20, 그 뒤로는 그 종목을 더 폴하지 않음.
	if got, want := src.count("StrictMinuteCandles"), 20; got != want {
		t.Fatalf("candle pages = %d; want %d (one page per symbol-minute until the label)", got, want)
	}
}

func TestARestartDoesNotProposeTheSameSetupAgainAndStillLabels(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	bars := append(breakoutScenario(), ohlcv{10390, 10395, 10380, 10390, 100}, ohlcv{10390, 10395, 10380, 10390, 100})
	src := newFakeSource(bars, clk.Now)
	dir := t.TempDir()
	first := newTestProbe(t, src, clk, dir)
	if err := first.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	runMinutes(t, first, clk, 1, 19) // 09:19 에 PROPOSED(봉 17 적재), 라벨은 아직
	_ = first.ledger.Close()
	before := 0
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindSignal && row.Phase == "PROPOSED" {
			before++
		}
	}
	if before != 1 {
		t.Fatalf("the first run left %d PROPOSED rows; the restart must start after the proposal", before)
	}

	second := newTestProbe(t, src, clk, dir)
	if err := second.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	runMinutes(t, second, clk, 20, 22)
	proposals, labels := 0, 0
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindSignal && row.Phase == "PROPOSED" {
			proposals++
		}
		if row.Kind == kindLabel {
			labels++
		}
	}
	if proposals != 1 || labels != 1 {
		t.Fatalf("after a restart: %d PROPOSED rows, %d labels; want 1 and 1", proposals, labels)
	}
	if got := src.count("Rankings"); got != 1 {
		t.Fatalf("the restart re-ranked the universe (%d ranking reads); it must reuse the session's universe row", got)
	}
}

func TestArmedRetriesAtMostFiveMinutesThenExpires(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	bars := breakoutScenario()[:18]
	for i := 0; i < 8; i++ {
		bars = append(bars, ohlcv{10070, 10075, 10065, 10070, 100})
	}
	src := newFakeSource(bars, clk.Now)
	src.bid = "10000" // 스프레드 6,976 ppm — 매번 SPREAD_TOO_WIDE
	dir := t.TempDir()
	p := newTestProbe(t, src, clk, dir)
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	runMinutes(t, p, clk, 1, 26)
	attempts := 0
	expired := false
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindSignal && row.Phase == "ARMED" && row.Attempt > 0 {
			attempts++
			if row.Refusal != "SPREAD_TOO_WIDE" || row.SpreadPPM != 6976 {
				t.Errorf("attempt row %+v", row)
			}
			expired = expired || row.Done == doneArmedExpired
		}
	}
	if attempts != 5 || !expired {
		t.Fatalf("ARMED attempts = %d, expired %v; want 5 and expired", attempts, expired)
	}
	if got := src.count("StrictOrderbookTop"); got != 6 {
		t.Fatalf("book reads = %d; want 1 start + 5 attempts", got)
	}
}

func TestARateLimitOnTheBookSuspendsQuoteReads(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen}
	src := newFakeSource(nil, clk.Now)
	src.rateOnce["StrictOrderbookTop"] = true
	paced := newPacedSource(src, 0, nil, 5*time.Minute, clk.Now, func(context.Context, time.Duration) error { return nil })
	if _, err := paced.StrictOrderbookTop(context.Background(), "KR", "005930"); !errors.Is(err, official.ErrRateLimited) {
		t.Fatalf("first read: %v; want the rate limit", err)
	}
	if _, err := paced.StrictOrderbookTop(context.Background(), "KR", "005930"); !errors.Is(err, errQuoteSuspended) {
		t.Fatalf("second read inside the window: %v; want suspended (no broker call)", err)
	}
	if got := src.count("StrictOrderbookTop"); got != 1 {
		t.Fatalf("book calls = %d; want 1", got)
	}
	clk.set(testOpen.Add(5*time.Minute + time.Second))
	if _, err := paced.StrictOrderbookTop(context.Background(), "KR", "005930"); err != nil {
		t.Fatalf("after the window: %v", err)
	}
	if got := paced.readsByGroup()[groupMarketData]; got != 2 {
		t.Fatalf("MARKET_DATA reads counted = %d; want 2 broker calls", got)
	}
}

// TestThePollLowerBoundKeepsOnePagePerMinute 는 폴 하한(가장 새 저장 봉 − 5 분)을 고정함 — 하한이 개장이면 세션이 길어질수록
// 페이지가 늘어 호출 예산(§6)이 두 배가 됨. 페이지를 10 봉으로 줄여 그 차이를 드러냄.
func TestThePollLowerBoundKeepsOnePagePerMinute(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	var bars []ohlcv
	for i := 0; i < 30; i++ {
		bars = append(bars, ohlcv{10000, 10010, 9990, 10000, 100})
	}
	src := newFakeSource(bars, clk.Now)
	src.pageSize = 10
	p := newTestProbe(t, src, clk, t.TempDir())
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	runMinutes(t, p, clk, 1, 25)
	if got := src.count("StrictMinuteCandles"); got != 25 {
		t.Fatalf("candle pages over 25 minutes = %d; want 25 (one per minute)", got)
	}
	series, err := p.loadSeries(context.Background(), "005930", clk.Now())
	// 09:25:05 의 폴은 이름표 09:25 까지 보고 가장 새 봉(09:24 개장)을 남기므로 저장은 09:00..09:23 의 24 봉.
	if err != nil || len(series.Bars) != 24 {
		t.Fatalf("stored %d bars (%v); want 24 contiguous bars (09:00..09:23)", len(series.Bars), err)
	}
}

// TestAGapIsWrittenOnceAndCountedAlways 는 결손 행의 중복 제거(원장 폭주 금지)와 계수(조용한 생략 금지)를 함께 고정함.
func TestAGapIsWrittenOnceAndCountedAlways(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen}
	dir := t.TempDir()
	p := newTestProbe(t, newFakeSource(nil, clk.Now), clk, dir)
	for i := 0; i < 3; i++ {
		if err := p.gap("005930", "bars", errors.New("READER_ERROR reading page 1: rate limited")); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.gap("005930", "quote", errors.New("QUOTE_READER_ERROR")); err != nil {
		t.Fatal(err)
	}
	gaps := 0
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindGap {
			gaps++
		}
	}
	if gaps != 2 || p.counts["gap:bars"] != 3 || p.counts["gap:quote"] != 1 {
		t.Fatalf("gap rows %d, counts %v; want 2 rows and bars 3 · quote 1", gaps, p.counts)
	}
}
