# Risk Pattern Report: `exitContextForFill`

Source: `internal/journal/exit_state.go`
AST source SHA-256: `f3895fb41abc09f4de2aad1eceeeff1b39ab17ed658b2dc74e02bf7727b46f86`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
