package obs_test

// mode_test.go covers add-core-domain task 3.3: the transition alert, the
// structured log line, and the loop that has to terminate between the two.

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// TestAModeTransitionIsCritical: the grading, stated where a reader of the table
// will look for it.
func TestAModeTransitionIsCritical(t *testing.T) {
	if got := obs.SeverityOf(obs.EventOperatingMode); got != obs.SeverityCritical {
		t.Fatalf("SeverityOf(%s) = %s, want critical — an automatic tightening that "+
			"nobody is told about is an engine that silently stopped trading",
			obs.EventOperatingMode, got)
	}
}

// TestTheTransitionAlertIsDurableAndNamesBothModes runs the announcement through
// the whole notifier: outbox row first, then the send, with the fields an
// operator needs to act without opening the journal.
func TestTheTransitionAlertIsDurableAndNamesBothModes(t *testing.T) {
	pub := &failingPublisher{}
	n, j, gate := newNotifier(t, pub)
	ctx := context.Background()

	err := n.AnnounceOperatingMode(ctx, journal.ModeNormal, journal.OperatingModeRecord{
		AccountRef: "acct-7",
		Mode:       journal.ModeEntryBlocked,
		Cause:      journal.ModeTriggerDailyLossLimit,
		Actor:      journal.ModeActorAuto,
		CreatedAt:  obsNow,
	})
	if err != nil {
		t.Fatalf("AnnounceOperatingMode: %v", err)
	}

	if pub.callCount() != 1 {
		t.Fatalf("publish calls = %d, want 1", pub.callCount())
	}
	pub.mu.Lock()
	msg := pub.messages[0]
	pub.mu.Unlock()

	if msg.Severity != obs.SeverityCritical {
		t.Fatalf("severity = %s, want critical", msg.Severity)
	}
	for _, want := range []string{journal.ModeNormal, journal.ModeEntryBlocked} {
		if !strings.Contains(msg.Title, want) {
			t.Fatalf("title %q does not name %s; an operator has to know what it moved from", msg.Title, want)
		}
	}
	for _, want := range []string{
		journal.ModeTriggerDailyLossLimit, // why
		journal.ModeActorAuto,             // who
		"risk-reducing",                   // what is still allowed (§0.3)
	} {
		if !strings.Contains(msg.Body, want) {
			t.Fatalf("body %q does not carry %q", msg.Body, want)
		}
	}

	// Durable: delivered, so nothing is left pending, and no gate latch.
	remaining, err := j.UndeliveredCount(ctx)
	if err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("undelivered = %d after a successful send", remaining)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a delivered alert blocked entries: %v", rejected)
	}
}

// TestTheTransitionLogLineIsCountable: the variable parts are fields, not prose,
// because every question about modes is a counting question — how many automatic
// tightenings today, which trigger fires most.
func TestTheTransitionLogLineIsCountable(t *testing.T) {
	var buf bytes.Buffer
	n := &obs.Notifier{
		Log: obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true, Clock: clock.NewFake(obsNow)}),
	}

	if err := n.AnnounceOperatingMode(context.Background(), journal.ModeEntryBlocked,
		journal.OperatingModeRecord{
			AccountRef: "acct-7",
			Mode:       journal.ModeNormal,
			Cause:      "credential rotated | approved-by: sre-1",
			Actor:      journal.ModeActorOperator,
			CreatedAt:  obsNow,
		}); err != nil {
		t.Fatalf("AnnounceOperatingMode: %v", err)
	}

	var line map[string]any
	first := strings.SplitN(strings.TrimSpace(buf.String()), "\n", 2)[0]
	if err := json.Unmarshal([]byte(first), &line); err != nil {
		t.Fatalf("the log line is not JSON (%q): %v", first, err)
	}
	want := map[string]string{
		obs.FieldEvent:     string(obs.EventOperatingMode),
		obs.FieldAccount:   "acct-7",
		obs.FieldFromState: journal.ModeEntryBlocked,
		obs.FieldToState:   journal.ModeNormal,
		obs.FieldActor:     journal.ModeActorOperator,
	}
	for field, value := range want {
		got, ok := line[field]
		if !ok {
			t.Fatalf("the line has no %q field: %v", field, line)
		}
		if got != value {
			t.Fatalf("%s = %v, want %q", field, got, value)
		}
	}
	if reason, _ := line[obs.FieldReason].(string); !strings.Contains(reason, "approved-by: sre-1") {
		t.Fatalf("reason = %q, want the approval that justified the relaxation", reason)
	}
}

// TestARelaxationIsAnnouncedAsARelaxation: the direction is in the alert, because
// "the mode changed" is the one thing an operator already knows and the direction
// is what decides whether they have to do something.
func TestARelaxationIsAnnouncedAsARelaxation(t *testing.T) {
	pub := &failingPublisher{}
	n, _, _ := newNotifier(t, pub)

	if err := n.AnnounceOperatingMode(context.Background(), journal.ModeHaltAll,
		journal.OperatingModeRecord{
			AccountRef: "acct-7", Mode: journal.ModeEntryBlocked,
			Cause: "incident closed | approved-by: sre-1", Actor: journal.ModeActorOperator,
		}); err != nil {
		t.Fatalf("AnnounceOperatingMode: %v", err)
	}
	pub.mu.Lock()
	msg := pub.messages[0]
	pub.mu.Unlock()

	if !strings.Contains(msg.Title, "relaxed") {
		t.Fatalf("title %q does not say the mode was relaxed", msg.Title)
	}
	// …and the entry permission it implies is spelled out, both ways.
	if !strings.Contains(msg.Body, "refused") {
		t.Fatalf("body %q does not say ENTRY_BLOCKED still refuses entries", msg.Body)
	}
}

