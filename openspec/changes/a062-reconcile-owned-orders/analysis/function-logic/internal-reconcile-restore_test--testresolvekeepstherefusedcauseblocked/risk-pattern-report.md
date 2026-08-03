# Risk Pattern Report: `TestResolveKeepsTheRefusedCauseBlocked`

Source: `internal/reconcile/restore_test.go`
AST source SHA-256: `06075e0e4501b78ee04e55e617309bf70a7b1a025c31d7a496cd9396161bc2ab`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
