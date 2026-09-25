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
| B1 | decision is not exposure-raising | none | nil — risk-reducing bypass | `TestAnExitNeedsNoReservation` |
| B2 | legacy reservation read fails | none | Guardian reservation missing | no test executes this body (Wave 2A coverage) |
| B3 | iterate the decision's aggregate reservations | none | continue | package suite |
| B4 | a HELD aggregate reservation triggers q_final revalidation | none | continue into `RevalidateQFinalAdmission` | `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker` |
| B5 | revalidation returns an error | none | refusal | `TestRevokedDecisionIsRefusedAtTheLastMoment` |
| B6 | the error is `ErrDecisionNotFound` | none | Guardian-missing reason kept | `TestRevokedDecisionIsRefusedAtTheLastMoment` |

Wave 2A (2026-09-25): AST re-extracted at HEAD (889–918, was 745–774); body text identical to the 2026-08-04
revision, alignment identical B1–B6. Rows re-described from the AST source lines; the earlier table grouped
B3/B4 and B5/B6 and listed B4 twice.

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
