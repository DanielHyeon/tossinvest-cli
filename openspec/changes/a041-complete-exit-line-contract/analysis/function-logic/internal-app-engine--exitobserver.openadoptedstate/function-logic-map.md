# Function Logic Map: `ExitObserver.openAdoptedState`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| adopted position | durable adoption with policy/baseline seed | journal adoption record | fail without fallback to entry decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | completed duplicate state | none | empty state, nil | existing adoption recovery test |
| B2 | open/read failure | none | wrapped error | existing adoption tests |
| B3 | opened | structured audit log | state | existing adoption tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OpenAdoptedExitState` | atomically seed from adoption | duplicate completed is distinct from error | CodeGraph + AST |

## State mutations and fallbacks

- Existing function is unchanged; current-file AST is refreshed because exitloop changed elsewhere.

## Safety conclusion

- Safe edit boundary: no behavior change.
- High-risk impact: yes — adopted protection opening.
