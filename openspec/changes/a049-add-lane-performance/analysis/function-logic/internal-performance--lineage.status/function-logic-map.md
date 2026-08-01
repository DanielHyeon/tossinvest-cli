# Function Logic Map: `Lineage.Status`

- Source: `internal/performance/model.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | validated journal-derived trades, observations, query/window, and derived-store state | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `range` at `internal/performance/model.go:62`: `for _, value := range []string{` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance`, `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B2 | AST `if` at `internal/performance/model.go:67`: `if strings.TrimSpace(value) == "" {` | local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured | condition determines the documented success/error/assertion path | `TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance`, `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMeasureSideAdjustsBuyAndSellMarkoutsWithCostsAndProvenance`, `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |

## State mutations and fallbacks

- local analytics values or derived SQLite state only; corrupt/unknown values remain errors or not-measured.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/model.go` function `Lineage.Status` and its documented derived/test state.
- High-risk impact: analytics only; no order, toggle, broker, or LIVE capability.
