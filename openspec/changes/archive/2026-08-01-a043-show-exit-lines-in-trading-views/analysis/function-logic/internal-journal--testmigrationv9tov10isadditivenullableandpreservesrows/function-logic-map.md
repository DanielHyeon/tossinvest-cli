# Function Logic Map: `TestMigrationV9ToV10IsAdditiveNullableAndPreservesRows`

- Source: `internal/journal/migration_v10_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v9 database and v8 rows | historical migration fixture | migration plan | head reaches v11; v10 columns stay additive |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | close/version/row/column assertions | test-only database | fail test | named migration test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| journal migration | preserve v10 claims at current head | explicit failure | AST + named test |

## State mutations and fallbacks

- Head assertion uses `SchemaVersion`; column assertions remain v10-specific.

## Safety conclusion

- Safe edit boundary: migration regression test.
- High-risk impact: no production mutation.
