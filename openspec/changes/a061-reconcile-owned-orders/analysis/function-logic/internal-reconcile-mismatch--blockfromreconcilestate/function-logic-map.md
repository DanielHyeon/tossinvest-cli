# Function Logic Map: `blockFromReconcileState`

Source: `internal/reconcile/mismatch.go`  
Function: `blockFromReconcileState`  
Signature: `blockFromReconcileState(params=1, results=1)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters in `blockFromReconcileState(params=1, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:804 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/mismatch.go:807 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/mismatch.go:810 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `execgw.ReconcileReasonFor`: returned errors and state follow the mapped branches.
- `state.AccountWide`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 8 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
