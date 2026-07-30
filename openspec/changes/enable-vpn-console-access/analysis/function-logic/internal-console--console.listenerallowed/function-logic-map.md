# Function Logic Map: `Console.listenerAllowed`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`85d2bb460f96627d062ed9cfbccfd64ca13ad3de1dee21d0af3d3d70e8e70178`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | local listener is loopback; remote listener TCP address exactly matches the validated bind | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 565 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B2 | existing if branch at line 568 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B3 | existing if branch at line 572 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B4 | existing if branch at line 576 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B5 | existing if branch at line 580 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| loopbackOnly, netip.AddrFromSlice | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- close/refuse a listener whose real address differs from configuration.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: close/refuse a listener whose real address differs from configuration.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
