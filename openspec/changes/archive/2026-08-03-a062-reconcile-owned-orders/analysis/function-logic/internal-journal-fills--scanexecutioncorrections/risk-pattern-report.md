# Risk Pattern Report: `scanExecutionCorrections`

Source: `internal/journal/fills.go`
AST source SHA-256: `000918b94c8c3f776b611421c412e4604086fc4cbee2fd0e7c21fe0dd46454c0`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
