# Function Logic Map: `TestIncludeOnlyAdoptionRequiresAPriceReader`

Source: `internal/app/engine/reconcileloop_test.go`
Function: `TestIncludeOnlyAdoptionRequiresAPriceReader`
Signature: `TestIncludeOnlyAdoptionRequiresAPriceReader(params=1, results=0)`
Source SHA-256: `feb0b59737a7c47e4ead572b77c9f2b591273fa6bd61744850a60c87830d6342`
Revision: `current`

## Inputs and invariants

- Inputs are `TestIncludeOnlyAdoptionRequiresAPriceReader(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/app/engine/reconcileloop_test.go:435 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/app/engine/reconcileloop_test.go:446 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `clock.NewFake`: errors and state follow mapped branches.
- `journal.Open`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `filepath.Join`: errors and state follow mapped branches.
- `t.TempDir`: errors and state follow mapped branches.
- `journal.FixedFSProber`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `t.Cleanup`: errors and state follow mapped branches.
- `j.Close`: errors and state follow mapped branches.
- `engine.NewReconcileDriver`: errors and state follow mapped branches.
- `errors.Is`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 4; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
