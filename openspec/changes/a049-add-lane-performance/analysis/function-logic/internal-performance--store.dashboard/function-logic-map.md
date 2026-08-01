# Function Logic Map: `Store.Dashboard`

- Source: `internal/performance/query.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| query | explicit AsOf, 1..90 days, fixed market/lane/complete filters, positive sample minimum | typed `Query` | invalid period fails before SQL |
| result size | <=10,000 trades (+1 sentinel) | implementation bound | larger result fails closed instead of truncating |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at line 112: `if err != nil {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B2 | AST `if` at line 118: `if err != nil {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B3 | AST `if` at line 122: `if err != nil {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B4 | AST `range` at line 125: `for _, aggregate := range view.Aggregates {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B5 | AST `if` at line 126: `if aggregate.Status == StatusInsufficientSample {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B6 | AST `if` at line 131: `if err := s.db.QueryRowContext(ctx, \`SELECT` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B7 | AST `range` at line 137: `for _, trade := range trades {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B8 | AST `range` at line 138: `for _, key := range []string{"markout_5", "markout_15", "markout_30", "slippage", "mfe", "mae"} {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |
| B9 | AST `if` at line 139: `if trade.Statuses[key] != StatusComplete {` | read-only query and local view counters only | explicit success/error/continue path; no invented fallback | `TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates`, `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`, `TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `queryTrades` | bounded indexed read of filtered trades/latest snapshots | errors returned; no fallback scan | current HEAD + AST |
| `aggregateTrades` | derives ordered portfolio/lane summaries | operates on bounded slice | tests |
| filtered state SQL | counts complete/link-missing under identical predicates | one indexed query | tests |

## State mutations and fallbacks

- Read-only; never loads price observations or performs a quote poll.
- No silent truncation, no unbounded period, and no policy recommendation/apply action.

## Safety conclusion

- Safe edit boundary: bounded derived analytics query.
- High-risk impact: read correctness/performance only.
