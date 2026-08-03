# Function Logic Map: `TestSchemaTablesAndColumns`

Source: `internal/journal/schema_test.go`  
Function: `TestSchemaTablesAndColumns`  
Signature: `TestSchemaTablesAndColumns(params=1, results=0)`  
Source SHA-256: `5e56ff9da74a1775d91251b7360bbf9bddecbcdd2ee5c7f2063ab3d9213cb396`

## Inputs and invariants

- Inputs are the parameters in `TestSchemaTablesAndColumns(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/schema_test.go:123 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | for | internal/journal/schema_test.go:127 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/schema_test.go:129 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/schema_test.go:134 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/schema_test.go:138 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | range | internal/journal/schema_test.go:256 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/schema_test.go:259 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `openTestJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `j.db.QueryContext`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `rows.Next`: returned errors and state follow the mapped branches.
- `rows.Scan`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `rows.Err`: returned errors and state follow the mapped branches.
- `rows.Close`: returned errors and state follow the mapped branches.
- `strings.Join`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `tableColumns`: returned errors and state follow the mapped branches.
- `sort.Strings`: returned errors and state follow the mapped branches.
- `t.Errorf`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 9 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
