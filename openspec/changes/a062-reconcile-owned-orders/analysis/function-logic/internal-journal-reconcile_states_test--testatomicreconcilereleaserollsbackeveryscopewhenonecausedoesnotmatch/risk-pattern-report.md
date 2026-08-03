# Risk Pattern Report: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`

Source: `internal/journal/reconcile_states_test.go`
AST source SHA-256: `d2de10c4ae8c4d15346e190fa03cbc7bd4db7648bd7b4ab102272225dd1785a6`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
