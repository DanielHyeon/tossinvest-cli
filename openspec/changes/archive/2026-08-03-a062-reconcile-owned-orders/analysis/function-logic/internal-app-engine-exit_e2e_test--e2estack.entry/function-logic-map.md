# Function Logic Map: `e2eStack.entry`

Source: `internal/app/engine/exit_e2e_test.go`
Function: `e2eStack.entry`
Signature: `e2eStack.entry(params=4, results=1)`
Source SHA-256: `8cc6877572d4602364bcaf443e33aff26036804732de2e6fc0ace8a60aecdc2d`
Revision: `current`

## Inputs and invariants

- Inputs are `e2eStack.entry(params=4, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/exit_e2e_test.go:274 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/exit_e2e_test.go:278 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/app/engine/exit_e2e_test.go:304 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/app/engine/exit_e2e_test.go:307 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/app/engine/exit_e2e_test.go:310 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/app/engine/exit_e2e_test.go:313 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/app/engine/exit_e2e_test.go:319 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `s.t.Helper`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `execgw.EncodeLimits`: errors and state follow mapped branches.
- `execgw.Bound`: errors and state follow mapped branches.
- `s.t.Fatalf`: errors and state follow mapped branches.
- `s.engine.Journal.RecordDecision`: errors and state follow mapped branches.
- `s.clk.Now`: errors and state follow mapped branches.
- `Add`: errors and state follow mapped branches.
- `s.engine.Journal.Prepare`: errors and state follow mapped branches.
- `journal.DeriveClientOrderID`: errors and state follow mapped branches.
- `attempt.MarkDispatchStarted`: errors and state follow mapped branches.
- `attempt.MarkAcked`: errors and state follow mapped branches.
- `attempt.Settle`: errors and state follow mapped branches.
- `s.fill`: errors and state follow mapped branches.
- `s.engine.Journal.CurrentPosition`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 9; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
