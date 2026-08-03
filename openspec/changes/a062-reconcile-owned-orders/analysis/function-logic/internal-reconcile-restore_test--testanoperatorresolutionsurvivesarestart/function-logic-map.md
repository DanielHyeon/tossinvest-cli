# Function Logic Map: `TestAnOperatorResolutionSurvivesARestart`

Source: `internal/reconcile/restore_test.go`
Function: `TestAnOperatorResolutionSurvivesARestart`
Signature: `TestAnOperatorResolutionSurvivesARestart(params=1, results=0)`
Source SHA-256: `06361705cca4cd1d8cfd0263dff7b47ea9c661cac3c4b09bd164ba91b75c67f4`
Revision: `base`

## Inputs and invariants

- Inputs are `TestAnOperatorResolutionSurvivesARestart(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/reconcile/restore_test.go:144 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/reconcile/restore_test.go:147 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/reconcile/restore_test.go:150 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/reconcile/restore_test.go:157 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/reconcile/restore_test.go:160 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/reconcile/restore_test.go:163 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `context.Background`: errors and state follow mapped branches.
- `clock.NewFake`: errors and state follow mapped branches.
- `t.TempDir`: errors and state follow mapped branches.
- `openJournalAt`: errors and state follow mapped branches.
- `trackerOn`: errors and state follow mapped branches.
- `noStaleGate`: errors and state follow mapped branches.
- `tracker.Observe`: errors and state follow mapped branches.
- `mismatchDiff`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `tracker.Resolve`: errors and state follow mapped branches.
- `first.Close`: errors and state follow mapped branches.
- `restarted.Restore`: errors and state follow mapped branches.
- `restartedGate.CheckEntryFor`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `restarted.Blocks`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 13; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
