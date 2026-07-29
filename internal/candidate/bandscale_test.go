package candidate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"math/big"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// bandscale_test.go is the change refine-extended-shadow-bands.
//
// The shadow record for `extended` existed, measured 131 candidates of 156, and
// answered no question: every one of the 131 landed in the same bucket, because the
// finest edge on the scale was 10% and `bandCrossings` compares with `>=`. The
// arithmetic held, the measured count was honest, and nothing failed. What was
// missing was the only thing the record is for — a distribution somebody could
// approve a threshold from.
//
// So these tests are about the scale rather than the instrument: that the sign
// change is on it, that the ordinary intraday range is cut into pieces, and that the
// five values already reported against are still there and still mean what they
// meant. The one thing they must not do is turn a scale value into a decision, and
// TestNoFunctionThatProducesAVerdictCanSeeAShadowBand is what stops that.

// anExtendedBand is a measured extended band from `first` to `last`.
func anExtendedBand(t *testing.T, first, last string) ShadowBand {
	t.Helper()
	at := t0.Add(30 * time.Second)
	band := MeasureExtendedBand(anExpansion(first, last, t0), aCandidate(t0), at,
		VetoThresholds{NearHighDistancePct: DefaultNearHighThresholdPct})
	if !band.Measured {
		t.Fatalf("%s → %s produced an unmeasured band (%s); the row is not testing what it says",
			first, last, band.Reason())
	}
	return band
}

// crossedBands is the set of bands a record reached, in scale order.
func crossedBands(b ShadowBand) []string {
	var out []string
	for _, c := range b.Crossings {
		if c.Crossed {
			out = append(out, c.Band)
		}
	}
	return out
}

// TestADeclineAndASmallGainAreNotTheSameRecord is the sign change (task 2.1).
//
// With every edge at 10 or above, a candidate 20% below the price it was first seen
// at and a candidate 9% above it produced the identical record: crossed nothing.
// Both are "not a chase risk" and they are not the same fact — one has not run yet
// and the other has already given it back — and the person choosing a threshold is
// reading the distribution, not the conclusion.
//
// What band `0` divides is "declined" from "did not decline", NOT decline from gain:
// the comparison is `>=`, so a candidate exactly at its first price crosses it. That
// is the correct place for the line. A record that has given nothing back belongs
// with the ones that have gained, because the question the scale is asked is how far
// a candidate has run.
func TestADeclineAndASmallGainAreNotTheSameRecord(t *testing.T) {
	declined := anExtendedBand(t, "10000", "8000") // −20%
	flat := anExtendedBand(t, "10000", "10000")    // 0%
	gained := anExtendedBand(t, "10000", "10900")  // +9%

	if declined.Value != "-20" || flat.Value != "0" || gained.Value != "9" {
		t.Fatalf("values = %q/%q/%q, want -20/0/9; the premise of this test is the arithmetic",
			declined.Value, flat.Value, gained.Value)
	}

	for _, pair := range []struct{ a, b ShadowBand }{
		{declined, flat}, {declined, gained}, {flat, gained},
	} {
		if fmt.Sprint(crossedBands(pair.a)) == fmt.Sprint(crossedBands(pair.b)) {
			t.Errorf("%s%% and %s%% produced the same record (%v); a scale that cannot tell a "+
				"decline from a gain records both as \"has not run\" and a threshold derived "+
				"from it was derived from a distribution that never happened",
				pair.a.Value, pair.b.Value, crossedBands(pair.a))
		}
	}

	// Named separately from the pairwise check, because the pairwise check would
	// also pass if `0` moved and something else happened to separate them.
	if declined.Crossed("0") {
		t.Errorf("a candidate 20%% below its first price crossed band 0")
	}
	if !flat.Crossed("0") {
		t.Errorf("a candidate exactly at its first price did not cross band 0; the comparison " +
			"is >=, and band 0 divides \"declined\" from \"did not decline\"")
	}
}

// TestTheOrdinaryIntradayRangeIsResolved is task 2.2.
//
// band.go already says 10 is "inside a normal day's range for a name that is on a
// trading-value ranking at all". Everything a candidate does on an ordinary day was
// therefore one bucket. Five gains inside that range have to produce five records.
func TestTheOrdinaryIntradayRangeIsResolved(t *testing.T) {
	seen := map[string]string{}
	for _, tc := range []struct{ last, gain string }{
		{"10100", "1"}, {"10300", "3"}, {"10500", "5"}, {"10700", "7"}, {"10900", "9"},
	} {
		band := anExtendedBand(t, "10000", tc.last)
		if band.Value != tc.gain {
			t.Fatalf("value = %q, want %q", band.Value, tc.gain)
		}
		key := fmt.Sprint(crossedBands(band))
		if other, clash := seen[key]; clash {
			t.Errorf("+%s%% and +%s%% both recorded %s; the scale does not resolve the range it "+
				"is supposed to resolve", other, tc.gain, key)
		}
		seen[key] = tc.gain
	}
}

