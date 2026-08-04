// Package obs is the engine's observability surface: structured logs and graded
// operator alerts.
//
// # Two channels, two guarantees
//
// The engine produces two kinds of output and they are not interchangeable:
//
//   - Structured logs are complete and local. Everything the engine does is
//     written, with a countable event name, and nothing about a log line can fail
//     in a way the engine has to react to.
//   - Alerts are selective and remote. They go to a human, over a network, and
//     they can be lost — so they are graded, and the critical grade is durable
//     (internal/journal's alert outbox) and gates trading when it cannot be
//     delivered.
//
// Conflating the two is how monitoring systems end up either silent or useless.
// engine-safety states the grading directly: critical events (IN_DOUBT,
// UNRESOLVED_IN_DOUBT, credential expiry imminent, permanent mismatch,
// UNKNOWN_BROKER_STATE) go through the outbox and block new entries on persistent
// failure; ordinary ones (fills, state transitions) are best-effort.
//
// # Heartbeats, because a dead process cannot report its own death
//
// The failure this system most needs to detect is the engine not running. No
// amount of error handling inside the engine covers that, so the engine publishes
// a heartbeat on a fixed interval and the *receiving* side alerts when one does
// not arrive. That inverts the responsibility onto something that is still alive.
package obs

import "strings"

// EventType names a countable engine event.
//
// The strings are a contract: they are the log's `event` field, the ntfy topic
// tag, and the outbox's event_type column. They are grouped by subject and
// appended to, never renamed — a dashboard or an alert rule keyed on one of these
// is a thing somebody built.
type EventType string

