package journal

// operating_mode.go is the durable operating mode (design D3, add-core-domain
// tasks 3.1–3.2; risk-management "Kill switch와 운영 모드 — 모드×클래스 표").
//
// # The journal is the authority, the gate is a view
//
// The mode decides whether the account may open exposure, and risk-management
// names the enforcement point exactly: 모드 전환은 journal 영속과 동시에 EntryGate
// 계좌 latch로 투영되고, Gateway의 기존 제출 시 재검사가 이를 소비한다. So there is
// one way to change the mode — TransitionOperatingMode — and it both appends the
// row and projects it. There is deliberately no exported "append without
// projecting" and no "project without appending": the first would leave a
// tightening the gate never learned about, and the second a block no restart
// could rebuild and no operator could find a record of.
//
// The order inside that flow is commit-then-project, and it is chosen for the
// relaxation direction. Projecting first would clear the gate's latch before the
// row that justifies clearing it is durable, so a crash in between would come
// back with an open gate and a journal that still says ENTRY_BLOCKED. The other
// direction — a tightening committed but not yet projected — is a window of
// microseconds inside one process, and RestoreOperatingModeProjection closes it
// on the next start.
//
// # Direction is asymmetric, and the column is what records it
//
// 보수 방향(NORMAL→ENTRY_BLOCKED→HALT_ALL)은 자동·즉시·durable, 완화는 사람
// 승인(§0.7)+audit. Three rules implement that here:
//
//   - AUTO may only tighten. An automatic relaxation is refused, because the
//     condition that raised the mode is not disproved by the code that raised it
//     no longer firing.
//   - HALT_ALL is never automatic (SHALL NOT). Its meaning is "an operator
//     decided", and a machine that could enter it would erase that meaning.
//   - A relaxation carries OPERATOR, an approval reference and an audit line, and
//     the audit line is written before the commit — a relaxation whose record is
//     missing is the exact state this path exists to prevent.
//
// # Concurrency: the conservative side wins
//
// 동시 적용 시 보수 우선(SHALL). An automatic tightening whose target is not more
// conservative than the mode already in force is a no-op that appends nothing
// and returns changed=false. That is what makes concurrent escalations
// order-independent — whichever runs last, the account ends up in the most
// conservative mode any of them asked for — and it is also what stops the
// alert/escalation feedback loop: an undelivered critical alert escalates to
// ENTRY_BLOCKED, which alerts, which may fail to deliver, which escalates to
// ENTRY_BLOCKED again — and the second escalation changes nothing and therefore
// announces nothing.
//
// # Broker-behaviour claims
//
// None. Every value here is a TossOS-internal judgement.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// AuditActionOperatingMode is the audit log's action string for a mode
// transition. The journal owns the string because the journal owns the event;
// internal/audit owns where it is written.
const AuditActionOperatingMode = "operating_mode.transition"

// Automatic tightening triggers (risk-management 자동 강화 트리거와 목적 상태의
// 열거 — SHALL). The enumeration is closed: a cause outside it cannot escalate,
// which is how "분석·성과 작업 실패는 트리거가 아니다(SHALL NOT)" is enforced
// structurally rather than by everybody remembering it.
const (
	// ModeTriggerDailyLossLimit — 일일 손실 한도 도달.
	ModeTriggerDailyLossLimit = "DAILY_LOSS_LIMIT_REACHED"
	// ModeTriggerCredentialRejected — 자격증명 실패(401/403).
	ModeTriggerCredentialRejected = "BROKER_AUTH_REJECTED"
	// ModeTriggerCriticalAlertUndelivered — critical 알림 outbox 전달 실패 지속.
	ModeTriggerCriticalAlertUndelivered = "CRITICAL_ALERT_UNDELIVERED"
	// ModeTriggerExitObservationOutage — exit 관측 두절 임계 초과 (exit-policy:
	// 관측 두절은 보호 부재와 같다).
	ModeTriggerExitObservationOutage = "EXIT_OBSERVATION_OUTAGE"
	// ModeTriggerReconcileCycleFailure — 대사 사이클 지속 실패 (risk-management,
	// add-engine-runtime: 대사에 눈이 먼 엔진은 새 진입을 받으면 안 된다).
	//
	// The loop keeps retrying; what this trigger changes is whether the account
	// accepts new exposure while the comparison against the broker is failing.
	ModeTriggerReconcileCycleFailure = "RECONCILE_CYCLE_FAILURE"
	// ModeTriggerFillDetectionFailure — 체결 감지 사이클 지속 실패
	// (add-engine-runtime). An engine that cannot see its own fills believes it
	// holds less than it does, which is the belief every sizing decision is made
	// against.
	ModeTriggerFillDetectionFailure = "FILL_DETECTION_CYCLE_FAILURE"
)

