# Function Logic Map: `TestSchemaIndexes`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sqlite schema | current required index set | schema golden | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | enumerate sqlite indexes and assert required v24 names | read-only test query | assertion | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite_master | list indexes | fatal on query/scan error | AST |

## State mutations and fallbacks

- Test-only read.

## Safety conclusion

- Safe edit boundary: expected index list only.
- High-risk impact: no production mutation.
