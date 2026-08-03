# Risk Pattern Report: `sweepOrphanedTerminals`

Source: `internal/journal/reservation_release.go`  
AST source SHA-256: `7e7ef60ba4a8325d9a2bfca513828195ff7741058af4ab4a9dc6d2d843334718`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
