# Function Logic Map: `ReadinessAdapter.Check`

- Source: `internal/protection/readiness_adapter.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request | KR/US, exact entry order type and canonical positive integral quantity | execgw mutation plan | typed refusal before transport |
| supervisor contract | sealed per-market session/trigger/replace/capability binding | engine production supervisor | corrupt/missing contract fails closed |
| prior checkpoint | empty on first check or exact same market generation/identity | first readiness decision | drift returns typed refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | adapter/provider/seal invalid | none | state-corrupt refusal | adapter corruption test |
| B2 | unsupported market or invalid order/quantity | none | invalid refusal | substitution matrix |
| B3 | provider error | none | provider-unavailable refusal | existing provider failure test |
| B4 | dispatch refuses exact merged scope | none | propagated typed refusal | dispatch substitution matrix |
| B5 | prior checkpoint differs | none | snapshot-drift refusal | drift test |
| B6 | exact current snapshot | none | sealed checkpoint | valid test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SnapshotProvider.Current` | obtain current immutable KR/US view | single read; error is fail-closed | current HEAD |
| `ReadinessSnapshot.Dispatch` | compare plan plus sealed supervisor contract | pure/no retry | current HEAD |

## State mutations and fallbacks

- No broker mutation, retry or toggle write. The adapter only narrows authority.

## Safety conclusion

- Safe edit boundary: merge plan-supplied fields with already sealed supervisor-only fields; never invent defaults.
- High-risk impact: yes — called twice around durable dispatch.
