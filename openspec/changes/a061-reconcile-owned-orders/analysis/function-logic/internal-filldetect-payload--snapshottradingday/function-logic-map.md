# Function Logic Map: `snapshotTradingDay`

Source: `internal/filldetect/payload.go`  
Function: `snapshotTradingDay`  
Signature: `snapshotTradingDay(params=2, results=1)`  
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`

## Inputs and invariants

- Inputs are the parameters in `snapshotTradingDay(params=2, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload.go:146 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/filldetect/payload.go:150 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | range | internal/filldetect/payload.go:153 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/filldetect/payload.go:154 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/filldetect/payload.go:156 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `trimPtr`: returned errors and state follow the mapped branches.
- `clock.ParseMarket`: returned errors and state follow the mapped branches.
- `time.Parse`: returned errors and state follow the mapped branches.
- `market.TradingDay`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 4 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
