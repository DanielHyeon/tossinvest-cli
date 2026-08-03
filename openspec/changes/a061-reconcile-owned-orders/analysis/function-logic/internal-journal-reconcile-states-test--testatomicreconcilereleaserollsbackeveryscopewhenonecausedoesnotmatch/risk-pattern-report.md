# Risk Pattern Report: `TestAtomicReconcileReleaseRollsBackEveryScopeWhenOneCauseDoesNotMatch`

Source: `internal/journal/reconcile_states_test.go`  
AST source SHA-256: `d2de10c4ae8c4d15346e190fa03cbc7bd4db7648bd7b4ab102272225dd1785a6`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
