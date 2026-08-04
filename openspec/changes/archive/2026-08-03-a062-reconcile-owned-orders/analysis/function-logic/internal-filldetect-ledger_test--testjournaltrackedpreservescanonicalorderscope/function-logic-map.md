# Function Logic Map: `TestJournalTrackedPreservesCanonicalOrderScope`

Source: `internal/filldetect/ledger_test.go`
Function: `TestJournalTrackedPreservesCanonicalOrderScope`
Signature: `TestJournalTrackedPreservesCanonicalOrderScope(params=1, results=0)`
Source SHA-256: `bec446ca895a16f15bd76144da044c9308c3d8d5f71e2959e04fc83b77fa78bb`
Revision: `current`

## Inputs and invariants

- Inputs are `TestJournalTrackedPreservesCanonicalOrderScope(params=1, results=0)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger_test.go:327 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/filldetect/ledger_test.go:331 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/filldetect/ledger_test.go:334 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | if | internal/filldetect/ledger_test.go:338 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `clock.NewFake`: errors and state follow mapped branches.
- `openLedgerJournal`: errors and state follow mapped branches.
- `filepath.Join`: errors and state follow mapped branches.
- `t.TempDir`: errors and state follow mapped branches.
- `recordConfirmedLedgerOrder`: errors and state follow mapped branches.
- `source.SelectedAccountRef`: errors and state follow mapped branches.
- `t.Fatalf`: errors and state follow mapped branches.
- `source.TrackedOrders`: errors and state follow mapped branches.
- `context.Background`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 6; return points: 0; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
