# Function Logic Map: `Chase.HasUnmeasured`

- Source: `internal/candidate/veto.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| three veto states | measured or unmeasured | `Chase.State` and D3 order | any unmeasured state returns true |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied order | none | inspect all codes | zero/dropped-input tests |
| B2 | state not measured | none | true immediately | `TestTheZeroChaseDoesNotPassTheVeto` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `Chase.State`, `VetoState.Measured` | preserve three-state semantics | pure | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Unknown code/state remains unmeasured.

## Safety conclusion

- Safe edit boundary: immutable ordering access only.
- High-risk impact: no; fail-closed result unchanged.
