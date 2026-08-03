# Risk Pattern Report: `NewReconcileDriver`

Source: `internal/app/engine/reconcileloop.go`
AST source SHA-256: `accaa4c5f6645d8af7be3f1cbcd9ec61a7efc9f1f022be26b39b53789d867763`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
