# Risk Pattern Report: `Journal.ResolveCurrentOrderID`

Source: `internal/journal/lineage.go`
AST source SHA-256: `bf26e9cfd6030033e99ec6ee2ceb53dd5843a0c4c25111e8761844c598bb9d73`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
