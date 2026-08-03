# Risk Pattern Report: `openEngineJournal`

Source: `internal/app/engine/engine.go`
AST source SHA-256: `401ab52518aac369f7567a60f711c4a019efad96ab0bcd7af1751155ba67e1f5`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
