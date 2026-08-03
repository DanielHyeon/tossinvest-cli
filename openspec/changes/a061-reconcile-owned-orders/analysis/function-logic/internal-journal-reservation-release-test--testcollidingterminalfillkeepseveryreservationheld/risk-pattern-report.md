# Risk Pattern Report: `TestCollidingTerminalFillKeepsEveryReservationHeld`

Source: `internal/journal/reservation_release_test.go`  
AST source SHA-256: `aa3b949774db057260930c4c4ccfacd2fbf88f15741f24d4476c756d28d592e7`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
