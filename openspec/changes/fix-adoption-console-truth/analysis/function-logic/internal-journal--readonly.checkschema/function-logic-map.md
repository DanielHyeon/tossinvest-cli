# Function Logic Map: `ReadOnly.checkSchema`

- Source: `internal/journal/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| read-only DB | existing SQLite file opened `mode=ro` | `OpenReadOnly` | typed failure and close |
| user_version | 0..current or future | SQLite pragma | too-new typed error |
| required tables/columns | tables in `readOnlyTables`; v9 `exit_states.policy_id` | read API query set | too-old typed error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `PRAGMA user_version` fails | none | wrapped inspection error | existing RO tests |
| B2 | version greater than build | none | `ErrSchemaTooNew` | existing direction test |
| B3 | required table absent | none | `ErrSchemaTooOld` | existing direction/attempt tests |
| B4 | required v9 column absent | none | `ErrSchemaTooOld` before query | new v8 compatibility test |
| B5 | metadata inspection fails | none | wrapped error | existing malformed DB behavior |
| B6 | all required shapes exist | records version only | nil | existing writer-visible tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `QueryRowContext` | read pragma and sqlite metadata | context-bound, no retry/write | CodeGraph + AST |

## State mutations and fallbacks

- Only `r.version` is assigned. Connection remains query-only.
- No migration, DDL, fallback database, or writable handle is introduced.

## Safety conclusion

- Safe edit boundary: extend schema-shape preflight with the exact column
  already required by read queries.
- High-risk impact: yes — journal compatibility; failure becomes earlier and
  typed, never more permissive.
