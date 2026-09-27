# Function Logic Map: `Gateway.checkReservation`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| decision | durable journal row; exposure-raising or risk-reducing | `LookupDecision` in submit | read failure/refusal stops dispatch |
| legacy reservations | at least one HELD for exposure raising | journal reservation ledger | none/read failure refuses |
| q_final admission | required only when the durable `RiskIntent.PolicyVersion` carries the q_final marker | immutable a066 final-decision/owner/reservation rows | missing/divergent/owner-released/bucket-not-HELD refuses |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.5 post-edit) |
|---|---|---|---|---|
| B1 | if at 890:2 | `if dec.SafetyClass != journal.SafetyClassExposureRaising {`; then `return nil` (line last changed by `a6a396ab`) | a066: risk-reducing decision returns before any reservation read | covered |
| B2 | if at 894:2 | `if err != nil {`; then `return reject(ReasonGuardianReservationMissing,` (line last changed by `a6a396ab`) | not a066 | NOT covered |
| B3 | range at 899:2 | `for _, reservation := range reservations {`; then `if reservation.Held() {` (line last changed by `a6a396ab`) | not a066 | covered |
| B4 | if at 900:3 | `if reservation.Held() {`; then `_, err := g.journal.RevalidateQFinalAdmission(ctx, dec.ID)` (line last changed by `a6a396ab`) | a066: HELD aggregate reservation triggers q_final revalidation | covered |
| B5 | if at 902:4 | `if err != nil {`; then `if errors.Is(err, journal.ErrDecisionNotFound) {` (line last changed by `a37d97f5`) | a066: q_final revalidation error refuses | covered |
| B6 | if at 903:5 | `if errors.Is(err, journal.ErrDecisionNotFound) {`; then `return reject(ReasonGuardianMissing,` (line last changed by `a37d97f5`) | a066: decision disappeared during revalidation keeps the Guardian-missing reason | covered |
| B7 | if at 908:5 | `if errors.Is(err, journal.ErrRiskBucketEntryLossLocked) {`; then `return reject(ReasonEntryLossLockActive,` (line last changed by `00000000`) | a066: entry loss lock refusal is reported as its own reason (entry_loss_lock_active), chosen by errors.Is on the journal sentinel type, not by text | covered |

5.5 post-edit (2026-09-27, HEAD `eecc5fe7` + working tree): AST 889–923, 6 → 7 branches. B1–B6 unchanged; new B7 maps the journal sentinel `ErrRiskBucketEntryLossLocked` (by `errors.Is`) to `entry_loss_lock_active` after the Guardian-missing check and before the mismatch fallback.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ReservationsForDecision` | preserve legacy aggregate hold gate | fail closed | current AST |
| `RevalidateQFinalAdmission` | detect the durable q_final marker and verify exact q_final/owner/aggregate/all-bucket authority | fail closed; no repair; unmarked legacy decisions return `(false, nil)` | current AST and journal contract |
| `errors.Is(err, journal.ErrRiskBucketEntryLossLocked)` | a066 5.5: name the entry-loss-lock refusal instead of reporting it as a bucket mismatch | type, never message text; checked after `ErrDecisionNotFound` | AST B7 + mutants M30/M31 |

## State mutations and fallbacks

- Read-only. Never releases or repairs reservations/owners.
- Risk-reducing decisions bypass both legacy and monetary admission checks.

## Safety conclusion

- Safe edit boundary: after proving legacy HELD, require exact q_final admission when the durable policy marker requires it.
- High-risk impact: yes — final exposure gate; every read/mismatch must refuse.
