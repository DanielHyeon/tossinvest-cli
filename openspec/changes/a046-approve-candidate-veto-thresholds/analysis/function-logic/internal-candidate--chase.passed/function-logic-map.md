# Function Logic Map: `Chase.Passed`

- Source: `internal/candidate/veto.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| three veto states | raised, clear, unmeasured | `Chase.State` and private D3 order | any non-clear returns false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied D3 order | none | continues through all codes | `TestRemovingAVetoCodeCannotRemoveItsVeto` |
| B2 | state is not clear | none | false immediately | zero/raised/unmeasured veto tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `Chase.State`, `VetoState.Clear` | fixed-order all-clear predicate | pure, no I/O/retry | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Unknown or missing state is unmeasured and therefore cannot pass.

## Safety conclusion

- Safe edit boundary: ordering source only changed to a private-copy accessor.
- High-risk impact: no; fail-closed semantics are unchanged and better protected.
