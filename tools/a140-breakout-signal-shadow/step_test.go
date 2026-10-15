package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// movingClock 은 시험이 분 단위로 옮기는 시계.
type movingClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *movingClock) Now() time.Time  { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *movingClock) set(t time.Time) { c.mu.Lock(); c.now = t; c.mu.Unlock() }

// advance 는 브로커 호출 하나가 걸리는 시간(fakeSource.tick).
func (c *movingClock) advance() { c.mu.Lock(); c.now = c.now.Add(50 * time.Millisecond); c.mu.Unlock() }

// sessionBars 는 시나리오 앞 18 봉(봉 17 = ARMED) 뒤에 평평한 봉 둘, 목표를 넘는 봉 20(09:20 개장 — 제안 09:19 뒤), 평평한 봉 둘.
// 봉 18 · 19 는 목표에 못 닿아 「제안 전 봉으로 라벨」 결함이 있으면 결과가 달라짐.
func sessionBars() []ohlcv {
	bars := append([]ohlcv(nil), breakoutScenario()[:18]...)
	flat := ohlcv{10070, 10080, 10060, 10075, 100}
	return append(bars, flat, flat, ohlcv{10075, 10400, 10070, 10390, 100}, flat, flat, flat)
}

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
	src := newFakeSource(sessionBars(), clk.Now)
	src.tick = clk.advance
	dir := t.TempDir()
	p := newTestProbe(t, src, clk, dir)
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	if got := src.count("StrictOrderbookTop"); got != 1 {
		t.Fatalf("session start read the book %d times; want 1 (the structural seal)", got)
	}
	runMinutes(t, p, clk, 1, 23)

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
	if len(labels) != 1 || labels[0].Outcome != "target" || labels[0].SetupID != proposed[0].SetupID ||
		labels[0].ExitOpenAt != testOpen.Add(20*time.Minute).UTC().Format(time.RFC3339) {
		t.Fatalf("labels = %+v; want one target on the bar that opened 09:20, after the 09:19 proposal", labels)
	}
	if want := uint64(testOpen.Add(20 * time.Minute).UnixMilli()); proposed[0].LabelFromMS != want {
		t.Fatalf("label_from_ms = %d; want the 09:20 open %d", proposed[0].LabelFromMS, want)
	}
	// 호가: 시작 1 + ARMED 시도 1 — 그 뒤에는 다시 읽지 않는다(setup 이 끝남).
	if got := src.count("StrictOrderbookTop"); got != 2 {
		t.Fatalf("book reads = %d; want 2", got)
	}
	// 분마다 정확히 1 페이지(폴 하한). 봉 k 는 k+2 분에 적재 — ARMED(봉 17)는 09:19, 라벨(봉 20)은 09:22,
	// 그 뒤로는 그 종목을 더 폴하지 않음(09:23 은 폴 0).
	if got, want := src.count("StrictMinuteCandles"), 22; got != want {
		t.Fatalf("candle pages = %d; want %d (one page per symbol-minute until the label)", got, want)
	}
}

func TestARestartDoesNotProposeTheSameSetupAgainAndStillLabels(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	src := newFakeSource(sessionBars(), clk.Now)
	src.tick = clk.advance
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
	runMinutes(t, second, clk, 20, 23)
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
	src.tick = clk.advance
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

// TestAGapIsOneRowPerSymbolStageAndMinute 는 결손 행의 단위(종목 · 단계 · 분 — spec 시나리오)와 계수를 고정함.
func TestAGapIsOneRowPerSymbolStageAndMinute(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(5 * time.Second)}
	dir := t.TempDir()
	p := newTestProbe(t, newFakeSource(nil, clk.Now), clk, dir)
	for i := 0; i < 3; i++ {
		if err := p.gap("005930", "bars", errors.New("READER_ERROR reading page 1: rate limited")); err != nil {
			t.Fatal(err)
		}
	}
	clk.set(testOpen.Add(time.Minute + 5*time.Second))
	if err := p.gap("005930", "bars", errors.New("READER_ERROR reading page 1: rate limited")); err != nil {
		t.Fatal(err)
	}
	if err := p.gap("005930", "quote", errors.New("QUOTE_READER_ERROR")); err != nil {
		t.Fatal(err)
	}
	var minutes []string
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindGap {
			minutes = append(minutes, row.Stage+"@"+row.Minute)
		}
	}
	want := []string{"bars@" + testOpen.UTC().Format(time.RFC3339), "bars@" + testOpen.Add(time.Minute).UTC().Format(time.RFC3339),
		"quote@" + testOpen.Add(time.Minute).UTC().Format(time.RFC3339)}
	if strings.Join(minutes, ",") != strings.Join(want, ",") || p.counts["gap:bars"] != 4 || p.counts["gap:quote"] != 1 {
		t.Fatalf("gap rows %v, counts %v; want %v and bars 4 · quote 1", minutes, p.counts, want)
	}
}

