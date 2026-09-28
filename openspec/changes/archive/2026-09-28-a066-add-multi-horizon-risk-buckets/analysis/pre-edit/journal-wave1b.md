# Journal Wave 1B Pre-Edit Evidence

- Base commit: `23794f8626a20691431d5452b76e800255b0ee74`
- Date: 2026-08-04
- Scope: schema v22 and dormant authoritative journal transaction primitives only.

## SDD sync

`make sdd-sync` reported CodeGraph already up to date. The subsequent advisory CodeGraphContext update
stalled after loading its database and was interrupted. `codegraph status .` then reported an up-to-date
index with 1,465 files, 25,617 nodes and 83,224 edges.

## CodeGraph hard evidence

- `Journal.RecordStrategyDecisionAndReserve` has one direct production caller
  (`RecordStrategyDecisionAndReserveWithRecollection`), calls the existing decision and reservation helpers,
  and has a three-node impact set. It remains unchanged and is not reused as an implicit a066 integration.
- `reserveRows` has three production callers (`Reserve`, `RecordDecisionAndReserve`, and
  `RecordStrategyDecisionAndReserve`) and a 25-node impact set. It remains unchanged; the a066 transaction
  only requires an exact already-HELD reservation reference.
- `Journal.applyMigration` is called only by `Journal.migrate` and has a four-node impact set. The migration
  engine remains unchanged; v22 is an additive migration entry.
- `ApplyPositionCampaignFill` has no indexed callers and a two-node impact set. It remains unchanged so the
  a065 campaign state machine and fill application ordering are preserved.
- Pure `riskbucket.ApplyFill` has test-only callers and a 15-node impact set. The journal adapter invokes it
  as a pure transition and does not edit it.
- Pure `riskbucket.AcquireOwner` has test-only callers and a seven-node impact set. SQLite uniqueness is the
  multi-process authority; the pure owner contract is used for validation semantics without runtime wiring.

## Existing-function Logic Map gate

No existing Go function body will be edited. The only existing production file change is the declarative
`SchemaVersion` constant and migration slice in `internal/journal/schema.go`; all transaction functions and
tests are new a066 files. Therefore no existing-function AST/Function Logic Map/Branch Test Map target exists
for this slice. Branch evidence for new functions is captured by RED/GREEN journal tests.

## Safety boundary

- No Guardian, Gateway, engine, broker, apply-hook, LIVE approval or operating-toggle integration.
- Existing reservation, v21 evidence lineage and v20 campaign rows are referenced but never rewritten.
- The v22 schema reserves append-only fill/event and conservative latch surfaces. Their authoritative
  transaction adapter remains pending and no fill path is wired by this checkpoint.

## RED/GREEN checkpoint

- RED: focused journal tests failed to compile on the absent schema v22,
  `CommitRiskBucketAdmission`, typed snapshot reference and stable replay/owner errors.
- GREEN: focused tests cover atomic migration rollback, legacy unknown/no backfill, journal-owned
  `q_final` calculation, referenced HELD reservation, five-bucket atomicity, exact idempotence,
  partial rollback, concurrent owner arbitration, orphan/snapshot drift and stable replay digest
  mismatch without mutation.
- CodeGraphContext remained stalled after loading its database and was terminated; CodeGraph hard
  evidence and its up-to-date fingerprint were retained because the advisory layer does not gate
  RED/GREEN work.

## Independent source-review RED/GREEN

- RED exposed silent immutable policy reuse, prospective-token-insensitive owner reuse, and the
  fixed event-sequence collision on a second same-owner admission.
- GREEN binds immutable policy/snapshot content to full-record digests, checks exact prospective
  owner identity and assigns owner-scoped monotonic sequences inside the write transaction.
- Commit and replay now call the same DB-derived canonical state builder. Its ordered preimage covers
  every decision/reservation identity, snapshot ID, quantity, HELD/FILLED/overage/state value and
  owner/scope latch; it rejects missing, extra, duplicate or invalid dimensions and 256-bit monetary
  aggregation overflow without repair.

## Final authority-binding hardening

- RED proved owner KR/US and symbol identities were not cross-checked against their corresponding
  bucket values, and reference digest/time fields could diverge from the sealed evidence actually
  consumed by `CalculateAdmission`.
- The pure package now exposes only a by-value `BucketEvidenceBinding` view. It copies private sealed
  evidence and snapshot binding fields; callers still cannot construct or mutate provenance seals.
- Journal preflight requires exact owner market/symbol bucket identity and exact policy/snapshot
  authority source, version, digest, observation/freshness interval and amount binding. Every failure
  occurs before owner, decision, policy, snapshot or reservation writes.

## Idempotence preimage hardening

- RED constructed a retry with the same transaction/reference/evidence digest and the same
  availability, caps and `q_final`, while changing snapshot limit/FILLED from `100/0` to `110/10`.
  The old preimage incorrectly returned idempotent success.
- GREEN adds the canonical ordered `BucketEvidenceBinding` list to the request preimage. It contains
  exact key, limit, FILLED, HELD, snapshot version and both sealed policy/snapshot evidence copies,
  including source/version/digest/times. The divergent retry now returns
  `ErrRiskBucketReplayMismatch` and writes no additional row.
