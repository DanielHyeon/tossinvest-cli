# Function Logic Map: `TestAFailingApplyHookRollsBackTheFill`

Source: `internal/journal/apply_hook_test.go`  
Function: `TestAFailingApplyHookRollsBackTheFill`  
Signature: `TestAFailingApplyHookRollsBackTheFill(params=1, results=0)`  
Source SHA-256: `26d73b9371960a62335c0be0eef4750f398ad5099dca7b88222de4a126e09ccd`

## Inputs and invariants

- Inputs are the parameters in `TestAFailingApplyHookRollsBackTheFill(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/apply_hook_test.go:151 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/journal/apply_hook_test.go:154 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/journal/apply_hook_test.go:169 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | if | internal/journal/apply_hook_test.go:175 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/journal/apply_hook_test.go:179 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/journal/apply_hook_test.go:182 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | if | internal/journal/apply_hook_test.go:187 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/journal/apply_hook_test.go:191 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/journal/apply_hook_test.go:197 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | if | internal/journal/apply_hook_test.go:200 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `applyHookFixture`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `errors.New`: returned errors and state follow the mapped branches.
- `j.SetApplyHooks`: returned errors and state follow the mapped branches.
- `tx.Exec`: returned errors and state follow the mapped branches.
- `t.Error`: returned errors and state follow the mapped branches.
- `t.Fatal`: returned errors and state follow the mapped branches.
- `j.RecordFill`: returned errors and state follow the mapped branches.
- `observation`: returned errors and state follow the mapped branches.
- `errors.Is`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- `j.LookupFill`: returned errors and state follow the mapped branches.
- `t.Errorf`: returned errors and state follow the mapped branches.
- `j.FillEvents`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `Scan`: returned errors and state follow the mapped branches.
- `j.db.QueryRowContext`: returned errors and state follow the mapped branches.
- `j.TrackedFillOrders`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 10 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
