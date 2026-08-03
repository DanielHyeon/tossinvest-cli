# Function Logic Map: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld`

Source: `internal/journal/nonce_test.go`  
Function: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld`  
Signature: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld(params=1, results=0)`  
Source SHA-256: `84f12358991f53b64e3f9dbebef2730533e6375e8ea55289c80b9cdfeb35487f`

## Inputs and invariants

- Inputs are the parameters in `TestPrunePreservesSpentNonceWhileItsReservationIsHeld(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/nonce_test.go:263 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/nonce_test.go:271 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/nonce_test.go:274 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `openReservationJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `recordEntryDecision`: returned errors and state follow the mapped branches.
- `j.Reserve`: returned errors and state follow the mapped branches.
- `exposureReserve`: returned errors and state follow the mapped branches.
- `mustVersion`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `spendNonce`: returned errors and state follow the mapped branches.
- `fake.Advance`: returned errors and state follow the mapped branches.
- `j.PruneSpentNonces`: returned errors and state follow the mapped branches.
- `fake.Now`: returned errors and state follow the mapped branches.
- `spentNonceCount`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 5 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
