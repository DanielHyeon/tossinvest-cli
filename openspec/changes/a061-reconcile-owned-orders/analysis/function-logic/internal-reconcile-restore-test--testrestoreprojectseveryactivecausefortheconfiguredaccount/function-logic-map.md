# Function Logic Map: `TestRestoreProjectsEveryActiveCauseForTheConfiguredAccount`

Source: `internal/reconcile/restore_test.go`  
Function: `TestRestoreProjectsEveryActiveCauseForTheConfiguredAccount`  
Signature: `TestRestoreProjectsEveryActiveCauseForTheConfiguredAccount(params=1, results=0)`  
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`

## Inputs and invariants

- Inputs are the parameters in `TestRestoreProjectsEveryActiveCauseForTheConfiguredAccount(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/reconcile/restore_test.go:373 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/restore_test.go:374 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/restore_test.go:381 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/restore_test.go:386 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | range | internal/reconcile/restore_test.go:390 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | range | internal/reconcile/restore_test.go:393 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/reconcile/restore_test.go:400 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/reconcile/restore_test.go:404 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | range | internal/reconcile/restore_test.go:408 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/reconcile/restore_test.go:414 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/reconcile/restore_test.go:419 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `context.Background`: returned errors and state follow the mapped branches.
- `clock.NewFake`: returned errors and state follow the mapped branches.
- `openJournal`: returned errors and state follow the mapped branches.
- `j.EnterReconcile`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `noStaleGate`: returned errors and state follow the mapped branches.
- `trackerOn`: returned errors and state follow the mapped branches.
- `tracker.Restore`: returned errors and state follow the mapped branches.
- `tracker.Blocks`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `t.Errorf`: returned errors and state follow the mapped branches.
- `gate.CheckEntryFor`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 11 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
