# Risk Pattern Report: `EntryGate.SetAuthorityRefresh`

Source: `internal/execgw/retry.go`
AST source SHA-256: `a549135ef2864ab05eb8168cfac899cad5052457189307c4ed3e3bee42e102d3`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
