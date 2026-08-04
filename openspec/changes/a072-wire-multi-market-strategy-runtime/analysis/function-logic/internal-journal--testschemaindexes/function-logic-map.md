# Function Logic Map: `TestSchemaIndexes`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| current SQLite schema | required lookup index names | current migration contract | test failure only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | enumerate indexes and assert every required v25 lookup path | read-only test queries | assertions | `TestSchemaIndexes` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite_master | index-name golden | fatal on storage/scan mismatch | AST |

## State mutations and fallbacks

- Test-only reads; no production mutation.

## Safety conclusion

- Safe edit boundary: append v25 index names only.
- High-risk impact: no runtime impact; protects CAS query paths.
