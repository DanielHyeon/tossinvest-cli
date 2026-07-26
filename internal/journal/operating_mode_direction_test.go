package journal

// operating_mode_direction_test.go is add-core-domain task 3.2: 전환 승인은 방향
// 비대칭이다(SHALL).
//
//	보수 방향 (NORMAL→ENTRY_BLOCKED→HALT_ALL)  자동·즉시·durable
//	완화                                        사람 승인(§0.7) + audit
//
// The tests are organised the way the spec's two sentences are: what the
// enumerated triggers may do, and what nothing automatic may do.

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// --- the conservative direction: automatic, immediate, durable ------------------

// TestEveryEnumeratedTriggerTargetsEntryBlocked is the enumeration itself
// (risk-management: 일일 손실 한도 도달 / 자격증명 실패(401/403) / critical 알림
// outbox 전달 실패 지속 / exit 관측 두절 임계 초과 / 대사 사이클 지속 실패 / 체결
// 감지 사이클 지속 실패 — 전부 ENTRY_BLOCKED).
//
// It runs over the whole trigger set rather than a list repeated here, so adding
// a trigger without deciding its target is a compile-time-visible omission
// (TargetModeForTrigger) and adding one that reaches HALT_ALL fails this test.
//
// The last two joined the table in add-engine-runtime: a runtime whose
// reconciliation or fill detection has failed five cycles running is blind to
// what the account holds, and an engine in that state must not take new exposure.
func TestEveryEnumeratedTriggerTargetsEntryBlocked(t *testing.T) {
	triggers := AutomaticTriggers()
	if len(triggers) != 6 {
		t.Fatalf("the spec enumerates six triggers, this build has %d: %v", len(triggers), triggers)
	}
	for _, want := range []string{
		ModeTriggerDailyLossLimit,
		ModeTriggerCredentialRejected,
		ModeTriggerCriticalAlertUndelivered,
		ModeTriggerExitObservationOutage,
		ModeTriggerReconcileCycleFailure,
		ModeTriggerFillDetectionFailure,
	} {
		if !AutomaticTrigger(want) {
			t.Errorf("%s is named by risk-management but is not a trigger here", want)
		}
	}
	for _, trigger := range triggers {
		target, ok := TargetModeForTrigger(trigger)
		if !ok {
			t.Fatalf("%s is listed but maps to nothing", trigger)
		}
		if target != ModeEntryBlocked {
			t.Fatalf("%s escalates to %s; every automatic target is ENTRY_BLOCKED (SHALL NOT — HALT_ALL 자동 진입 없음)",
				trigger, target)
		}
	}
}

// TestEachTriggerEscalatesImmediatelyAndDurably runs each one end to end: no
// approval, no operator, one row, and the row is on disk before the call
// returns.
func TestEachTriggerEscalatesImmediatelyAndDurably(t *testing.T) {
	for _, trigger := range AutomaticTriggers() {
		t.Run(trigger, func(t *testing.T) {
			j, projector := modeJournal(t)
			ctx := context.Background()

			rec, changed, err := j.EscalateOperatingMode(ctx, "acct-1", trigger, nil)
			if err != nil || !changed {
				t.Fatalf("changed=%v err=%v", changed, err)
			}
			if rec.Actor != ModeActorAuto {
				t.Fatalf("actor = %s, want %s — the conservative direction needs no person",
					rec.Actor, ModeActorAuto)
			}
			if rec.Mode != ModeEntryBlocked {
				t.Fatalf("mode = %s, want ENTRY_BLOCKED", rec.Mode)
			}
			if rec.Cause != trigger {
				t.Fatalf("cause = %q, want the trigger verbatim so the history says why", rec.Cause)
			}
			// Durable: a reader that never saw the call agrees.
			if got := currentMode(t, j, "acct-1"); got.Mode != ModeEntryBlocked || !got.Recorded {
				t.Fatalf("current mode %+v", got)
			}
			// …and projected, so "즉시" means the gate too.
			if seen := projector.records(); len(seen) != 1 || !seen[0].BlocksEntry() {
				t.Fatalf("projected %+v", seen)
			}
		})
	}
}

// --- what nothing automatic may do ----------------------------------------------

