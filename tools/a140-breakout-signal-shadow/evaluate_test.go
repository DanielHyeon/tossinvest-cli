package main

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// ---- 사이징 ----

func TestKRTickMatchesTheKRXScheduleUsedElsewhere(t *testing.T) {
	t.Parallel()
	// 같은 표의 셋째 사본이므로 기존 사본(verifylive.TickSize)과 경계마다 대조해 두 쪽을 함께 못 박음.
	for _, price := range []uint64{1, 1999, 2000, 4999, 5000, 19999, 20000, 49999, 50000, 199999, 200000, 499999, 500000, 2_000_000} {
		want := uint64(verifylive.TickSize(float64(price), "KR"))
		if got := krTickMinor(price); got != want {
			t.Errorf("krTickMinor(%d) = %d, verifylive says %d", price, got, want)
		}
	}
}

func TestCostRateIsTheKRRoundTripInPPM(t *testing.T) {
	t.Parallel()
	got, err := costRatePPM(costs.DefaultModel())
	// 매수 수수료 0.1% + 매도 수수료 0.1% + 매도 거래세 0.3% = 5,000 ppm(부동소수 잡음이 5,001 을 만들지 않아야 함).
	if err != nil || got != 5_000 {
		t.Fatalf("costRatePPM = %d, %v; want 5,000", got, err)
	}
}

func TestCostRateDropsFloatNoiseButKeepsRealFractions(t *testing.T) {
	t.Parallel()
	// 0.00102×1e6 = 1020.0000000000001 — 잡음으로 1 ppm 이 더 붙으면 안 됨. 0.0010005 는 진짜 소수(1000.5)라 올림.
	model, err := costs.NewModel(map[string]string{costs.KeyKRBuyCommissionRate: "0.00102", costs.KeyKRSellCommissionRate: "0.0010005",
		costs.KeyKRSellTaxRate: "0.003"})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := costRatePPM(model); err != nil || got != 1020+1001+3000 {
		t.Fatalf("costRatePPM = %d, %v; want 5,021", got, err)
	}
}

func TestTheTickIsTheOpeningRangeResistanceBand(t *testing.T) {
	t.Parallel()
	// 저항 20,000(틱 50)인데 마지막 종가는 19,900(틱 10) — 틱은 저항에서 정해야 세션 동안 고정됨.
	var bars []ohlcv
	for i := 0; i < 15; i++ {
		bars = append(bars, ohlcv{19950, 20000, 19900, 19950, 100})
	}
	bars = append(bars, ohlcv{19950, 19960, 19890, 19900, 100})
	got, err := evaluateSetup(seriesOf(bars), testSessionInfo(), "005930", testParams(), freshQuote(testOpen), uint64(testOpen.Add(17*time.Minute).UnixMilli()))
	if err != nil || got.Tick != 50 {
		t.Fatalf("tick = %d (%v); want 50 — the band of the 20,000 resistance", got.Tick, err)
	}
}

func TestPlanSizingBuildsTheTargetFromTheNetRR(t *testing.T) {
	t.Parallel()
	plan, err := planSizing(10070, 10070, 9950, 5_000, 1_500_000)
	if err != nil {
		t.Fatal(err)
	}
	// 왕복 = ceil(10070×0.005) = 51, 위험 = 10070 − 9950 + 51 = 171, 목표 = 10070 + 51 + ceil(171×1.5 = 256.5) = 10,378.
	if plan.Worst != 10070 || plan.RoundTrip != 51 || plan.Risk != 171 || plan.Target != 10378 {
		t.Fatalf("plan = %+v", plan)
	}
	// ask 가 진입 제안가보다 높으면 worst 는 ask.
	plan, err = planSizing(10070, 10080, 9950, 5_000, 1_500_000)
	if err != nil || plan.Worst != 10080 {
		t.Fatalf("plan with a higher ask = %+v, %v", plan, err)
	}
	if _, err := planSizing(10070, 10070, 10070, 5_000, 1_500_000); err == nil {
		t.Fatal("a stop at the entry was accepted")
	}
}

// ---- 평가 ----

