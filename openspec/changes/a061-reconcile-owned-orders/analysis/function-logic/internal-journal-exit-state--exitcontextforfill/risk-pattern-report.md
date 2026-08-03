# Risk Pattern Report: `exitContextForFill`

Source: `internal/journal/exit_state.go`  
AST source SHA-256: `f3895fb41abc09f4de2aad1eceeeff1b39ab17ed658b2dc74e02bf7727b46f86`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
