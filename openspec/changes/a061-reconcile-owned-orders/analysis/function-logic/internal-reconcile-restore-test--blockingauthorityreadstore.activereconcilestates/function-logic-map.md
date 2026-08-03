# Function Logic Map: `blockingAuthorityReadStore.ActiveReconcileStates`

Source: `internal/reconcile/restore_test.go`  
Function: `blockingAuthorityReadStore.ActiveReconcileStates`  
Signature: `blockingAuthorityReadStore.ActiveReconcileStates(params=1, results=2)`  
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`

## Inputs and invariants

- Inputs are the parameters in `blockingAuthorityReadStore.ActiveReconcileStates(params=1, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/reconcile/restore_test.go:51 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `s.ReconcileStore.ActiveReconcileStates`: returned errors and state follow the mapped branches.
- `close`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 1 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
