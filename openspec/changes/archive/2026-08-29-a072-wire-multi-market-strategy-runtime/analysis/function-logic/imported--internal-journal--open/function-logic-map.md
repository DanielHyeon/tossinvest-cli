# Function Logic Map: `Open`

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
| B1 | AST `if` at `internal/journal/journal.go:103`: `if path == "" {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B2 | AST `if` at `internal/journal/journal.go:105`: `if err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B3 | AST `if` at `internal/journal/journal.go:114`: `if prober == nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B4 | AST `if` at `internal/journal/journal.go:120`: `if err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B5 | AST `if` at `internal/journal/journal.go:124`: `if err := os.MkdirAll(dir, dataDirPerm); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B6 | AST `if` at `internal/journal/journal.go:129`: `if clk == nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B7 | AST `if` at `internal/journal/journal.go:133`: `if busy <= 0 {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B8 | AST `if` at `internal/journal/journal.go:138`: `if err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B9 | AST `if` at `internal/journal/journal.go:148`: `if err := db.PingContext(ctx); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B10 | AST `if` at `internal/journal/journal.go:158`: `if err := j.checkIntegrity(ctx); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B11 | AST `if` at `internal/journal/journal.go:163`: `if opts.migrationOverride != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B12 | AST `if` at `internal/journal/journal.go:166`: `if err := j.migrate(ctx, plan); err != nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B13 | AST `range` at `internal/journal/journal.go:172`: `for _, p := range []string{path, path + "-wal", path + "-shm"} {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| B14 | AST `if` at `internal/journal/journal.go:173`: `if _, statErr := os.Stat(p); statErr == nil {` | SQLite transaction or read state only; errors and missing evidence fail closed | condition determines the documented success/error/assertion path | `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `DefaultPath` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `filepath.Clean` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `filepath.Dir` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `SystemFSProber` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `CheckFilesystem` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `os.MkdirAll` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `fmt.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |
| `clock.System` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestMigrationV15TransactionBoundariesSurviveSIGKILL`, `TestFailedV15MigrationRollsBackColumnTriggerAndVersion`, migration package regressions |

## State mutations and fallbacks

- SQLite transaction or read state only; errors and missing evidence fail closed.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/journal.go` function `Open` and its documented derived/test state.
- High-risk impact: journal correctness is high-risk, but this function has no broker/order capability.
