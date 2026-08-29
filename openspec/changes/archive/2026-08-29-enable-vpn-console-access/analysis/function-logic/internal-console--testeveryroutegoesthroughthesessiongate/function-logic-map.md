# Function Logic Map: `TestEveryRouteGoesThroughTheSessionGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json` (`08f735ee79500678bf15ecd103e5af3608d672b47e3d3a64b1b8b4a9b7c99c49`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | only /login and fixed /healthz are public; all operational paths require session0 | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing range branch at line 278 | only the branch's documented state transition | existing return/error contract | `TestEveryRouteGoesThroughTheSessionGate` |
| B2 | existing if branch at line 279 | only the branch's documented state transition | existing return/error contract | `TestEveryRouteGoesThroughTheSessionGate` |
| B3 | existing if branch at line 309 | only the branch's documented state transition | existing return/error contract | `TestEveryRouteGoesThroughTheSessionGate` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| registeredRoutes | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- fail when a future route lacks authentication without an explicit public classification.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: fail when a future route lacks authentication without an explicit public classification.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
