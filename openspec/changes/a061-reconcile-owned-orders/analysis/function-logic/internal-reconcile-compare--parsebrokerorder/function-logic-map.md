# Function Logic Map: `parseBrokerOrder`

Source: `internal/reconcile/compare.go`  
Function: `parseBrokerOrder`  
Signature: `parseBrokerOrder(params=1, results=2)`  
Source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`

## Inputs and invariants

- Inputs are the parameters in `parseBrokerOrder(params=1, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/compare.go:639 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/reconcile/compare.go:642 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/reconcile/compare.go:658 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `json.Unmarshal`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `strings.ToLower`: returned errors and state follow the mapped branches.
- `brokerPayloadTradingDay`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `derefDecimal`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
