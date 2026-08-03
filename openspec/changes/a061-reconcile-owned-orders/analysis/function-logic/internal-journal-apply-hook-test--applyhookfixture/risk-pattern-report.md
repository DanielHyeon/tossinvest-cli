# Risk Pattern Report: `applyHookFixture`

Source: `internal/journal/apply_hook_test.go`  
AST source SHA-256: `26d73b9371960a62335c0be0eef4750f398ad5099dca7b88222de4a126e09ccd`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
