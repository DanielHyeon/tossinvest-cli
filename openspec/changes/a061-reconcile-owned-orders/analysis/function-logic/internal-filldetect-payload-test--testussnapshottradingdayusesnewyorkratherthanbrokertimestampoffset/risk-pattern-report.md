# Risk Pattern Report: `TestUSSnapshotTradingDayUsesNewYorkRatherThanBrokerTimestampOffset`

Source: `internal/filldetect/payload_test.go`  
AST source SHA-256: `efd2573ee79ff75e82219f73f6fdc135d5491b408ab180b28237e4b710d9061c`

- Risk class: journal/reconciliation ownership, persistence ordering, or verification helper.
- Live broker mutation introduced: none.
- Primary hazards reviewed: identifier reuse, cross-account/market/day leakage, legacy empty-scope ambiguity, non-atomic release, stale runtime projection, and reservation/nonce evidence loss.
- Controls: canonical scoped evidence, confirmed AMEND lineage validation, durable fail-closed conflict, transaction boundaries, engine lock, stable official snapshots, and targeted plus race/full-suite tests.
- Residual risk: broker evidence can still disagree legitimately; the system retains the reconciliation block for explicit operator review instead of guessing.
