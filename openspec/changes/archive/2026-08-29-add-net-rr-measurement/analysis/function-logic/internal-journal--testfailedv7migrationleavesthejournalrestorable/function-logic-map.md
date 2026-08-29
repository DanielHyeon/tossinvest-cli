# Function Logic Map: `TestFailedV7MigrationLeavesTheJournalRestorable`

- Source: `internal/journal/migration_v7_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `add-net-rr-measurement`

## Why this function is in scope

Asserts a step failing partway leaves the live journal at v6 and the backup restorable. `SchemaVersion` moves 7 → 8, so `openTestJournalAt` (which migrates to head) would have turned this file's tests into tests of a v6→v8 transition. Each call was changed to `openJournalAtSchema(t, path, 7)`, which pins the step under test. This is the documented precedent: `migration_v6_test.go` did the same when v7 landed, and its comment says so — "a later change adding v7 must not silently turn it into a test of a different transition".

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| on-disk journal | a genuine v6 database built by `openJournalAtSchema(t, path, 6)` | the migration test harness | `t.Fatal` — the fixture is the test |
| `migrationOverride` | the plan under test, pinned to target 7 | `journal.Options` (unexported, test-only) | n/a |
| injected clock | `migrationTestInstant`, so the backup filename is deterministic | `clock.NewFake` | n/a |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`if` @ internal/journal/migration_v7_test.go:254) | assertion guard at internal/journal/migration_v7_test.go:254 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B2 (`if` @ internal/journal/migration_v7_test.go:269) | assertion guard at internal/journal/migration_v7_test.go:269 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B3 (`if` @ internal/journal/migration_v7_test.go:273) | assertion guard at internal/journal/migration_v7_test.go:273 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B4 (`if` @ internal/journal/migration_v7_test.go:276) | assertion guard at internal/journal/migration_v7_test.go:276 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B5 (`if` @ internal/journal/migration_v7_test.go:283) | assertion guard at internal/journal/migration_v7_test.go:283 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B6 (`if` @ internal/journal/migration_v7_test.go:287) | assertion guard at internal/journal/migration_v7_test.go:287 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B7 (`if` @ internal/journal/migration_v7_test.go:292) | assertion guard at internal/journal/migration_v7_test.go:292 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B8 (`if` @ internal/journal/migration_v7_test.go:296) | assertion guard at internal/journal/migration_v7_test.go:296 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B9 (`if` @ internal/journal/migration_v7_test.go:299) | assertion guard at internal/journal/migration_v7_test.go:299 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B10 (`if` @ internal/journal/migration_v7_test.go:302) | assertion guard at internal/journal/migration_v7_test.go:302 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B11 (`if` @ internal/journal/migration_v7_test.go:311) | assertion guard at internal/journal/migration_v7_test.go:311 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B12 (`if` @ internal/journal/migration_v7_test.go:315) | assertion guard at internal/journal/migration_v7_test.go:315 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B13 (`if` @ internal/journal/migration_v7_test.go:318) | assertion guard at internal/journal/migration_v7_test.go:318 (`if`) | none — the test only reads | `t.Fatal`/`t.Error` on violation | this test is itself the required test |

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
