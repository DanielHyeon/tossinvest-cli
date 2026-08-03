# Risk Pattern Report: `newLedger`

Source: `internal/filldetect/detect_test.go`
AST source SHA-256: `7fe5825a894d212e278325c39d6b369d975ef46f006b913627daa8c7264e2e26`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
