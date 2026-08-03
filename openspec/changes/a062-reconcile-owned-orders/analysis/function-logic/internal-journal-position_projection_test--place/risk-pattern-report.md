# Risk Pattern Report: `place`

Source: `internal/journal/position_projection_test.go`
AST source SHA-256: `e1094b972b2f61b58d5665501165349c25b2a90624b2256090185b8eda37de35`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
