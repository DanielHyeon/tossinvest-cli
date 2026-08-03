# Function Logic Map: `Journal.ResolveCurrentOrderID`

Source: `internal/journal/lineage.go`  
Function: `Journal.ResolveCurrentOrderID`  
Signature: `Journal.ResolveCurrentOrderID(params=2, results=2)`  
Source SHA-256: `bf26e9cfd6030033e99ec6ee2ceb53dd5843a0c4c25111e8761844c598bb9d73`

## Inputs and invariants

- Inputs are the parameters in `Journal.ResolveCurrentOrderID(params=2, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:200 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | for | internal/journal/lineage.go:204 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/lineage.go:206 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | range | internal/journal/lineage.go:210 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/lineage.go:211 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/lineage.go:216 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/lineage.go:219 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `j.LineageChildren`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 6 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
