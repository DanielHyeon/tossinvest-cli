# Function Logic Map: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause`

Source: `internal/reconcile/restore_test.go`
Function: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause`
Signature: `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause(params=1, results=0)`
Source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`
Revision: `current`

## Inputs and invariants

- Inputs are `TestPermanentPromotionDoesNotOverwriteAnAccountWideForeignCause(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:565 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/reconcile/restore_test.go:572 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | for | internal/reconcile/restore_test.go:576 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/reconcile/restore_test.go:577 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | range | internal/reconcile/restore_test.go:583 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/reconcile/restore_test.go:584 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/reconcile/restore_test.go:589 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/reconcile/restore_test.go:592 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `context.Background`: errors and state follow mapped branches.
- `clock.NewFake`: errors and state follow mapped branches.
- `openJournal`: errors and state follow mapped branches.
- `j.EnterReconcile`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `trackerOn`: errors and state follow mapped branches.
- `noStaleGate`: errors and state follow mapped branches.
- `tracker.Restore`: errors and state follow mapped branches.
- `tracker.Observe`: errors and state follow mapped branches.
- `mismatchDiff`: errors and state follow mapped branches.
- `clk.Advance`: errors and state follow mapped branches.
- `tracker.Blocks`: errors and state follow mapped branches.
- `tracker.Permanent`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 11; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
