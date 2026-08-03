# Function Logic Map: `resolveFillOrigin`

Source: `internal/journal/position_projection.go`  
Function: `resolveFillOrigin`  
Signature: `resolveFillOrigin(params=3, results=3)`  
Source SHA-256: `ae74d3ba1b66a05360e7b5851248fd6814577fa0b34068a89f52c58c10644c7b`

## Inputs and invariants

- Inputs are the parameters in `resolveFillOrigin(params=3, results=3)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection.go:141 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | for | internal/journal/position_projection.go:153 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/position_projection.go:155 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/position_projection.go:162 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/position_projection.go:167 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | range | internal/journal/position_projection.go:181 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/position_projection.go:182 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/position_projection.go:190 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/position_projection.go:193 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | range | internal/journal/position_projection.go:195 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/journal/position_projection.go:199 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | if | internal/journal/position_projection.go:203 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B13 | if | internal/journal/position_projection.go:216 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B14 | if | internal/journal/position_projection.go:225 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `tx.Query`: returned errors and state follow the mapped branches.
- `string`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `rows.Close`: returned errors and state follow the mapped branches.
- `rows.Next`: returned errors and state follow the mapped branches.
- `rows.Scan`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `rows.Err`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `firstNonEmpty`: returned errors and state follow the mapped branches.
- `normaliseMarket`: returned errors and state follow the mapped branches.
- `normaliseSymbol`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `fmt.Sprintf`: returned errors and state follow the mapped branches.
- `enterReconcileScopeInTx`: returned errors and state follow the mapped branches.
- `position.RoleForSide`: returned errors and state follow the mapped branches.
- `hasReplaceSuccessor`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 26 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
