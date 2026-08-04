# Risk Pattern Report: `Attempt.resolveWithLineage`

Source: `internal/journal/lineage.go`
AST source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
