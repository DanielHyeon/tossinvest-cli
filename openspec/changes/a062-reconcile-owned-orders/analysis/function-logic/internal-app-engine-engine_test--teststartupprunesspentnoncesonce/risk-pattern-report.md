# Risk Pattern Report: `TestStartupPrunesSpentNoncesOnce`

Source: `internal/app/engine/engine_test.go`
AST source SHA-256: `dec52090cdbaa17f2a868d8e204c1755d88ec127a3f5158efc858405f41b83b7`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