// TestTheBandsAlreadyReportedAgainstDoNotMove is D4 (task 2.3).
//
// `crossed 10 0 · 20 0 · 30 0 · 50 0 · 100 0` has been read by people. Moving one of
// those values would make the same column name mean something else in their memory,
// which is a migration in a person's head for a record that has none on disk. New
// edges go below them; these five stay, in this order, at the end.
//
// Nothing checked this before: `"50"` could be deleted from the declaration and the
// whole suite still passed.
func TestTheBandsAlreadyReportedAgainstDoNotMove(t *testing.T) {
	reported := []string{"10", "20", "30", "50", "100"}
	got := ExtendedGainBands
	if len(got) < len(reported) {
		t.Fatalf("the scale is %v, shorter than the five values already reported against", got)
	}
	tail := got[len(got)-len(reported):]
	for i, want := range reported {
		if tail[i] != want {
			t.Errorf("the scale ends %v, want it to end %v — the five values already reported "+
				"against keep their values and their order (D4)", tail, reported)
			break
		}
	}

	// The scale is read in order by bandCrossings, by orderedCounts and by the
	// console, and a distribution rendered out of order is one nobody can read down.
	// This is a property of the list rather than a copy of it.
	for i := 1; i < len(got); i++ {
		prev, ok := new(big.Rat).SetString(got[i-1])
		if !ok {
			t.Fatalf("band %q is not a number; bandCrossings skips it silently and the column "+
				"disappears from the report without anything failing", got[i-1])
		}
		cur, ok := new(big.Rat).SetString(got[i])
		if !ok {
			t.Fatalf("band %q is not a number", got[i])
		}
		if prev.Cmp(cur) >= 0 {
			t.Errorf("the scale is not increasing at %s → %s (%v)", got[i-1], got[i], got)
		}
	}
}

// TestTheTallyHasOneColumnPerBandAndCountsThemRight is task 2.4.
//
// The existing invariant (Total == Measured + Σ NotMeasured) holds for any number of
// bands and says nothing about this change. What has to hold is that every band of
// the code is a column even at zero, and that the counts are the counts.
func TestTheTallyHasOneColumnPerBandAndCountsThemRight(t *testing.T) {
	in := []ShadowBand{
		anExtendedBand(t, "10000", "8000"),  // −20
		anExtendedBand(t, "10000", "10000"), // 0
		anExtendedBand(t, "10000", "10300"), // +3
		anExtendedBand(t, "10000", "10900"), // +9
		anExtendedBand(t, "10000", "12000"), // +20
		{},                                  // the unassigned one
	}
	tally := TallyBands(VetoExtended, in)

	if len(tally.Crossed) != len(BandsFor(VetoExtended)) {
		t.Fatalf("crossed has %d columns for %d bands (%v); a band that is missing from the "+
			"tally renders as a band nobody crossed",
			len(tally.Crossed), len(BandsFor(VetoExtended)), tally.Crossed)
	}
	if tally.Total != len(in) || tally.Measured != len(in)-1 {
		t.Fatalf("total = %d, measured = %d, want %d and %d",
			tally.Total, tally.Measured, len(in), len(in)-1)
	}
	for band, want := range map[string]int{
		"0": 4, "2": 3, "4": 2, "6": 2, "8": 2, "10": 1, "20": 1, "30": 0, "50": 0, "100": 0,
	} {
		if got := tally.Crossed[band]; got != want {
			t.Errorf("crossed %s = %d, want %d (from -20, 0, +3, +9, +20)", band, got, want)
		}
	}
}

