# Function Logic Map: `ReconcileDriver.RunOnce`

Source: `internal/app/engine/reconcileloop.go`
Function: `ReconcileDriver.RunOnce`
Signature: `ReconcileDriver.RunOnce(params=1, results=1)`
Source SHA-256: `accaa4c5f6645d8af7be3f1cbcd9ec61a7efc9f1f022be26b39b53789d867763`
Revision: `current`

## Inputs and invariants

- Inputs are `ReconcileDriver.RunOnce(params=1, results=1)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop.go:387 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/reconcileloop.go:393 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/app/engine/reconcileloop.go:404 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/app/engine/reconcileloop.go:410 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/app/engine/reconcileloop.go:416 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/app/engine/reconcileloop.go:417 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | if | internal/app/engine/reconcileloop.go:425 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/app/engine/reconcileloop.go:426 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `unknown`: errors and state follow mapped branches.
- `d.note`: errors and state follow mapped branches.
- `d.stabilise`: errors and state follow mapped branches.
- `reconcile.LocalStateFromJournal`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `Compare`: errors and state follow mapped branches.
- `d.ingest.IngestExternalPositions`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `d.opts.Converge.ConvergeQuantities`: errors and state follow mapped branches.
- `d.opts.Tracker.Refresh`: errors and state follow mapped branches.
- `d.opts.Tracker.Observe`: errors and state follow mapped branches.
- `d.opts.Tracker.Blocks`: errors and state follow mapped branches.
- `d.judgeHoldings`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 18; return points: 5; deferred operations: 1.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
