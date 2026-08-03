# Function Logic Map: `TestAnAmendmentKeepsOneInstance`

Source: `internal/journal/position_projection_test.go`  
Function: `TestAnAmendmentKeepsOneInstance`  
Signature: `TestAnAmendmentKeepsOneInstance(params=1, results=0)`  
Source SHA-256: `e1094b972b2f61b58d5665501165349c25b2a90624b2256090185b8eda37de35`

## Inputs and invariants

- Inputs are the parameters in `TestAnAmendmentKeepsOneInstance(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:334 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/position_projection_test.go:346 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/position_projection_test.go:349 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/position_projection_test.go:352 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/position_projection_test.go:355 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/position_projection_test.go:365 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/position_projection_test.go:368 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/position_projection_test.go:374 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/position_projection_test.go:379 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/journal/position_projection_test.go:382 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/journal/position_projection_test.go:385 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | if | internal/journal/position_projection_test.go:388 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `projectingJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `place`: returned errors and state follow the mapped branches.
- `j.RecordFill`: returned errors and state follow the mapped branches.
- `fillOf`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `j.Prepare`: returned errors and state follow the mapped branches.
- `testIntentFor`: returned errors and state follow the mapped branches.
- `amend.MarkDispatchStarted`: returned errors and state follow the mapped branches.
- `amend.MarkAcked`: returned errors and state follow the mapped branches.
- `amend.ResolveConfirmedWithLineage`: returned errors and state follow the mapped branches.
- `terminalFill`: returned errors and state follow the mapped branches.
- `currentPosition`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `j.Positions`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `t.Errorf`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 14 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
