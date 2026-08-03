# Function Logic Map: `TestPruneRemovesRecordsOlderThanTheRetention`

Source: `internal/journal/nonce_test.go`  
Function: `TestPruneRemovesRecordsOlderThanTheRetention`  
Signature: `TestPruneRemovesRecordsOlderThanTheRetention(params=1, results=0)`  
Source SHA-256: `83fcf17c3cd3758fadd4f23e7f31e675b8e3a2df7d56d3cdd6e70b583e16f5e3`

## Inputs and invariants

- Inputs are the parameters represented by `TestPruneRemovesRecordsOlderThanTheRetention(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce_test.go:234 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/nonce_test.go:241 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/nonce_test.go:244 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/nonce_test.go:250 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/nonce_test.go:253 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `openTestJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `cancelDecision`: errors and returned state remain governed by the function's explicit branches.
- `boundAttempt`: errors and returned state remain governed by the function's explicit branches.
- `a.MarkDispatchStarted`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatal`: errors and returned state remain governed by the function's explicit branches.
- `testIssued`: errors and returned state remain governed by the function's explicit branches.
- `j.PruneSpentNonces`: errors and returned state remain governed by the function's explicit branches.
- `issued.Add`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `spentNonceCount`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 8 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
