# Function Logic Map: `Journal.RecordFill`

Source: `internal/journal/fills.go`  
Function: `Journal.RecordFill`  
Signature: `Journal.RecordFill(params=2, results=2)`  
Source SHA-256: `1a9973b325d8be62dd5d0cdebe10988ac90c6e2114d5f2e1f0b545482b141a65`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.RecordFill(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/fills.go:259 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/fills.go:263 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/fills.go:267 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/fills.go:279 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | switch | internal/journal/fills.go:285 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | case | internal/journal/fills.go:286 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | case | internal/journal/fills.go:287 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | case | internal/journal/fills.go:289 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/fills.go:293 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/fills.go:299 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | if | internal/journal/fills.go:304 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B12 | if | internal/journal/fills.go:311 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B13 | if | internal/journal/fills.go:320 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B14 | if | internal/journal/fills.go:324 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B15 | if | internal/journal/fills.go:332 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B16 | if | internal/journal/fills.go:339 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B17 | if | internal/journal/fills.go:340 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B18 | if | internal/journal/fills.go:356 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B19 | if | internal/journal/fills.go:379 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B20 | if | internal/journal/fills.go:380 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B21 | if | internal/journal/fills.go:395 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B22 | if | internal/journal/fills.go:396 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B23 | if | internal/journal/fills.go:434 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B24 | if | internal/journal/fills.go:442 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B25 | if | internal/journal/fills.go:446 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B26 | if | internal/journal/fills.go:458 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B27 | if | internal/journal/fills.go:459 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B28 | if | internal/journal/fills.go:464 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `strconv.ParseFloat`: errors and returned state remain governed by the function's explicit branches.
- `orZero`: errors and returned state remain governed by the function's explicit branches.
- `math.IsNaN`: errors and returned state remain governed by the function's explicit branches.
- `math.IsInf`: errors and returned state remain governed by the function's explicit branches.
- `j.nowString`: errors and returned state remain governed by the function's explicit branches.
- `j.db.BeginTx`: errors and returned state remain governed by the function's explicit branches.
- `tx.Rollback`: errors and returned state remain governed by the function's explicit branches.
- `scanFillSnapshot`: errors and returned state remain governed by the function's explicit branches.
- `tx.QueryRowContext`: errors and returned state remain governed by the function's explicit branches.
- `errors.Is`: errors and returned state remain governed by the function's explicit branches.
- `fillSnapshotScopeChanged`: errors and returned state remain governed by the function's explicit branches.
- `classifyFillRefusal`: errors and returned state remain governed by the function's explicit branches.
- `markFillRefused`: errors and returned state remain governed by the function's explicit branches.
- `alertsForOrder`: errors and returned state remain governed by the function's explicit branches.
- `tx.Commit`: errors and returned state remain governed by the function's explicit branches.
- `nearlyZero`: errors and returned state remain governed by the function's explicit branches.
- `sameSnapshot`: errors and returned state remain governed by the function's explicit branches.
- `tx.ExecContext`: errors and returned state remain governed by the function's explicit branches.
- `strings.ToUpper`: errors and returned state remain governed by the function's explicit branches.
- `boolToInt`: errors and returned state remain governed by the function's explicit branches.
- `recordExecutionCorrection`: errors and returned state remain governed by the function's explicit branches.
- `firstNonEmpty`: errors and returned state remain governed by the function's explicit branches.
- `decimalString`: errors and returned state remain governed by the function's explicit branches.
- `resolveFillOrigin`: errors and returned state remain governed by the function's explicit branches.
- `ownershipHandle.invalidate`: errors and returned state remain governed by the function's explicit branches.
- `releaseReservationsForOrder`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Sprintf`: errors and returned state remain governed by the function's explicit branches.
- `j.runApplyHooks`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 40 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
