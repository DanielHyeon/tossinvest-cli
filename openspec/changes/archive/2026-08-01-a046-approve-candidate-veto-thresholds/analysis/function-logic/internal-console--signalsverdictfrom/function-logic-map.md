# Function Logic Map: `signalsVerdictFrom`

- Source: `internal/console/signals.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| candidate chase verdict | three D3 states | candidate package | vetoed wins, then passed, otherwise unchecked |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | verdict classification switch | local render value | one case returned | signals tests |
| B2 | any veto raised | none | vetoed row | raised render test |
| B3 | all measured and clear | none | passed row | clear render test |
| B4 | otherwise | none | unchecked row | nobody-measured test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `NotMeasured`, `Vetoed`, `Passed` | measured detail and precedence | pure | CodeGraph + AST |

## State mutations and fallbacks

- Local display value only; no activation or mutation. Unmeasured falls to unchecked, never passed.

## Safety conclusion

- Safe edit boundary: immutable code-count accessor.
- High-risk impact: no; render-only.
