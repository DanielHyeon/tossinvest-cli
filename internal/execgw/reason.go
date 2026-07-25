package execgw

import "fmt"

// ReasonCode is a stable, machine-readable reason for a refusal or an outcome.
//
// The strings are a contract, not debug text: they are written into the journal's
// reason_code column, into structured logs, and into operator alerts, and Phase 2's
// ledger imports them. Renaming one is a breaking change to records that already
// exist on disk; add a new code instead.
type ReasonCode string

const (
	// --- guardian (engine-safety "ExecutionGateway 봉인") --------------------

	// ReasonGuardianMissing: no GuardianDecision, or one without a nonce.
	ReasonGuardianMissing ReasonCode = "guardian_decision_missing"
	// ReasonGuardianExpired: the decision's expiry passed before the broker call.
	ReasonGuardianExpired ReasonCode = "guardian_decision_expired"
	// ReasonGuardianIntentMismatch: the decision authorises a different intent.
	ReasonGuardianIntentMismatch ReasonCode = "guardian_intent_hash_mismatch"
	// ReasonGuardianNonceReused: the one-shot nonce was already spent.
	ReasonGuardianNonceReused ReasonCode = "guardian_nonce_reused"
	// ReasonGuardianLimitsUnset: the limit snapshot is empty. A risk-increasing
	// mutation authorised by "no limits" is not authorised at all.
	ReasonGuardianLimitsUnset ReasonCode = "guardian_limits_unset"
	// ReasonGuardianLimitExceeded: the mutation is bigger than the snapshot allows.
	ReasonGuardianLimitExceeded ReasonCode = "guardian_limit_exceeded"

	// --- request / policy ---------------------------------------------------

	// ReasonInvalidRequest: the mutation request itself is not recordable
	// (unknown market, missing symbol, non-positive quantity).
	ReasonInvalidRequest ReasonCode = "invalid_mutation_request"
	// ReasonPolicyDisabled: the user's trading config refuses this action.
	ReasonPolicyDisabled ReasonCode = "trading_policy_disabled"
	// ReasonUnsupportedOrderType: the order shape is outside what the official
	// path supports.
	ReasonUnsupportedOrderType ReasonCode = "unsupported_order_type"

	// --- broker outcomes ----------------------------------------------------

	// ReasonBrokerAccepted: the broker acknowledged the mutation and named an order.
	ReasonBrokerAccepted ReasonCode = "broker_accepted"
	// ReasonBrokerRejected: a well-formed refusal that definitively did not execute.
	ReasonBrokerRejected ReasonCode = "broker_rejected"
	// ReasonBrokerOutcomeUnknown: the transport or the status does not prove
	// whether the mutation executed.
	ReasonBrokerOutcomeUnknown ReasonCode = "broker_outcome_unknown"

	// --- entry gate (retry matrix, task 2.6) --------------------------------

	// ReasonBrokerAuthRejected: the broker refused the credential (401/403).
	// Latching, not auto-clearing: a credential that came back is not evidence
	// the operator knows what happened.
	ReasonBrokerAuthRejected ReasonCode = "broker_auth_rejected"
	// ReasonQueryStale: a required read is older than its staleness threshold.
	// Auto-clears when the read succeeds again.
	ReasonQueryStale ReasonCode = "required_query_stale"
	// ReasonUnresolvedInDoubt: an attempt ended UNRESOLVED_IN_DOUBT. Operator
	// resolution only.
	ReasonUnresolvedInDoubt ReasonCode = "unresolved_in_doubt"
	// ReasonAlertUndelivered: reserved for task 4.3 — a critical alert that could
	// not be delivered blocks new entries. Declared here so the enum is stable
	// before the alerting path exists.
	ReasonAlertUndelivered ReasonCode = "critical_alert_undelivered"
)

// RejectedError is a refusal produced by the gateway itself: the mutation was
// journalled and closed without any broker contact.
type RejectedError struct {
	Reason ReasonCode
	Detail string
}

func (e *RejectedError) Error() string {
	if e == nil {
		return "execgw: rejected"
	}
	if e.Detail == "" {
		return fmt.Sprintf("execgw: rejected (%s)", e.Reason)
	}
	return fmt.Sprintf("execgw: rejected (%s): %s", e.Reason, e.Detail)
}

func reject(reason ReasonCode, format string, args ...any) *RejectedError {
	return &RejectedError{Reason: reason, Detail: fmt.Sprintf(format, args...)}
}
