# Function Logic Map: `Console.mutating`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`85d2bb460f96627d062ed9cfbccfd64ca13ad3de1dee21d0af3d3d70e8e70178`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | POST, then remote same-origin, then form/CSRF, then handler | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 798 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |
| B2 | existing if branch at line 804 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |
| B3 | existing if branch at line 809 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |
| B4 | existing if branch at line 813 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| remote.sameOrigin, ParseForm, tokenEqual | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- all independent request gates must pass before any state-changing handler executes.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: all independent request gates must pass before any state-changing handler executes.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
