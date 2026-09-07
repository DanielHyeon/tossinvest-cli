# Function Logic Map: `TestSchemaIndexes`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sqlite indexes | required lookup/uniqueness index golden | migration | test failure on missing index |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | query/scan/assert branches | read-only test queries | test failure | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite metadata query | index golden | fatal on error | AST |

## State mutations and fallbacks

- Test-only; validates order identity/predecessor unique indexes.

## Safety conclusion

- Safe edit boundary: schema index evidence.
- High-risk impact: journal uniqueness contract test.
