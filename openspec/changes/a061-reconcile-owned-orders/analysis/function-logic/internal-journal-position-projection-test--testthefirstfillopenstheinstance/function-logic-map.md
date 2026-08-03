# Function Logic Map: `TestTheFirstFillOpensTheInstance`

Source: `internal/journal/position_projection_test.go`  
Function: `TestTheFirstFillOpensTheInstance`  
Signature: `TestTheFirstFillOpensTheInstance(params=1, results=0)`  
Source SHA-256: `6ab3463bdc484584a3e1dc23b86cabc42fa737122966e7ed57b96ec78bd1572f`

## Inputs and invariants

- Inputs are the parameters represented by `TestTheFirstFillOpensTheInstance(params=1, results=0)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:132 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/journal/position_projection_test.go:136 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/journal/position_projection_test.go:141 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/journal/position_projection_test.go:144 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/journal/position_projection_test.go:147 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/journal/position_projection_test.go:150 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/journal/position_projection_test.go:153 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/journal/position_projection_test.go:156 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B9 | if | internal/journal/position_projection_test.go:159 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B10 | if | internal/journal/position_projection_test.go:162 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `projectingJournal`: errors and returned state remain governed by the function's explicit branches.
- `context.Background`: errors and returned state remain governed by the function's explicit branches.
- `place`: errors and returned state remain governed by the function's explicit branches.
- `j.CurrentPosition`: errors and returned state remain governed by the function's explicit branches.
- `errors.Is`: errors and returned state remain governed by the function's explicit branches.
- `t.Fatalf`: errors and returned state remain governed by the function's explicit branches.
- `j.RecordFill`: errors and returned state remain governed by the function's explicit branches.
- `fillOf`: errors and returned state remain governed by the function's explicit branches.
- `currentPosition`: errors and returned state remain governed by the function's explicit branches.
- `t.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `t.Error`: errors and returned state remain governed by the function's explicit branches.
- `PositionID`: errors and returned state remain governed by the function's explicit branches.
- `p.ExitEligible`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 6 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
