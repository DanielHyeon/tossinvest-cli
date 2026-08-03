# Risk Pattern Report: `OrderIdentity.less`

Source: `internal/reconcile/compare.go`
AST source SHA-256: `36ce21d173549fe4b957c6132a56993887fb62dfe3acaa7c9afd39a6e61154b2`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
