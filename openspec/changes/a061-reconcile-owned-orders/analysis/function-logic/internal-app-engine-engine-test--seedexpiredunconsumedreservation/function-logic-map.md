# Function Logic Map: `seedExpiredUnconsumedReservation`

Source: `internal/app/engine/engine_test.go`  
Function: `seedExpiredUnconsumedReservation`  
Signature: `seedExpiredUnconsumedReservation(params=2, results=1)`  
Source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`

## Inputs and invariants

- Inputs are the parameters in `seedExpiredUnconsumedReservation(params=2, results=1)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine_test.go:495 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/app/engine/engine_test.go:511 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/app/engine/engine_test.go:515 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/app/engine/engine_test.go:519 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `t.Helper`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `Add`: returned errors and state follow the mapped branches.
- `UTC`: returned errors and state follow the mapped branches.
- `time.Now`: returned errors and state follow the mapped branches.
- `journal.Open`: returned errors and state follow the mapped branches.
- `filepath.Join`: returned errors and state follow the mapped branches.
- `clock.NewFake`: returned errors and state follow the mapped branches.
- `journal.FixedFSProber`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `j.Close`: returned errors and state follow the mapped branches.
- `j.RecordDecision`: returned errors and state follow the mapped branches.
- `issued.Add`: returned errors and state follow the mapped branches.
- `j.ReservationVersion`: returned errors and state follow the mapped branches.
- `j.Reserve`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 7 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
