# Risk Pattern Report: `TestStartupPrunesSpentNoncesOnce`

Source: `internal/app/engine/engine_test.go`  
AST source SHA-256: `dec52090cdbaa17f2a868d8e204c1755d88ec127a3f5158efc858405f41b83b7`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
