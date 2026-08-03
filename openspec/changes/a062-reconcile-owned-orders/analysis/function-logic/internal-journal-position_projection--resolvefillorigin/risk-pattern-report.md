# Risk Pattern Report: `resolveFillOrigin`

Source: `internal/journal/position_projection.go`
AST source SHA-256: `5b62d380b860e67884ceb5f5e217a12fa0d8a190c54190f370ae711133002748`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
