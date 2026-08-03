# Function Logic Map: `fillMarketFromCurrency`

Source: `internal/filldetect/payload.go`  
Function: `fillMarketFromCurrency`  
Signature: `fillMarketFromCurrency(params=1, results=1)`  
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`

## Inputs and invariants

- Inputs are the parameters in `fillMarketFromCurrency(params=1, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/filldetect/payload.go:134 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | case | internal/filldetect/payload.go:135 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | case | internal/filldetect/payload.go:137 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | case | internal/filldetect/payload.go:139 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 0 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
