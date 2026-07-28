package candidate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// band_test.go is task 4.10.
//
// D18 found that two of the three vetoes have no threshold anywhere in the
// repository, and refused to invent one. The rule it applied instead is the one
// §3.5 already applies to the acceleration: record the crossings against a band of
// values and judge nothing. A band is a measuring instrument and may be chosen
// without a source; a veto threshold may not.
//
// So the tests here have two jobs. The first is that the record exists even where
// the veto cannot run. The second is the harder one: that a band can never become
// the veto it is standing in for.

// aSighting is a measured first sighting at the given position.
func aSighting(rank, total int, at time.Time) Sighting {
	return MeasureFirstSighting(
		aCandidate(at),
		storedFirstRank(rank, total, at, SourceOfficialTradingValue),
		[]Observation{rankedObs(at, SourceOfficialTradingValue, rank, total)})
}

// anExpansion is a measured expansion from `first` to `last`.
func anExpansion(first, last string, at time.Time) Expansion {
	return MeasureExpansion(
		storedBaseline(first, at, SourceOfficialPrices),
		[]Observation{
			levelObs(at, SourceOfficialPrices, first),
			levelObs(at.Add(30*time.Second), SourceOfficialPrices, last),
		})
}

// --- the record exists where the veto cannot run --------------------------------

// TestAVetoWithNoThresholdStillLeavesAShadowRecord is D18's whole point.
//
// seen_late and extended are THRESHOLD_ABSENT and will stay that way until a month
// of measurement lets a person approve a number. If nothing were recorded in the
// meantime, that month would produce nothing to approve from — the shadow record is
// what turns "we have no threshold" into "we will have one".
func TestAVetoWithNoThresholdStillLeavesAShadowRecord(t *testing.T) {
	// The contract fixes no value for either, so this is what §5 actually passes.
	th := VetoThresholds{NearHighDistancePct: DefaultNearHighThresholdPct}

	sighting := aSighting(12, 100, t0)
	if state := AssessSeenLate(sighting, th); state.Reason() != VetoThresholdAbsent {
		t.Fatalf("seen_late = %v, want unmeasured (%s) — the premise of this test",
			state, VetoThresholdAbsent)
	}
	band := MeasureSeenLateBand(sighting)
	if !band.Measured {
		t.Fatalf("the seen_late band is %v; a veto with no threshold and no shadow record "+
			"produces nothing to derive a threshold from", band.Reason())
	}
	if band.Value != "88" {
		t.Errorf("seen_late band value = %q, want %q (12th of 100)", band.Value, "88")
	}
	if !band.Crossed("80") || band.Crossed("90") {
		t.Errorf("crossings = %v, want 80 crossed and 90 not", band.Crossings)
	}

	expansion := anExpansion("10000", "13000", t0)
	if state := AssessExtended(expansion, aCandidate(t0), t0.Add(30*time.Second), th); state.Reason() != VetoThresholdAbsent {
		t.Fatalf("extended = %v, want unmeasured (%s) — the premise of this test",
			state, VetoThresholdAbsent)
	}
	eband := MeasureExtendedBand(expansion, aCandidate(t0), t0.Add(30*time.Second), th)
	if !eband.Measured {
		t.Fatalf("the extended band is %v", eband.Reason())
	}
	if eband.Value != "30" {
		t.Errorf("extended band value = %q, want %q (10000 to 13000)", eband.Value, "30")
	}
	if !eband.Crossed("30") || eband.Crossed("50") {
		t.Errorf("crossings = %v, want 30 crossed and 50 not", eband.Crossings)
	}
}

// TestTheZeroValueOfAShadowBandIsUnmeasured.
//
// The §3 review found Acceleration{}.Computed() == true — an unassigned struct
// reporting itself as measured — inside the section built to stop exactly that. A
// band is filled in per candidate from a map or a slice, so the same three
// spellings reach it, and a band that read as "measured, crossed nothing" would put
// a candidate nobody looked at in the quiet bucket.
func TestTheZeroValueOfAShadowBandIsUnmeasured(t *testing.T) {
	var b ShadowBand
	if b.Measured {
		t.Error("the zero ShadowBand reports itself as measured")
	}
	if b.Reason() != VetoNotEvaluated {
		t.Errorf("reason = %q, want %q — a reason code that is sometimes empty is one an "+
			"operator learns to skip", b.Reason(), VetoNotEvaluated)
	}
	for _, band := range SeenLatePercentileBands {
		if b.Crossed(band) {
			t.Errorf("an unmeasured band crossed %s; measuring nothing is not crossing", band)
		}
	}
}

