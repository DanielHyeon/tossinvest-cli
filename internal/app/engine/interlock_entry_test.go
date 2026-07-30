package engine_test

// interlock_entry_test.go is the half of clause 6 that stayed in the interlock
// (change interlock-gates-entry-not-exit).
//
// The refusal itself moved to the mutation chokepoint (internal/execgw's
// protection_test.go). What the interlock keeps is the *report*: whether entry is
// permitted, and the operator-facing fact that a build with no broker-resident
// protection protects only while this process is alive.
//
// The tests are separate from interlock_test.go's refusal table on purpose. That
// table is about the eight clauses that still refuse a start, and it must stay
// readable as exactly that list.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// TestTheGateStartsWithoutBrokerSideProtection is the inversion this change is.
//
// Every other clause is satisfied and the protective-order marker is not — which
// is every operator's situation until T2.3 lands. Before this change that
// combination produced no engine at all, so the exit observer never ran and the
// account's holdings carried no stop. The clause's own stated purpose was to stop
// *automatic entry*; refusing the runtime went past it.
func TestTheGateStartsWithoutBrokerSideProtection(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	// openGateEngine, not openProtectedGateEngine: no seam, the shipped shape.
	eng, err := openGateEngine(t, dir, srv, matchedGuardian())
	if err != nil {
		t.Fatalf("clauses 1-5 are satisfied, so the runtime must start: %v", err)
	}
	if eng == nil {
		t.Fatal("a verified start must return an engine")
	}
	if !eng.Automation.Verified {
		t.Error("Verified = false: the loops read this, and false means no loop set")
	}
	if eng.Guardian == nil {
		t.Error("the injected Guardian must be published — the exit observer issues through it")
	}
}

// TestAnUnwiredProfileReportsEntryNotPermitted: starting is not permission.
//
// The two facts are now separate fields because they answer different questions.
// "Did the interlock pass" decides whether loops run; "is entry permitted"
// decides what those loops may do, and an operator reading a dashboard has to be
// able to see the second one without inferring it from the absence of a crash.
func TestAnUnwiredProfileReportsEntryNotPermitted(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, matchedGuardian())
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if eng.Automation.EntryPermitted {
		t.Error("EntryPermitted = true on a build with no broker-resident protective execution")
	}
	if eng.Automation.Protection != engine.ProtectionUnwired {
		t.Errorf("Protection = %q, want %q — the marker did not change meaning, only its effect",
			eng.Automation.Protection, engine.ProtectionUnwired)
	}
}

// TestEntryIsPermittedOnlyWhenProtectionIsWired is the same assertion from the
// other side, so that "EntryPermitted is always false" cannot pass for a correct
// implementation.
func TestEntryIsPermittedOnlyWhenProtectionIsWired(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openProtectedGateEngine(t, dir, srv, matchedGuardian())
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if !eng.Automation.EntryPermitted {
		t.Error("EntryPermitted = false with the protective marker satisfied")
	}
}

// TestTheGateOffPathIsUntouched is §0.3 stated as a test.
//
// This change alters what a gate-ON start does. A gate-OFF start must be byte
// identical: same absence of verification, same absence of entry permission, same
// reported marker.
func TestTheGateOffPathIsUntouched(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("a gate-off engine must start: %v", err)
	}
	if eng.Automation.Enabled || eng.Automation.Verified || eng.Automation.EntryPermitted {
		t.Errorf("gate off must verify and permit nothing: %+v", eng.Automation)
	}
	if eng.Automation.Protection != engine.ProtectionUnwired {
		t.Errorf("Protection = %q", eng.Automation.Protection)
	}
}

// TestTheStartupSaysProtectionDiesWithTheProcess is design D6.
//
// A runtime that comes up while protection is unwired is a new state, and one an
// operator can misread as "the broker is holding my stop". The structured
// operating-mode record has carried `protection: UNWIRED` since 5.2, but nobody
// reads a JSON field for reassurance — so the decision line has to say, in words,
// that the stop lives and dies with this process.
//
// It is a sentence, not a prompt. No typed confirmation, no extra approval step.
func TestTheStartupSaysProtectionDiesWithTheProcess(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	var logs strings.Builder
	if _, err := openGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err != nil {
		t.Fatalf("start: %v", err)
	}

	got := logs.String()
	for _, want := range []string{"UNWIRED", "프로세스가 죽으면 보호도 사라진다"} {
		if !strings.Contains(got, want) {
			t.Errorf("the gate decision record does not mention %q; an operator reading it "+
				"cannot tell that the stop does not survive a crash\n%s", want, got)
		}
	}
}
