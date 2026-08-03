# Risk Pattern Report: `TestAnOperatorResolutionSurvivesARestart`

Source: `internal/reconcile/restore_test.go`
AST source SHA-256: `06361705cca4cd1d8cfd0263dff7b47ea9c661cac3c4b09bd164ba91b75c67f4`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
