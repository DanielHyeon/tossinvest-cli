# Risk Pattern Report: `EntryGate.CheckEntryFor`

Source: `internal/execgw/symbolgate.go`
AST source SHA-256: `46d559bec6ca59e70876da33face5ae8dadce7d832c5c32e105dd2add8cc8617`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
