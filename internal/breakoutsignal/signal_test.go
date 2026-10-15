package breakoutsignal

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// 시험 픽스처: 2026-08-14 KRX 정규장 09:00 KST 개장.
var testOpen = time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC) // 09:00 KST

const testSession = "KRX:2026-08-14"

// ohlcv 는 봉 하나의 값(minor 단위)임.
type ohlcv struct{ o, h, l, c, v uint64 }

// seriesOf 는 개장 분부터 빈 분 없이 이어진 봉 묶음을 만듦.
func seriesOf(bars ...ohlcv) strategyevidence.BarSeries {
	records := make([]strategyevidence.ClosedBarRecord, len(bars))
	for i, b := range bars {
		records[i] = recordAt(i, b)
	}
	return strategyevidence.BarSeries{Bars: records}
}

func recordAt(minute int, b ohlcv) strategyevidence.ClosedBarRecord {
	openAt := testOpen.Add(time.Duration(minute) * time.Minute)
	return strategyevidence.ClosedBarRecord{
		EvidenceID:       "ev-" + openAt.Format("1504"),
		RevisionIdentity: "r1",
		PayloadDigest:    "sha256:" + strings.Repeat("a", 64),
		Payload: strategyevidence.ClosedBar1mPayload{
			Market: "KR", Symbol: "005930", SessionID: testSession, IntervalMS: 60_000,
			OpenAtMS: uint64(openAt.UnixMilli()), Closed: true, RegularSession: true, Currency: "KRW",
			OpenMinor: b.o, HighMinor: b.h, LowMinor: b.l, CloseMinor: b.c, Volume: b.v, Revision: 1,
		},
	}
}

