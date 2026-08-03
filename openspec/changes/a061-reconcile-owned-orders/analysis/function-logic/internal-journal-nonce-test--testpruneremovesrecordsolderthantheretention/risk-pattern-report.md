# Risk Pattern Report: `TestPruneRemovesRecordsOlderThanTheRetention`

Source: `internal/journal/nonce_test.go`  
AST source SHA-256: `83fcf17c3cd3758fadd4f23e7f31e675b8e3a2df7d56d3cdd6e70b583e16f5e3`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
