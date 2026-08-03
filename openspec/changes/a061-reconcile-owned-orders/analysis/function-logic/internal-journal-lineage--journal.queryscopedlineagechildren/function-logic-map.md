# Function Logic Map: `Journal.queryScopedLineageChildren`

Source: `internal/journal/lineage.go`  
Function: `Journal.queryScopedLineageChildren`  
Signature: `Journal.queryScopedLineageChildren(params=5, results=2)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters in `Journal.queryScopedLineageChildren(params=5, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:423 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | for | internal/journal/lineage.go:429 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/lineage.go:431 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/lineage.go:436 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `j.db.QueryContext`: returned errors and state follow the mapped branches.
- `string`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `rows.Close`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `rows.Next`: returned errors and state follow the mapped branches.
- `rows.Scan`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `rows.Err`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
