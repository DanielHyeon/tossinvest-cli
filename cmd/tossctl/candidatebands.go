package main

// candidatebands.go is the shadow-band half of the discovery scan report.
//
// It is a separate file from candidate.go for two reasons, and the second one is
// the one that matters.
//
// The first is size: candidate.go carries the commands, the whole report struct and
// every other rendering, and the band rows are a self-contained cluster with their
// own rule about ordering.
//
// The second is that these are new functions rather than edits to old ones, and the
// Function Logic Map gate asks for evidence about existing functions a diff hunk
// touches. A new helper wedged between two old ones drags them in by adjacency, and
// the evidence then describes functions nobody changed. The change's tasks state the
// same placement rule for new tests. What stayed in candidate.go is what this change
// actually modified: buildCandidateReport, writeCandidateTable and orderedCounts.
//
// # What the band rows have to get right
//
// The scale is ordered and numeric, and both surfaces lost that in their own way.
// JSON emitted `crossed` as an object, so encoding/json sorted ten numeric keys as
// strings — `"0","10","100","2",…`, with 100 in third place — on the one surface a
// person deriving a threshold actually pipes somewhere. The terminal joined ten
// counts onto one row of 83 columns and let the terminal fold it, which restarts at
// column 0 underneath the section labels and reads as a new field.

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// bandCount is one scale edge and how many measured records reached it.
//
// A list and not a map, because encoding/json sorts map keys as strings and the
// scale is numeric. With five edges the two orders happened to agree; with ten they
// do not, and the map form emits `{"0":118,"10":0,"100":0,"2":44,…}` — 100 in third
// place. That is the surface somebody piping this into a distribution actually
// reads, which is to say it is the surface the whole shadow record exists for. The
// console has rendered these as an ordered list since it was written
// (signalsBandTalliesFrom); this is the terminal's JSON catching up to it.
type bandCount struct {
	Band  string `json:"band"`
	Count int    `json:"count"`
}

// bandQuantile is one position in the distribution of the measured values.
type bandQuantile struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// shadowBandReport is one veto code's shadow record as this surface renders it.
type shadowBandReport struct {
	Note     string `json:"note"`
	Total    int    `json:"total"`
	Measured int    `json:"measured"`
	// Crossed is per edge in scale order; see bandCount.
	Crossed []bandCount `json:"crossed"`
	// Quantiles are where the values themselves fell. The crossings answer "how
	// many were above each edge" and are useless when every answer agrees, which is
	// the state that produced this change: 131 measured, every crossing zero, and
	// no way to tell a distribution sitting at +9 from one sitting at −1.
	Quantiles    []bandQuantile `json:"quantiles"`
	QuantileBase int            `json:"quantile_base"`
	// Alarm is set when every measured record produced the same crossings.
	//
	// spec candidate-discovery: a record whose measured values all fall in one
	// bucket meets the obligation to record and answers nothing, and that state must
	// not be shown as an ordinary number. It was found by a person reading
	// `crossed 10 0 · 20 0 · 30 0 · 50 0 · 100 0` twice.
	Alarm       string         `json:"alarm,omitempty"`
	NotMeasured map[string]int `json:"not_measured"`
}

// orderedCountParts is orderedCounts before it is joined, for a row that may fold.
func orderedCountParts(keys []string, counts map[string]int) []string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, counts[k]))
	}
	return parts
}

// bandCountParts is the same for a shadow band's already-ordered list.
func bandCountParts(in []bandCount) []string {
	parts := make([]string, 0, len(in))
	for _, c := range in {
		parts = append(parts, fmt.Sprintf("%s %d", c.Band, c.Count))
	}
	return parts
}

// quantileParts renders the distribution's positions.
func quantileParts(in []bandQuantile) []string {
	parts := make([]string, 0, len(in))
	for _, q := range in {
		parts = append(parts, fmt.Sprintf("%s %s", q.Name, q.Value))
	}
	return parts
}

// reportWidth is the column this report folds its count rows at.
//
// Eighty, the width a terminal is still assumed to have when nothing says
// otherwise. It became load-bearing when the extended scale went from five edges to
// ten: the row measured 83 columns against a fixture and 92 against the live tally,
// and a row folded by the terminal wraps to column 0, which puts a continuation
// under the section labels and reads as a new field rather than as more of this one.
const reportWidth = 80

// wrapCounts renders `parts` after `label`, folding at reportWidth with the
// continuation lines indented to the label's width.
//
// A folded line ends in ` ·` so that "the list continues" is visible without
// counting columns — the same reason the separator is there in the first place.
// That marker is two columns and they are budgeted here rather than added
// afterwards: a line filled exactly to the width and then given a marker is a line
// two columns past it, which is the whole failure this function exists to prevent,
// reintroduced by the fix for it.
func wrapCounts(label string, parts []string) []string {
	const sep = " · "
	const mark = " ·"
	if len(parts) == 0 {
		return []string{label + "none"}
	}
	indent := strings.Repeat(" ", utf8.RuneCountInString(label))
	var lines []string
	line := label + parts[0]
	for i, part := range parts[1:] {
		// Room for the marker unless this part is the last one there is; if it is,
		// this line will never be folded and needs no room for one.
		room := reportWidth - len(mark)
		if i == len(parts)-2 {
			room = reportWidth
		}
		if utf8.RuneCountInString(line)+len(sep)+utf8.RuneCountInString(part) <= room {
			line += sep + part
			continue
		}
		lines = append(lines, line+mark)
		line = indent + part
	}
	return append(lines, line)
}
