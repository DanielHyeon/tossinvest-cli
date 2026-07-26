package journal

// operating_mode_test.go covers add-core-domain task 3.1: the operating_modes
// repository, the mode×class table, conservative precedence, and the projection
// that makes a transition and its enforcement one act.
//
// The gate-side half (a real *execgw.EntryGate as the projector, and the restart
// that rebuilds it) is internal/execgw/modegate_test.go — from here the projector
// is a recorder, because this package must not import the gateway.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// recordingProjector captures what the journal projected, in order.
type recordingProjector struct {
	mu   sync.Mutex
	seen []OperatingModeRecord
}

func (p *recordingProjector) ProjectOperatingMode(rec OperatingModeRecord) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.seen = append(p.seen, rec)
}

func (p *recordingProjector) records() []OperatingModeRecord {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]OperatingModeRecord, len(p.seen))
	copy(out, p.seen)
	return out
}

func modeJournal(t *testing.T) (*Journal, *recordingProjector) {
	t.Helper()
	j := openTestJournal(t)
	p := &recordingProjector{}
	if err := j.SetModeProjector(p); err != nil {
		t.Fatalf("SetModeProjector: %v", err)
	}
	return j, p
}

func currentMode(t *testing.T, j *Journal, account string) ModeSnapshot {
	t.Helper()
	snapshot, err := j.CurrentOperatingMode(context.Background(), account)
	if err != nil {
		t.Fatalf("CurrentOperatingMode(%s): %v", account, err)
	}
	return snapshot
}

// --- the table ----------------------------------------------------------------

// TestModeClassTableIsComplete enumerates every cell of risk-management's
// mode×class table, including the reserved column.
func TestModeClassTableIsComplete(t *testing.T) {
	cases := []struct {
		mode  string
		class string
		want  bool
		why   string
	}{
		{ModeNormal, SafetyClassExposureRaising, true, "NORMAL 허용"},
		{ModeNormal, SafetyClassRiskReducing, true, "NORMAL 허용"},
		{ModeNormal, SafetyClassProtectionWeakening, true, "NORMAL 허용(audit) — 열 예약"},

		{ModeEntryBlocked, SafetyClassExposureRaising, false, "ENTRY_BLOCKED 거부"},
		{ModeEntryBlocked, SafetyClassRiskReducing, true, "청산은 어느 모드에서도 막히지 않는다 (§0.3)"},
		{ModeEntryBlocked, SafetyClassProtectionWeakening, true, "ENTRY_BLOCKED 허용(audit) — 열 예약"},

		{ModeHaltAll, SafetyClassExposureRaising, false, "HALT_ALL 거부"},
		{ModeHaltAll, SafetyClassRiskReducing, true, "수동 flatten-all은 모든 모드에서 통과한다 (§0.3)"},
		{ModeHaltAll, SafetyClassProtectionWeakening, false, "HALT_ALL의 추가 규칙"},

		// Fail-closed edges: neither an unknown mode nor an unknown class is a
		// permission. A typo in a persisted enum must not read as NORMAL.
		{"EXIT_ONLY", SafetyClassRiskReducing, false, "EXIT_ONLY는 존재하지 않는다 (D3)"},
		{"normal", SafetyClassExposureRaising, false, "대소문자가 다른 값은 다른 값이다"},
		{ModeNormal, "SOMETHING_ELSE", false, "열거 밖 클래스"},
		{"", "", false, "빈 값"},
	}
	for _, tc := range cases {
		if got := ModeAllows(tc.mode, tc.class); got != tc.want {
			t.Errorf("ModeAllows(%q, %q) = %v, want %v — %s", tc.mode, tc.class, got, tc.want, tc.why)
		}
	}
}