const (
	// --- order lifecycle ----------------------------------------------------

	// EventOrderIntent is an intent recorded in the journal, before dispatch.
	EventOrderIntent EventType = "order.intent_recorded"
	// EventOrderDispatched is a mutation sent to the broker.
	EventOrderDispatched EventType = "order.dispatched"
	// EventOrderConfirmed is a mutation the broker acknowledged.
	EventOrderConfirmed EventType = "order.confirmed"
	// EventOrderRejected is a mutation the broker definitively refused.
	EventOrderRejected EventType = "order.rejected"
	// EventOrderNotDispatched is a mutation that provably never left.
	EventOrderNotDispatched EventType = "order.not_dispatched"
	// EventOrderInDoubt is a mutation whose outcome the transport could not
	// establish. CRITICAL.
	EventOrderInDoubt EventType = "order.in_doubt"
	// EventOrderUnresolved is an IN_DOUBT attempt the resolver could not settle.
	// CRITICAL.
	EventOrderUnresolved EventType = "order.unresolved_in_doubt"

	// --- fills --------------------------------------------------------------

	// EventFillObserved is a positive fill delta.
	EventFillObserved EventType = "fill.observed"
	// EventFillRefused is a snapshot the fill ledger refused (a shrinking
	// cumulative quantity, an out-of-order observation). CRITICAL.
	EventFillRefused EventType = "fill.refused"
	// EventFillSLOViolated is fill detection falling behind its objective.
	EventFillSLOViolated EventType = "fill.slo_violated"

	// --- broker state -------------------------------------------------------

	// EventBrokerStateUnknown is a derivation that fell through to
	// UNKNOWN_BROKER_STATE. CRITICAL.
	EventBrokerStateUnknown EventType = "broker.state_unknown"
	// EventBrokerAuthRejected is a 401/403. CRITICAL — an engine with a dead
	// credential cannot exit a position either.
	EventBrokerAuthRejected EventType = "broker.auth_rejected"
	// EventCredentialExpiring is a capability attestation or credential about to
	// lapse. CRITICAL, because the alternative is finding out at the next
	// restart, which is too late.
	EventCredentialExpiring EventType = "broker.credential_expiring"

	// --- reconciliation -----------------------------------------------------

	// EventReconcileClean is an agreeing reconciliation.
	EventReconcileClean EventType = "reconcile.clean"
	// EventReconcileMismatch is a disagreement within its retry budget.
	EventReconcileMismatch EventType = "reconcile.mismatch"
	// EventReconcilePermanent is a disagreement that survived its retries.
	// CRITICAL.
	EventReconcilePermanent EventType = "reconcile.mismatch_permanent"
	// EventRecoveryComplete is the restart recovery sequence finishing.
	EventRecoveryComplete EventType = "reconcile.recovery_complete"

	// --- flatten ------------------------------------------------------------

	// EventFlattenStarted is a flatten saga beginning. CRITICAL: somebody or
	// something decided to exit every position, and that is never routine.
	EventFlattenStarted EventType = "flatten.started"
	// EventFlattenProgress is one saga step completing.
	EventFlattenProgress EventType = "flatten.progress"
	// EventFlattenComplete is a saga finishing with nothing left.
	EventFlattenComplete EventType = "flatten.complete"
	// EventFlattenStalled is a saga that could not finish. CRITICAL.
	EventFlattenStalled EventType = "flatten.stalled"

	// --- engine lifecycle ---------------------------------------------------

	// EventEngineStarted is the engine coming up.
	EventEngineStarted EventType = "engine.started"
	// EventEngineHeartbeat is the periodic "still alive" publish.
	EventEngineHeartbeat EventType = "engine.heartbeat"
	// EventEntryBlocked is the entry gate refusing new exposure.
	EventEntryBlocked EventType = "engine.entry_blocked"
	// EventAlertUndelivered is a critical alert that could not be delivered.
	// CRITICAL, and the one that blocks entries by itself.
	EventAlertUndelivered EventType = "engine.alert_undelivered"
	// EventEngineLoopFailed is a supervised loop that returned for a reason other
	// than the runtime being cancelled (add-engine-runtime: 방어적 종료 계약).
	// CRITICAL: the landed loops do not return, so one that did has hit a state
	// nobody wrote a recovery for — and the runtime stops the rest rather than
	// leaving a partially alive engine trading on half its senses.
	EventEngineLoopFailed EventType = "engine.loop_failed"
	// EventEngineLoopDegraded is a supervised loop that is alive but whose cycles
	// have failed consecutively past the threshold (add-engine-runtime: 지속 열화
	// 임계). CRITICAL, and it is what carries the automatic ENTRY_BLOCKED: the
	// loop keeps retrying, and the account stops taking new exposure meanwhile.
	EventEngineLoopDegraded EventType = "engine.loop_degraded"

	// --- exit policy ----------------------------------------------------------
	//
	// The grading rule across this group is one question: does the condition mean
	// a position is *not being protected*? Where the answer is yes the event is
	// critical, because until 2c puts a stop on the broker the only protection an
	// open position has is this loop's judgement.

	// EventExitObservationOutage is price observation failing past the staleness
	// threshold (exit-policy: 관측 두절 60초 → critical 알림 + ENTRY_BLOCKED).
	// CRITICAL: with no broker-resident stop, an unobserved position is an
	// unprotected one.
	EventExitObservationOutage EventType = "exit.observation_outage"
	// EventExitJudgementRefused is an evaluation that could not run at all — an
	// invariant violation in the stored state, or a price the policy cannot use
	// (exitpolicy.ErrRefused). CRITICAL, and distinct from "nothing to do": the
	// position went unjudged.
	EventExitJudgementRefused EventType = "exit.judgement_refused"
	// EventExitCycleFailed is one exit observation cycle that ended on a failure
	// (change a074). Normal, and the grade is the decision.
	//
	// ExitCycle.Err has no single meaning: a ledger write that lost a race, a
	// break-even that could not be priced for one symbol, a judgement transaction
	// that quarantined a generation. Grading that critical would send a transient
	// SQLite error to the durable outbox, and an outbox entry that cannot be
	// delivered latches the entry gate and escalates the account to ENTRY_BLOCKED.
	// A blip must not be able to stop a live account trading.
	//
	// So this is the counting channel, and it is complete: every failed cycle
	// produces one line. The conditions that genuinely mean a position is
	// unprotected have their own critical events — observation outage, judgement
	// refused, snapshot quarantined, liquidation delayed — and those are what
	// reach a human.
	EventExitCycleFailed EventType = "exit.cycle_failed"
	// EventExitSnapshotQuarantined is the moment a stored exit snapshot generation
	// is quarantined (change a074). CRITICAL, by the same rule as the rest of this
	// group: a quarantined generation is refused outright by the observation loop,
	// so the position is not judged at all and its stop is not evaluated.
	//
	// It is deliberately separate from EventExitJudgementRefused, which reports the
	// *consequence* on every later cycle. Three things made the consequence alone
	// insufficient. It arrives one cycle late, because the quarantine is written
	// inside the judgement transaction and only the next working set sees it. It
	// carries none of the quarantine's identity — version, reason, evidence — so
	// nothing in it distinguishes "unreadable price" from "this generation is
	// permanently out of the judgement set until a human lifts it". And it is
	// latched per position, so a position already refused for another reason could
	// be quarantined without producing a single line.
	EventExitSnapshotQuarantined EventType = "exit.snapshot_quarantined"
	// EventExitProposalRefused is a liquidation or take-profit the Guardian or the
	// gateway declined to submit. CRITICAL — the protection the policy asked for
	// did not happen.
	EventExitProposalRefused EventType = "exit.proposal_refused"
	// EventExitProposalCapped is a liquidation bounded by the RECONCILE confirmed
	// floor, with the remainder left pending (exit-policy: 캡 발생은 알림된다).
	// Normal: part of the exit did go, and the remainder re-proposes itself.
	EventExitProposalCapped EventType = "exit.proposal_capped"
	// EventExitLiquidationDelayed is a breach liquidation held back past the
	// resolution bound by an unsettled attempt on the same symbol (exit-policy:
	// 지연이 유계를 넘으면 critical 알림). CRITICAL.
	EventExitLiquidationDelayed EventType = "exit.liquidation_delayed"
	// EventExitPositionUnmanaged is a held position the exit policy will not
	// manage: neither an entry decision nor an adoption record, so no stop to
	// build a baseline from (exit-policy: entry 결정도 편입 기록도 없는 포지션 …
	// 발견 시 알림). Normal — somebody trading their own account by hand is not a
	// malfunction.
	//
	// It is raised regardless of `adoption.enabled` (design A4): the landed
	// behaviour is that an operator is told when the engine is trading beside a
	// position it will not protect, and a feature toggle must not silence it.
	// What the toggle changes is how often it has anything to say.
	EventExitPositionUnmanaged EventType = "exit.position_unmanaged"
	// EventExitPositionAdopted is an externally acquired holding taken into exit
	// management (change adopt-external-positions). It carries the observation the
	// synthetic t0 was built from and the stop derived from it, because those two
	// numbers are the whole of what the engine will now protect the position to.
	//
	// Normal, and it *replaces* the unmanaged alert for that position rather than
	// arriving beside it: the operator's question is "is this protected", and two
	// events answering it differently within one cycle is worse than either.
	EventExitPositionAdopted EventType = "exit.position_adopted"
	// EventExitPositionClosedExternally is a position the exit policy was managing
	// that went to zero outside the engine (design A7). The exit state is
	// completed with an ADJUSTMENT_CLOSED event and no trade outcome is frozen.
	//
	// Normal, for the same reason the fold is: a person selling their own shares
	// is not a malfunction, and grading it critical would mean an engine with no
	// alert transport configured stops opening positions every time its owner
	// takes a profit by hand.
	EventExitPositionClosedExternally EventType = "exit.position_closed_externally"

	// --- operating mode -------------------------------------------------------

	// EventOperatingMode is a committed operating-mode transition
	// (add-core-domain task 3.3). CRITICAL in both directions: a tightening
	// means the engine stopped opening positions on its own, and a relaxation
	// means somebody re-enabled entries on a live account.
	EventOperatingMode EventType = "engine.operating_mode"

	// --- measurement ----------------------------------------------------------
	//
	// One rule governs this group and it is a SHALL NOT: measurement failures are
	// never critical. risk-management's "분석·성과 작업 실패는 트리거가 아니다" is
	// the requirement, and the mechanism it protects against is this file's own
	// grading table — an event listed there goes to the durable outbox, and an
	// outbox entry that cannot be delivered latches the entry gate and escalates
	// the operating mode to ENTRY_BLOCKED, which only an operator can lift.
	//
	// So a typo in a v8 column name must not be able to stop a live account
	// trading. Membership in criticalEvents is the only switch that could do that,
	// and measurement_test.go asserts the non-membership rather than trusting it.

	// EventEntryObservationFailed is an entry verdict whose observation could not
	// be written (change add-net-rr-measurement). The verdict itself stands: the
	// observation is outside the issuance transaction precisely so that losing it
	// cannot roll a decision back.
	//
	// Normal, deliberately. What is lost is a row in an analysis table; nothing
	// about the account's safety changed, and the loss is separately counted in a
	// store that does not share a failure domain with the one that just failed
	// (internal/measure/degrade).
	EventEntryObservationFailed EventType = "measurement.entry_observation_failed"
	// EventEntryObservationReconstructed is a lost observation rebuilt from a
	// decision preimage. Normal: it is the repair working, and the row carries a
	// marker saying its ratios were recomputed under today's rates rather than the
	// ones the verdict used.
	EventEntryObservationReconstructed EventType = "measurement.entry_observation_reconstructed"

	// --- Phase 2, reserved --------------------------------------------------
	//
	// Declared now so the enum is stable before the features exist: a consumer
	// written against this list should not have to change when they land.

	// EventKillSwitch is reserved for the Phase 2 kill switch.
	EventKillSwitch EventType = "engine.kill_switch"
)

