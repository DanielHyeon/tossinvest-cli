# Risk Pattern Report: `TestAFailingApplyHookRollsBackTheFill`

Source: `internal/journal/apply_hook_test.go`
AST source SHA-256: `26d73b9371960a62335c0be0eef4750f398ad5099dca7b88222de4a126e09ccd`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
