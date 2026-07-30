package verifylive

// hold_test.go pins the property a trigger measurement cannot exist without: an
// object an unfinished measurement is waiting on is not a cleanup target.
//
// The rule under test is one sentence — *the gate a line names must decide after
// that line* — and every test here is a record shaped to put a decision on one
// side of a hold or the other. Records are hand-built for the same reason
// cleanup_test.go builds them: the failure being guarded against is an ordering
// between runs days apart, and the harness runs one invocation at a time.
//
// The regression direction matters as much as the new behaviour. Before this file
// existed, an order was an unconditional cleanup target (cleanup.go) and a
// conditional was gated on conditional-cancel alone. Both of those judgements must
// survive unchanged for every record written before the hold fields existed —
// that is what TestALegacyRecordIsJudgedExactlyAsBefore is for, and it is the
// evidence that this change does not alter what the live account sees.

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing"
)

// held is an artifact a step declared it is waiting on.
func held(kind, id, symbol string, gate StepID, chain string) Artifact {
	return Artifact{
		Kind: kind, ID: id, Symbol: symbol,
		Deliberate: true, HeldUntil: gate, ChainID: chain,
	}
}

// --- P1: orders can be held at all ---------------------------------------------

// TestAHeldOrderIsNotACleanupTarget is the whole reason this change exists.
//
// A conditional order that fires produces a child market sell, and that child has
// to be *allowed to fill* — "not cancelling it" is the content of the
// measurement. Today every order this tool created is an unconditional cleanup
// target (cleanup.go cleanupFrom), so the next run would put the child on the
// approval list and cancel it. That is M37's shape with a different kind.
func TestAHeldOrderIsNotACleanupTarget(t *testing.T) {
	entries := []Entry{
		{Kind: KindStep, StepID: StepConditionalTrigger, Verdict: VerdictAwaitingRestart,
			Artifacts: []Artifact{held(KindOrder, "child-1", "333430", StepConditionalTrigger, "chain-A")}},
	}

	if got := PendingCleanup(entries); len(got) != 0 {
		t.Fatalf("the prologue would cancel the child order the trigger measurement has to watch fill; "+
			"the step that placed it has not reached a terminal verdict: %+v", got)
	}
}

// TestAHeldConditionalIsNotACleanupTarget is the same rule on the kind that
// already had a guard, stated through the new field rather than through the
// hardcoded conditional-cancel default.
func TestAHeldConditionalIsNotACleanupTarget(t *testing.T) {
	entries := []Entry{
		{Kind: KindStep, StepID: StepConditionalTrigger, Verdict: VerdictAwaitingRestart,
			Artifacts: []Artifact{held(KindConditional, "co-1", "333430", StepConditionalTrigger, "chain-A")}},
	}

	if got := PendingCleanup(entries); len(got) != 0 {
		t.Fatalf("a conditional held by an unfinished step must not be a cleanup target: %+v", got)
	}
}

// --- release ------------------------------------------------------------------

// TestAHoldEndsWhenItsGateDecides keeps the guard from becoming a leak. A held
// object whose gate has finished is an ordinary leftover and must reach the
// approval list, or verify-clears-leftovers' deadlock grows back in a new place.
func TestAHoldEndsWhenItsGateDecides(t *testing.T) {
	entries := []Entry{
		{Kind: KindStep, StepID: StepConditionalTrigger, Verdict: VerdictAwaitingRestart,
			Artifacts: []Artifact{held(KindOrder, "child-1", "333430", StepConditionalTrigger, "chain-A")}},
		{Kind: KindStep, StepID: StepConditionalTrigger, Verdict: VerdictFail},
	}

	got := PendingCleanup(entries)
	if len(got) != 1 || got[0].ID != "child-1" {
		t.Fatalf("once the holding step has a terminal verdict recorded after the hold, the object is a "+
			"leftover and must be offered for cancellation, got %+v", got)
	}
}

// TestAGateVerdictOlderThanTheHoldDoesNotRelease is M37's rule applied to the
// general case: a verdict recorded before a line is not a verdict about what that
// line said.
func TestAGateVerdictOlderThanTheHoldDoesNotRelease(t *testing.T) {
	entries := []Entry{
		{Kind: KindStep, StepID: StepConditionalTrigger, Verdict: VerdictFail},
		{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass,
			Artifacts: []Artifact{held(KindOrder, "child-1", "333430", StepConditionalTrigger, "chain-A")}},
	}

	if got := PendingCleanup(entries); len(got) != 0 {
		t.Fatalf("the gate's verdict predates the hold, so it decided nothing about this object: %+v", got)
	}
}

