# Function Logic Map: `releaseControllingStore.ReleaseReconciles`

Source: `internal/reconcile/restore_test.go`  
Function: `releaseControllingStore.ReleaseReconciles`  
Signature: `releaseControllingStore.ReleaseReconciles(params=2, results=2)`  
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`

## Inputs and invariants

- Inputs are the parameters in `releaseControllingStore.ReleaseReconciles(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:105 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/restore_test.go:108 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `append`: returned errors and state follow the mapped branches.
- `errors.New`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 1 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
