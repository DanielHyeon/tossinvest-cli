package candidate

// sightingsource_skip_test.go drives TallySightingSources' one refusal directly.
//
// The reducer exists to say *whose* readings the refusals came from, so the one
// thing it must not do is invent an attribution. A verdict whose sighting carries no
// source is skipped rather than filed under the empty id — a row under "" is read as
// a source, and a screen that lists it sends an operator looking for an endpoint
// nobody named.
//
// It was unpinned: the wired tests all go through Assess, which fills the source from
// a stored position, so no fixture ever produced a sourceless sighting and deleting
// the skip left the suite green. Every candidate whose first rank was never recorded
// has one — that is the session-start state this change deliberately creates more of.

import "testing"

// sightingFrom is one verdict's sighting, as the reducer sees it.
func sightingFrom(id SourceID, measured bool, why VetoUnmeasured) Verdict {
	return Verdict{Sighting: Sighting{Source: id, Measured: measured, Why: why}}
}

// TestASightingWithNoSourceIsNotASourceCalledEmpty.
func TestASightingWithNoSourceIsNotASourceCalledEmpty(t *testing.T) {
	got := TallySightingSources([]Verdict{
		sightingFrom(SourceOfficialTradingValue, true, ""),
		// No stored position, so nothing dated it and nothing named who read it. This
		// is what every candidate looks like between the tick that promoted it and
		// the tick that qualifies its first sighting.
		sightingFrom("", false, VetoNoFirstRank),
		sightingFrom("", false, VetoNoFirstRank),
	})
	if len(got) != 1 {
		t.Fatalf("%d sources in the reduction, want 1; only one of these three verdicts says "+
			"who read it: %+v", len(got), got)
	}
	if got[0].Source != SourceOfficialTradingValue {
		t.Fatalf("the one entry is attributed to %q", got[0].Source)
	}
	if got[0].Total != 1 || got[0].Measured != 1 {
		t.Errorf("the entry = %d total and %d measured, want 1 and 1. The two sourceless "+
			"verdicts were counted against a source that did not produce them",
			got[0].Total, got[0].Measured)
	}
}

// TestAnUnmeasuredSightingWithNoReasonIsStillCountedUnderOne is the neighbouring
// guarantee, and it is why the skip above is the only one.
//
// A sighting that names a source is counted whatever else it is missing: an
// unmeasured one with no reason at all lands under NOT_EVALUATED rather than under
// an empty key, so the census adds up to Total for every source in it.
func TestAnUnmeasuredSightingWithNoReasonIsStillCountedUnderOne(t *testing.T) {
	got := TallySightingSources([]Verdict{
		sightingFrom(SourceWTSPopular, false, ""),
		sightingFrom(SourceWTSPopular, false, VetoReadingTruncated),
	})
	if len(got) != 1 {
		t.Fatalf("%d sources in the reduction, want 1", len(got))
	}
	entry := got[0]
	if entry.Total != 2 || entry.Measured != 0 {
		t.Fatalf("the entry = %d total and %d measured, want 2 and 0", entry.Total, entry.Measured)
	}
	if entry.NotMeasured[VetoNotEvaluated] != 1 {
		t.Errorf("the reasonless refusal was filed under %v; an unmeasured sighting counted "+
			"with no bucket is exactly the unattributable count this reducer removes one level "+
			"up", entry.NotMeasured)
	}
	sum := 0
	for _, n := range entry.NotMeasured {
		sum += n
	}
	if entry.Measured+sum != entry.Total {
		t.Errorf("measured %d + refusals %d != total %d; the census has to account for every "+
			"sighting it counted", entry.Measured, sum, entry.Total)
	}
}
