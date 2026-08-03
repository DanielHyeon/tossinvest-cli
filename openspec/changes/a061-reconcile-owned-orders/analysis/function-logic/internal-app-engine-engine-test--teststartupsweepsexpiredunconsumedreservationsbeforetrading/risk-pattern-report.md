# Risk Pattern Report: `TestStartupSweepsExpiredUnconsumedReservationsBeforeTrading`

Source: `internal/app/engine/engine_test.go`  
AST source SHA-256: `2ece46493d087d62d38a888ab2a3da4be554ce268f85d8e1ce09b0db18d8e0b1`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
