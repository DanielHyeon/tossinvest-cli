# Function Logic Map: `sweepOrphanedTerminals`

Source: `internal/journal/reservation_release.go`  
Function: `sweepOrphanedTerminals`  
Signature: `sweepOrphanedTerminals(params=3, results=2)`  
Source SHA-256: `7e7ef60ba4a8325d9a2bfca513828195ff7741058af4ab4a9dc6d2d843334718`

## Inputs and invariants

- Inputs are the parameters in `sweepOrphanedTerminals(params=3, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/reservation_release.go:518 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/reservation_release.go:559 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | for | internal/journal/reservation_release.go:563 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/reservation_release.go:565 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/reservation_release.go:573 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | range | internal/journal/reservation_release.go:582 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/reservation_release.go:583 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/reservation_release.go:588 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/reservation_release.go:597 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `releaseWhere`: returned errors and state follow the mapped branches.
- `string`: returned errors and state follow the mapped branches.
- `tx.QueryContext`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `rows.Next`: returned errors and state follow the mapped branches.
- `rows.Scan`: returned errors and state follow the mapped branches.
- `rows.Close`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `rows.Err`: returned errors and state follow the mapped branches.
- `applyTx.invalidate`: returned errors and state follow the mapped branches.
- `fmt.Sprintf`: returned errors and state follow the mapped branches.
- `enterReconcileScopeInTx`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 10 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
