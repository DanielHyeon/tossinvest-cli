# Function Logic Map: `Chase.NotMeasured`

- Source: `internal/candidate/veto.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| three veto states | measured or unmeasured | private D3 order | append every unmeasured code |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied order | local slice only | ordered result | zero/dropped-input tests |
| B2 | state unmeasured | append local code | continue | `TestTheZeroChaseDoesNotPassTheVeto` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `Chase.State`, `Measured` | ordered missing list | pure | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only a fresh result slice; unrecognised state remains unmeasured.

## Safety conclusion

- Safe edit boundary: immutable ordering access.
- High-risk impact: no; missing measurements remain visible.
