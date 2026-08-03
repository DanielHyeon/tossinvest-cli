# Risk Pattern Report: `TestBlankPayloadIdFallsBackToTheTrackedId`

Source: `internal/filldetect/payload_test.go`
AST source SHA-256: `2a3179003b761f34a7ba63d94ba7f3c439689cc48bb00c04a23837a23f97fa9a`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
