package execgw_test

// modegate_test.go is the enforcement half of add-core-domain task 3.1: the
// operating mode reaches the sealed submission sequence as an EntryGate latch,
// and the sequence itself does not change.
//
// The tests deliberately go through *the gateway*, not through the gate. A test
// that only asked CheckEntry would pass even if nothing consulted the gate at
// submission time, and "the mode is enforced" is a claim about the submission
// path.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// modeAuditor is the audit sink a relaxation needs.
type modeAuditor struct{ lines []string }

func (a *modeAuditor) RecordAction(action, setting, value, detail string) error {
	a.lines = append(a.lines, action+"|"+setting+"|"+value+"|"+detail)
	return nil
}

func openJournalAt(t *testing.T, clk clock.Clock, path string) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open(%s): %v", path, err)
	}
	return j
}

// modeRig wires the three parts the way the engine does: the journal is the
// authority, the gate is its projection, and the gateway consults the gate.
type modeRig struct {
	gw     *execgw.Gateway
	j      *journal.Journal
	gate   *execgw.EntryGate
	clk    *clock.Fake
	broker *fakeBroker
}

func newModeRig(t *testing.T, path string, result domain.MutationResult) modeRig {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournalAt(t, clk, path)
	t.Cleanup(func() { _ = j.Close() })

	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	if err := j.SetModeProjector(gate); err != nil {
		t.Fatalf("SetModeProjector: %v", err)
	}
	broker := &fakeBroker{result: result}
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker),
		Clock: clk, AccountRef: "acct-7", Source: "test", Entry: gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return modeRig{gw: gw, j: j, gate: gate, clk: clk, broker: broker}
}

func placeUnder(t *testing.T, rig modeRig, symbol string) (execgw.Outcome, error) {
	t.Helper()
	intent := placeIntent()
	intent.Symbol = symbol
	return rig.gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   intent,
		Decision: entryDecision(t, rig.j, rig.clk, intent, testLimits()),
	})
}

// TestAnAutomaticTighteningRefusesTheNextPlace is the whole of "모드 전환 = 영속 +
// EntryGate 투영, landed checkEntry가 소비": nothing in this test touches the
// gate, and the next place is refused.
func TestAnAutomaticTighteningRefusesTheNextPlace(t *testing.T) {
	rig := newModeRig(t, filepath.Join(t.TempDir(), "journal.db"),
		domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"})

	if _, err := placeUnder(t, rig, "005930"); err != nil {
		t.Fatalf("precondition: NORMAL trades: %v", err)
	}

	if _, changed, err := rig.j.EscalateOperatingMode(context.Background(), "acct-7",
		journal.ModeTriggerDailyLossLimit, nil); err != nil || !changed {
		t.Fatalf("escalating: changed=%v err=%v", changed, err)
	}

	before, _, _ := rig.broker.totals()
	_, err := placeUnder(t, rig, "000660")

	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("err = %v, want a gateway refusal", err)
	}
	if rejected.Reason != execgw.ReasonOperatingModeBlocked {
		t.Fatalf("reason = %s, want %s", rejected.Reason, execgw.ReasonOperatingModeBlocked)
	}
	// The refusal travels the ordinary latch path: nothing was sent.
	if after, _, _ := rig.broker.totals(); after != before {
		t.Fatalf("the broker was called %d times during a blocked mode", after-before)
	}
	// The detail names the mode and its cause, so an operator reading a refused
	// attempt does not have to go and look the mode up.
	for _, want := range []string{journal.ModeEntryBlocked, journal.ModeTriggerDailyLossLimit} {
		if !strings.Contains(rejected.Detail, want) {
			t.Fatalf("detail %q does not name %s", rejected.Detail, want)
		}
	}
}

