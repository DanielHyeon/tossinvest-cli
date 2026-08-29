# Function Logic Map: `TestOlderBuildRefusesTheV8Journal`

- Source: `internal/journal/migration_v8_test.go`
- Qualified function: `TestOlderBuildRefusesTheV8Journal`
- AST evidence: `ast.json` (`47390b93f0a39f2a46256ea58f99f024192dce4c9953c39906e44aeded5ceb09`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `internal/journal/migration_v8_test.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/journal/migration_v8_test.go:232` — if err := j.Close(); err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestOlderBuildRefusesTheV8Journal`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B2 | `if` at `internal/journal/migration_v8_test.go:242` — if err == nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestOlderBuildRefusesTheV8Journal`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B3 | `if` at `internal/journal/migration_v8_test.go:245` — if !errors.Is(err, ErrSchemaTooNew) { | only mutations visible in this branch and its callees | existing return/error contract | `TestOlderBuildRefusesTheV8Journal`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B4 | `if` at `internal/journal/migration_v8_test.go:248` — if !strings.Contains(err.Error(), "8") \|\| !strings.Contains(err.Error(), "7") { | only mutations visible in this branch and its callees | existing return/error contract | `TestOlderBuildRefusesTheV8Journal`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B5 | `if` at `internal/journal/migration_v8_test.go:255` — if got := countRows(t, reopened.db, []string{"positions"})["positions"]; got != 1 { | only mutations visible in this branch and its callees | existing return/error contract | `TestOlderBuildRefusesTheV8Journal`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |
| B6 | `if` at `internal/journal/migration_v8_test.go:258` — if backups := backupsIn(t, filepath.Dir(path)); len(backups) != 0 { | only mutations visible in this branch and its callees | existing return/error contract | `TestOlderBuildRefusesTheV8Journal`; `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.Join`, `t.TempDir`, `openJournalAtSchema`, `seedV7Rows`, `j.Close`, `t.Fatal`, `Open`, `context.Background`, `clock.NewFake`, `FixedFSProber`, `migrationsThrough`, `errors.Is` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 7 assignment(s) and 0 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: yes; current AST hash and affected-package tests are required.
