# Risk Pattern Report: `parseSnapshot`

Source: `internal/filldetect/payload.go`
AST source SHA-256: `564abf540ee18280e610ef6910202dbb746846d7869072ce7112b45d72649508`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
