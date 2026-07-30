# Function Logic Map: `TestFailedV8MigrationLeavesTheJournalRestorable`

- Source: `internal/journal/migration_v8_test.go`
- Qualified function: `TestFailedV8MigrationLeavesTheJournalRestorable`
- AST evidence: `ast.json` (`47390b93f0a39f2a46256ea58f99f024192dce4c9953c39906e44aeded5ceb09`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/migration_v8_test.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/migration_v8_test.go:327` — if err := old.Close(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B2 | `if` at `internal/journal/migration_v8_test.go:342` — if err == nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B3 | `if` at `internal/journal/migration_v8_test.go:346` — if len(backups) != 1 { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B4 | `if` at `internal/journal/migration_v8_test.go:349` — if !strings.Contains(err.Error(), backups[0]) { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B5 | `if` at `internal/journal/migration_v8_test.go:356` — if got := countRows(t, survivor.db, v7AllTables); !sameCounts(got, before) { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B6 | `if` at `internal/journal/migration_v8_test.go:360` — if err := survivor.db.QueryRowContext(ctx, | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B7 | `if` at `internal/journal/migration_v8_test.go:365` — if halfBuilt != 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B8 | `if` at `internal/journal/migration_v8_test.go:369` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B9 | `if` at `internal/journal/migration_v8_test.go:372` — if version != 7 { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B10 | `if` at `internal/journal/migration_v8_test.go:375` — if err := survivor.Close(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B11 | `if` at `internal/journal/migration_v8_test.go:384` — if got := countRows(t, restored.db, v7AllTables); !sameCounts(got, before) { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B12 | `if` at `internal/journal/migration_v8_test.go:388` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B13 | `if` at `internal/journal/migration_v8_test.go:391` — if version != 8 { | only mutations visible in this branch and its callees | existing return/error contract | `TestFailedV8MigrationLeavesTheJournalRestorable`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.TempDir`, `filepath.Join`, `context.Background`, `openJournalAtSchema`, `seedV7Rows`, `countRows`, `old.Close`, `t.Fatal`, `append`, `migrationsThrough`, `Open`, `clock.NewFake` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 17 assignment(s) and 0 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
