# Function Logic Map: `Tracker.syncGate`

Source: `internal/reconcile/mismatch.go`  
Function: `Tracker.syncGate`  
Signature: `Tracker.syncGate(params=1, results=0)`  
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`

## Inputs and invariants

- Inputs are the parameters in `Tracker.syncGate(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:879 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | range | internal/reconcile/mismatch.go:883 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/mismatch.go:884 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | range | internal/reconcile/mismatch.go:894 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/reconcile/mismatch.go:895 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | range | internal/reconcile/mismatch.go:901 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/reconcile/mismatch.go:902 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/reconcile/mismatch.go:908 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | range | internal/reconcile/mismatch.go:913 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/reconcile/mismatch.go:917 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | else | internal/reconcile/mismatch.go:919 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `string`: returned errors and state follow the mapped branches.
- `t.Gate.BlockSymbol`: returned errors and state follow the mapped branches.
- `t.Gate.SymbolBlocks`: returned errors and state follow the mapped branches.
- `isReconcileReason`: returned errors and state follow the mapped branches.
- `t.Gate.ClearSymbol`: returned errors and state follow the mapped branches.
- `t.Gate.Block`: returned errors and state follow the mapped branches.
- `t.Gate.Clear`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
