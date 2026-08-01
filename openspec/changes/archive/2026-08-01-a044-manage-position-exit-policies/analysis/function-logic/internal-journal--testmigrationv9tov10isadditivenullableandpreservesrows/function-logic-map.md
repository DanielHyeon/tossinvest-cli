# Function Logic Map: `TestMigrationV9ToV10IsAdditiveNullableAndPreservesRows`

- Source: `internal/journal/migration_v10_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| migration fixture | valid test/domain fixture | schema contract | fail test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1..Bn | each AST branch | head migration preserves v10 evidence | assertion/error | branch map below |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournalAt` | enforce the mapped contract | fail closed; no automatic retry | CodeGraph + AST |

## State mutations and fallbacks

- head migration preserves v10 evidence.

## Safety conclusion

- Safe edit boundary: not-applicable test expectation update.
- High-risk impact: no.
