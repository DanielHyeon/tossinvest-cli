# Risk Pattern Report: `Journal.guardTrackedFillIdentity`

Source: `internal/journal/fills.go`  
AST source SHA-256: `8ee09a6b042e305d9e8d913eb86beb14f874d034f8ad8974ca488f8080699e9a`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
