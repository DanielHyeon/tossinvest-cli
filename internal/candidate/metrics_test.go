package candidate

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// openStoreOn is openStore with a clock a test can move. Section 3 measures
// elapsed time, so every test in this file manufactures its own.
func openStoreOn(t *testing.T, clk clock.Clock) *Store {
	t.Helper()
	s, err := Open(context.Background(), Options{
		Path:     filepath.Join(t.TempDir(), "candidates.db"),
		FSProber: FixedFSProber(FSInfo{Name: "ext4"}),
		Clock:    clk,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// valueObs is one reading of a cumulative trading value, ranked so the store's
// own validation is satisfied.
func valueObs(symbol string, at time.Time, source SourceID, value string) Observation {
	return Observation{
		Market: MarketKR, Symbol: symbol, Source: source, ObservedAt: at,
		Reported: Reported{Rank: 1, RankTotal: 100, TradingValue: value},
	}
}

// rankObs is one ranked reading with no cumulative figure on it.
func rankObs(market, symbol string, at time.Time, source SourceID,
	rank, total int, newly bool) Observation {
	return Observation{
		Market: market, Symbol: symbol, Source: source, ObservedAt: at,
		Reported: Reported{Rank: rank, RankTotal: total, NewlyListed: newly},
	}
}

// seriesOf builds a one-source series from readings a test wrote by hand.
func seriesOf(t *testing.T, in ...Observation) SourceSeries {
	t.Helper()
	s, err := NewSourceSeries(in)
	if err != nil {
		t.Fatalf("NewSourceSeries: %v", err)
	}
	return s
}

// --- task 3.1: the clock is injected, and it is the only one -------------------

// TestTheElapsedTimeComesFromTheClockATestCanMove is D5 as something that fails.
//
// Every metric below is a quantity divided by an elapsed time, so a test that
// cannot manufacture elapsed time cannot test any of them — it would have to wait
// ten real minutes to check a ten-minute gap, which means nobody ever checks it.
// The store already owns the injected clock; this pins that the metrics path
// takes its instants from there and from caller arguments, never from the wall.
func TestTheElapsedTimeComesFromTheClockATestCanMove(t *testing.T) {
	fake := clock.NewFake(t0)
	s := openStoreOn(t, fake)
	ctx := context.Background()

	if err := s.RecordObservations(ctx, []Observation{
		valueObs("005930", s.Now(), SourceOfficialTradingValue, "0"),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	// Ten minutes, manufactured.
	fake.Advance(10 * time.Minute)
	if err := s.RecordObservations(ctx, []Observation{
		valueObs("005930", s.Now(), SourceOfficialTradingValue, "600"),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}

	series, err := s.SourceSeries(ctx, MarketKR, "005930", SourceOfficialTradingValue, time.Time{})
	if err != nil {
		t.Fatalf("SourceSeries: %v", err)
	}
	window, reason := Rate(series, FieldTradingValue, s.Now(), DefaultAccelerationWindow)
	if reason != "" {
		t.Fatalf("Rate over a ten-minute gap was not computed: %s", reason)
	}
	if window.Seconds != "600" {
		t.Errorf("elapsed = %q, want %q — the metrics did not see the ten minutes the clock made",
			window.Seconds, "600")
	}
	if window.Rate != "1" {
		t.Errorf("rate = %q, want %q (600 over 600 seconds)", window.Rate, "1")
	}
}

// TestNothingInThisPackageAsksTheWallClockWhatTimeItIs keeps the injected clock
// from being one of two.
//
// A single time.Now() anywhere in this package makes the time axis untestable at
// that point, and it does it silently: the arithmetic still runs, the numbers
// still look like numbers, and a test that advances a fake clock by ten minutes
// measures nothing. Cheap to state, impossible to notice in review.
//
// It parses rather than greps. The first version searched for the text and failed
// on candidate.go and store.go — both of which say "rather than from time.Now()"
// in a comment explaining that they do not. A rule enforced by substring match
// fires on the documentation of itself, and the fix for that is to delete the
// sentence, which is the wrong outcome.
//
// The detector is wallClockCalls, which resolves the import by path and matches
// every name in wallClockNames rather than the identifier `time` and the single
// name `Now` — the §3 review measured that `time.Since`, `time.Until`,
// `time.After`, `time.Tick`, `time.NewTimer`, `time.NewTicker`, `time.AfterFunc`,
// `time.Sleep` and `import gotime "time"` all passed the original.
func TestNothingInThisPackageAsksTheWallClockWhatTimeItIs(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing the package: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("parsed no non-test files; the check would pass vacuously")
	}
	files := 0
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			files++
			for _, call := range wallClockCalls(fset, file) {
				t.Errorf("%s:%s reaches the wall clock; instants come from Store.Now() or "+
					"from a caller's argument and waits come from the injected clock, or the "+
					"time axis this package measures cannot be driven by a test",
					filepath.Base(name), call)
			}
		}
	}
	if files == 0 {
		t.Fatal("inspected no files")
	}
}

// --- task 3.2: a rate is per second, not per observation -----------------------

// TestAOneMinuteGapAndATenMinuteGapAreNotTheSameRate is the spec scenario
// "불규칙한 간격에서도 가속도가 비교 가능하다".
//
// Scan spacing is irregular by design — 429 backoff, an operator starting and
// stopping the watch, the engine yield in Schedule.Every — so a delta per
// observation is a delta per unknown amount of time. Two symbols that took in the
// same money, one over a minute and one over ten, are not doing the same thing,
// and a metric that says they are makes the densely sampled stretches look calm.
func TestAOneMinuteGapAndATenMinuteGapAreNotTheSameRate(t *testing.T) {
	fake := clock.NewFake(t0)
	s := openStoreOn(t, fake)
	ctx := context.Background()

	record := func(symbol, value string) {
		t.Helper()
		if err := s.RecordObservations(ctx, []Observation{
			valueObs(symbol, s.Now(), SourceOfficialTradingValue, value),
		}); err != nil {
			t.Fatalf("RecordObservations: %v", err)
		}
	}
	record("FAST", "0")
	record("SLOW", "0")
	fake.Advance(time.Minute)
	record("FAST", "600")
	fake.Advance(9 * time.Minute)
	record("SLOW", "600")

	rateOf := func(symbol string, at time.Time) Window {
		t.Helper()
		series, err := s.SourceSeries(ctx, MarketKR, symbol, SourceOfficialTradingValue, time.Time{})
		if err != nil {
			t.Fatalf("SourceSeries(%s): %v", symbol, err)
		}
		w, reason := Rate(series, FieldTradingValue, at, DefaultAccelerationWindow)
		if reason != "" {
			t.Fatalf("Rate(%s) was not computed: %s", symbol, reason)
		}
		return w
	}
	fast := rateOf("FAST", t0.Add(time.Minute))
	slow := rateOf("SLOW", t0.Add(10*time.Minute))

	if fast.Delta != slow.Delta {
		t.Fatalf("the two symbols took in %q and %q; the test wants the same absolute increase",
			fast.Delta, slow.Delta)
	}
	if fast.Rate == slow.Rate {
		t.Fatalf("both symbols report a rate of %q — a per-observation delta calls a "+
			"one-minute move and a ten-minute move the same thing", fast.Rate)
	}
	if fast.Rate != "10" || slow.Rate != "1" {
		t.Errorf("rates = fast %q slow %q, want %q and %q (600 over 60s and over 600s)",
			fast.Rate, slow.Rate, "10", "1")
	}
	if fast.Seconds != "60" || slow.Seconds != "600" {
		t.Errorf("elapsed = fast %q slow %q, want %q and %q", fast.Seconds, slow.Seconds, "60", "600")
	}
}

// --- task 3.3: acceleration is a ratio of rates, and both lengths are recorded --

// TestAWindowStretchedByBackoffIsNotAcceleration is the spec scenario "늘어난
// 창이 가속도를 부풀리지 않는다", and D9's whole argument.
//
// The draft definition was recent-window value ÷ prior-window value. The same
// contract sets the 429 backoff at 30→60→120→300 seconds, so the first retreat
// turns a 30-second window into a 90-second one and a ratio of sums triples on
// its own. The 1.8 threshold would then be reacting to our polling accident
// rather than to the market — and backoff is not an edge case here, it is a thing
// the design guarantees will happen.
func TestAWindowStretchedByBackoffIsNotAcceleration(t *testing.T) {
	series := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "0"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "100"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "200"),
		// The 429 lands here: the next reading arrives 90 seconds later, and three
		// windows' worth of money arrives with it.
		valueObs("005930", t0.Add(150*time.Second), SourceOfficialTradingValue, "500"),
	)

	got := Accelerate(series, FieldTradingValue, t0.Add(150*time.Second), DefaultAccelerationWindow)
	if !got.Computed() {
		t.Fatalf("acceleration was not computed: %s", got.Reason)
	}
	if got.Ratio != "1" {
		t.Errorf("acceleration = %q, want %q — a ratio of sums reads a stretched window as a "+
			"faster market", got.Ratio, "1")
	}
	// D9: 두 창의 실제 길이를 함께 기록한다.
	if got.Recent.Seconds != "90" {
		t.Errorf("recent window = %q seconds, want %q — the backoff is invisible in the record",
			got.Recent.Seconds, "90")
	}
	if got.Prior.Seconds != "30" {
		t.Errorf("prior window = %q seconds, want %q", got.Prior.Seconds, "30")
	}
	if got.Crossed("1.3") {
		t.Error("a market that did not accelerate crossed the 1.3 shadow threshold")
	}
}

// TestTheSameRatioFromTwoWindowLengthsIsTwoDifferentFacts is the sentence after
// D9's formula: 30/30초에서 나온 1.8배와 30/90초에서 나온 1.8배를 읽는 사람이
// 구분할 수 있어야 한다.
//
// Normalising by elapsed time makes the two ratios comparable; it does not make
// them equally trustworthy. A 1.8 measured against a 90-second denominator has
// three times the averaging in it, and a reader deciding whether to believe the
// number needs to see that without re-deriving it from timestamps.
func TestTheSameRatioFromTwoWindowLengthsIsTwoDifferentFacts(t *testing.T) {
	tight := seriesOf(t,
		valueObs("AAA", t0, SourceOfficialTradingValue, "0"),
		valueObs("AAA", t0.Add(30*time.Second), SourceOfficialTradingValue, "100"),
		valueObs("AAA", t0.Add(60*time.Second), SourceOfficialTradingValue, "280"),
	)
	stretched := seriesOf(t,
		valueObs("BBB", t0, SourceOfficialTradingValue, "0"),
		valueObs("BBB", t0.Add(90*time.Second), SourceOfficialTradingValue, "300"),
		valueObs("BBB", t0.Add(120*time.Second), SourceOfficialTradingValue, "480"),
	)

	a := Accelerate(tight, FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	b := Accelerate(stretched, FieldTradingValue, t0.Add(120*time.Second), DefaultAccelerationWindow)

	if a.Ratio != "1.8" || b.Ratio != "1.8" {
		t.Fatalf("ratios = %q and %q, want both %q — the test needs them equal to be about "+
			"anything else", a.Ratio, b.Ratio, "1.8")
	}
	if a.Prior.Seconds != "30" {
		t.Errorf("tight prior window = %q seconds, want %q", a.Prior.Seconds, "30")
	}
	if b.Prior.Seconds != "90" {
		t.Errorf("stretched prior window = %q seconds, want %q — two identical ratios are "+
			"indistinguishable in the record", b.Prior.Seconds, "90")
	}
}

// --- task 3.4: no prior window means no acceleration ----------------------------

// TestTheFirstWindowHasNoPriorToAccelerateAgainst is the spec scenario "첫 관측은
// 가속도를 만들지 않는다".
//
// A missing denominator defaulted to zero produces +Inf, and +Inf clears every
// threshold it is ever compared against. That is a maximal early signal
// manufactured out of having just started — precisely the fake earliness D3
// exists to refuse. Not computing something is not the same as passing it.
func TestTheFirstWindowHasNoPriorToAccelerateAgainst(t *testing.T) {
	series := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1600"),
	)

	got := Accelerate(series, FieldTradingValue, t0.Add(30*time.Second), DefaultAccelerationWindow)
	if got.Computed() {
		t.Fatalf("acceleration = %q on a series with no prior window; there was nothing to "+
			"divide by", got.Ratio)
	}
	if got.Reason != NotComputedWarmingUp {
		t.Errorf("reason = %q, want %q", got.Reason, NotComputedWarmingUp)
	}
	if got.Ratio != "" {
		t.Errorf("ratio = %q, want empty — an uncomputed value must not have a spelling",
			got.Ratio)
	}
	for _, c := range got.Crossings {
		if c.Crossed {
			t.Errorf("a warming-up candidate crossed the %s shadow threshold", c.Threshold)
		}
	}
}

// --- task 3.4b: an unusable denominator is not a pass either ---------------------

// TestAnUnusableDenominatorIsNotAPassAndSaysWhichOne is D9's 2026-07-28
// correction: WARMING_UP means "there is no prior window", and there are three
// further ways to have one and not be able to divide by it. All three produce
// +Inf or a sign flip, and all three clear every shadow threshold.
//
// The reasons are recorded apart from each other, and apart from WARMING_UP, for
// the reason D3 gives for splitting seen_late from extended: the remedies differ.
// Many halted denominators means the source panel is wrong for this market; many
// reversals means a session boundary is being read as intraday; many warming-ups
// means the scan simply started, and nothing needs fixing at all.
func TestAnUnusableDenominatorIsNotAPassAndSaysWhichOne(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values []string
		want   NotComputed
		why    string
	}{
		{
			name:   "the prior window had no trades at all",
			values: []string{"1000", "1000", "1500"},
			want:   NotComputedPriorRateZero,
			why:    "a halted or illiquid symbol divides by zero",
		},
		{
			name:   "the cumulative figure went backwards",
			values: []string{"2000", "1900", "1000"},
			want:   NotComputedReversed,
			why: "trading value is an intraday cumulative, so a session boundary or a source " +
				"restart walks it back — and a ratio of two negatives is positive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			series := seriesOf(t,
				valueObs("005930", t0, SourceOfficialTradingValue, tc.values[0]),
				valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, tc.values[1]),
				valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, tc.values[2]),
			)
			got := Accelerate(series, FieldTradingValue, t0.Add(60*time.Second),
				DefaultAccelerationWindow)

			if got.Computed() {
				t.Fatalf("acceleration = %q; %s", got.Ratio, tc.why)
			}
			if got.Reason != tc.want {
				t.Errorf("reason = %q, want %q", got.Reason, tc.want)
			}
			if got.Reason == NotComputedWarmingUp {
				t.Error("recorded as warming up; that reason means the scan just started and " +
					"nothing needs fixing, which is the opposite of this case")
			}
			for _, c := range got.Crossings {
				if c.Crossed {
					t.Errorf("crossed the %s shadow threshold on an unusable denominator",
						c.Threshold)
				}
			}
		})
	}
}

