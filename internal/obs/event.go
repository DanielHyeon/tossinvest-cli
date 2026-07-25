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

	// --- Phase 2, reserved --------------------------------------------------
	//
	// Declared now so the enum is stable before the features exist: a consumer
	// written against this list should not have to change when they land.

	// EventKillSwitch is reserved for the Phase 2 kill switch.
	EventKillSwitch EventType = "engine.kill_switch"
	// EventOperatingMode is reserved for Phase 2 operating-mode changes.
	EventOperatingMode EventType = "engine.operating_mode"
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