// flat 은 같은 봉 n 개.
func flat(n int, b ohlcv) []ohlcv {
	out := make([]ohlcv, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func refusalCode(t *testing.T, err error) RefusalCode {
	t.Helper()
	var refusal *Refusal
	if !errors.As(err, &refusal) {
		t.Fatalf("error %v is not a typed *Refusal", err)
	}
	return refusal.Code
}

// ---- 윗꼬리 비율 ----

func TestUpperWickRoundsUpAndIsZeroOnAFlatBar(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		bar  ohlcv
		want uint64
	}{
		// (110−105)/(110−100) = 0.5 정확히.
		{"exact half", ohlcv{100, 110, 100, 105, 1}, 500_000},
		// (10−7)/(10−1)·1e6 = 333,333.33… → 올림 333,334.
		{"rounds up", ohlcv{5, 10, 1, 7, 1}, 333_334},
		// 시가가 더 높으면 max(open, close) 는 시가.
		{"open above close", ohlcv{108, 110, 100, 102, 1}, 200_000},
		// 고가 = max(open, close) 면 윗꼬리 0.
		{"no wick", ohlcv{100, 110, 100, 110, 1}, 0},
		// 고가 = 저가 = 0 으로 나누지 않고 0.
		{"flat bar", ohlcv{100, 100, 100, 100, 1}, 0},
	}
	for _, tc := range cases {
		got, err := UpperWickRangePPM(Bar{OpenMinor: tc.bar.o, HighMinor: tc.bar.h, LowMinor: tc.bar.l, CloseMinor: tc.bar.c, Volume: tc.bar.v})
		if err != nil || got != tc.want {
			t.Errorf("%s: UpperWickRangePPM = %d, %v; want %d", tc.name, got, err, tc.want)
		}
	}
}

func TestUpperWickRefusesAnImpossibleBar(t *testing.T) {
	t.Parallel()
	for _, bar := range []Bar{
		{OpenMinor: 100, HighMinor: 90, LowMinor: 95, CloseMinor: 92},   // 고가 < 저가
		{OpenMinor: 100, HighMinor: 110, LowMinor: 95, CloseMinor: 120}, // 종가 > 고가
		{OpenMinor: 0, HighMinor: 110, LowMinor: 95, CloseMinor: 100},   // 가격 0
	} {
		if _, err := UpperWickRangePPM(bar); refusalCode(t, err) != RefusalBarInvalid {
			t.Errorf("bar %+v: want %s, got %v", bar, RefusalBarInvalid, err)
		}
	}
}

// ---- ATR ----

func TestOpeningRangeATRIsWilderFourteenOverTheFifteenBars(t *testing.T) {
	t.Parallel()
	// TR₁ = H−L = 10, TR₂..TR₁₄ = 10(이전 종가가 범위 안), TR₁₅ = 24(이전 종가 105 에서 고가 129 까지).
	bars := flat(14, ohlcv{100, 110, 100, 105, 10})
	bars = append(bars, ohlcv{120, 129, 120, 125, 10})
	in := barsOf(seriesOf(bars...))
	got, err := OpeningRangeATR(in)
	// ATR₁₄ = 10, ATR₁₅ = (13·10 + 24)/14 = 154/14 = 11.0 → 11.
	if err != nil || got != 11 {
		t.Fatalf("OpeningRangeATR = %d, %v; want 11", got, err)
	}
}

func TestOpeningRangeATRKeepsTheRationalUntilTheEnd(t *testing.T) {
	t.Parallel()
	// TR 이 1,2,1,2,… 이면 ATR₁₄ = 21/14 = 1.5(정수로 먼저 내리면 1).
	// ATR₁₅ = (13·1.5 + 3)/14 = 22.5/14 = 1.607… → 1. 먼저 내리면 (13·1 + 3)/14 = 1.14 → 1 이라 같은 답이 되므로
	// 큰 값으로 차이를 드러냄: TR 을 ×7 하면 ATR₁₄ = 10.5, ATR₁₅ = (136.5 + TR₁₅=21)/14 = 11.25 → 11; 먼저 내리면 (130+21)/14 = 10.78 → 10.
	var bars []ohlcv
	for i := 0; i < 14; i++ {
		width := uint64(7)
		if i%2 == 1 {
			width = 14
		}
		bars = append(bars, ohlcv{1000, 1000 + width, 1000, 1000, 1})
	}
	bars = append(bars, ohlcv{1000, 1021, 1000, 1000, 1})
	got, err := OpeningRangeATR(barsOf(seriesOf(bars...)))
	if err != nil || got != 11 {
		t.Fatalf("OpeningRangeATR = %d, %v; want 11 (a floor before the last step gives 10)", got, err)
	}
}

func TestOpeningRangeATRIsFixedOnceTheRangeIsLocked(t *testing.T) {
	t.Parallel()
	base := flat(15, ohlcv{100, 110, 100, 105, 10})
	long := append(append([]ohlcv(nil), base...), flat(25, ohlcv{100, 190, 100, 150, 10})...)
	short := append(append([]ohlcv(nil), base...), flat(5, ohlcv{100, 190, 100, 150, 10})...)
	a, errA := Derive(seriesOf(long...), testOpen)
	b, errB := Derive(seriesOf(short...), testOpen)
	if errA != nil || errB != nil || a.ATRMinor != b.ATRMinor || a.ATRMinor != 10 {
		t.Fatalf("ATR after 40 bars = %d (%v), after 20 bars = %d (%v); want both 10", a.ATRMinor, errA, b.ATRMinor, errB)
	}
}

func TestOpeningRangeATRRefusesAnIncompleteRangeAndAZeroATR(t *testing.T) {
	t.Parallel()
	if _, err := OpeningRangeATR(barsOf(seriesOf(flat(14, ohlcv{100, 110, 100, 105, 10})...))); refusalCode(t, err) != RefusalOpeningRangeIncomplete {
		t.Errorf("14 bars: want %s, got %v", RefusalOpeningRangeIncomplete, err)
	}
	if _, err := OpeningRangeATR(barsOf(seriesOf(flat(15, ohlcv{100, 100, 100, 100, 10})...))); refusalCode(t, err) != RefusalATRZero {
		t.Errorf("flat bars: want %s, got %v", RefusalATRZero, err)
	}
}

// ---- RVOL ----

func TestRVOLIsUndefinedInsideTheOpeningRange(t *testing.T) {
	t.Parallel()
	in := barsOf(seriesOf(flat(20, ohlcv{100, 110, 100, 105, 10})...))
	for _, index := range []int{0, 9, 14} {
		if _, err := RVOLPPM(in, index); refusalCode(t, err) != RefusalRVOLUndefined {
			t.Errorf("bar index %d: want %s, got %v", index, RefusalRVOLUndefined, err)
		}
	}
	if got, err := RVOLPPM(in, 15); err != nil || got != 1_000_000 {
		t.Errorf("bar index 15 (first after the range): RVOL = %d, %v; want 1,000,000", got, err)
	}
}

func TestRVOLUsesTheFifteenBarsBeforeAndRoundsDown(t *testing.T) {
	t.Parallel()
	bars := flat(15, ohlcv{100, 110, 100, 105, 3})
	// 직전 15 봉 합 45, 평균 3. 거래량 10 → 10/3 = 3.333… → 3,333,333.
	bars = append(bars, ohlcv{100, 110, 100, 105, 10})
	// 다음 봉의 창은 봉 1..15(거래량 3×14 + 10 = 52). 거래량 13 → 13×15/52 = 3.75 → 3,750,000.
	bars = append(bars, ohlcv{100, 110, 100, 105, 13})
	in := barsOf(seriesOf(bars...))
	if got, err := RVOLPPM(in, 15); err != nil || got != 3_333_333 {
		t.Errorf("RVOL[15] = %d, %v; want 3,333,333", got, err)
	}
	if got, err := RVOLPPM(in, 16); err != nil || got != 3_750_000 {
		t.Errorf("RVOL[16] = %d, %v; want 3,750,000", got, err)
	}
}

func TestRVOLRefusesAZeroVolumeWindowAndSurvivesHugeVolumes(t *testing.T) {
	t.Parallel()
	zero := barsOf(seriesOf(append(flat(15, ohlcv{100, 110, 100, 105, 0}), ohlcv{100, 110, 100, 105, 5})...))
	if _, err := RVOLPPM(zero, 15); refusalCode(t, err) != RefusalDenominatorZero {
		t.Errorf("zero-volume window: want %s, got %v", RefusalDenominatorZero, err)
	}
	// vol × 1e6 × 15 는 uint64 를 넘지만 결과(2×)는 넘지 않는다.
	big := uint64(math.MaxUint64 / 4)
	huge := barsOf(seriesOf(append(flat(15, ohlcv{100, 110, 100, 105, big / 15}), ohlcv{100, 110, 100, 105, 2 * (big / 15)})...))
	if got, err := RVOLPPM(huge, 15); err != nil || got != 2_000_000 {
		t.Errorf("huge volumes: RVOL = %d, %v; want 2,000,000", got, err)
	}
}

func TestVolumeExpandedIsRVOLAtLeastOne(t *testing.T) {
	t.Parallel()
	if VolumeExpanded(999_999) || !VolumeExpanded(1_000_000) || !VolumeExpanded(2_500_000) {
		t.Fatal("VolumeExpanded must be RVOL ≥ 1,000,000 ppm exactly")
	}
}

// ---- 묶음 검사 ----

func TestDeriveRefusesAGapAndASeriesThatDoesNotStartAtTheOpen(t *testing.T) {
	t.Parallel()
	bars := flat(20, ohlcv{100, 110, 100, 105, 10})
	gapped := seriesOf(bars...)
	gapped.Bars = append(gapped.Bars[:10:10], gapped.Bars[11:]...) // 09:10 봉이 빠짐
	if _, err := Derive(gapped, testOpen); refusalCode(t, err) != RefusalMinuteGap {
		t.Errorf("gap: want %s, got %v", RefusalMinuteGap, err)
	} else if !strings.Contains(err.Error(), "09:10") && !strings.Contains(err.Error(), "00:10") {
		t.Errorf("gap refusal %q does not name the missing minute", err)
	}
	late := seriesOf(bars...)
	late.Bars = late.Bars[1:]
	if _, err := Derive(late, testOpen); refusalCode(t, err) != RefusalNotAtOpen {
		t.Errorf("late start: want %s, got %v", RefusalNotAtOpen, err)
	}
}

func TestDeriveRefusesAMixedOrUnclosedSeries(t *testing.T) {
	t.Parallel()
	mutations := map[string]func(*strategyevidence.ClosedBar1mPayload){
		"other symbol":      func(p *strategyevidence.ClosedBar1mPayload) { p.Symbol = "000660" },
		"other session":     func(p *strategyevidence.ClosedBar1mPayload) { p.SessionID = "KRX:2026-08-13" },
		"not closed":        func(p *strategyevidence.ClosedBar1mPayload) { p.Closed = false },
		"not regular":       func(p *strategyevidence.ClosedBar1mPayload) { p.RegularSession = false },
		"other interval":    func(p *strategyevidence.ClosedBar1mPayload) { p.IntervalMS = 300_000 },
		"impossible prices": func(p *strategyevidence.ClosedBar1mPayload) { p.LowMinor = p.HighMinor + 1 },
	}
	for name, mutate := range mutations {
		series := seriesOf(flat(20, ohlcv{100, 110, 100, 105, 10})...)
		mutate(&series.Bars[17].Payload)
		if _, err := Derive(series, testOpen); err == nil {
			t.Errorf("%s: Derive accepted the series", name)
		} else {
			_ = refusalCode(t, err)
		}
	}
}

func TestDeriveIsDeterministicAndTheDigestFollowsTheInput(t *testing.T) {
	t.Parallel()
	bars := append(flat(15, ohlcv{100, 110, 100, 105, 10}), flat(10, ohlcv{105, 118, 104, 117, 30})...)
	a, errA := Derive(seriesOf(bars...), testOpen)
	b, errB := Derive(seriesOf(bars...), testOpen)
	if errA != nil || errB != nil {
		t.Fatalf("Derive: %v / %v", errA, errB)
	}
	if a.Digest == "" || a.Digest != b.Digest || a.Version != Version {
		t.Fatalf("digests %q / %q (version %q); want equal, non-empty, %q", a.Digest, b.Digest, a.Version, Version)
	}
	changed := seriesOf(bars...)
	changed.Bars[20].Payload.Volume++
	c, err := Derive(changed, testOpen)
	if err != nil || c.Digest == a.Digest {
		t.Fatalf("a volume change kept the digest %q (%v)", c.Digest, err)
	}
}

func TestDeriveCarriesNoRVOLForOpeningRangeBars(t *testing.T) {
	t.Parallel()
	bars := append(flat(15, ohlcv{100, 110, 100, 105, 10}), ohlcv{105, 118, 104, 117, 30})
	m, err := Derive(seriesOf(bars...), testOpen)
	if err != nil {
		t.Fatal(err)
	}
	for i, bar := range m.Bars {
		if i < OpeningRangeBars && (bar.RVOLDefined || bar.RVOLPPM != 0 || bar.VolumeExpanded) {
			t.Errorf("opening-range bar %d carries RVOL %+v", i, bar)
		}
		if i >= OpeningRangeBars && !bar.RVOLDefined {
			t.Errorf("bar %d after the range has no RVOL", i)
		}
	}
	if got := m.Bars[15]; got.RVOLPPM != 3_000_000 || !got.VolumeExpanded {
		t.Errorf("bar 15 = %+v; want RVOL 3,000,000 and expanded", got)
	}
}

// ---- breakoutlane 결속 ----

// TestLaneBarsAreAcceptedByTheFrozenEvaluator 는 우리가 만든 봉이 동결 평가기의 구조 검사를 통과하고,
// 평가기의 opening range 가 우리의 OpeningRangeBars 와 같음을 행동으로 못 박음(평가기 상수는 내보내지 않음).
func TestLaneBarsAreAcceptedByTheFrozenEvaluator(t *testing.T) {
	t.Parallel()
	bars := append(flat(15, ohlcv{100, 110, 100, 105, 10}), flat(3, ohlcv{105, 109, 104, 106, 10})...)
	m, err := Derive(seriesOf(bars...), testOpen)
	if err != nil {
		t.Fatal(err)
	}
	lane, err := m.LaneBars()
	if err != nil || len(lane) != len(bars) {
		t.Fatalf("LaneBars = %d bars, %v", len(lane), err)
	}
	snapshot, err := breakoutlane.NewEvidenceSnapshot(testEvidenceInput(t, lane, m.ATRMinor))
	if err != nil {
		t.Fatalf("NewEvidenceSnapshot refused our bars: %v", err)
	}
	if phase := breakoutlane.Evaluate(snapshot, nil).Phase(); phase != "RANGE_LOCKED" {
		t.Fatalf("phase = %s; want RANGE_LOCKED", phase)
	}
	// 평가기는 opening range 미만이면 스냅숏을 거절한다 — 그 수가 우리의 OpeningRangeBars 와 같아야 한다.
	if _, err := breakoutlane.NewEvidenceSnapshot(testEvidenceInput(t, lane[:OpeningRangeBars-1], m.ATRMinor)); err == nil {
		t.Fatal("the evaluator accepted fewer than OpeningRangeBars bars; the two opening ranges disagree")
	}
}

func testEvidenceInput(t *testing.T, bars []breakoutlane.ClosedBar, atr uint64) breakoutlane.EvidenceInput {
	t.Helper()
	cfgIn := breakoutlane.V1ConfigInput{Version: "v1", TickMinor: 1, OpeningRangeMinutes: 15, BreakoutBufferPPM: 100_000,
		RetestTolerancePPM: 250_000, TimeoutKR: 8, TimeoutUS: 10, RVOLMinPPM: 1_500_000, UpperWickRangeMaxPPM: 350_000,
		MaxQuoteAgeMS: 5_000, MaxSpreadPPM: 3_000, MaxEntryDriftPPM: 5_000}
	cfgIn.Digest = breakoutlane.V1ConfigDigest(cfgIn)
	cfg, err := breakoutlane.NewV1Config(cfgIn)
	if err != nil {
		t.Fatal(err)
	}
	q := breakoutlane.QuoteSealInput{BidMinor: 105, AskMinor: 106, LastMinor: 106, SourceObservedAtMS: 1, ReceivedAtMS: 2, Currency: "KRW"}
	q.Digest = breakoutlane.QuoteSealDigest(q)
	quote, err := breakoutlane.NewQuoteSeal(q)
	if err != nil {
		t.Fatal(err)
	}
	f := breakoutlane.FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "KRW", Direction: breakoutlane.FXAccountToInstrument,
		RateNum: 1, RateDen: 1, Scale: 1, AsOfMS: 0, FreshUntilMS: math.MaxInt64}
	f.Digest = breakoutlane.FXSealDigest(f)
	fx, err := breakoutlane.NewFXSeal(f)
	if err != nil {
		t.Fatal(err)
	}
	return breakoutlane.EvidenceInput{Market: breakoutlane.MarketKR, Symbol: "005930", SessionID: testSession, CalendarVersion: "sha256:x",
		LaneID: breakoutlane.KRLaneID, LaneVersion: breakoutlane.LaneVersionV1, Config: cfg, Bars: bars, ATRMinor: atr,
		EvaluatedAtMS: 3, Quote: quote, FX: fx}
}

