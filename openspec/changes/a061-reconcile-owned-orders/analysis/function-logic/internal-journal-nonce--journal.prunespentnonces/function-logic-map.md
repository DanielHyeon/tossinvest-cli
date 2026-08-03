# Function Logic Map: `Journal.PruneSpentNonces`

Source: `internal/journal/nonce.go`  
Function: `Journal.PruneSpentNonces`  
Signature: `Journal.PruneSpentNonces(params=3, results=2)`  
Source SHA-256: `1466fddb8d43a5481cdc10f06b53c09a340862ab91907b7ccc70e40d35b7959c`

## Inputs and invariants

- Inputs are the parameters represented by `Journal.PruneSpentNonces(params=3, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce.go:151 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/nonce.go:155 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/nonce.go:158 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/nonce.go:172 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/nonce.go:176 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `j.MaxDecisionTTL`: errors and returned state remain governed by the function's explicit branches.
- `formatJournalTime`: errors and returned state remain governed by the function's explicit branches.
- `now.Add`: errors and returned state remain governed by the function's explicit branches.
- `j.db.ExecContext`: errors and returned state remain governed by the function's explicit branches.
- `res.RowsAffected`: errors and returned state remain governed by the function's explicit branches.
- `int`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 4 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
