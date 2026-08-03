# Risk Pattern Report: `TestSchemaIndexes`

Source: `internal/journal/schema_test.go`  
AST source SHA-256: `9d6a1364f76a4dc3c6e9854066182ca18e59be53458d64e42f13a9e74df95d7b`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
