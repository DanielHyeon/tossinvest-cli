# Risk Pattern Report: `TestPreOwnerTerminalEvidenceCannotReleaseAFutureReservation`

Source: `internal/journal/reservation_sweep_test.go`
AST source SHA-256: `dcfcc7de8f1e832e67d19949578ce40bf3d79d549ad6f99de6f5a714cfcd011e`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
