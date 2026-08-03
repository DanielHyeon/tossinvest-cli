# Status — a064-add-multi-market-strategy-evidence

- Date: 2026-08-04
- State: implementation in progress
- Completed in evidence-source slice: frozen SEC/OpenDART fixture contracts, KRX zero-call unavailable gate,
  sealed deployment-policy minting, bounded official pagination, shared rate/concurrency budget, secret boundary,
  immutable complete-batch commit gate, trusted ingestion clock and adversarial evidence-core hardening.
- Completed in Wave 1C: v21 nullable snapshot-ID/digest-only journal lineage, exact immutable replay,
  dormant SELECT-only read ports, full-Header snapshot digest binding, scope/dual-cutoff replay revalidation,
  database-level partial-reference guards, and independent KR/US failure scope without runtime activation.
- Focused verification: a064 journal/evidence race tests PASS; full non-race journal/evidence package tests PASS;
  `go vet ./internal/journal ./internal/strategyevidence` PASS; strict OpenSpec validation PASS.
- Still open: repository-wide `make sdd-check`/test/vet/validate/gate and final independent review. The full
  journal race suite exceeded its 10-minute test timeout while still executing SQLite migrations; targeted
  a064 race coverage passed and no race report was emitted.
- Safety: no broker/order/toggle path was added; no LIVE authority was granted; KRX performs zero calls.
