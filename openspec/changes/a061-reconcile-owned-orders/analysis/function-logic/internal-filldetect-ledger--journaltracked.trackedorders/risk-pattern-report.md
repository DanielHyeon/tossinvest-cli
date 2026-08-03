# Risk Pattern Report: `JournalTracked.TrackedOrders`

Source: `internal/filldetect/ledger.go`
AST source SHA-256: `75966bf1507d412f48b88efddab42f045fff31f45c76bc9cf43dfc0a634242c4`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