// TestACorrectedBarKeepsItsLaneIdentity 는 봉 정정(판 2)이 lane 봉 신원 · setupID 를 바꾸지 않고 digest 만 바꿈을 봄.
func TestACorrectedBarKeepsItsLaneIdentity(t *testing.T) {
	t.Parallel()
	bars := append(flat(15, ohlcv{100, 110, 100, 105, 10}), flat(3, ohlcv{105, 109, 104, 106, 10})...)
	original := seriesOf(bars...)
	corrected := seriesOf(bars...)
	corrected.Bars[0].EvidenceID, corrected.Bars[0].RevisionIdentity = "ev-0900-r2", "r2"
	corrected.Bars[0].Payload.Revision = 2
	corrected.Bars[0].Payload.Volume = 11
	a, errA := Derive(original, testOpen)
	b, errB := Derive(corrected, testOpen)
	if errA != nil || errB != nil {
		t.Fatalf("Derive: %v / %v", errA, errB)
	}
	if a.Digest == b.Digest {
		t.Fatal("a corrected bar kept the metrics digest")
	}
	laneA, _ := a.LaneBars()
	laneB, _ := b.LaneBars()
	setupA := breakoutlane.Evaluate(mustSnapshot(t, laneA, a.ATRMinor), nil).SetupID()
	setupB := breakoutlane.Evaluate(mustSnapshot(t, laneB, b.ATRMinor), nil).SetupID()
	if setupA == "" || setupA != setupB {
		t.Fatalf("setup identity changed with a bar revision: %q → %q", setupA, setupB)
	}
	if got := LaneBarID(corrected.Bars[0].Payload); got != testSession+":"+strconv.FormatUint(corrected.Bars[0].Payload.OpenAtMS, 10) {
		t.Fatalf("LaneBarID = %q", got)
	}
}

