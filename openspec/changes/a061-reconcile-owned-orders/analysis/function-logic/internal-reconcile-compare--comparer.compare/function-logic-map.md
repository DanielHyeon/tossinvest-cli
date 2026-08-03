# Function Logic Map: `Comparer.Compare`

Source: `internal/reconcile/compare.go`  
Function: `Comparer.Compare`  
Signature: `Comparer.Compare(params=2, results=1)`  
Source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`

## Inputs and invariants

- Inputs are the parameters in `Comparer.Compare(params=2, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:362 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | range | internal/reconcile/compare.go:368 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | range | internal/reconcile/compare.go:374 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/compare.go:375 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | range | internal/reconcile/compare.go:380 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/reconcile/compare.go:382 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | range | internal/reconcile/compare.go:389 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/reconcile/compare.go:394 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/reconcile/compare.go:397 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | switch | internal/reconcile/compare.go:403 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | case | internal/reconcile/compare.go:404 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | case | internal/reconcile/compare.go:406 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B13 | case | internal/reconcile/compare.go:412 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B14 | case | internal/reconcile/compare.go:416 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B15 | range | internal/reconcile/compare.go:426 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B16 | range | internal/reconcile/compare.go:427 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B17 | if | internal/reconcile/compare.go:429 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B18 | range | internal/reconcile/compare.go:442 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B19 | if | internal/reconcile/compare.go:443 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B20 | if | internal/reconcile/compare.go:447 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B21 | range | internal/reconcile/compare.go:454 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B22 | if | internal/reconcile/compare.go:455 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B23 | range | internal/reconcile/compare.go:463 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B24 | if | internal/reconcile/compare.go:464 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B25 | if | internal/reconcile/compare.go:469 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `snap.AsOf.IsZero`: returned errors and state follow the mapped branches.
- `Format`: returned errors and state follow the mapped branches.
- `snap.AsOf.UTC`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `sort.Strings`: returned errors and state follow the mapped branches.
- `canonicalDecimal`: returned errors and state follow the mapped branches.
- `isZeroDecimal`: returned errors and state follow the mapped branches.
- `quantitiesAgree`: returned errors and state follow the mapped branches.
- `localOrdersForComparison`: returned errors and state follow the mapped branches.
- `brokerOrderIdentityForLocal`: returned errors and state follow the mapped branches.
- `localOrder.Identity`: returned errors and state follow the mapped branches.
- `identitiesCompatible`: returned errors and state follow the mapped branches.
- `sort.Slice`: returned errors and state follow the mapped branches.
- `less`: returned errors and state follow the mapped branches.
- `missing.Identity`: returned errors and state follow the mapped branches.
- `brokerOrderIdentity`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 36 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
