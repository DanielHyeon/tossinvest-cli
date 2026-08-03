# Risk Pattern Report: `newJournalDetector`

Source: `internal/filldetect/ledger_test.go`  
AST source SHA-256: `3fa79a4b73613bae1f9ee086608012e2dca32a6f85b5b9128b916016fdfe571a`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
