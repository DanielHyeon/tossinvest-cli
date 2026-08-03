# Risk Pattern Report: `TestSchemaIndexes`

Source: `internal/journal/schema_test.go`  
AST source SHA-256: `5e56ff9da74a1775d91251b7360bbf9bddecbcdd2ee5c7f2063ab3d9213cb396`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
