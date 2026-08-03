# Risk Pattern Report: `buildGateway`

Source: `internal/app/engine/gateway.go`
AST source SHA-256: `3dead101adcc3b89767975b14f72de7246909ac0ef3f909e3928ebed2637ee8b`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
