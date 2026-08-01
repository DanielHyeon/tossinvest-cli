# Function Logic Map: `appendExitEventTx`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| event row | position and optional evaluation | journal caller | encode/insert error, transaction rolls back |
| lifecycle generation | current exit-state binding | same transaction | absence/mismatch fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | evaluated event | encode full saved/recomputed/effective tuple | error or insert | snapshot tests |
| B2 | legacy/state event | nullable evaluation evidence | insert | exit-state tests |
| B3 | encoding/insert fails | none committed | error | crash-hook tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `encodeStoredSnapshot` | persist immutable evidence | error aborts caller transaction | CodeGraph + AST |
| `tx.ExecContext` | append event | caller owns commit | CodeGraph + AST |

## State mutations and fallbacks

- Append-only. New lifecycle generation is derived from the locked exit row, not trusted from a caller.

## Safety conclusion

- Safe edit boundary: add the derived lifecycle column without weakening existing snapshot tuple.
- High-risk impact: yes — audit/event attribution boundary.
