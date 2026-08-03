# Function Logic Map: `sweepOrphanedTerminals`

Source: `internal/journal/reservation_release.go`  
Function: `sweepOrphanedTerminals`  
Signature: `sweepOrphanedTerminals(params=3, results=2)`  
Source SHA-256: `7e7ef60ba4a8325d9a2bfca513828195ff7741058af4ab4a9dc6d2d843334718`

## Inputs and invariants

- Inputs are the parameters represented by `sweepOrphanedTerminals(params=3, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_release.go:518 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/reservation_release.go:559 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | for | internal/journal/reservation_release.go:563 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/reservation_release.go:565 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/reservation_release.go:573 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | range | internal/journal/reservation_release.go:582 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/reservation_release.go:583 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/reservation_release.go:588 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/reservation_release.go:597 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `releaseWhere`: errors and returned state remain governed by the function's explicit branches.
- `string`: errors and returned state remain governed by the function's explicit branches.
- `tx.QueryContext`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `rows.Next`: errors and returned state remain governed by the function's explicit branches.
- `rows.Scan`: errors and returned state remain governed by the function's explicit branches.
- `rows.Close`: errors and returned state remain governed by the function's explicit branches.
- `append`: errors and returned state remain governed by the function's explicit branches.
- `rows.Err`: errors and returned state remain governed by the function's explicit branches.
- `applyTx.invalidate`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Sprintf`: errors and returned state remain governed by the function's explicit branches.
- `enterReconcileScopeInTx`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 10 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
