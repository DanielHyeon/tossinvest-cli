# Function Logic Map: `Journal.ResolveCurrentOrderIDScoped`

Source: `internal/journal/lineage.go`  
Function: `Journal.ResolveCurrentOrderIDScoped`  
Signature: `Journal.ResolveCurrentOrderIDScoped(params=3, results=2)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters in `Journal.ResolveCurrentOrderIDScoped(params=3, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:313 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/lineage.go:317 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | for | internal/journal/lineage.go:323 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/lineage.go:325 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | switch | internal/journal/lineage.go:328 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | case | internal/journal/lineage.go:329 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | case | internal/journal/lineage.go:331 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | case | internal/journal/lineage.go:333 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/lineage.go:338 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `scope.canonical`: returned errors and state follow the mapped branches.
- `j.scopedLineageChildren`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `j.recordScopedLineageConflict`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 7 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
