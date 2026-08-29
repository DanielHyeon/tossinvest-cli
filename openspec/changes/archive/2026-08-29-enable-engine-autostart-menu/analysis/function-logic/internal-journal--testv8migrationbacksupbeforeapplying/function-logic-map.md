# Function Logic Map: `TestV8MigrationBacksUpBeforeApplying`

- Source: `internal/journal/migration_v8_test.go`
- Qualified function: `TestV8MigrationBacksUpBeforeApplying`
- AST evidence: `ast.json` (`47390b93f0a39f2a46256ea58f99f024192dce4c9953c39906e44aeded5ceb09`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/migration_v8_test.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/migration_v8_test.go:274` — if err := old.Close(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B2 | `if` at `internal/journal/migration_v8_test.go:281` — if len(backups) != 1 { | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B3 | `if` at `internal/journal/migration_v8_test.go:285` — if want := "journal.db.v7-pre-v8.20260330T003000Z.bak"; filepath.Base(backup) != want { | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B4 | `if` at `internal/journal/migration_v8_test.go:290` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B5 | `if` at `internal/journal/migration_v8_test.go:293` — if perm := info.Mode().Perm(); perm&0o077 != 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B6 | `if` at `internal/journal/migration_v8_test.go:298` — if err := j.db.QueryRowContext(ctx, | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B7 | `if` at `internal/journal/migration_v8_test.go:302` — if recorded != backup { | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B8 | `if` at `internal/journal/migration_v8_test.go:305` — if err := j.db.QueryRowContext(ctx, | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B9 | `if` at `internal/journal/migration_v8_test.go:309` — if recordedVersion != "7" { | only mutations visible in this branch and its callees | existing return/error contract | `TestV8MigrationBacksUpBeforeApplying`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.TempDir`, `filepath.Join`, `context.Background`, `openJournalAtSchema`, `seedV7Rows`, `countRows`, `old.Close`, `t.Fatal`, `backupsIn`, `len`, `t.Fatalf`, `filepath.Base` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 14 assignment(s) and 0 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
