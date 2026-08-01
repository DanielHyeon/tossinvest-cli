# Function Logic Map: `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric`

- Source: `internal/console/performance_history_test.go`
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
| B1 | AST `range` at `internal/console/performance_history_test.go:41`: `for _, want := range []string{` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| B2 | AST `if` at `internal/console/performance_history_test.go:48`: `if !strings.Contains(page, want) {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| B3 | AST `if` at `internal/console/performance_history_test.go:52`: `if reader.calls != 1 \|\| reader.query.PeriodDays != 30 \|\| reader.query.Market != performance.AllMarkets \|\|` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| B4 | AST `range` at `internal/console/performance_history_test.go:56`: `for _, forbidden := range []string{` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| B5 | AST `if` at `internal/console/performance_history_test.go:61`: `if strings.Contains(strings.ToLower(page), forbidden) {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| `h.authenticate` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| `h.page` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| `strings.Contains` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| `t.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| `t.Fatalf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |
| `strings.ToLower` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/console/performance_history_test.go` function `TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric` and its documented derived/test state.
- High-risk impact: no runtime authority.
