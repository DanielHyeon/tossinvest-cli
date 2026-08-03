# Risk Pattern Report: `buildGateway`

Source: `internal/app/engine/gateway.go`  
AST source SHA-256: `3dead101adcc3b89767975b14f72de7246909ac0ef3f909e3928ebed2637ee8b`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
