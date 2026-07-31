# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sqlite schema metadata | current isolated journal | sqlite_master/pragma | fatal assertion on drift |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | query, scan, compare tables and required columns | read-only metadata checks | fatal assertion | schema contract test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite_master/pragma_table_info | pin schema v10 quarantine table and columns | no retry; test fails closed | AST |

## State mutations and fallbacks

- Expected table list adds the generation quarantine ledger.

## Safety conclusion

- Safe edit boundary: schema contract assertion only.
- High-risk impact: no direct runtime behavior; protects the high-risk ledger schema.
