# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current SQLite schema | exact table and column sets | released migrations + current v25 | test failure only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | enumerate tables, inspect each expected column set and fail on missing/extra shape | read-only test queries | assertions | `TestSchemaTablesAndColumns` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite_master / table_info | current schema golden | fatal on storage/scan mismatch | AST |

## State mutations and fallbacks

- Test-only reads; no production state mutation.

## Safety conclusion

- Safe edit boundary: add v25 tables/columns to the literal golden.
- High-risk impact: no runtime impact; protects high-risk journal schema.
