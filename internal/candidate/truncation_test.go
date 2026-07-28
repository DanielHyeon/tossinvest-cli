package candidate

// truncation_test.go is section 3: a reading that arrived short is not a list, so
// the position in it is not a percentile.
//
// # The arithmetic
//
// The percentile is (total − rank) ÷ total and `total` is what arrived. A reading
// that asked for a hundred rows and got three therefore puts its top row at the
// 66.7th percentile — mid-list, measured, unremarkable — and its bottom row at 0.
// A one-row reading is the extreme of the same thing: its only row, the top of the
// list, comes out at percentile 0, the very bottom.
//
// # Why the bar is not a minimum row count
//
// The obvious repair is "refuse any reading below N rows". N would be a number
// nobody derived, and it would also be wrong in both directions: the WTS popularity
// list is thirty rows and complete at thirty, while a hundred-row request that
// returned forty is broken at forty. Length does not decide it. What decides it is
// whether the source got what it asked for, and both of those numbers are the
// source's own.
//
// candidatesrc.go has carried the warning since it was written — "a caller that
// asked for 150 and got 100 would compute rank percentiles against a list length
// that never existed" — one field away from where the defect was.

import "testing"

// TestTheZeroTruncationFactIsUnknownRatherThanWhole is the same zero-value rule the
// new-entrant fact is held to. A reading nobody recorded a request for has not been
// found complete.
func TestTheZeroTruncationFactIsUnknownRatherThanWhole(t *testing.T) {
	var zero Truncation
	if zero.Known() || zero.Yes() || zero.No() {
		t.Errorf("the zero Truncation answers known=%v yes=%v no=%v; a reading nobody "+
			"recorded a request for has not been measured as whole",
			zero.Known(), zero.Yes(), zero.No())
	}
	if zero != TruncationUnknown() || zero.String() != "unknown" {
		t.Errorf("the zero Truncation is %q and TruncationUnknown() is %q; they have to be "+
			"the same value", zero, TruncationUnknown())
	}
	if fields := exportedFieldNames(Truncation{}); len(fields) != 0 {
		t.Errorf("Truncation exports %v; a composite literal outside this package could then "+
			"spell a whole reading nobody measured", fields)
	}
}

// TestTruncationComparesTheTwoNumbersTheSourceDeclared.
func TestTruncationComparesTheTwoNumbersTheSourceDeclared(t *testing.T) {
	for _, tc := range []struct {
		name               string
		requested, arrived int
		want               Truncation
	}{
		{"a full official ranking", 100, 100, WholeReading()},
		{"a full WTS popularity list, which is thirty rows and complete", 30, 30, WholeReading()},
		{"a hundred requested, three arrived", 100, 3, TruncatedReading()},
		{"a hundred requested, one arrived", 100, 1, TruncatedReading()},
		{"a hundred requested, none arrived", 100, 0, TruncatedReading()},
		// More than asked for is also a list nobody can name the length of, and it is
		// a contract or wiring fault rather than a market event.
		{"more arrived than were asked for", 30, 31, TruncatedReading()},
		// Nothing recorded. Not "asked for nothing" — that is not a reading anything
		// could produce, and reading it that way would mark the whole stored history
		// truncated on a comparison nobody made.
		{"no request recorded", 0, 100, TruncationUnknown()},
		{"a negative request, which validate refuses at the boundary", -1, 100, TruncationUnknown()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := truncationOf(tc.requested, tc.arrived); got != tc.want {
				t.Errorf("truncationOf(%d, %d) = %s, want %s",
					tc.requested, tc.arrived, got, tc.want)
			}
		})
	}
}

// TestATruncatedReadingsPositionIsNotAPercentile is task 3.2's RED.
//
// Three of a hundred requested. The top row would read 66.7 — which is not a
// wrong-looking number, and that is the point: it is the middle of a list, so
// seen_late would report the symbol as one we caught early.
func TestATruncatedReadingsPositionIsNotAPercentile(t *testing.T) {
	c := aCandidate(t0)
	first := storedFirstRank(1, 3, t0, SourceOfficialTradingValue)
	first.Requested = 100

	got := MeasureFirstSighting(c, first, nil)
	if got.Measured {
		t.Fatalf("the top row of a reading that asked for 100 and received 3 was measured "+
			"at %s%%; that is a percentile of a list that never existed", got.PercentilePct)
	}
	if got.Reason() != VetoReadingTruncated {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoReadingTruncated)
	}
	if !got.Truncation.Yes() {
		t.Errorf("the refused sighting's truncation fact = %s, want truncated", got.Truncation)
	}
	if got.Rank != 1 || got.RankTotal != 3 {
		t.Errorf("the refused sighting reports %d of %d; it still has to show what it read",
			got.Rank, got.RankTotal)
	}
	if state := AssessSeenLate(got, seenLateThresholds()); state.Clear() || state.Dangerous() {
		t.Errorf("seen_late = %v on a truncated reading; unmeasured is not a pass (D10)", state)
	}
}

