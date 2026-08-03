# Risk Pattern Report: `EntryGate.SetAuthorityRefresh`

Source: `internal/execgw/retry.go`  
AST source SHA-256: `a549135ef2864ab05eb8168cfac899cad5052457189307c4ed3e3bee42e102d3`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
