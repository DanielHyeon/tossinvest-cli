# Function Logic Map: `OrderLineageScope.canonical`

Source: `internal/journal/lineage.go`  
Function: `OrderLineageScope.canonical`  
Signature: `OrderLineageScope.canonical(params=0, results=2)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters in `OrderLineageScope.canonical(params=0, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | switch | internal/journal/lineage.go:72 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | case | internal/journal/lineage.go:73 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | case | internal/journal/lineage.go:75 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | case | internal/journal/lineage.go:77 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | case | internal/journal/lineage.go:79 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | case | internal/journal/lineage.go:81 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `normaliseMarket`: returned errors and state follow the mapped branches.
- `normaliseSymbol`: returned errors and state follow the mapped branches.
- `strings.ToUpper`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
