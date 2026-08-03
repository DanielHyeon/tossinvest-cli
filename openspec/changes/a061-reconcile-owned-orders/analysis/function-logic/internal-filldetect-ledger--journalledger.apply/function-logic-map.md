# Function Logic Map: `JournalLedger.Apply`

Source: `internal/filldetect/ledger.go`  
Function: `JournalLedger.Apply`  
Signature: `JournalLedger.Apply(params=2, results=2)`  
Source SHA-256: `75966bf1507d412f48b88efddab42f045fff31f45c76bc9cf43dfc0a634242c4`

## Inputs and invariants

- Inputs are the parameters in `JournalLedger.Apply(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger.go:37 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/filldetect/ledger.go:42 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/filldetect/ledger.go:45 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/filldetect/ledger.go:48 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/filldetect/ledger.go:76 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/filldetect/ledger.go:81 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/filldetect/ledger.go:84 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/filldetect/ledger.go:85 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/filldetect/ledger.go:97 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/filldetect/ledger.go:102 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/filldetect/ledger.go:106 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `string`: returned errors and state follow the mapped branches.
- `decimalString`: returned errors and state follow the mapped branches.
- `journal.RFC3339`: returned errors and state follow the mapped branches.
- `l.Journal.RecordFill`: returned errors and state follow the mapped branches.
- `l.Refresh`: returned errors and state follow the mapped branches.
- `time.Parse`: returned errors and state follow the mapped branches.
- `ts.UTC`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 12 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
