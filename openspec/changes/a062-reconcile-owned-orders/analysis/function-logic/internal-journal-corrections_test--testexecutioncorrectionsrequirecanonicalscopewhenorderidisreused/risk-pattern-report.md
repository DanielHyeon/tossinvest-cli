# Risk Pattern Report: `TestExecutionCorrectionsRequireCanonicalScopeWhenOrderIDIsReused`

Source: `internal/journal/corrections_test.go`
AST source SHA-256: `74650be142c65dda4f8e2f7347a7fcb84adb51d3d268094a6793d235b761dfb4`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
