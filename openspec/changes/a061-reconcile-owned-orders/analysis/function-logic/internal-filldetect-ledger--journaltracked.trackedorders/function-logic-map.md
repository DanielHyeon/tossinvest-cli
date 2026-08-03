# Function Logic Map: `JournalTracked.TrackedOrders`

Source: `internal/filldetect/ledger.go`  
Function: `JournalTracked.TrackedOrders`  
Signature: `JournalTracked.TrackedOrders(params=1, results=2)`  
Source SHA-256: `75966bf1507d412f48b88efddab42f045fff31f45c76bc9cf43dfc0a634242c4`

## Inputs and invariants

- Inputs are the parameters in `JournalTracked.TrackedOrders(params=1, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger.go:126 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/filldetect/ledger.go:129 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/filldetect/ledger.go:133 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | range | internal/filldetect/ledger.go:137 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `t.Journal.TrackedFillOrders`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 3 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
