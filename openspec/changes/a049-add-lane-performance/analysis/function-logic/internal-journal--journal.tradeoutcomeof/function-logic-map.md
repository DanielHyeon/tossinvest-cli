# Function Logic Map: `Journal.TradeOutcomeOf`

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
| B1 | AST `if` at `internal/journal/trade_outcomes.go:684`: `if err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |
| B2 | AST `if` at `internal/journal/trade_outcomes.go:688`: `if !rows.Next() {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.QueryContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |
| `strings.TrimSpace` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |
| `rows.Close` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |
| `rows.Next` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |
| `scanTradeOutcome` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` |

## State mutations and fallbacks

- SQLite transaction or read state only; errors and missing evidence fail closed.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/trade_outcomes.go` function `Journal.TradeOutcomeOf` and its documented derived/test state.
- High-risk impact: journal correctness is high-risk, but this function has no broker/order capability.
