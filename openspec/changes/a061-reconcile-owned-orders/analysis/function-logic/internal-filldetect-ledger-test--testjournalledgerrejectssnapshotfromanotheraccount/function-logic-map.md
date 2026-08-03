# Function Logic Map: `TestJournalLedgerRejectsSnapshotFromAnotherAccount`

Source: `internal/filldetect/ledger_test.go`  
Function: `TestJournalLedgerRejectsSnapshotFromAnotherAccount`  
Signature: `TestJournalLedgerRejectsSnapshotFromAnotherAccount(params=1, results=0)`  
Source SHA-256: `bec446ca895a16f15bd76144da044c9308c3d8d5f71e2959e04fc83b77fa78bb`

## Inputs and invariants

- Inputs are the parameters in `TestJournalLedgerRejectsSnapshotFromAnotherAccount(params=1, results=0)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/ledger_test.go:352 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `clock.NewFake`: returned errors and state follow the mapped branches.
- `openLedgerJournal`: returned errors and state follow the mapped branches.
- `filepath.Join`: returned errors and state follow the mapped branches.
- `t.TempDir`: returned errors and state follow the mapped branches.
- `ledger.Apply`: returned errors and state follow the mapped branches.
- `context.Background`: returned errors and state follow the mapped branches.
- `strings.Contains`: returned errors and state follow the mapped branches.
- `err.Error`: returned errors and state follow the mapped branches.
- `t.Fatalf`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 4 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
