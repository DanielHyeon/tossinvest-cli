# Function Logic Map: `freezeTradeOutcomeTx`

- Source: `internal/journal/trade_outcomes.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` / `positionID` | active close transaction / exact projected position id | `ApplyTx`, `positions.id` | computation or insert failure returns `false`; caller's fill/close transaction is never failed by analytics |
| `model` | configured bounded cost model | `internal/costs` | unconfigured/invalid pricing produces no outcome row and no fabricated zero cost |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `computeTradeOutcome` returns `ok=false` | none | `false` | `TestAnUnconfiguredCostModelDoesNotFreezeAnOutcome` |
| B2 | exact outcome and cost are available | append one `trade_outcomes` row inside caller transaction | insert success boolean | `TestFutureTradeOutcomeFreezesCostTotalAndLegacyRowsRemainNull` |
| B3 | insert constraint/storage error | transaction statement fails; no partial outcome | `false` without delaying/rolling back broker fill path | `TestTradeOutcomeFailureCannotRollBackClosingFill` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `computeTradeOutcome` | derive immutable PnL, R and exact total cost once | no retry; false means not honestly measurable | current HEAD + `trade_outcomes_test.go` |
| `ApplyTx.Exec` | append the close outcome atomically with projection close | error is deliberately converted to false | AST + close-path regression tests |

## State mutations and fallbacks

- Only the nullable `cost_total` column is added to the existing insert; no historical row is rewritten.
- Analytics remains subordinate to close processing: inability to measure never blocks a fill or emergency reduction.

## Safety conclusion

- Safe edit boundary: additive nullable field on the existing frozen row; no order/gate/protection behavior changes.
- High-risk impact: yes — journal close transaction, pinned by failure-isolation and migration tests.
