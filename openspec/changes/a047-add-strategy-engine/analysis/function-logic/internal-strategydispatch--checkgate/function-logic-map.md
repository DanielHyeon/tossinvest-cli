# Function Logic Map: `checkGate`

- Source: `internal/strategydispatch/dispatch.go`
- CodeGraph callers/callees: initial and leased checks inside `dispatchValidated`
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| decision binding | exact 60-field snapshot | opaque DecisionRecord + activation snapshot | activation refusal |
| operational blockers | authority-owned booleans | lane/kill/protection/reconcile/scheduler/gate stores | first stable refusal |
| order settings | fixed LIMIT/KRW + manifest digest | server manifest | activation refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | activation/decision/order mismatch | none | activation | exhaustive binding + precedence table |
| B2 | lane desired/effective OFF | none | lane off | gate table |
| B3 | kill switch | none | kill | gate table |
| B4 | protection unwired | none | protection | gate table |
| B5 | reconciliation unhealthy | none | reconcile | gate table |
| B6 | scheduler invalid | none | scheduler | gate table |
| B7 | autostart disabled | none | autostart | gate table |
| B8 | gate closed | none | gate | gate table |
| B9 | LIVE unapproved | none | live | gate table |
| Success | all exact | none | empty reason | post-validation core test |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| `DecisionBinding` | exact full-record value conversion | activation refusal | 60/60 reflection test |

## State mutations and fallbacks

- Pure ordered predicate. No state mutation or fallback; the first blocker is stable.

## Safety conclusion

- Pure ordered predicate. No mutation, I/O, fallback, or user input.