// TestAnUnmeasurableInputLeavesNoBandCrossings.
//
// The counting form of the same rule. A band computed from an input that could not
// be measured carries no crossings at all, so it cannot contribute to a bucket —
// and TallyBands then has to account for it under its reason.
func TestAnUnmeasurableInputLeavesNoBandCrossings(t *testing.T) {
	// A first sighting with no stored rank: the ordinary state before a scan has
	// written the column, and after a prune.
	unmeasured := MeasureFirstSighting(aCandidate(t0), FirstRank{}, nil)
	band := MeasureSeenLateBand(unmeasured)
	if band.Measured {
		t.Fatal("a sighting nobody could measure produced a measured band")
	}
	if band.Reason() != VetoNoObservations {
		t.Errorf("reason = %q, want %q — the band has to name the same missing input the "+
			"veto does or an operator gets two different stories", band.Reason(), VetoNoObservations)
	}
	if len(band.Crossings) != 0 {
		t.Errorf("crossings = %v, want none", band.Crossings)
	}
	if band.Value != "" {
		t.Errorf("value = %q, want empty — an unmeasured band carries its inputs, never arithmetic",
			band.Value)
	}
}

// TestABandNamesTheSameMissingInputTheVetoWould is the drift guard.
//
// The band applies the same identity and age checks the veto does — a baseline from
// a dead life, a baseline later than the first sighting, a price too old to be about
// now — because a band built from those inputs would record a plausible wrong number
// into the very dataset a threshold is going to be derived from. The two chains are
// written twice (the veto's decides, the band's records), so this pins them
// together the way fsguard_drift_test.go pins the filesystem allowlist.
func TestABandNamesTheSameMissingInputTheVetoWould(t *testing.T) {
	// A threshold that parses, so the veto's own threshold reasons never fire and
	// what is left is exactly the input chain.
	th := VetoThresholds{ExtendedGainPct: "50", NearHighDistancePct: DefaultNearHighThresholdPct}
	at := t0.Add(30 * time.Second)

	for _, tc := range []struct {
		name      string
		expansion Expansion
		candidate Candidate
		at        time.Time
	}{
		{"an unmeasured expansion", Expansion{}, aCandidate(t0), at},
		{"no candidate summary", anExpansion("10000", "13000", t0), Candidate{}, at},
		{
			"a baseline later than the first sighting",
			anExpansion("10000", "13000", t0.Add(2*DefaultStalenessTTL)),
			aCandidate(t0), t0.Add(2*DefaultStalenessTTL + 30*time.Second),
		},
		{
			"a baseline from a life that ended",
			anExpansion("10000", "13000", t0),
			aCandidate(t0.Add(2 * DefaultStalenessTTL)), at,
		},
		{
			"a latest price older than the ceiling",
			anExpansion("10000", "13000", t0),
			aCandidate(t0), t0.Add(MaxVetoInputAge + time.Minute),
		},
		{"no instant to measure ages against", anExpansion("10000", "13000", t0), aCandidate(t0), time.Time{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			veto := AssessExtended(tc.expansion, tc.candidate, tc.at, th)
			band := MeasureExtendedBand(tc.expansion, tc.candidate, tc.at, th)
			if veto.Measured() {
				t.Fatalf("the veto measured this input; the row is not testing what it says")
			}
			if band.Measured {
				t.Fatalf("the veto could not measure this input and the band did; the band is "+
					"recording a number the veto refused (%s)", veto.Reason())
			}
			if band.Reason() != veto.Reason() {
				t.Errorf("band reason = %q, veto reason = %q — the two chains have drifted and "+
					"the screen now tells two stories about one candidate",
					band.Reason(), veto.Reason())
			}
		})
	}
}

// --- the tally ------------------------------------------------------------------

// TestTheBandTallyAccountsForEveryCandidate is CrossingTally's invariant applied
// here.
//
// The §3 review found TallyCrossings dropping the unmeasured candidates out of
// both of its halves, which flipped a screen's default reading from "mostly
// unchecked" to "all measured and quiet". Total == Measured + Σ NotMeasured makes
// that an arithmetic failure instead of a number that is quietly short.
func TestTheBandTallyAccountsForEveryCandidate(t *testing.T) {
	in := []ShadowBand{
		MeasureSeenLateBand(aSighting(12, 100, t0)),
		MeasureSeenLateBand(aSighting(140, 150, t0)),
		MeasureSeenLateBand(MeasureFirstSighting(aCandidate(t0), FirstRank{}, nil)),
		{}, // the unassigned one: a map miss, an unfilled slot
	}
	tally := TallyBands(VetoSeenLate, in)

	if tally.Total != len(in) {
		t.Errorf("total = %d, want %d", tally.Total, len(in))
	}
	sum := tally.Measured
	for _, n := range tally.NotMeasured {
		sum += n
	}
	if sum != tally.Total {
		t.Errorf("measured %d + not measured %d != total %d; a candidate fell out of both halves",
			tally.Measured, sum-tally.Measured, tally.Total)
	}
	if tally.NotMeasured[VetoNotEvaluated] != 1 {
		t.Errorf("NOT_EVALUATED = %d, want 1 — the unassigned band has to land somewhere",
			tally.NotMeasured[VetoNotEvaluated])
	}
	// Keys are exactly the code's bands even at zero, so a column that is missing
	// cannot read as a column nobody needed.
	if len(tally.Crossed) != len(SeenLatePercentileBands) {
		t.Errorf("crossed keys = %v, want exactly %v", tally.Crossed, SeenLatePercentileBands)
	}
	// 12/100 is the 88th percentile and 140/150 is the 6.67th: one crosses 80, and
	// neither crosses 90.
	if tally.Crossed["80"] != 1 {
		t.Errorf("crossed 80 = %d, want 1", tally.Crossed["80"])
	}
	if tally.Crossed["90"] != 0 {
		t.Errorf("crossed 90 = %d, want 0", tally.Crossed["90"])
	}
}

