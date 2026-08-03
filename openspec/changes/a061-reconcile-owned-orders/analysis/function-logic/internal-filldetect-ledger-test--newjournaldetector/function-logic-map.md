# Function Logic Map: `newJournalDetector`

Source: `internal/filldetect/ledger_test.go`  
Function: `newJournalDetector`  
Signature: `newJournalDetector(params=5, results=2)`  
Source SHA-256: `3fa79a4b73613bae1f9ee086608012e2dca32a6f85b5b9128b916016fdfe571a`

## Inputs and invariants

- Inputs are the parameters represented by `newJournalDetector(params=5, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/filldetect/ledger_test.go:38 | Execute the function contract without an alternate branch. |

## Calls and live bindings

- `t.Helper`: errors and returned state remain governed by the function's explicit branches.
- `openLedgerJournal`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 1 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
