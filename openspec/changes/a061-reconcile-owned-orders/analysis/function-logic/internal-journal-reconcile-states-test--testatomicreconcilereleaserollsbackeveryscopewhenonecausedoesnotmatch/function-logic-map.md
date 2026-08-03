# Function Logic Map: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`

Source: `internal/journal/reconcile_states_test.go`  
Function: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`  
Signature: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch(params=1, results=0)`  
Source SHA-256: `d2de10c4ae8c4d15346e190fa03cbc7bd4db7648bd7b4ab102272225dd1785a6`

## Inputs and invariants

- Inputs are the parameters in `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/journal/reconcile_states_test.go:296 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/reconcile_states_test.go:300 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/reconcile_states_test.go:310 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/reconcile_states_test.go:314 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/reconcile_states_test.go:317 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | range | internal/journal/reconcile_states_test.go:320 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/reconcile_states_test.go:321 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `openReservationJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `j.EnterReconcile`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `j.ReleaseReconciles`: returned errors and state follow the mapped branches.
- `strings.Contains`: returned errors and state follow the mapped branches.
- `err.Error`: returned errors and state follow the mapped branches.
- `j.ActiveReconcileStates`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `state.ReleasedAt.IsZero`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
