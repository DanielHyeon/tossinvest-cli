# Function Logic Map: `reserveConfirmedSweepOrder`

Source: `internal/journal/reservation_sweep_test.go`  
Function: `reserveConfirmedSweepOrder`  
Signature: `reserveConfirmedSweepOrder(params=6, results=1)`  
Source SHA-256: `489ced650b9f96b30419ecc9c21f7911145af1a973372a7f167edb61cfcbdcab`

## Inputs and invariants

- Inputs are the parameters represented by `reserveConfirmedSweepOrder(params=6, results=1)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_sweep_test.go:136 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `t.Helper`: errors and returned state remain governed by the function's explicit branches.
- `recordEntryDecision`: errors and returned state remain governed by the function's explicit branches.
- `j.Reserve`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `exposureReserve`: errors and returned state remain governed by the function's explicit branches.
- `mustVersion`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `confirmAttempt`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 2 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
