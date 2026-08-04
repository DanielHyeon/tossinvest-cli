# Function Logic Map: `Open`

- Source: `internal/performance/store.go`
- Qualified function: `Open`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The path names a separate rebuildable performance database, never the trading journal. The function
must accept versions 0..`SchemaVersion`, migrate monotonically, refuse newer schemas, and close the DB
on every failure.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | database open/configuration fails | no usable store | error | invalid-path/open tests |
| B2 | schema-version query fails | closes DB | error | corrupt metadata tests |
| B3 | on-disk version is newer | closes DB; no downgrade | `ErrSchemaTooNew` | `TestStoreRefusesANewerDerivedSchema` |
| B4-B5 | version below v1 and migration fails/succeeds | transactional v1 migration or close | error/continue | migration crash/rollback tests |
| B6-B7 | version below current and attribution v2 migration fails/succeeds | transactional v2 migration or close | error/continue | SIGKILL and preservation tests |
| B8 | file permission hardening fails | closes DB | error | secure-file tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openDatabase` | open dedicated SQLite store with required pragmas | no journal fallback | store tests |
| `SchemaVersion` | detect migration/downgrade boundary | read failure closes | migration tests |
| `migrate` / `migrateAttributionV2` | monotonic transactional schemas | all-or-none; no retry/downgrade | crash suites |
| `securePerformanceFiles` | restrict DB/WAL/SHM permissions | failure closes | file-mode tests |

## State mutations and fallbacks

Mutates only the separate performance DB schema and filesystem modes. It never writes trading state,
orders, activation or lane configuration. No newer-schema downgrade or destructive rebuild fallback exists.

## Safety conclusion

- Safe edit boundary: close on every error and preserve append-only performance records across v1→v2.
- High-risk impact: moderate — derived analytics must be rebuildable and cannot affect execution authority.
