# Risk Pattern Report: `Journal.PositionProvenance`

Source: `internal/journal/provenance.go`
AST source SHA-256: `77a75f3139467ec1119dd5ad6d06b36505766294a45aabdc70fa4930583c8bc9`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
