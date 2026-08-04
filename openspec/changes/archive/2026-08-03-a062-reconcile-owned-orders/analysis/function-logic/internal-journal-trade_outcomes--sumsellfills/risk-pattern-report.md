# Risk Pattern Report: `sumSellFills`

Source: `internal/journal/trade_outcomes.go`
AST source SHA-256: `0bd43abf96107b9998ea7c9bc6c6655f162b1af3ae96490daadc78fb354a4958`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
