# Function Logic Map: `ReconcileDriver.RunOnce`

Source: `internal/app/engine/reconcileloop.go`  
Function: `ReconcileDriver.RunOnce`  
Signature: `ReconcileDriver.RunOnce(params=1, results=1)`  
Source SHA-256: `accaa4c5f6645d8af7be3f1cbcd9ec61a7efc9f1f022be26b39b53789d867763`

## Inputs and invariants

- Inputs are the parameters in `ReconcileDriver.RunOnce(params=1, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop.go:387 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/app/engine/reconcileloop.go:393 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/app/engine/reconcileloop.go:404 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/app/engine/reconcileloop.go:410 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/app/engine/reconcileloop.go:416 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/app/engine/reconcileloop.go:417 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/app/engine/reconcileloop.go:425 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/app/engine/reconcileloop.go:426 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `d.note`: returned errors and state follow the mapped branches.
- `d.stabilise`: returned errors and state follow the mapped branches.
- `reconcile.LocalStateFromJournal`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `Compare`: returned errors and state follow the mapped branches.
- `d.ingest.IngestExternalPositions`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `d.opts.Converge.ConvergeQuantities`: returned errors and state follow the mapped branches.
- `d.opts.Tracker.Refresh`: returned errors and state follow the mapped branches.
- `d.opts.Tracker.Observe`: returned errors and state follow the mapped branches.
- `d.opts.Tracker.Blocks`: returned errors and state follow the mapped branches.
- `d.judgeHoldings`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 18 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
