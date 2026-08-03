# Risk Pattern Report: `TestActiveForeignCauseIsCountedAndBlocksBeforePriceRead`

Source: `internal/app/engine/reconcileloop_test.go`
AST source SHA-256: `feb0b59737a7c47e4ead572b77c9f2b591273fa6bd61744850a60c87830d6342`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