// TestAnArmedSetupExpiresByTimeWhenQuotesKeepFailing 는 ARMED 창이 시도 수가 아니라 시간임을 고정함(코드 리뷰 P2-2).
func TestAnArmedSetupExpiresByTimeWhenQuotesKeepFailing(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	bars := breakoutScenario()[:18]
	for i := 0; i < 10; i++ {
		bars = append(bars, ohlcv{10070, 10075, 10065, 10070, 100})
	}
	src := newFakeSource(bars, clk.Now)
	src.tick = clk.advance
	dir := t.TempDir()
	p := newTestProbe(t, src, clk, dir)
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	src.failAlways["StrictOrderbookTop"] = true // 시작 뒤 모든 호가 읽기가 실패
	runMinutes(t, p, clk, 1, 27)
	expired := 0
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindSignal && row.Done == doneArmedExpired {
			expired++
			if row.Attempt != 0 || row.ArmedSince == "" {
				t.Errorf("expiry row %+v; want no attempt and the armed_since instant", row)
			}
		}
		if row.Kind == kindSignal && row.Phase == "PROPOSED" {
			t.Fatalf("a proposal without a fresh quote: %+v", row)
		}
	}
	if expired != 1 {
		t.Fatalf("armed_expired rows = %d; want exactly 1 once five minutes passed", expired)
	}
}

// TestABarGroupGiveUpSkipsTheRestOfTheMinute 는 봉 그룹 재시도를 다 쓰면 그 분의 나머지 종목 봉 폴을 건너뜀을 고정함(§2.1-10).
func TestABarGroupGiveUpSkipsTheRestOfTheMinute(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	src := newFakeSource(breakoutScenario(), clk.Now)
	src.ranking = append(src.ranking, domain.RankingItem{Rank: 2, Symbol: "000660", Currency: "KRW", LastPrice: 10000})
	dir := t.TempDir()
	p := newTestProbe(t, src, clk, dir)
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	src.failAlways["StrictMinuteCandles"] = true
	runMinutes(t, p, clk, 1, 1)
	if got := src.count("StrictMinuteCandles"); got != 3 {
		t.Fatalf("candle calls = %d; want 3 (one symbol's attempts, the other skipped)", got)
	}
	skipped := false
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindGap && row.Symbol == "000660" && strings.HasPrefix(row.Reason, "skipped") {
			skipped = true
		}
	}
	if !skipped {
		t.Fatal("the skipped symbol left no gap row")
	}
}

// TestARankingGiveUpLeavesAGapRow 는 랭킹 포기가 결손 행으로 남음을 고정함(spec — 코드 리뷰 P2-5).
func TestARankingGiveUpLeavesAGapRow(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen}
	src := newFakeSource(nil, clk.Now)
	src.failAlways["Rankings"] = true
	dir := t.TempDir()
	p := newTestProbe(t, src, clk, dir)
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err == nil {
		t.Fatal("the session started without a universe")
	}
	rows := rowsOf(t, dir)
	if len(rows) != 1 || rows[0].Kind != kindGap || rows[0].Stage != "ranking" {
		t.Fatalf("rows = %+v; want one ranking gap", rows)
	}
}

// TestAnotherLotInTheSameLedgerIsRefused 는 설정이 다른 판의 행을 섞지 않음을 고정함(코드 리뷰 P3-2).
func TestAnotherLotInTheSameLedgerIsRefused(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen}
	dir := t.TempDir()
	first := newTestProbe(t, newFakeSource(nil, clk.Now), clk, dir)
	if err := first.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	second := newTestProbe(t, newFakeSource(nil, clk.Now), clk, dir)
	second.cfg.Params.RetestTolerancePPM = 100_000
	if err := second.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err == nil || !strings.Contains(err.Error(), "lot") {
		t.Fatalf("a different lot joined the session: %v", err)
	}
}

// TestALabelSeesPastAGapAfterTheProposal 는 라벨이 이어진 앞부분이 아니라 전체 묶음을 봄을 고정함(코드 리뷰 P2-6) — 제안 뒤에
// 빈 분이 생겨도 그 뒤의 목표 도달을 잃지 않음.
func TestALabelSeesPastAGapAfterTheProposal(t *testing.T) {
	t.Parallel()
	clk := &movingClock{now: testOpen.Add(-time.Minute)}
	src := newFakeSource(sessionBars(), clk.Now)
	src.tick = clk.advance
	src.missing = map[int]bool{19: true} // 09:19 봉이 오지 않음 — 제안(09:19) 뒤의 빈 분
	dir := t.TempDir()
	p := newTestProbe(t, src, clk, dir)
	if err := p.startSession(context.Background(), testCalendar(t), sessionSourceCalendar); err != nil {
		t.Fatal(err)
	}
	runMinutes(t, p, clk, 1, 23)
	labels := 0
	for _, row := range rowsOf(t, dir) {
		if row.Kind == kindLabel && row.Outcome == "target" {
			labels++
		}
	}
	if labels != 1 {
		t.Fatalf("target labels = %d; want 1 — the 09:20 bar past the gap reaches the target", labels)
	}
}
