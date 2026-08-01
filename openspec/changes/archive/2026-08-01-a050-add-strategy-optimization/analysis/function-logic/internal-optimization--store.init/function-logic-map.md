# Function Logic Map: `Store.init`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| SQLite user version and lifecycle tables | version is supported; migration is atomic; append-only tables cannot be updated/deleted | database transaction | error and rollback; never silently downgrade/upgrade unknown schema |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | schema is newer than this binary | none | refuse Open | `TestOpenRefusesNewerSchema` |
| B2 | migration SQL fails | transaction rollback | wrapped error | migration atomicity test |
| B3 | migration succeeds | set `user_version`, add tables/triggers | initial snapshot step | `TestSchemaMigrationIsVersionedAndAppendOnly` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `migrateSchema` | creates schema and append-only triggers atomically | transaction has no retry; error aborts Open | AST B1-B3 |
| `ensureInitialSnapshot` | establishes version 1 exactly once | transactional CAS | CodeGraph impact |

## State mutations and fallbacks

- DDL migration and user-version write are a single transaction. Application tables are append-only; only `optimization_control` advances current version.

## Safety conclusion

- Safe edit boundary: control-store schema only.
- High-risk impact: no LIVE/journal authority is introduced.
