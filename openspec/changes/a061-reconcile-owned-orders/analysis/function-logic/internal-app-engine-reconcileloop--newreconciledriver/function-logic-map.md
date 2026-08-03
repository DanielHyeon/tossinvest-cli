# Function Logic Map: `NewReconcileDriver`

Source: `internal/app/engine/reconcileloop.go`  
Function: `NewReconcileDriver`  
Signature: `NewReconcileDriver(params=1, results=2)`  
Source SHA-256: `accaa4c5f6645d8af7be3f1cbcd9ec61a7efc9f1f022be26b39b53789d867763`

## Inputs and invariants

- Inputs are the parameters in `NewReconcileDriver(params=1, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/app/engine/reconcileloop.go:280 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | case | internal/app/engine/reconcileloop.go:281 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | case | internal/app/engine/reconcileloop.go:283 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | case | internal/app/engine/reconcileloop.go:285 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | case | internal/app/engine/reconcileloop.go:287 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | case | internal/app/engine/reconcileloop.go:289 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | case | internal/app/engine/reconcileloop.go:292 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/app/engine/reconcileloop.go:296 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/app/engine/reconcileloop.go:297 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/app/engine/reconcileloop.go:309 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `exitpolicy.CommonPolicyByID`: returned errors and state follow the mapped branches.
- `clock.System`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 8 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
