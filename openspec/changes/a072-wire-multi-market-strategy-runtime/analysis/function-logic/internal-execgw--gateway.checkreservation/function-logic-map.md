# Function Logic Map: `Gateway.checkReservation`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context and durable decision | exposure-raising decisions require one exact HELD aggregate plus current q_final authority | journal, never caller | typed reservation/risk-bucket refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | decision is not exposure-raising | none | allow without reservation | exit/reduction tests |
| B2 | reservation read fails | none | reservation-missing refusal | failclosed tests |
| B3-B5 | a HELD aggregate exists and q_final revalidation succeeds/fails | read-only journal checks | allow or typed missing/risk-bucket refusal | reservation/q_final tests |
| B6 | zero HELD aggregate reservations | none | reservation-missing refusal | reservation gate tests |
| future boundary | exact frozen FX identity must also be proven for q_final strategy entries | no mutation | mismatch refusal before lease `SUBMITTING` | paired KR/US Gateway FX matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.ReservationsForDecision` | find real aggregate holds | read failure refuses | current source/tests |
| `Journal.RevalidateQFinalAdmission` | live-join q_final owner, aggregate and five monetary holds | missing/corrupt authority refuses | current source/tests |

## State mutations and fallbacks

- Read-only. It neither releases nor transfers reservations.
- FX-envelope validation belongs beside this check at the final Gateway callback because it also
  needs the opaque request evidence; the decision alone must not reconstruct authority from JSON.

## Safety conclusion

- Safe edit boundary: preserve existing aggregate/q_final read semantics; add no caller-supplied
  reservation or FX scalar authority.
- High-risk impact: **yes** — last-moment account headroom proof.
