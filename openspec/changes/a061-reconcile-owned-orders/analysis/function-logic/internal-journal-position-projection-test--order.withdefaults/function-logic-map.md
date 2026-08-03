# Function Logic Map: `order.withDefaults`

Source: `internal/journal/position_projection_test.go`  
Function: `order.withDefaults`  
Signature: `order.withDefaults(params=0, results=1)`  
Source SHA-256: `e1094b972b2f61b58d5665501165349c25b2a90624b2256090185b8eda37de35`

## Inputs and invariants

- Inputs are the parameters in `order.withDefaults(params=0, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:36 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/position_projection_test.go:39 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/position_projection_test.go:42 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/position_projection_test.go:45 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/position_projection_test.go:48 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/position_projection_test.go:51 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- No outbound call; behavior is local and deterministic.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 6 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
