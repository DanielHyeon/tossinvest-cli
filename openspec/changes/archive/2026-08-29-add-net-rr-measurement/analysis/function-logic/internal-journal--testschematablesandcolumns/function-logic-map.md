# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `add-net-rr-measurement`

## Why this function is in scope

Pins the full table list and the column list of the tables Phase 2's ledger import reads. This change added one entry — `entry_decision_observations` — in sorted position.

The test exists to make a schema addition a deliberate act rather than a silent one, so editing it *is* the intended workflow for an additive change. What must not happen is an edit that removes or reorders an existing entry, and the diff is a single inserted line.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `wantTables` | every table a head-version journal holds, sorted | this literal — it is the contract | `t.Fatalf` listing both sides |
| `wantColumns` | per-table column lists for the ledger-import tables | this literal | `t.Errorf` per table |
| the opened journal | migrated to head by `openTestJournal` | `journal.Open` | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`if` @ internal/journal/schema_test.go:112) | assertion guard at internal/journal/schema_test.go:112 (`if`) | none — the test only reads `sqlite_master` and `pragma_table_info` | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B2 (`for` @ internal/journal/schema_test.go:116) | assertion guard at internal/journal/schema_test.go:116 (`for`) | none — the test only reads `sqlite_master` and `pragma_table_info` | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B3 (`if` @ internal/journal/schema_test.go:118) | assertion guard at internal/journal/schema_test.go:118 (`if`) | none — the test only reads `sqlite_master` and `pragma_table_info` | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B4 (`if` @ internal/journal/schema_test.go:123) | assertion guard at internal/journal/schema_test.go:123 (`if`) | none — the test only reads `sqlite_master` and `pragma_table_info` | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B5 (`if` @ internal/journal/schema_test.go:127) | assertion guard at internal/journal/schema_test.go:127 (`if`) | none — the test only reads `sqlite_master` and `pragma_table_info` | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B6 (`range` @ internal/journal/schema_test.go:223) | assertion guard at internal/journal/schema_test.go:223 (`range`) | none — the test only reads `sqlite_master` and `pragma_table_info` | `t.Fatal`/`t.Error` on violation | this test is itself the required test |
| B7 (`if` @ internal/journal/schema_test.go:226) | assertion guard at internal/journal/schema_test.go:226 (`if`) | none — the test only reads `sqlite_master` and `pragma_table_info` | `t.Fatal`/`t.Error` on violation | this test is itself the required test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournal` | a head-version journal | `t.Fatal` | schema_test.go:20 |
| `db.QueryContext` on `sqlite_master` | the actual table list | `t.Fatal` | schema_test.go:106 |

## State mutations and fallbacks

- Writes only inside `t.TempDir()`.
- No existing entry was removed, renamed or reordered — the diff is one inserted line.

## Safety conclusion

- Safe edit boundary: the `wantTables` literal. `TestV8IsPurelyAdditive` independently proves the additive claim by comparing every v7 schema object's DDL byte-for-byte against the same object in a v8 journal, so this list is not the only guard.
- High-risk impact: **no**. Test-only; the schema change it accompanies is one new table and three new indexes, with no ALTER of a released table.