// TestProtectionWeakeningIsRefusedByTheDecisionRecorderInEveryMode asserts the
// landed enforcement rather than duplicating it.
//
// The table above reserves the PROTECTION_WEAKENING column and marks two cells
// 허용(audit) — but this build issues none, because RecordDecision refuses the
// class outright. That refusal is stricter than the table in NORMAL and
// ENTRY_BLOCKED and identical to it in HALT_ALL, which is the safe direction; the
// point of this test is that the reserved column cannot quietly become effective
// without somebody changing decision.go and seeing this case fail.
func TestProtectionWeakeningIsRefusedByTheDecisionRecorderInEveryMode(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	const account = "acct-1"

	for _, mode := range []string{ModeNormal, ModeEntryBlocked, ModeHaltAll} {
		if mode != ModeNormal {
			if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
				AccountRef: account, Mode: mode, Actor: ModeActorOperator,
				Cause: "operator drill", Approval: "runbook-7", Auditor: &recordingAuditor{},
			}); err != nil {
				t.Fatalf("entering %s: %v", mode, err)
			}
		}
		_, err := j.RecordDecision(ctx, DecisionRequest{
			ID: "d-pw-" + mode, AccountRef: account, SafetyClass: SafetyClassProtectionWeakening,
		})
		if !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("in %s: RecordDecision(PROTECTION_WEAKENING) = %v, want a refusal", mode, err)
		}
	}
}

func TestModeRankAndConservativePrecedence(t *testing.T) {
	if got := MoreConservativeMode(ModeNormal, ModeEntryBlocked); got != ModeEntryBlocked {
		t.Errorf("MoreConservativeMode(NORMAL, ENTRY_BLOCKED) = %s", got)
	}
	if got := MoreConservativeMode(ModeHaltAll, ModeEntryBlocked); got != ModeHaltAll {
		t.Errorf("MoreConservativeMode(HALT_ALL, ENTRY_BLOCKED) = %s", got)
	}
	if got := MoreConservativeMode(ModeNormal, ModeNormal); got != ModeNormal {
		t.Errorf("MoreConservativeMode(NORMAL, NORMAL) = %s", got)
	}
	// An unrecognised mode wins: "we do not know what this means" cannot be the
	// reading that lets an entry through.
	if got := MoreConservativeMode("WHAT", ModeHaltAll); got != "WHAT" {
		t.Errorf("MoreConservativeMode(unknown, HALT_ALL) = %s, want the unknown one", got)
	}
	if _, known := ModeRank("EXIT_ONLY"); known {
		t.Error("EXIT_ONLY is not a mode this build has (D3)")
	}
}

// --- persistence and projection -------------------------------------------------

func TestATransitionIsPersistedAndProjectedInOneFlow(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()

	rec, changed, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, nil)
	if err != nil || !changed {
		t.Fatalf("EscalateOperatingMode: changed=%v err=%v", changed, err)
	}
	if rec.Mode != ModeEntryBlocked || rec.Actor != ModeActorAuto {
		t.Fatalf("recorded %+v, want ENTRY_BLOCKED by AUTO", rec)
	}

	if got := currentMode(t, j, "acct-1"); got.Mode != ModeEntryBlocked || !got.Recorded ||
		got.Cause != ModeTriggerDailyLossLimit {
		t.Fatalf("current mode %+v", got)
	}

	// The projection is not a second call the caller has to remember.
	seen := projector.records()
	if len(seen) != 1 || seen[0].Mode != ModeEntryBlocked || !seen[0].BlocksEntry() {
		t.Fatalf("projected %+v, want one blocking ENTRY_BLOCKED record", seen)
	}
}

func TestNoModeRowMeansNormal(t *testing.T) {
	j, _ := modeJournal(t)
	got := currentMode(t, j, "acct-never-touched")
	if got.Mode != ModeNormal || got.Recorded || got.BlocksEntry() {
		t.Fatalf("an account with no history is %+v, want an unrecorded NORMAL", got)
	}
	if !got.Since.IsZero() {
		t.Fatalf("Since = %s, want zero for an account with no row", got.Since)
	}
}

