# Function Logic Map: `Journal.applyMigration`

- Source: `internal/journal/journal.go`
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
| B1 | AST `if` at `internal/journal/journal.go:275`: `if err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B2 | AST `if` at `internal/journal/journal.go:280`: `if _, err := tx.ExecContext(ctx, m.SQL); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B3 | AST `if` at `internal/journal/journal.go:284`: `if _, err := tx.ExecContext(ctx,` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B4 | AST `if` at `internal/journal/journal.go:290`: `if _, err := tx.ExecContext(ctx,` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B5 | AST `if` at `internal/journal/journal.go:295`: `if _, err := tx.ExecContext(ctx,` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B6 | AST `if` at `internal/journal/journal.go:300`: `if j.migrationHook != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B7 | AST `if` at `internal/journal/journal.go:304`: `if _, err := tx.ExecContext(ctx, "PRAGMA user_version = "+strconv.Itoa(m.Version)); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B8 | AST `if` at `internal/journal/journal.go:307`: `if j.migrationHook != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B9 | AST `if` at `internal/journal/journal.go:310`: `if err := tx.Commit(); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.BeginTx` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `tx.Rollback` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `tx.ExecContext` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `Format` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `UTC` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `j.clk.Now` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `strconv.Itoa` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |

## State mutations and fallbacks

- SQLite transaction or read state only; errors and missing evidence fail closed.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/journal.go` function `Journal.applyMigration` and its documented derived/test state.
- High-risk impact: journal correctness is high-risk, but this function has no broker/order capability.
