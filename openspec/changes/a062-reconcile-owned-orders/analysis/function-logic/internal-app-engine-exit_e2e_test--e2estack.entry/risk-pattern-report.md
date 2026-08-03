# Risk Pattern Report: `e2eStack.entry`

Source: `internal/app/engine/exit_e2e_test.go`
AST source SHA-256: `8cc6877572d4602364bcaf443e33aff26036804732de2e6fc0ace8a60aecdc2d`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
