# Function Logic Map: `Journal.TradeOutcomes`

- Source: `internal/journal/trade_outcomes.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | context, journal/transaction state, persisted lineage and schema version | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/journal/trade_outcomes.go:704`: `if err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| B2 | AST `for` at `internal/journal/trade_outcomes.go:710`: `for rows.Next() {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| B3 | AST `if` at `internal/journal/trade_outcomes.go:712`: `if err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| B4 | AST `if` at `internal/journal/trade_outcomes.go:717`: `if err := rows.Err(); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.QueryContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| `strings.TrimSpace` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| `rows.Close` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| `rows.Next` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| `scanTradeOutcome` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| `append` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |
| `rows.Err` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeOutcomesAreScopedToTheAccount`, `TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost` |

## State mutations and fallbacks

- SQLite transaction or read state only; errors and missing evidence fail closed.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/trade_outcomes.go` function `Journal.TradeOutcomes` and its documented derived/test state.
- High-risk impact: journal correctness is high-risk, but this function has no broker/order capability.
