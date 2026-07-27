// Package risk is the Guardian judgement chain: the fixed sequence every
// automated intent passes before a decision may be issued.
//
// It is the artifact risk-management asks for — "모든 자동 진입 의도는 Guardian
// 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서
// 정지한다" — and nothing else in this package does anything else.
//
// # What this package is not
//
//   - It is not the issuer. Evaluate answers "may this happen"; recording the
//     decision and holding the reservation is one atomic journal transaction
//     that belongs to task 4.1, and it is deliberately not reachable from here.
//   - It does not read the journal, the broker or the clock. Every input arrives
//     in Input, which is what makes the chain a pure function of a snapshot: the
//     same snapshot always produces the same verdict, and a test can construct
//     any account state without a database.
//   - It carries no signal-layer input (SHALL — 체인의 모든 입력은 의도 필드·
//     브로커 스냅샷·journal 상태에서 온다: 구조적 RR 계산·등급배수는 P3), and no
//     protective- or conditional-order code (2c).
//
// # The order is the interface
//
// Which check runs first decides which reason an operator reads after a
// refusal, so the order is pinned by a test and documented in
// docs/guardian-chain.md alongside the StockOS mapping. The authority for it is
// the risk-management table, not StockOS's `evaluate_guardian`: the original
// chain mixes in steps TossOS excludes, and the minimum reward:risk is a check
// the original does not have at all.
package risk

// ReasonCode is a stable, machine-readable reason for a chain verdict.
//
// The strings are a contract in the same sense execgw.ReasonCode's are: they
// reach journal rows, structured logs and operator alerts, and a rename rewrites
// records that already exist. Add a code rather than repurpose one.
//
// They are UPPER_SNAKE because that is the vocabulary risk-management writes
// them in, and because the visible difference from execgw's lower_snake gateway
// codes is useful: a Guardian refusal ("this must not be attempted") and a
// gateway refusal ("this attempt cannot be submitted") are different events, and
// task 2.4 maps between them explicitly rather than by string coincidence.
type ReasonCode string

