# Function Logic Map: `Context.ReconcileDriver`

- Source: `internal/app/engine/reconcileloop.go`
- AST evidence: `ast.json` (`157aa37d842a4ab0379b0364a9590d18d5b3ef27b9a655dd3e6ed803120dcc29`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | Context dependencies, freshness and adoption settings are validated | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 328 | only the branch's documented state transition | existing return/error contract | `TestContextReconcileDriverPolicy` |
| B2 | existing if branch at line 331 | only the branch's documented state transition | existing return/error contract | `TestContextReconcileDriverPolicy` |
| B3 | existing if branch at line 342 | only the branch's documented state transition | existing return/error contract | `TestContextReconcileDriverPolicy` |
| B4 | existing if branch at line 345 | only the branch's documented state transition | existing return/error contract | `TestContextReconcileDriverPolicy` |
| B5 | existing if branch at line 348 | only the branch's documented state transition | existing return/error contract | `TestContextReconcileDriverPolicy` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| NewReconcileDriver | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- inject the same startup common policy into adoption requests.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: inject the same startup common policy into adoption requests.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
