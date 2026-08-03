# Risk Pattern Report: `TestPrunePreservesSpentNonceWhileItsReservationIsHeld`

Source: `internal/journal/nonce_test.go`  
AST source SHA-256: `84f12358991f53b64e3f9dbebef2730533e6375e8ea55289c80b9cdfeb35487f`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
