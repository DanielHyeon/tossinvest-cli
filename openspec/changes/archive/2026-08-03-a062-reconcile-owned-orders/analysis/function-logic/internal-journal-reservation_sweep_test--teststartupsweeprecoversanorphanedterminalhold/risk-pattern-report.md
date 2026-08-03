# Risk Pattern Report: `TestStartupSweepRecoversAnOrphanedTerminalHold`

Source: `internal/journal/reservation_sweep_test.go`
AST source SHA-256: `40bf008737b54e15935e4ad2855e1c09d2f42fd3b6a6f975efd9e2d2e074d7cc`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
