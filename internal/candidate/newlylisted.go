package candidate

// newlylisted.go holds the two facts a ranking source knows about its own reading
// and used to throw away: whether a symbol was in the source's previous reading,
// and whether the reading arrived whole.
//
// # Why they are types rather than a bool and an int
//
// `NewlyListed bool` is what this package shipped. It was declared on Row, copied
// through Reported, persisted, read back, carried on Sighting and rendered on
// /signals — five layers — and no production source ever assigned it, so the fact
// was structurally false everywhere and the screen's new-entrant marker could not
// appear. Nothing failed, which is the whole point: a bool has no spelling for "we
// did not measure this", so an unmeasured fact and a measured negative are the same
// value and the difference cannot be noticed.
//
// That is the shape this package has already refused four times — VetoState,
// Sighting, Expansion and ShadowBand each grew a three-state answer for the same
// reason — and the rule those four settled is that **unmeasured has to be the zero
// value**. It is what a struct literal that omits the field produces, what a map
// miss produces and what a NULL column scans into, so the accident produces the
// honest answer instead of the flattering one.
//
// The fields are unexported for VetoState's reason: a composite literal in another
// package — a fixture, a rendering, a future lane — must not be able to spell "the
// source said no" without a source having said it.
//
// # Why the first reading of a source is unknown and not "yes"
//
// A source with no previous reading is not looking at a list every symbol just
// joined; it is looking at a list for the first time. "Absent from the previous
// reading" is vacuously true of every row and means nothing about the market — it
// means the process just started. Recording that as `yes` would make the first scan
// after every restart declare the whole panel newly listed, and seen_late would
// then be measured against a fact about us.

// NewlyListed is a source's three-state report that a symbol was absent from that
// source's previous reading of the same list.
//
// It is recorded as its own fact and never folded into a rank velocity by
// pretending the symbol was previously last (D8, correction 1).
type NewlyListed struct {
	known bool
	newly bool
}

// NewlyListedUnknown is the state of a fact nobody measured, and it is the zero
// value. The constructor exists so a caller can say so out loud; a caller who says
// nothing produces the same value.
func NewlyListedUnknown() NewlyListed { return NewlyListed{} }

// NewlyListedYes: the source held a previous reading and this symbol was not in it.
func NewlyListedYes() NewlyListed { return NewlyListed{known: true, newly: true} }

// NewlyListedNo: the source held a previous reading and this symbol was in it.
func NewlyListedNo() NewlyListed { return NewlyListed{known: true} }

// Known reports that the source had an answer at all.
func (n NewlyListed) Known() bool { return n.known }

// Yes reports that the symbol was measured as absent from the previous reading.
// It requires the measurement, so an unassigned value answers false to this and to
// No alike — neither of them alone can be read as the other's negation.
func (n NewlyListed) Yes() bool { return n.known && n.newly }

// No reports that the symbol was measured as present in the previous reading.
func (n NewlyListed) No() bool { return n.known && !n.newly }

// String is what a screen and a stored row spell. An unknown state names itself.
func (n NewlyListed) String() string {
	switch {
	case n.Yes():
		return "yes"
	case n.No():
		return "no"
	default:
		return "unknown"
	}
}

// newlyListedFromStore reads the stored spelling back. Anything that is not one of
// the two measured words is unknown, including NULL and including a value written
// by a build that spelled it differently — an unrecognised state is unmeasured, the
// same rule Chase.State applies to an unrecognised veto code.
func newlyListedFromStore(text string, valid bool) NewlyListed {
	if !valid {
		return NewlyListedUnknown()
	}
	switch text {
	case "yes":
		return NewlyListedYes()
	case "no":
		return NewlyListedNo()
	default:
		return NewlyListedUnknown()
	}
}

// newlyListedToStore maps the fact onto a nullable TEXT column: NULL for unknown,
// so a row written before this fact existed reads back as unmeasured rather than as
// a negative somebody recorded.
func newlyListedToStore(n NewlyListed) any {
	if !n.Known() {
		return nil
	}
	return n.String()
}

// Truncation is a source's three-state report that a reading arrived short.
//
// A ranking source knows how many rows it asked for — officialRanking.count is 100
// and wtsPopular.size is 30 — and used to discard that number, so RankTotal was the
// count that arrived and a reading of three rows out of a hundred requested was
// indistinguishable from a three-row list. The percentile is (total − rank) ÷ total,
// so the top of that reading came out at 66.7: a normalisation against a list length
// that never existed. candidatesrc.go's own comment warned about exactly this shape
// — "a caller that asked for 150 and got 100 would compute rank percentiles against
// a list length that never existed" — one field over from where it happened.
//
// Unknown is the zero value for NewlyListed's reason, and it is what every row
// written before this fact existed carries.
type Truncation struct {
	known     bool
	truncated bool
}

// TruncationUnknown is the zero value: nobody recorded how many rows were asked for.
func TruncationUnknown() Truncation { return Truncation{} }

// TruncatedReading: the source asked for more rows than arrived.
func TruncatedReading() Truncation { return Truncation{known: true, truncated: true} }

// WholeReading: the source asked for exactly as many rows as arrived.
func WholeReading() Truncation { return Truncation{known: true} }

// Known reports that the requested row count was recorded beside the arrived one.
func (t Truncation) Known() bool { return t.known }

// Yes reports that the reading was measured as short. It requires the measurement.
func (t Truncation) Yes() bool { return t.known && t.truncated }

// No reports that the reading was measured as whole.
func (t Truncation) No() bool { return t.known && !t.truncated }

// String names the state.
func (t Truncation) String() string {
	switch {
	case t.Yes():
		return "truncated"
	case t.No():
		return "whole"
	default:
		return "unknown"
	}
}

// truncationOf compares what a source asked for with what arrived.
//
// A non-positive `requested` is unknown rather than "asked for nothing": zero is
// what a row written before this fact existed carries and what a source that does
// not declare a size reports, and reading it as a request for no rows would make
// every such reading truncated — a refusal that would then cover the whole stored
// history for a reason nobody measured.
//
// The comparison is inequality rather than `arrived < requested`. A source that
// returned MORE rows than it asked for has also produced a list nobody can name the
// length of, and it is a wiring or contract fault rather than a market event.
func truncationOf(requested, arrived int) Truncation {
	if requested <= 0 {
		return TruncationUnknown()
	}
	if requested != arrived {
		return TruncatedReading()
	}
	return WholeReading()
}
