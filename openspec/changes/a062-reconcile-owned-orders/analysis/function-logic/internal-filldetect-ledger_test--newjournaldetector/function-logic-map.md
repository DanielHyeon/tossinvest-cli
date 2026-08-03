# Function Logic Map: `newJournalDetector`

Source: `internal/filldetect/ledger_test.go`
Function: `newJournalDetector`
Signature: `newJournalDetector(params=5, results=2)`
Source SHA-256: `bec446ca895a16f15bd76144da044c9308c3d8d5f71e2959e04fc83b77fa78bb`
Revision: `current`

## Inputs and invariants

- Inputs are `newJournalDetector(params=5, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | happy path | internal/filldetect/ledger_test.go:38 | Preserve the source-bound happy path and propagated errors. |

## Calls and live bindings

- `t.Helper`: errors and state follow mapped branches.
- `openLedgerJournal`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 1; return points: 1; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
