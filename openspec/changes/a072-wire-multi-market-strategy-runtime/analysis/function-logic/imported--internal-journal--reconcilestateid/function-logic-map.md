# Function Logic Map: `reconcileStateID`

- Source: `internal/journal/reconcile_states.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account/symbol/market/cause/time | normalized scope identity and journal time | EnterReconcile | deterministic length-prefixed digest |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each identity component | hash write only | stable ID | market coexistence test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SHA-256 | derive collision-resistant journal ID | deterministic, no error exposed | AST |

## State mutations and fallbacks

- Market becomes part of the identity so simultaneous KR/US observations cannot collide.

## Safety conclusion

- Safe edit boundary: add normalized market component to existing length-prefixed sequence.
- High-risk impact: yes — durable row primary key identity.
