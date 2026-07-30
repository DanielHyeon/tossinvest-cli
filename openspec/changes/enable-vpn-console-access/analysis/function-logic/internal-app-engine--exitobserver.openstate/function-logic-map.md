# Function Logic Map: `ExitObserver.openState`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | exit-eligible position; adopted and engine-entered origins remain distinct | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 513 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |
| B2 | existing if branch at line 517 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |
| B3 | existing if branch at line 522 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |
| B4 | existing if branch at line 527 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |
| B5 | existing if branch at line 534 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |
| B6 | existing switch branch at line 544 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |
| B7 | existing case branch at line 545 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |
| B8 | existing case branch at line 548 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpens` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| LookupDecision, ParsePreimage, OpenExitState, openAdoptedState | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- snapshot the startup common policy only when a new state is opened; existing state is never rebound.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: snapshot the startup common policy only when a new state is opened; existing state is never rebound.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
