package candidate

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// bandbehaviour_test.go is the requirement's behavioural half, and it is here
// because the structural half has now been walked around three times.
//
// # Why a second kind of check exists at all
//
//	2026-07-28  the selector did not look at functions returning a Verdict
//	2026-07-29  the vocabulary was construction-only, so one hop passed
//	2026-07-29  the rule exempted functions that return nothing, so a void
//	            function with a *Verdict parameter passed — no helper, no hop,
//	            an idiomatic Go one-liner
//
// Three failures, three different families, and every one of them appeared the same
// way: as a check that passed. A check built by enumerating shapes answers only
// about the shapes somebody thought of, and the shapes nobody thought of are exactly
// the ones a bypass is made of. The guards are kept — they name the offending
// function, which is worth a great deal when one fires — but they cannot be the
// whole guarantee.
//
// The sharpest evidence is in the change's review: with the third bypass in place
// and its band changed from "6" to "0" — a value today's KR data actually crosses —
// both structural guards still passed and two behavioural tests failed. The shadow
// record really was deciding vetoes, and what noticed was behaviour, not structure.
//
// # What this file asserts instead
//
// The property from the spec delta: move the scale and nothing else, and no
// candidate's veto verdict changes. It enumerates no path. Any route that reads the
// scale at the moment the verdict is decided — reflection, type aliases, embedding,
// pointer parameters, package variables, channels, closures, unsafe — comes out as a
// changed verdict, because the assertion is about the answer rather than about the
// route. What is covered is a moment rather than a list of language features: the
// scale being read where the verdict is decided. Everything below is outside it.
//
// # What it does NOT cover, stated here so nobody has to infer it
//
// Four boundaries, each measured on this worktree on 2026-07-29 by putting the line
// in, building, running every guard and every test, and taking it out again.
//
// First, a read of a field that does not depend on the scale. ShadowBand.Value and
// ShadowBand.Measured are computed from the candidate's own arithmetic, so
// `if v.ExtendedBand.Value >= "6"` decides a veto and moving the scale does not
// change what it decides — this file stays green on it. The structural guards are
// what catch that one, which is the concrete reason the requirement asks for both
// and calls them complements rather than alternatives.
//
// Second, a scale value read before the swap and kept — the sharpest of the four:
//
//	var zzEdge = ExtendedGainBands[3]   // package level, so init reads "6"
//	// and in assessOne
//	if exceeds, ok := v.Expansion.GainExceeds(zzEdge); ok && exceeds {
//		v.Chase.Extended = RaisedVeto()
//	}
//
// go build clean, four structural guards green, both behavioural tests green, go vet
// silent. A number lifted straight out of the production scale decided a veto and
// nothing failed. Ordering is not the fix: appending
// {"the production scale, run last", ExtendedGainBands, SeenLatePercentileBands} as a
// seventh variant left both tests green, because init runs before onScale can install
// any variant, so no order of variants reaches a value that was taken already. The
// same value read live where the verdict is decided — a helper returning
// ExtendedGainBands[3], called on that line — fails this file while the structural
// guards stay green. The boundary is a moment and not a shape: a path that reads the
// scale when the verdict is decided is caught whatever its route, and a path that
// read the scale earlier and kept the value is not. The one fix that is not more
// shape enumeration is to make the scale an argument the caller hands in instead of
// package state, so there is nothing to capture at init — a real refactor, not done
// in this change, written down there rather than closed.
//
// Third, anything outside this package. One line in internal/console's render path
// changed what a verdict displays while build, the four structural guards, both
// behavioural tests, the full test suite and go vet were all green. This property
// measures what the producing package returns, not what a consumer does with the
// verdict afterwards. Hiding the scale does not narrow that: BandsFor(code) hands
// back the slice itself, so the values are reachable from any package in the binary
// and unexporting the two variables would change nothing.
//
// Fourth, the structural guards are bounded in their own way. They walk
// verdict-producing FuncDecls, so a package-level var declaration can name the scale
// without any guard firing — that is the door the second boundary goes through — and
// writesOutward does not enumerate every channel a function can answer through. Both
// are written down in the change rather than closed: the first boundary says the
// guards catch it, and they do, while being partial themselves.
//
// Neither layer closes every path and neither is entitled to say it does. This list
// is the record the requirement asks each layer to keep, not an apology for one.

// scaleVariant is one way of spelling the two shadow scales.
type scaleVariant struct {
	name     string
	extended []string
	seenLate []string
}

