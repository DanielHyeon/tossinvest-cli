package engine_test

// a064 task 3: the moment protection stops is an event of its own.
//
// Before this change the only trace a quarantine left was the *next* cycle's
// judgement refusal. In the operational ledger on 2026-08-03 that cost five
// seconds of silence three times over, and it cost the identity of the
// quarantine every time: the refusal alert carries no version, no reason and no
// evidence, so nothing in it says whether a human has to lift something.

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// openLadderState is the shared fixture: one held position under the common
// ladder, with one clean cycle behind it so a canonical snapshot exists.
func openLadderState(t *testing.T, h *exitHarness, symbol string) journal.Position {
	t.Helper()
	p := h.entry(symbol, "10", "70000", "68000", "70000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder,
		EntryPrice: "70000", InitialStop: "68000",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	return p
}

func openLedger(t *testing.T, h *exitHarness) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+h.journal.Path())
	if err != nil {
		t.Fatalf("opening the ledger for the fixture: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func quarantineAlerts(h *exitHarness) []obs.Event {
	var out []obs.Event
	for _, e := range h.alerts.events {
		if e.Type == obs.EventExitSnapshotQuarantined {
			out = append(out, e)
		}
	}
	return out
}

// TestALegacyIdentityQuarantineIsAnnouncedWhenItIsCreated covers workingSet B18.
func TestALegacyIdentityQuarantineIsAnnouncedWhenItIsCreated(t *testing.T) {
	h := newExitHarness(t, nil)
	p := openLadderState(t, h, "005930")
	db := openLedger(t, h)
	// The fixture from exit_snapshot_isolation_test.go: a legacy row whose policy
	// id has no pinned meaning, which is what managedPolicyIdentity refuses.
	if _, err := db.Exec(`UPDATE exit_states SET policy_id='UNKNOWN_LEGACY', snapshot_status=NULL,
		policy_version=NULL, policy_digest=NULL, snapshot_id=NULL, decision_id=NULL,
		observation_id=NULL, position_generation=NULL, next_target=NULL, next_protection=NULL,
		last_observation_source=NULL, last_observed_at=NULL, snapshot_action=NULL,
		snapshot_ratio=NULL, projected_quantity=NULL, state_only=NULL, suppressed_reason=NULL,
		effective_snapshot_json=NULL WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}

	h.quote("005930", 70500)
	h.observe()

	events := quarantineAlerts(h)
	if len(events) != 1 {
		t.Fatalf("quarantine announcements = %d, want exactly one", len(events))
	}
	if got := events[0].Fields["reason"]; got != "legacy_policy_identity_unknown" {
		t.Fatalf("reason = %v", got)
	}
}

// TestACorruptSnapshotQuarantineIsAnnouncedWhenItIsCreated covers workingSet B11.
func TestACorruptSnapshotQuarantineIsAnnouncedWhenItIsCreated(t *testing.T) {
	h := newExitHarness(t, nil)
	p := openLadderState(t, h, "005930")
	h.quote("005930", 70500)
	h.observe()

	// A v10 tuple with its status column missing is the "partial_snapshot_tuple"
	// corruption: the row proves a snapshot was written and cannot say what it is.
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET snapshot_status=NULL WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}

	// Three cycles, not one. The corruption check runs *before* the active
	// quarantine check, so this path calls QuarantineExitSnapshot on every cycle
	// and gets the existing row back each time — it is where the announcement
	// latch actually carries load. Without it this would alert every five seconds
	// forever, and with no publisher wired each of those is a gate block.
	for i := 0; i < 3; i++ {
		h.quote("005930", float64(70800+i*100))
		h.observe()
	}

	events := quarantineAlerts(h)
	if len(events) != 1 {
		t.Fatalf("quarantine announcements = %d, want exactly one across three cycles", len(events))
	}
	if got := events[0].Fields["reason"]; got != "stored_snapshot_corrupt" {
		t.Fatalf("reason = %v", got)
	}
}

// TestAnAmbiguousRecoveryQuarantineIsAnnouncedInTheSameCycle covers record B9 —
// the path all three live quarantines came through.
func TestAnAmbiguousRecoveryQuarantineIsAnnouncedInTheSameCycle(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := openLadderState(t, h, "005930")
	h.quote("005930", 70500)
	h.observe()

	// The two records of the entry price now disagree. The stored snapshot is
	// untouched and still verifies against its own derivation; what changed is
	// the column the *next* evaluation recomputes from. That is exactly what
	// "recovery candidate identity mismatch" means, and it is reached without
	// forging a snapshot — which a063 (issues I4) established is unreachable.
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET entry_price='69000' WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}

	h.quote("005930", 70900)
	cycle := h.observe()

	events := quarantineAlerts(h)
	if len(events) != 1 {
		t.Fatalf("quarantine announcements = %d, want exactly one in the creating cycle: cycle=%+v",
			len(events), cycle)
	}
	if got := events[0].Fields["reason"]; got != "ambiguous_recovery" {
		t.Fatalf("reason = %v", got)
	}
	// The ledger agrees the quarantine is real and the announcement names its
	// version, so an operator can take it straight to the release screen.
	q, active, err := h.journal.ActiveExitSnapshotQuarantine(context.Background(), p.ID, p.InstanceSeq)
	if err != nil || !active {
		t.Fatalf("active quarantine = %v err=%v", active, err)
	}
	if got := events[0].Fields["quarantine_version"]; got != q.Version {
		t.Fatalf("announced version = %v, ledger version = %d", got, q.Version)
	}
}

// TestAQuarantineAnnouncementCarriesItsIdentity covers task 3.2.
func TestAQuarantineAnnouncementCarriesItsIdentity(t *testing.T) {
	h := newExitHarness(t, nil)
	p := openLadderState(t, h, "005930")
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET policy_id='UNKNOWN_LEGACY', snapshot_status=NULL,
		policy_version=NULL, policy_digest=NULL, snapshot_id=NULL, decision_id=NULL,
		observation_id=NULL, position_generation=NULL, effective_snapshot_json=NULL
		WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 70500)
	h.observe()

	events := quarantineAlerts(h)
	if len(events) != 1 {
		t.Fatalf("quarantine announcements = %d", len(events))
	}
	event := events[0]
	for _, field := range []string{
		obs.FieldSymbol, "position_id", "position_generation",
		"quarantine_version", "reason", "evidence", "quarantined_at",
	} {
		if _, ok := event.Fields[field]; !ok {
			t.Errorf("announcement is missing field %q: %+v", field, event.Fields)
		}
	}
	if event.Fields["position_id"] != p.ID {
		t.Errorf("position_id = %v, want %s", event.Fields["position_id"], p.ID)
	}
	if event.Fields["position_generation"] != p.InstanceSeq {
		t.Errorf("generation = %v, want %d", event.Fields["position_generation"], p.InstanceSeq)
	}
	// The key is what the outbox deduplicates on, so it has to distinguish two
	// quarantines of the same position.
	if !strings.Contains(event.Key, p.ID) {
		t.Errorf("event key %q does not name the position", event.Key)
	}
}

// TestAnAlreadyActiveQuarantineIsNotAnnouncedAgain covers workingSet B15/B17 and
// task 3.7.
func TestAnAlreadyActiveQuarantineIsNotAnnouncedAgain(t *testing.T) {
	h := newExitHarness(t, nil)
	p := openLadderState(t, h, "005930")
	ctx := context.Background()
	if _, err := h.journal.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq,
		"ambiguous_recovery", "exitpolicy: recovery candidate identity mismatch"); err != nil {
		t.Fatalf("QuarantineExitSnapshot: %v", err)
	}

	for i := 0; i < 3; i++ {
		h.quote("005930", float64(70500+i*100))
		h.observe()
	}

	if events := quarantineAlerts(h); len(events) != 0 {
		t.Fatalf("a quarantine this loop did not create was announced %d time(s)", len(events))
	}
	// The refusal alert is still raised: nothing about the announcement replaces
	// the existing contract.
	if _, ok := h.alerts.first(obs.EventExitJudgementRefused); !ok {
		t.Fatal("a quarantined position lost its judgement-refused alert")
	}
}

// TestANewQuarantineVersionIsAnnouncedAgain covers task 3.8 and, with it, the
// case where the judgement-refusal latch is already set — which is precisely how
// a quarantine could previously happen with no line at all.
func TestANewQuarantineVersionIsAnnouncedAgain(t *testing.T) {
	h := newExitHarness(t, nil)
	p := openLadderState(t, h, "005930")
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET policy_id='UNKNOWN_LEGACY', snapshot_status=NULL,
		policy_version=NULL, policy_digest=NULL, snapshot_id=NULL, decision_id=NULL,
		observation_id=NULL, position_generation=NULL, effective_snapshot_json=NULL
		WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 70500)
	h.observe()
	if len(quarantineAlerts(h)) != 1 {
		t.Fatalf("first quarantine was not announced once: %d", len(quarantineAlerts(h)))
	}
	refusalsAfterFirst := h.alerts.count(obs.EventExitJudgementRefused)

	// a063's release, then the same unresolved cause quarantines it again under a
	// new version. The judgement-refusal latch is still set from the first round.
	ctx := context.Background()
	if err := h.journal.ReleaseExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq, 1,
		journal.QuarantineReleaseHumanRepair, "LOCAL_OPERATOR released quarantine v1"); err != nil {
		t.Fatalf("ReleaseExitSnapshotQuarantine: %v", err)
	}
	h.quote("005930", 70800)
	h.observe()

	events := quarantineAlerts(h)
	if len(events) != 2 {
		t.Fatalf("announcements = %d, want a second one for the new version", len(events))
	}
	if events[1].Fields["quarantine_version"] == events[0].Fields["quarantine_version"] {
		t.Fatalf("the second announcement carries the first version: %v", events[1].Fields)
	}
	if h.alerts.count(obs.EventExitJudgementRefused) != refusalsAfterFirst {
		t.Fatal("fixture drifted: the judgement-refusal latch was expected to still be set")
	}
}

// TestAnnouncingAQuarantineDoesNotChangeTheWorkingSet covers Pre-Edit 선언 3.
func TestAnnouncingAQuarantineDoesNotChangeTheWorkingSet(t *testing.T) {
	run := func(t *testing.T, withAlerts bool) (engine.ExitCycle, journal.ExitState) {
		t.Helper()
		h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
			if !withAlerts {
				o.Alerts = nil
			}
		})
		p := openLadderState(t, h, "005930")
		db := openLedger(t, h)
		if _, err := db.Exec(`UPDATE exit_states SET policy_id='UNKNOWN_LEGACY', snapshot_status=NULL,
			policy_version=NULL, policy_digest=NULL, snapshot_id=NULL, decision_id=NULL,
			observation_id=NULL, position_generation=NULL, effective_snapshot_json=NULL
			WHERE position_id=?`, p.ID); err != nil {
			t.Fatal(err)
		}
		h.quote("005930", 70500)
		return h.observe(), h.state(p.ID)
	}

	withAlerts, stateWith := run(t, true)
	without, stateWithout := run(t, false)
	if withAlerts != without {
		t.Fatalf("the cycle differs with alerts wired:\n with    = %+v\n without = %+v", withAlerts, without)
	}
	if stateWith.HighWater != stateWithout.HighWater || stateWith.Baseline != stateWithout.Baseline ||
		stateWith.ActiveRung != stateWithout.ActiveRung {
		t.Fatalf("the stored state differs with alerts wired:\n with    = %+v\n without = %+v",
			stateWith, stateWithout)
	}
}

// TestAQuarantineAnnouncementWritesNothing is Pre-Edit 선언 3 at the record path:
// announcing is observation, and observation that writes is not observation.
func TestAQuarantineAnnouncementWritesNothing(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := openLadderState(t, h, "005930")
	h.quote("005930", 70500)
	h.observe()
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET entry_price='69000' WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}

	before := h.state(p.ID)
	h.quote("005930", 70900)
	h.observe()
	after := h.state(p.ID)

	if len(quarantineAlerts(h)) != 1 {
		t.Fatalf("fixture drifted: announcements = %d", len(quarantineAlerts(h)))
	}
	if before.HighWater != after.HighWater || before.Baseline != after.Baseline ||
		before.ActiveRung != after.ActiveRung || before.RatchetLevel != after.RatchetLevel {
		t.Fatalf("the announcement moved the protection state:\n before = %+v\n after  = %+v", before, after)
	}
	if len(h.submit.places) != 0 {
		t.Fatalf("a quarantined position reached the order path: %+v", h.submit.places)
	}
}

// TestAQuarantineIsAnnouncedOnceAcrossBothPaths pins that the latch is shared:
// record announces the creation, and the working set that sees the same row on
// every later cycle does not announce it again.
func TestAQuarantineIsAnnouncedOnceAcrossBothPaths(t *testing.T) {
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) {
		policy := exitpolicy.DefaultLadderPolicy()
		o.Ladder = &policy
	})
	p := openLadderState(t, h, "005930")
	h.quote("005930", 70500)
	h.observe()
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET entry_price='69000' WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		h.quote("005930", float64(70900+i*100))
		h.observe()
	}

	if events := quarantineAlerts(h); len(events) != 1 {
		t.Fatalf("announcements = %d across four cycles, want exactly one", len(events))
	}
}