// TestTheTallyCarriesTheDistributionAndNotOnlyTheCrossings is task 2.7.
//
// The crossings are counted against edges, and this change exists because a set of
// edges produced 131 identical answers. The quantiles do not depend on the edges
// being right, which is the property that makes the next spacing decision an
// observation instead of another guess.
//
// Nearest rank, so every figure is a value some candidate actually had. With five
// values that is min, the 2nd, the 3rd, the 4th and the 5th.
func TestTheTallyCarriesTheDistributionAndNotOnlyTheCrossings(t *testing.T) {
	in := []ShadowBand{
		anExtendedBand(t, "10000", "8000"),  // −20
		anExtendedBand(t, "10000", "12000"), // +20 — out of order on purpose
		anExtendedBand(t, "10000", "10000"), // 0
		anExtendedBand(t, "10000", "10900"), // +9
		anExtendedBand(t, "10000", "10300"), // +3
		{},                                  // unmeasured: carries no value
	}
	tally := TallyBands(VetoExtended, in)

	if tally.QuantileBase != 5 {
		t.Errorf("quantile base = %d, want 5 — the unmeasured record has no value to place",
			tally.QuantileBase)
	}
	want := map[string]string{"min": "-20", "p25": "0", "median": "3", "p75": "9", "max": "20"}
	if len(tally.Quantiles) != len(BandQuantilePoints) {
		t.Fatalf("quantiles = %v, want %d positions", tally.Quantiles, len(BandQuantilePoints))
	}
	for i, q := range tally.Quantiles {
		if q.Name != BandQuantilePoints[i].Name {
			t.Errorf("quantile %d is %q, want %q — the order is the order they are read in",
				i, q.Name, BandQuantilePoints[i].Name)
		}
		if q.Value != want[q.Name] {
			t.Errorf("%s = %q, want %q (from -20, 0, +3, +9, +20)", q.Name, q.Value, want[q.Name])
		}
	}

	// The distribution has to survive the collapse this change is about: if every
	// value falls in one bucket the crossings are all equal and the quantiles are
	// still the only thing that says where they were.
	flat := TallyBands(VetoExtended, []ShadowBand{
		anExtendedBand(t, "10000", "10010"), // +0.1
		anExtendedBand(t, "10000", "10030"), // +0.3
	})
	for _, c := range flat.Crossed {
		if c != 0 && c != flat.Measured {
			t.Fatalf("the premise failed: %v is not a collapsed set of crossings", flat.Crossed)
		}
	}
	if flat.Quantiles[0].Value == flat.Quantiles[len(flat.Quantiles)-1].Value {
		t.Errorf("min and max are both %q for values that differ; the surface that was supposed "+
			"to survive the collapse collapsed with it", flat.Quantiles[0].Value)
	}
}

// TestAnUnmeasuredTallyHasNoQuantiles.
//
// Same rule as ShadowBand.Value: an unmeasured record carries its inputs and never
// arithmetic. A tally of nothing reporting `median 0` would put a number where there
// is no measurement, and 0 is the one number that reads as an answer.
func TestAnUnmeasuredTallyHasNoQuantiles(t *testing.T) {
	tally := TallyBands(VetoExtended, []ShadowBand{{}, {}})
	if tally.QuantileBase != 0 || len(tally.Quantiles) != 0 {
		t.Errorf("base = %d, quantiles = %v, want none of either",
			tally.QuantileBase, tally.Quantiles)
	}
}

// TestANegativeValueRendersTheWayTheRestOfThePackageSaysItDoes is task 3.4.
//
// Band 0 puts negative values on a surface for the first time, so the rendering of
// one is worth pinning rather than assuming. formatDecimal truncates towards zero,
// which for a negative non-terminating value is not the floor: the printed figure is
// up to one unit in the last place closer to zero than the exact one. distancePct
// already documents this for the package's other negative value.
//
// It cannot move a crossing, because bandCrossings compares the exact rational and
// never the string. It can move a quantile by that last unit, because TallyBands
// reads the rendered value back — which is the right way round, since the quantiles
// are meant to describe the numbers a person actually sees.
func TestANegativeValueRendersTheWayTheRestOfThePackageSaysItDoes(t *testing.T) {
	band := anExtendedBand(t, "3", "2") // −100/3 %, which has no exact decimal
	if !strings.HasPrefix(band.Value, "-33.333333333333") {
		t.Errorf("value = %q, want a truncation of -33.3…; the sign is on the surface now and "+
			"a negative that renders as something else would be read as a gain", band.Value)
	}
	if band.Crossed("0") {
		t.Errorf("a candidate a third below its first price crossed band 0 (%q)", band.Value)
	}
	// The crossing is decided on the exact rational, so the truncation cannot reach
	// it. Pinned here because band 0 is the first edge a rendering could straddle.
	tiny := anExtendedBand(t, "300000000000000000", "299999999999999999")
	if tiny.Crossed("0") {
		t.Errorf("a value a hair below zero crossed band 0 (%q); the crossing is being decided "+
			"on the rendering rather than the arithmetic", tiny.Value)
	}
}

