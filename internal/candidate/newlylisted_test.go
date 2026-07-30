package candidate

// newlylisted_test.go is task 1.1: the new-entrant fact is three-state and
// `unknown` is its zero value.
//
// The test that matters here is the one about the value nobody assigned. A bool
// would make that value read as "the source told us this symbol was already in
// its previous reading", which is a measurement nobody made — and it is the shape
// this package has refused four times already (VetoState, Sighting, Expansion,
// ShadowBand). Every one of those refusals was written down; this one is written
// down as a test because the field it replaces spent five layers being exactly
// that bool and nothing ever failed.

import (
	"reflect"
	"testing"
)

// exportedFieldNames is every field of a struct that another package could name in
// a composite literal.
func exportedFieldNames(v any) []string {
	typ := reflect.TypeOf(v)
	var out []string
	for i := 0; i < typ.NumField(); i++ {
		if f := typ.Field(i); f.IsExported() {
			out = append(out, f.Name)
		}
	}
	return out
}

// TestTheUnassignedNewEntrantFactIsUnknownRatherThanAlreadyListed is the RED.
//
// `var n NewlyListed` is what a struct literal that omits the field produces, what
// a map miss produces and what a scan of a NULL column produces. If it answers No
// — "the source reported this symbol was already there" — then the three states
// have been folded back into two and the fold is invisible.
func TestTheUnassignedNewEntrantFactIsUnknownRatherThanAlreadyListed(t *testing.T) {
	var zero NewlyListed
	if zero.Known() {
		t.Error("the zero NewlyListed reports itself as known; a fact nobody measured " +
			"cannot be a fact somebody reported")
	}
	if zero.Yes() || zero.No() {
		t.Errorf("the zero NewlyListed answers yes=%v no=%v; both affirmative predicates "+
			"must require the measurement, exactly as VetoState.Clear does",
			zero.Yes(), zero.No())
	}
	if zero != NewlyListedUnknown() {
		t.Error("NewlyListedUnknown() is not the zero value, so a field nobody assigned and " +
			"a field explicitly set to unknown are two different values")
	}
	if got := zero.String(); got != "unknown" {
		t.Errorf("the zero NewlyListed prints %q, want %q — an unnamed state is one an "+
			"operator learns to skip", got, "unknown")
	}
}

// TestTheThreeNewEntrantStatesAreDistinguishable pins the two measured answers
// apart from each other and from the unknown one.
func TestTheThreeNewEntrantStatesAreDistinguishable(t *testing.T) {
	for _, tc := range []struct {
		name           string
		got            NewlyListed
		known, yes, no bool
		text           string
	}{
		{"unknown", NewlyListedUnknown(), false, false, false, "unknown"},
		{"yes", NewlyListedYes(), true, true, false, "yes"},
		{"no", NewlyListedNo(), true, false, true, "no"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.Known() != tc.known || tc.got.Yes() != tc.yes || tc.got.No() != tc.no {
				t.Errorf("known=%v yes=%v no=%v, want %v %v %v",
					tc.got.Known(), tc.got.Yes(), tc.got.No(), tc.known, tc.yes, tc.no)
			}
			if tc.got.String() != tc.text {
				t.Errorf("String() = %q, want %q", tc.got.String(), tc.text)
			}
		})
	}
}

// TestTheNewEntrantFactCannotBeBuiltFromOutsideThisPackage.
//
// The fields are unexported for VetoState's reason: a composite literal in another
// package — a fixture, a §5 rendering, a future lane — must not be able to
// manufacture a "the source said no" that no source said. This asserts the shape
// rather than the call site, because the call site that would do it does not exist
// yet and that is precisely when the guard has to be written.
func TestTheNewEntrantFactCannotBeBuiltFromOutsideThisPackage(t *testing.T) {
	if fields := exportedFieldNames(NewlyListed{}); len(fields) != 0 {
		t.Errorf("NewlyListed exports %v; an exported field is a way to spell a measured "+
			"answer without measuring one", fields)
	}
}

// TestARowThatNobodyToldAboutThePreviousReadingIsUnknown is the same rule at the
// three layers the fact travels through.
//
// Row is what a source returns, Reported is what the store holds, and Sighting is
// what the veto reads. A literal omitting the field at any of the three used to
// mean "already listed", which is how the fact managed to be declared in five
// places and assigned in none.
func TestARowThatNobodyToldAboutThePreviousReadingIsUnknown(t *testing.T) {
	if (Row{Symbol: "005930", Rank: 1, RankTotal: 100}).NewlyListed.Known() {
		t.Error("a Row built without the fact claims to carry it")
	}
	if (Reported{Rank: 1, RankTotal: 100}).NewlyListed.Known() {
		t.Error("a Reported built without the fact claims to carry it")
	}
	if (Sighting{Measured: true, Rank: 1, RankTotal: 100}).NewlyListed.Known() {
		t.Error("a Sighting built without the fact claims to carry it")
	}
	if (FirstRank{Rank: 1, Total: 100}).NewlyListed.Known() {
		t.Error("a FirstRank built without the fact claims to carry it")
	}
}
