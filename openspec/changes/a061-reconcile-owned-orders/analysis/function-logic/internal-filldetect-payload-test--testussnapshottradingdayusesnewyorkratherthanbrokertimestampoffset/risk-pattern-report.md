# Risk Pattern Report: `TestUSSnapshotTradingDayUsesNewYorkRatherThanBrokerTimestampOffset`

Source: `internal/filldetect/payload_test.go`  
AST source SHA-256: `2a3179003b761f34a7ba63d94ba7f3c439689cc48bb00c04a23837a23f97fa9a`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.