// TestAFailedCancelAfterTheHoldStillReleases is the M22 guard.
//
// design.md D2 rejected an earlier definition that measured from "the most recent
// line that declared a hold" rather than from the line Outstanding selected.
// Under that definition a cancel that failed *after* the hold could be undone by
// any later mention, and a leftover whose cancel failed would stop being a
// cleanup target — which is exactly the deadlock verify-clears-leftovers exists to
// remove. Measuring from the line that names the gate makes that impossible.
func TestAFailedCancelAfterTheHoldStillReleases(t *testing.T) {
	entries := []Entry{
		{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass,
			Artifacts: []Artifact{held(KindConditional, "co-1", "333430", StepConditionalCancel, "chain-A")}},
		{Kind: KindStep, StepID: StepConditionalCancel, Verdict: VerdictFail},
	}

	got := PendingCleanup(entries)
	if len(got) != 1 || got[0].ID != "co-1" {
		t.Fatalf("a conditional whose cancel failed after it was held is a leftover; refusing to clean it up "+
			"would rebuild the deadlock verify-clears-leftovers removed, got %+v", got)
	}
}

// TestAReDeclaredHoldOutlivesAnOlderVerdict is the case that separates the new
// definition from the old one (design.md D2, first row of the table).
//
// The cancel failed, and *after* that a later step declared the object held
// again. The newest line naming a gate is the authority, and its gate has not
// decided since. Recovery is not lost: a redo of the gate writes a newer verdict
// and the object is released.
func TestAReDeclaredHoldOutlivesAnOlderVerdict(t *testing.T) {
	entries := []Entry{
		{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass,
			Artifacts: []Artifact{held(KindConditional, "co-1", "333430", StepConditionalCancel, "chain-A")}},
		{Kind: KindStep, StepID: StepConditionalCancel, Verdict: VerdictFail},
		{Kind: KindStep, StepID: StepConditionalPersist, Verdict: VerdictPass,
			Artifacts: []Artifact{held(KindConditional, "co-1", "333430", StepConditionalCancel, "chain-A")}},
	}

	if got := PendingCleanup(entries); len(got) != 0 {
		t.Fatalf("the newest line holds this object until conditional-cancel decides again, and the only "+
			"verdict on record predates that line: %+v", got)
	}

	released := append(append([]Entry{}, entries...),
		Entry{Kind: KindStep, StepID: StepConditionalCancel, Verdict: VerdictFail})
	if got := PendingCleanup(released); len(got) != 1 {
		t.Fatalf("a newer gate verdict must release the hold, or a redo could not recover the account: %+v", got)
	}
}

// --- the legacy record must be judged identically -------------------------------

// TestALegacyRecordIsJudgedExactlyAsBefore is this change's central safety claim:
// every line written before the hold fields existed keeps the judgement it has
// today. The two shapes below are the two rules cleanup.go carries, taken from
// the real records on the account.
//
// The baseline is the *current HEAD's* judgement, not what actually happened on
// 2026-07-28: the conditional cancelled that day was cancelled because the bug
// verify-reopens-conditional-chain fixed was still present. Replaying that record
// through today's code correctly refuses.
func TestALegacyRecordIsJudgedExactlyAsBefore(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		want    []string
	}{
		{
			name: "조건주문은 conditional-cancel이 gate다 — 그 판정이 뒤에 있으면 대상",
			entries: []Entry{
				{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass, Artifacts: []Artifact{
					{Kind: KindConditional, ID: "grLKqiGuCVS7mj", Symbol: "333430", Deliberate: true},
				}},
				{Kind: KindStep, StepID: StepConditionalCancel, Verdict: VerdictFail},
			},
			want: []string{"grLKqiGuCVS7mj"},
		},
		{
			name: "조건주문 — 판정이 앞에 있으면 대상이 아니다 (M37)",
			entries: []Entry{
				{Kind: KindStep, StepID: StepConditionalCancel, Verdict: VerdictSkipped},
				{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass, Artifacts: []Artifact{
					{Kind: KindConditional, ID: "grLKqiGuCVS7mj", Symbol: "333430", Deliberate: true},
				}},
			},
			want: nil,
		},
		{
			name: "주문은 gate가 없다 — 언제나 대상",
			entries: []Entry{
				{Kind: KindStep, StepID: StepConditionalCancel, Verdict: VerdictSkipped},
				{Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictFail, Artifacts: []Artifact{
					{Kind: KindOrder, ID: "OsBakht", Symbol: "005930"},
				}},
			},
			want: []string{"OsBakht"},
		},
		{
			name: "조건주문 — cancel 줄이 아예 없으면 권한이 없다",
			entries: []Entry{
				{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass, Artifacts: []Artifact{
					{Kind: KindConditional, ID: "co-1", Symbol: "333430", Deliberate: true},
				}},
			},
			want: nil,
		},
		{
			name: "조건주문 — Deliberate가 아니어도 같은 gate를 쓴다",
			entries: []Entry{
				{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass, Artifacts: []Artifact{
					{Kind: KindConditional, ID: "co-1", Symbol: "333430"},
				}},
				{Kind: KindStep, StepID: StepConditionalCancel, Verdict: VerdictFail},
			},
			want: []string{"co-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ids []string
			for _, a := range PendingCleanup(tc.entries) {
				ids = append(ids, a.ID)
			}
			if strings.Join(ids, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("legacy judgement changed: got %v, want %v", ids, tc.want)
			}
		})
	}
}

