package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// candidatebands_test.go is the change refine-extended-shadow-bands, terminal half.
//
// The extended scale went from five edges to ten and the row that renders it is one
// `·`-joined line. Doubling the number of columns on a row nobody had measured is
// the kind of change that is invisible in a unit test and obvious the first time
// somebody reads it in an 80-column terminal, so the width is measured here.

// aBandTallyOfTen is a tally shaped like the session that produced this change: 131
// measured of 156, three-digit counts at the low edges, zeros at the top.
func aBandTallyOfTen() candidate.BandTally {
	counts := map[string]int{
		"0": 118, "2": 44, "4": 21, "6": 12, "8": 7,
		"10": 0, "20": 0, "30": 0, "50": 0, "100": 0,
	}
	tally := candidate.BandTally{
		Code:        candidate.VetoExtended,
		Total:       156,
		Measured:    131,
		Crossed:     map[string]int{},
		NotMeasured: map[candidate.VetoUnmeasured]int{candidate.VetoNoObservations: 25},
		Quantiles: []candidate.BandQuantile{
			{Name: "min", Value: "-12.35"}, {Name: "p25", Value: "-0.42"},
			{Name: "median", Value: "0.91"}, {Name: "p75", Value: "2.68"},
			{Name: "max", Value: "41.07"},
		},
		QuantileBase: 131,
	}
	for _, band := range candidate.BandsFor(candidate.VetoExtended) {
		tally.Crossed[band] = counts[band]
	}
	return tally
}

// renderBandSection is the report's shadow-band block for one tally.
func renderBandSection(t *testing.T, tally candidate.BandTally) string {
	t.Helper()
	res := candidate.CycleResult{
		Market: "KR",
		At:     time.Date(2026, 7, 29, 4, 30, 0, 0, time.UTC),
		Bands:  map[candidate.VetoCode]candidate.BandTally{candidate.VetoExtended: tally},
	}
	var buf bytes.Buffer
	if err := writeCandidateTable(&buf, res, buildCandidateReport(res)); err != nil {
		t.Fatalf("rendering the report: %v", err)
	}
	out := buf.String()
	start := strings.Index(out, "shadow bands — extended")
	if start < 0 {
		t.Fatalf("the report has no extended band section:\n%s", out)
	}
	section := out[start:]
	if end := strings.Index(section, "\nretention"); end >= 0 {
		section = section[:end]
	}
	return section
}

// TestTheShadowBandRowsStayInsideEightyColumns is task 2.6.
//
// D7 predicted this row would be fine because the renderer is driven by BandsFor and
// makes no width assumption. The renderer makes none; the terminal does. Ten edges
// with three-digit counts is 83 columns and the live tally measured 92, and a row
// the terminal folds wraps to column 0 — under the section labels, where it reads as
// a new field rather than as the rest of this one.
//
// So the report folds it itself, at reportWidth, with the continuation indented to
// the label. This test is the measurement D7 asked for, kept.
func TestTheShadowBandRowsStayInsideEightyColumns(t *testing.T) {
	section := renderBandSection(t, aBandTallyOfTen())
	t.Logf("the rendered section (task 2.6):\n%s", section)

	// The list rows and their continuations, which are the rows this change widened.
	// Prose rows are excluded on purpose and the exclusion is not an oversight: the
	// section heading has been 112 columns since it was written and the collapse
	// alarm is a sentence in the style of tallyAlarm's. Prose soft-wraps at a word
	// boundary and stays readable; a `·`-joined list of counts wrapped by the
	// terminal restarts at column 0, under the section labels, and reads as a new
	// field. Both widths are recorded in the change's issues.md.
	folded, continuation := 0, false
	for _, line := range strings.Split(section, "\n") {
		isList := strings.Contains(line, " · ") || strings.HasSuffix(line, " ·") || continuation
		continuation = strings.HasSuffix(line, " ·")
		if !isList {
			continue
		}
		if width := utf8.RuneCountInString(line); width > reportWidth {
			t.Errorf("a count row is %d columns wide, past %d:\n%s", width, reportWidth, line)
		}
		if strings.HasSuffix(line, " ·") {
			folded++
		}
	}
	// Without this the test would also pass if the row silently lost its columns.
	for _, band := range candidate.BandsFor(candidate.VetoExtended) {
		if !strings.Contains(section, band+" ") {
			t.Errorf("band %s is not in the rendered section; folding dropped a column:\n%s",
				band, section)
		}
	}
	if folded == 0 {
		t.Errorf("nothing folded in a section built to be too wide for one line; the fold is "+
			"not being exercised and the width above proves nothing:\n%s", section)
	}
}

