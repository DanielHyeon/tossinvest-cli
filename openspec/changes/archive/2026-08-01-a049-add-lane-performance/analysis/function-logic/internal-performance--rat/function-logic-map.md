# Function Logic Map: `rat`

- Source: `internal/performance/model.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| persisted decimal text | canonical finite decimal, exponent-free | immutable performance DB | return explicit error; never return numeric zero for malformed bytes |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `decimal` succeeds | none | parsed rational | aggregate regression tests |
| B2 | empty/malformed/exponent form | none | `invalid persisted decimal` error | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decimal` | strict parser shared with validation | boolean failure is promoted to error | model tests |

## State mutations and fallbacks

- Pure parser; no mutation, polling, fallback, or trading capability.

## Safety conclusion

- Safe edit boundary: change silent-zero fallback into explicit error propagation.
- High-risk impact: medium reporting-integrity improvement; no live path.