// TestABandOfAnotherCodeIsNotCountedAsMeasured.
//
// The keys and the measurement are per code, so a seen_late band arriving in an
// extended tally is a wiring defect. It is counted — the invariant above holds for
// every input — and it is counted as unmeasured, because nobody measured *this*
// code for that candidate.
func TestABandOfAnotherCodeIsNotCountedAsMeasured(t *testing.T) {
	tally := TallyBands(VetoExtended, []ShadowBand{MeasureSeenLateBand(aSighting(12, 100, t0))})
	if tally.Measured != 0 {
		t.Errorf("measured = %d, want 0 — a seen_late percentile was counted as an extended gain",
			tally.Measured)
	}
	if tally.Total != 1 {
		t.Errorf("total = %d, want 1", tally.Total)
	}
}

// --- a band can never become a veto ---------------------------------------------

// TestAShadowBandCannotBeReadAsAVeto.
//
// The type carries no affirmative predicate at all: no Dangerous, no Clear, no
// Raised. `if !chase.NearHigh.Dangerous() { score += 2 }` is the one-line mistake
// the §4 review left for section 5, and the band's version of it — treating a
// crossing as a veto and a non-crossing as safety — has no spelling here to make.
//
// Crossed is the only predicate, it needs a band to be named, and an unmeasured
// band answers false to every band there is. So "crossed nothing" and "we never
// looked" are the same answer from Crossed and different answers from Measured,
// which is why a caller has to read Measured.
func TestAShadowBandCannotBeReadAsAVeto(t *testing.T) {
	band := MeasureSeenLateBand(aSighting(12, 100, t0))
	typ := reflect.TypeOf(band)
	for _, forbidden := range []string{"Dangerous", "Clear", "Raised", "Vetoed", "Passed", "State"} {
		if _, ok := typ.MethodByName(forbidden); ok {
			t.Errorf("ShadowBand has a %s method; a band is a measuring instrument and the "+
				"vocabulary of a verdict is what turns one into the other", forbidden)
		}
	}
}

// TestNoFunctionThatProducesAVerdictCanSeeAShadowBand is the structural half.
//
// TestTheVetoCannotSeeAScoreToBeOffsetBy closes the input direction: a band cannot
// be reached from VetoInputs, because the band types are in that test's score set.
// This closes the other one — no function anywhere in this package that returns a
// VetoState or a Chase may so much as mention a band. There is therefore no
// conversion to call, and adding one fails here rather than reading as a reasonable
// diff.
func TestNoFunctionThatProducesAVerdictCanSeeAShadowBand(t *testing.T) {
	bandNames := map[string]bool{
		"ShadowBand": true, "BandCrossing": true, "BandTally": true,
		"MeasureSeenLateBand": true, "MeasureExtendedBand": true, "TallyBands": true,
		"SeenLatePercentileBands": true, "ExtendedGainBands": true, "BandsFor": true,
	}
	verdicts := map[string]bool{"VetoState": true, "Chase": true}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing the package: %v", err)
	}
	checked := 0
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Type.Results == nil {
					continue
				}
				produces := false
				for _, result := range fn.Type.Results.List {
					if ident, ok := result.Type.(*ast.Ident); ok && verdicts[ident.Name] {
						produces = true
					}
				}
				if !produces {
					continue
				}
				checked++
				ast.Inspect(fn, func(n ast.Node) bool {
					ident, ok := n.(*ast.Ident)
					if ok && bandNames[ident.Name] {
						t.Errorf("%s:%s produces a verdict and mentions %s; a shadow band is a "+
							"measurement and a veto is a decision, and D18 says the second one "+
							"waits for a person to approve a number",
							filepath.Base(name), fn.Name.Name, ident.Name)
					}
					return true
				})
			}
		}
	}
	if checked == 0 {
		t.Fatal("no function producing a VetoState or a Chase was found; the check would " +
			"pass vacuously")
	}
	t.Logf("checked %d verdict-producing functions", checked)
}