// TestWrapCountsKeepsEveryPartAndStaysInsideTheWidth is the fold itself.
//
// The section test above proves the fold happens for one realistic tally. This
// proves the properties that make it safe for any tally: nothing is dropped, nothing
// is reordered, continuations line up under the first item rather than at column 0,
// and a folded line is marked as folded.
func TestWrapCountsKeepsEveryPartAndStaysInsideTheWidth(t *testing.T) {
	for _, tc := range []struct {
		name  string
		label string
		parts []string
	}{
		{"nothing", "  crossed      ", nil},
		{"one", "  crossed      ", []string{"0 1"}},
		{"the five-edge scale", "  crossed      ",
			[]string{"10 0", "20 0", "30 0", "50 0", "100 0"}},
		{"the ten-edge scale at three digits", "  crossed      ", []string{
			"0 118", "2 44", "4 21", "6 12", "8 7", "10 0", "20 0", "30 0", "50 0", "100 0"}},
		{"counts wide enough to fold twice", "  crossed      ", []string{
			"0 118000", "2 440000", "4 210000", "6 120000", "8 700000",
			"10 100000", "20 200000", "30 300000", "50 500000", "100 900000"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines := wrapCounts(tc.label, tc.parts)
			indent := strings.Repeat(" ", utf8.RuneCountInString(tc.label))
			for i, line := range lines {
				if width := utf8.RuneCountInString(line); width > reportWidth {
					t.Errorf("line %d is %d columns: %q", i, width, line)
				}
				want := tc.label
				if i > 0 {
					want = indent
				}
				if !strings.HasPrefix(line, want) {
					t.Errorf("line %d does not begin at the label's column: %q", i, line)
				}
				if i < len(lines)-1 && !strings.HasSuffix(line, " ·") {
					t.Errorf("line %d is continued and does not say so: %q", i, line)
				}
			}
			if len(tc.parts) == 0 {
				if len(lines) != 1 || !strings.HasSuffix(lines[0], "none") {
					t.Errorf("an empty row rendered as %q, want the label and \"none\"", lines)
				}
				return
			}
			// Every part, once, in order — a fold that loses or reorders a column is
			// the failure this whole section is about, one level down.
			var got []string
			for i, line := range lines {
				body := strings.TrimPrefix(line, tc.label)
				if i > 0 {
					body = strings.TrimPrefix(line, indent)
				}
				got = append(got, strings.Split(strings.TrimSuffix(body, " ·"), " · ")...)
			}
			if fmt.Sprint(got) != fmt.Sprint(tc.parts) {
				t.Errorf("the folded row reads %v, want %v (from %q)", got, tc.parts, lines)
			}
		})
	}
}

// TestTheReportShowsWhereTheValuesWereAndNotOnlyHowManyCrossed is task 2.7 on the
// terminal.
//
// The crossings are counted against edges. The session that produced this change
// reported `measured 131 of 156, crossed … all 0`, which is the row a distribution
// sitting at −1 and one sitting at +9 would both produce, and ShadowBand.Value —
// computed for every candidate — appeared on no surface at all.
func TestTheReportShowsWhereTheValuesWereAndNotOnlyHowManyCrossed(t *testing.T) {
	section := renderBandSection(t, aBandTallyOfTen())
	for _, want := range []string{"min -12.35", "median 0.91", "max 41.07"} {
		if !strings.Contains(section, want) {
			t.Errorf("the report does not show %q; without the distribution the next spacing "+
				"decision is another guess:\n%s", want, section)
		}
	}

	// A tally that measured nothing says so rather than printing a zero. 0 is the
	// one value that reads as an answer.
	empty := candidate.TallyBands(candidate.VetoExtended, []candidate.ShadowBand{{}})
	blank := renderBandSection(t, empty)
	if !strings.Contains(blank, "values       none") {
		t.Errorf("an unmeasured tally does not say its distribution is empty:\n%s", blank)
	}
}

// TestTheReportRaisesTheAlarmForAScaleThatResolvedNothing is the spec's alarm
// scenario on this surface.
//
// The judgement is candidate.BandTally.Collapsed and is tested there. What is tested
// here is that it reaches the screen, because the state it names was invisible for a
// month with the arithmetic intact.
func TestTheReportRaisesTheAlarmForAScaleThatResolvedNothing(t *testing.T) {
	collapsed := aBandTallyOfTen()
	for _, band := range candidate.BandsFor(candidate.VetoExtended) {
		collapsed.Crossed[band] = 0
	}
	section := renderBandSection(t, collapsed)
	t.Logf("the collapsed rendering (task 2.6):\n%s", section)
	if !strings.Contains(section, "ALARM") {
		t.Errorf("131 measured records that all produced the same crossings render as an "+
			"ordinary row:\n%s", section)
	}
	// Above the counts: they are what a reader takes away and they looked fine.
	if strings.Index(section, "ALARM") > strings.Index(section, "crossed") {
		t.Errorf("the alarm is printed below the counts it is about:\n%s", section)
	}

	if plain := renderBandSection(t, aBandTallyOfTen()); strings.Contains(plain, "ALARM") {
		t.Errorf("a tally whose bands separate its records raises the alarm:\n%s", plain)
	}
}

// TestABandRowNobodyCrossedStillOccupiesItsColumn.
//
// Existing behaviour, pinned because the fold is new: a wrapped row must not become
// a row that omits its zeros. A missing column reads as a column nobody needed.
func TestABandRowNobodyCrossedStillOccupiesItsColumn(t *testing.T) {
	tally := aBandTallyOfTen()
	for band := range tally.Crossed {
		tally.Crossed[band] = 0
	}
	section := renderBandSection(t, tally)
	for _, band := range candidate.BandsFor(candidate.VetoExtended) {
		if !strings.Contains(section, band+" 0") {
			t.Errorf("band %s has no column in a tally where nobody crossed anything:\n%s",
				band, section)
		}
	}
}