// TestHaltAllIsNeverAutomatic: HALT_ALL은 운영자 결정(SHALL NOT — 자동 진입 없음).
//
// Two routes are closed, because there are two: naming HALT_ALL directly, and
// hoping some trigger maps to it.
func TestHaltAllIsNeverAutomatic(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()

	for _, trigger := range AutomaticTriggers() {
		_, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
			AccountRef: "acct-1", Mode: ModeHaltAll, Actor: ModeActorAuto, Cause: trigger,
		})
		if !errors.Is(err, ErrHaltAllIsNeverAutomatic) {
			t.Fatalf("AUTO → HALT_ALL with trigger %s returned %v, want ErrHaltAllIsNeverAutomatic",
				trigger, err)
		}
	}
	if got := currentMode(t, j, "acct-1"); got.Recorded {
		t.Fatalf("a refused automatic HALT_ALL wrote a row: %+v", got)
	}
	if len(projector.records()) != 0 {
		t.Fatal("a refused transition still projected")
	}

	// The operator route is open, and that difference is the whole reason
	// HALT_ALL exists inside this change (D3): it behaves like ENTRY_BLOCKED and
	// means something else.
	rec, changed, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeHaltAll, Actor: ModeActorOperator,
		Cause: "operator halted the account after an incident", Auditor: &recordingAuditor{},
	})
	if err != nil || !changed || rec.Mode != ModeHaltAll {
		t.Fatalf("operator → HALT_ALL: rec=%+v changed=%v err=%v", rec, changed, err)
	}
}

// TestAnalyticsFailureIsNotATrigger: 분석·성과 작업 실패는 트리거가 아니다
// (SHALL NOT).
//
// A performance aggregation that keeps failing is a reason to retry the
// aggregation. Letting it tighten the account would mean a reporting bug can
// stop the engine trading, which inverts what the mode is for.
//
// The distinction the enumeration draws is execution vs analysis, not "a loop
// failed": reconciliation and fill detection ARE triggers (they are how the
// engine learns what the account holds) and an aggregation job is not.
func TestAnalyticsFailureIsNotATrigger(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()

	for _, cause := range []string{
		"PERFORMANCE_AGGREGATION_FAILED",
		"TRADE_OUTCOME_RETENTION_SWEEP_FAILED",
		"analytics job crashed",
		"", // and no cause at all
	} {
		_, changed, err := j.EscalateOperatingMode(ctx, "acct-1", cause, nil)
		if err == nil {
			t.Fatalf("cause %q escalated the account", cause)
		}
		if cause != "" && !errors.Is(err, ErrModeTriggerUnknown) {
			t.Fatalf("cause %q returned %v, want ErrModeTriggerUnknown", cause, err)
		}
		if changed {
			t.Fatalf("cause %q reported a change", cause)
		}
		// The refusal names the enumerated set, so an operator reading the log
		// can tell a typo from a genuinely unlisted condition.
		if cause != "" && !strings.Contains(err.Error(), ModeTriggerDailyLossLimit) {
			t.Fatalf("the refusal %q does not list the enumerated triggers", err)
		}
	}
	if got := currentMode(t, j, "acct-1"); got.Recorded || got.Mode != ModeNormal {
		t.Fatalf("an analytics failure changed the mode to %+v; only 분석 재시도만 수행된다", got)
	}
	if len(projector.records()) != 0 {
		t.Fatal("an analytics failure reached the gate")
	}
}

// TestAnAutomaticRelaxationIsRefused: the direction is one-way for AUTO. The
// condition that raised the mode is not disproved by the code that raised it no
// longer firing.
func TestAnAutomaticRelaxationIsRefused(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()

	if _, _, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}

	// Directly, with a valid trigger as the cause.
	_, changed, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeNormal, Actor: ModeActorAuto,
		Cause: ModeTriggerDailyLossLimit,
	})
	if !errors.Is(err, ErrModeRelaxationRequiresOperator) {
		t.Fatalf("AUTO → NORMAL returned %v, want ErrModeRelaxationRequiresOperator", err)
	}
	if changed {
		t.Fatal("a refused relaxation reported a change")
	}
	if got := currentMode(t, j, "acct-1"); got.Mode != ModeEntryBlocked {
		t.Fatalf("mode = %s, want ENTRY_BLOCKED to stand", got.Mode)
	}

	// And an approval attached to an AUTO request does not buy the relaxation:
	// the approval is a human's, and AUTO is not one.
	if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeNormal, Actor: ModeActorAuto,
		Cause: ModeTriggerDailyLossLimit, Approval: "sre-1", Auditor: &recordingAuditor{},
	}); !errors.Is(err, ErrModeRelaxationRequiresOperator) {
		t.Fatalf("AUTO with an approval returned %v, want a refusal", err)
	}
}

