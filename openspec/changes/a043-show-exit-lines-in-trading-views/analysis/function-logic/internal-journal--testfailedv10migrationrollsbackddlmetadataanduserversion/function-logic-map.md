# Function Logic Map: `TestFailedV10MigrationRollsBackDDLMetadataAndUserVersion`

- Source: `internal/journal/migration_v10_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broken v10 plan from v9 | deterministic failure | migration recovery contract | rollback v9; restored backup reaches v11 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B12 | failure, backup, rollback, restore assertions | test-only database/files | fail test | named migration test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| migration backup helpers | preserve v10 rollback claim | explicit failure | AST + named test |

## State mutations and fallbacks

- Broken step stays v10-specific; only fresh-build final version is v11.

## Safety conclusion

- Safe edit boundary: migration recovery test.
- High-risk impact: no production mutation.