// TestAWindowThatTookNoTimeHasNoRate is D9's third unusable denominator, at the
// level where it is still reachable.
//
// Two readings carrying the same instant divide a real delta by zero seconds. The
// per-source series of task 3.4c removes the cause D9 named — two sources
// reporting one scan's instant — but the guard belongs on the measurement itself,
// because that is the primitive every window goes through and a duplicated row or
// a replayed scan produces the same pair.
func TestAWindowThatTookNoTimeHasNoRate(t *testing.T) {
	from := valueObs("005930", t0, SourceOfficialTradingValue, "1000")
	to := valueObs("005930", t0, SourceOfficialTradingValue, "1200")

	got, reason := MeasureWindow(from, to, FieldTradingValue)
	if reason != NotComputedZeroElapsed {
		t.Errorf("reason = %q, want %q — a rate of %q per second was manufactured out of "+
			"no elapsed time", reason, NotComputedZeroElapsed, got.Rate)
	}
	if reason == NotComputedWarmingUp {
		t.Error("recorded as warming up; there is a prior reading, it is just unusable")
	}
	if got.Rate != "" {
		t.Errorf("rate = %q, want empty", got.Rate)
	}
}

// TestADuplicatedRowInAShortSeriesIsWarmingUpAndSaysSo was
// TestAPriorWindowThatTookNoTimeCrossesNothing, which was aimed at the wrong
// thing and passed for the wrong reason (§3 review, P1-1).
//
// It intended to reach ZERO_ELAPSED_SECONDS through Accelerate, and cannot:
// Accelerate anchors `mid` at or before end − window and `start` at or before
// mid − window, so with a positive window both windows are at least `window`
// long and neither elapsed guard can fire. What this series actually produces is
// WARMING_UP with both windows empty, and the old assertions — !Computed() and no
// crossings — are satisfied by WARMING_UP, so nothing was being checked.
//
// The reason it actually gets is now asserted. The guard it was aimed at lives on
// MeasureWindow, where TestAWindowThatTookNoTimeHasNoRate reaches it, and the real
// hazard a duplicated row carries is
// TestADuplicatedRowDoesNotSilentlyChangeTheAcceleration.
func TestADuplicatedRowInAShortSeriesIsWarmingUpAndSaysSo(t *testing.T) {
	series := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		// Same instant, same source: a replayed scan, or a body listing the symbol
		// twice.
		valueObs("005930", t0, SourceOfficialTradingValue, "1200"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1500"),
	)

	got := Accelerate(series, FieldTradingValue, t0.Add(30*time.Second), DefaultAccelerationWindow)
	if got.Computed() {
		t.Fatalf("acceleration = %q from a series holding one window of history", got.Ratio)
	}
	if got.Reason != NotComputedWarmingUp {
		t.Errorf("reason = %q, want %q — a duplicate does not shorten the history, it "+
			"just does not lengthen it", got.Reason, NotComputedWarmingUp)
	}
	for _, c := range got.Crossings {
		if c.Crossed {
			t.Errorf("crossed the %s shadow threshold", c.Threshold)
		}
	}
}