// --- the relaxation direction: approval + audit -----------------------------------

func TestRelaxationRequiresAnApproval(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	if _, _, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}

	_, changed, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeNormal, Actor: ModeActorOperator,
		Cause: "I looked and it is fine", Auditor: &recordingAuditor{},
	})
	if !errors.Is(err, ErrModeApprovalRequired) {
		t.Fatalf("err = %v, want ErrModeApprovalRequired", err)
	}
	if changed {
		t.Fatal("a relaxation without an approval reported a change")
	}
	if got := currentMode(t, j, "acct-1"); got.Mode != ModeEntryBlocked {
		t.Fatalf("mode = %s", got.Mode)
	}
}

func TestRelaxationRequiresAnAuditLog(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	if _, _, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}

	_, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeNormal, Actor: ModeActorOperator,
		Cause: "the loss was a mispriced snapshot", Approval: "sre-1",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want a refusal — §0.5 requires the change to be traceable", err)
	}
	if got := currentMode(t, j, "acct-1"); got.Mode != ModeEntryBlocked {
		t.Fatalf("mode = %s", got.Mode)
	}
}

// TestTheApprovedRelaxationIsAuditedWithBothValues: the audit line has to be
// usable on its own — engine-safety asks for before-value, after-value, time and
// subject, and a line that only says "the mode changed" is not that.
func TestTheApprovedRelaxationIsAuditedWithBothValues(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()
	auditor := &recordingAuditor{}

	if _, _, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	rec, changed, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeNormal, Actor: ModeActorOperator,
		Cause: "credential rotated and a read succeeded", Approval: "ticket-4471",
		Auditor: auditor,
	})
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	if len(auditor.actions) != 1 {
		t.Fatalf("audit actions = %v, want one", auditor.actions)
	}
	line := auditor.actions[0]
	for _, want := range []string{
		AuditActionOperatingMode, // the action
		"acct-1",                 // the subject account
		ModeNormal,               // the new value
		ModeEntryBlocked,         // the old value
		ModeActorOperator,        // who
		"ticket-4471",            // the approval
		"credential rotated",     // why
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("audit line %q does not carry %q", line, want)
		}
	}
	// The approval is in the durable row too, not only in the audit file: the
	// journal must be able to answer "who approved this" on its own.
	if !strings.Contains(rec.Cause, "ticket-4471") {
		t.Fatalf("recorded cause %q drops the approval", rec.Cause)
	}
}

// TestAFailedAuditWriteAbortsTheRelaxation: the audit line goes in before the
// commit, the same way the operator reservation release does. A relaxation whose
// record is missing is exactly the state the approval requirement exists to
// prevent.
func TestAFailedAuditWriteAbortsTheRelaxation(t *testing.T) {
	j, projector := modeJournal(t)
	ctx := context.Background()
	if _, _, err := j.EscalateOperatingMode(ctx, "acct-1", ModeTriggerDailyLossLimit, nil); err != nil {
		t.Fatalf("escalating: %v", err)
	}
	before := len(projector.records())

	auditor := &recordingAuditor{err: errors.New("the audit disk is full")}
	_, changed, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeNormal, Actor: ModeActorOperator,
		Cause: "resolved", Approval: "sre-1", Auditor: auditor,
	})
	if err == nil {
		t.Fatal("a failed audit write let the relaxation through")
	}
	if changed {
		t.Fatal("a failed audit write reported a change")
	}
	if got := currentMode(t, j, "acct-1"); got.Mode != ModeEntryBlocked {
		t.Fatalf("mode = %s, want the block to stand", got.Mode)
	}
	if len(projector.records()) != before {
		t.Fatal("a rolled-back relaxation still cleared the gate")
	}
}

