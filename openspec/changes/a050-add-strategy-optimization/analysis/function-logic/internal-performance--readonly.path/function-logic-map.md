# Function Logic Map: `ReadOnly.Path`

- Source: `internal/performance/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receiver | nil or initialized | `OpenReadOnly` | nil returns empty path; initialized returns cleaned path |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil receiver | none | empty string | path accessor coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure accessor | no error/retry | AST |

## State mutations and fallbacks

- No mutation or I/O; empty string is only the nil-receiver sentinel.

## Safety conclusion

- Safe edit boundary: read-only diagnostic accessor.
- High-risk impact: no.
