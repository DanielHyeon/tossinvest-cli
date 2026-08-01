# Function Logic Map: `Journal.RecordExitJudgement`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| legacy caller judgement | same contract as `RecordExitJudgementResult` | a041 snapshot, a042 recovery spec | returns delegated error and discards typed result |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | delegated success/error | none beyond delegated transaction | error only | all existing journal judgement tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RecordExitJudgementResult` | compatibility wrapper delegates the complete transaction | exact error propagation | CodeGraph + AST |

## State mutations and fallbacks

- No independent state or fallback logic; the typed result API and transaction helper own all decisions.

## Safety conclusion

- Safe edit boundary: journal-only atomic write; broker submission remains after commit in engine.
- High-risk impact: yes; crash and race tests required.
