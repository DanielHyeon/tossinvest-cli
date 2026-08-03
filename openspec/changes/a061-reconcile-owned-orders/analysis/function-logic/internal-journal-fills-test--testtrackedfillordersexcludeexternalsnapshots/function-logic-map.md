# Function Logic Map: `TestTrackedFillOrdersExcludeExternalSnapshots`

Source: `internal/journal/fills_test.go`  
Function: `TestTrackedFillOrdersExcludeExternalSnapshots`  
Signature: `TestTrackedFillOrdersExcludeExternalSnapshots(params=1, results=0)`  
Source SHA-256: `e322e6a62817b22a0ed66fb2c17067e2d8707c87e0ae69c648fa3bfc7c766c56`

## Inputs and invariants

- Inputs are the parameters in `TestTrackedFillOrdersExcludeExternalSnapshots(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills_test.go:415 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/fills_test.go:418 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/fills_test.go:423 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/fills_test.go:426 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `openTestJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `j.RecordFill`: returned errors and state follow the mapped branches.
- `observation`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `j.LookupFill`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `j.TrackedFillOrders`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
