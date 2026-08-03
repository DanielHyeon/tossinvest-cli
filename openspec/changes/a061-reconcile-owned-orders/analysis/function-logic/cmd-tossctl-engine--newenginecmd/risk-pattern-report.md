# Risk Pattern Report: `newEngineCmd`

Source: `cmd/tossctl/engine.go`
AST source SHA-256: `45414562be8a352d2183fb2dfc0985154e0eea5ce781e167eb6800841c495451`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
