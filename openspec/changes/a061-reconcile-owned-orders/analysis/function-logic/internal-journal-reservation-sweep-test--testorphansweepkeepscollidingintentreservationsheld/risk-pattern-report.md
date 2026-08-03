# Risk Pattern Report: `TestOrphanSweepKeepsCollidingIntentReservationsHeld`

Source: `internal/journal/reservation_sweep_test.go`  
AST source SHA-256: `489ced650b9f96b30419ecc9c21f7911145af1a973372a7f167edb61cfcbdcab`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
