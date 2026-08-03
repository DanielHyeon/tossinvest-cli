# Function Logic Map: `brokerPayloadTradingDay`

Source: `internal/reconcile/compare.go`  
Function: `brokerPayloadTradingDay`  
Signature: `brokerPayloadTradingDay(params=1, results=1)`  
Source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`

## Inputs and invariants

- Inputs are the parameters in `brokerPayloadTradingDay(params=1, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:665 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/compare.go:669 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/compare.go:673 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/reconcile/compare.go:675 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/reconcile/compare.go:676 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/reconcile/compare.go:684 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `time.Parse`: returned errors and state follow the mapped branches.
- `marketclock.ParseMarket`: returned errors and state follow the mapped branches.
- `market.TradingDay`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
