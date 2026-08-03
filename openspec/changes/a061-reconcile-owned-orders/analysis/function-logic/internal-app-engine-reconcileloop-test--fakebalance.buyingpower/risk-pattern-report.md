# Risk Pattern Report: `fakeBalance.BuyingPower`

Source: `internal/app/engine/reconcileloop_test.go`  
AST source SHA-256: `f7244f04d716230ddc2536f8e219958c52b86a6b899cf6f4df45fa09962f961e`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
