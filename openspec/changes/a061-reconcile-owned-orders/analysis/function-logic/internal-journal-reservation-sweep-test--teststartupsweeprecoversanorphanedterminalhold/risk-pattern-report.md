# Risk Pattern Report: `TestStartupSweepRecoversAnOrphanedTerminalHold`

Source: `internal/journal/reservation_sweep_test.go`  
AST source SHA-256: `40bf008737b54e15935e4ad2855e1c09d2f42fd3b6a6f975efd9e2d2e074d7cc`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