// --- task 3.4c: the series is one source's, and cannot be otherwise -------------

// TestASeriesCannotBeDifferencedAcrossTwoSources is D9's "시계열은 원천마다
// 따로다", and the paragraph after it explaining why it had to be written down:
// the natural implementation is the wrong one, it compiles, and it produces
// numbers that look right.
//
// TradingValue is a different cumulative per source — different unit, different
// aggregation window, different fallback status. Differencing one source's
// reading against another's measures the gap between the two sources, and because
// the two track each other loosely the result is plausible in both magnitude and
// sign. Here the interleaved answer is a flat 1.0 while the official source's own
// series is crossing the lowest shadow threshold: the mixing did not produce
// nonsense, it produced calm.
func TestASeriesCannotBeDifferencedAcrossTwoSources(t *testing.T) {
	s := openStoreOn(t, clock.NewFake(t0))
	ctx := context.Background()

	// Two cumulative trading values for one symbol, tracking each other loosely
	// and sampled out of phase, exactly as two panels of a real scan would.
	official := []Observation{
		valueObs("005930", t0, SourceOfficialTradingValue, "10000000000"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "10600000000"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "11400000000"),
	}
	wts := []Observation{
		valueObs("005930", t0.Add(15*time.Second), SourceWTSPopular, "10300000000"),
		valueObs("005930", t0.Add(45*time.Second), SourceWTSPopular, "11000000000"),
	}
	if err := s.RecordObservations(ctx, append(append([]Observation{}, official...), wts...)); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}

	// A series built from every source's rows is refused rather than averaged.
	if _, err := NewSourceSeries(append(append([]Observation{}, official...), wts...)); err == nil {
		t.Error("a series accepted rows from two sources; differencing them measures the " +
			"difference between the sources, not the market")
	}
	// And so is a single window across the pair, which is the same mistake reached
	// without going through a series at all.
	if _, reason := MeasureWindow(official[1], wts[1], FieldTradingValue); reason !=
		NotComputedSourceMismatch {
		t.Errorf("a window measured from %s to %s = %q, want %q",
			official[1].Source, wts[1].Source, reason, NotComputedSourceMismatch)
	}

	series, err := s.SourceSeries(ctx, MarketKR, "005930", SourceOfficialTradingValue, time.Time{})
	if err != nil {
		t.Fatalf("SourceSeries: %v", err)
	}
	if series.Len() != len(official) {
		t.Fatalf("the official series carries %d readings, want %d — it picked up another "+
			"source's rows", series.Len(), len(official))
	}

	got := Accelerate(series, FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	if !got.Computed() {
		t.Fatalf("acceleration was not computed: %s", got.Reason)
	}
	// 0.8e9 over 30s against 0.6e9 over 30s. Interleaving the two sources instead
	// gives a flat 1.0 — every alternating step is +0.4e9 over 15s — so the mixing
	// does not produce nonsense, it produces calm.
	if got.Ratio != "1.333333333333" {
		t.Errorf("acceleration = %q, want %q", got.Ratio, "1.333333333333")
	}
	if !got.Crossed("1.3") {
		t.Errorf("acceleration = %q and did not cross 1.3; the official source's own series "+
			"is accelerating and the interleaved reading reports calm", got.Ratio)
	}
}

// --- task 3.5: the five shadow thresholds are recorded, and judged by nothing ---

// TestEveryShadowThresholdIsRecordedAndTheBoundaryIsExact.
//
// This change records and does not decide (design.md "결정된 계약값": 이 change는
// 이 값들로 판정하지 않는다). Recording all five is what lets T3.2 derive the lane
// thresholds from data instead of asserting them, so a partial record is a
// decision taken early by omission.
//
// The boundary is `>=` and it is pinned on exact decimals rather than floats,
// because a threshold that moves with the last bit of a float64 is a threshold
// nobody can reproduce from the record.
func TestEveryShadowThresholdIsRecordedAndTheBoundaryIsExact(t *testing.T) {
	at := func(recent string) Acceleration {
		series := seriesOf(t,
			valueObs("005930", t0, SourceOfficialTradingValue, "0"),
			valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "100"),
			valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, recent),
		)
		return Accelerate(series, FieldTradingValue, t0.Add(60*time.Second),
			DefaultAccelerationWindow)
	}

	exactly := at("280") // +180 against +100 — exactly 1.8
	if exactly.Ratio != "1.8" {
		t.Fatalf("ratio = %q, want %q", exactly.Ratio, "1.8")
	}
	if len(exactly.Crossings) != len(ShadowThresholds) {
		t.Fatalf("recorded %d thresholds, want all %d",
			len(exactly.Crossings), len(ShadowThresholds))
	}
	want := map[string]bool{"1.3": true, "1.5": true, "1.8": true, "2.0": false, "2.5": false}
	for _, c := range exactly.Crossings {
		if c.Crossed != want[c.Threshold] {
			t.Errorf("%s crossed = %v, want %v", c.Threshold, c.Crossed, want[c.Threshold])
		}
	}

	just := at("279.99") // +179.99 against +100 — 1.7999
	if just.Ratio != "1.7999" {
		t.Fatalf("ratio = %q, want %q", just.Ratio, "1.7999")
	}
	if just.Crossed("1.8") {
		t.Error("1.7999 crossed the 1.8 threshold")
	}
	if !just.Crossed("1.5") {
		t.Error("1.7999 did not cross the 1.5 threshold")
	}

	// And a ratio that is exactly a threshold has to reach it. `>=` is only a
	// decidable rule if the arithmetic under it is exact: +1.3 against +1.0 in
	// float64 is 1.2999999999999998, which misses its own threshold — a crossing
	// dropped from the shadow record by the last bit of a double. The lane
	// thresholds in T3.2 are supposed to be derived from these counts.
	series := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "0"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "2.3"),
	)
	exact := Accelerate(series, FieldTradingValue, t0.Add(60*time.Second),
		DefaultAccelerationWindow)
	if exact.Ratio != "1.3" {
		t.Fatalf("ratio = %q, want %q", exact.Ratio, "1.3")
	}
	if !exact.Crossed("1.3") {
		t.Error("a ratio of exactly 1.3 did not cross the 1.3 threshold")
	}
	if exact.Crossed("1.5") {
		t.Error("1.3 crossed the 1.5 threshold")
	}
}

