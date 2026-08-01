# Function Logic Map: `signalsVetoesFrom`

- Source: `internal/console/signals.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| chase states | raised, clear, unmeasured | candidate package and D3 order | each code renders exactly one state |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied D3 codes | local slice append | ordered cells | signals tests |
| B2 | classify state | local cell | one branch | signals tests |
| B3 | dangerous | set Danger | continue | raised render test |
| B4 | clear | set Clear | continue | clear render test |
| B5 | otherwise | attach reason | continue | unmeasured render test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `State`, state predicates | stable three-state cells | pure | CodeGraph + AST |

## State mutations and fallbacks

- Fresh display slice only. Unknown state renders a reason instead of appearing clear.

## Safety conclusion

- Safe edit boundary: immutable order accessor.
- High-risk impact: no; read-only UI projection.
