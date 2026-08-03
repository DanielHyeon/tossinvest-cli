# Risk Pattern Report: `Snapshot.Digest`

Source: `internal/reconcile/snapshot.go`
AST source SHA-256: `827f148d49ae878bd1acb64327dbd5545cebe9a576e130305255e06861e1b8e3`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