// Errors this file returns. All wrap ErrInvalidRequest, so a caller that already
// distinguishes "the request was wrong" keeps working while a caller that needs
// the specific refusal can ask for it.
var (
	// ErrModeRelaxationRequiresOperator — an automatic relaxation was attempted.
	ErrModeRelaxationRequiresOperator = fmt.Errorf(
		"%w: relaxing the operating mode requires an operator", ErrInvalidRequest)
	// ErrHaltAllIsNeverAutomatic — an automatic transition targeted HALT_ALL.
	ErrHaltAllIsNeverAutomatic = fmt.Errorf(
		"%w: HALT_ALL is an operator decision and has no automatic trigger", ErrInvalidRequest)
	// ErrModeTriggerUnknown — an automatic tightening named a cause that is not
	// one of the enumerated triggers.
	ErrModeTriggerUnknown = fmt.Errorf(
		"%w: that cause is not an enumerated automatic tightening trigger", ErrInvalidRequest)
	// ErrModeApprovalRequired — a relaxation carried no approval reference.
	ErrModeApprovalRequired = fmt.Errorf(
		"%w: relaxing the operating mode requires a recorded approval", ErrInvalidRequest)
)

// ModeAuditor records a mode transition where an operator can find it.
//
// One method, satisfied by *audit.Log, for the same reason ReservationAuditor is:
// a storage layer that knows where the operator log lives is a storage layer with
// an opinion about deployment.
type ModeAuditor interface {
	RecordAction(action, setting, value, detail string) error
}

// ModeProjector receives every committed transition.
//
// execgw's EntryGate implements it. The journal holds the interface rather than
// the gate so the dependency points the way it already does (execgw → journal),
// and so a test can observe the projection without a gate.
type ModeProjector interface {
	ProjectOperatingMode(rec OperatingModeRecord)
}

// ModeAnnouncer is told about a transition that actually changed something.
//
// obs.Notifier implements it. It is separate from the projector because the two
// have different failure semantics: a projection cannot fail and must happen, an
// announcement can fail and the transition still stands.
type ModeAnnouncer interface {
	AnnounceOperatingMode(ctx context.Context, previous string, rec OperatingModeRecord) error
}

// ErrModeAnnouncementFailed means the transition is committed and projected but
// the alert did not go out. It is returned *with* a valid record and
// changed=true: retrying the transition would be wrong (it already happened),
// and swallowing it would hide a silent alerting path.
var ErrModeAnnouncementFailed = errors.New("journal: the operating-mode transition was recorded but not announced")

// ValidOperatingMode reports whether a mode is one this build writes.
func ValidOperatingMode(mode string) bool {
	switch mode {
	case ModeNormal, ModeEntryBlocked, ModeHaltAll:
		return true
	default:
		return false
	}
}

// ValidModeActor reports whether an actor is one this build writes.
func ValidModeActor(actor string) bool {
	switch actor {
	case ModeActorAuto, ModeActorOperator:
		return true
	default:
		return false
	}
}

// ModeRank orders the modes by how conservative they are: NORMAL 0,
// ENTRY_BLOCKED 1, HALT_ALL 2. A higher rank permits strictly less.
//
// The bool is false for a mode this build does not know. Callers treat that as
// the most conservative reading rather than as NORMAL — an unrecognised persisted
// enum is a state nobody wrote a release procedure for.
func ModeRank(mode string) (int, bool) {
	switch mode {
	case ModeNormal:
		return 0, true
	case ModeEntryBlocked:
		return 1, true
	case ModeHaltAll:
		return 2, true
	default:
		return 0, false
	}
}

