# Function Logic Map: `TestPartialIndexPredicates`

- Source: `internal/journal/execution_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sqlite schema | current v24 index DDL | migration contract | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | load each expected index and assert required DDL fragments | read-only test query | assertion | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite_master | inspect exact persisted DDL | fatal/error assertions | AST |

## State mutations and fallbacks

- Test-only read.

## Safety conclusion

- Safe edit boundary: golden names/predicates only.
- High-risk impact: no production mutation; protects high-risk schema.
