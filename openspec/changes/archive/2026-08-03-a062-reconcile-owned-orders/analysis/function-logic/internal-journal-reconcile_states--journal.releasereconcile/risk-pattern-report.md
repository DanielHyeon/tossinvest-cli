# Risk Pattern Report: `Journal.ReleaseReconcile`

Source: `internal/journal/reconcile_states.go`
AST source SHA-256: `f07e1a91c10a72e1226e5cf5328d461def19b571714145d31ccb838c2e402e19`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.
