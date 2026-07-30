# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json` (`08f735ee79500678bf15ecd103e5af3608d672b47e3d3a64b1b8b4a9b7c99c49`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | logout joins every existing state-changing POST in the closed CSRF allowlist | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing range branch at line 393 | only the branch's documented state transition | existing return/error contract | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |
| B2 | existing switch branch at line 395 | only the branch's documented state transition | existing return/error contract | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |
| B3 | existing case branch at line 396 | only the branch's documented state transition | existing return/error contract | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |
| B4 | existing case branch at line 398 | only the branch's documented state transition | existing return/error contract | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |
| B5 | existing range branch at line 402 | only the branch's documented state transition | existing return/error contract | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |
| B6 | existing if branch at line 403 | only the branch's documented state transition | existing return/error contract | `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| registeredRoutes | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- fail on either an ungated mutation or a read accidentally made POST-only.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: fail on either an ungated mutation or a read accidentally made POST-only.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
