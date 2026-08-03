# Function Logic Map: `refusingReconcileRelease.ReleaseReconcile`

Source: `internal/app/engine/reconcileloop_test.go`
Function: `refusingReconcileRelease.ReleaseReconcile`
Signature: `refusingReconcileRelease.ReleaseReconcile(params=2, results=3)`
Source SHA-256: `feb0b59737a7c47e4ead572b77c9f2b591273fa6bd61744850a60c87830d6342`
Revision: `current`

## Inputs and invariants

- Inputs are `refusingReconcileRelease.ReleaseReconcile(params=2, results=3)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/app/engine/reconcileloop_test.go:100 | Preserve the source-bound happy path and propagated errors. |

## Calls and live bindings

- No nested call is present; behavior is source-local and source-hash bound.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 0; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
