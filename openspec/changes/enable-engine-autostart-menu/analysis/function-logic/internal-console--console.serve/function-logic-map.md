# Function Logic Map: `Console.Serve`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`da1cbb194372c0f20b926357a65085ebb20021744f530209782e971d0357c254`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | listener must match selected local/remote mode before banner or serve | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 470 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B2 | existing if branch at line 486 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B3 | existing if branch at line 496 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B4 | existing select branch at line 504 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B5 | existing if branch at line 506 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B6 | existing if branch at line 518 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B7 | existing if branch at line 523 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| listenerAllowed, http.Server.Serve, ServeTLS, Shutdown | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- use TLS 1.3 and bounded server timeouts remotely while preserving settle-before-shutdown.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: use TLS 1.3 and bounded server timeouts remotely while preserving settle-before-shutdown.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
