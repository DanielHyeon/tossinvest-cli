# Function Logic Map: `ExitObserver.openAdoptedState`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | adoption record is the sole t0 and recovery authority | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing switch branch at line 571 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpensAdopted` |
| B2 | existing case branch at line 572 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpensAdopted` |
| B3 | existing case branch at line 575 | only the branch's documented state transition | existing return/error contract | `TestExitObserverOpensAdopted` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| OpenAdoptedExitState | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- recover policy ID from the committed adoption record, never from later config.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: recover policy ID from the committed adoption record, never from later config.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
