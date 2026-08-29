# Function Logic Map: `Console.Serve`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`85d2bb460f96627d062ed9cfbccfd64ca13ad3de1dee21d0af3d3d70e8e70178`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | listener must match selected local/remote mode before banner or serve | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 485 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B2 | existing if branch at line 501 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B3 | existing if branch at line 511 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B4 | existing select branch at line 519 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B5 | existing if branch at line 521 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B6 | existing if branch at line 533 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B7 | existing if branch at line 538 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |

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
