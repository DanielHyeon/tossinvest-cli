# Function Logic Map: `Snapshot.Digest`

Source: `internal/reconcile/snapshot.go`  
Function: `Snapshot.Digest`  
Signature: `Snapshot.Digest(params=0, results=1)`  
Source SHA-256: `827f148d49ae878bd1acb64327dbd5545cebe9a576e130305255e06861e1b8e3`

## Inputs and invariants

- Inputs are the parameters in `Snapshot.Digest(params=0, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | range | internal/reconcile/snapshot.go:169 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | range | internal/reconcile/snapshot.go:180 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | range | internal/reconcile/snapshot.go:187 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `b.WriteString`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `sort.Slice`: returned errors and state follow the mapped branches.
- `less`: returned errors and state follow the mapped branches.
- `brokerOrderIdentity`: returned errors and state follow the mapped branches.
- `fmt.Fprintf`: returned errors and state follow the mapped branches.
- `canonicalDecimal`: returned errors and state follow the mapped branches.
- `b.String`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 4 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