// MoreConservativeMode returns whichever of two modes permits less.
//
// An unrecognised mode wins, which is the fail-closed direction: 동시 적용 시
// 보수 우선(SHALL), and "we do not know what this mode means" cannot be the
// reading that lets an entry through.
func MoreConservativeMode(a, b string) string {
	rankA, okA := ModeRank(a)
	rankB, okB := ModeRank(b)
	switch {
	case !okA:
		return a
	case !okB:
		return b
	case rankB > rankA:
		return b
	default:
		return a
	}
}

// ModeAllows is the mode×class table (risk-management), as one function.
//
//	| 모드           | EXPOSURE_RAISING | RISK_REDUCING | PROTECTION_WEAKENING |
//	| NORMAL         | 허용             | 허용          | 허용(audit)          |
//	| ENTRY_BLOCKED  | 거부             | 허용          | 허용(audit)          |
//	| HALT_ALL       | 거부             | 허용          | 거부                 |
//
// Two notes the table needs and a column heading cannot carry:
//
//   - RISK_REDUCING is 허용 in every row, HALT_ALL included, and 수동 flatten-all은
//     모든 모드에서 통과한다(§0.3). A mode that could trap a position is a mode
//     that makes the account less safe, not more.
//   - The PROTECTION_WEAKENING column is *reserved*. This build issues none —
//     RecordDecision refuses the class outright (decision.go) — so the 허용(audit)
//     cells only acquire effect when the protective-order change introduces the
//     issuer. Until then the journal's refusal is stricter than this table, and
//     that is the safe direction.
//
// An unrecognised mode or class allows nothing.
func ModeAllows(mode, safetyClass string) bool {
	if _, known := ModeRank(mode); !known {
		return false
	}
	switch safetyClass {
	case SafetyClassRiskReducing:
		return true
	case SafetyClassExposureRaising:
		return mode == ModeNormal
	case SafetyClassProtectionWeakening:
		return mode != ModeHaltAll
	default:
		return false
	}
}

// OperatingModeRecord is one stored operating_modes row.
type OperatingModeRecord struct {
	ID         string
	AccountRef string
	Mode       string
	// Cause is why the transition happened: an enumerated trigger for AUTO, or
	// the operator's reason plus their approval reference.
	Cause     string
	Actor     string
	CreatedAt time.Time
}

// BlocksEntry reports whether this mode refuses EXPOSURE_RAISING.
func (r OperatingModeRecord) BlocksEntry() bool {
	return !ModeAllows(r.Mode, SafetyClassExposureRaising)
}

// ModeSnapshot is the shared read surface: the current mode of one account, and
// what it permits.
//
// It is the one view the Gateway, the Guardian chain's issuer and the flatten
// path read, so that three consumers cannot end up with three answers about the
// same account (risk-management: 모드·kill switch·이력은 journal 영속·재시작
// 유지). It is a value, not a handle — nothing here can change the mode.
type ModeSnapshot struct {
	AccountRef string
	Mode       string
	Cause      string
	Actor      string
	// Since is when the mode was entered. Zero when no row exists.
	Since time.Time
	// Recorded is false when the account has no operating_modes row at all.
	// Mode is then NORMAL: an account nobody has ever tightened is not blocked,
	// and defaulting the other way would mean no engine could ever start.
	Recorded bool
}

// BlocksEntry reports whether new exposure is refused.
func (s ModeSnapshot) BlocksEntry() bool {
	return !ModeAllows(s.Mode, SafetyClassExposureRaising)
}

// Allows applies the mode×class table to this snapshot.
func (s ModeSnapshot) Allows(safetyClass string) bool { return ModeAllows(s.Mode, safetyClass) }

// SetModeProjector binds the projection target, once, at wiring time.
//
// Rebinding is refused for the same reason SetApplyHooks refuses it: two
// projections of one mode history is not a configuration, and the second would
// silently become the truth from an arbitrary transition onwards.
func (j *Journal) SetModeProjector(p ModeProjector) error {
	if p == nil {
		return fmt.Errorf("%w: SetModeProjector was given no projector", ErrInvalidRequest)
	}
	j.modeMu.Lock()
	defer j.modeMu.Unlock()
	if j.modeProjector != nil {
		return fmt.Errorf("%w: the operating-mode projector is already bound; "+
			"a second projection of one mode history is not a configuration", ErrInvalidRequest)
	}
	j.modeProjector = p
	return nil
}

