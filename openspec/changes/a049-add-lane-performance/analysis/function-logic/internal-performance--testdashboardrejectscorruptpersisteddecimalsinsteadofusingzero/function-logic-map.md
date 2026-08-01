# Function Logic Map: `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero`

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
| B1 | AST `range` at `internal/performance/store_test.go:121`: `for _, test := range []struct {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| B2 | AST `if` at `internal/performance/store_test.go:126`: `if _, err := store.db.Exec(\`UPDATE performance_trades SET realized_pnl_after_costs='broken'\`); err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| B3 | AST `if` at `internal/performance/store_test.go:131`: `if _, err := store.db.Exec(\`UPDATE performance_trades SET realized_r='broken'\`); err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| B4 | AST `if` at `internal/performance/store_test.go:136`: `if _, err := store.db.Exec(\`UPDATE metric_observations SET value='broken'` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| B5 | AST `if` at `internal/performance/store_test.go:147`: `if _, err := store.Collect(context.Background(), trade, []Observation{{` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| B6 | AST `if` at `internal/performance/store_test.go:154`: `if _, err := store.Dashboard(context.Background(), DefaultQuery(now)); err == nil \|\|` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `store.db.Exec` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| `t.Fatal` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| `t.Run` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| `openTestStore` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| `time.Date` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| `measuredTrade` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| `now.Add` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |
| `store.Collect` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/store_test.go` function `TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero` and its documented derived/test state.
- High-risk impact: no runtime authority.
