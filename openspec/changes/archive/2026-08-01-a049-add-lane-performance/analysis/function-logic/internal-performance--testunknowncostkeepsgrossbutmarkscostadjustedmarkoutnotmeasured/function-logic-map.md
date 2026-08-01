# Function Logic Map: `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured`

- Source: `internal/performance/model_test.go`
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
| B1 | AST `if` at `internal/performance/model_test.go:152`: `if err := trade.validate(); err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| B2 | AST `if` at `internal/performance/model_test.go:160`: `if unknown.Status != StatusNotMeasured \|\| unknown.GrossPct != "5" \|\| unknown.CostAdjustedPct != "" \|\|` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| B3 | AST `if` at `internal/performance/model_test.go:167`: `if zero.Status != StatusComplete \|\| zero.GrossPct != "5" \|\| zero.CostAdjustedPct != "5" {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| B4 | AST `range` at `internal/performance/model_test.go:171`: `for _, invalid := range []string{"broken", "-0.01"} {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| B5 | AST `if` at `internal/performance/model_test.go:173`: `if err := trade.validate(); err == nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| `measuredTrade` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| `trade.validate` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| `t.Fatalf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| `at.Add` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| `Markout` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| `Measure` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |
| `t.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/model_test.go` function `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` and its documented derived/test state.
- High-risk impact: no runtime authority.
