# Function Logic Map: `TestFailedV10MigrationRollsBackDDLMetadataAndUserVersion`

- Source: `internal/journal/migration_v10_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broken migration fixture | valid test/domain fixture | durability contract | fail test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1..Bn | each AST branch | rollback and head restore | assertion/error | branch map below |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `restoreBackup` | enforce the mapped contract | fail closed; no automatic retry | CodeGraph + AST |

## State mutations and fallbacks

- rollback and head restore.

## Safety conclusion

- Safe edit boundary: not-applicable test expectation update.
- High-risk impact: no.
