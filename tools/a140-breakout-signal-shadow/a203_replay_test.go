package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// ---- 시간대 거절(결정 M3) — 경계 양쪽 · 서머타임 양쪽 ----

func TestA203RefusalWindowBoundaries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		at      time.Time
		refused bool
	}{
		{"wed 07:59 KST (US closed: tue 17:59 EST)", kst(2026, 11, 25, 7, 59), false},
		{"wed 08:00 KST", kst(2026, 11, 25, 8, 0), true},
		{"wed 19:59 KST", kst(2026, 11, 25, 19, 59), true},
		{"wed 20:00 KST (US: wed 06:00 EST)", kst(2026, 11, 25, 20, 0), false},
		{"wed 23:29 KST = 09:29 EST", kst(2026, 11, 25, 23, 29), false},
		{"wed 23:30 KST = 09:30 EST open", kst(2026, 11, 25, 23, 30), true},
		{"thu 05:59 KST = wed 15:59 EST", kst(2026, 11, 26, 5, 59), true},
		{"thu 06:00 KST = wed 16:00 EST close", kst(2026, 11, 26, 6, 0), false},
		{"jul wed 22:29 KST = 09:29 EDT", kst(2026, 7, 15, 22, 29), false},
		{"jul wed 22:30 KST = 09:30 EDT open", kst(2026, 7, 15, 22, 30), true},
		{"jul thu 04:59 KST = wed 15:59 EDT", kst(2026, 7, 16, 4, 59), true},
		{"jul thu 05:00 KST = wed 16:00 EDT close", kst(2026, 7, 16, 5, 0), false},
		{"sat 05:59 KST = fri 15:59 EST", kst(2026, 11, 28, 5, 59), true},
		{"sat 06:00 KST = fri 16:00 EST", kst(2026, 11, 28, 6, 0), false},
		{"jul sat 04:59 KST = fri 15:59 EDT", kst(2026, 7, 18, 4, 59), true},
		{"jul sat 05:00 KST = fri 16:00 EDT", kst(2026, 7, 18, 5, 0), false},
		{"sat 10:00 KST", kst(2026, 11, 28, 10, 0), false},
		{"sun 23:30 KST = sun 09:30 EST", kst(2026, 11, 29, 23, 30), false},
		{"mon 07:59 KST = sun 17:59 EST", kst(2026, 11, 30, 7, 59), false},
		// 미국 서머타임 시작(2026-03-08 일) 다음 월요일 · 끝(2026-11-01 일) 다음 월요일.
		{"dst start mon 22:29 KST", kst(2026, 3, 9, 22, 29), false},
		{"dst start mon 22:30 KST = 09:30 EDT", kst(2026, 3, 9, 22, 30), true},
		{"dst end mon 22:30 KST = 08:30 EST", kst(2026, 11, 2, 22, 30), false},
		{"dst end mon 23:30 KST = 09:30 EST", kst(2026, 11, 2, 23, 30), true},
		{"fri 07:59:59 KST", kst(2026, 11, 27, 7, 59).Add(59 * time.Second), false},
	}
	for _, c := range cases {
		if got, why := a203Refused(c.at); got != c.refused {
			t.Errorf("%s: refused=%v (%s), want %v", c.name, got, why, c.refused)
		}
	}
	if next := a203NextChange(kst(2026, 11, 28, 5, 0)); !next.Equal(kst(2026, 11, 28, 6, 0)) {
		t.Errorf("next change from sat 05:00 KST = %s, want 06:00", next.In(seoul))
	}
}

// ---- 전송 관문: 간격 · 캐시 · 정지 ----

func newTestFetcher(t *testing.T, live a203Live, clock *a203Clock, root string, maxGets int) *a203Fetcher {
	t.Helper()
	if live == nil {
		return newA203Fetcher(nil, root, clock.now, clock.sleep, time.Second, maxGets)
	}
	return newA203Fetcher(live, root, clock.now, clock.sleep, time.Second, maxGets)
}

