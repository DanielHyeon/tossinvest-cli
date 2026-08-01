# Function Logic Map: `TestSchemaIndexes`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| migrated test journal | schema head v11 | migration plan | fail test on query/scan/missing index |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | query, row scan, rows error, expected-index loop and missing index | test-only reads | fatal/error | TestSchemaIndexes |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite schema catalog | assert lookup indexes | errors fail test | AST + schema test |

## State mutations and fallbacks

- Test-only schema assertion adds the v11 partial index.

## Safety conclusion

- Safe edit boundary: test only.
- High-risk impact: no.
