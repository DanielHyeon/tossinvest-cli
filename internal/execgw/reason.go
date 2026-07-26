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

	// --- the persisted decision contract (extend-execution-contract 1.4/1.5) --
	//
	// Appended, never reordered: the codes above are already on disk.

	// ReasonGuardianDecisionTampered: the persisted decision row is not
	// internally consistent — the hash does not cover the preimage, the preimage
	// is not canonical, the class and the preimage schema disagree, or the row
	// carries something no writer in this build produces. It is a distinct code
	// from a plain mismatch because it means the *record* is wrong, not the
	// order, and no operator should try to fix that by re-submitting.
	ReasonGuardianDecisionTampered ReasonCode = "guardian_decision_tampered"
	// ReasonGuardianKeyMismatch: the idempotency key the decision carries is not
	// the one derived from it, is missing on a place, or is present on a
	// mutation that cannot send one.
	ReasonGuardianKeyMismatch ReasonCode = "guardian_idempotency_key_mismatch"
	// ReasonGuardianClassMismatch: the decision's safety class and the shape of
	// the mutation disagree — an EXPOSURE_RAISING decision for something that
	// reduces exposure, or a RISK_REDUCING decision for a buy. A class that only
	// the caller asserts is a limit bypass waiting to be used.
	ReasonGuardianClassMismatch ReasonCode = "guardian_safety_class_mismatch"

	// --- idempotent replay (extend-execution-contract 2.1/2.2) ---------------
	//
	// Appended, never reordered. These describe the *replay* — the resend of a
	// stored body under its stored idempotency key — and none of them is a
	// statement about whether the original order exists. That distinction is why
	// the replay path has its own vocabulary instead of reusing the dispatch
	// codes above.

	// ReasonReplayRecovered: the replay returned the original order's identity
	// and the attempt is CONFIRMED. No second order was created — inside the
	// key's ten-minute window the same key cannot make one (openapi).
	ReasonReplayRecovered ReasonCode = "replay_identity_recovered"
	// ReasonReplayInProgress: `409 request-in-progress` (openapi). The broker is
	// still processing the original request, so the replay established nothing
	// and does not consume the cap.
	ReasonReplayInProgress ReasonCode = "replay_request_in_progress"
	// ReasonReplayKeyConflict: the key does not name this attempt's order —
	// `422 idempotency-key-conflict` (openapi), or a 2xx echoing a different
	// key. It is a defect in this program and is never FAILED_CONFIRMED: the
	// attempt parks for an operator.
	ReasonReplayKeyConflict ReasonCode = "replay_idempotency_key_conflict"
	// ReasonReplayNotAttested: the replay capability attestation is not in
	// place, so nothing was resent. Default state until the capability is
	// measured [미측정 — 2b 전 비활성].
	ReasonReplayNotAttested ReasonCode = "replay_not_attested"
	// ReasonReplayExpired: too little of the idempotency key's documented
	// ten-minute life is left for a replay to be safe. Re-checked before every
	// individual send, never once per procedure.
	ReasonReplayExpired ReasonCode = "replay_key_window_expired"
	// ReasonReplayExhausted: the per-attempt replay cap is spent, or the broker
	// stayed "in progress" past the wait bound. The query fallback takes over.
	ReasonReplayExhausted ReasonCode = "replay_attempts_exhausted"
	// ReasonReplayIneligible: the attempt is not one a replay is defined for —
	// wrong state, no stored body, no key, or a mutation kind that carries none.
	ReasonReplayIneligible ReasonCode = "replay_ineligible"

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
	// journal rows on disk.
	//
	// Two of these clear themselves and three do not, and the split is the same
	// one the gate already makes. A condition a later observation disproves
	// (recovery finished; detection caught up) is released automatically. A
	// condition that means our model of the account was wrong
	// (UNKNOWN_BROKER_STATE, a mismatch that survived its retries) stays latched
	// until a human has looked, because "it stopped reporting the contradiction"
	// is not evidence that the contradiction was resolved.

	// ReasonRecoveryIncomplete: the restart recovery sequence has not finished.
	// Until it has, the engine does not know what it already has on the account,
	// so new exposure is refused. Clears when recovery completes.
	ReasonRecoveryIncomplete ReasonCode = "recovery_incomplete"
	// ReasonFillDetectionSLO: fill detection has been slower than its objective
	// for longer than the grace period. Clears when the measurement recovers.
	ReasonFillDetectionSLO ReasonCode = "fill_detection_slo_violated"
	// ReasonBrokerStateUnknown: a broker snapshot could not be reconciled with
	// what was already observed (a shrinking cumulative fill, an out-of-order
	// snapshot, a status this build does not understand). Fail-closed, and
	// operator-cleared: the specific rule that fired travels in the detail.
	ReasonBrokerStateUnknown ReasonCode = "unknown_broker_state"
	// ReasonReconcileMismatch: local state and the account disagree beyond the
	// documented tolerance. Clears on a successful reconciliation.
	ReasonReconcileMismatch ReasonCode = "reconciliation_mismatch"
	// ReasonReconcilePermanent: reconciliation failed the configured number of
	// times in a row. Operator resolution only — the automatic retry has already
	// been shown not to work.
	ReasonReconcilePermanent ReasonCode = "reconciliation_mismatch_permanent"

	// --- operating mode (add-core-domain task 3.1) ---------------------------

	// ReasonOperatingModeBlocked: the account's operating mode refuses new
	// exposure (ENTRY_BLOCKED or HALT_ALL). It is a projection of the journal's
	// latest `operating_modes` row, not a judgement made here — the journal is
	// the authority and this latch is how the sealed submission sequence
	// consumes it without changing.
	//
	// It never auto-clears: relaxing the mode is a human decision with an audit
	// line (§0.7), and the release path is a mode transition, not anything the
	// gate can do on its own.
	ReasonOperatingModeBlocked ReasonCode = "operating_mode_blocked"

	// --- flatten (task 4.4) --------------------------------------------------

	// ReasonFlattenInProgress: a flatten-all saga is running or has run. New
	// exposure is refused for the duration and afterwards.
	//
	// It does not auto-clear, and that is the point rather than an oversight:
	// somebody decided to exit every position on the account, and deciding to
	// start again is a separate human decision (§0.7). Exits are untouched —
	// nothing about this code path can refuse a cancel or a liquidation (§0.3).
	ReasonFlattenInProgress ReasonCode = "flatten_in_progress"

	// --- the reservation as authority (add-core-domain task 5.1) -------------

	// ReasonGuardianReservationMissing: an EXPOSURE_RAISING decision reached
	// submission without a HELD risk reservation.
	//
	// The aggregate limits are enforced by the reservation ledger, not by the
	// decision (engine-safety: "총계 한도의 최종 권위는 예약 트랜잭션"), so an entry
	// decision that never took the headroom it needs was never paid for. The
	// atomic issuance API makes that state unreachable from the issuer's side;
	// this code is the other side of the same statement, and the two together are
	// what make "예약 없는 진입 결정은 거부된다" true rather than merely intended.
	//
	// RISK_REDUCING decisions never carry one: an exit lowers the aggregate it
	// would otherwise be reserving against, and a limit that could refuse a
	// liquidation is a trap (§0.3).
	ReasonGuardianReservationMissing ReasonCode = "guardian_reservation_missing"
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
