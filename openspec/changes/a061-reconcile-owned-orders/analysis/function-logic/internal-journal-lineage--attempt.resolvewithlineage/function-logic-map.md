# Function Logic Map: `Attempt.resolveWithLineage`

Source: `internal/journal/lineage.go`  
Function: `Attempt.resolveWithLineage`  
Signature: `Attempt.resolveWithLineage(params=4, results=1)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters represented by `Attempt.resolveWithLineage(params=4, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:133 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/lineage.go:139 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/lineage.go:141 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/lineage.go:146 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/lineage.go:162 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/lineage.go:166 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/lineage.go:169 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/lineage.go:173 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/lineage.go:180 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/lineage.go:216 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B11 | if | internal/journal/lineage.go:221 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B12 | if | internal/journal/lineage.go:225 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B13 | if | internal/journal/lineage.go:230 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `a.mu.Lock`: errors and returned state remain governed by the function's explicit branches.
- `a.mu.Unlock`: errors and returned state remain governed by the function's explicit branches.
- `a.j.nowString`: errors and returned state remain governed by the function's explicit branches.
- `a.j.db.BeginTx`: errors and returned state remain governed by the function's explicit branches.
- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `tx.Rollback`: errors and returned state remain governed by the function's explicit branches.
- `Scan`: errors and returned state remain governed by the function's explicit branches.
- `tx.QueryRowContext`: errors and returned state remain governed by the function's explicit branches.
- `errors.Is`: errors and returned state remain governed by the function's explicit branches.
- `checkTransitionAllowed`: errors and returned state remain governed by the function's explicit branches.
- `tx.ExecContext`: errors and returned state remain governed by the function's explicit branches.
- `string`: errors and returned state remain governed by the function's explicit branches.
- `res.RowsAffected`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `scoped.RowsAffected`: errors and returned state remain governed by the function's explicit branches.
- `tx.Commit`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 13 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
