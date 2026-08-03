# Function Logic Map: `JournalLedger.Apply`

Source: `internal/filldetect/ledger.go`  
Function: `JournalLedger.Apply`  
Signature: `JournalLedger.Apply(params=2, results=2)`  
Source SHA-256: `59f7c08036ddcbab21f9b9e938856c39bec98b2f38c64b39ba9ce334986547db`

## Inputs and invariants

- Inputs are the parameters represented by `JournalLedger.Apply(params=2, results=2)` and any receiver state.
- Opaque broker identifiers remain byte-preserving; account, market, trading-day, symbol, and side comparisons use the canonical scope defined by the change.
- Broker mutation is outside this function-map change; reconciliation and journal ambiguity fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger.go:37 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B2 | if | internal/filldetect/ledger.go:63 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B3 | if | internal/filldetect/ledger.go:68 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B4 | if | internal/filldetect/ledger.go:71 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B5 | if | internal/filldetect/ledger.go:72 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B6 | if | internal/filldetect/ledger.go:84 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B7 | if | internal/filldetect/ledger.go:89 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |
| B8 | if | internal/filldetect/ledger.go:93 | Preserve the explicit condition, early return, and fail-closed error behavior at this branch. |

## Calls and live bindings

- `fmt.Errorf`: errors and returned state remain governed by the function's explicit branches.
- `string`: errors and returned state remain governed by the function's explicit branches.
- `decimalString`: errors and returned state remain governed by the function's explicit branches.
- `strings.TrimSpace`: errors and returned state remain governed by the function's explicit branches.
- `journal.RFC3339`: errors and returned state remain governed by the function's explicit branches.
- `l.Journal.RecordFill`: errors and returned state remain governed by the function's explicit branches.
- `l.Refresh`: errors and returned state remain governed by the function's explicit branches.
- `time.Parse`: errors and returned state remain governed by the function's explicit branches.
- `ts.UTC`: errors and returned state remain governed by the function's explicit branches.
- Runtime configuration and official-read bindings are passed by callers; this function does not broaden live-trading authority.

## State mutations and fallbacks

- 10 assignment point(s) are AST-bound; durable writes and in-memory publication occur only through the calls and successful paths listed above.
- Errors propagate without synthesizing ownership, clearing a durable block, or releasing held reservations early.
- Migration fallback accepts legacy empty scope only where the surrounding canonical evidence proves an unambiguous owner.

## Safety conclusion

The current AST, every branch ID, and the regression/full-suite verification are bound to this source hash. Ambiguity remains durable and fail-closed, while external broker observations cannot become engine-owned without positive local evidence.
