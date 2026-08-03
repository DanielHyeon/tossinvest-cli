# Risk Pattern Report: `insertNonConfirmedFillAttempt`

Source: `internal/journal/provenance_test.go`
AST source SHA-256: `42fabaf43a4709cf94fadf4c29b9daeb6e6967779c17161f242ab6184d7e5003`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
