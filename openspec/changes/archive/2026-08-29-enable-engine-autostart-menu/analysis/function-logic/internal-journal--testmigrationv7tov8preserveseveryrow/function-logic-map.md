# Function Logic Map: `TestMigrationV7ToV8PreservesEveryRow`

- Source: `internal/journal/migration_v8_test.go`
- Qualified function: `TestMigrationV7ToV8PreservesEveryRow`
- AST evidence: `ast.json` (`47390b93f0a39f2a46256ea58f99f024192dce4c9953c39906e44aeded5ceb09`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/migration_v8_test.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/migration_v8_test.go:74` — if err := old.db.QueryRowContext(ctx, | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B2 | `if` at `internal/journal/migration_v8_test.go:79` — if err := old.Close(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B3 | `if` at `internal/journal/migration_v8_test.go:85` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B4 | `if` at `internal/journal/migration_v8_test.go:88` — if version != 8 { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B5 | `range` at `internal/journal/migration_v8_test.go:93` — for _, table := range v7AllTables { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B6 | `if` at `internal/journal/migration_v8_test.go:94` — if before[table] != after[table] { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B7 | `if` at `internal/journal/migration_v8_test.go:97` — if after[table] == 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B8 | `if` at `internal/journal/migration_v8_test.go:106` — if err := j.db.QueryRowContext(ctx, | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B9 | `if` at `internal/journal/migration_v8_test.go:111` — if keptPreimage != preimage { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B10 | `if` at `internal/journal/migration_v8_test.go:114` — if keptHash != hash { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B11 | `range` at `internal/journal/migration_v8_test.go:118` — for _, table := range v8Tables { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B12 | `if` at `internal/journal/migration_v8_test.go:120` — if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B13 | `if` at `internal/journal/migration_v8_test.go:124` — if n != 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B14 | `if` at `internal/journal/migration_v8_test.go:128` — if err := j.checkIntegrity(ctx); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestMigrationV7ToV8PreservesEveryRow`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.Join`, `t.TempDir`, `context.Background`, `openJournalAtSchema`, `seedV7Rows`, `countRows`, `Scan`, `old.db.QueryRowContext`, `t.Fatal`, `old.Close`, `j.SchemaVersion`, `t.Fatalf` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 12 assignment(s) and 0 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
