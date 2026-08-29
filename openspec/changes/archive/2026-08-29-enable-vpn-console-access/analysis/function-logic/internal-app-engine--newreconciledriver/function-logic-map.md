# Function Logic Map: `NewReconcileDriver`

- Source: `internal/app/engine/reconcileloop.go`
- AST evidence: `ast.json` (`157aa37d842a4ab0379b0364a9590d18d5b3ef27b9a655dd3e6ed803120dcc29`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing switch branch at line 279 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B2 | existing case branch at line 280 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B3 | existing case branch at line 282 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B4 | existing case branch at line 284 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B5 | existing case branch at line 286 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B6 | existing case branch at line 288 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B7 | existing case branch at line 291 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B8 | existing if branch at line 295 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B9 | existing if branch at line 296 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |
| B10 | existing if branch at line 308 | only the branch's documented state transition | existing return/error contract | `TestNewReconcileDriver` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST-listed callees | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- preserve existing fail-closed behavior.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: preserve existing fail-closed behavior.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
