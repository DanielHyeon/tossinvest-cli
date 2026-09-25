# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sqlite schema | current `SchemaVersion` (32 at HEAD 2026-09-25) table/column golden; a066 owns only its v22–v24 rows, later rows belong to the changes that added v25–v32 | migration contract | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | enumerate tables/columns and compare exact sets | read-only test queries | assertions | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite schema pragmas | inspect tables and columns | fatal on query errors | AST |

## State mutations and fallbacks

- Test-only read.

## Safety conclusion

- Safe edit boundary: add columns/tables to golden only. Wave 2A (2026-09-25) re-extracted the AST at HEAD: the body grew 78–349 → 78–397 because other changes appended v25–v32 golden rows; `difflib` alignment of old/new branches (kind + source line) is identical B1–B7.
- High-risk impact: no production mutation.
