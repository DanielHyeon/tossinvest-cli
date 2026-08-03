# Function Logic Map: `Tracker.Restore`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.Restore`  
Signature: `Tracker.Restore(params=1, results=1)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters in `Tracker.Restore(params=1, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:578 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/mismatch.go:582 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | range | internal/reconcile/mismatch.go:589 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/mismatch.go:590 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/reconcile/mismatch.go:607 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/reconcile/mismatch.go:610 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `t.Journal.ActiveReconcileStates`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `blockFromReconcileState`: returned errors and state follow the mapped branches.
- `block.Key`: returned errors and state follow the mapped branches.
- `t.mu.Lock`: returned errors and state follow the mapped branches.
- `t.maxFailures`: returned errors and state follow the mapped branches.
- `t.Gate.RebuildReconcileProjection`: returned errors and state follow the mapped branches.
- `t.mu.Unlock`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 12 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