func (j *Journal) modeProjectorRef() ModeProjector {
	j.modeMu.RLock()
	defer j.modeMu.RUnlock()
	return j.modeProjector
}

// TransitionModeRequest asks for a mode change.
type TransitionModeRequest struct {
	// ID is optional; an empty one is derived from the account, the target, the
	// cause and the account's transition count.
	ID         string
	AccountRef string
	// Mode is the target: NORMAL, ENTRY_BLOCKED or HALT_ALL.
	Mode string
	// Cause is required. An automatic tightening must name one of the
	// enumerated triggers; an operator names their reason.
	Cause string
	// Actor is AUTO or OPERATOR.
	Actor string
	// Approval is the human approval reference (§0.7) — a ticket, a name, a
	// runbook step. Required for a relaxation and ignored otherwise.
	Approval string
	// Auditor receives the audit line. Required for a relaxation: the audit
	// record and the relaxation are one act, and one without the other is the
	// state this path exists to prevent.
	Auditor ModeAuditor
	// Announcer is told about a transition that changed something. Optional.
	Announcer ModeAnnouncer
}

// TransitionOperatingMode records a mode change, projects it onto the entry gate
// and announces it.
//
// The bool reports whether anything changed. It is false — with no row appended,
// no projection and no announcement — when an automatic tightening asks for a
// mode no more conservative than the one already in force. That is the
// conservative-precedence rule, and it is also what makes repeated triggers
// idempotent.
func (j *Journal) TransitionOperatingMode(ctx context.Context, req TransitionModeRequest) (OperatingModeRecord, bool, error) {
	account := strings.TrimSpace(req.AccountRef)
	mode := strings.TrimSpace(req.Mode)
	cause := strings.TrimSpace(req.Cause)
	actor := strings.TrimSpace(req.Actor)
	approval := strings.TrimSpace(req.Approval)

	switch {
	case account == "":
		return OperatingModeRecord{}, false, fmt.Errorf(
			"%w: an operating mode is scoped to an account; none was named", ErrInvalidRequest)
	case !ValidOperatingMode(mode):
		return OperatingModeRecord{}, false, fmt.Errorf(
			"%w: operating mode %q is not one this build writes", ErrInvalidRequest, req.Mode)
	case !ValidModeActor(actor):
		return OperatingModeRecord{}, false, fmt.Errorf(
			"%w: mode actor %q is neither %s nor %s", ErrInvalidRequest, req.Actor,
			ModeActorAuto, ModeActorOperator)
	case cause == "":
		return OperatingModeRecord{}, false, fmt.Errorf(
			"%w: a mode transition names why; a change nobody can explain afterwards is not auditable",
			ErrInvalidRequest)
	}
	// The automatic direction rules, decided before any state is read: they are
	// properties of the request, not of the account.
	if actor == ModeActorAuto {
		switch mode {
		case ModeHaltAll:
			return OperatingModeRecord{}, false, ErrHaltAllIsNeverAutomatic
		case ModeNormal:
			// NORMAL is the least conservative mode, so an automatic request for
			// it is a relaxation whatever the account is currently in.
			return OperatingModeRecord{}, false, fmt.Errorf(
				"%w: %s asked for %s", ErrModeRelaxationRequiresOperator, ModeActorAuto, ModeNormal)
		}
		if !AutomaticTrigger(cause) {
			return OperatingModeRecord{}, false, fmt.Errorf(
				"%w: %q (enumerated triggers: %s)", ErrModeTriggerUnknown, cause,
				strings.Join(AutomaticTriggers(), ", "))
		}
	}

	now := j.clk.Now().UTC()
	nowText := formatJournalTime(now)

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return OperatingModeRecord{}, false, fmt.Errorf(
			"journal: starting the operating-mode transition for %s: %w", account, err)
	}
	defer tx.Rollback()

	current, err := currentModeTx(ctx, tx, account)
	if err != nil {
		return OperatingModeRecord{}, false, err
	}

	// Direction is decided against the mode actually in force, inside this
	// transaction — never against what the caller believed it was.
	direction, err := modeDirection(current.Mode, mode)
	if err != nil {
		return OperatingModeRecord{}, false, err
	}
	switch {
	case direction == 0:
		// Already there. Not an error, and not a row: a trigger that fires every
		// poll cycle is the normal case, and one row per cycle would bury the
		// transition that mattered. Reporting changed=false is also what stops
		// the announce/escalate feedback loop (file header).
		return OperatingModeRecord{}, false, nil

	case direction < 0 && actor == ModeActorAuto:
		// Conservative precedence: the account is in a stricter mode than the
		// trigger asks for (HALT_ALL, say, while a daily-loss trigger asks for
		// ENTRY_BLOCKED). The stricter mode stands and nothing is recorded.
		return OperatingModeRecord{}, false, nil

	case direction < 0:
		if approval == "" {
			return OperatingModeRecord{}, false, fmt.Errorf(
				"%w: %s → %s", ErrModeApprovalRequired, current.Mode, mode)
		}
		if req.Auditor == nil {
			return OperatingModeRecord{}, false, fmt.Errorf(
				"%w: relaxing %s → %s requires an audit log; §0.5 requires the change to be traceable",
				ErrInvalidRequest, current.Mode, mode)
		}
		cause = cause + " | approved-by: " + approval
	}

	seq, err := modeRowCount(ctx, tx, account)
	if err != nil {
		return OperatingModeRecord{}, false, err
	}
	id := strings.TrimSpace(req.ID)
	if id == "" {
		id = operatingModeID(account, mode, cause, actor, nowText, seq)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO operating_modes (id, account_ref, mode, cause, actor, created_at)
		 VALUES (?,?,?,?,?,?)`,
		id, account, mode, cause, actor, nowText); err != nil {
		return OperatingModeRecord{}, false, fmt.Errorf(
			"journal: recording the %s → %s transition for %s: %w", current.Mode, mode, account, err)
	}

	record := OperatingModeRecord{
		ID: id, AccountRef: account, Mode: mode, Cause: cause, Actor: actor, CreatedAt: now,
	}

	// The audit line goes in before the commit, exactly as the operator
	// reservation release does: a relaxation whose record is missing is the state
	// this path exists to make impossible, and a failed audit write must abort
	// the transition rather than outlive it.
	if req.Auditor != nil {
		if err := req.Auditor.RecordAction(AuditActionOperatingMode, "operating_mode:"+account,
			mode, actor+" — "+cause+" (from "+current.Mode+")"); err != nil {
			return OperatingModeRecord{}, false, fmt.Errorf(
				"journal: recording the %s → %s transition in the audit log (nothing was changed): %w",
				current.Mode, mode, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return OperatingModeRecord{}, false, fmt.Errorf(
			"journal: committing the %s → %s transition for %s: %w", current.Mode, mode, account, err)
	}

	// Durable first, then projected: see the file header on why this order and
	// not the other one.
	if p := j.modeProjectorRef(); p != nil {
		p.ProjectOperatingMode(record)
	}
	if req.Announcer != nil {
		if err := req.Announcer.AnnounceOperatingMode(ctx, current.Mode, record); err != nil {
			return record, true, fmt.Errorf("%w (%s → %s): %w",
				ErrModeAnnouncementFailed, current.Mode, mode, err)
		}
	}
	return record, true, nil
}

// EscalateOperatingMode is the automatic conservative tightening.
//
// It takes a trigger rather than a target because the mapping is the spec's, not
// the caller's: 자동 강화 트리거와 목적 상태의 열거(SHALL) — 일일 손실 한도 도달 /
// 자격증명 실패 / critical 알림 전달 실패 지속 / exit 관측 두절 / 대사 사이클 지속
// 실패 / 체결 감지 사이클 지속 실패, **전부** ENTRY_BLOCKED. A caller cannot ask for
// HALT_ALL by naming a trigger, because no trigger maps to it.
//
// It needs no approval and no operator: 보수 방향은 자동·즉시·durable.
func (j *Journal) EscalateOperatingMode(ctx context.Context, accountRef, trigger string,
	announcer ModeAnnouncer) (OperatingModeRecord, bool, error) {
	target, ok := TargetModeForTrigger(trigger)
	if !ok {
		return OperatingModeRecord{}, false, fmt.Errorf(
			"%w: %q (enumerated triggers: %s)", ErrModeTriggerUnknown, trigger,
			strings.Join(AutomaticTriggers(), ", "))
	}
	return j.TransitionOperatingMode(ctx, TransitionModeRequest{
		AccountRef: accountRef,
		Mode:       target,
		Cause:      trigger,
		Actor:      ModeActorAuto,
		Announcer:  announcer,
	})
}

// AutomaticTriggers lists the enumerated automatic tightening triggers.
func AutomaticTriggers() []string {
	return []string{
		ModeTriggerDailyLossLimit,
		ModeTriggerCredentialRejected,
		ModeTriggerCriticalAlertUndelivered,
		ModeTriggerExitObservationOutage,
		ModeTriggerReconcileCycleFailure,
		ModeTriggerFillDetectionFailure,
	}
}

// AutomaticTrigger reports whether a cause is an enumerated trigger.
func AutomaticTrigger(cause string) bool {
	_, ok := TargetModeForTrigger(cause)
	return ok
}

// TargetModeForTrigger maps a trigger onto the mode it escalates to.
//
// Every trigger maps to ENTRY_BLOCKED — 전부 메인 스펙의 "신규 진입 차단"과
// 정합하며 HALT_ALL 자동 진입은 없다(SHALL NOT). The function exists rather than a
// constant so that adding a trigger is a visible edit to one table, and so a test
// can assert over the whole set that none of them reaches HALT_ALL.
func TargetModeForTrigger(trigger string) (string, bool) {
	switch strings.TrimSpace(trigger) {
	case ModeTriggerDailyLossLimit,
		ModeTriggerCredentialRejected,
		ModeTriggerCriticalAlertUndelivered,
		ModeTriggerExitObservationOutage,
		ModeTriggerReconcileCycleFailure,
		ModeTriggerFillDetectionFailure:
		return ModeEntryBlocked, true
	default:
		return "", false
	}
}

// CurrentOperatingMode returns the account's mode: the latest row, or NORMAL when
// there is none.
func (j *Journal) CurrentOperatingMode(ctx context.Context, accountRef string) (ModeSnapshot, error) {
	account := strings.TrimSpace(accountRef)
	if account == "" {
		return ModeSnapshot{}, fmt.Errorf(
			"%w: an operating mode is scoped to an account; none was named", ErrInvalidRequest)
	}
	return currentModeFromRow(
		j.db.QueryRowContext(ctx, operatingModeSelect+" WHERE account_ref = ?"+modeLatestOrder, account),
		account)
}

// RestoreOperatingModeProjection rebuilds the entry gate's mode latch from the
// journal.
//
// This is the restart half of "모드·kill switch·이력은 journal 영속·재시작
// 유지(SHALL)". A process that stopped in ENTRY_BLOCKED comes back with an empty
// latch map, and without this call it would trade — the mode row would still be
// there, correct and unread.
//
// It also closes the commit-then-project window: a transition that committed and
// crashed before projecting is projected here.
func (j *Journal) RestoreOperatingModeProjection(ctx context.Context, accountRef string) (ModeSnapshot, error) {
	snapshot, err := j.CurrentOperatingMode(ctx, accountRef)
	if err != nil {
		return ModeSnapshot{}, err
	}
	if p := j.modeProjectorRef(); p != nil {
		p.ProjectOperatingMode(OperatingModeRecord{
			AccountRef: snapshot.AccountRef,
			Mode:       snapshot.Mode,
			Cause:      snapshot.Cause,
			Actor:      snapshot.Actor,
			CreatedAt:  snapshot.Since,
		})
	}
	return snapshot, nil
}

// OperatingModeHistory returns every transition an account has made, oldest
// first. Append-only: this is the whole record, and nothing rewrites it.
func (j *Journal) OperatingModeHistory(ctx context.Context, accountRef string) ([]OperatingModeRecord, error) {
	rows, err := j.db.QueryContext(ctx,
		operatingModeSelect+" WHERE account_ref = ? ORDER BY created_at, rowid",
		strings.TrimSpace(accountRef))
	if err != nil {
		return nil, fmt.Errorf("journal: listing operating-mode history: %w", err)
	}
	defer rows.Close()

	var out []OperatingModeRecord
	for rows.Next() {
		record, err := scanOperatingMode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing operating-mode history: %w", err)
	}
	return out, nil
}

const operatingModeSelect = `SELECT id, account_ref, mode, cause, actor, created_at FROM operating_modes`

// modeLatestOrder is the "current mode" ordering, and it is (created_at, rowid)
// rather than created_at alone.
//
// Journal timestamps are second-resolution (decision.go formatJournalTime), so
// two transitions inside one second are indistinguishable by time. Ordering by
// time alone would let a relaxation that happened *before* a conservative
// tightening read as the later one — the fail-open direction, in the one place
// it must not be. (issues.md 2026-07-26, task 0.1 → 3.1.)
const modeLatestOrder = " ORDER BY created_at DESC, rowid DESC LIMIT 1"

// modeDirection compares the target against the mode in force: +1 tightening,
// 0 no change, −1 relaxation.
//
// A persisted mode this build does not recognise is read as the most
// conservative state there is, so every move away from it is a relaxation and
// needs an operator. That is the same fail-closed reading ModeAllows gives it.
func modeDirection(current, target string) (int, error) {
	targetRank, ok := ModeRank(target)
	if !ok {
		return 0, fmt.Errorf("%w: operating mode %q is not one this build writes", ErrInvalidRequest, target)
	}
	currentRank, known := ModeRank(current)
	if !known {
		return -1, nil
	}
	switch {
	case targetRank > currentRank:
		return 1, nil
	case targetRank < currentRank:
		return -1, nil
	default:
		return 0, nil
	}
}

// currentModeFromRow turns the latest-row query into a snapshot, defaulting an
// account with no rows to NORMAL.
func currentModeFromRow(row rowScanner, account string) (ModeSnapshot, error) {
	record, err := scanOperatingMode(row)
	switch {
	case errors.Is(err, errNoOperatingMode):
		// No row: the account has never been tightened.
		return ModeSnapshot{AccountRef: account, Mode: ModeNormal}, nil
	case err != nil:
		return ModeSnapshot{}, err
	}
	return ModeSnapshot{
		AccountRef: record.AccountRef,
		Mode:       record.Mode,
		Cause:      record.Cause,
		Actor:      record.Actor,
		Since:      record.CreatedAt,
		Recorded:   true,
	}, nil
}

// currentModeTx reads the mode inside the transition's own transaction, so the
// direction check and the append see the same world.
func currentModeTx(ctx context.Context, tx *sql.Tx, account string) (ModeSnapshot, error) {
	return currentModeFromRow(
		tx.QueryRowContext(ctx, operatingModeSelect+" WHERE account_ref = ?"+modeLatestOrder, account),
		account)
}

func modeRowCount(ctx context.Context, tx *sql.Tx, account string) (int, error) {
	var count int
	if err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM operating_modes WHERE account_ref = ?", account).Scan(&count); err != nil {
		return 0, fmt.Errorf("journal: counting operating-mode rows for %s: %w", account, err)
	}
	return count, nil
}

var errNoOperatingMode = errors.New("journal: no operating-mode row")

func scanOperatingMode(row rowScanner) (OperatingModeRecord, error) {
	var (
		record    OperatingModeRecord
		createdAt string
	)
	if err := row.Scan(&record.ID, &record.AccountRef, &record.Mode,
		&record.Cause, &record.Actor, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OperatingModeRecord{}, errNoOperatingMode
		}
		return OperatingModeRecord{}, fmt.Errorf("journal: reading an operating-mode row: %w", err)
	}
	when, err := parseJournalTime(createdAt)
	if err != nil {
		return OperatingModeRecord{}, fmt.Errorf(
			"journal: operating-mode row %s carries an unreadable timestamp: %w", record.ID, err)
	}
	record.CreatedAt = when
	return record, nil
}

// operatingModeID derives a stable id for a transition that did not bring one.
//
// The account's transition count is in the hash because journal timestamps are
// second-resolution: two transitions in one second with the same target and
// cause would otherwise derive the same primary key and the second would be
// refused by SQLite rather than recorded.
func operatingModeID(account, mode, cause, actor, at string, seq int) string {
	h := sha256.New()
	for _, part := range []string{account, mode, cause, actor, at, fmt.Sprint(seq)} {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	return "mode-" + hex.EncodeToString(h.Sum(nil))[:24]
}
