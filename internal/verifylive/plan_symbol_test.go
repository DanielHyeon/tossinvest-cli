package verifylive

// plan_symbol_test.go covers one property: the approval list names the object the
// step will actually act on.
//
// A batch approval is a person agreeing to a list. The list carries a symbol on
// every line and Authorises compares it exactly, so a plan that named the wrong
// object had two possible outcomes and both are bad — the run stops on something
// the operator did approve (2026-07-29, twice), or, if the comparison were loosened
// to make that stop go away, a request reaches the broker for an object nobody read.
// The fix is neither of those: the plan resolves the same object the step resolves.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// conditionalPlanLines returns the plan's lines for the two steps that act on an
// already-registered conditional order.
func conditionalPlanLines(p Plan) []PlannedMutation {
	var out []PlannedMutation
	for _, m := range p.Mutations {
		if m.Step == StepConditionalModify || m.Step == StepConditionalCancel {
			out = append(out, m)
		}
	}
	return out
}

// TestThePlanNamesTheConditionalItWillModify is the 2026-07-29 failure, as a test.
//
// The account holds 333430 and the run's probe symbol is 005930 — the shape the
// console produces, where the buy-side probes go to a fixed KR symbol and the
// conditional lives on whatever the account happens to hold. The conditional steps
// act on the registered conditional, so that is what their lines have to say.
func TestThePlanNamesTheConditionalItWillModify(t *testing.T) {
	broker := newFakeBroker().withHolding("333430", 2)
	h := newHarness(t, broker, alwaysConfirm())

	// One invocation registers the conditional and halts at the persistence check,
	// leaving it alive — the state the record was in when the operator pressed
	// [재측정].
	if _, err := h.run(Options{Symbol: "005930", HoldingSymbol: "333430"}); err != nil {
		t.Logf("first invocation: %v", err)
	}
	if h.verdict(StepConditionalPersist) != VerdictAwaitingRestart {
		t.Fatalf("this test needs the run to stop at conditional-persist, got %q", h.verdict(StepConditionalPersist))
	}

	broker.restart()
	plan := h.runner(t, Options{Symbol: "005930", HoldingSymbol: "333430"}).Plan(context.Background())

	lines := conditionalPlanLines(plan)
	if len(lines) == 0 {
		t.Fatalf("the resumed plan lists neither conditional step:\n%+v", plan.Excluded)
	}
	for _, m := range lines {
		if !strings.EqualFold(m.Symbol, "333430") {
			t.Errorf("%s is planned against %q, but the step acts on the registered conditional order on "+
				"333430. The run would stop on a request the operator approved.", m.Step, m.Symbol)
		}
	}
}

// TestAFreshRunPlansTheConditionalStepsAgainstTheHolding — the same property when
// nothing is registered yet. The conditional the modify will act on is the one
// conditional-register is about to create, and that is the holding.
func TestAFreshRunPlansTheConditionalStepsAgainstTheHolding(t *testing.T) {
	broker := newFakeBroker().withHolding("333430", 2)
	h := newHarness(t, broker, alwaysConfirm())

	plan := h.runner(t, Options{Symbol: "005930", HoldingSymbol: "333430"}).Plan(context.Background())

	lines := conditionalPlanLines(plan)
	if len(lines) == 0 {
		t.Fatalf("a fresh plan lists neither conditional step:\n%+v", plan.Excluded)
	}
	for _, m := range lines {
		if !strings.EqualFold(m.Symbol, "333430") {
			t.Errorf("%s is planned against %q; conditional-register will register on the holding 333430 "+
				"and this step acts on what it registered", m.Step, m.Symbol)
		}
	}
}

