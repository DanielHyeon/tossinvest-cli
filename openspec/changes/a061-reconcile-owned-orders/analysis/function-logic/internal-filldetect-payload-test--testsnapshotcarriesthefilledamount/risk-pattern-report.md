# Risk Pattern Report: `TestSnapshotCarriesTheFilledAmount`

Source: `internal/filldetect/payload_test.go`  
AST source SHA-256: `14be87b64e23ea392eff45b6daf15ee72d8307260b6d46185aac6ed748d5d9c2`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