// TestTheLatestRowIsOrderedByCreatedAtThenRowid is issues.md's task-0.1 finding,
// as a behaviour.
//
// Journal timestamps are second-resolution, so two transitions inside one second
// share a created_at. Ordering by created_at alone leaves the pair ambiguous, and
// SQLite is free to return either — which in this table means a relaxation could
// read as the current mode after a tightening that came later. The fixed clock
// makes that collision certain rather than occasional.
func TestTheLatestRowIsOrderedByCreatedAtThenRowid(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	const account = "acct-1"
	auditor := &recordingAuditor{}

	if _, _, err := j.EscalateOperatingMode(ctx, account, ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: account, Mode: ModeNormal, Actor: ModeActorOperator,
		Cause: "credential rotated", Approval: "sre-1", Auditor: auditor,
	}); err != nil {
		t.Fatalf("relaxing: %v", err)
	}
	// …and back up again, still inside the same second.
	if _, _, err := j.EscalateOperatingMode(ctx, account, ModeTriggerExitObservationOutage, nil); err != nil {
		t.Fatalf("re-escalating: %v", err)
	}

	history, err := j.OperatingModeHistory(ctx, account)
	if err != nil {
		t.Fatalf("OperatingModeHistory: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("history has %d rows, want 3: %+v", len(history), history)
	}
	// The premise: all three share a timestamp, so created_at cannot order them.
	for _, rec := range history[1:] {
		if !rec.CreatedAt.Equal(history[0].CreatedAt) {
			t.Fatalf("the fixed clock did not produce one second: %s vs %s",
				rec.CreatedAt, history[0].CreatedAt)
		}
	}
	if got := currentMode(t, j, account); got.Mode != ModeEntryBlocked {
		t.Fatalf("current mode is %s; ordering by created_at alone let the relaxation win", got.Mode)
	}
	// History stays in insertion order too, for the same reason.
	wantModes := []string{ModeEntryBlocked, ModeNormal, ModeEntryBlocked}
	for i, want := range wantModes {
		if history[i].Mode != want {
			t.Fatalf("history[%d] = %s, want %s", i, history[i].Mode, want)
		}
	}
}

// TestConservativePrecedenceKeepsTheStricterMode: 동시 적용 시 보수 우선(SHALL).
func TestConservativePrecedenceKeepsTheStricterMode(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()
	const account = "acct-1"

	if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: account, Mode: ModeHaltAll, Actor: ModeActorOperator,
		Cause: "operator halted the account", Auditor: &recordingAuditor{},
	}); err != nil {
		t.Fatalf("entering HALT_ALL: %v", err)
	}
	projected := len(projector.records())

	// A daily-loss trigger now fires. Its target is ENTRY_BLOCKED, which permits
	// strictly more than HALT_ALL, so it must not take effect — and it is not an
	// error either, because a trigger firing while an operator has already halted
	// the account is ordinary.
	rec, changed, err := j.EscalateOperatingMode(ctx, account, ModeTriggerDailyLossLimit, nil)
	if err != nil {
		t.Fatalf("EscalateOperatingMode during HALT_ALL: %v", err)
	}
	if changed || rec.Mode != "" {
		t.Fatalf("a trigger relaxed HALT_ALL to %+v", rec)
	}
	if got := currentMode(t, j, account); got.Mode != ModeHaltAll {
		t.Fatalf("current mode = %s, want HALT_ALL to stand", got.Mode)
	}
	if len(projector.records()) != projected {
		t.Fatal("a no-op transition still projected; the gate would flap")
	}
	history, _ := j.OperatingModeHistory(ctx, account)
	if len(history) != 1 {
		t.Fatalf("a no-op appended a row: %+v", history)
	}
}

func TestARepeatedTriggerAppendsNothing(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, changed, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerExitObservationOutage, nil)
		if err != nil {
			t.Fatalf("escalation %d: %v", i, err)
		}
		if want := i == 0; changed != want {
			t.Fatalf("escalation %d reported changed=%v, want %v", i, changed, want)
		}
	}
	history, _ := j.OperatingModeHistory(context.Background(), "acct-1")
	if len(history) != 1 {
		t.Fatalf("a poll-cycle trigger wrote %d rows; the history would bury the one that mattered", len(history))
	}
	if len(projector.records()) != 1 {
		t.Fatalf("projected %d times for one transition", len(projector.records()))
	}
}

// TestConcurrentEscalationsConvergeOnTheStrictestMode is the "동시 적용" case with
// real concurrency: whichever order they land in, the account ends up in the most
// conservative mode any of them asked for, and no caller sees an error.
func TestConcurrentEscalationsConvergeOnTheStrictestMode(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	const account = "acct-1"

	triggers := AutomaticTriggers()
	var wg sync.WaitGroup
	errs := make([]error, len(triggers)*4)
	for i := 0; i < len(errs); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, err := j.EscalateOperatingMode(ctx, account, triggers[i%len(triggers)], nil)
			errs[i] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent escalation %d: %v", i, err)
		}
	}
	if got := currentMode(t, j, account); got.Mode != ModeEntryBlocked {
		t.Fatalf("current mode = %s, want ENTRY_BLOCKED", got.Mode)
	}
	history, _ := j.OperatingModeHistory(ctx, account)
	if len(history) != 1 {
		t.Fatalf("%d rows for one effective transition: %+v", len(history), history)
	}
}

