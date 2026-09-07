# Function Logic Map: `ReadOnly.checkSchema`

- Source: `internal/journal/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.db` | query-only SQLite handle | `OpenReadOnly` | return wrapped inspection error |
| `PRAGMA user_version` | `0..SchemaVersion` for this build | SQLite header | `ErrSchemaTooNew` when newer |
| required tables/columns | released schema prerequisites | `readOnlyTables`, `readOnlyColumns`, and version-gated v20 requirements | `ErrSchemaTooOld` with missing identities |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | reading `user_version` fails | none | wrapped read error | existing damaged/open tests |
| B2 | version exceeds `SchemaVersion` | none | `ErrSchemaTooNew` | `TestOpenReadOnlyRejectsNewerSchema` |
| B3 | a baseline table is missing | append local `missing` only | `ErrSchemaTooOld` after scan | existing pre-v6 table tests |
| B4 | baseline table inspection fails | none | wrapped inspection error | query/open error coverage |
| B5 | baseline column is missing | append local `missing` only | `ErrSchemaTooOld` after scan | v8/v9/v14 compatibility tests |
| B6 | version is at least 20 and a campaign table/column is missing | append local `missing` only | `ErrSchemaTooOld` before any campaign query | new damaged-v20 preflight tests |
| B7 | all prerequisites exist | assign already-read `r.version` | nil | current-schema readonly tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `QueryRowContext` | inspect version/table/column existence | no retry; first driver error is returned | AST |
| `errors.Is` | distinguish missing rows | pure | AST |
| `strings.Join` | stable missing-schema detail | pure | AST |

## State mutations and fallbacks

- No database mutation, migration, broker call, runtime toggle, or fallback query.
- The v20 check is intentionally after released baseline checks so specific v8/v9/v14 failure contracts remain authoritative.

## Safety conclusion

- Safe edit boundary: additive version-gated preflight only; the query-only DSN and public handle remain unchanged.
- High-risk impact: yes (journal observability), but no trading-path side effect.
