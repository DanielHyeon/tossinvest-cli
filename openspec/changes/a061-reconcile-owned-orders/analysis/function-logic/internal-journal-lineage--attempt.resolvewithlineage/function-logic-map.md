# Function Logic Map: `Attempt.resolveWithLineage`

Source: `internal/journal/lineage.go`  
Function: `Attempt.resolveWithLineage`  
Signature: `Attempt.resolveWithLineage(params=4, results=1)`  
Source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

## Inputs and invariants

- Inputs are the parameters in `Attempt.resolveWithLineage(params=4, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/lineage.go:133 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/lineage.go:139 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/lineage.go:141 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/lineage.go:146 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/lineage.go:162 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/lineage.go:166 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/lineage.go:169 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/lineage.go:173 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/lineage.go:180 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/journal/lineage.go:216 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/journal/lineage.go:221 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | if | internal/journal/lineage.go:225 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B13 | if | internal/journal/lineage.go:230 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `a.mu.Lock`: returned errors and state follow the mapped branches.
- `a.mu.Unlock`: returned errors and state follow the mapped branches.
- `a.j.nowString`: returned errors and state follow the mapped branches.
- `a.j.db.BeginTx`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `tx.Rollback`: returned errors and state follow the mapped branches.
- `Scan`: returned errors and state follow the mapped branches.
- `tx.QueryRowContext`: returned errors and state follow the mapped branches.
- `errors.Is`: returned errors and state follow the mapped branches.
- `checkTransitionAllowed`: returned errors and state follow the mapped branches.
- `tx.ExecContext`: returned errors and state follow the mapped branches.
- `string`: returned errors and state follow the mapped branches.
- `res.RowsAffected`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `scoped.RowsAffected`: returned errors and state follow the mapped branches.
- `tx.Commit`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 13 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
