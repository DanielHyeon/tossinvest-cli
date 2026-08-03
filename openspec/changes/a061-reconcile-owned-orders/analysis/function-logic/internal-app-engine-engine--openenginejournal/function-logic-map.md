# Function Logic Map: `openEngineJournal`

Source: `internal/app/engine/engine.go`  
Function: `openEngineJournal`  
Signature: `openEngineJournal(params=3, results=2)`  
Source SHA-256: `401ab52518aac369f7567a60f711c4a019efad96ab0bcd7af1751155ba67e1f5`

## Inputs and invariants

- Inputs are the parameters in `openEngineJournal(params=3, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/engine.go:590 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/app/engine/engine.go:598 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/app/engine/engine.go:606 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/app/engine/engine.go:622 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/app/engine/engine.go:625 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `filepath.Join`: returned errors and state follow the mapped branches.
- `journal.Open`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `j.SweepReservations`: returned errors and state follow the mapped branches.
- `j.Close`: returned errors and state follow the mapped branches.
- `j.MaxDecisionTTL`: returned errors and state follow the mapped branches.
- `j.PruneSpentNonces`: returned errors and state follow the mapped branches.
- `clk.Now`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 10 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
