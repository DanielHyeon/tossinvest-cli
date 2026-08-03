# Function Logic Map: `TestAFailingApplyHookRollsBackTheFill`

Source: `internal/journal/apply_hook_test.go`
Function: `TestAFailingApplyHookRollsBackTheFill`
Signature: `TestAFailingApplyHookRollsBackTheFill(params=1, results=0)`
Source SHA-256: `26d73b9371960a62335c0be0eef4750f398ad5099dca7b88222de4a126e09ccd`
Revision: `current`

## Inputs and invariants

- Inputs are `TestAFailingApplyHookRollsBackTheFill(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/apply_hook_test.go:151 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/apply_hook_test.go:154 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/apply_hook_test.go:169 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/apply_hook_test.go:175 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/apply_hook_test.go:179 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/apply_hook_test.go:182 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/apply_hook_test.go:187 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/apply_hook_test.go:191 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/apply_hook_test.go:197 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/apply_hook_test.go:200 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `applyHookFixture`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `errors.New`: errors and state follow mapped branches.
- `j.SetApplyHooks`: errors and state follow mapped branches.
- `tx.Exec`: errors and state follow mapped branches.
- `t.Error`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `j.RecordFill`: errors and state follow mapped branches.
- `observation`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `j.LookupFill`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- `j.FillEvents`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `Scan`: errors and state follow mapped branches.
- `j.db.QueryRowContext`: errors and state follow mapped branches.
- `j.TrackedFillOrders`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 10; return points: 3; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
