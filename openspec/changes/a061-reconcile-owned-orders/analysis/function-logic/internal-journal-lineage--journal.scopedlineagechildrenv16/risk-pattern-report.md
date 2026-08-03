# Risk Pattern Report: `Journal.scopedLineageChildrenV16`

Source: `internal/journal/lineage.go`  
AST source SHA-256: `73943302679524a29931771062a92c6140e53ffd5724c54620eab50f1740508a`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
