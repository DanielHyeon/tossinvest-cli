# Risk Pattern Report: `TestAnOperatorResolutionSurvivesARestart`

Source: `internal/reconcile/restore_test.go`  
AST source SHA-256: `06361705cca4cd1d8cfd0263dff7b47ea9c661cac3c4b09bd164ba91b75c67f4`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
