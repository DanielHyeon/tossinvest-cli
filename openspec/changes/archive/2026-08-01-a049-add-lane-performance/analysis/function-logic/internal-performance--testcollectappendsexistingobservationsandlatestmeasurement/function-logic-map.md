# Function Logic Map: `TestCollectAppendsExistingObservationsAndLatestMeasurement`

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
| B1 | base-revision AST `if` at line 56: `if err != nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| B2 | base-revision AST `if` at line 59: `if got.Markout(30).GrossPct != "7" {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| B3 | base-revision AST `if` at line 63: `if err := store.db.QueryRow(\`SELECT count(*) FROM price_observations\`).Scan(&observations); err != nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| B4 | base-revision AST `if` at line 66: `if err := store.db.QueryRow(\`SELECT count(*) FROM measurement_snapshots\`).Scan(&snapshots); err != nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| B5 | base-revision AST `if` at line 69: `if observations != 3 \|\| snapshots != 1 {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| B6 | base-revision AST `if` at line 72: `if _, err := store.Collect(ctx, trade, rows, at.Add(2*time.Hour)); err != nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `context.Background` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| `openTestStore` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| `time.Date` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| `measuredTrade` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| `at.Add` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| `store.Collect` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| `t.Fatal` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |
| `got.Markout` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestCollectAppendsExistingObservationsAndLatestMeasurement` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/store_test.go` function `TestCollectAppendsExistingObservationsAndLatestMeasurement` and its documented derived/test state.
- High-risk impact: no runtime authority.