// Severity is the alert grade.
type Severity string

const (
	// SeverityCritical is durable: written to the outbox before any send, and
	// blocking new entries while it stays undelivered.
	SeverityCritical Severity = "critical"
	// SeverityNormal is best-effort: sent if it can be, dropped if it cannot.
	SeverityNormal Severity = "normal"
)

// criticalEvents is the grading table. It is a map rather than a switch so a test
// can assert that every event engine-safety names as critical is in it.
var criticalEvents = map[EventType]bool{
	EventOrderInDoubt:       true,
	EventOrderUnresolved:    true,
	EventFillRefused:        true,
	EventBrokerStateUnknown: true,
	EventBrokerAuthRejected: true,
	EventCredentialExpiring: true,
	EventReconcilePermanent: true,
	EventFlattenStarted:     true,
	EventFlattenStalled:     true,
	EventAlertUndelivered:   true,
	EventOperatingMode:      true,
	EventEngineLoopFailed:   true,
	EventEngineLoopDegraded: true,

	EventExitObservationOutage:   true,
	EventExitJudgementRefused:    true,
	EventExitSnapshotQuarantined: true,
	EventExitProposalRefused:     true,
	EventExitLiquidationDelayed:  true,
}

// SeverityOf grades an event.
//
// An event type this build does not know is graded normal, not critical. That
// looks like the wrong direction until you consider what the alternative does: an
// unrecognised string would durably enqueue an alert nobody can act on and, on
// delivery failure, stop the engine trading. Unknown strings come from typos and
// from newer components, and neither is a reason to halt a live account.
// Genuinely critical conditions are named in the table above.
func SeverityOf(t EventType) Severity {
	if criticalEvents[t] {
		return SeverityCritical
	}
	return SeverityNormal
}

// CriticalEvents lists the critical event types, for tests and documentation.
func CriticalEvents() []EventType {
	out := make([]EventType, 0, len(criticalEvents))
	for t := range criticalEvents {
		out = append(out, t)
	}
	return out
}

// Subject is the leading part of an event name ("order", "fill", "reconcile", …).
func (t EventType) Subject() string {
	name := string(t)
	if i := strings.Index(name, "."); i > 0 {
		return name[:i]
	}
	return name
}
