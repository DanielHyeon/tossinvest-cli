# Function Logic Map: `scanExitProgress`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| position id | exact position | judgement transaction | typed not-found/error |\n| lifecycle generation | legacy default 1 or positive explicit value | exit state row | later generation guard refuses invalid scope |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | missing/query error | none | typed/wrapped error | exit judgement suite |\n| B3 | active rung present | hydrate | result | ratchet suite |\n| B4-B5 | effective snapshot present/invalid | decode or fail | result/error | snapshot suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tx.QueryRowContext` | read position + lifecycle progress | same transaction | AST |\n| `decodeStoredSnapshot` | validate effective tuple | fail closed | AST |

## State mutations and fallbacks

- Read-only hydration inside the judgement transaction.

## Safety conclusion

- Safe edit boundary: carry lifecycle generation alongside existing monotone progress without touching guarded execution columns.
- High-risk impact: yes
