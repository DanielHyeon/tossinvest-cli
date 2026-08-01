# Function Logic Map: `Trade.validate`

- Source: `internal/performance/model.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| identity/lineage anchor | non-empty trade and position id | journal outcome/position PK | return validation error |
| market/side/times | known market, BUY/SELL, ordered entry/close | exact journal adapter | return validation error |
| decimal metrics | positive entry price/quantity; signed PnL/R | journal frozen outcome | return validation error |
| `CostTotal` | empty means journal v14 legacy/unmeasured; otherwise non-negative decimal | nullable journal v15 field | negative/malformed non-empty value is rejected |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `switch` at line 223: `switch {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B2 | AST `case` at line 224: `case strings.TrimSpace(t.ID) == "":` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B3 | AST `case` at line 226: `case strings.TrimSpace(t.Lineage.PositionID) == "":` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B4 | AST `case` at line 228: `case strings.TrimSpace(t.Market) == "":` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B5 | AST `case` at line 230: `case t.Side != SideBuy && t.Side != SideSell:` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B6 | AST `case` at line 232: `case t.EntryAt.IsZero() \|\| t.ClosedAt.IsZero() \|\| t.ClosedAt.Before(t.EntryAt):` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B7 | AST `range` at line 235: `for name, value := range map[string]string{"pnl after costs": t.RealizedPnLAfterCosts, "realized r": t.RealizedR} {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B8 | AST `if` at line 236: `if _, ok := decimal(value); !ok {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B9 | AST `range` at line 240: `for name, value := range map[string]string{"entry price": t.EntryPrice, "quantity": t.Quantity} {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B10 | AST `if` at line 241: `if parsed, ok := decimal(value); !ok \|\| parsed.Sign() <= 0 {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B11 | AST `if` at line 245: `if strings.TrimSpace(t.CostTotal) != "" {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B12 | AST `if` at line 246: `if cost, ok := decimal(t.CostTotal); !ok \|\| cost.Sign() < 0 {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B13 | AST `if` at line 250: `if strings.TrimSpace(t.DecisionPrice) != "" {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |
| B14 | AST `if` at line 251: `if decision, ok := decimal(t.DecisionPrice); !ok \|\| decision.Sign() <= 0 {` | none; validation returns before persistence | explicit success/error/continue path; no invented fallback | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`, `TestUnknownCostKeepsGrossButMarksCostAdjustedMarkoutNotMeasured` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decimal` / `positiveDecimal` | strict non-exponent decimal parsing | no default and no rounding | model tests |

## State mutations and fallbacks

- Validation does not write or poll. Nullable cost affects only cost-adjusted metrics, whose empty value is rendered `not_measured`.

## Safety conclusion

- Safe edit boundary: widen only empty cost from invalid to explicitly unmeasured.
- High-risk impact: no trading authority; medium data-integrity risk pinned by nullable-cost tests.