// TestHaltAllStillPermitsAnExit is the RISK_REDUCING row of the mode×class table,
// through the gateway: 청산은 세 모드 전부 허용, 수동 flatten-all 포함(§0.3).
func TestHaltAllStillPermitsAnExit(t *testing.T) {
	rig := newModeRig(t, filepath.Join(t.TempDir(), "journal.db"),
		domain.MutationResult{Kind: "cancel", Status: "accepted", OrderID: "O-9"})

	if _, _, err := rig.j.TransitionOperatingMode(context.Background(), journal.TransitionModeRequest{
		AccountRef: "acct-7", Mode: journal.ModeHaltAll, Actor: journal.ModeActorOperator,
		Cause: "operator halted the account", Auditor: &modeAuditor{},
	}); err != nil {
		t.Fatalf("entering HALT_ALL: %v", err)
	}
	if _, blocked := rig.gate.OperatingModeBlocked(); !blocked {
		t.Fatal("HALT_ALL did not project onto the gate")
	}

	out, err := rig.gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, rig.j, rig.clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	})
	if err != nil {
		t.Fatalf("HALT_ALL must never trap a position (§0.3): %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Fatalf("state = %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
}

// TestARelaxationClearsTheLatch: the projection runs in the release direction
// too, or an approved relaxation would leave the account blocked with no latch
// anybody could point at.
func TestARelaxationClearsTheLatch(t *testing.T) {
	rig := newModeRig(t, filepath.Join(t.TempDir(), "journal.db"),
		domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-2"})
	ctx := context.Background()

	if _, _, err := rig.j.EscalateOperatingMode(ctx, "acct-7",
		journal.ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	if _, blocked := rig.gate.OperatingModeBlocked(); !blocked {
		t.Fatal("the escalation did not latch the gate")
	}

	auditor := &modeAuditor{}
	if _, changed, err := rig.j.TransitionOperatingMode(ctx, journal.TransitionModeRequest{
		AccountRef: "acct-7", Mode: journal.ModeNormal, Actor: journal.ModeActorOperator,
		Cause: "credential rotated and verified", Approval: "sre-1", Auditor: auditor,
	}); err != nil || !changed {
		t.Fatalf("relaxing: changed=%v err=%v", changed, err)
	}
	if detail, blocked := rig.gate.OperatingModeBlocked(); blocked {
		t.Fatalf("the latch survived the relaxation: %s", detail)
	}
	if _, err := placeUnder(t, rig, "005930"); err != nil {
		t.Fatalf("trading did not resume after the relaxation: %v", err)
	}
	if len(auditor.lines) != 1 {
		t.Fatalf("audit lines = %v, want exactly one", auditor.lines)
	}
}

// TestTheRestartRebuildsTheModeLatch: a process that stopped in ENTRY_BLOCKED
// must not come back trading. The journal row is the authority and the new gate
// starts empty, so the restore is what closes the gap.
func TestTheRestartRebuildsTheModeLatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	first := newModeRig(t, path, domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"})
	if _, _, err := first.j.EscalateOperatingMode(ctx, "acct-7",
		journal.ModeTriggerExitObservationOutage, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	if err := first.j.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	restarted := newModeRig(t, path, domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-2"})
	if _, blocked := restarted.gate.OperatingModeBlocked(); blocked {
		t.Fatal("opening the journal latched by itself; the restore must be an explicit startup step")
	}

	if _, err := restarted.j.RestoreOperatingModeProjection(ctx, "acct-7"); err != nil {
		t.Fatalf("RestoreOperatingModeProjection: %v", err)
	}
	_, err := placeUnder(t, restarted, "005930")
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonOperatingModeBlocked {
		t.Fatalf("a restart lost the mode block: err = %v", err)
	}
}

// TestTheModeLatchIsReplacedRatherThanAccumulated: NORMAL → ENTRY_BLOCKED →
// HALT_ALL leaves one latch whose detail describes the mode actually in force.
func TestTheModeLatchIsReplacedRatherThanAccumulated(t *testing.T) {
	rig := newModeRig(t, filepath.Join(t.TempDir(), "journal.db"), domain.MutationResult{})
	ctx := context.Background()

	if _, _, err := rig.j.EscalateOperatingMode(ctx, "acct-7",
		journal.ModeTriggerDailyLossLimit, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	if _, _, err := rig.j.TransitionOperatingMode(ctx, journal.TransitionModeRequest{
		AccountRef: "acct-7", Mode: journal.ModeHaltAll, Actor: journal.ModeActorOperator,
		Cause: "operator halted after the loss limit", Auditor: &modeAuditor{},
	}); err != nil {
		t.Fatalf("tightening to HALT_ALL: %v", err)
	}

	detail, blocked := rig.gate.OperatingModeBlocked()
	if !blocked {
		t.Fatal("HALT_ALL is not latched")
	}
	if !strings.Contains(detail, journal.ModeHaltAll) {
		t.Fatalf("latch detail %q describes a mode that is no longer in force", detail)
	}
	blocks := rig.gate.Blocks()
	if len(blocks) != 1 {
		t.Fatalf("the gate carries %d latches, want one: %v", len(blocks), blocks)
	}
}

// TestTheModeSnapshotIsTheSharedView: Gateway, Guardian and flatten read one
// value, so they cannot disagree about the same account.
func TestTheModeSnapshotIsTheSharedView(t *testing.T) {
	rig := newModeRig(t, filepath.Join(t.TempDir(), "journal.db"), domain.MutationResult{})
	ctx := context.Background()

	if _, _, err := rig.j.EscalateOperatingMode(ctx, "acct-7",
		journal.ModeTriggerCriticalAlertUndelivered, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	snapshot, err := rig.j.CurrentOperatingMode(ctx, "acct-7")
	if err != nil {
		t.Fatalf("CurrentOperatingMode: %v", err)
	}

	// The Gateway's view (the projected latch) and the read surface agree…
	_, latched := rig.gate.OperatingModeBlocked()
	if snapshot.BlocksEntry() != latched {
		t.Fatalf("snapshot says BlocksEntry=%v, the gate says %v", snapshot.BlocksEntry(), latched)
	}
	// …the Guardian issuer's view is the same value, mapped onto the chain's
	// mode field…
	if snapshot.Mode != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s", snapshot.Mode)
	}
	// …and flatten's question has one answer in every mode.
	if !snapshot.Allows(journal.SafetyClassRiskReducing) {
		t.Fatal("a mode refused a risk-reducing mutation (§0.3)")
	}
}
