# Function Logic Map: `TestTheFirstFillOpensTheInstance`

Source: `internal/journal/position_projection_test.go`  
Function: `TestTheFirstFillOpensTheInstance`  
Signature: `TestTheFirstFillOpensTheInstance(params=1, results=0)`  
Source SHA-256: `6ab3463bdc484584a3e1dc23b86cabc42fa737122966e7ed57b96ec78bd1572f`

## Inputs and invariants

- Inputs are the parameters in `TestTheFirstFillOpensTheInstance(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:132 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/position_projection_test.go:136 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/position_projection_test.go:141 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/position_projection_test.go:144 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/position_projection_test.go:147 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/position_projection_test.go:150 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/position_projection_test.go:153 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/position_projection_test.go:156 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/position_projection_test.go:159 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/journal/position_projection_test.go:162 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `projectingJournal`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `place`: returned errors and state follow the mapped branches.
- `j.CurrentPosition`: returned errors and state follow the mapped branches.
- `errors.Is`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `j.RecordFill`: returned errors and state follow the mapped branches.
- `fillOf`: returned errors and state follow the mapped branches.
- `currentPosition`: returned errors and state follow the mapped branches.
- `t.Errorf`: returned errors and state follow the mapped branches.
- `t.Error`: returned errors and state follow the mapped branches.
- `PositionID`: returned errors and state follow the mapped branches.
- `p.ExitEligible`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 6 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
