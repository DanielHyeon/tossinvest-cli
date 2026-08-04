# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sqlite schema | v24 table/column golden | migration contract | test failure |

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

- Safe edit boundary: add v24 columns/tables to golden only.
- High-risk impact: no production mutation.
