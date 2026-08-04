# Risk Pattern Report: `TestSchemaIndexes`

Source: `internal/journal/schema_test.go`
AST source SHA-256: `50b6ce71056a4e67c8dbb344ab84bd429c6490b1ebc7f25fe7bc7d15e7c7c1be`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
