// Package exitpolicy is the local protection decision: given what a position was
// opened at and what its price has been observed to do, what the protected
// baseline now is and whether anything should be proposed.
//
// It is pure. It reads no clock, opens no connection, imports no journal and
// submits no order — every input arrives as a decimal string and every output is
// a value. That is deliberate and load-bearing: the same evaluation has to be
// replayable from a persisted row after a crash, and a decision function that
// consulted the outside world could not be replayed.
//
// # The two policies
//
// A position has exactly one (exit-policy: 포지션당 정책 하나; design D5). They
// are alternative execution paths in the original and must never contest one
// baseline.
//
//   - RATCHET — [EvaluateRatchet]. A baseline that steps up on R triggers, with
//     one 40 % partial take-profit. Ported from StockOS
//     `packages/trading/stockos_trading/exit/baseline_ratchet.py` and
//     `exit/protected_stop_candidate.py`.
//   - LADDER — [EvaluateLadder]. A rung table of percentage targets, each rung
//     raising a protected stop and optionally taking part of the remainder.
//     Ported from StockOS `packages/trading/stockos_trading/profit_ladder.py`.
//
// # The probe is a watermark, not the last price
//
// Both policies judge *attainment* against `high_water` — the running maximum of
// the observed prices — and judge *breach* against the price just observed. The
// original's `high_since_entry` is the same idea (baseline_ratchet.py:87
// `probe_price = inputs.high_since_entry or inputs.current_price`), restored
// here as a required field rather than an optional one, which is what makes the
// level a monotone function of the observations.
//
// The limitation is stated rather than hidden (exit-policy SHALL): the watermark
// is the maximum of the *samples*. A trigger brushed between two observations is
// missed, and the observation period is what defines that tolerance. The upside
// of that is recovered at the next sample because the watermark never falls; the
// downside — a baseline breached and recovered between two samples — is not, and
// is named in design.md's risk list as the window 2c's broker-resident stop
// closes.
//
// # Numbers
//
// Every price, ratio and R multiple is a decimal string, and every computation
// is a `math/big` rational over those strings, for the reason internal/risk gives
// (contract.go): the rules here have boundaries at exact values — ">= 0.4R",
// "> the previous baseline" — and a boundary decided by a float's binary
// expansion is not the boundary anybody specified.
//
// Where a result cannot terminate in decimal (an R multiple, a ratio
// reconstructed from quantities) it is rendered at a fixed scale, and the
// rounding direction is chosen so the error cannot weaken protection: a
// protective price rounds **up**, a taken ratio rounds **up** (a larger recorded
// take suppresses more future proposals). See decimal.go.
//
// # Provenance of the numbers
//
// The ratchet trigger table (+0.4R → −0.5R, +0.8R → real break-even, +1.0R →
// 40 % partial, +1.2R → +0.3R, +2.0R → +0.8R) is baseline_ratchet.py:32-41.
// The ladder rung defaults (1.5/2.5/4.0/6.0 % targets, 0/1.0/2.0/3.5 % locks,
// 0/0.25/0.25/1.0 partials) are profit_ladder.py:165-176. Both are
// `[미검증 — StockOS KOSPI 튜닝값]`: they were tuned on a different broker and a
// different market and TossOS has measured neither. They are defaults, they are
// overridable by configuration, and §0.9 permits changing them only in the
// conservative direction until a tracer measures them.
//
// # What is deliberately not here
//
//   - The observation loop, the price poll and the submission of a proposal to
//     Guardian (task 7.4). This package decides; it does not act.
//   - The confirmed-floor cap on a liquidation and the alerts (task 7.5).
//   - Persistence. The journal owns `exit_states` and `exit_events`; the values
//     in and out of this package are the columns of those tables and nothing
//     here knows they exist.
//   - Every signal-derived candidate the original composes a stop from — VWAP,
//     EMA9, ATR trail, distribution lock, the fixed and R-multiple floors. They
//     are the signal layer, which this change admits zero inputs from (P3).
//   - The original's STOP_FIRST same-bar fill model, which needs OHLC input that
//     only a backtest has (exit-policy: 이 change의 SHALL이 아니다).
package exitpolicy
