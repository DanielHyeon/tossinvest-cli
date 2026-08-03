# Function Logic Map: `e2eStack.fill`

Source: `internal/app/engine/exit_e2e_test.go`
Function: `e2eStack.fill`
Signature: `e2eStack.fill(params=7, results=0)`
Source SHA-256: `8cc6877572d4602364bcaf443e33aff26036804732de2e6fc0ace8a60aecdc2d`
Revision: `current`

## Inputs and invariants

- Inputs are `e2eStack.fill(params=7, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/exit_e2e_test.go:330 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/exit_e2e_test.go:333 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `s.t.Helper`: errors and state follow mapped branches.
- `s.engine.Journal.RecordFill`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `Format`: errors and state follow mapped branches.
- `s.clk.Now`: errors and state follow mapped branches.
- `s.t.Fatalf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 3; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
