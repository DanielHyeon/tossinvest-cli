# Function Logic Map: `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero`

- Source: `internal/performance/store_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | test fixture and assertions | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/performance/store_test.go:82`: `if _, err := store.Collect(context.Background(), trade, []Observation{{` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B2 | AST `if` at `internal/performance/store_test.go:89`: `if err := store.db.QueryRow(\`SELECT cost_total FROM performance_trades WHERE id=?\`, trade.ID).Scan(&cost); err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B3 | AST `if` at `internal/performance/store_test.go:92`: `if cost.Valid {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B4 | AST `if` at `internal/performance/store_test.go:97`: `if err := store.db.QueryRow(\`SELECT status,gross_value,cost_adjusted_value` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B5 | AST `if` at `internal/performance/store_test.go:101`: `if status != string(StatusNotMeasured) \|\| gross.String != "5" \|\| adjusted.Valid {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B6 | AST `if` at `internal/performance/store_test.go:107`: `if err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B7 | AST `if` at `internal/performance/store_test.go:110`: `if view.States.NotMeasured == 0 \|\| len(view.Aggregates) != 1 {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B8 | AST `range` at `internal/performance/store_test.go:113`: `for _, metric := range view.Aggregates[0].Metrics {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| B9 | AST `if` at `internal/performance/store_test.go:114`: `if metric.Key == "markout_5" && (metric.Status != StatusNotMeasured \|\| metric.Value != "") {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestStore` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| `time.Date` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| `measuredTrade` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| `store.Collect` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| `context.Background` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| `at.Add` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| `t.Fatal` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |
| `Scan` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/store_test.go` function `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` and its documented derived/test state.
- High-risk impact: no runtime authority.