// TestANotComputedAccelerationCrossesNothingAndIsCountedApart is the counting
// half of task 3.5, and the same shape as D10.
//
// A candidate whose acceleration could not be computed has not passed anything.
// If it is tallied as a crossing the shadow record fills with maximal signals
// produced by missing denominators, and the whole point of recording five
// thresholds — deriving lane thresholds from real data later — is gone. It is
// counted under its reason instead, which is also the number that says whether
// the source panel needs fixing.
func TestANotComputedAccelerationCrossesNothingAndIsCountedApart(t *testing.T) {
	warming := Accelerate(seriesOf(t,
		valueObs("AAA", t0, SourceOfficialTradingValue, "1000"),
		valueObs("AAA", t0.Add(30*time.Second), SourceOfficialTradingValue, "1600"),
	), FieldTradingValue, t0.Add(30*time.Second), DefaultAccelerationWindow)

	halted := Accelerate(seriesOf(t,
		valueObs("BBB", t0, SourceOfficialTradingValue, "1000"),
		valueObs("BBB", t0.Add(30*time.Second), SourceOfficialTradingValue, "1000"),
		valueObs("BBB", t0.Add(60*time.Second), SourceOfficialTradingValue, "1500"),
	), FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)

	real := Accelerate(seriesOf(t,
		valueObs("CCC", t0, SourceOfficialTradingValue, "0"),
		valueObs("CCC", t0.Add(30*time.Second), SourceOfficialTradingValue, "100"),
		valueObs("CCC", t0.Add(60*time.Second), SourceOfficialTradingValue, "280"),
	), FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)

	tally := TallyCrossings([]Acceleration{warming, halted, real})
	for _, th := range []string{"1.3", "1.5", "1.8"} {
		if tally.Crossed[th] != 1 {
			t.Errorf("%s crossed by %d candidates, want 1 — only one of the three was measured",
				th, tally.Crossed[th])
		}
	}
	for _, th := range []string{"2.0", "2.5"} {
		if tally.Crossed[th] != 0 {
			t.Errorf("%s crossed by %d candidates, want 0", th, tally.Crossed[th])
		}
	}
	if tally.NotComputed[NotComputedWarmingUp] != 1 {
		t.Errorf("warming up counted %d times, want 1", tally.NotComputed[NotComputedWarmingUp])
	}
	if tally.NotComputed[NotComputedPriorRateZero] != 1 {
		t.Errorf("halted denominator counted %d times, want 1",
			tally.NotComputed[NotComputedPriorRateZero])
	}
}

// --- task 3.8: rank percentile, and the fact that is not a rank -----------------

// TestANewlyListedSymbolDoesNotClimbFromLastPlace is D8 correction 1.
//
// A symbol appearing in a list for the first time has no previous percentile.
// Filling that gap with "previously last" is the natural move and it converts
// having no evidence into having the strongest evidence available: the recorded
// gain becomes the symbol's own percentile, so anything entering above the bottom
// fifth clears the 20%p contract threshold on its first sighting.
//
// The list boundary is what makes this constant rather than occasional — the rows
// around 140–150 of 150 churn in and out on every scan — and a symbol whose volume
// has just spiked does not enter at the boundary, it enters mid-list. So the path
// is taken every scan and the manufactured gain is routinely maximal.
//
// The fix is not a better fill. It is to record newly_listed as its own fact and
// leave the gain absent, which is the same rule as every other absent value in
// this package: 값을 지어내지 않고 없음을 없음으로 남긴다.
func TestANewlyListedSymbolDoesNotClimbFromLastPlace(t *testing.T) {
	for _, tc := range []struct {
		name string
		rank int
	}{
		{"entering mid-list, where a spiking symbol enters", 45},
		{"entering at the boundary, which churns every scan", 148},
	} {
		t.Run(tc.name, func(t *testing.T) {
			series := seriesOf(t,
				rankObs(MarketKR, "005930", t0.Add(30*time.Second),
					SourceOfficialTradingValue, tc.rank, 150, true),
			)
			got := RankChange(series, t0.Add(30*time.Second), DefaultAccelerationWindow)

			if got.Computed() {
				t.Fatalf("percentile gain = %q for a symbol with no previous reading; it was "+
					"credited with a climb from last place", got.PercentileGain)
			}
			if got.Reason != NotComputedNoPriorRank {
				t.Errorf("reason = %q, want %q", got.Reason, NotComputedNoPriorRank)
			}
			if got.PercentileGain != "" {
				t.Errorf("gain = %q, want empty — absent is not zero and it is not a climb",
					got.PercentileGain)
			}
			// The separate fact survives, because it is the one that says whether we
			// are arriving late: a symbol that first appears at 12th and one that
			// first appears at 148th are different events (D8).
			if !got.NewlyListed {
				t.Error("newly_listed was dropped; it is the fact that replaces the invented gain")
			}
			if got.Rank != tc.rank || got.RankTotal != 150 {
				t.Errorf("rank = %d of %d, want %d of 150", got.Rank, got.RankTotal, tc.rank)
			}
		})
	}
}

// TestTheSameNumberOfPlacesIsADifferentMoveInADifferentList is why the move is a
// percentile and not a count of positions.
//
// The KR panel returns 150 rows and the US panel 100, so "up thirty places" is
// two different distances (D8). A raw position delta compared against one
// threshold silently applies a stricter bar to the longer list.
func TestTheSameNumberOfPlacesIsADifferentMoveInADifferentList(t *testing.T) {
	kr := seriesOf(t,
		rankObs(MarketKR, "005930", t0, SourceOfficialTradingValue, 70, 150, false),
		rankObs(MarketKR, "005930", t0.Add(30*time.Second),
			SourceOfficialTradingValue, 40, 150, false),
	)
	us := seriesOf(t,
		rankObs(MarketUS, "AAPL", t0, SourceOfficialTradingValue, 70, 100, false),
		rankObs(MarketUS, "AAPL", t0.Add(30*time.Second),
			SourceOfficialTradingValue, 40, 100, false),
	)

	krMove := RankChange(kr, t0.Add(30*time.Second), DefaultAccelerationWindow)
	usMove := RankChange(us, t0.Add(30*time.Second), DefaultAccelerationWindow)
	if !krMove.Computed() || !usMove.Computed() {
		t.Fatalf("rank moves were not computed: %q / %q", krMove.Reason, usMove.Reason)
	}
	if krMove.PercentileGain != "20" {
		t.Errorf("KR gain = %q, want %q (thirty places of a hundred and fifty)",
			krMove.PercentileGain, "20")
	}
	if usMove.PercentileGain != "30" {
		t.Errorf("US gain = %q, want %q (thirty places of a hundred)",
			usMove.PercentileGain, "30")
	}
	if krMove.PercentileGain == usMove.PercentileGain {
		t.Error("thirty places is the same move in a 150-row list and a 100-row one; " +
			"the gain is not normalised by the list length")
	}
}

// TestAChurningSymbolIsMeasuredAgainstWhatWeActuallySaw is the other half of the
// newly_listed rule: the flag suppresses nothing.
//
// A symbol that dropped off the source's own previous reading and came back comes
// with newly_listed set, but we still hold a reading of our own to measure
// against. Refusing the gain there would throw away a real measurement in the name
// of the rule that exists to stop inventing one, and the boundary rows this
// affects are exactly the rows the record needs to be honest about.
func TestAChurningSymbolIsMeasuredAgainstWhatWeActuallySaw(t *testing.T) {
	series := seriesOf(t,
		rankObs(MarketKR, "005930", t0, SourceOfficialTradingValue, 149, 150, false),
		// Gone from the source's previous reading, back now, two places higher.
		rankObs(MarketKR, "005930", t0.Add(60*time.Second),
			SourceOfficialTradingValue, 147, 150, true),
	)

	got := RankChange(series, t0.Add(60*time.Second), DefaultAccelerationWindow)
	if !got.Computed() {
		t.Fatalf("a symbol with a stored previous reading was not measured: %s", got.Reason)
	}
	if !got.NewlyListed {
		t.Error("newly_listed was dropped; the record has to show the series has a hole in it")
	}
	// Two places of a hundred and fifty. Not the 2 that "previously last" would
	// have credited it with, and nowhere near a threshold either way.
	if got.PercentileGain != "1.333333333333" {
		t.Errorf("gain = %q, want %q (two places of a hundred and fifty)",
			got.PercentileGain, "1.333333333333")
	}
}

