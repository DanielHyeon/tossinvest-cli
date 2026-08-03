# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sqlite schema | exact v20 table/column golden | migration | test failure on drift |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | query/scan/compare branches | read-only test queries | test failure | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite metadata queries | schema golden | fatal on error | AST |

## State mutations and fallbacks

- Test-only; no production mutation beyond isolated test migration.

## Safety conclusion

- Safe edit boundary: additive schema evidence.
- High-risk impact: journal schema contract test.
