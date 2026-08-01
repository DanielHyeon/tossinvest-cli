# Function Logic Map: `parseStoredTime`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| stored RFC3339Nano string | canonical UTC, parseable, nonzero | persisted DB value | error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | parse/zero/noncanonical input | none | invalid timestamp | corruption tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Parse` | canonical timestamp decode | parse error propagates as helper error | AST |

## State mutations and fallbacks

- Pure validation helper.

## Safety conclusion

- Safe edit boundary: read integrity.
- High-risk impact: no.
