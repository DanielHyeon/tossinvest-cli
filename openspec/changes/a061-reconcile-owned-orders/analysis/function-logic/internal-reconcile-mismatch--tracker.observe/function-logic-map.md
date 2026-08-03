# Function Logic Map: `Tracker.Observe`

Source: `internal/reconcile/mismatch.go`
Function: `Tracker.Observe`
Signature: `Tracker.Observe(params=2, results=2)`
Source SHA-256: `a0ffbb279e773f7648b0a844e4bb783fdd671125003f4eb8619a827ed0688b9f`
Revision: `current`

## Inputs and invariants

- Inputs are `Tracker.Observe(params=2, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/mismatch.go:366 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/reconcile/mismatch.go:373 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | else | internal/reconcile/mismatch.go:396 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | range | internal/reconcile/mismatch.go:380 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/reconcile/mismatch.go:381 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/reconcile/mismatch.go:388 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | range | internal/reconcile/mismatch.go:398 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/reconcile/mismatch.go:399 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/reconcile/mismatch.go:405 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | if | internal/reconcile/mismatch.go:418 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | range | internal/reconcile/mismatch.go:440 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | range | internal/reconcile/mismatch.go:443 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | if | internal/reconcile/mismatch.go:444 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | range | internal/reconcile/mismatch.go:449 | Preserve the condition, error propagation, and fail-closed behavior. |
| B15 | range | internal/reconcile/mismatch.go:453 | Preserve the condition, error propagation, and fail-closed behavior. |
| B16 | range | internal/reconcile/mismatch.go:456 | Preserve the condition, error propagation, and fail-closed behavior. |
| B17 | if | internal/reconcile/mismatch.go:462 | Preserve the condition, error propagation, and fail-closed behavior. |
| B18 | else | internal/reconcile/mismatch.go:475 | Preserve the condition, error propagation, and fail-closed behavior. |
| B19 | range | internal/reconcile/mismatch.go:467 | Preserve the condition, error propagation, and fail-closed behavior. |
| B20 | if | internal/reconcile/mismatch.go:468 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `Now`: errors and state follow mapped branches.
- `t.clock`: errors and state follow mapped branches.
- `t.interval`: errors and state follow mapped branches.
- `t.maxFailures`: errors and state follow mapped branches.
- `t.mu.Lock`: errors and state follow mapped branches.
- `diff.BlocksEntry`: errors and state follow mapped branches.
- `strings.ToUpper`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `blocksFor`: errors and state follow mapped branches.
- `block.Key`: errors and state follow mapped branches.
- `fmt.Sprintf`: errors and state follow mapped branches.
- `permanent.Key`: errors and state follow mapped branches.
- `sortBlocks`: errors and state follow mapped branches.
- `t.syncGate`: errors and state follow mapped branches.
- `t.snapshotBlocks`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `t.persist`: errors and state follow mapped branches.
- `delete`: errors and state follow mapped branches.
- `hasPermanentQuantityAccountBlock`: errors and state follow mapped branches.
- `now.Add`: errors and state follow mapped branches.
- `t.mu.Unlock`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 39; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