func mustSnapshot(t *testing.T, bars []breakoutlane.ClosedBar, atr uint64) breakoutlane.EvidenceSnapshot {
	t.Helper()
	snapshot, err := breakoutlane.NewEvidenceSnapshot(testEvidenceInput(t, bars, atr))
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

// ---- 구조 ----

// TestThePackageUsesNoFloatClockOrNetwork 는 spec 「부동소수점 · 시계 · 네트워크 … 를 써서는 안 된다」를 AST 로 고정함.
func TestThePackageUsesNoFloatClockOrNetwork(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	forbiddenImports := map[string]bool{"net": true, "net/http": true, "os": true, "math/rand": true, "math/rand/v2": true}
	seen := 0
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		seen++
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range file.Imports {
			if p := strings.Trim(spec.Path.Value, `"`); forbiddenImports[p] {
				t.Errorf("%s imports %s", path, p)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.Ident:
				if n.Name == "float64" || n.Name == "float32" {
					t.Errorf("%s uses %s", path, n.Name)
				}
			case *ast.SelectorExpr:
				if x, ok := n.X.(*ast.Ident); ok && x.Name == "time" && (n.Sel.Name == "Now" || n.Sel.Name == "Since" || n.Sel.Name == "Sleep") {
					t.Errorf("%s reads the clock via time.%s", path, n.Sel.Name)
				}
			case *ast.BasicLit:
				if n.Kind == token.FLOAT {
					t.Errorf("%s has a float literal %s", path, n.Value)
				}
			}
			return true
		})
	}
	if seen == 0 {
		t.Fatal("no source files were inspected")
	}
}

// barsOf 는 묶음을 순수 입력으로 옮김(시험 도우미 — Derive 의 첫 단계와 같은 변환).
func barsOf(series strategyevidence.BarSeries) []Bar {
	out := make([]Bar, len(series.Bars))
	for i, record := range series.Bars {
		out[i] = barFromPayload(record.Payload)
	}
	return out
}
