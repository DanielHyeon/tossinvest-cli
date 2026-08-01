# Function Logic Map: `TestAReadWhoseErrorWasDroppedIsNotAPass`

- Source: `internal/candidate/veto_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| dropped observation read | no usable observations | D10 three-state contract | verdict remains unmeasured and never passes |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | dropped read produced rows/error mismatch | testing failure | continue | this test |
| B2 | verdict passes | testing failure | continue | this test |
| B3 | inspect copied D3 order | test inspection | continue | this test |
| B4 | missing state has empty reason | testing failure | continue | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| measurement/evaluation functions and `OrderedVetoCodes` | protect dropped-error path | test-only | AST |

## State mutations and fallbacks

- Fixture only; dropped data never falls back to clear.

## Safety conclusion

- Safe edit boundary: expected-order accessor.
- High-risk impact: no; regression test.
