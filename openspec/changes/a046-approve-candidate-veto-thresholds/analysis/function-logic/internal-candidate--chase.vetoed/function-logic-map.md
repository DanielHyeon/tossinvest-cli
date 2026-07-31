# Function Logic Map: `Chase.Vetoed`

- Source: `internal/candidate/veto.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| three veto states | raised, clear, unmeasured | `Chase.State` and private D3 order | only a dangerous state returns true |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied D3 order | none | inspect each state | `TestRemovingAVetoCodeCannotRemoveItsVeto` |
| B2 | dangerous state found | none | true immediately | raised-veto tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `Chase.State`, `VetoState.Dangerous` | fixed-order any-raised predicate | pure | CodeGraph + AST |

## State mutations and fallbacks

- No mutation or numeric fallback.

## Safety conclusion

- Safe edit boundary: replace exported ordering read with copy accessor.
- High-risk impact: no; raised detection is unchanged.
