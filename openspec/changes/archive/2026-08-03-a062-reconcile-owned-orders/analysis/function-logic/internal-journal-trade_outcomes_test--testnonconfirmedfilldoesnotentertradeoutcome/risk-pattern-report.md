# Risk Pattern Report: `TestNonConfirmedFillDoesNotEnterTradeOutcome`

Source: `internal/journal/trade_outcomes_test.go`
AST source SHA-256: `1723845dbc5c11be31276e125b182a5f02a9401abb17356649aca0f50858ada2`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
