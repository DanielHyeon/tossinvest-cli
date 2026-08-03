# Risk Pattern Report: `TestDerivedTerminalFillReleasesTheHold`

Source: `internal/journal/reservation_release_test.go`
AST source SHA-256: `d00f0ca0ea14ea4b020f324d064436b0a1de2fa05e8005b1d973e7626d0d5fa7`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
