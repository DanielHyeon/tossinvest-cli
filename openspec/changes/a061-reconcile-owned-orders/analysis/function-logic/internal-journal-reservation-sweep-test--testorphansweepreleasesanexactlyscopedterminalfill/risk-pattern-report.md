# Risk Pattern Report: `TestOrphanSweepReleasesAnExactlyScopedTerminalFill`

Source: `internal/journal/reservation_sweep_test.go`  
AST source SHA-256: `270775c1e9bf78356ecab224999f302311e6d690e28c55e853e260c86b098503`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