// --- §3 review, P0-1: an unassigned metric is not a measured one ----------------

// TestTheZeroValueOfEveryMetricInThisFileIsUnmeasured is metrics.go's twin of
// level_test.go's TestTheZeroValueOfEveryLevelMetricIsUnmeasured, and it is here
// because this file inverted the rule its own header states: Computed() was
// Reason == "", so an Acceleration nobody assigned reported itself as measured,
// with no ratio and no crossings — "we measured it and it crossed nothing".
//
// It covers every exported type in the file rather than the two that were wrong,
// because the next one added is the one that gets it wrong next.
func TestTheZeroValueOfEveryMetricInThisFileIsUnmeasured(t *testing.T) {
	var a Acceleration
	if a.Computed() {
		t.Errorf("the zero Acceleration reports itself as computed with ratio %q and %d "+
			"crossings; a map miss or an unfilled slot would read as a candidate we "+
			"measured and found to have crossed nothing", a.Ratio, len(a.Crossings))
	}
	if a.Why() != NotComputedUnset {
		t.Errorf("the zero Acceleration's reason = %q, want %q — the unset bucket needs a "+
			"name or it cannot be counted", a.Why(), NotComputedUnset)
	}
	for _, th := range ShadowThresholds {
		if a.Crossed(th) {
			t.Errorf("the zero Acceleration crossed %s", th)
		}
	}

	var m RankMove
	if m.Computed() {
		t.Errorf("the zero RankMove reports itself as computed with gain %q", m.PercentileGain)
	}
	if m.Why() != NotComputedUnset {
		t.Errorf("the zero RankMove's reason = %q, want %q", m.Why(), NotComputedUnset)
	}

	var w Window
	if w.Measured() {
		t.Error("the zero Window reports a rate")
	}
	if w.Seconds != "" || w.Delta != "" || w.Rate != "" {
		t.Errorf("the zero Window carries seconds %q delta %q rate %q; absent is spelled "+
			"empty here, never \"0\"", w.Seconds, w.Delta, w.Rate)
	}

	var c ThresholdCrossing
	if c.Crossed {
		t.Error("the zero ThresholdCrossing reports a crossing")
	}

	var s SourceSeries
	if s.Len() != 0 {
		t.Errorf("the zero SourceSeries holds %d readings", s.Len())
	}
	if s.Unusable() != NotComputedNoObservations {
		t.Errorf("the zero SourceSeries reports %q, want %q", s.Unusable(),
			NotComputedNoObservations)
	}

	var tally CrossingTally
	if tally.Total != 0 || tally.Computed != 0 {
		t.Errorf("the zero CrossingTally counts %d candidates and %d computed",
			tally.Total, tally.Computed)
	}
}