// TestTheOneRowReadingIsCaughtByTheSameRefusal is task 3.3.
//
// percentileOf(1, 1) is 0: the only row of a one-row reading reads as the bottom of
// its own list, which clears every seen_late threshold there is. This is not given
// a rule of its own, and that is the point of D4 — a separate minimum row count
// would be a number nobody derived, and this case is already covered by the two
// numbers the source declared.
func TestTheOneRowReadingIsCaughtByTheSameRefusal(t *testing.T) {
	// The arithmetic that makes the case dangerous, stated so that the refusal below
	// is visibly about something rather than about nothing.
	if got := formatDecimal(percentileOf(1, 1)); got != "0" {
		t.Fatalf("percentileOf(1, 1) = %s, want 0 — if this is no longer the bottom of the "+
			"list then the case this test exists for has changed shape", got)
	}

	c := aCandidate(t0)
	first := storedFirstRank(1, 1, t0, SourceOfficialTradingValue)
	first.Requested = 100

	got := MeasureFirstSighting(c, first, nil)
	if got.Measured {
		t.Fatalf("the only row of a one-row reading was measured at %s%%, which is the "+
			"bottom of the list for a symbol that was the whole of it", got.PercentilePct)
	}
	if got.Reason() != VetoReadingTruncated {
		t.Errorf("reason = %q, want %q — and it is the truncation reason rather than a "+
			"minimum-length one, because no minimum was invented", got.Reason(), VetoReadingTruncated)
	}
}

// TestAShortListThatIsNotTruncatedIsStillMeasurable is the other side of the same
// rule, and it is why the refusal is not a length bar.
//
// The WTS popularity list is thirty rows. Thirty of thirty requested is a complete
// reading of a short list, and its positions are perfectly good percentiles. A
// minimum row count set anywhere above thirty would have thrown away one of the two
// ranking sources entirely, silently, under a reason that said "truncated".
func TestAShortListThatIsNotTruncatedIsStillMeasurable(t *testing.T) {
	c := aCandidate(t0)
	first := storedFirstRank(3, 30, t0, SourceWTSPopular)
	first.Requested = 30

	got := MeasureFirstSighting(c, first, nil)
	if !got.Measured {
		t.Fatalf("a complete thirty-row reading was refused as %s; thirty is the whole of "+
			"that list and the source got exactly what it asked for", got.Reason())
	}
	if !got.Truncation.No() {
		t.Errorf("truncation fact = %s, want whole", got.Truncation)
	}
	if got.PercentilePct != "90" {
		t.Errorf("percentile = %q, want 90 (3rd of 30)", got.PercentilePct)
	}
}

// TestAPositionWithNoRecordedRequestIsRefusedByTheOtherRuleRatherThanThisOne.
//
// Truncation is unknown for every row written before schema 4, and unknown is not
// truncated — nothing is concluded from it. Those rows are refused anyway, under the
// new-entrant reason, because that fact is unknown for them too. The distinction
// matters for what an operator is told: "this reading came back short" and "nobody
// recorded what this reading asked for" are different problems.
func TestAPositionWithNoRecordedRequestIsRefusedByTheOtherRuleRatherThanThisOne(t *testing.T) {
	c := aCandidate(t0)
	first := storedFirstRank(4, 100, t0, SourceOfficialTradingValue)
	first.Requested = 0

	got := MeasureFirstSighting(c, first, nil)
	if got.Truncation.Known() {
		t.Errorf("a position with no recorded request reports truncation %s; nothing was "+
			"measured about it", got.Truncation)
	}
	if !got.Measured {
		t.Errorf("the position was refused as %s; an unknown request is not a truncated "+
			"reading and must not be treated as one", got.Reason())
	}
}

// TestTheTruncationRefusalNamesNoMinimumRowCount is D4's rule stated against the
// source, for the reason task 3.3 gives: the repair that looks identical from the
// outside is a floor, and a floor is a number nobody derived.
func TestTheTruncationRefusalNamesNoMinimumRowCount(t *testing.T) {
	// truncationOf is the whole of the decision. If it ever compares `arrived`
	// against anything other than `requested`, a floor has arrived.
	for _, arrived := range []int{1, 2, 3, 5, 8, 10, 29, 30, 99, 100} {
		if got := truncationOf(arrived, arrived); !got.No() {
			t.Errorf("a complete reading of %d rows was called %s; the length of the list is "+
				"not what decides this", arrived, got)
		}
	}
}
