package position

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// table_test.go walks the transition table (task 6.1: 전이표 전체는 design
// 산출물이며 태스크는 표의 전 행을 테스트한다).
//
// Three things are asserted and none of them is enough alone:
//
//  1. every row is *driven* — Apply is called with concrete inputs that classify
//     into that row, and the state, quantity and refusal it produces are the
//     ones the row promises;
//  2. every reachable combination of the four inputs has a row, and no
//     combination has two;
//  3. the table in the package comment says the same thing as the table in the
//     code, so the design artifact cannot rot away from the program.

// caseFor builds an instance and an event that classify into one row.
//
// The quantities are chosen so the movement is the row's own: FLAT and CLOSED
// hold nothing, everything else holds 10, and the delta is picked against that.
func caseFor(row Row) (Instance, Event, bool) {
	held := "10"
	if row.State == Flat || row.State == Closed {
		held = "0"
	}

	var delta string
	switch row.Movement {
	case MoveNone:
		delta = "0"
	case MoveAdds:
		delta = "4"
	case MoveReduces:
		delta = "4"
	case MoveFlattens:
		delta = held
	case MoveOvershoots:
		delta = "14"
	}
	// REDUCES and FLATTENS need something to reduce; a row that asks for them
	// against an empty instance is one of the twelve unreachable combinations.
	if held == "0" && (row.Movement == MoveReduces || row.Movement == MoveFlattens) {
		return Instance{}, Event{}, false
	}

	// The order's own numbers decide the disposition. Ordered 100, so a fill of
	// 4 never completes it by arithmetic; the three dispositions are then
	// separated by the terminal flag and the replace edge.
	ev := Event{
		Role:              row.Role,
		Delta:             delta,
		OrderQuantity:     "100",
		OrderFilled:       "4",
		PrevOrderFilled:   "0",
		OrderAvgPrice:     "500",
		PrevOrderAvgPrice: "",
	}
	if row.Movement == MoveNone {
		ev.OrderFilled, ev.PrevOrderFilled = "0", "0"
	}
	switch row.Disposition {
	case Working:
	case Done:
		ev.Terminal = true
	case Succeeded:
		ev.Terminal, ev.HasSuccessor = true, true
	}
	return Instance{State: row.State, Quantity: held, AvgPrice: "500"}, ev, true
}

// TestEveryTransitionRowIsDriven runs Apply once per row and checks that the row
// the classification picked is the row under test — which is what makes this a
// test of the table and not of a re-derivation of it.
func TestEveryTransitionRowIsDriven(t *testing.T) {
	t.Parallel()

	driven := map[string]bool{}
	for _, row := range Table {
		inst, ev, constructible := caseFor(row)
		if !constructible {
			t.Errorf("row %s asks for movement %s against an empty instance; that combination "+
				"cannot occur and must not be in the table", row.ID, row.Movement)
			continue
		}
		out, err := Apply(inst, ev)
		if err != nil {
			t.Fatalf("row %s: Apply: %v", row.ID, err)
		}
		if out.Row != row.ID {
			t.Fatalf("inputs for row %s classified into %s (movement %s, disposition %s)",
				row.ID, out.Row, out.Movement, out.Disposition)
		}
		driven[row.ID] = true

		if !row.Allowed() {
			if out.Refusal != row.Refusal {
				t.Errorf("row %s refusal = %q, want %q", row.ID, out.Refusal, row.Refusal)
			}
			if out.Next != inst.State {
				t.Errorf("row %s moved the state to %s; a refused transition transitions nothing",
					row.ID, out.Next)
			}
			if out.Quantity != inst.Quantity || out.AvgPrice != inst.AvgPrice {
				t.Errorf("row %s changed the instance to (%s, %s); 산식 보정 금지",
					row.ID, out.Quantity, out.AvgPrice)
			}
			if out.Reason == "" {
				t.Errorf("row %s refused without saying why", row.ID)
			}
			continue
		}

		if out.Refusal != RefusalNone {
			t.Errorf("row %s is an allowed transition but refused with %q", row.ID, out.Refusal)
		}
		if out.Next != row.Next {
			t.Errorf("row %s next = %s, want %s", row.ID, out.Next, row.Next)
		}
		if out.NewInstance != row.NewInstance {
			t.Errorf("row %s NewInstance = %v, want %v", row.ID, out.NewInstance, row.NewInstance)
		}
		wantQuantity := expectedQuantity(row, inst.Quantity, ev.Delta)
		if out.Quantity != wantQuantity {
			t.Errorf("row %s quantity = %s, want %s", row.ID, out.Quantity, wantQuantity)
		}
		// The one invariant that must hold on every allowed row: a CLOSED
		// instance holds nothing, and an instance that holds nothing is CLOSED
		// or was never opened.
		if out.Next == Closed && out.Quantity != "0" {
			t.Errorf("row %s reached CLOSED holding %s", row.ID, out.Quantity)
		}
	}
	if len(driven) != len(Table) {
		t.Fatalf("drove %d of %d rows", len(driven), len(Table))
	}
	if len(Table) != 96 {
		t.Fatalf("the table has %d rows, want the 96 the package comment enumerates", len(Table))
	}
}

