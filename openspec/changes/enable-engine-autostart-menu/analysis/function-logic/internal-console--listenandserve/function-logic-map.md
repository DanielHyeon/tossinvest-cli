# Function Logic Map: `ListenAndServe`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`da1cbb194372c0f20b926357a65085ebb20021744f530209782e971d0357c254`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated Console chooses exactly one listener constructor | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 534 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B2 | existing if branch at line 538 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B3 | existing else branch at line 540 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B4 | existing if branch at line 543 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| New, Listen, ListenOn, Serve | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- partial remote configuration never falls back to local or exposed HTTP.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: partial remote configuration never falls back to local or exposed HTTP.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
