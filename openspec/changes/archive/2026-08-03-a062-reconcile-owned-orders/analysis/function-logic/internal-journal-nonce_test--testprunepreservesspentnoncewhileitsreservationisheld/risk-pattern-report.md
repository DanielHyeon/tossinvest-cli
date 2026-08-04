# Risk Pattern Report: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld`

Source: `internal/journal/nonce_test.go`
AST source SHA-256: `84f12358991f53b64e3f9dbebef2730533e6375e8ea55289c80b9cdfeb35487f`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