// TestTheModeIsRestoredAfterARestart: 모드·kill switch·이력은 journal 영속·재시작
// 유지(SHALL). A fresh process has an empty projection until it asks.
func TestTheModeIsRestoredAfterARestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	first := openTestJournalAt(t, path)
	if err := first.SetModeProjector(&recordingProjector{}); err != nil {
		t.Fatalf("SetModeProjector: %v", err)
	}
	if _, _, err := first.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	restarted := openTestJournalAt(t, path)
	projector := &recordingProjector{}
	if err := restarted.SetModeProjector(projector); err != nil {
		t.Fatalf("SetModeProjector after restart: %v", err)
	}
	if len(projector.records()) != 0 {
		t.Fatal("opening the journal projected by itself; the restore must be an explicit step")
	}

	snapshot, err := restarted.RestoreOperatingModeProjection(ctx, "acct-1")
	if err != nil {
		t.Fatalf("RestoreOperatingModeProjection: %v", err)
	}
	if snapshot.Mode != ModeEntryBlocked || !snapshot.BlocksEntry() {
		t.Fatalf("restored %+v, want a blocking ENTRY_BLOCKED", snapshot)
	}
	seen := projector.records()
	if len(seen) != 1 || !seen[0].BlocksEntry() {
		t.Fatalf("the restart did not re-project the block: %+v", seen)
	}
}

// TestRestoringANormalAccountProjectsTheReleaseToo: the restore has to run the
// clearing direction as well, or a gate left latched by a previous life would
// survive a relaxation it never saw.
func TestRestoringANormalAccountProjectsTheReleaseToo(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()

	snapshot, err := j.RestoreOperatingModeProjection(ctx, "acct-1")
	if err != nil {
		t.Fatalf("RestoreOperatingModeProjection: %v", err)
	}
	if snapshot.Mode != ModeNormal {
		t.Fatalf("mode = %s, want NORMAL", snapshot.Mode)
	}
	seen := projector.records()
	if len(seen) != 1 || seen[0].BlocksEntry() {
		t.Fatalf("projected %+v, want one non-blocking NORMAL record", seen)
	}
}

// --- announcement (task 3.3) ------------------------------------------------------

// failingAnnouncer stands in for an alert path that cannot even make the alert
// durable — a full disk under the outbox, say. (A merely *undelivered* alert is
// not an error here: obs.Notifier handles that by latching the gate, and
// internal/obs/mode_test.go covers it.)
type failingAnnouncer struct {
	calls int
	err   error
}

func (a *failingAnnouncer) AnnounceOperatingMode(_ context.Context, _ string, _ OperatingModeRecord) error {
	a.calls++
	return a.err
}

// TestAFailedAnnouncementDoesNotUndoTheTransition.
//
// The tightening is durable and projected before anything is announced, and it
// stays that way: rolling a mode change back because nobody could be told about
// it would make the safety mechanism depend on the notification transport, which
// is the one dependency it must not have. The caller learns through a distinct
// sentinel so it does not retry a transition that already happened.
func TestAFailedAnnouncementDoesNotUndoTheTransition(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()
	announcer := &failingAnnouncer{err: errors.New("the outbox disk is full")}

	rec, changed, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, announcer)
	if !errors.Is(err, ErrModeAnnouncementFailed) {
		t.Fatalf("err = %v, want ErrModeAnnouncementFailed", err)
	}
	if !changed || rec.Mode != ModeEntryBlocked {
		t.Fatalf("rec=%+v changed=%v — the transition happened and must be reported as such", rec, changed)
	}
	if got := currentMode(t, j, "acct-1"); got.Mode != ModeEntryBlocked {
		t.Fatalf("mode = %s, want the tightening to stand", got.Mode)
	}
	if seen := projector.records(); len(seen) != 1 || !seen[0].BlocksEntry() {
		t.Fatalf("projected %+v, want the gate latched despite the announcement", seen)
	}
}

