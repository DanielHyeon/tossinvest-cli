# Function Logic Map: `TestTheFirstFillOpensTheInstance`

Source: `internal/journal/position_projection_test.go`
Function: `TestTheFirstFillOpensTheInstance`
Signature: `TestTheFirstFillOpensTheInstance(params=1, results=0)`
Source SHA-256: `6ab3463bdc484584a3e1dc23b86cabc42fa737122966e7ed57b96ec78bd1572f`
Revision: `base`

## Inputs and invariants

- Inputs are `TestTheFirstFillOpensTheInstance(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/position_projection_test.go:132 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/position_projection_test.go:136 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/position_projection_test.go:141 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/position_projection_test.go:144 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/position_projection_test.go:147 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/position_projection_test.go:150 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/position_projection_test.go:153 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/position_projection_test.go:156 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/position_projection_test.go:159 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/position_projection_test.go:162 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `projectingJournal`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `place`: errors and state follow mapped branches.
- `j.CurrentPosition`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `j.RecordFill`: errors and state follow mapped branches.
- `fillOf`: errors and state follow mapped branches.
- `currentPosition`: errors and state follow mapped branches.
- `t.Errorf`: errors and state follow mapped branches.
- `t.Error`: errors and state follow mapped branches.
- `PositionID`: errors and state follow mapped branches.
- `p.ExitEligible`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 6; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
