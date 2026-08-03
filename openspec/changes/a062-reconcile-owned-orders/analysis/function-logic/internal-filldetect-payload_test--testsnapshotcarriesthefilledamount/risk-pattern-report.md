# Risk Pattern Report: `TestSnapshotCarriesTheFilledAmount`

Source: `internal/filldetect/payload_test.go`
AST source SHA-256: `14be87b64e23ea392eff45b6daf15ee72d8307260b6d46185aac6ed748d5d9c2`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