func TestANoOpTransitionIsNotAnnounced(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	announcer := &failingAnnouncer{}

	if _, changed, err := j.EscalateOperatingMode(ctx, "acct-1",
		ModeTriggerDailyLossLimit, announcer); err != nil || !changed {
		t.Fatalf("first escalation: changed=%v err=%v", changed, err)
	}
	if announcer.calls != 1 {
		t.Fatalf("announcements = %d, want 1", announcer.calls)
	}
	if _, changed, err := j.EscalateOperatingMode(ctx, "acct-1",
		ModeTriggerCredentialRejected, announcer); err != nil || changed {
		t.Fatalf("second escalation: changed=%v err=%v", changed, err)
	}
	if announcer.calls != 1 {
		t.Fatalf("announcements = %d; a transition that did not happen was announced", announcer.calls)
	}
}

// TestARefusedTransitionIsNotAnnounced: the announcement is downstream of the
// commit, so a request the rules refused never reaches it.
func TestARefusedTransitionIsNotAnnounced(t *testing.T) {
	j, _ := modeJournal(t)
	announcer := &failingAnnouncer{}

	if _, _, err := j.TransitionOperatingMode(context.Background(), TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeHaltAll, Actor: ModeActorAuto,
		Cause: ModeTriggerDailyLossLimit, Announcer: announcer,
	}); !errors.Is(err, ErrHaltAllIsNeverAutomatic) {
		t.Fatalf("err = %v", err)
	}
	if announcer.calls != 0 {
		t.Fatalf("announcements = %d, want none for a refused transition", announcer.calls)
	}
}

func TestModeProjectorIsBoundOnce(t *testing.T) {
	j := openTestJournal(t)
	if err := j.SetModeProjector(nil); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("SetModeProjector(nil) = %v, want a refusal", err)
	}
	if err := j.SetModeProjector(&recordingProjector{}); err != nil {
		t.Fatalf("first bind: %v", err)
	}
	if err := j.SetModeProjector(&recordingProjector{}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("rebinding = %v, want a refusal", err)
	}
}

// --- the shared read surface (task 3.3) --------------------------------------------

// TestTheSnapshotIsWhatThreeConsumersRead: 모드 스냅샷 배선(Gateway·Guardian·
// flatten 동일 뷰).
//
// Each of the three asks a different question of the same value, and the point
// of one surface is that they cannot end up with three answers about one account.
func TestTheSnapshotIsWhatThreeConsumersRead(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	const account = "acct-1"

	if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: account, Mode: ModeHaltAll, Actor: ModeActorOperator,
		Cause: "operator halted after an incident", Auditor: &recordingAuditor{},
	}); err != nil {
		t.Fatalf("entering HALT_ALL: %v", err)
	}
	snapshot := currentMode(t, j, account)

	// Gateway: may this submission raise exposure?
	if snapshot.Allows(SafetyClassExposureRaising) || !snapshot.BlocksEntry() {
		t.Fatal("HALT_ALL must refuse EXPOSURE_RAISING")
	}
	// Guardian: the chain's mode rung reads the same string, and it is one the
	// chain recognises (internal/risk's three constants are these three).
	if _, known := ModeRank(snapshot.Mode); !known {
		t.Fatalf("mode %q is not one the chain can read", snapshot.Mode)
	}
	// flatten: 수동 flatten-all은 모든 모드에서 통과한다(§0.3).
	if !snapshot.Allows(SafetyClassRiskReducing) {
		t.Fatal("HALT_ALL refused a risk-reducing mutation")
	}
	// …and the surface carries the provenance an operator needs, so none of the
	// three has to go and join the history itself.
	if snapshot.Actor != ModeActorOperator || snapshot.Cause == "" || snapshot.Since.IsZero() {
		t.Fatalf("snapshot %+v does not say who, why and when", snapshot)
	}
}