// TestTheThreeNaturalWaysToProduceAnUnassignedMetricAreAllUnmeasured is the same
// rule at the call sites §4 and §5 are about to spell it at. Each of these is the
// obvious way to write the thing, and each produced a measured-looking value.
func TestTheThreeNaturalWaysToProduceAnUnassignedMetricAreAllUnmeasured(t *testing.T) {
	// A map miss: the candidate was not in the hot queue, so nothing measured it.
	byKey := map[Key]Acceleration{}
	if byKey[Key{Market: MarketKR, Symbol: "005930"}].Computed() {
		t.Error("a map miss reads as a measured acceleration")
	}

	// A slice sized to the candidate list and filled only where a read succeeded.
	slots := make([]Acceleration, 3)
	slots[0] = Accelerate(seriesOf(t,
		valueObs("AAA", t0, SourceOfficialTradingValue, "0"),
		valueObs("AAA", t0.Add(30*time.Second), SourceOfficialTradingValue, "100"),
		valueObs("AAA", t0.Add(60*time.Second), SourceOfficialTradingValue, "280"),
	), FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	for i, a := range slots[1:] {
		if a.Computed() {
			t.Errorf("slot %d was never filled and reads as measured", i+1)
		}
	}

	// And the tally over it accounts for the two nobody filled.
	tally := TallyCrossings(slots)
	if tally.NotComputed[NotComputedUnset] != 2 {
		t.Errorf("%d unfilled slots counted under %q, want 2",
			tally.NotComputed[NotComputedUnset], NotComputedUnset)
	}
	if tally.Crossed["1.8"] != 1 {
		t.Errorf("1.8 crossed by %d, want 1 — only the filled slot measured anything",
			tally.Crossed["1.8"])
	}
}

// TestACandidateNobodyMeasuredIsAccountedForInTheTally is P0-2: the unset bucket
// falls out of both halves of the tally.
func TestACandidateNobodyMeasuredIsAccountedForInTheTally(t *testing.T) {
	tally := TallyCrossings([]Acceleration{{}})
	crossed := 0
	for _, n := range tally.Crossed {
		crossed += n
	}
	notComputed := 0
	for _, n := range tally.NotComputed {
		notComputed += n
	}
	if crossed+notComputed != 1 {
		t.Errorf("one candidate went in and %d came out (crossed %d, not-computed %d); "+
			"the one nobody measured is in neither half",
			crossed+notComputed, crossed, notComputed)
	}
}

// TestATallyCountsEachCandidateOnceAndOnlyAgainstThresholdsItKnows.
func TestATallyCountsEachCandidateOnceAndOnlyAgainstThresholdsItKnows(t *testing.T) {
	tally := TallyCrossings([]Acceleration{{
		Ratio: "2.0",
		Crossings: []ThresholdCrossing{
			{Threshold: "1.3", Crossed: true},
			{Threshold: "1.3", Crossed: true},
			{Threshold: "9.9", Crossed: true},
		},
	}})
	if tally.Crossed["1.3"] != 1 {
		t.Errorf("1.3 crossed by %d candidates, want 1 — one candidate was counted twice",
			tally.Crossed["1.3"])
	}
	if _, ok := tally.Crossed["9.9"]; ok {
		t.Errorf("the tally grew a %q key; a threshold nobody shadowed is not a column in "+
			"the shadow record", "9.9")
	}
}

// --- §3 review, P1-1: a replayed row must not move the answer -------------------

// TestADuplicatedRowDoesNotSilentlyChangeTheAcceleration.
//
// Two rows of one source sharing an instant is a replayed scan or a body that
// listed the symbol twice. Resolving that tie to the later row skips the earlier
// one, so the prior window is measured from a figure the source published second
// — and the acceleration changes with no reason code and nothing in the record
// saying it happened.
func TestADuplicatedRowDoesNotSilentlyChangeTheAcceleration(t *testing.T) {
	clean := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1500"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "2000"),
	)
	doubled := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		// The replay: same source, same instant, a figure published after the first.
		valueObs("005930", t0, SourceOfficialTradingValue, "1200"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1500"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "2000"),
	)

	want := Accelerate(clean, FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	got := Accelerate(doubled, FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	if want.Ratio != "1" {
		t.Fatalf("the undoubled series accelerated %q, want %q", want.Ratio, "1")
	}
	if got.Ratio != want.Ratio {
		t.Errorf("the doubled series accelerated %q against the same market's %q, and said "+
			"nothing about why (reason %q)", got.Ratio, want.Ratio, got.Reason)
	}
}

// TestANonPositiveWindowIsACallerBugAndNotAMarketCondition.
//
// ZERO_ELAPSED_SECONDS says two readings carried the same instant — a replayed
// scan, a doubled row. A window argument of zero or less is nobody's market: it
// is a caller that computed a duration wrong, and it belongs with SOURCE_MISMATCH
// rather than with the conditions an operator would go looking for in the panel.
func TestANonPositiveWindowIsACallerBugAndNotAMarketCondition(t *testing.T) {
	series := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1500"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "2000"),
	)
	for _, window := range []time.Duration{0, -30 * time.Second} {
		got := Accelerate(series, FieldTradingValue, t0.Add(60*time.Second), window)
		if got.Reason == NotComputedZeroElapsed {
			t.Errorf("window %s reported %q, which describes the data; the window is the "+
				"caller's argument", window, got.Reason)
		}
		if got.Computed() {
			t.Errorf("window %s produced a ratio of %q", window, got.Ratio)
		}
	}
}

// --- §3 review, P1-2: three causes are not one ----------------------------------

// TestASeriesThatWasRefusedCannotComeBackAsWarmingUp.
//
// NewSourceSeries refuses a mixed series and returns the zero value beside the
// error. A caller that drops the error — the natural spelling in a scan loop that
// must keep going — then measures the zero value, and every metric answers
// WARMING_UP: the one reason that means nothing needs fixing.
func TestASeriesThatWasRefusedCannotComeBackAsWarmingUp(t *testing.T) {
	mixed, err := NewSourceSeries([]Observation{
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		valueObs("005930", t0.Add(30*time.Second), SourceWTSPopular, "1500"),
	})
	if err == nil {
		t.Fatal("a mixed series was accepted")
	}
	got := Accelerate(mixed, FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	if got.Reason == NotComputedWarmingUp {
		t.Errorf("a refused series reports %q — a source composition defect is indistinguishable "+
			"from a scan that just started", got.Reason)
	}
	if got.Computed() {
		t.Errorf("a refused series produced a ratio of %q", got.Ratio)
	}
}

// TestASeriesWithNoReadingsAtAllSaysSo is D9's correction on the same axis:
// 전자가 많으면 원천 구성 문제이고 후자가 많으면 그냥 방금 시작한 것. A source that has
// never reported this symbol and a symbol whose second window has not filled yet
// are different facts with different remedies. level.go got this right with
// LevelNoObservations.
func TestASeriesWithNoReadingsAtAllSaysSo(t *testing.T) {
	var empty SourceSeries
	got := Accelerate(empty, FieldTradingValue, t0, DefaultAccelerationWindow)
	if got.Reason == NotComputedWarmingUp {
		t.Errorf("a series with no readings reports %q; nothing was ever observed for it, "+
			"which is a panel question rather than a young scan", got.Reason)
	}
	if _, reason := Rate(empty, FieldTradingValue, t0, DefaultAccelerationWindow); reason ==
		NotComputedWarmingUp {
		t.Errorf("Rate over a series with no readings reports %q", reason)
	}
}

// TestATallyAccountsForEveryCandidateItWasGiven is the identity that makes the
// disjointness of the two halves checkable instead of asserted.
//
// Total == Computed + Σ NotComputed. The shape it refuses is the one P0-2 found:
// a candidate that is in neither half, so the not-computed count — the number D9
// wants an operator to read to tell a source problem from a young scan — is short
// by exactly the candidates nobody measured.
func TestATallyAccountsForEveryCandidateItWasGiven(t *testing.T) {
	measured := Accelerate(seriesOf(t,
		valueObs("AAA", t0, SourceOfficialTradingValue, "0"),
		valueObs("AAA", t0.Add(30*time.Second), SourceOfficialTradingValue, "100"),
		valueObs("AAA", t0.Add(60*time.Second), SourceOfficialTradingValue, "280"),
	), FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	warming := Accelerate(seriesOf(t,
		valueObs("BBB", t0, SourceOfficialTradingValue, "1000"),
		valueObs("BBB", t0.Add(30*time.Second), SourceOfficialTradingValue, "1600"),
	), FieldTradingValue, t0.Add(30*time.Second), DefaultAccelerationWindow)
	halted := Accelerate(seriesOf(t,
		valueObs("CCC", t0, SourceOfficialTradingValue, "1000"),
		valueObs("CCC", t0.Add(30*time.Second), SourceOfficialTradingValue, "1000"),
		valueObs("CCC", t0.Add(60*time.Second), SourceOfficialTradingValue, "1500"),
	), FieldTradingValue, t0.Add(60*time.Second), DefaultAccelerationWindow)
	mixed, _ := NewSourceSeries([]Observation{
		valueObs("DDD", t0, SourceOfficialTradingValue, "1000"),
		valueObs("DDD", t0.Add(30*time.Second), SourceWTSPopular, "1500"),
	})
	refused := Accelerate(mixed, FieldTradingValue, t0.Add(60*time.Second),
		DefaultAccelerationWindow)

	in := []Acceleration{measured, warming, halted, refused, {} /* nobody measured this one */}
	tally := TallyCrossings(in)

	if tally.Total != len(in) {
		t.Errorf("tally.Total = %d, want %d", tally.Total, len(in))
	}
	sum := tally.Computed
	for _, n := range tally.NotComputed {
		sum += n
	}
	if tally.Total != sum {
		t.Errorf("Total = %d but Computed + Σ NotComputed = %d; %d candidates are in "+
			"neither half of the record", tally.Total, sum, tally.Total-sum)
	}
	for reason, want := range map[NotComputed]int{
		NotComputedWarmingUp:     1,
		NotComputedPriorRateZero: 1,
		NotComputedMixedSeries:   1,
		NotComputedUnset:         1,
	} {
		if tally.NotComputed[reason] != want {
			t.Errorf("%s counted %d times, want %d", reason, tally.NotComputed[reason], want)
		}
	}
	if tally.Computed != 1 {
		t.Errorf("computed = %d, want 1", tally.Computed)
	}
}

// --- §3 review, P1-2: the reasons a series cannot be measured are distinct ------

// TestTheThreeWaysAMeasurementCannotStartAreNamedApart.
//
// D9's correction is that WARMING_UP must not absorb causes with different
// remedies — 전자가 많으면 원천 구성 문제이고 후자가 많으면 그냥 방금 시작한 것 — and
// level.go drew the same three lines with LevelNoObservations,
// LevelMixedCandidates and the rest.
func TestTheThreeWaysAMeasurementCannotStartAreNamedApart(t *testing.T) {
	full := seriesOf(t,
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1500"),
		valueObs("005930", t0.Add(60*time.Second), SourceOfficialTradingValue, "2000"),
	)
	mixed, err := NewSourceSeries([]Observation{
		valueObs("005930", t0, SourceOfficialTradingValue, "1000"),
		valueObs("005930", t0.Add(30*time.Second), SourceWTSPopular, "1500"),
	})
	if err == nil {
		t.Fatal("a mixed series was accepted")
	}

	for _, tc := range []struct {
		name   string
		series SourceSeries
		at     time.Time
		want   NotComputed
	}{
		{
			name:   "the source has never reported this symbol",
			series: SourceSeries{},
			at:     t0.Add(60 * time.Second),
			want:   NotComputedNoObservations,
		},
		{
			name:   "the series was refused at construction and the error was dropped",
			series: mixed,
			at:     t0.Add(60 * time.Second),
			want:   NotComputedMixedSeries,
		},
		{
			name:   "every reading is later than the instant asked about",
			series: full,
			at:     t0.Add(-time.Hour),
			want:   NotComputedReadingsAllLater,
		},
		{
			// One reading, so both metrics reach back past the start of the series.
			// This is the only one of the four that means nothing needs fixing.
			name:   "there is a newest reading and no history behind it",
			series: seriesOf(t, valueObs("005930", t0, SourceOfficialTradingValue, "1000")),
			at:     t0,
			want:   NotComputedWarmingUp,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Accelerate(tc.series, FieldTradingValue, tc.at, DefaultAccelerationWindow)
			if got.Why() != tc.want {
				t.Errorf("acceleration reason = %q, want %q", got.Why(), tc.want)
			}
			if _, reason := Rate(tc.series, FieldTradingValue, tc.at,
				DefaultAccelerationWindow); reason != tc.want {
				t.Errorf("rate reason = %q, want %q", reason, tc.want)
			}
		})
	}
}

// --- §3 review, P1-3: the reach is bounded, and the span is recorded ------------

// TestARankMoveRecordsTheSpanItWasMeasuredOver is the protection.
//
// The knob is rank.percentile_gain_30s_pct: 20. A prior reading from the previous
// session produces the same eighty-point gain as one from thirty seconds ago, and
// nothing in the answer said which — RankMove had no Seconds at all, unlike
// Window, which D9 gave one precisely so a reader would not have to re-derive the
// span from timestamps. The bound below is a backstop; this is what actually keeps
// a five-minute move from being read as a thirty-second one.
func TestARankMoveRecordsTheSpanItWasMeasuredOver(t *testing.T) {
	for _, tc := range []struct {
		gap  time.Duration
		want string
	}{
		{30 * time.Second, "30"},
		{90 * time.Second, "90"},
		{5 * time.Minute, "300"},
	} {
		got := rankMoveOver(t, tc.gap)
		if !got.Computed() {
			t.Fatalf("a %s move was not computed: %s", tc.gap, got.Why())
		}
		if got.PercentileGain != "80" {
			t.Errorf("a %s move reported a gain of %q, want %q", tc.gap, got.PercentileGain, "80")
		}
		if got.Seconds != tc.want {
			t.Errorf("a %s move recorded seconds %q, want %q — two identical gains measured "+
				"over different spans are indistinguishable in the record", tc.gap,
				got.Seconds, tc.want)
		}
	}
}

// TestTheWholePlannedBackoffLadderStillProducesARankMove is the Manager's ruling
// on the bound's shape, as something that fails.
//
// The contract's 429 ladder is 30→60→120→300 seconds and the design guarantees it
// will be walked. A bound at a small multiple of `window` refuses at the second
// step, which leaves rank velocity unmeasured during precisely the retreats that
// are planned operation — the state DefaultStalenessTTL was set to avoid for
// first_seen_at, one metric across. Seconds carries the coarseness; refusing
// carries nothing.
func TestTheWholePlannedBackoffLadderStillProducesARankMove(t *testing.T) {
	for _, gap := range []time.Duration{
		30 * time.Second, 60 * time.Second, 120 * time.Second, 300 * time.Second,
	} {
		if got := rankMoveOver(t, gap); !got.Computed() {
			t.Errorf("a %s gap — a planned backoff step, not an anomaly — was refused as %q",
				gap, got.Why())
		}
	}
}

// TestAPriorRankOlderThanTheBackoffCeilingIsRefusedUnderItsOwnReason is the
// backstop, at its boundary.
//
// Past ten minutes the gap is not explained by the planned ladder at all: it is a
// market close, a dead process or a session boundary. That is a different kind of
// fact rather than a coarser measurement of the same one, and Seconds cannot
// rescue it. The boundary is `>=`, matching DefaultStalenessTTL's — store_test.go
// pins a candidate as cooled *at* the TTL rather than a second after, and two
// constants pinned to one number must not disagree about which side of themselves
// they sit on.
func TestAPriorRankOlderThanTheBackoffCeilingIsRefusedUnderItsOwnReason(t *testing.T) {
	if got := rankMoveOver(t, MaxRankPriorAge-time.Second); !got.Computed() {
		t.Errorf("a gap one second inside the bound was refused as %q", got.Why())
	}

	for _, gap := range []time.Duration{
		MaxRankPriorAge, MaxRankPriorAge + time.Second,
		30 * time.Minute, 9 * time.Hour, 40 * time.Hour,
	} {
		got := rankMoveOver(t, gap)
		if got.Computed() {
			t.Errorf("a %s gap reported a gain of %q", gap, got.PercentileGain)
		}
		if got.Why() != NotComputedPriorRankTooOld {
			t.Errorf("a %s gap reported %q, want %q — a prior rank too old to use is not the "+
				"same fact as not having one", gap, got.Why(), NotComputedPriorRankTooOld)
		}
		if got.Seconds == "" || got.From.IsZero() {
			t.Errorf("a %s gap recorded seconds %q and from %v; the refusal has to show how "+
				"old the only prior reading was, which is what says whether this was a close, "+
				"a dead process or a session boundary", gap, got.Seconds, got.From)
		}
		if got.PercentileGain != "" {
			t.Errorf("gain = %q, want empty", got.PercentileGain)
		}
	}
}

// TestTheRankAgeBoundAndTheStalenessTTLShareOneDerivation is the cross-reference
// that keeps the two from drifting.
//
// Both are twice the longest planned rate-limit backoff (300s), against the same
// argument: a retreat that long is normal operation, so neither the candidate nor
// its rank velocity may be discarded while one is in progress. Moving one without
// the other means someone re-derived half of a single decision. Separating them is
// allowed — it just has to be argued for here rather than by editing a literal.
func TestTheRankAgeBoundAndTheStalenessTTLShareOneDerivation(t *testing.T) {
	if MaxRankPriorAge != DefaultStalenessTTL {
		t.Errorf("MaxRankPriorAge = %s and DefaultStalenessTTL = %s; they are twice the same "+
			"300s backoff ceiling and one has moved without the other",
			MaxRankPriorAge, DefaultStalenessTTL)
	}
	if MaxRankPriorAge != 2*300*time.Second {
		t.Errorf("MaxRankPriorAge = %s, want twice the 300s backoff ceiling (%s)",
			MaxRankPriorAge, 2*300*time.Second)
	}
}

// rankMoveOver is one symbol rising 90th → 10th of 100 across `gap`, measured at
// the newer reading. The gain is 80 points whatever the gap is, which is the whole
// difficulty: only Seconds and the bound distinguish the cases.
func rankMoveOver(t *testing.T, gap time.Duration) RankMove {
	t.Helper()
	series := seriesOf(t,
		rankObs(MarketKR, "005930", t0, SourceOfficialTradingValue, 90, 100, false),
		rankObs(MarketKR, "005930", t0.Add(gap), SourceOfficialTradingValue, 10, 100, false),
	)
	return RankChange(series, t0.Add(gap), DefaultAccelerationWindow)
}

// --- §3 review, P2-2: which side of the window was wrong (D6) -------------------

// TestAWindowCarriesTheSourcesOwnFiguresBesideTheDelta.
//
// D6: 원천이 준 값과 우리가 계산한 값을 구분해 저장한다 — when a figure turns out to be
// wrong the record has to say whether it came off the wire or out of this file.
// level.go carries FirstPrice and LastPrice for exactly this; Window kept the
// derived Delta and dropped both cumulatives, so an unreadable one left nothing
// behind at all and a suspicious delta could not be checked against its ends.
func TestAWindowCarriesTheSourcesOwnFiguresBesideTheDelta(t *testing.T) {
	from := valueObs("005930", t0, SourceOfficialTradingValue, "1000")
	to := valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1500")

	got, reason := MeasureWindow(from, to, FieldTradingValue)
	if reason != "" {
		t.Fatalf("window was not measured: %s", reason)
	}
	if got.FromFigure != "1000" || got.ToFigure != "1500" {
		t.Errorf("figures = %q → %q, want %q → %q", got.FromFigure, got.ToFigure, "1000", "1500")
	}

	// The case the pair earns its keep in: the delta is absent and the reason names
	// a kind of failure, but only the raw strings say which end and what it sent.
	broken := valueObs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, "1,500")
	got, reason = MeasureWindow(from, broken, FieldTradingValue)
	if reason != NotComputedFigureUnreadable {
		t.Fatalf("reason = %q, want %q", reason, NotComputedFigureUnreadable)
	}
	if got.ToFigure != "1,500" {
		t.Errorf("the unreadable figure = %q, want %q — the record cannot say what the "+
			"source actually sent", got.ToFigure, "1,500")
	}
	if got.FromFigure != "1000" {
		t.Errorf("the readable end = %q, want %q — the record cannot say which end was "+
			"the problem", got.FromFigure, "1000")
	}
	if got.Delta != "" {
		t.Errorf("delta = %q, want empty", got.Delta)
	}
}

// --- §3 review, P1-4: the wall clock has more than one spelling ------------------

// wallClockNames are the ways code in this package could reach the wall clock.
//
// The guard used to be `time.Now` alone, and a mutation test found that every
// other name on this list passed it. That matters here rather than in the
// abstract: §5's watch loop and its 429 backoff are the code about to be written
// into this package, and `time.After`, `time.NewTicker` and `time.Sleep` are what
// a backoff ladder is naturally spelled with. Each of them binds work to a clock
// no test can move, which is the same defect as reading one.
var wallClockNames = map[string]bool{
	"Now":       true,
	"Since":     true,
	"Until":     true,
	"After":     true,
	"Tick":      true,
	"NewTimer":  true,
	"NewTicker": true,
	"AfterFunc": true,
	"Sleep":     true,
}

// wallClockCalls returns every reference in `file` to one of wallClockNames on the
// standard library's time package, as "line:spelling".
//
// The import is resolved by *path*, not by the identifier `time`. A check that
// looks for the identifier is defeated by `import gotime "time"`, which is a
// rename anyone may make for any reason and which the old check could not see at
// all. A dot import is handled too, because it removes the qualifier entirely.
func wallClockCalls(fset *token.FileSet, file *ast.File) []string {
	locals := map[string]bool{}
	dot := false
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != "time" {
			continue
		}
		switch {
		case spec.Name == nil:
			locals["time"] = true
		case spec.Name.Name == ".":
			dot = true
		case spec.Name.Name == "_":
			// Imported for its side effects; nothing is reachable through it.
		default:
			locals[spec.Name.Name] = true
		}
	}
	if len(locals) == 0 && !dot {
		return nil
	}

	var found []string
	ast.Inspect(file, func(n ast.Node) bool {
		switch e := n.(type) {
		case *ast.SelectorExpr:
			ident, ok := e.X.(*ast.Ident)
			if ok && locals[ident.Name] && wallClockNames[e.Sel.Name] {
				found = append(found, fmt.Sprintf("%d:%s.%s",
					fset.Position(e.Pos()).Line, ident.Name, e.Sel.Name))
			}
		case *ast.CallExpr:
			// A dot import spells the same call with no qualifier at all.
			if ident, ok := e.Fun.(*ast.Ident); ok && dot && wallClockNames[ident.Name] {
				found = append(found, fmt.Sprintf("%d:%s",
					fset.Position(e.Pos()).Line, ident.Name))
			}
		}
		return true
	})
	return found
}

