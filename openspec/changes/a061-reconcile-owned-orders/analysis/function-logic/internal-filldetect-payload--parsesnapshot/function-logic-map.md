# Function Logic Map: `parseSnapshot`

Source: `internal/filldetect/payload.go`  
Function: `parseSnapshot`  
Signature: `parseSnapshot(params=3, results=2)`  
Source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`

## Inputs and invariants

- Inputs are the parameters in `parseSnapshot(params=3, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/payload.go:73 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/filldetect/payload.go:78 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/filldetect/payload.go:87 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/filldetect/payload.go:88 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/filldetect/payload.go:95 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/filldetect/payload.go:119 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/filldetect/payload.go:123 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `unwrapResult`: returned errors and state follow the mapped branches.
- `json.Unmarshal`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `parseDecimal`: returned errors and state follow the mapped branches.
- `trimPtr`: returned errors and state follow the mapped branches.
- `parseTime`: returned errors and state follow the mapped branches.
- `fillMarketFromCurrency`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `snapshotTradingDay`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 15 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