// scaleVariants are the scales this property is measured over.
//
// They are chosen so that no single band value can be assumed present or absent:
// every edge of the production scale is missing from at least one variant and
// present in at least one, which is what makes a smuggled `Crossed("6")` — or
// `Crossed("0")`, the value today's data actually crosses — come out as a different
// verdict rather than as a different record nobody compares.
func scaleVariants() []scaleVariant {
	return []scaleVariant{
		{"the production scale", ExtendedGainBands, SeenLatePercentileBands},
		{"no scale at all", nil, nil},
		{"a single edge", []string{"0"}, []string{"50"}},
		{
			"a grid that shares no value with the production scale",
			[]string{"1", "3", "5", "7", "9", "11", "21", "31", "51", "101"},
			[]string{"51", "71", "81", "91", "96"},
		},
		{"one integer at a time from -30 to 130", integerScale(-30, 130), integerScale(0, 100)},
		{"the production scale in reverse", reversed(ExtendedGainBands), reversed(SeenLatePercentileBands)},
	}
}

func integerScale(from, to int) []string {
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, strconv.Itoa(i))
	}
	return out
}

func reversed(in []string) []string {
	out := make([]string, 0, len(in))
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return out
}

// onScale runs f with the two scales replaced, and puts them back.
//
// The scales are package variables and this package's tests do not run in parallel,
// which is what makes swapping them the honest way to ask the question: the
// production code reads exactly what it reads in a scan, through whatever path it
// reads it, including paths no test knows the name of.
func onScale(t *testing.T, v scaleVariant, f func()) {
	t.Helper()
	extended, seenLate := ExtendedGainBands, SeenLatePercentileBands
	defer func() { ExtendedGainBands, SeenLatePercentileBands = extended, seenLate }()
	ExtendedGainBands, SeenLatePercentileBands = v.extended, v.seenLate
	f()
}

// aScaleFixture is one candidate with a first sighting at rank/total and a price
// that moved from first to last.
func aScaleFixture(rank, total int, first, last string) (Summary, []Observation) {
	rows := []Observation{
		obs("005930", t0, SourceOfficialTradingValue,
			Reported{Rank: rank, RankTotal: total, Price: first}),
		obs("005930", t0.Add(30*time.Second), SourceOfficialPrices, Reported{Price: last}),
	}
	return Summary{
		Candidate: aCandidate(t0),
		Baseline:  storedBaseline(first, t0, SourceOfficialPrices),
		FirstRank: storedFirstRank(rank, total, t0, SourceOfficialTradingValue),
	}, rows
}

// scaleFixtures spread the candidates across the range the scales cut up: below the
// first price, exactly at it, inside an ordinary day, and a double.
var scaleFixtures = []struct {
	name              string
	rank, total       int
	first, last, gain string
}{
	{"down a fifth", 5, 100, "10000", "8000", "-20"},
	{"exactly flat", 30, 100, "10000", "10000", "0"},
	{"up three", 60, 100, "10000", "10300", "3"},
	{"up seven", 85, 100, "10000", "10700", "7"},
	{"up nine", 92, 100, "10000", "10900", "9"},
	{"doubled and more", 98, 100, "10000", "22000", "120"},
}

// TestChangingTheScaleChangesNoVerdict is the requirement's behavioural property.
//
// One candidate at a time, through the function where a Chase and two shadow
// records are in scope together, over six spellings of the scale. The verdict has to
// be the same object every time.
//
// The failure this is aimed at does not look like a bug in a diff. It looks like
// `annotate(&v)`.
func TestChangingTheScaleChangesNoVerdict(t *testing.T) {
	at := t0.Add(time.Minute)
	th := VetoThresholds{NearHighDistancePct: DefaultNearHighThresholdPct}

	type reading struct {
		chase  Chase
		record string
	}
	got := map[string]map[string]reading{} // fixture → variant → reading

	for _, variant := range scaleVariants() {
		onScale(t, variant, func() {
			for _, fx := range scaleFixtures {
				summary, rows := aScaleFixture(fx.rank, fx.total, fx.first, fx.last)
				v := assessOne(summary, rows, at, th, DefaultAccelerationWindow)
				if !v.ExtendedBand.Measured || !v.SeenLateBand.Measured {
					t.Fatalf("%s on %q: extended measured=%v (%s), seen_late measured=%v (%s); "+
						"an unmeasured record carries no crossings and this property would be "+
						"asserted over nothing",
						fx.name, variant.name, v.ExtendedBand.Measured, v.ExtendedBand.Reason(),
						v.SeenLateBand.Measured, v.SeenLateBand.Reason())
				}
				if v.ExtendedBand.Value != fx.gain {
					t.Fatalf("%s: gain = %q, want %q — the fixture is not the candidate it says",
						fx.name, v.ExtendedBand.Value, fx.gain)
				}
				if got[fx.name] == nil {
					got[fx.name] = map[string]reading{}
				}
				got[fx.name][variant.name] = reading{
					chase: v.Chase,
					record: fmt.Sprintf("seen_late=%v extended=%v",
						v.SeenLateBand.Crossings, v.ExtendedBand.Crossings),
				}
			}
		})
	}

	base := scaleVariants()[0].name
	moved := 0
	for _, fx := range scaleFixtures {
		want := got[fx.name][base]
		for _, variant := range scaleVariants()[1:] {
			have := got[fx.name][variant.name]
			if have.record != want.record {
				moved++
			}
			if reflect.DeepEqual(have.chase, want.chase) {
				continue
			}
			t.Errorf("%s: the veto verdict changed when the only thing that moved was the "+
				"shadow scale (%s → %s).\n  verdict on the production scale: %+v\n"+
				"  verdict on %q:                  %+v\n"+
				"A scale value has reached a decision. Nothing about the candidate changed — "+
				"not its price, not its position, not its age — so whatever read the scale is "+
				"the path the requirement says must not exist, and this test does not know "+
				"where it is. The structural guards in band_test.go and bandguard_test.go are "+
				"the ones that name it; if they are green, the path is a shape they do not "+
				"enumerate and the rule they carry is what has to grow",
				fx.name, base, variant.name, want.chase, variant.name, have.chase)
		}
	}

	// Without this the whole property is vacuous: if every variant produced the same
	// record, "the verdict did not change" would be a statement about nothing having
	// happened rather than about nothing having leaked.
	if moved == 0 {
		t.Fatal("no variant changed any shadow record, so this test asserted that an unchanged " +
			"input produced an unchanged verdict — which is arithmetic, not a property. Either " +
			"the fixtures stopped being measurable or the scales stopped being read")
	}
	t.Logf("%d of %d (fixture, variant) pairs recorded a different scale reading",
		moved, len(scaleFixtures)*(len(scaleVariants())-1))
}

