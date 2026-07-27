package position

// table.go is the transition table as data.
//
// It is written as a literal rather than as a switch on purpose. A switch is a
// program that answers questions about transitions; a table is the set of
// transitions, and only a table can be walked to prove that every combination
// has exactly one answer and that every answer was tested. The same rows, in the
// same order, are in the package comment — that copy is the design artifact the
// spec asks for (전이표 전체는 design 산출물), and TestTheDocumentedTableMatchesTheData
// fails when the two drift apart.

// Row is one transition: the four inputs on the left, what happens on the right.
type Row struct {
	ID          string
	State       State
	Role        Role
	Movement    Movement
	Disposition Disposition
	// Next is the state after. Ignored when Refusal is set — a refused
	// transition leaves the instance exactly where it was.
	Next State
	// NewInstance marks the rows that open a fresh position instance rather than
	// advancing the one that is there.
	NewInstance bool
	// Refusal is empty for an allowed transition.
	Refusal Refusal
}

// Allowed reports whether the row is a transition rather than a refusal.
func (r Row) Allowed() bool { return r.Refusal == RefusalNone }

// Table is every transition, 96 rows: the six states against the two roles,
// against the movements each role can produce, against the three dispositions.
//
// FLAT and CLOSED hold nothing, so an EXIT against them can only be NONE (delta
// zero) or OVERSHOOTS (any delta at all is more than nothing) — REDUCES and
// FLATTENS need 0 < Δ ≤ quantity and are therefore unreachable, not omitted.
var Table = []Row{
	// --- ENTRY, Δ = 0 -------------------------------------------------------
	// A correction, or an order that ended having filled nothing. Neither moves
	// quantity, so neither can be a forbidden transition; what can move is the
	// cost basis and whether the order the state was waiting on is over.
	{ID: "E01", State: Flat, Role: Entry, Movement: MoveNone, Disposition: Working, Next: Flat},
	{ID: "E02", State: Flat, Role: Entry, Movement: MoveNone, Disposition: Done, Next: Flat},
	{ID: "E03", State: Flat, Role: Entry, Movement: MoveNone, Disposition: Succeeded, Next: Flat},
	{ID: "E04", State: Opening, Role: Entry, Movement: MoveNone, Disposition: Working, Next: Opening},
	{ID: "E05", State: Opening, Role: Entry, Movement: MoveNone, Disposition: Done, Next: Open},
	{ID: "E06", State: Opening, Role: Entry, Movement: MoveNone, Disposition: Succeeded, Next: Opening},
	{ID: "E07", State: Open, Role: Entry, Movement: MoveNone, Disposition: Working, Next: Scaling},
	{ID: "E08", State: Open, Role: Entry, Movement: MoveNone, Disposition: Done, Next: Open},
	{ID: "E09", State: Open, Role: Entry, Movement: MoveNone, Disposition: Succeeded, Next: Scaling},
	{ID: "E10", State: Scaling, Role: Entry, Movement: MoveNone, Disposition: Working, Next: Scaling},
	{ID: "E11", State: Scaling, Role: Entry, Movement: MoveNone, Disposition: Done, Next: Open},
	{ID: "E12", State: Scaling, Role: Entry, Movement: MoveNone, Disposition: Succeeded, Next: Scaling},
	{ID: "E13", State: Closing, Role: Entry, Movement: MoveNone, Disposition: Working, Next: Closing},
	{ID: "E14", State: Closing, Role: Entry, Movement: MoveNone, Disposition: Done, Next: Closing},
	{ID: "E15", State: Closing, Role: Entry, Movement: MoveNone, Disposition: Succeeded, Next: Closing},
	{ID: "E16", State: Closed, Role: Entry, Movement: MoveNone, Disposition: Working, Next: Closed},
	{ID: "E17", State: Closed, Role: Entry, Movement: MoveNone, Disposition: Done, Next: Closed},
	{ID: "E18", State: Closed, Role: Entry, Movement: MoveNone, Disposition: Succeeded, Next: Closed},

	// --- ENTRY, Δ > 0 -------------------------------------------------------
	{ID: "E19", State: Flat, Role: Entry, Movement: MoveAdds, Disposition: Working, Next: Opening, NewInstance: true},
	// 즉시 전량체결: the first observation already carries the whole order, and
	// FLAT→OPEN direct is a row of the table rather than an error.
	{ID: "E20", State: Flat, Role: Entry, Movement: MoveAdds, Disposition: Done, Next: Open, NewInstance: true},
	{ID: "E21", State: Flat, Role: Entry, Movement: MoveAdds, Disposition: Succeeded, Next: Opening, NewInstance: true},
	{ID: "E22", State: Opening, Role: Entry, Movement: MoveAdds, Disposition: Working, Next: Opening},
	{ID: "E23", State: Opening, Role: Entry, Movement: MoveAdds, Disposition: Done, Next: Open},
	{ID: "E24", State: Opening, Role: Entry, Movement: MoveAdds, Disposition: Succeeded, Next: Opening},
	{ID: "E25", State: Open, Role: Entry, Movement: MoveAdds, Disposition: Working, Next: Scaling},
	{ID: "E26", State: Open, Role: Entry, Movement: MoveAdds, Disposition: Done, Next: Open},
	{ID: "E27", State: Open, Role: Entry, Movement: MoveAdds, Disposition: Succeeded, Next: Scaling},
	{ID: "E28", State: Scaling, Role: Entry, Movement: MoveAdds, Disposition: Working, Next: Scaling},
	{ID: "E29", State: Scaling, Role: Entry, Movement: MoveAdds, Disposition: Done, Next: Open},
	{ID: "E30", State: Scaling, Role: Entry, Movement: MoveAdds, Disposition: Succeeded, Next: Scaling},
	{ID: "E31", State: Closing, Role: Entry, Movement: MoveAdds, Disposition: Working, Refusal: RefusalEntryWhileClosing},
	{ID: "E32", State: Closing, Role: Entry, Movement: MoveAdds, Disposition: Done, Refusal: RefusalEntryWhileClosing},
	{ID: "E33", State: Closing, Role: Entry, Movement: MoveAdds, Disposition: Succeeded, Refusal: RefusalEntryWhileClosing},
	// 재진입: CLOSED is final, so the buy opens the next instance rather than
	// resurrecting the one that closed.
	{ID: "E34", State: Closed, Role: Entry, Movement: MoveAdds, Disposition: Working, Next: Opening, NewInstance: true},
	{ID: "E35", State: Closed, Role: Entry, Movement: MoveAdds, Disposition: Done, Next: Open, NewInstance: true},
	{ID: "E36", State: Closed, Role: Entry, Movement: MoveAdds, Disposition: Succeeded, Next: Opening, NewInstance: true},

	// --- EXIT, Δ = 0 --------------------------------------------------------
	// The state still moves: an exit order that is working means the position is
	// CLOSING even before it fills, and one that ended without filling hands the
	// state back to whatever the entry side was doing.
	{ID: "X01", State: Flat, Role: Exit, Movement: MoveNone, Disposition: Working, Next: Flat},
	{ID: "X02", State: Flat, Role: Exit, Movement: MoveNone, Disposition: Done, Next: Flat},
	{ID: "X03", State: Flat, Role: Exit, Movement: MoveNone, Disposition: Succeeded, Next: Flat},
	{ID: "X04", State: Opening, Role: Exit, Movement: MoveNone, Disposition: Working, Next: Closing},
	{ID: "X05", State: Opening, Role: Exit, Movement: MoveNone, Disposition: Done, Next: Opening},
	{ID: "X06", State: Opening, Role: Exit, Movement: MoveNone, Disposition: Succeeded, Next: Closing},
	{ID: "X07", State: Open, Role: Exit, Movement: MoveNone, Disposition: Working, Next: Closing},
	{ID: "X08", State: Open, Role: Exit, Movement: MoveNone, Disposition: Done, Next: Open},
	{ID: "X09", State: Open, Role: Exit, Movement: MoveNone, Disposition: Succeeded, Next: Closing},
	{ID: "X10", State: Scaling, Role: Exit, Movement: MoveNone, Disposition: Working, Next: Closing},
	{ID: "X11", State: Scaling, Role: Exit, Movement: MoveNone, Disposition: Done, Next: Scaling},
	{ID: "X12", State: Scaling, Role: Exit, Movement: MoveNone, Disposition: Succeeded, Next: Closing},
	{ID: "X13", State: Closing, Role: Exit, Movement: MoveNone, Disposition: Working, Next: Closing},
	{ID: "X14", State: Closing, Role: Exit, Movement: MoveNone, Disposition: Done, Next: Open},
	{ID: "X15", State: Closing, Role: Exit, Movement: MoveNone, Disposition: Succeeded, Next: Closing},
	{ID: "X16", State: Closed, Role: Exit, Movement: MoveNone, Disposition: Working, Next: Closed},
	{ID: "X17", State: Closed, Role: Exit, Movement: MoveNone, Disposition: Done, Next: Closed},
	{ID: "X18", State: Closed, Role: Exit, Movement: MoveNone, Disposition: Succeeded, Next: Closed},

	// --- EXIT, 0 < Δ < quantity --------------------------------------------
	// 매도 체결 귀속: the sell lands on the open instance and reduces it. Whether
	// the position is still CLOSING afterwards is the exit order's disposition,
	// not the quantity's.
	{ID: "X19", State: Opening, Role: Exit, Movement: MoveReduces, Disposition: Working, Next: Closing},
	{ID: "X20", State: Opening, Role: Exit, Movement: MoveReduces, Disposition: Done, Next: Opening},
	{ID: "X21", State: Opening, Role: Exit, Movement: MoveReduces, Disposition: Succeeded, Next: Closing},
	{ID: "X22", State: Open, Role: Exit, Movement: MoveReduces, Disposition: Working, Next: Closing},
	{ID: "X23", State: Open, Role: Exit, Movement: MoveReduces, Disposition: Done, Next: Open},
	{ID: "X24", State: Open, Role: Exit, Movement: MoveReduces, Disposition: Succeeded, Next: Closing},
	{ID: "X25", State: Scaling, Role: Exit, Movement: MoveReduces, Disposition: Working, Next: Closing},
	{ID: "X26", State: Scaling, Role: Exit, Movement: MoveReduces, Disposition: Done, Next: Scaling},
	{ID: "X27", State: Scaling, Role: Exit, Movement: MoveReduces, Disposition: Succeeded, Next: Closing},
	{ID: "X28", State: Closing, Role: Exit, Movement: MoveReduces, Disposition: Working, Next: Closing},
	{ID: "X29", State: Closing, Role: Exit, Movement: MoveReduces, Disposition: Done, Next: Open},
	{ID: "X30", State: Closing, Role: Exit, Movement: MoveReduces, Disposition: Succeeded, Next: Closing},

	// --- EXIT, Δ = quantity -------------------------------------------------
	// Flat is flat. It does not matter what else was in flight: nothing is held,
	// so the instance is over and the next entry opens a new one.
	{ID: "X31", State: Opening, Role: Exit, Movement: MoveFlattens, Disposition: Working, Next: Closed},
	{ID: "X32", State: Opening, Role: Exit, Movement: MoveFlattens, Disposition: Done, Next: Closed},
	{ID: "X33", State: Opening, Role: Exit, Movement: MoveFlattens, Disposition: Succeeded, Next: Closed},
	{ID: "X34", State: Open, Role: Exit, Movement: MoveFlattens, Disposition: Working, Next: Closed},
	{ID: "X35", State: Open, Role: Exit, Movement: MoveFlattens, Disposition: Done, Next: Closed},
	{ID: "X36", State: Open, Role: Exit, Movement: MoveFlattens, Disposition: Succeeded, Next: Closed},
	{ID: "X37", State: Scaling, Role: Exit, Movement: MoveFlattens, Disposition: Working, Next: Closed},
	{ID: "X38", State: Scaling, Role: Exit, Movement: MoveFlattens, Disposition: Done, Next: Closed},
	{ID: "X39", State: Scaling, Role: Exit, Movement: MoveFlattens, Disposition: Succeeded, Next: Closed},
	{ID: "X40", State: Closing, Role: Exit, Movement: MoveFlattens, Disposition: Working, Next: Closed},
	{ID: "X41", State: Closing, Role: Exit, Movement: MoveFlattens, Disposition: Done, Next: Closed},
	{ID: "X42", State: Closing, Role: Exit, Movement: MoveFlattens, Disposition: Succeeded, Next: Closed},

	// --- EXIT, Δ > quantity -------------------------------------------------
	// Three different facts, three different refusals: nothing was ever held,
	// the instance is already closed, or more was sold than is held. They are
	// named apart because the operator's next move differs for each.
	{ID: "X43", State: Flat, Role: Exit, Movement: MoveOvershoots, Disposition: Working, Refusal: RefusalUnattributedSell},
	{ID: "X44", State: Flat, Role: Exit, Movement: MoveOvershoots, Disposition: Done, Refusal: RefusalUnattributedSell},
	{ID: "X45", State: Flat, Role: Exit, Movement: MoveOvershoots, Disposition: Succeeded, Refusal: RefusalUnattributedSell},
	{ID: "X46", State: Opening, Role: Exit, Movement: MoveOvershoots, Disposition: Working, Refusal: RefusalOversell},
	{ID: "X47", State: Opening, Role: Exit, Movement: MoveOvershoots, Disposition: Done, Refusal: RefusalOversell},
	{ID: "X48", State: Opening, Role: Exit, Movement: MoveOvershoots, Disposition: Succeeded, Refusal: RefusalOversell},
	{ID: "X49", State: Open, Role: Exit, Movement: MoveOvershoots, Disposition: Working, Refusal: RefusalOversell},
	{ID: "X50", State: Open, Role: Exit, Movement: MoveOvershoots, Disposition: Done, Refusal: RefusalOversell},
	{ID: "X51", State: Open, Role: Exit, Movement: MoveOvershoots, Disposition: Succeeded, Refusal: RefusalOversell},
	{ID: "X52", State: Scaling, Role: Exit, Movement: MoveOvershoots, Disposition: Working, Refusal: RefusalOversell},
	{ID: "X53", State: Scaling, Role: Exit, Movement: MoveOvershoots, Disposition: Done, Refusal: RefusalOversell},
	{ID: "X54", State: Scaling, Role: Exit, Movement: MoveOvershoots, Disposition: Succeeded, Refusal: RefusalOversell},
	{ID: "X55", State: Closing, Role: Exit, Movement: MoveOvershoots, Disposition: Working, Refusal: RefusalOversell},
	{ID: "X56", State: Closing, Role: Exit, Movement: MoveOvershoots, Disposition: Done, Refusal: RefusalOversell},
	{ID: "X57", State: Closing, Role: Exit, Movement: MoveOvershoots, Disposition: Succeeded, Refusal: RefusalOversell},
	{ID: "X58", State: Closed, Role: Exit, Movement: MoveOvershoots, Disposition: Working, Refusal: RefusalSellOnClosed},
	{ID: "X59", State: Closed, Role: Exit, Movement: MoveOvershoots, Disposition: Done, Refusal: RefusalSellOnClosed},
	{ID: "X60", State: Closed, Role: Exit, Movement: MoveOvershoots, Disposition: Succeeded, Refusal: RefusalSellOnClosed},
}

// tableIndex is Table keyed by its four inputs. It is built once, from the same
// literal, so there is exactly one place a row can be written.
var tableIndex = func() map[transitionKey]Row {
	index := make(map[transitionKey]Row, len(Table))
	for _, row := range Table {
		index[transitionKey{row.State, row.Role, row.Movement, row.Disposition}] = row
	}
	return index
}()

type transitionKey struct {
	state       State
	role        Role
	movement    Movement
	disposition Disposition
}

// Lookup returns the row for one combination of the four inputs.
func Lookup(state State, role Role, movement Movement, disposition Disposition) (Row, bool) {
	row, ok := tableIndex[transitionKey{state, role, movement, disposition}]
	return row, ok
}
