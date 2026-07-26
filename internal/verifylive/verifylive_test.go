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
		text := strings.ToLower(strings.Join(s.Procedure, " "))
		if !strings.Contains(text, "cancel") {
			t.Errorf("%s mutates but its procedure never mentions cancelling what it creates:\n  %s",
				s.ID, strings.Join(s.Procedure, "\n  "))
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
	first := strings.ToUpper(s.Procedure[0])
	if !strings.Contains(first, "SECOND LIVE ORDER") {
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
