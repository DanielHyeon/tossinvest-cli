# Function Logic Map: `OpenReadOnly`

- Source: `internal/performance/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| performance DB path | existing regular clean-checkpoint SQLite file; no WAL/SHM sidecars | profile composition plus filesystem | typed missing/WAL/change error; creates nothing |
| schema version | exactly current `performance.SchemaVersion` | persisted `PRAGMA user_version` | old/new schema rejected without migration |
| file identity | main DB size and mtime unchanged across validation | filesystem stat | changed DB rejected as unavailable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | immutable read-only DB open/identity validation fails | no writer side effect | propagate typed error | missing DB and active WAL tests |
| B2 | schema version query fails | closes local read handle | wrapped read error | malformed DB case |
| B3 | switch on schema relationship | none | current continues; mismatch rejects | current/old/new schema tests |
| B4 | schema older than build | closes deferred handle | `ErrSchemaTooOld` | `TestOpenReadOnlyRejectsOldSchemaWithoutMigrating` |
| B5 | schema newer than build | closes deferred handle | `ErrSchemaTooNew` | newer schema compatibility test |
| B6 | file/sidecar identity changed while validating | closes deferred handle | typed change/WAL error | immutable change/WAL coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openImmutablePerformanceDB` | opens `mode=ro`, `immutable=1`, `query_only` after clean-sidecar identity check | one attempt; all errors fail closed | read-only tests and AST |
| `PRAGMA user_version` | compatibility check without migration | background context, one query, no retry | old-schema no-migration test |
| `unchangedImmutablePerformanceDB` | prevents accepting a changing/stale immutable snapshot | one post-open check | WAL/change tests |

## State mutations and fallbacks

- The SQLite handle is always closed before return. Successful output stores only path/lifecycle state; it owns no long-lived writer or collector.
- No directory, DB, sidecar, schema, journal mode, or trading state is created or changed; any uncertainty is unavailable.

## Safety conclusion

- Safe edit boundary: read-only derived performance capability from a clean, exact-schema checkpoint.
- High-risk impact: yes; stale evidence could influence optimization, so WAL/schema/file drift fails closed.
