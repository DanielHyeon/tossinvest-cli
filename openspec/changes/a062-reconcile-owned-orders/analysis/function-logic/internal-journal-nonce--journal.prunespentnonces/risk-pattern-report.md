# Risk Pattern Report: `Journal.PruneSpentNonces`

Source: `internal/journal/nonce.go`
AST source SHA-256: `1466fddb8d43a5481cdc10f06b53c09a340862ab91907b7ccc70e40d35b7959c`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