func expectedQuantity(row Row, held, delta string) string {
	base := held
	if row.NewInstance {
		base = "0"
	}
	switch row.Role {
	case Entry:
		switch {
		case delta == "0":
			return base
		case base == "0":
			return delta
		case base == "10" && delta == "4":
			return "14"
		}
	case Exit:
		switch {
		case delta == "0":
			return base
		case base == "10" && delta == "4":
			return "6"
		case base == delta:
			return "0"
		}
	}
	return "?" // an unreachable combination; the assertion above will report it
}

// TestTheTableCoversEveryReachableCombination is the completeness half. A
// missing row would come back from Apply as an error on a live fill; a duplicate
// would mean two rules for one event and the winner would be whichever was
// written last.
func TestTheTableCoversEveryReachableCombination(t *testing.T) {
	t.Parallel()

	seen := map[transitionKey]string{}
	for _, row := range Table {
		key := transitionKey{row.State, row.Role, row.Movement, row.Disposition}
		if first, dup := seen[key]; dup {
			t.Errorf("rows %s and %s both claim (%s, %s, %s, %s)",
				first, row.ID, row.State, row.Role, row.Movement, row.Disposition)
		}
		seen[key] = row.ID
	}

	states := []State{Flat, Opening, Open, Scaling, Closing, Closed}
	dispositions := []Disposition{Working, Done, Succeeded}
	want := 0
	for _, state := range states {
		empty := state == Flat || state == Closed
		for _, role := range []Role{Entry, Exit} {
			movements := []Movement{MoveNone, MoveAdds}
			if role == Exit {
				movements = []Movement{MoveNone, MoveReduces, MoveFlattens, MoveOvershoots}
				if empty {
					// Nothing is held, so any sell at all overshoots.
					movements = []Movement{MoveNone, MoveOvershoots}
				}
			}
			for _, movement := range movements {
				for _, disposition := range dispositions {
					want++
					if _, ok := Lookup(state, role, movement, disposition); !ok {
						t.Errorf("no row for (%s, %s, %s, %s)", state, role, movement, disposition)
					}
				}
			}
		}
	}
	if want != len(Table) {
		t.Fatalf("the reachable combinations are %d but the table has %d rows", want, len(Table))
	}
}

// TestUnreachableMovementsCannotBeConstructed is the other half of "unreachable
// rather than omitted": the twelve combinations the table leaves out are the
// ones the classifier cannot produce, not the ones nobody thought about.
func TestUnreachableMovementsCannotBeConstructed(t *testing.T) {
	t.Parallel()

	for _, delta := range []string{"0", "1", "1000000", "0.0001"} {
		movement, err := classifyMovement(Exit, delta, "0")
		if err != nil {
			t.Fatalf("classifying a sell of %s against nothing: %v", delta, err)
		}
		if movement == MoveReduces || movement == MoveFlattens {
			t.Fatalf("a sell of %s against a held quantity of 0 classified as %s; "+
				"REDUCES and FLATTENS both need something to reduce", delta, movement)
		}
	}
}

// TestTheDocumentedTableMatchesTheData compares the package comment's table with
// the literal. The comment is the design artifact the spec asks for, and an
// artifact that no longer matches the program is worse than none: a reviewer
// checking the rule would be reading fiction.
func TestTheDocumentedTableMatchesTheData(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "doc.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing doc.go: %v", err)
	}
	var comment string
	ast.Inspect(file, func(ast.Node) bool { return true })
	if file.Doc == nil {
		t.Fatal("doc.go has no package comment; the design artifact is missing")
	}
	comment = file.Doc.Text()

	// Each documented row is "  Exx  STATE  DISPOSITION  NEXT …" — the id, then
	// the fields, whitespace separated.
	line := regexp.MustCompile(`(?m)^\t([EX]\d{2})\s+(\S+)\s+(\S+)\s+(\S+)(.*)$`)
	documented := map[string][]string{}
	for _, m := range line.FindAllStringSubmatch(comment, -1) {
		documented[m[1]] = []string{m[2], m[3], m[4], strings.TrimSpace(m[5])}
	}
	if len(documented) != len(Table) {
		var missing []string
		for _, row := range Table {
			if _, ok := documented[row.ID]; !ok {
				missing = append(missing, row.ID)
			}
		}
		sort.Strings(missing)
		t.Fatalf("the package comment documents %d rows and the table has %d; missing %v",
			len(documented), len(Table), missing)
	}

	for _, row := range Table {
		got := documented[row.ID]
		if State(got[0]) != row.State {
			t.Errorf("row %s: comment says state %s, table says %s", row.ID, got[0], row.State)
		}
		if Disposition(got[1]) != row.Disposition {
			t.Errorf("row %s: comment says disposition %s, table says %s", row.ID, got[1], row.Disposition)
		}
		wantNext := string(row.Next)
		if !row.Allowed() {
			wantNext = "RECONCILE"
			if !strings.Contains(got[3], string(row.Refusal)) {
				t.Errorf("row %s: comment does not name the refusal %s", row.ID, row.Refusal)
			}
		}
		if got[2] != wantNext {
			t.Errorf("row %s: comment says next %s, table says %s", row.ID, got[2], wantNext)
		}
	}
}
