# Risk Pattern Report: `TestTheFirstFillOpensTheInstance`

Source: `internal/journal/position_projection_test.go`
AST source SHA-256: `6ab3463bdc484584a3e1dc23b86cabc42fa737122966e7ed57b96ec78bd1572f`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
