# Function Logic Map: `TestTheZeroChaseDoesNotPassTheVeto`

- Source: `internal/candidate/veto_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| zero `Chase` | all states unmeasured | D10 fail-closed contract | test fails if it passes, loses missing codes, or appears vetoed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | zero chase passes | testing error | continue | this test |
| B2 | zero chase lacks unmeasured | testing error | continue | this test |
| B3 | missing-code count differs from copied order | testing error | continue | this test |
| B4 | zero chase reports vetoed | testing error | return | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| chase predicates, `OrderedVetoCodes` | assert zero semantics and fixed cardinality | test-only | AST |

## State mutations and fallbacks

- Test state only; accessor replacement does not change assertions.

## Safety conclusion

- Safe edit boundary: derive expected count from immutable copy.
- High-risk impact: no; regression test.
