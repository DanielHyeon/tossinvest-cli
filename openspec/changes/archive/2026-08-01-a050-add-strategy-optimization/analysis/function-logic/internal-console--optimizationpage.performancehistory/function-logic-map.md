# Function Logic Map: `optimizationPage.PerformanceHistory`

- Source: `internal/console/optimization_view.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.Selected` | one fixed optimization category | parsed server request | pure false for every other category |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| happy | compare selected category with `performance-history` | none | boolean template guard | category navigation render tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| equality comparison | render only the selected category body | pure/no retry | AST |

## State mutations and fallbacks

- No mutation, IO, fallback, or operating-authority binding.

## Safety conclusion

- Safe edit boundary: presentation-only category predicate.
- High-risk impact: no.