// TestTheConditionalStepsReachTheBrokerWhenTheProbeSymbolDiffers is the claim of
// this change, measured end to end rather than field by field.
//
// Two invocations, because the persistence step cannot pass inside the process that
// registered the conditional. The second one has to get through modify and cancel,
// and the account has to be left as the verification found it.
func TestTheConditionalStepsReachTheBrokerWhenTheProbeSymbolDiffers(t *testing.T) {
	broker := newFakeBroker().withHolding("333430", 2)
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{Symbol: "005930", HoldingSymbol: "333430"}); err != nil {
		t.Logf("first invocation: %v", err)
	}
	broker.restart()
	if _, err := h.run(Options{Symbol: "005930", HoldingSymbol: "333430"}); err != nil {
		t.Logf("second invocation: %v", err)
	}

	for _, id := range []StepID{StepConditionalPersist, StepConditionalModify, StepConditionalCancel} {
		if v := h.verdict(id); v != VerdictPass {
			e, _ := LastEntry(h.entries(), id)
			t.Errorf("%s = %q, want pass: %s", id, v, e.Reason)
		}
	}
	if out := Outstanding(h.entries()); len(out) > 0 {
		t.Errorf("the verification left %+v on the account; conditional-cancel is the step that removes it", out)
	}
}

// TestAMutatingStepWithNoNameableTargetIsNotPlanned.
//
// MarketOf("") is "US" (six digits or it is not KR), so an empty symbol slips past
// the market check on a US run in a way it cannot on a KR one. A line with no symbol
// is a line that does not say what it is a request about, and the operator is being
// asked to approve it.
// The runner is built without the harness because the harness fills a probe symbol
// in for a caller that gave none, and the state under test is the one where the
// account could not supply one: consoleProbeSymbol returns the first usable US
// holding, which is empty on a US account holding nothing.
func TestAMutatingStepWithNoNameableTargetIsNotPlanned(t *testing.T) {
	recorder, err := OpenRecorder(t.TempDir() + "/verify.jsonl")
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	t.Cleanup(func() { _ = recorder.Close() })

	r, err := New(Options{
		Broker:       newFakeBroker(),
		Recorder:     recorder,
		Confirm:      func(Mutation) error { return nil },
		ConfirmBatch: func(Batch) error { return nil },
		AccountRef:   "123-45-678901",
		Market:       MarketUS,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	plan := r.Plan(context.Background())

	for _, m := range plan.Mutations {
		if strings.TrimSpace(m.Symbol) == "" {
			t.Errorf("the approval list carries a line with no symbol: %+v", m)
		}
	}
	if len(plan.Mutations) != 0 {
		t.Errorf("a run that can name no target planned %d live request(s)", len(plan.Mutations))
	}
	if len(plan.Excluded) == 0 {
		t.Error("the plan drops every mutating step without saying why")
	}
}

// TestAPlanLineWithoutASymbolAuthorisesNothingWithOne.
//
// The old comparison skipped the check when the planned line carried no symbol,
// which made such a line authorise a request for any symbol at all. A cleanup line
// for an artifact whose symbol was never recorded still authorises its own
// symbol-less cancel, which is the only case that ever relied on the skip.
func TestAPlanLineWithoutASymbolAuthorisesNothingWithOne(t *testing.T) {
	p := Plan{Mutations: []PlannedMutation{{
		Step: StepCleanup,
		Kind: MutateCancelConditional,
	}}}

	if p.Authorises(StepCleanup, MutateCancelConditional, "333430", "", 0) {
		t.Error("a line that names no symbol authorised a request for 333430 — a list that says nothing " +
			"about the object cannot be a list of what was approved")
	}
	if !p.Authorises(StepCleanup, MutateCancelConditional, "", "", 0) {
		t.Error("a symbol-less line no longer authorises its own symbol-less request; an old record whose " +
			"artifact carried no symbol can no longer be cleaned up")
	}
}

// TestTheConditionalStepsStillRefuseAnotherSymbol — the rail this change does not
// touch. A request for something other than what the list carries is still refused,
// nothing is sent, and the run stops.
func TestTheConditionalStepsStillRefuseAnotherSymbol(t *testing.T) {
	p := Plan{Mutations: []PlannedMutation{{
		Step: StepConditionalModify, Kind: MutateModifyConditional,
		Symbol: "333430", Side: "sell", Quantity: "1 share", MaxQuantity: 1,
	}}}

	if p.Authorises(StepConditionalModify, MutateModifyConditional, "005930", "sell", 1) {
		t.Error("the plan authorised a modify for a symbol it does not carry")
	}
	if p.Authorises(StepConditionalModify, MutateModifyConditional, "333430", "sell", 2) {
		t.Error("the plan authorised more shares than the line's ceiling")
	}
	if !p.Authorises(StepConditionalModify, MutateModifyConditional, "333430", "sell", 1) {
		t.Error("the plan refused the request it was built from")
	}
}

// TestEveryMutatingStepThatActsOnTheLiveConditionalDeclaresIt is the drift guard.
//
// The defect this change fixes was exactly a catalogue that did not say what its
// step bodies do: conditional-modify and conditional-cancel resolved their target
// through liveConditional while the catalogue let the plan resolve it through the
// run's probe symbol. Nothing connected the two, so nothing failed until the steps
// became reachable. This reads the connection out of the source: a mutating step
// whose body reads the live conditional has to declare ActsOnConditional, and a step
// that declares it has to be one that reads it.
func TestEveryMutatingStepThatActsOnTheLiveConditionalDeclaresIt(t *testing.T) {
	fset := token.NewFileSet()

	steps, err := parser.ParseFile(fset, "steps.go", mustRead(t, "steps.go"), 0)
	if err != nil {
		t.Fatalf("parse steps.go: %v", err)
	}
	catalogue, err := parser.ParseFile(fset, "verifylive.go", mustRead(t, "verifylive.go"), 0)
	if err != nil {
		t.Fatalf("parse verifylive.go: %v", err)
	}

	// StepConditionalModify -> "conditional-modify", read off the const block so
	// this test carries no hand-kept copy of the catalogue's identifiers.
	byConst := map[string]StepID{}
	ast.Inspect(catalogue, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		ident, ok := spec.Type.(*ast.Ident)
		if !ok || ident.Name != "StepID" || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		lit, ok := spec.Values[0].(*ast.BasicLit)
		if !ok {
			return true
		}
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		byConst[spec.Names[0].Name] = StepID(value)
		return true
	})
	if len(byConst) == 0 {
		t.Fatal("no StepID constants found; this test cannot check anything")
	}

	// The dispatch switch is the only place a step is bound to its body.
	handlerOf := map[string]StepID{}
	for _, decl := range steps.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "dispatch" || fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			clause, ok := n.(*ast.CaseClause)
			if !ok {
				return true
			}
			var ids []StepID
			for _, label := range clause.List {
				if ident, ok := label.(*ast.Ident); ok {
					if id, known := byConst[ident.Name]; known {
						ids = append(ids, id)
					}
				}
			}
			ast.Inspect(clause, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				for _, id := range ids {
					handlerOf[sel.Sel.Name] = id
				}
				return true
			})
			return true
		})
	}
	if len(handlerOf) == 0 {
		t.Fatal("dispatch bound no step to a body; this test cannot check anything")
	}

	reads := map[StepID]bool{}
	for _, decl := range steps.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		id, bound := handlerOf[fn.Name.Name]
		if !bound {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "liveConditional" {
				reads[id] = true
			}
			return true
		})
	}
	if !reads[StepConditionalModify] || !reads[StepConditionalCancel] {
		t.Fatalf("the scan did not find the two step bodies known to read the live conditional: %+v", reads)
	}

	for _, step := range Steps() {
		switch {
		case step.Mutates && reads[step.ID] && !step.ActsOnConditional:
			t.Errorf("%s sends a live request against the registered conditional order but the catalogue "+
				"does not declare ActsOnConditional, so the approval list will name the run's probe "+
				"symbol instead", step.ID)
		case step.ActsOnConditional && !reads[step.ID]:
			t.Errorf("%s declares ActsOnConditional but its body never reads the live conditional; the "+
				"plan would name an object the step does not touch", step.ID)
		}
	}
}

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}
