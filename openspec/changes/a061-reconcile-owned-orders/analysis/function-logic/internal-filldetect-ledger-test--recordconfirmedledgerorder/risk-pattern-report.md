# Risk Pattern Report: `recordConfirmedLedgerOrder`

Source: `internal/filldetect/ledger_test.go`  
AST source SHA-256: `bec446ca895a16f15bd76144da044c9308c3d8d5f71e2959e04fc83b77fa78bb`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