func TestA203FetcherPacesCachesAndResumes(t *testing.T) {
	t.Parallel()
	clock := &a203Clock{t: kst(2026, 11, 28, 10, 0)}
	live := newA203FakeLive(clock)
	root := t.TempDir()
	f := newTestFetcher(t, live, clock, root, 100)
	for _, date := range []string{"2026-11-25", "2026-11-24", "2026-11-23"} {
		if _, err := f.calendar(context.Background(), date); err != nil {
			t.Fatal(err)
		}
	}
	if live.count() != 3 || f.gets != 3 {
		t.Fatalf("live calls %d · gets %d, want 3", live.count(), f.gets)
	}
	for i := 1; i < len(live.sentAt); i++ {
		if gap := live.sentAt[i].Sub(live.sentAt[i-1]); gap < time.Second {
			t.Fatalf("send %d came %s after the previous one, want >= 1s", i, gap)
		}
	}
	// 이어하기: 같은 캐시로 다시 — 받은 것은 GET 0, 새 날짜만 보냄.
	g := newTestFetcher(t, live, clock, root, 100)
	for _, date := range []string{"2026-11-25", "2026-11-24", "2026-11-23", "2026-11-20"} {
		if _, err := g.calendar(context.Background(), date); err != nil {
			t.Fatal(err)
		}
	}
	if live.count() != 4 || g.gets != 1 || g.hits != 3 {
		t.Fatalf("resume: live calls %d · gets %d · hits %d, want 4 · 1 · 3", live.count(), g.gets, g.hits)
	}
	// 재생(live 없음)은 캐시에 없는 것을 보내지 않고 errA203NotCached.
	off := newTestFetcher(t, nil, clock, root, 100)
	if _, err := off.calendar(context.Background(), "2026-11-19"); !errors.Is(err, errA203NotCached) {
		t.Fatalf("offline miss = %v", err)
	}
	info, err := os.Stat(filepath.Join(root, "calendar", "2026-11-25.json"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("cache file mode %v err %v, want 0600", info.Mode().Perm(), err)
	}
	if dir, _ := os.Stat(filepath.Join(root, "calendar")); dir.Mode().Perm() != 0o700 {
		t.Fatalf("cache dir mode %v, want 0700", dir.Mode().Perm())
	}
}

func TestA203FetcherRefusesInsideTheWindowAfterWaiting(t *testing.T) {
	t.Parallel()
	// 07:59:59.5 KST 수요일 첫 전송은 허용, 1 초 간격을 기다리면 08:00:00.5 — 기다린 뒤에 보므로 보내지 않음.
	clock := &a203Clock{t: kst(2026, 11, 25, 7, 59).Add(59500 * time.Millisecond)}
	live := newA203FakeLive(clock)
	f := newTestFetcher(t, live, clock, t.TempDir(), 100)
	if _, err := f.calendar(context.Background(), "2026-11-24"); err != nil {
		t.Fatal(err)
	}
	_, err := f.calendar(context.Background(), "2026-11-23")
	var stop *a203StopError
	if !errors.As(err, &stop) || !strings.Contains(stop.Reason, "time window refused") || live.count() != 1 {
		t.Fatalf("err=%v live calls %d, want a time-window stop after 1 send", err, live.count())
	}
	// 멈춘 뒤에는 시각이 풀려도 같은 실행에서는 보내지 않음(sticky).
	clock.t = kst(2026, 11, 28, 10, 0)
	if _, err := f.calendar(context.Background(), "2026-11-20"); !errors.As(err, &stop) || live.count() != 1 {
		t.Fatalf("after a stop: err=%v live calls %d", err, live.count())
	}
}

func TestA203FetcherStopsOnTheFirstErrorAndCap(t *testing.T) {
	t.Parallel()
	clock := &a203Clock{t: kst(2026, 11, 28, 10, 0)}
	live := newA203FakeLive(clock)
	live.failOn[2] = fmt.Errorf("fake: %w", official.ErrRateLimited)
	root := t.TempDir()
	f := newTestFetcher(t, live, clock, root, 100)
	if _, err := f.calendar(context.Background(), "2026-11-25"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.calendar(context.Background(), "2026-11-24"); !errors.Is(err, official.ErrRateLimited) {
		t.Fatalf("second = %v, want the rate limit", err)
	}
	if _, err := f.calendar(context.Background(), "2026-11-23"); err == nil || live.count() != 2 {
		t.Fatalf("after 429: err=%v live calls %d, want no third send", err, live.count())
	}
	if _, err := os.Stat(filepath.Join(root, "calendar", "2026-11-24.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a failed GET was cached: %v", err)
	}
	capped := newTestFetcher(t, live, clock, t.TempDir(), 1)
	if _, err := capped.calendar(context.Background(), "2026-11-25"); err != nil {
		t.Fatal(err)
	}
	var stop *a203StopError
	if _, err := capped.calendar(context.Background(), "2026-11-24"); !errors.As(err, &stop) || !strings.Contains(stop.Reason, "cap 1") {
		t.Fatalf("cap: %v", err)
	}
}

func TestA203MinuteCacheDropsTheRateBudget(t *testing.T) {
	t.Parallel()
	clock := &a203Clock{t: kst(2026, 11, 28, 10, 0)}
	live := &budgetPage{}
	root := t.TempDir()
	f := newTestFetcher(t, live, clock, root, 10)
	page, err := f.minute(context.Background(), "KR", "TST001", 200, "2026-08-14T16:00:00.000+09:00")
	if err != nil || !page.Budget.Exhausted() {
		t.Fatalf("live page: %v exhausted %v — the collector must still see the live budget", err, page.Budget.Exhausted())
	}
	off := newTestFetcher(t, nil, clock, root, 10)
	cached, err := off.minute(context.Background(), "KR", "TST001", 200, "2026-08-14T16:00:00.000+09:00")
	if err != nil || cached.Budget.Exhausted() || cached.BodyDigest != page.BodyDigest || len(cached.Candles) != 1 {
		t.Fatalf("cached page: %v exhausted %v digest %q", err, cached.Budget.Exhausted(), cached.BodyDigest)
	}
	if _, err := off.minute(context.Background(), "KR", "../x00", 200, "x"); err == nil {
		t.Fatal("a path-escaping symbol was accepted")
	}
}

type budgetPage struct{ a203FakeLive }

func (b *budgetPage) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	return official.StrictMinutePage{Market: market, Symbol: symbol, Terminal: true, StatusCode: 200, BodyDigest: digestFor("b"),
		Candles: []official.RawMinuteCandle{{Timestamp: "2026-08-14T09:01:00.000+09:00", Open: "1", High: "1", Low: "1", Close: "1", Volume: "1", Currency: "KRW"}},
		Budget:  official.RateBudget{Reported: true, Limit: 10, Remaining: 0}}, nil
}

// ---- M9 · 종료 코드 · dispatch ----

func TestA203PopulationMismatchSendsNothingAndExitsTwo(t *testing.T) {
	stateDir := t.TempDir()
	a203WritePA(t, stateDir, fakeCodes(3))
	a203PAFingerprint = strings.Repeat("0", 64) // 사전등록 지문과 다른 파일
	opened := 0
	o := a203TestOptions(t, stateDir, &a203Clock{t: kst(2026, 11, 28, 10, 0)}, nil)
	o.openLive = func(ctx context.Context, _ string) (context.Context, a203Live, func() error, error) {
		opened++
		return ctx, nil, nil, errors.New("must not open")
	}
	for _, command := range []string{"sessions", "collect", "report"} {
		o.command = command
		if err := runA203(context.Background(), o); !errors.Is(err, errA203Population) || opened != 0 {
			t.Fatalf("%s: err=%v opened %d", command, err, opened)
		}
	}
	code := -1
	old := a203Exit
	a203Exit = func(c int) { code = c }
	t.Cleanup(func() { a203Exit = old })
	err := replayCommand(context.Background(), []string{"sessions", "--state-dir", stateDir, "--dry-run"}, io.Discard)
	if code != 2 || !errors.Is(err, errA203Population) {
		t.Fatalf("exit code %d err %v, want 2", code, err)
	}
}

func TestA203DispatchRoutesReplay(t *testing.T) {
	t.Parallel()
	if err := dispatch(context.Background(), nil, io.Discard); err == nil || !strings.Contains(err.Error(), `"replay"`) {
		t.Fatalf("no args: %v", err)
	}
	if err := dispatch(context.Background(), []string{"bogus"}, io.Discard); err == nil || !strings.Contains(err.Error(), `"replay"`) {
		t.Fatalf("unknown: %v", err)
	}
	if err := dispatch(context.Background(), []string{"replay"}, io.Discard); err == nil || !strings.Contains(err.Error(), "replay needs a step") {
		t.Fatalf("replay: %v", err)
	}
	if err := dispatch(context.Background(), []string{"report", "--ledger", filepath.Join(t.TempDir(), "none.jsonl")}, io.Discard); err != nil {
		t.Fatalf("report still routes: %v", err)
	}
	if _, err := parseA203Options([]string{"collect", "--sample", "main", "--sessions", "x"}, io.Discard); err == nil || !strings.Contains(err.Error(), "--config-dir") {
		t.Fatalf("collect without --config-dir: %v", err)
	}
}

// ---- 끝에서 끝: sessions → collect(control · main) → report, 가짜 브로커 ----

func a203TestOptions(t *testing.T, stateDir string, clock *a203Clock, live *a203FakeLive) a203Options {
	t.Helper()
	o := a203Options{stateDir: stateDir, configDir: "fake-profile", anchor: "2026-10-23", mainSessions: 3, warmup: 21, maxGets: 10_000,
		now: clock.now, sleep: clock.sleep, stdout: io.Discard, stderr: io.Discard,
		sessionsFile: filepath.Join(stateDir, "sessions.txt"), a140Ledger: filepath.Join(stateDir, "a140.jsonl")}
	o.openLive = func(ctx context.Context, dir string) (context.Context, a203Live, func() error, error) {
		if dir != "fake-profile" {
			return ctx, nil, nil, errors.New("unexpected profile")
		}
		return ctx, live, func() error { return nil }, nil
	}
	return o
}

// a203Scenario 는 TST001 의 하루(40 봉): opening range → 돌파 → 되돌림 → 되찾음(ARMED, k=17) → 봉 19 시가 = 진입가 → 봉 20 이 목표를 넘음.
func a203Scenario() []ohlcv {
	bars := breakoutScenario()[:18]
	bars = append(bars, ohlcv{10070, 10080, 10060, 10075, 100}, ohlcv{10070, 10080, 10060, 10075, 100}, ohlcv{10075, 10400, 10070, 10390, 100})
	for len(bars) < 40 {
		bars = append(bars, ohlcv{10390, 10400, 10380, 10390, 100})
	}
	return bars
}

func writeA140Fixture(t *testing.T, path string, universe []string) {
	t.Helper()
	led, err := openLedger(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	defer led.Close()
	params := testParams()
	params.NotionalCapKRW = 500_000
	for _, date := range a203ControlDates {
		rows := []ledgerRow{{Kind: kindUniverse, Session: "KRX:" + date, Symbols: universe, Params: &params}}
		for _, code := range universe {
			phase := "RANGE_LOCKED"
			if code == "TST001" {
				phase = "PROPOSED"
			}
			rows = append(rows, ledgerRow{Kind: kindSignal, Session: "KRX:" + date, Symbol: code, Phase: phase, SetupID: date + code, Ask: 10070, Worst: 10070, RoundTrip: 51})
		}
		rows = append(rows, ledgerRow{Kind: kindLabel, Session: "KRX:" + date, Symbol: "TST001", SetupID: date + "TST001", ExitMinor: 10378})
		for _, row := range rows {
			if err := led.write(row); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.Rename(filepath.Join(filepath.Dir(path), ledgerFileName), path); err != nil {
		t.Fatal(err)
	}
}

func TestA203EndToEndWithAFakeBroker(t *testing.T) {
	stateDir := t.TempDir()
	codes := fakeCodes(5)
	a203WritePA(t, stateDir, codes[:4])
	clock := &a203Clock{t: kst(2026, 11, 28, 10, 0)} // 토요일 — 허용 창
	live := newA203FakeLive(clock)
	live.closeAt = [2]int{9, 40} // 짧은 세션(40 봉) — 적재 시간만 줄임, 페이지 1 장
	live.ranking = []string{"TST004", "TST005"}
	live.bars = func(symbol, date string) []ohlcv {
		if symbol == "TST001" {
			return a203Scenario()
		}
		return nil
	}
	live.volume = func(symbol string, day int) float64 { return float64(1000 + 100*int(symbol[5]-'0') + 10*(day%7)) }
	o := a203TestOptions(t, stateDir, clock, live)
	writeA140Fixture(t, o.a140Ledger, codes[:3])
	ctx := context.Background()

	// 1) dry-run 은 전송 0.
	var out bytes.Buffer
	o.command, o.dryRun, o.stdout = "sessions", true, &out
	if err := runA203(ctx, o); err != nil || live.count() != 0 || !strings.Contains(out.String(), "sends 0") {
		t.Fatalf("sessions dry-run: %v calls %d\n%s", err, live.count(), out.String())
	}
	// 2) sessions — 세션마다 달력 1 회.
	o.dryRun, o.stdout = false, io.Discard
	if err := runA203(ctx, o); err != nil {
		t.Fatal(err)
	}
	sessionsGets := live.count()
	if sessionsGets != 3+5+21 {
		t.Fatalf("calendar GETs %d, want 29", sessionsGets)
	}
	listed, _ := os.ReadFile(o.sessionsFile)
	if !strings.Contains(string(listed), "2026-10-16 control 09:00 09:40") || !strings.Contains(string(listed), "2026-10-15 main") {
		t.Fatalf("sessions file:\n%s", listed)
	}
	// 3) collect dry-run(전송 0) → collect control → collect main → 다시 collect(GET 0).
	for _, sample := range []string{a203SampleControl, a203SampleMain} {
		o.command, o.sample, o.dryRun = "collect", sample, true
		out.Reset()
		o.stdout = &out
		before := live.count()
		if err := runA203(ctx, o); err != nil || live.count() != before {
			t.Fatalf("collect %s dry-run: %v sent %d", sample, err, live.count()-before)
		}
		o.dryRun, o.stdout = false, io.Discard
		if err := runA203(ctx, o); err != nil {
			t.Fatalf("collect %s: %v", sample, err)
		}
	}
	collected := live.count()
	if err := runA203(ctx, o); err != nil || live.count() != collected {
		t.Fatalf("re-collect sent %d", live.count()-collected)
	}
	// 4) report — 캐시만, 전송 0.
	o.openLive = nil
	for _, sample := range []string{a203SampleControl, a203SampleMain} {
		o.command, o.sample, o.out = "report", sample, filepath.Join(stateDir, "report-"+sample+".txt")
		if err := runA203(ctx, o); err != nil {
			t.Fatalf("report %s: %v", sample, err)
		}
		raw, _ := os.ReadFile(o.out)
		text := string(raw)
		if !strings.HasPrefix(text, a203QuoteLimitLine+"\n"+a203UniverseLimitLine+"\n") {
			t.Fatalf("report %s does not start with the two limit lines:\n%s", sample, text)
		}
		for _, code := range codes {
			if strings.Contains(text, code) {
				t.Fatalf("report %s names the symbol %s", sample, code)
			}
		}
		if sample == a203SampleControl && !strings.Contains(text, "ARMED 전 종결 일치 15 / 15") {
			t.Fatalf("control report:\n%s", text)
		}
		if sample == a203SampleMain && !strings.Contains(text, "P0 판정: INSUFFICIENT") {
			t.Fatalf("main report:\n%s", text)
		}
	}
	if live.count() != collected {
		t.Fatalf("report sent %d", live.count()-collected)
	}
	// a204 추가분: main 첫 세션 d 의 종목군 d−1 페이지가 캐시에 있음, 종목군 파일이 있음.
	var uf a203UniverseFile
	raw, err := os.ReadFile(filepath.Join(stateDir, "universe", "uc-main.json"))
	if err != nil || json.Unmarshal(raw, &uf) != nil || uf.Prior["2026-10-13"] != "2026-10-12" || len(uf.Universe["2026-10-13"]) != 5 {
		t.Fatalf("universe file: %v %+v", err, uf)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "cache", "minute", "TST005", "2026-10-12T101000.000+0900-c200.json")); err != nil {
		t.Fatalf("a204 d-1 first page not cached: %v", err)
	}
	// 실행 기록: sessions 1 · collect control 1 · collect main 2(dry-run · report 는 기록 안 함).
	if runs, err := a203ReadRuns(stateDir); err != nil || len(runs) != 4 || runs[3].Gets != 0 {
		t.Fatalf("runs %v %+v", err, runs)
	}
}

func TestA203CheckpointOnARefusedRun(t *testing.T) {
	stateDir := t.TempDir()
	a203WritePA(t, stateDir, fakeCodes(2))
	clock := &a203Clock{t: kst(2026, 11, 27, 10, 0)} // 금요일 10:00 KST — 거절
	live := newA203FakeLive(clock)
	o := a203TestOptions(t, stateDir, clock, live)
	o.command = "sessions"
	err := runA203(context.Background(), o)
	var stop *a203StopError
	if !errors.As(err, &stop) || live.count() != 0 {
		t.Fatalf("err=%v calls %d", err, live.count())
	}
	var cp a203RunRecord
	raw, err := os.ReadFile(filepath.Join(stateDir, "checkpoint.json"))
	if err != nil || json.Unmarshal(raw, &cp) != nil || !strings.Contains(cp.Stop, "time window refused") || cp.Gets != 0 {
		t.Fatalf("checkpoint %v %+v", err, cp)
	}
}

// 수집 중 캐시 파일을 못 읽으면(깨진 파일) 결손으로 세고 넘어가지 않고 멈춤 — 이어하기마다 같은 자리가 영구 결손이 되지 않게(리뷰 지적).
func TestA203CollectStopsOnAnUnreadableCacheFile(t *testing.T) {
	stateDir := t.TempDir()
	codes := fakeCodes(3)
	a203WritePA(t, stateDir, codes)
	clock := &a203Clock{t: kst(2026, 11, 28, 10, 0)}
	live := newA203FakeLive(clock)
	live.closeAt = [2]int{9, 40}
	o := a203TestOptions(t, stateDir, clock, live)
	writeA140Fixture(t, o.a140Ledger, codes[:1])
	ctx := context.Background()
	o.command = "sessions"
	if err := runA203(ctx, o); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(stateDir, "cache", "minute", "TST001", "2026-10-15T101000.000+0900-c200.json")
	if err := os.MkdirAll(filepath.Dir(broken), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(broken, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	o.command, o.sample = "collect", a203SampleMain
	if err := runA203(ctx, o); err == nil || !strings.Contains(err.Error(), "unreadable") {
		t.Fatalf("collect over a broken cache file: %v", err)
	}
}
