# Risk Pattern Report: `EntryGate.CheckEntryFor`

Source: `internal/execgw/symbolgate.go`  
AST source SHA-256: `46d559bec6ca59e70876da33face5ae8dadce7d832c5c32e105dd2add8cc8617`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
