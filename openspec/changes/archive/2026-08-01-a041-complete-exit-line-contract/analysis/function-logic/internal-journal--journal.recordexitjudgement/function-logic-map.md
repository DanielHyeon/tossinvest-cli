# Function Logic Map: `Journal.RecordExitJudgement`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement | monotone state plus optional complete provenance/proposal | exitpolicy snapshot | invalid/duplicate write refused atomically |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | missing ID/invalid provenance/proposal | no transaction | invalid request | provenance tests |
| B2 | completed or decreasing state | rollback | state error | existing monotonicity tests |
| B3 | proposal pending concurrently | rollback loser | `ErrProposalPending` | concurrent engine/journal test |
| B4 | valid judgement | update, arm, append event | commit | existing journal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanExitProgress` | lock/read current state inside immediate transaction | read errors abort | CodeGraph + AST |
| `armExitProposalTx` | exactly-once pending arm | concurrent loser gets pending error | CodeGraph + AST |
| `appendExitEventTx` | append history in same transaction | failure rolls back state and arm | CodeGraph + AST |

## State mutations and fallbacks

- Provenance remains typed on judgement/proposal for a042; the existing transaction remains the dedup authority.

## Safety conclusion

- Safe edit boundary: validate matching snapshot/decision provenance before existing transactional writes.
- High-risk impact: yes — atomic proposal arming.
