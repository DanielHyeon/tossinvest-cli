# Function Logic Map: `positionPolicyReconcileScope`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| scope | account, market, symbol, or future/unknown reconcile scope | reconcile tracker block | maps known scopes exactly; unknown value remains typed unknown rather than widening coverage |
| output | transport-neutral `positionpolicy.ReconcileScope` | a052 read model | never grants resolution or changes the tracker |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `switch` at line 121 | `switch scope {` | none beyond calls listed below | early return/error path nearby | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |
| B2 | `case` at line 122 | `case reconcile.ScopeAccount:` | none beyond calls listed below | early return/error path nearby | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |
| B3 | `case` at line 124 | `case reconcile.ScopeMarket:` | none beyond calls listed below | early return/error path nearby | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |
| B4 | `case` at line 126 | `case reconcile.ScopeSymbol:` | none beyond calls listed below | early return/error path nearby | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |
| B5 | `case` at line 128 | `default:` | none beyond calls listed below | early return/error path nearby | TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicy.ReconcileScope` | execute the explicit dependency at line 129 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- Pure mapping with no assignment, journal access, tracker mutation, or network call.
- Known scopes preserve their exact coverage. The default preserves the raw typed value so downstream coverage logic can treat it as unknown/fail-closed.
- No reconciliation-resolution authority is created.

## Safety conclusion

- Safe edit boundary: Only the approved read/projection or fail-closed adoption boundary may change; no order placement, reconciliation resolution, or live toggle mutation is authorized.
- High-risk impact: yes; adoption provenance, reconciliation blocking, or persisted position lifecycle is money-sensitive.
