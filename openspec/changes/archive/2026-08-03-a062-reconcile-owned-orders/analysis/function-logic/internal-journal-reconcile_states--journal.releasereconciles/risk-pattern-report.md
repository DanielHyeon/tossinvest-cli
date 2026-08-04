# Risk Pattern Report: `Journal.ReleaseReconciles`

Source: `internal/journal/reconcile_states.go`
AST source SHA-256: `1a5e5aa3d3c37c940bb43adaebb05b8585256908cf2b28f0112da141ede1eb08`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
