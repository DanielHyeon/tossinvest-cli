# Risk Pattern Report: `TestDerivedTerminalFillReleasesTheHold`

Source: `internal/journal/reservation_release_test.go`  
AST source SHA-256: `8ff5b3579d3d6fbc39eebac1594e9acbfcbec475dc7be5ccd470f6cce753fb07`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