// --- the record has to carry it -------------------------------------------------

// TestAHeldArtifactSurvivesTheRecord: the fields are only useful if a later
// process reads them back, which is the entire point — the holding process is
// usually gone by the time cleanup runs.
func TestAHeldArtifactSurvivesTheRecord(t *testing.T) {
	dir := t.TempDir()
	rec, err := OpenRecorder(dir + "/verify.jsonl")
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	want := held(KindOrder, "child-1", "333430", StepConditionalTrigger, "chain-A")
	if err := rec.Append(Entry{
		FormatVersion: RecordFormatVersion, Kind: KindStep, StepID: StepConditionalTrigger,
		Verdict: VerdictAwaitingRestart, Artifacts: []Artifact{want},
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	entries, err := LoadEntries(dir + "/verify.jsonl")
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	if len(entries) != 1 || len(entries[0].Artifacts) != 1 {
		t.Fatalf("record did not round trip: %+v", entries)
	}
	got := entries[0].Artifacts[0]
	if got.HeldUntil != want.HeldUntil || got.ChainID != want.ChainID {
		t.Fatalf("the hold did not survive the record: got HeldUntil=%q ChainID=%q, want %q/%q",
			got.HeldUntil, got.ChainID, want.HeldUntil, want.ChainID)
	}
}

// TestAnUnheldArtifactWritesNoHoldFields keeps the record readable for a human
// and keeps old readers unaffected: the fields are omitempty, so a line for an
// object nobody is waiting on looks exactly as it does today.
func TestAnUnheldArtifactWritesNoHoldFields(t *testing.T) {
	raw, err := json.Marshal(Artifact{Kind: KindOrder, ID: "o-1", Symbol: "005930"})
	if err != nil {
		t.Fatalf("marshalling an artifact: %v", err)
	}
	for _, field := range []string{"held_until", "chain_id"} {
		if strings.Contains(string(raw), field) {
			t.Errorf("an unheld artifact wrote %q into the record: %s", field, raw)
		}
	}
}

// --- the modify carries the chain ------------------------------------------------

// TestAModifiedConditionalKeepsTheChainAndTheHold: a conditional modify issues a
// new identifier and invalidates the old one immediately (measurements.md M19,
// M40 — observed in both markets). The replacement must arrive already held, or
// there is a window in which the successor of a held object is an ordinary
// leftover.
func TestAModifiedConditionalKeepsTheChainAndTheHold(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 2)
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("first invocation: %v", err)
	}
	broker.restart()
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Logf("second invocation: %v", err)
	}

	var chains []string
	var heldCount int
	for _, e := range h.entries() {
		if e.StepID != StepConditionalModify {
			continue
		}
		for _, a := range e.Artifacts {
			if a.Cancelled {
				continue
			}
			if a.ChainID == "" {
				t.Errorf("the conditional the modify created carries no chain, so the record cannot say it "+
					"continues the one that was replaced: %+v", a)
			}
			if a.HeldUntil == "" {
				t.Errorf("the replacement arrived unheld: %+v", a)
			}
			chains = append(chains, a.ChainID)
			heldCount++
		}
	}
	if heldCount == 0 {
		t.Fatal("this test needs the modify step to have created a conditional")
	}

	for _, e := range h.entries() {
		if e.StepID != StepConditionalRegister {
			continue
		}
		for _, a := range e.Artifacts {
			if a.ChainID == "" {
				t.Fatalf("the registered conditional carries no chain: %+v", a)
			}
			for _, c := range chains {
				if c != a.ChainID {
					t.Errorf("the modify started a new chain %q instead of continuing %q — the record then "+
						"cannot reconstruct that these are one protection", c, a.ChainID)
				}
			}
		}
	}
}

