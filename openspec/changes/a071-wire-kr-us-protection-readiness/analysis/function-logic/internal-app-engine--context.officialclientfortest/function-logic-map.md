# Function Logic Map: `Context.OfficialClientForTest`

- Source: `internal/app/engine/export_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test Context | initialized engine context | engine assembly | nil only when test built an invalid context |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | accessor called | none | returns sealed official client | existing official transport isolation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct test-only field access | no error/retry | CodeGraph + AST |

## State mutations and fallbacks

- No mutation; `_test.go` accessor is absent from production binaries.

## Safety conclusion

- Safe edit boundary: retain read-only test seam
- High-risk impact: no (test-only)