// TestATallyThatResolvedNothingSaysSo is the spec's alarm scenario.
//
// spec candidate-discovery: when a shadow tally has one or more measured records and
// all of them produced the same crossings, the reporting surface must not show it as
// an ordinary number. That state is the whole reason this change exists and it was
// found by a person reading `crossed 10 0 · 20 0 · 30 0 · 50 0 · 100 0` twice — the
// invariants held, the counts were honest, nothing failed.
//
// The check is arithmetic on the tally and needs no idea of what the scale was for,
// which is what makes it possible at all.
func TestATallyThatResolvedNothingSaysSo(t *testing.T) {
	// The state of 2026-07-28, reproduced: everything below the finest edge.
	collapsed := TallyBands(VetoExtended, []ShadowBand{
		anExtendedBand(t, "10000", "10010"), // +0.1
		anExtendedBand(t, "10000", "10030"), // +0.3
		anExtendedBand(t, "10000", "10050"), // +0.5
	})
	if !collapsed.Collapsed() {
		t.Errorf("three records that all crossed exactly band 0 are not reported as collapsed "+
			"(%v); a scale that separates nothing is what nobody can pick a threshold out of",
			collapsed.Crossed)
	}
	if collapsed.CollapsedAlarm() == "" {
		t.Error("a collapsed tally carries no sentence; a boolean nobody renders is a state " +
			"that stays invisible")
	}

	// One band crossed by some and not others is a scale that separated something.
	resolved := TallyBands(VetoExtended, []ShadowBand{
		anExtendedBand(t, "10000", "10010"), // +0.1
		anExtendedBand(t, "10000", "10300"), // +3
	})
	if resolved.Collapsed() {
		t.Errorf("a tally where band 2 separates two records is reported as collapsed (%v)",
			resolved.Crossed)
	}
	if resolved.CollapsedAlarm() != "" {
		t.Errorf("a tally that resolved something carries an alarm: %q", resolved.CollapsedAlarm())
	}

	// Measuring nothing is not the same failure and must not raise this one: the
	// `measured 0 of N` row already says nobody was measured, and an alarm there
	// would be permanently on, which is how an alarm stops being read.
	none := TallyBands(VetoExtended, []ShadowBand{{}, {}})
	if none.Collapsed() {
		t.Error("a tally that measured nothing raises the collapse alarm; an alarm that is " +
			"always on is one nobody reads")
	}
}

// --- the scale decides nothing -----------------------------------------------------

// TestMeasureExtendedBandNeverReadsTheVetoThreshold is task 1.4.
//
// MeasureExtendedBand takes a VetoThresholds and uses part of it: th.inputAge() is
// the ceiling on how old the latest price may be, and the band applies it because a
// price too old to be about now produces a plausible wrong number in the very
// dataset a threshold will be derived from. What it must never read is
// ExtendedGainPct — the threshold itself — because reading it is the moment the
// record starts deciding.
//
// Two checks, because either alone is weak. The first reads the source: a mention is
// a defect even if no test exercises the path. The second runs the function with
// thresholds that differ in nothing but that field, which catches a read that
// arrives through a helper.
func TestMeasureExtendedBandNeverReadsTheVetoThreshold(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "band.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing band.go: %v", err)
	}
	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "MeasureExtendedBand" {
			continue
		}
		found = true
		ast.Inspect(fn, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == "ExtendedGainPct" {
				t.Errorf("MeasureExtendedBand mentions ExtendedGainPct; the scale is an "+
					"instrument and the threshold is a decision, and the whole reason the "+
					"signature carries VetoThresholds is that everything in it EXCEPT that "+
					"field is checked (%s)", fset.Position(id.Pos()))
			}
			return true
		})
	}
	if !found {
		t.Fatal("MeasureExtendedBand was not found in band.go; the check would pass vacuously")
	}

	// The differential half. inputAge is held equal so the only difference is the
	// threshold, and a band that changed with it would be comparing against a number
	// nobody approved.
	at := t0.Add(30 * time.Second)
	expansion := anExpansion("10000", "10600", t0)
	base := VetoThresholds{NearHighDistancePct: DefaultNearHighThresholdPct}
	withThreshold := base
	withThreshold.ExtendedGainPct = "6"
	a := MeasureExtendedBand(expansion, aCandidate(t0), at, base)
	b := MeasureExtendedBand(expansion, aCandidate(t0), at, withThreshold)
	if fmt.Sprint(a) != fmt.Sprint(b) {
		t.Errorf("the band changed when ExtendedGainPct went from %q to %q:\n  %v\n  %v",
			base.ExtendedGainPct, withThreshold.ExtendedGainPct, a, b)
	}
}

