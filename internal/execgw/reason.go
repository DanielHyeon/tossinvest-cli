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
	// ReasonSymbolInFlight: another mutation on this symbol has not settled.
	// The engine holds at most one in-flight mutation per symbol, which is what
	// makes IN_DOUBT fingerprint matching unique by construction.
	ReasonSymbolInFlight ReasonCode = "symbol_mutation_in_flight"

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

	// --- fail-closed branches (task 2.10) -----------------------------------

	// ReasonBalanceInsufficient: the currency balance does not cover the buy.
	// The engine never converts currency to close the gap.
	ReasonBalanceInsufficient ReasonCode = "currency_balance_insufficient"
	// ReasonBalanceUnavailable: the balance could not be read. Not knowing what
	// is in the account is not permission to spend it.
	ReasonBalanceUnavailable ReasonCode = "currency_balance_unavailable"
	// ReasonInteractiveAuthRequired: the broker wants a human to authenticate.
	// There is no automatic answer to this by design.
	ReasonInteractiveAuthRequired ReasonCode = "interactive_auth_required"
	// ReasonFXConsentRequired: settling would need a currency conversion the
	// operator has to consent to.
	ReasonFXConsentRequired ReasonCode = "fx_consent_required"
	// ReasonFundingRequired: the account needs a deposit first.
	ReasonFundingRequired ReasonCode = "funding_required"

	// --- fill detection & reconciliation (tasks 3.1-3.6) ---------------------
	//
	// Appended, never reordered: the codes above are already written into
	// journal rows on disk. Unlike the auth latch, every code in this block
	// describes a condition that a later successful observation disproves, so
	// they auto-clear — except the permanent-mismatch one, which by definition
	// means observation has stopped helping.

	// ReasonRecoveryIncomplete: the restart recovery sequence has not finished.
	// Until it has, the engine does not know what it already has on the account,
	// so new exposure is refused. Clears when recovery completes.
	ReasonRecoveryIncomplete ReasonCode = "recovery_incomplete"
	// ReasonFillDetectionSLO: fill detection has been slower than its objective
	// for longer than the grace period. Clears when the measurement recovers.
	ReasonFillDetectionSLO ReasonCode = "fill_detection_slo_violated"
	// ReasonBrokerStateUnknown: a broker snapshot could not be reconciled with
	// what was already observed (a shrinking cumulative fill, an out-of-order
	// snapshot, a status this build does not understand). Fail-closed.
	ReasonBrokerStateUnknown ReasonCode = "unknown_broker_state"
	// ReasonReconcileMismatch: local state and the account disagree beyond the
	// documented tolerance. Clears on a successful reconciliation.
	ReasonReconcileMismatch ReasonCode = "reconciliation_mismatch"
	// ReasonReconcilePermanent: reconciliation failed the configured number of
	// times in a row. Operator resolution only — the automatic retry has already
	// been shown not to work.
	ReasonReconcilePermanent ReasonCode = "reconciliation_mismatch_permanent"
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