// TestTheAlertEscalationLoopTerminates is the property named in mode.go's header.
//
// Announcing a tightening can fail to deliver; a failed critical delivery latches
// the gate and is itself an escalation trigger. If that escalation produced
// another transition, it would announce, fail, and escalate again. It does not,
// because the account is already in the mode the trigger targets and the no-op
// rule appends nothing — so this test is really an assertion about
// journal.TransitionOperatingMode, exercised from the side that would loop.
func TestTheAlertEscalationLoopTerminates(t *testing.T) {
	pub := &failingPublisher{fail: true}
	n, j, gate := newNotifier(t, pub)
	ctx := context.Background()
	const account = "acct-7"

	if err := j.SetModeProjector(gate); err != nil {
		t.Fatalf("SetModeProjector: %v", err)
	}

	// A daily-loss trigger escalates, and the journal announces through the
	// notifier as part of the same flow.
	//
	// The announcement returns nil even though every send failed: the Notifier's
	// contract is that an undelivered *critical* alert is already handled — by
	// latching the gate — and only a failed outbox write is the caller's problem.
	// That is what keeps a broken notification transport from looking like a
	// broken mode transition.
	_, changed, err := j.EscalateOperatingMode(ctx, account, journal.ModeTriggerDailyLossLimit, n)
	if err != nil {
		t.Fatalf("EscalateOperatingMode: %v", err)
	}
	if !changed {
		t.Fatal("an undelivered alert must not undo the transition; the mode is the safety mechanism")
	}
	// The transition stands, durably, and the gate carries both latches.
	if snapshot, err := j.CurrentOperatingMode(ctx, account); err != nil ||
		snapshot.Mode != journal.ModeEntryBlocked {
		t.Fatalf("mode = %+v err=%v, want a durable ENTRY_BLOCKED", snapshot, err)
	}
	if _, blocked := gate.OperatingModeBlocked(); !blocked {
		t.Fatal("the mode did not project")
	}
	if _, undelivered := gate.Blocks()[execgw.ReasonAlertUndelivered]; !undelivered {
		t.Fatal("a critical alert that could not be delivered must latch the gate")
	}

	// Now the second lap: sustained outbox failure is an escalation trigger.
	before, err := j.OperatingModeHistory(ctx, account)
	if err != nil {
		t.Fatalf("OperatingModeHistory: %v", err)
	}
	_, changedAgain, err := j.EscalateOperatingMode(ctx, account,
		journal.ModeTriggerCriticalAlertUndelivered, n)
	if err != nil {
		t.Fatalf("the second escalation errored: %v", err)
	}
	if changedAgain {
		t.Fatal("the outbox-failure trigger produced a second transition; the loop would not close")
	}
	after, _ := j.OperatingModeHistory(ctx, account)
	if len(after) != len(before) {
		t.Fatalf("history grew from %d to %d rows on a no-op", len(before), len(after))
	}
}

// TestANoOpTransitionAnnouncesNothing is the same property from the journal's
// side, without a notifier that can fail.
func TestANoOpTransitionAnnouncesNothing(t *testing.T) {
	pub := &failingPublisher{}
	n, j, _ := newNotifier(t, pub)
	ctx := context.Background()

	if _, changed, err := j.EscalateOperatingMode(ctx, "acct-7",
		journal.ModeTriggerDailyLossLimit, n); err != nil || !changed {
		t.Fatalf("first escalation: changed=%v err=%v", changed, err)
	}
	if pub.callCount() != 1 {
		t.Fatalf("publish calls = %d after one transition, want 1", pub.callCount())
	}

	for i := 0; i < 4; i++ {
		if _, changed, err := j.EscalateOperatingMode(ctx, "acct-7",
			journal.ModeTriggerDailyLossLimit, n); err != nil || changed {
			t.Fatalf("repeat %d: changed=%v err=%v", i, changed, err)
		}
	}
	if pub.callCount() != 1 {
		t.Fatalf("publish calls = %d; a trigger that fires every poll cycle would page "+
			"an operator every poll cycle", pub.callCount())
	}
}

// TestAnnouncingWithoutANotifierIsSafe: a nil announcer is how an engine wired
// without alerting behaves, and the transition must not depend on the transport.
func TestAnnouncingWithoutANotifierIsSafe(t *testing.T) {
	var n *obs.Notifier
	if err := n.AnnounceOperatingMode(context.Background(), journal.ModeNormal,
		journal.OperatingModeRecord{AccountRef: "acct-7", Mode: journal.ModeEntryBlocked}); err != nil {
		t.Fatalf("a nil notifier must be a no-op, got %v", err)
	}
}
