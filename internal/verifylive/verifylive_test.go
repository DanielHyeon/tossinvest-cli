package verifylive

// verifylive_test.go covers the step catalogue itself.
//
// The catalogue is the procedure an operator reads before deciding whether to run
// any of this against their own money, so the properties asserted here are about
// it being *readable and honest*: no step without a stated proof, no mutating
// step without a procedure that says how the exposure ends, no dependency on a
// step that does not exist.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// TestOfficialClientSatisfiesBroker. If this stops compiling, the tool cannot be
// wired to the real API, which is the only thing it is for.
func TestOfficialClientSatisfiesBroker(t *testing.T) {
	var _ Broker = (*official.Client)(nil)
}

func TestStepCatalogueIsWellFormed(t *testing.T) {
	seen := map[StepID]bool{}
	for _, s := range Steps() {
		if s.ID == "" {
			t.Fatalf("a step has no id: %+v", s)
		}
		if seen[s.ID] {
			t.Errorf("%s appears twice in the catalogue", s.ID)
		}
		seen[s.ID] = true

		if strings.TrimSpace(s.Title) == "" {
			t.Errorf("%s has no title", s.ID)
		}
		if strings.TrimSpace(s.Proves) == "" {
			t.Errorf("%s does not say what it proves; an operator cannot judge whether to run it", s.ID)
		}
		if len(s.Tasks) == 0 {
			t.Errorf("%s feeds no measurement task, so nothing downstream consumes it", s.ID)
		}
		if len(s.Procedure) == 0 {
			t.Errorf("%s has no procedure", s.ID)
		}
	}
	for _, s := range Steps() {
		for _, dep := range s.DependsOn {
			if !seen[dep] {
				t.Errorf("%s depends on %s, which is not in the catalogue", s.ID, dep)
			}
		}
	}
}

// TestEveryMutatingStepSaysHowTheExposureEnds.
//
// The justification for placing live orders at all is that each one is cancelled
// inside the step that created it. A mutating step whose procedure never mentions
// cancelling is a step that leaves something behind — and the one deliberate
// exception, the conditional order that has to outlive the process, says so in
// its own words.
func TestEveryMutatingStepSaysHowTheExposureEnds(t *testing.T) {
	for _, s := range Steps() {
		if !s.Mutates {
			continue
		}
		// The procedure is in the display language (task 1.8 ③); "취소" is what
		// "cancel" now reads as, and the property being asserted is unchanged.
		text := strings.ToLower(strings.Join(s.Procedure, " "))
		if !strings.Contains(text, "취소") && !strings.Contains(text, "cancel") {
			t.Errorf("%s mutates but its procedure never mentions cancelling what it creates:\n  %s",
				s.ID, strings.Join(s.Procedure, "\n  "))
		}
	}
}

// TestEveryMutatingStepDeclaresWhatItSends.
//
// Step.Mutations is the step's permission slip: it is what the batch approval
// enumerates and what mutate.go's guard authorises. A step that mutates without
// declaring anything could send nothing at all under the default model, and a step
// that declares something without being marked as mutating would be a mutation the
// catalogue's own [mutating] tag hides.
func TestEveryMutatingStepDeclaresWhatItSends(t *testing.T) {
	for _, s := range Steps() {
		if s.Mutates != (len(s.Mutations) > 0) {
			t.Errorf("%s: Mutates=%v but declares %d mutation(s); the two must agree",
				s.ID, s.Mutates, len(s.Mutations))
		}
		for i, m := range s.Mutations {
			if strings.TrimSpace(string(m.Kind)) == "" {
				t.Errorf("%s mutation %d has no kind", s.ID, i)
			}
			if strings.TrimSpace(m.Ends) == "" {
				t.Errorf("%s mutation %d (%s) does not say how the exposure ends, so it cannot be approved",
					s.ID, i, m.Kind)
			}
			if m.Kind.Verb() == string(m.Kind) {
				t.Errorf("%s mutation %d (%s) has no readable verb, so the summary would print a constant",
					s.ID, i, m.Kind)
			}
			// Anything that creates or moves an order has to say how it is priced;
			// a cancel has nothing to price.
			needsPrice := m.Kind != MutateCancelOrder && m.Kind != MutateCancelConditional
			if needsPrice && m.Pricing == PriceNone {
				t.Errorf("%s mutation %d (%s) states no pricing basis", s.ID, i, m.Kind)
			}
			if !needsPrice && m.Quantity != QuantityNone {
				t.Errorf("%s mutation %d (%s) carries a quantity rule although a cancel has no quantity",
					s.ID, i, m.Kind)
			}
		}
	}
}

