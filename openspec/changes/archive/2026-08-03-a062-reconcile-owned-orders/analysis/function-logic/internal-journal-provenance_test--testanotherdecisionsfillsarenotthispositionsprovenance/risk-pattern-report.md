# Risk Pattern Report: `TestAnotherDecisionsFillsAreNotThisPositionsProvenance`

Source: `internal/journal/provenance_test.go`
AST source SHA-256: `3a77145080d4963658125cee4e7ae33db9f1c6c76d329aa893f589748775f301`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
