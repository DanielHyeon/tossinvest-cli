# Risk Pattern Report: `TestJournalTrackedPreservesCanonicalOrderScope`

Source: `internal/filldetect/ledger_test.go`
AST source SHA-256: `bec446ca895a16f15bd76144da044c9308c3d8d5f71e2959e04fc83b77fa78bb`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
