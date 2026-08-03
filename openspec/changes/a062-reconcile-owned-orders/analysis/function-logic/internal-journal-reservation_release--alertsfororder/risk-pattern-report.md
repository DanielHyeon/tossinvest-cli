# Risk Pattern Report: `alertsForOrder`

Source: `internal/journal/reservation_release.go`
AST source SHA-256: `d61f428958e0ac5ba535af8148bdcb40d25f2c1893be6feb8f95b1e9af7b2ff2`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
