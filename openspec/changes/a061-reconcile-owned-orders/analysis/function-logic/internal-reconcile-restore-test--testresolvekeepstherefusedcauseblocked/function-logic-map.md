# Function Logic Map: `TestResolveKeepsTheRefusedCauseBlocked`

Source: `internal/reconcile/restore_test.go`  
Function: `TestResolveKeepsTheRefusedCauseBlocked`  
Signature: `TestResolveKeepsTheRefusedCauseBlocked(params=1, results=0)`  
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`

## Inputs and invariants

- Inputs are the parameters in `TestResolveKeepsTheRefusedCauseBlocked(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:533 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/restore_test.go:541 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/restore_test.go:547 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/restore_test.go:550 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/reconcile/restore_test.go:553 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/reconcile/restore_test.go:556 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `context.Background`: returned errors and state follow the mapped branches.
- `clock.NewFake`: returned errors and state follow the mapped branches.
- `openJournal`: returned errors and state follow the mapped branches.
- `j.EnterReconcile`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `noStaleGate`: returned errors and state follow the mapped branches.
- `trackerOn`: returned errors and state follow the mapped branches.
- `tracker.Restore`: returned errors and state follow the mapped branches.
- `tracker.Resolve`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `tracker.Blocks`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `gate.CheckEntryFor`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 11 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