// bandIdentityFiles are the non-test files allowed to name a shadow band type.
//
// Four: where the record is defined, where it is attached to a verdict, and the two
// screens that display it. A store, a migration or a journal is not on this list.
var bandIdentityFiles = map[string]bool{
	"internal/candidate/band.go":  true,
	"internal/candidate/watch.go": true,
	"cmd/tossctl/candidate.go":    true,
	"internal/console/signals.go": true,
}

// TestAShadowBandIsNotPersistedAnywhere is D6 as a test rather than a sentence
// (task 2.5).
//
// This change moves the scale, and the reason it needs no migration, no backfill and
// no schema version is that a band is computed on every pass and stored nowhere. A
// grep for a column called `band` proves nothing — the column could be called
// anything — so what is asserted is the identity: the three band types are named
// only where a record is made and where it is shown.
//
// If a later change does persist one, this test is what it meets, and the question
// it will have to answer is what the stored rows mean after the scale moves under
// them.
func TestAShadowBandIsNotPersistedAnywhere(t *testing.T) {
	types := map[string]bool{"ShadowBand": true, "BandTally": true, "BandCrossing": true}
	root := filepath.Join("..", "..")
	seen := map[string]int{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "bin", "openspec", "docs":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel := filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator)))
		fset := token.NewFileSet()
		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return fmt.Errorf("parsing %s: %w", rel, perr)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok || !types[id.Name] {
				return true
			}
			seen[rel]++
			if !bandIdentityFiles[rel] {
				t.Errorf("%s names %s. A shadow band is recomputed on every assessment and "+
					"stored nowhere, which is why moving the scale needs no migration. A file "+
					"outside the record and the two screens either persists one or decides "+
					"with one, and both change what this change is allowed to do",
					rel, id.Name)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking the tree: %v", err)
	}
	// Without this the check passes when the walk matches nothing at all — the same
	// empty-set failure the snapshot guard next door was written to refuse.
	if len(seen) == 0 {
		t.Fatal("no file in the tree names a shadow band type; the walk stopped matching " +
			"rather than the types stopping being used")
	}
	for _, required := range []string{"internal/candidate/band.go", "internal/console/signals.go"} {
		if seen[required] == 0 {
			t.Errorf("%s names no shadow band type; the walk is not reaching it", required)
		}
	}
}

// TestATallyWithNoScaleIsNotReportedAsCollapsed.
//
// issues I3 records an obligation this change leaves behind: the day extended's
// threshold is approved, BandsFor(VetoExtended) has to start returning nil the way
// near_high's already does, because a live veto with a shadow record beside it is an
// invitation to read the shadow instead. On that day TallyBands seeds Crossed from
// an empty scale, the loop inside Collapsed runs zero times, and "every band was
// crossed by all of them or none of them" is vacuously true for any tally with a
// single measured record in it.
//
// The alarm would then be lit on that code permanently. That is worse than not
// having it: a light that is always on is the thing that teaches an operator to skip
// the block it is in, and this alarm now renders on a page somebody leaves open.
//
// A scale that does not exist did not fail to resolve anything. There is no
// instrument here, so there is no instrument failure to report — which is a
// different statement from the one the sentence would make.
func TestATallyWithNoScaleIsNotReportedAsCollapsed(t *testing.T) {
	if len(BandsFor(VetoNearHigh)) != 0 {
		t.Fatal("near_high has a scale now; this test was written around it having none, " +
			"which is also the shape BandsFor(VetoExtended) has to take under I3")
	}
	tally := TallyBands(VetoNearHigh, []ShadowBand{
		{Code: VetoNearHigh, Measured: true, Value: "1"},
		{Code: VetoNearHigh, Measured: true, Value: "2"},
	})
	if tally.Measured != 2 {
		t.Fatalf("measured = %d, want 2 — the fixture is not exercising the branch", tally.Measured)
	}
	if tally.Collapsed() {
		t.Errorf("a tally with no bands at all reports itself collapsed (crossed = %v); the "+
			"sentence claims every measured record produced the same crossings when there "+
			"were no crossings for any of them to produce", tally.Crossed)
	}
	if got := tally.CollapsedAlarm(); got != "" {
		t.Errorf("and it carries an alarm sentence: %q", got)
	}
}