// TestTheWallClockGuardSeesEverySpellingOfTheWallClock is the guard's own test.
//
// The rule below is only worth what its detector catches, and the detector was
// matching one selector name on one identifier. Each row here is a spelling that
// compiles inside this package and passed the old check.
func TestTheWallClockGuardSeesEverySpellingOfTheWallClock(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"time.Now", `import "time"
func f() { _ = time.Now() }`, true},
		{"time.Since", `import "time"
func f() { _ = time.Since(time.Time{}) }`, true},
		{"time.Until", `import "time"
func f() { _ = time.Until(time.Time{}) }`, true},
		{"time.After", `import "time"
func f() { <-time.After(time.Second) }`, true},
		{"time.Tick", `import "time"
func f() { <-time.Tick(time.Second) }`, true},
		{"time.NewTimer", `import "time"
func f() { _ = time.NewTimer(time.Second) }`, true},
		{"time.NewTicker", `import "time"
func f() { _ = time.NewTicker(time.Second) }`, true},
		{"time.AfterFunc", `import "time"
func f() { _ = time.AfterFunc(time.Second, func() {}) }`, true},
		{"time.Sleep", `import "time"
func f() { time.Sleep(time.Second) }`, true},
		{"a renamed import", `import gotime "time"
func f() { _ = gotime.Now() }`, true},
		{"a dot import", `import . "time"
func f() { _ = Now() }`, true},

		// And the things a guard must not fire on, which is why it parses rather
		// than greps: the first version matched text and fired on the comments
		// explaining that the code does not do this.
		{"a comment saying it does not", `import "time"
// This takes its instant from Store.Now() rather than from time.Now().
func f() time.Duration { return time.Second }`, false},
		{"the injected clock", `import "time"
type clk interface{ Now() time.Time }
func f(c clk) time.Time { return c.Now() }`, false},
		{"a struct field of the same name", `import "time"
type s struct{ Now time.Time }
func f(v s) time.Time { return v.Now }`, false},
		{"time used only for its types and constants", `import "time"
func f() time.Duration { return 30 * time.Second }`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "probe.go", "package candidate\n"+tc.body,
				parser.ParseComments)
			if err != nil {
				t.Fatalf("parsing the probe: %v", err)
			}
			got := wallClockCalls(fset, file)
			if (len(got) > 0) != tc.want {
				t.Errorf("guard found %v, want a finding = %v — this spelling reaches a clock "+
					"no test can move", got, tc.want)
			}
		})
	}
}

