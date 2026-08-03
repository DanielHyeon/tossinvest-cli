# Risk Pattern Report: `Attempt.transition`

Source: `internal/journal/durability.go`
AST source SHA-256: `29ec7a1849fade446c9125a5f604ab37ea23080b240472390cf4f5b3c534b1e9`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
