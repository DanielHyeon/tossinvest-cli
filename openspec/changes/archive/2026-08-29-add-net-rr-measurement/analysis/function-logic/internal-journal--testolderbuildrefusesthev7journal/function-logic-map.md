# Function Logic Map: `TestOlderBuildRefusesTheV7Journal`

- Source: `internal/journal/migration_v7_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `add-net-rr-measurement`

## Why this function is in scope

Asserts a v6 binary refuses a v7 journal with `ErrSchemaTooNew` and takes no backup. `SchemaVersion` moves 7 → 8, so `openTestJournalAt` (which migrates to head) would have turned this file's tests into tests of a v6→v8 transition. Each call was changed to `openJournalAtSchema(t, path, 7)`, which pins the step under test. This is the documented precedent: `migration_v6_test.go` did the same when v7 landed, and its comment says so — "a later change adding v7 must not silently turn it into a test of a different transition".

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| on-disk journal | a genuine v6 database built by `openJournalAtSchema(t, path, 6)` | the migration test harness | `t.Fatal` — the fixture is the test |
| `migrationOverride` | the plan under test, pinned to target 7 | `journal.Options` (unexported, test-only) | n/a |
| injected clock | `migrationTestInstant`, so the backup filename is deterministic | `clock.NewFake` | n/a |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`if` @ internal/journal/migration_v7_test.go:159) | assertion guard at internal/journal/migration_v7_test.go:159 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B2 (`if` @ internal/journal/migration_v7_test.go:169) | assertion guard at internal/journal/migration_v7_test.go:169 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B3 (`if` @ internal/journal/migration_v7_test.go:172) | assertion guard at internal/journal/migration_v7_test.go:172 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B4 (`if` @ internal/journal/migration_v7_test.go:175) | assertion guard at internal/journal/migration_v7_test.go:175 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B5 (`if` @ internal/journal/migration_v7_test.go:182) | assertion guard at internal/journal/migration_v7_test.go:182 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B6 (`if` @ internal/journal/migration_v7_test.go:185) | assertion guard at internal/journal/migration_v7_test.go:185 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openJournalAtSchema` | builds and opens a journal at a pinned version | `t.Fatal` on error | migration_v5_test.go:46 |
| `countRows` / `sameCounts` | row preservation | `t.Fatal` | migration_v5_test.go |
| `backupsIn` / `assertBackupAtVersion` / `restoreBackup` | the backup contract | `t.Fatal` | migration_v5_test.go |

## State mutations and fallbacks

- Writes only inside `t.TempDir()`. No shared state, no production path.
- The edit changed **which version the journal is migrated to**, not a single assertion. Every `t.Errorf` message and every compared value is unchanged.

## Safety conclusion

- Safe edit boundary: the call that selects the migration target. Diff is `openTestJournalAt(t, path)` → `openJournalAtSchema(t, path, 7)` plus the matching expected-version literals.
- High-risk impact: **no**. This is test-only code; the production migration path is unchanged and v7's step SQL is untouched (`TestV8IsPurelyAdditive` proves the v7 schema objects are byte-identical in a v8 journal).
- Justification is required by task 7.3 and recorded there: without this edit these four tests would silently become v6→v8 tests, leaving the v6→v7 step untested.
