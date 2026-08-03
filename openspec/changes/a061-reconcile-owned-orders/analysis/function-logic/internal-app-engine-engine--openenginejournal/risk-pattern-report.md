# Risk Pattern Report: `openEngineJournal`

Source: `internal/app/engine/engine.go`  
AST source SHA-256: `401ab52518aac369f7567a60f711c4a019efad96ab0bcd7af1751155ba67e1f5`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
