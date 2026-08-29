# Function Logic Map: `ExitObserver.checkLadderPolicyStillFits`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json` (`435abbc323679864d61b0d9c12a8c1ee6a0f239d5fd0b78d7a1c8de6d7342f3e`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | active rung must exist in the selected policy and its lock must not exceed stored baseline | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 785 | only the branch's documented state transition | existing return/error contract | `TestCheckLadderPolicyStillFits` |
| B2 | existing if branch at line 788 | only the branch's documented state transition | existing return/error contract | `TestCheckLadderPolicyStillFits` |
| B3 | existing if branch at line 795 | only the branch's documented state transition | existing return/error contract | `TestCheckLadderPolicyStillFits` |
| B4 | existing if branch at line 799 | only the branch's documented state transition | existing return/error contract | `TestCheckLadderPolicyStillFits` |
| B5 | existing if branch at line 802 | only the branch's documented state transition | existing return/error contract | `TestCheckLadderPolicyStillFits` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| LockPrice, CompareDecimal | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- accept the resolved per-state policy as input while preserving all fail-closed checks.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: accept the resolved per-state policy as input while preserving all fail-closed checks.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