// TestFlattenIsNeverGatedByTheMode is §0.3 at the saga's own layer: the flatten
// record does not consult the mode, in any mode, and a flatten started before a
// HALT_ALL keeps running through it.
func TestFlattenIsNeverGatedByTheMode(t *testing.T) {
	ctx := context.Background()
	for _, mode := range []string{ModeNormal, ModeEntryBlocked, ModeHaltAll} {
		t.Run(mode, func(t *testing.T) {
			j, _ := modeJournal(t)
			if mode != ModeNormal {
				if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
					AccountRef: "acct-1", Mode: mode, Actor: ModeActorOperator,
					Cause: "test fixture", Auditor: &recordingAuditor{},
				}); err != nil {
					t.Fatalf("reaching %s: %v", mode, err)
				}
			}
			saga, err := j.StartFlatten(ctx, FlattenSaga{ID: "fl-1", AccountRef: "acct-1",
				Reason: "operator flatten", Operator: "sre-1"})
			if err != nil {
				t.Fatalf("in %s: StartFlatten: %v", mode, err)
			}
			if !saga.Active() {
				t.Fatalf("in %s: the saga did not start", mode)
			}
			if _, err := j.AddFlattenStep(ctx, FlattenStep{
				SagaID: saga.ID, Kind: FlattenStepLiquidate, Market: "kr", Symbol: "005930",
				Side: "SELL", Quantity: "9", State: FlattenStepPending,
			}); err != nil {
				t.Fatalf("in %s: AddFlattenStep: %v", mode, err)
			}
		})
	}
}

// --- request validation ----------------------------------------------------------

func TestUnusableTransitionRequestsAreRefused(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	base := TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeEntryBlocked, Actor: ModeActorAuto,
		Cause: ModeTriggerDailyLossLimit,
	}

	cases := []struct {
		name string
		mut  func(*TransitionModeRequest)
	}{
		{"no account", func(r *TransitionModeRequest) { r.AccountRef = "  " }},
		{"unknown mode", func(r *TransitionModeRequest) { r.Mode = "EXIT_ONLY" }},
		{"empty mode", func(r *TransitionModeRequest) { r.Mode = "" }},
		{"unknown actor", func(r *TransitionModeRequest) { r.Actor = "SYSTEM" }},
		{"empty actor", func(r *TransitionModeRequest) { r.Actor = "" }},
		{"no cause", func(r *TransitionModeRequest) { r.Cause = "   " }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := base
			tc.mut(&req)
			_, changed, err := j.TransitionOperatingMode(ctx, req)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("err = %v, want ErrInvalidRequest", err)
			}
			if changed {
				t.Fatal("a refused request reported a change")
			}
		})
	}
	if got := currentMode(t, j, "acct-1"); got.Recorded {
		t.Fatalf("a refused request wrote a row: %+v", got)
	}
}

func TestCurrentOperatingModeNeedsAnAccount(t *testing.T) {
	j, _ := modeJournal(t)
	if _, err := j.CurrentOperatingMode(context.Background(), " "); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

func TestModesAreScopedToTheirAccount(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()

	if _, _, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, nil); err != nil {
		t.Fatalf("escalating acct-1: %v", err)
	}
	if got := currentMode(t, j, "acct-2"); got.Mode != ModeNormal || got.Recorded {
		t.Fatalf("acct-2 inherited acct-1's mode: %+v", got)
	}
}

func TestTransitionIDsDoNotCollideInsideOneSecond(t *testing.T) {
	// The clock is fixed, so two transitions with the same target and cause
	// derive from identical inputs but for the sequence number. Without it the
	// second INSERT would hit the primary key.
	j, _ := modeJournal(t)
	ctx := context.Background()
	const account = "acct-1"
	auditor := &recordingAuditor{}

	for i := 0; i < 3; i++ {
		if _, _, err := j.EscalateOperatingMode(ctx, account, ModeTriggerDailyLossLimit, nil); err != nil {
			t.Fatalf("escalation %d: %v", i, err)
		}
		if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
			AccountRef: account, Mode: ModeNormal, Actor: ModeActorOperator,
			Cause: "resolved", Approval: "sre-1", Auditor: auditor,
		}); err != nil {
			t.Fatalf("relaxation %d: %v", i, err)
		}
	}
	history, _ := j.OperatingModeHistory(ctx, account)
	if len(history) != 6 {
		t.Fatalf("history has %d rows, want 6", len(history))
	}
	seen := map[string]bool{}
	for _, rec := range history {
		if seen[rec.ID] {
			t.Fatalf("duplicate transition id %s", rec.ID)
		}
		seen[rec.ID] = true
		if !strings.HasPrefix(rec.ID, "mode-") {
			t.Fatalf("id %q does not carry its kind", rec.ID)
		}
	}
}