const (
	// ReasonAllowed: every check passed. The chain says nothing about whether
	// the reservation will succeed — 총계 한도의 최종 권위는 예약 트랜잭션이다.
	ReasonAllowed ReasonCode = "ALLOWED"

	// ReasonInputUnavailable: an input the chain needs is absent, unreadable or
	// in a currency it cannot compare. Fail-closed: 정의되지 않은 양에 예약을
	// 걸어서는 안 된다. It is a distinct code from every policy refusal because
	// the operator action is different — fix the wiring, not the limits.
	ReasonInputUnavailable ReasonCode = "INPUT_UNAVAILABLE"

	// --- 1. kill switch and operating mode ---------------------------------

	// ReasonKillSwitchActive: the kill switch is on. BLOCK-ONLY — it refuses new
	// entries and never causes a liquidation, in either direction.
	ReasonKillSwitchActive ReasonCode = "KILL_SWITCH_ACTIVE"
	// ReasonOperatingModeBlocked: the account's operating mode refuses
	// EXPOSURE_RAISING (ENTRY_BLOCKED or HALT_ALL), or is a value this build does
	// not recognise.
	ReasonOperatingModeBlocked ReasonCode = "OPERATING_MODE_BLOCKED"

	// --- 2. entry gate latch -----------------------------------------------

	// ReasonEntryGateBlocked: the entry latch is set (401/403 · SLO violation ·
	// RECONCILE · recovery incomplete). The latch's own reason travels in the
	// decision's detail; task 2.4 owns the mapping.
	ReasonEntryGateBlocked ReasonCode = "ENTRY_GATE_BLOCKED"

	// --- 3. allowlist -------------------------------------------------------

	// ReasonSymbolNotAllowed: the symbol is not on the allowlist. This is the
	// structural replacement for StockOS's instrument-class blocks
	// (leveraged/inverse, ETF/ETN) — 분류 소스 `[미측정]`, P3 재설계.
	ReasonSymbolNotAllowed ReasonCode = "SYMBOL_NOT_ALLOWED"

	// --- 4. the stop contract ------------------------------------------------

	// ReasonStopMissing: no stop price. No Stop = No Trade, and the refusal
	// happens before any sizing arithmetic runs.
	ReasonStopMissing ReasonCode = "STOP_MISSING"
	// ReasonStopNotBelowEntry: the stop does not protect a long. TossOS is
	// long-only, so a protective stop is strictly below the entry.
	ReasonStopNotBelowEntry ReasonCode = "STOP_NOT_BELOW_ENTRY"
	// ReasonTargetNotAboveEntry: the target is at or below the entry.
	ReasonTargetNotAboveEntry ReasonCode = "TARGET_NOT_ABOVE_ENTRY"
	// ReasonInvalidTargetStop: a target or stop that is absent, unreadable,
	// non-finite or not a positive price. It is separate from the two ordering
	// codes above because the value is unusable rather than badly placed.
	ReasonInvalidTargetStop ReasonCode = "INVALID_TARGET_STOP"
	// ReasonTargetBelowBreakEven: the target is below the cost-inclusive
	// break-even sell price — a trade that loses money when it works.
	ReasonTargetBelowBreakEven ReasonCode = "TARGET_BELOW_BREAK_EVEN"

	// --- 5. size --------------------------------------------------------------

	// ReasonInvalidOrderSize: the quantity is absent, not a positive whole
	// number, or above what the per-trade risk budget allows over this stop.
	ReasonInvalidOrderSize ReasonCode = "INVALID_ORDER_SIZE"
	// ReasonMaxOrderExceeded: the order is past a configured per-order ceiling
	// (quantity or notional). The ceiling is inclusive — reaching it is allowed.
	ReasonMaxOrderExceeded ReasonCode = "MAX_ORDER_EXCEEDED"
	// ReasonSellExceedsHoldings: a sell larger than the holding, which would open
	// a short. Reductions are reduce-only (SHALL NOT — short 노출 구조적 금지).
	ReasonSellExceedsHoldings ReasonCode = "SELL_EXCEEDS_HOLDINGS"

	// --- 6. minimum reward:risk ------------------------------------------------

	// ReasonMinRRNotMet: the **gross** ratio (target − entry) / (entry − stop) is
	// below Policy.MinRewardRisk, or cannot be computed. A ratio nobody could
	// compute is not a ratio of zero.
	//
	// Gross, and only gross. The net ratio — (target − B) / (B − stop), where B is
	// the cost-inclusive break-even — exists beside it as an *observation*
	// (risk.NetRewardRisk, change add-net-rr-measurement) and no code raises this
	// reason on account of it. There is deliberately no second code: the closest
	// thing to a net-basis refusal the chain has is TARGET_BELOW_BREAK_EVEN at the
	// stop-contract rung, which catches the case where costs eat the whole target.
	ReasonMinRRNotMet ReasonCode = "MIN_RR_NOT_MET"

	// --- 7. cash ---------------------------------------------------------------

	// ReasonInsufficientCash: available cash does not cover the notional plus the
	// over-estimated cost. StockOS's second code (INSUFFICIENT_CASH_AFTER_COSTS)
	// is merged into this one — see issues.md I4.
	ReasonInsufficientCash ReasonCode = "INSUFFICIENT_CASH"

	// --- 8. same-day re-entry ----------------------------------------------------

	// ReasonPendingBuyOrderBlocked: an unfilled buy on this symbol already exists.
	ReasonPendingBuyOrderBlocked ReasonCode = "PENDING_BUY_ORDER_BLOCKED"
	// ReasonSameDayReentryBlocked: the symbol reached its same-day entry limit
	// (default 2 `[미검증]`).
	ReasonSameDayReentryBlocked ReasonCode = "SAME_DAY_REENTRY_BLOCKED"
	// ReasonSameDayReentryCooldownActive: the cooldown since the last entry on
	// this symbol has not elapsed (default 30분 `[미검증]`).
	ReasonSameDayReentryCooldownActive ReasonCode = "SAME_DAY_REENTRY_COOLDOWN_ACTIVE"

	// --- 9. aggregates -------------------------------------------------------------

	// ReasonOpenExposureExceeded: total open exposure would reach the ceiling.
	// The aggregate boundary is ≥ — reaching it blocks (예약 실장과 일치).
	ReasonOpenExposureExceeded ReasonCode = "OPEN_EXPOSURE_EXCEEDED"
	// ReasonDailyLossLimitReached: the day's realised loss reached the absolute
	// ceiling or the capital-ratio ceiling, or the account's equity is not
	// positive (계좌자본 0 이하이면 즉시 차단).
	ReasonDailyLossLimitReached ReasonCode = "DAILY_LOSS_LIMIT_REACHED"

	// --- 10. duplicate ---------------------------------------------------------------

	// ReasonDuplicateOrder: this intent duplicates one already in flight.
	ReasonDuplicateOrder ReasonCode = "DUPLICATE_ORDER"
)

// allReasonCodes is the whole vocabulary, in chain order. It exists so a test
// can assert that the set has not silently grown or been respelled.
var allReasonCodes = []ReasonCode{
	ReasonAllowed,
	ReasonInputUnavailable,
	ReasonKillSwitchActive,
	ReasonOperatingModeBlocked,
	ReasonEntryGateBlocked,
	ReasonSymbolNotAllowed,
	ReasonStopMissing,
	ReasonStopNotBelowEntry,
	ReasonTargetNotAboveEntry,
	ReasonInvalidTargetStop,
	ReasonTargetBelowBreakEven,
	ReasonInvalidOrderSize,
	ReasonMaxOrderExceeded,
	ReasonSellExceedsHoldings,
	ReasonMinRRNotMet,
	ReasonInsufficientCash,
	ReasonPendingBuyOrderBlocked,
	ReasonSameDayReentryBlocked,
	ReasonSameDayReentryCooldownActive,
	ReasonOpenExposureExceeded,
	ReasonDailyLossLimitReached,
	ReasonDuplicateOrder,
}

// AllReasonCodes returns every reason code this package can produce.
func AllReasonCodes() []ReasonCode {
	out := make([]ReasonCode, len(allReasonCodes))
	copy(out, allReasonCodes)
	return out
}