func TestTheScenarioWalksTheFunnelAndProposesOnAFreshQuote(t *testing.T) {
	t.Parallel()
	bars := breakoutScenario()
	stale := freshQuote(testOpen) // 개장 때 봉인한 호가 — ARMED 판정에는 묵음
	cases := []struct {
		bars  int
		phase string
	}{{14, ""}, {15, "RANGE_LOCKED"}, {16, "RETEST_WAIT"}, {17, "RETEST_WAIT"}, {18, "ARMED"}}
	for _, tc := range cases {
		now := testOpen.Add(time.Duration(tc.bars+1) * time.Minute)
		got, err := evaluateSetup(seriesOf(bars[:tc.bars]), testSessionInfo(), "005930", testParams(), stale, uint64(now.UnixMilli()))
		if tc.phase == "" {
			if err == nil || got.Phase != "" {
				t.Errorf("%d bars: want a refusal before the opening range closes, got %+v, %v", tc.bars, got, err)
			}
			continue
		}
		if err != nil || got.Phase != tc.phase {
			t.Errorf("%d bars: phase %q (%v), want %q", tc.bars, got.Phase, err, tc.phase)
		}
	}
	// ARMED 를 묵은 호가로 보면 QUOTE_STALE — 최종 판정이 아님.
	at := testOpen.Add(19 * time.Minute)
	armed, _ := evaluateSetup(seriesOf(bars[:18]), testSessionInfo(), "005930", testParams(), stale, uint64(at.UnixMilli()))
	if armed.Refusal != "QUOTE_STALE" || armed.ArmedIndex != 17 {
		t.Fatalf("stale ARMED = %+v; want QUOTE_STALE at bar 17", armed)
	}
	// 새 호가 → PROPOSED, 계획 숫자는 sizing 시험과 같다.
	proposed, err := evaluateSetup(seriesOf(bars[:18]), testSessionInfo(), "005930", testParams(), freshQuote(at), uint64(at.UnixMilli()))
	if err != nil || proposed.Phase != "PROPOSED" || proposed.Refusal != "" {
		t.Fatalf("fresh ARMED = %+v, %v; want PROPOSED", proposed, err)
	}
	if proposed.Plan.Entry != 10070 || proposed.Plan.Stop != 9950 || proposed.Plan.Target != 10378 || proposed.Quantity != 99 {
		t.Fatalf("plan = %+v quantity %d; want entry 10,070 · stop 9,950 · target 10,378 · 99 shares (notional 1,000,000 / 10,070)", proposed.Plan, proposed.Quantity)
	}
	if proposed.Tick != 10 || proposed.ATRMinor != 100 || proposed.Resistance != 10050 || proposed.RangeLow != 9950 {
		t.Fatalf("tick %d atr %d resistance %d low %d", proposed.Tick, proposed.ATRMinor, proposed.Resistance, proposed.RangeLow)
	}
	if proposed.SpreadPPM != 994 || proposed.DriftPPM != 0 || proposed.AgeMS != 1000 {
		t.Fatalf("measured spread %d drift %d age %d; want 994 · 0 · 1000", proposed.SpreadPPM, proposed.DriftPPM, proposed.AgeMS)
	}
	// 더 긴 묶음(뒤 봉이 붙음)에서도 setup 신원 · ARMED 위치가 같다.
	later, _ := evaluateSetup(seriesOf(bars), testSessionInfo(), "005930", testParams(), freshQuote(at), uint64(at.UnixMilli()))
	if later.SetupID != proposed.SetupID || later.ArmedIndex != 17 {
		t.Fatalf("a longer series changed the setup %q→%q or the armed bar %d", proposed.SetupID, later.SetupID, later.ArmedIndex)
	}
}

func TestAWideQuoteIsRecordedAsSpreadTooWideWithItsMeasurement(t *testing.T) {
	t.Parallel()
	at := testOpen.Add(19 * time.Minute)
	wide := freshQuote(at)
	wide.Bid = 10000 // (10070 − 10000)/10035 = 6,976 ppm > 3,000
	got, err := evaluateSetup(seriesOf(breakoutScenario()[:18]), testSessionInfo(), "005930", testParams(), wide, uint64(at.UnixMilli()))
	if err != nil || got.Phase != "ARMED" || got.Refusal != "SPREAD_TOO_WIDE" || got.SpreadPPM != 6976 {
		t.Fatalf("wide quote = %+v, %v", got, err)
	}
}

func TestOnlyTheContiguousPrefixFromTheOpenIsEvaluated(t *testing.T) {
	t.Parallel()
	series := seriesOf(breakoutScenario())
	series.Bars = append(series.Bars[:16:16], series.Bars[17:]...) // 09:16 봉이 빠짐
	prefix, missing, gapped := contiguousPrefix(series, testOpen)
	if !gapped || len(prefix.Bars) != 16 || !missing.Equal(testOpen.Add(16*time.Minute)) {
		t.Fatalf("prefix %d bars, missing %v, gapped %v", len(prefix.Bars), missing, gapped)
	}
	whole, _, gapped := contiguousPrefix(seriesOf(breakoutScenario()), testOpen)
	if gapped || len(whole.Bars) != len(breakoutScenario()) {
		t.Fatalf("a whole series was cut to %d bars", len(whole.Bars))
	}
}

// ---- 생산 DB 가드 ----

