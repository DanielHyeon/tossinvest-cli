# Function Logic Map: `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`

Source: `internal/app/engine/reconcileloop_test.go`  
Function: `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`  
Signature: `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead(params=1, results=0)`  
Source SHA-256: `feb0b59737a7c47e4ead572b77c9f2b591273fa6bd61744850a60c87830d6342`

## Inputs and invariants

- Inputs are the parameters in `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop_test.go:382 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/app/engine/reconcileloop_test.go:389 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/app/engine/reconcileloop_test.go:392 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/app/engine/reconcileloop_test.go:395 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `newDriverHarness`: returned errors and state follow the mapped branches.
- `h.holds`: returned errors and state follow the mapped branches.
- `h.journal.EnterReconcile`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `h.cycle`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 3 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
