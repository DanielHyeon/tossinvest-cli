# Function Logic Map: `Attempt.transition`

Source: `internal/journal/durability.go`
Function: `Attempt.transition`
Signature: `Attempt.transition(params=3, results=1)`
Source SHA-256: `29ec7a1849fade446c9125a5f604ab37ea23080b240472390cf4f5b3c534b1e9`
Revision: `current`

## Inputs and invariants

- Inputs are `Attempt.transition(params=3, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/journal/durability.go:554 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/journal/durability.go:560 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/journal/durability.go:562 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/journal/durability.go:567 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/journal/durability.go:570 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/journal/durability.go:571 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/journal/durability.go:575 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/journal/durability.go:577 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/journal/durability.go:578 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/journal/durability.go:584 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/journal/durability.go:604 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/journal/durability.go:611 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/journal/durability.go:615 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | if | internal/journal/durability.go:619 | Preserve the condition, error propagation, and fail-closed behavior. |
| B15 | if | internal/journal/durability.go:623 | Preserve the condition, error propagation, and fail-closed behavior. |
| B16 | if | internal/journal/durability.go:627 | Preserve the condition, error propagation, and fail-closed behavior. |
| B17 | if | internal/journal/durability.go:635 | Preserve the condition, error propagation, and fail-closed behavior. |
| B18 | if | internal/journal/durability.go:639 | Preserve the condition, error propagation, and fail-closed behavior. |
| B19 | if | internal/journal/durability.go:642 | Preserve the condition, error propagation, and fail-closed behavior. |
| B20 | if | internal/journal/durability.go:649 | Preserve the condition, error propagation, and fail-closed behavior. |
| B21 | if | internal/journal/durability.go:662 | Preserve the condition, error propagation, and fail-closed behavior. |
| B22 | if | internal/journal/durability.go:663 | Preserve the condition, error propagation, and fail-closed behavior. |
| B23 | if | internal/journal/durability.go:668 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `a.mu.Lock`: errors and state follow mapped branches.
- `a.mu.Unlock`: errors and state follow mapped branches.
- `a.j.nowString`: errors and state follow mapped branches.
- `a.j.db.BeginTx`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `tx.Rollback`: errors and state follow mapped branches.
- `Scan`: errors and state follow mapped branches.
- `tx.QueryRowContext`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- `checkTransitionAllowed`: errors and state follow mapped branches.
- `consumeDecisionNonce`: errors and state follow mapped branches.
- `journalTimeStrictlyAfter`: errors and state follow mapped branches.
- `string`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `tx.ExecContext`: errors and state follow mapped branches.
- `strings.Join`: errors and state follow mapped branches.
- `res.RowsAffected`: errors and state follow mapped branches.
- `releasesReservations`: errors and state follow mapped branches.
- `releaseReservationsForAttempt`: errors and state follow mapped branches.
- `fmt.Sprintf`: errors and state follow mapped branches.
- `tx.Commit`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 28; return points: 15; deferred operations: 2.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