// TestOnlyTheConditionalRegistrationOutlivesItsStep. Everything else this tool
// sends is cleaned up where it was created, and the catalogue is where that claim
// has to be true before the runtime can honour it.
func TestOnlyTheConditionalRegistrationOutlivesItsStep(t *testing.T) {
	for _, s := range Steps() {
		for _, m := range s.Mutations {
			outlives := strings.Contains(strings.ToLower(m.Ends), "outlive")
			if outlives && m.Kind != MutateRegisterConditional {
				t.Errorf("%s: a %s mutation claims to outlive its step: %s", s.ID, m.Kind, m.Ends)
			}
			if m.Kind == MutateRegisterConditional && !outlives {
				t.Errorf("%s: the conditional registration does not warn that it is left alive: %s", s.ID, m.Ends)
			}
		}
	}
}

// TestMutationKindsAreDistinctAndReadable — the kinds are the join between the
// catalogue, the summary and the guard, so a duplicate verb would make two
// different requests look like one on the list a person approves.
func TestMutationKindsAreDistinctAndReadable(t *testing.T) {
	kinds := []MutationKind{
		MutatePlaceOrder, MutateReplayOrder, MutateConflictProbe, MutateCancelOrder, MutateAmendOrder,
		MutateRegisterConditional, MutateReplayConditional, MutateModifyConditional, MutateCancelConditional,
	}
	seen := map[string]MutationKind{}
	for _, k := range kinds {
		verb := k.Verb()
		if verb == string(k) {
			t.Errorf("%s has no verb of its own", k)
		}
		if prior, dup := seen[verb]; dup {
			t.Errorf("%s and %s render as the same line: %q", prior, k, verb)
		}
		seen[verb] = k
	}
	// Every kind the catalogue uses is one of the above.
	for _, s := range Steps() {
		for _, m := range s.Mutations {
			if _, ok := seen[m.Kind.Verb()]; !ok {
				t.Errorf("%s declares %s, which is not a known mutation kind", s.ID, m.Kind)
			}
		}
	}
}

// TestTTLEdgeStepIsOptInAndWarnsFirst is the tasks.md requirement in a test:
// "유효 창 경계(2.7)는 의도적 이중 주문 절차임을 단계 안내문에 명시하고 기본 생략".
func TestTTLEdgeStepIsOptInAndWarnsFirst(t *testing.T) {
	s, ok := StepByID(StepIdempotencyTTLEdge)
	if !ok {
		t.Fatal("the validity-window step is not in the catalogue")
	}
	if s.OptIn != FlagIncludeTTLEdge {
		t.Errorf("OptIn = %q, want %q — it must not run by default", s.OptIn, FlagIncludeTTLEdge)
	}
	if !strings.Contains(s.Procedure[0], "두 번째 라이브 주문") {
		t.Errorf("the step's first procedure line must say it creates a second live order, got %q", s.Procedure[0])
	}
}

// TestTriggerObservationIsDeferredNotPromised. Task 2.5 wants the trigger
// observation carried out in a separate session under market conditions this tool
// cannot manufacture. Recording it as deferred is what puts it on task 2.6's
// no-automatic-entry list instead of leaving a silent gap.
func TestTriggerObservationIsDeferredNotPromised(t *testing.T) {
	s, ok := StepByID(StepConditionalTrigger)
	if !ok {
		t.Fatal("the trigger step is not in the catalogue")
	}
	if s.Deferred == "" {
		t.Fatal("the trigger step must be marked deferred")
	}
	if s.Mutates {
		t.Error("the trigger step must not mutate: it is an observation nobody can force")
	}
}

// TestStepsNeedingAHoldingNeverBuyOne.
//
// "never buy-to-create a holding automatically" is the rule; this asserts the
// catalogue does not describe one. A step that needed a holding and offered to
// create it would be a step that opens a position to test a tool.
func TestStepsNeedingAHoldingNeverBuyOne(t *testing.T) {
	for _, s := range Steps() {
		if !s.NeedsHolding {
			continue
		}
		text := strings.ToLower(strings.Join(s.Procedure, " "))
		if strings.Contains(text, "buy") && !strings.Contains(text, "buy-side") {
			t.Errorf("%s needs a holding and its procedure mentions buying: %s", s.ID, text)
		}
	}
}

// TestMutationMethodsMatchTheBrokerInterface keeps the list static_test.go greps
// for in step with the interface it describes.
func TestMutationMethodsMatchTheBrokerInterface(t *testing.T) {
	// The interface's mutation surface, spelled out. Adding a seventh mutation to
	// Broker without adding it to MutationMethods would leave it un-greppable.
	want := []string{
		"PlaceOrder", "CancelOrder", "ModifyOrder",
		"CreateConditionalOrder", "ModifyConditionalOrderRef", "CancelConditionalOrder",
	}
	got := MutationMethods()
	if len(got) != len(want) {
		t.Fatalf("MutationMethods has %d entries, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("MutationMethods[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