// TestChangingTheScaleChangesNoVerdictAcrossAWholeScan is the same property over the
// public entry point and a real store.
//
// assessOne is where the two are in scope together and it is therefore the likeliest
// place; it is not the only one. Assess is what a scan calls, so this reads the
// verdicts the way anything downstream of discovery reads them, off candidates a
// Cycle actually wrote.
func TestChangingTheScaleChangesNoVerdictAcrossAWholeScan(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	// Two readings so the candidates have a first price and a later, different one.
	// Without the second every gain is 0 and the scale has nothing to resolve.
	src := &scriptedSource{
		id: SourceOfficialTradingValue,
		readings: [][]Row{
			{
				pricedRow("005930", 5, 100, "10000"),
				pricedRow("000660", 30, 100, "10000"),
				pricedRow("035420", 60, 100, "10000"),
				pricedRow("051910", 85, 100, "10000"),
				pricedRow("207940", 98, 100, "10000"),
			},
			{
				pricedRow("005930", 5, 100, "8000"),
				pricedRow("000660", 30, 100, "10000"),
				pricedRow("035420", 60, 100, "10300"),
				pricedRow("051910", 85, 100, "10700"),
				pricedRow("207940", 98, 100, "22000"),
			},
		},
	}
	opts := cycleOpts(MarketKR, src)
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	clk.Advance(30 * time.Second)
	opts.Schedule = NewSchedule(nil)
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("second Cycle: %v", err)
	}

	assessOpts := AssessOptions{
		Market:     MarketKR,
		At:         clk.Now(),
		Thresholds: VetoThresholds{NearHighDistancePct: DefaultNearHighThresholdPct},
	}

	chases := map[string]map[string]Chase{} // variant → symbol → chase
	records := map[string]map[string]string{}
	for _, variant := range scaleVariants() {
		onScale(t, variant, func() {
			verdicts, err := Assess(ctx, s, assessOpts)
			if err != nil {
				t.Fatalf("Assess on %q: %v", variant.name, err)
			}
			if len(verdicts) != len(src.readings[0]) {
				t.Fatalf("Assess returned %d verdicts on %q, want %d",
					len(verdicts), variant.name, len(src.readings[0]))
			}
			chases[variant.name] = map[string]Chase{}
			records[variant.name] = map[string]string{}
			for _, v := range verdicts {
				chases[variant.name][v.Summary.Symbol] = v.Chase
				records[variant.name][v.Summary.Symbol] = fmt.Sprintf("%v|%v",
					v.SeenLateBand.Crossings, v.ExtendedBand.Crossings)
			}
		})
	}

	base := scaleVariants()[0].name
	measured, moved := 0, 0
	for symbol, want := range chases[base] {
		if records[base][symbol] != "[]|[]" {
			measured++
		}
		for _, variant := range scaleVariants()[1:] {
			if records[variant.name][symbol] != records[base][symbol] {
				moved++
			}
			if reflect.DeepEqual(chases[variant.name][symbol], want) {
				continue
			}
			t.Errorf("%s: the veto verdict off a real store changed when only the shadow scale "+
				"moved (%s → %s):\n  %+v\n  %+v",
				symbol, base, variant.name, want, chases[variant.name][symbol])
		}
	}
	if measured == 0 || moved == 0 {
		t.Fatalf("measured records = %d, variants that moved a record = %d; both have to be "+
			"non-zero or this scan asserted nothing", measured, moved)
	}
}
