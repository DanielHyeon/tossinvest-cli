# Risk Pattern Report: `EntryGate.CheckEntryFor`

Source: `internal/execgw/symbolgate.go`  
AST source SHA-256: `46d559bec6ca59e70876da33face5ae8dadce7d832c5c32e105dd2add8cc8617`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
