# Function Logic Map: `Journal.PruneSpentNonces`

Source: `internal/journal/nonce.go`  
Function: `Journal.PruneSpentNonces`  
Signature: `Journal.PruneSpentNonces(params=3, results=2)`  
Source SHA-256: `1466fddb8d43a5481cdc10f06b53c09a340862ab91907b7ccc70e40d35b7959c`

## Inputs and invariants

- Inputs are the parameters in `Journal.PruneSpentNonces(params=3, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce.go:151 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/nonce.go:155 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/nonce.go:158 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/nonce.go:172 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/nonce.go:176 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `j.MaxDecisionTTL`: returned errors and state follow the mapped branches.
- `formatJournalTime`: returned errors and state follow the mapped branches.
- `now.Add`: returned errors and state follow the mapped branches.
- `j.db.ExecContext`: returned errors and state follow the mapped branches.
- `res.RowsAffected`: returned errors and state follow the mapped branches.
- `int`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 4 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
