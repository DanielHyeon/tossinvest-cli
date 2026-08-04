# Function Logic Map: `NewReadinessAdapter`

- Source: `internal/protection/readiness_adapter.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provider/account/profile | non-nil provider and non-empty exact identity | engine assembly | constructor error |
| paired supervisor contracts | exact KR and US sealed contract, no defaulting | production supervisor assembly | constructor error; default adapter remains paired UNWIRED |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | provider/identity absent | none | constructor error | adapter constructor test |
| B2 | production contract corrupt/incomplete | none | constructor error | contract seal test |
| B3 | valid dependencies | immutable adapter seal | adapter | valid adapter test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| readiness adapter seal | bind account/profile and paired contract identities | pure/no retry | current HEAD |

## State mutations and fallbacks

- Constructor has no broker mutation and cannot turn a default snapshot WIRED.

## Safety conclusion

- Safe edit boundary: require explicit sealed contracts for production constructor; keep a separate default-UNWIRED constructor.
- High-risk impact: yes — an adapter is consumed at the exposure boundary.
