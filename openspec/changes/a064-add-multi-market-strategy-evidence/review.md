# Review — a064-add-multi-market-strategy-evidence

- Date: 2026-08-03
- Stage: proposal freeze; production implementation not started
- Voices: Manager scope/safety review, independent data/engineering/security review, final semantic re-review

## Findings and disposition

- **Accepted:** historical queries need both `source_available_at <= evaluation_at` and
  `ingested_at <= ingestion_cutoff`; effective dates cannot stand in for source availability.
- **Accepted:** high-volume source payloads belong in append-only `evidence.db`; the trading journal
  stores only the consumed immutable snapshot ID/digest so ingestion cannot contend with exit work.
- **Accepted:** revision identity excludes payload digest. A different digest for the same authoritative
  revision is quarantined as `SOURCE_REVISION_CONFLICT`, not stored as a second valid revision.
- **Accepted:** official source policy freezes endpoint/schema, absolute request window, pagination,
  bytes/concurrency/deadline/retry and secret sanitation. Missing policy means disabled/zero calls;
  KRX without an official programmatic contract remains `SOURCE_UNAVAILABLE`.

## Verification

- Strict OpenSpec validation: PASS.
- Final independent semantic re-review: PASS, no open blocker.
- KR/US failure scopes remain independent; no broker, LIVE approval or operating-toggle authority added.

## Verdict

Proposal freeze approved for RED implementation. Source credentials and numeric production budgets remain
deployment inputs; their absence keeps the affected source disabled and is not an inferred approval.
