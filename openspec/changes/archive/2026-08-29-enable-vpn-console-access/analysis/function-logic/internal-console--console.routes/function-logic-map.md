# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`85d2bb460f96627d062ed9cfbccfd64ca13ad3de1dee21d0af3d3d70e8e70178`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | login/logout exist only in authenticated remote mode; health and every operational route retain their static wrapper classification | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 674 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkConsoleNeedsNoApplicationSession` |
| B2 | existing if branch at line 739 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkConsoleNeedsNoApplicationSession` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| http.ServeMux.HandleFunc, session0, mutating, remote.security | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- remove login lifecycle from trusted-network mode while preserving network middleware and mutating gates.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: remove login lifecycle from trusted-network mode while preserving network middleware and mutating gates.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
