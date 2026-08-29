# Function Logic Map: `ExitObserver.judgeLadder`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | stored state policy ID selects one immutable registry policy; adopted origin is retained | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 731 | only the branch's documented state transition | existing return/error contract | `TestExitObserverLadder` |
| B2 | existing if branch at line 735 | only the branch's documented state transition | existing return/error contract | `TestExitObserverLadder` |
| B3 | existing if branch at line 740 | only the branch's documented state transition | existing return/error contract | `TestExitObserverLadder` |
| B4 | existing if branch at line 741 | only the branch's documented state transition | existing return/error contract | `TestExitObserverLadder` |
| B5 | existing if branch at line 762 | only the branch's documented state transition | existing return/error contract | `TestExitObserverLadder` |
| B6 | existing if branch at line 768 | only the branch's documented state transition | existing return/error contract | `TestExitObserverLadder` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| checkLadderPolicyStillFits, RungIndex, EvaluateLadder, record | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- unknown/mismatched snapshots refuse judgement; record/Guardian/submit ordering remains untouched.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: unknown/mismatched snapshots refuse judgement; record/Guardian/submit ordering remains untouched.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
