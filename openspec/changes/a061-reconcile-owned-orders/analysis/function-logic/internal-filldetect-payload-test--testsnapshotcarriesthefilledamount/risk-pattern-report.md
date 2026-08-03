# Risk Pattern Report: `TestSnapshotCarriesTheFilledAmount`

Source: `internal/filldetect/payload_test.go`  
AST source SHA-256: `14be87b64e23ea392eff45b6daf15ee72d8307260b6d46185aac6ed748d5d9c2`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
