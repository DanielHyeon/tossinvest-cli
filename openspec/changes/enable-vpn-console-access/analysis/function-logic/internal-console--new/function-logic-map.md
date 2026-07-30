# Function Logic Map: `New`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`85d2bb460f96627d062ed9cfbccfd64ca13ad3de1dee21d0af3d3d70e8e70178`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | required verification seam plus zero local or fully validated remote configuration | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 379 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B2 | existing if branch at line 383 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B3 | existing if branch at line 387 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B4 | existing if branch at line 399 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B5 | existing if branch at line 402 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |
| B6 | existing if branch at line 405 | only the branch's documented state transition | existing return/error contract | `TestRemoteConfigurationIsAllOrNothing` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| newRemoteRuntime, token generation, route construction | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- refuse partial remote configuration before constructing an HTTP handler.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: refuse partial remote configuration before constructing an HTTP handler.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
