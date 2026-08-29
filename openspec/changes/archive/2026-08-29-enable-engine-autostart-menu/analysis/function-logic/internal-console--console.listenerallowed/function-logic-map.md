# Function Logic Map: `Console.listenerAllowed`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`da1cbb194372c0f20b926357a65085ebb20021744f530209782e971d0357c254`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | local listener is loopback; remote listener TCP address exactly matches the validated bind | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 550 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B2 | existing if branch at line 553 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B3 | existing if branch at line 557 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B4 | existing if branch at line 561 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B5 | existing if branch at line 565 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |

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
