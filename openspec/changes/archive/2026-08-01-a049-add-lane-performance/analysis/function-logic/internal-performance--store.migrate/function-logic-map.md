# Function Logic Map: `Store.migrate`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current/target version | 0..SchemaVersion; newer is refused | SQLite `user_version` | fail closed without schema mutation |
| schema DDL | complete v1 schema executed in one immediate transaction | compiled `schemaV1` | rollback leaves version 0 and no derived tables |
| phase hook | test-only no-op callback | package test harness | SIGKILL proves every intermediate phase is all-or-none |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | schema-version read fails | none | error | DB error contract |
| B2 | current > target | none | `ErrSchemaTooNew` | newer-schema test |
| B3 | current == target | none | success | reopen/replay tests |
| B4 | BeginTx fails | none | wrapped error | DB error contract |
| B5 | DDL fails/dies after DDL | temporary tx only | error/all-or-none | failed migration + phase SIGKILL tests |
| B6 | version assignment fails/dies after version | temporary tx only | error/all-or-none | phase SIGKILL test |
| B7 | commit fails | rollback/recovery | error; otherwise schema+version durable | migration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SchemaVersion` | reads authoritative SQLite version | read-only | current HEAD + AST |
| `BeginTx`/`ExecContext`/`Commit` | atomic schema creation and version publish | any error returned; no retry | migration tests |

## State mutations and fallbacks

- Only the rebuildable `performance.db` is mutated. The authoritative journal is not opened.
- Test phase hooks cannot create production behavior because their default is a no-op and tests install them only in the child process.

## Safety conclusion

- Safe edit boundary: derived database v1 bootstrap transaction.
- High-risk impact: persistence integrity; no order/LIVE capability.
