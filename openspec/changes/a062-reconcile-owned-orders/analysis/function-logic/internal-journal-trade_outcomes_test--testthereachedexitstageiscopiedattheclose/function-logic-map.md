# Function Logic Map: `TestTheReachedExitStageIsCopiedAtTheClose`

Source: `internal/journal/trade_outcomes_test.go`
Function: `TestTheReachedExitStageIsCopiedAtTheClose`
Signature: `TestTheReachedExitStageIsCopiedAtTheClose(params=1, results=0)`
Source SHA-256: `cbd539d47024cf89c59d18b54a014449c9b080a1cfa54eeed9e8f43449f2c2c3`
Revision: `base`

## Inputs and invariants

- Inputs are `TestTheReachedExitStageIsCopiedAtTheClose(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/trade_outcomes_test.go:167 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/trade_outcomes_test.go:171 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/trade_outcomes_test.go:176 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/trade_outcomes_test.go:187 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/trade_outcomes_test.go:192 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/trade_outcomes_test.go:195 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `outcomeFixture`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `place`: errors and state follow mapped branches.
- `j.RecordFill`: errors and state follow mapped branches.
- `terminalFill`: errors and state follow mapped branches.
- `t.Fatal`: errors and state follow mapped branches.
- `currentPosition`: errors and state follow mapped branches.
- `j.OpenExitState`: errors and state follow mapped branches.
- `j.RecordExitJudgement`: errors and state follow mapped branches.
- `outcomeOf`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 10; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
