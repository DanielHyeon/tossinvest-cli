# Function Logic Map: `ListenOn`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`da1cbb194372c0f20b926357a65085ebb20021744f530209782e971d0357c254`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated IP-literal bind and port | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 449 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B2 | existing if branch at line 453 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |
| B3 | existing if branch at line 457 | only the branch's documented state transition | existing return/error contract | `TestRemoteListenerMustMatchTheValidatedBind` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| netip.ParseAddr, net.Listen | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- force tcp4/tcp6 explicitly so wildcard binds cannot silently change address family.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: force tcp4/tcp6 explicitly so wildcard binds cannot silently change address family.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
