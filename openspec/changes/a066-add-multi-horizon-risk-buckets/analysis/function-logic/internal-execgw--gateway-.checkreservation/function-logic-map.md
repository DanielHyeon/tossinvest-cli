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

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | decision is risk-reducing | none | allow without reservation | existing exit/cancel tests |
| B2 | legacy reservation read fails | none | Guardian reservation missing | existing reservation tests |
| B3/B4 | a HELD legacy reservation exists | none | continue | existing reservation tests |
| B5/B6 | marked q_final admission is missing/divergent or decision disappears during revalidation | none | stable risk-bucket or Guardian-missing refusal | q_final missing-admission and last-moment tests |
| B4 | legacy or exact marked q_final authority remains HELD | none | allow | q_final issuance and gateway success tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ReservationsForDecision` | preserve legacy aggregate hold gate | fail closed | current AST |
| `RevalidateQFinalAdmission` | detect the durable q_final marker and verify exact q_final/owner/aggregate/all-bucket authority | fail closed; no repair; unmarked legacy decisions return `(false, nil)` | current AST and journal contract |

## State mutations and fallbacks

- Read-only. Never releases or repairs reservations/owners.
- Risk-reducing decisions bypass both legacy and monetary admission checks.

## Safety conclusion

- Safe edit boundary: after proving legacy HELD, require exact q_final admission when the durable policy marker requires it.
- High-risk impact: yes — final exposure gate; every read/mismatch must refuse.
