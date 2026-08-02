# Function Logic Map: `positionRow.HasDetail`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Position row evidence | exit, reference, journal, or reason | local read model | false omits empty detail disclosure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless boolean expression | none | true when any detail source exists | positions reference tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `HasExitReference` | includes non-actionable reference-only rows | pure predicate | a053 console tests |
| `Reason` | includes row-specific management explanation | pure predicate | existing portfolio tests |

## State mutations and fallbacks

- No mutation or fallback; the method only controls whether read-only details render.

## Safety conclusion

- Safe edit boundary: presentation predicate only.
- High-risk impact: no; it creates no prices, actions, or capabilities.
