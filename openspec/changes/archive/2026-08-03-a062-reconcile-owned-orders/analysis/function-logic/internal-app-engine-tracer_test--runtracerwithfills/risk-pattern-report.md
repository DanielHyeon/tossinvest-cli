# Risk Pattern Report: `runTracerWithFills`

Source: `internal/app/engine/tracer_test.go`
AST source SHA-256: `6d50eb9c3d64746ce4b3430c56a3b714fe00cf852325806b2bdc1cf73014e582`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
