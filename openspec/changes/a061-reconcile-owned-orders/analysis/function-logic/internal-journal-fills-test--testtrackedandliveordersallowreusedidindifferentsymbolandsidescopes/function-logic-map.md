# Function Logic Map: `TestTrackedAndLiveOrdersAllowReusedIDInDifferentSymbolAndSideScopes`

Source: `internal/journal/fills_test.go`  
Function: `TestTrackedAndLiveOrdersAllowReusedIDInDifferentSymbolAndSideScopes`  
Signature: `TestTrackedAndLiveOrdersAllowReusedIDInDifferentSymbolAndSideScopes(params=1, results=0)`  
Source SHA-256: `e322e6a62817b22a0ed66fb2c17067e2d8707c87e0ae69c648fa3bfc7c766c56`

## Inputs and invariants

- Inputs are the parameters in `TestTrackedAndLiveOrdersAllowReusedIDInDifferentSymbolAndSideScopes(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills_test.go:763 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/fills_test.go:766 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/fills_test.go:769 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/fills_test.go:772 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/fills_test.go:779 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/fills_test.go:784 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/fills_test.go:789 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/fills_test.go:792 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | range | internal/journal/fills_test.go:795 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/journal/fills_test.go:797 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/journal/fills_test.go:800 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `openTestJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `j.Prepare`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `attempt.MarkDispatchStarted`: returned errors and state follow the mapped branches.
- `attempt.MarkAcked`: returned errors and state follow the mapped branches.
- `attempt.Settle`: returned errors and state follow the mapped branches.
- `confirm`: returned errors and state follow the mapped branches.
- `observation`: returned errors and state follow the mapped branches.
- `j.RecordFill`: returned errors and state follow the mapped branches.
- `j.TrackedFillOrders`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `j.LiveOrdersForSymbol`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 14 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
