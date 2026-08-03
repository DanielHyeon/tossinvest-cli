# Risk Pattern Report: `TestScopedTerminalStateDoesNotHideAnotherAccountsLiveReusedOrder`

Source: `internal/journal/fills_test.go`  
AST source SHA-256: `e322e6a62817b22a0ed66fb2c17067e2d8707c87e0ae69c648fa3bfc7c766c56`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
