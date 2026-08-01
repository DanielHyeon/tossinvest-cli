# Function Logic Map: `ReadOnly.Close`

- Source: `internal/performance/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receiver | nil or initialized reader | lifecycle owner | nil is idempotent success; initialized reader becomes closed under lock |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil receiver | none | nil | nil/idempotent close coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mutex lock/unlock | serializes close against reads | blocking only for in-flight local read | race and lifecycle tests |

## State mutations and fallbacks

- Sets only `closed=true`; later reads fail. No DB writer, journal, broker, or trading state is touched.

## Safety conclusion

- Safe edit boundary: in-process read capability lifecycle.
- High-risk impact: no.
