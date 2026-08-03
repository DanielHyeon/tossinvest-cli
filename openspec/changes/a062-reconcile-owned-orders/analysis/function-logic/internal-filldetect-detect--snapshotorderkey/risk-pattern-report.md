# Risk Pattern Report: `snapshotOrderKey`

Source: `internal/filldetect/detect.go`
AST source SHA-256: `5441296826821097f82da79215934616d295c31644f24c8c4126d5778594fb2b`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