func TestTheEvidenceGuardRefusesEveryWayToReachTheProductionFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	prodDir := filepath.Join(dir, "prod")
	mustMkdir(t, prodDir)
	prod := filepath.Join(prodDir, "evidence.db")
	mustWrite(t, prod)
	mustWrite(t, filepath.Join(prodDir, "journal.db"))
	ownDir := filepath.Join(dir, "own")
	mustMkdir(t, ownDir)

	hard := filepath.Join(ownDir, "hard.sqlite")
	if err := os.Link(prod, hard); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}
	soft := filepath.Join(ownDir, "soft.sqlite")
	if err := os.Symlink(prod, soft); err != nil {
		t.Fatal(err)
	}
	softDir := filepath.Join(dir, "linkdir")
	if err := os.Symlink(prodDir, softDir); err != nil {
		t.Fatal(err)
	}
	for name, target := range map[string]string{
		"the production file":         prod,
		"any file named evidence.db":  filepath.Join(ownDir, "evidence.db"),
		"a file next to a journal":    filepath.Join(prodDir, "shadow.sqlite"),
		"a hard link":                 hard,
		"a symlink":                   soft,
		"a path through a linked dir": filepath.Join(softDir, "shadow.sqlite"),
	} {
		if err := checkEvidencePath(target, []string{prod}); err == nil {
			t.Errorf("%s (%s) was accepted", name, target)
		}
	}
	if err := checkEvidencePath(filepath.Join(ownDir, "shadow-evidence.sqlite"), []string{prod}); err != nil {
		t.Errorf("a private new file was refused: %v", err)
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// ---- 라벨 ----

func TestTheLabelerTakesTheFirstTouchAndPrefersTheStopOnATie(t *testing.T) {
	t.Parallel()
	plan := sizingPlan{Entry: 10070, Ask: 10070, Worst: 10070, Stop: 9950, RoundTrip: 51, Risk: 171, Target: 10378}
	bars := breakoutScenario()
	got, ok := labelProposal(seriesOf(bars), 18, plan, false)
	if !ok || got.Outcome != "target" || got.ExitMinor != 10378 || got.RPPM != 1_502_923 {
		t.Fatalf("target label = %+v, %v; want target at 10,378 and R 1,502,923 ppm", got, ok)
	}
	tie := append(append([]ohlcv(nil), bars[:18]...), ohlcv{10070, 10400, 9900, 10000, 100})
	got, ok = labelProposal(seriesOf(tie), 18, plan, false)
	if !ok || got.Outcome != "stop" || got.ExitMinor != 9950 || got.RPPM >= 0 {
		t.Fatalf("tie label = %+v, %v; want stop", got, ok)
	}
	flatAfter := append(append([]ohlcv(nil), bars[:18]...), ohlcv{10070, 10100, 10060, 10090, 100})
	if _, ok := labelProposal(seriesOf(flatAfter), 18, plan, false); ok {
		t.Fatal("an untouched proposal was labeled before the session end")
	}
	got, ok = labelProposal(seriesOf(flatAfter), 18, plan, true)
	// 마감가 10,090: (10090 − 10070 − 51)/171 = −0.181… → −181,287 ppm(내림).
	if !ok || got.Outcome != "close" || got.ExitMinor != 10090 || got.RPPM != -181_287 {
		t.Fatalf("close label = %+v, %v", got, ok)
	}
}

// ---- 보고서 ----

func TestTheReportWithholdsOutcomeVerdictsBelowThirtyAndNamesTheVacuousRR(t *testing.T) {
	t.Parallel()
	rows := []ledgerRow{
		{Kind: kindUniverse, Session: testSession, Symbols: []string{"005930", "000660"}},
		{Kind: kindSignal, Session: testSession, Symbol: "005930", Phase: "RANGE_LOCKED"},
		{Kind: kindSignal, Session: testSession, Symbol: "005930", Phase: "RETEST_WAIT", RVOL12: true, RVOL20: true},
		{Kind: kindSignal, Session: testSession, Symbol: "005930", Phase: "PROPOSED", SetupID: "s1", Attempt: 1},
		{Kind: kindSignal, Session: testSession, Symbol: "000660", Phase: "ARMED", Refusal: "SPREAD_TOO_WIDE", SetupID: "s2", Attempt: 1},
		{Kind: kindLabel, Session: testSession, Symbol: "005930", SetupID: "s1", Outcome: "target", RPPM: 1_500_000},
		{Kind: kindGap, Session: testSession, Symbol: "000660", Stage: "bars", Reason: "MINUTE_GAP"},
		{Kind: kindBudget, Session: testSession, Reads: map[string]int{"MARKET_DATA_CHART": 30, "MARKET_DATA": 2}},
	}
	report := renderReport(rows)
	for _, want := range []string{testSession, "PROPOSED", "SPREAD_TOO_WIDE", "MINUTE_GAP", "n<30", "보류", "구성상 공허", "MARKET_DATA_CHART"} {
		if !strings.Contains(report, want) {
			t.Errorf("report does not mention %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "승률 100") {
		t.Errorf("report gave a win rate on one label:\n%s", report)
	}
}

func TestRPPMDoesNotOverflowOnLargePrices(t *testing.T) {
	t.Parallel()
	plan := sizingPlan{Worst: math.MaxUint32, RoundTrip: 1, Risk: 1, Stop: 1, Target: math.MaxUint32 + 10}
	if got := rMultiplePPM(math.MaxUint32+10, plan); got <= 0 {
		t.Fatalf("rMultiplePPM = %d", got)
	}
}
