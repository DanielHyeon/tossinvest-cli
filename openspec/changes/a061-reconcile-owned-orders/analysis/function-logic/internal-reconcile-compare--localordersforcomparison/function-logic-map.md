# Function Logic Map: `localOrdersForComparison`

Source: `internal/reconcile/compare.go`  
Function: `localOrdersForComparison`  
Signature: `localOrdersForComparison(params=1, results=1)`  
Source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`

## Inputs and invariants

- Inputs are the parameters in `localOrdersForComparison(params=1, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:480 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | range | internal/reconcile/compare.go:482 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | range | internal/reconcile/compare.go:497 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/compare.go:498 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `identity.canonical`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `sort.Slice`: returned errors and state follow the mapped branches.
- `less`: returned errors and state follow the mapped branches.
- `orders.Identity`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 12 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
