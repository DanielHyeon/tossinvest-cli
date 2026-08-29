# Function Logic Map: `Console.mutating`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`da1cbb194372c0f20b926357a65085ebb20021744f530209782e971d0357c254`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | POST, then remote same-origin, then form/CSRF, then handler | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 771 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |
| B2 | existing if branch at line 777 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |
| B3 | existing if branch at line 782 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |
| B4 | existing if branch at line 786 | only the branch's documented state transition | existing return/error contract | `TestRemotePeerHostOriginAndCSRFAreIndependentGates` |

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
