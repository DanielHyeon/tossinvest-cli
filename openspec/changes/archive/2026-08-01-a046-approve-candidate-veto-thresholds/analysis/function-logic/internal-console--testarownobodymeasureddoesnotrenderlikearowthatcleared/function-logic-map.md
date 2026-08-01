# Function Logic Map: `TestARowNobodyMeasuredDoesNotRenderLikeARowThatCleared`

- Source: `internal/console/signals_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| unmeasured and clear chase rows | two explicit fixtures | console three-state contract | test fails if labels/cells/cardinality collapse |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | unmeasured verdict label wrong | testing failure | continue | this test |
| B2 | unmeasured detail wrong | testing failure | continue | this test |
| B3 | cell count wrong | testing failure | continue | this test |
| B4 | inspect unmeasured cells | test inspection | continue | this test |
| B5 | cell lacks reason/has clear | testing failure | continue | this test |
| B6 | inspect clear cells | test inspection | continue | this test |
| B7 | clear cell wrong | testing failure | continue | this test |
| B8 | clear verdict label wrong | testing failure | continue | this test |
| B9 | clear marker count differs from copied order | testing failure | return | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| console render helpers and `OrderedVetoCodes` | compare unmeasured vs clear projection | test-only | AST |

## State mutations and fallbacks

- Test render strings only; no production state.

## Safety conclusion

- Safe edit boundary: expected cardinality uses immutable accessor.
- High-risk impact: no; UI regression test.
