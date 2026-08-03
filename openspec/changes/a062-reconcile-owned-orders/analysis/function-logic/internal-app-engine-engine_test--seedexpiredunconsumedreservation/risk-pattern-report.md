# Risk Pattern Report: `seedExpiredUnconsumedReservation`

Source: `internal/app/engine/engine_test.go`
AST source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
