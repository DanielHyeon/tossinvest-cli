# Risk Pattern Report: `Snapshot.Digest`

Source: `internal/reconcile/snapshot.go`  
AST source SHA-256: `827f148d49ae878bd1acb64327dbd5545cebe9a576e130305255e06861e1b8e3`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
