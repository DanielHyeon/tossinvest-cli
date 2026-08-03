# Function Logic Map: `TestAnOperatorResolutionSurvivesARestart`

Source: `internal/reconcile/restore_test.go`  
Function: `TestAnOperatorResolutionSurvivesARestart`  
Signature: `TestAnOperatorResolutionSurvivesARestart(params=1, results=0)`  
Source SHA-256: `06361705cca4cd1d8cfd0263dff7b47ea9c661cac3c4b09bd164ba91b75c67f4`

## Inputs and invariants

- Inputs are the parameters in `TestAnOperatorResolutionSurvivesARestart(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:144 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/restore_test.go:147 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/restore_test.go:150 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/restore_test.go:157 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/reconcile/restore_test.go:160 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/reconcile/restore_test.go:163 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `context.Background`: returned errors and state follow the mapped branches.
- `clock.NewFake`: returned errors and state follow the mapped branches.
- `t.TempDir`: returned errors and state follow the mapped branches.
- `openJournalAt`: returned errors and state follow the mapped branches.
- `trackerOn`: returned errors and state follow the mapped branches.
- `noStaleGate`: returned errors and state follow the mapped branches.
- `tracker.Observe`: returned errors and state follow the mapped branches.
- `mismatchDiff`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `tracker.Resolve`: returned errors and state follow the mapped branches.
- `first.Close`: returned errors and state follow the mapped branches.
- `restarted.Restore`: returned errors and state follow the mapped branches.
- `restartedGate.CheckEntryFor`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `restarted.Blocks`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 13 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