// --- static guards ---------------------------------------------------------------

// TestTheCleanupDecisionDoesNotReadAClock pins design.md D3.
//
// A time-based release would put "cancel a live broker object because enough time
// passed" back into the tool, which is M37's shape, and it would fire in the
// middle of a trigger measurement's waiting window — where a long wait means the
// market has not moved yet, not that anything failed. The decision is positional
// and has to stay that way; a comment saying so is what drifted last time.
func TestTheCleanupDecisionDoesNotReadAClock(t *testing.T) {
	decide := map[string]bool{"cleanupFrom": true, "heldAfter": true, "holdGate": true}
	banned := []string{"Now", "CreatedAt", "CancelledAt", "Since", "After", "Sub"}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "cleanup.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing cleanup.go: %v", err)
	}

	seen := map[string]bool{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || !decide[fn.Name.Name] {
			continue
		}
		seen[fn.Name.Name] = true
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			for _, b := range banned {
				if sel.Sel.Name == b {
					t.Errorf("%s reads %s: the cleanup decision must be positional (design.md D3)",
						fn.Name.Name, b)
				}
			}
			return true
		})
	}
	for name := range decide {
		if !seen[name] {
			t.Fatalf("%s was not found in cleanup.go — this guard is asserting nothing", name)
		}
	}
}

// TestEveryHoldGateIsACatalogueStep closes the one way fail-closed turns into a
// silent deadlock (review.md A6): a gate naming a step that does not exist is
// never settled, so the object it holds is never released and never explained.
func TestEveryHoldGateIsACatalogueStep(t *testing.T) {
	known := map[StepID]bool{StepCleanup: true}
	for _, s := range Steps() {
		known[s.ID] = true
	}

	fset := token.NewFileSet()
	pkg, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing the package: %v", err)
	}

	constOf := map[string]StepID{}
	for _, p := range pkg {
		for _, file := range p.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.CONST {
					continue
				}
				for _, spec := range gen.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
						continue
					}
					lit, ok := vs.Values[0].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					constOf[vs.Names[0].Name] = StepID(strings.Trim(lit.Value, `"`))
				}
			}
		}
	}
	if len(constOf) == 0 {
		t.Fatal("no string constants were read — this guard is asserting nothing")
	}

	// Two spellings reach the field: the markHeld argument every step uses, and a
	// composite literal if anything ever builds an Artifact directly.
	gate := func(where string, expr ast.Expr) {
		ident, ok := expr.(*ast.Ident)
		if !ok {
			t.Errorf("%s: the hold gate is an expression this guard cannot resolve (%T) — it has to be a "+
				"catalogue step read from a constant, or a typo becomes a permanent hold", where, expr)
			return
		}
		if id, found := constOf[ident.Name]; !found || !known[id] {
			t.Errorf("%s: the hold gate is %s (%q), which is not a step in the catalogue — nothing will ever "+
				"settle it and the object it holds is never released", where, ident.Name, id)
		}
	}

	checked := 0
	for _, p := range pkg {
		for name, file := range p.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.CallExpr:
					sel, ok := node.Fun.(*ast.SelectorExpr)
					if !ok || sel.Sel.Name != "markHeld" || len(node.Args) < 3 {
						return true
					}
					checked++
					gate(name, node.Args[2])
				case *ast.KeyValueExpr:
					key, ok := node.Key.(*ast.Ident)
					if !ok || key.Name != "HeldUntil" {
						return true
					}
					checked++
					gate(name, node.Value)
				}
				return true
			})
		}
	}
	if checked == 0 {
		t.Fatal("no hold gate was found in the package — this guard is asserting nothing")
	}
}
