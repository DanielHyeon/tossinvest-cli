# Function Logic Map: `TestSchemaIndexes`

Source: `internal/journal/schema_test.go`  
Function: `TestSchemaIndexes`  
Signature: `TestSchemaIndexes(params=1, results=0)`  
Source SHA-256: `9d6a1364f76a4dc3c6e9854066182ca18e59be53458d64e42f13a9e74df95d7b`

## Inputs and invariants

- Inputs are the parameters represented by `TestSchemaIndexes(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/schema_test.go:507 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | for | internal/journal/schema_test.go:512 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/schema_test.go:514 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/schema_test.go:519 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | range | internal/journal/schema_test.go:522 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/schema_test.go:551 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openTestJournal`: errors and returned state remain governed by the function's explicit branches.
- `j.db.QueryContext`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `rows.Close`: errors and returned state remain governed by the function's explicit branches.
- `rows.Next`: errors and returned state remain governed by the function's explicit branches.
- `rows.Scan`: errors and returned state remain governed by the function's explicit branches.
- `rows.Err`: errors and returned state remain governed by the function's explicit branches.
- `t.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `keysOf`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 6 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
