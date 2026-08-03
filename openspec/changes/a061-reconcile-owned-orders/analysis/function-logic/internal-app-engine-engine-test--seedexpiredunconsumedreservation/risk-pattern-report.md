# Risk Pattern Report: `seedExpiredUnconsumedReservation`

Source: `internal/app/engine/engine_test.go`  
AST source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