// --- §3 review, P2-1: the rendering never overstates the arithmetic --------------

// TestFormatDecimalTruncatesAndNeverRoundsTowardsAThreshold pins the rule
// formatDecimal's doc comment states and nothing tested.
//
// The rendered string sits beside the crossing flags, and the flags are computed
// on the exact rational. A rendering that could round up would let the two
// disagree in the one direction that matters — a printed ratio claiming a
// threshold the arithmetic did not reach — and mutating the truncation to
// round-to-nearest left the whole package green.
func TestFormatDecimalTruncatesAndNeverRoundsTowardsAThreshold(t *testing.T) {
	rat := func(num, den int64) *big.Rat { return big.NewRat(num, den) }

	for _, tc := range []struct {
		name  string
		value *big.Rat
		want  string
	}{
		{"a terminating ratio is exact and not padded", rat(18, 10), "1.8"},
		{"a whole number keeps no point", rat(600, 1), "600"},
		{"a third truncates down", rat(1, 3), "0.333333333333"},
		{"two thirds truncates rather than rounding to ...667", rat(2, 3), "0.666666666666"},
		{"a negative truncates towards zero", rat(-1, 3), "-0.333333333333"},
		{
			// 1.8 − 1/(3×10^12). Non-terminating, and inside half a unit of the
			// twelfth digit of the 1.8 shadow threshold: rounding to nearest renders
			// it "1.8" beside a flag saying it did not cross 1.8.
			name:  "just under a threshold, where rounding would print the threshold",
			value: new(big.Rat).SetFrac(big.NewInt(5399999999999), big.NewInt(3000000000000)),
			want:  "1.799999999999",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := formatDecimal(tc.value)
			if got != tc.want {
				t.Errorf("formatDecimal(%s) = %q, want %q", tc.value.RatString(), got, tc.want)
			}
			// And the rule the string is there to keep, stated as arithmetic: what is
			// printed never has a larger magnitude than what was computed.
			back, ok := new(big.Rat).SetString(got)
			if !ok {
				t.Fatalf("the rendering %q is not a decimal", got)
			}
			if new(big.Rat).Abs(back).Cmp(new(big.Rat).Abs(tc.value)) > 0 {
				t.Errorf("formatDecimal(%s) = %q, which is further from zero than the value it "+
					"renders", tc.value.RatString(), got)
			}
		})
	}

	// The pairing that makes it matter: a ratio below a shadow threshold must not
	// render as a string at or above it, whatever the flags say.
	below := new(big.Rat).SetFrac(big.NewInt(5399999999999), big.NewInt(3000000000000))
	for _, th := range ShadowThresholds {
		bar, _ := new(big.Rat).SetString(th)
		if below.Cmp(bar) >= 0 {
			continue
		}
		rendered, ok := new(big.Rat).SetString(formatDecimal(below))
		if !ok {
			t.Fatalf("the rendering %q is not a decimal", formatDecimal(below))
		}
		if rendered.Cmp(bar) >= 0 {
			t.Errorf("a ratio below %s rendered as %q, which reads as %s reached — beside a "+
				"crossing flag that says it was not", th, formatDecimal(below), th)
		}
	}
}
