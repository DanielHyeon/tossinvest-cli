# Function Logic Map: `Tracker.persist`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.persist`  
Signature: `Tracker.persist(params=2, results=2)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters in `Tracker.persist(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:514 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | range | internal/reconcile/mismatch.go:524 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/mismatch.go:532 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/mismatch.go:536 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | range | internal/reconcile/mismatch.go:545 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/reconcile/mismatch.go:553 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/reconcile/mismatch.go:557 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `append`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `firstNonEmpty`: returned errors and state follow the mapped branches.
- `t.Journal.EnterReconcile`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `scopeLabel`: returned errors and state follow the mapped branches.
- `blockFromReconcileState`: returned errors and state follow the mapped branches.
- `t.Journal.ReleaseReconcile`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 8 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