// TestAnOperatorTighteningNeedsNoApproval: the asymmetry runs both ways. Making
// the account safer is never gated on paperwork — that is §0.9's direction rule,
// and a tightening an operator cannot apply immediately is a tightening that
// happens too late.
func TestAnOperatorTighteningNeedsNoApproval(t *testing.T) {
	j, _ := modeJournal(t)
	ctx := context.Background()

	rec, changed, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-1", Mode: ModeHaltAll, Actor: ModeActorOperator,
		Cause: "market-wide halt",
	})
	if err != nil || !changed || rec.Mode != ModeHaltAll {
		t.Fatalf("rec=%+v changed=%v err=%v", rec, changed, err)
	}
	// …and it is still audited when an auditor is wired, because §0.5 tracks
	// every operating-mode change, not only the loosening ones.
	auditor := &recordingAuditor{}
	if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: "acct-2", Mode: ModeEntryBlocked, Actor: ModeActorOperator,
		Cause: "pre-emptive block", Auditor: auditor,
	}); err != nil {
		t.Fatalf("tightening acct-2: %v", err)
	}
	if len(auditor.actions) != 1 {
		t.Fatalf("audit actions = %v, want the tightening recorded too", auditor.actions)
	}
}

// TestTheFullTransitionMatrix walks every (from, to, actor) combination once, so
// no direction rule is asserted only by the case that happened to be written.
func TestTheFullTransitionMatrix(t *testing.T) {
	modes := []string{ModeNormal, ModeEntryBlocked, ModeHaltAll}

	for _, from := range modes {
		for _, to := range modes {
			for _, actor := range []string{ModeActorAuto, ModeActorOperator} {
				name := from + "→" + to + "/" + actor
				t.Run(name, func(t *testing.T) {
					j, _ := modeJournal(t)
					ctx := context.Background()
					const account = "acct-1"

					if from != ModeNormal {
						if _, _, err := j.TransitionOperatingMode(ctx, TransitionModeRequest{
							AccountRef: account, Mode: from, Actor: ModeActorOperator,
							Cause: "test fixture", Auditor: &recordingAuditor{},
						}); err != nil {
							t.Fatalf("reaching %s: %v", from, err)
						}
					}

					fromRank, _ := ModeRank(from)
					toRank, _ := ModeRank(to)
					req := TransitionModeRequest{
						AccountRef: account, Mode: to, Actor: actor,
						Cause: "operator reason", Approval: "sre-1", Auditor: &recordingAuditor{},
					}
					if actor == ModeActorAuto {
						req.Cause = ModeTriggerDailyLossLimit
						req.Approval = ""
					}
					_, changed, err := j.TransitionOperatingMode(ctx, req)

					switch {
					case actor == ModeActorAuto && to == ModeHaltAll:
						if !errors.Is(err, ErrHaltAllIsNeverAutomatic) {
							t.Fatalf("err = %v, want ErrHaltAllIsNeverAutomatic", err)
						}
					case actor == ModeActorAuto && to == ModeNormal:
						if !errors.Is(err, ErrModeRelaxationRequiresOperator) {
							t.Fatalf("err = %v, want ErrModeRelaxationRequiresOperator", err)
						}
					case actor == ModeActorAuto:
						// to == ENTRY_BLOCKED: tighten from NORMAL, no-op otherwise.
						if err != nil {
							t.Fatalf("err = %v, want nil", err)
						}
						if want := fromRank < toRank; changed != want {
							t.Fatalf("changed = %v, want %v", changed, want)
						}
					default:
						if err != nil {
							t.Fatalf("err = %v, want nil", err)
						}
						if want := fromRank != toRank; changed != want {
							t.Fatalf("changed = %v, want %v", changed, want)
						}
					}

					// Whatever happened, the account is never in a less
					// conservative mode than the rules allow.
					got := currentMode(t, j, account)
					wantMode := from
					if changed {
						wantMode = to
					}
					if got.Mode != wantMode {
						t.Fatalf("mode = %s, want %s", got.Mode, wantMode)
					}
				})
			}
		}
	}
}
