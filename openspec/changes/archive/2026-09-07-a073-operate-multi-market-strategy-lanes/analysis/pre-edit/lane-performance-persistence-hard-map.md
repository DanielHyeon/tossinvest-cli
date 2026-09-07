# Lane-performance persistence/read adapter hard map (pre-edit)

- Base: `425bf8dbf2f02fd40af3743e08d1d830a32e70c0`
- Owned code: `internal/performance`, `internal/performancejournal`
- Forbidden code: `internal/journal`, execution gateway, engine, strategy flow

## Hard evidence

1. `performance.Open` delegates all versions to `openWithSchema`; `migrate` executes one supplied schema when
   `user_version < targetVersion`. Re-running the released v1 DDL against a v1 database is not a valid v2 migration.
2. `performance.OpenReadOnly` requires an exact `SchemaVersion` and opens an immutable, query-only snapshot only
   after WAL sidecars are absent.
3. `performance.BuildDerivedAttributionStore` is the authoritative pure validator/projector for exact composite
   KR/US lineage, fill corrections, cost/FX evidence and quantity/basis conservation.
4. `journal.ReadOnly.ClosedStrategyTradeSources` exposes closed outcomes and optional exact legacy strategy lineage.
   `PositionCampaignLineage` and `PositionCampaign` expose campaign identity and lane/version, but the read-only
   API exposes neither campaign-leg/fill-event history nor fee/tax/FX evidence. The adapter must therefore preserve
   these absences as `link_missing`/`not_measured`; it must not synthesize zeros, current FX or symbol/time joins.

## Function Logic Map: `performance.Open` / migration dispatch

Input path → validate/secure SQLite writer → read `user_version` →

- newer than build: typed refusal, no write;
- v0: apply released v1 DDL atomically, record v1;
- v1: apply additive attribution v2 DDL atomically, record v2;
- v2: no-op;
- any failed phase: rollback both schema rows and version.

Branch tests: fresh v2, v1→v2 preservation, v2 reopen, too-new refusal, injected v2 DDL failure rollback.

## Function Logic Map: attribution rebuild persistence

Input rebuild → validate account/rebuild identity and bounds → normalize and deterministically sort/deduplicate evidence →
pure-project exact positions/fills → append explicitly unavailable rows only with exact observed lineage → reject duplicate
composite keys → canonical encode/digest source evidence and complete row envelopes → immediate transaction → compare current
head →

- same rebuild ID + same digest: revalidate the complete persisted generation, then idempotent replay;
- same rebuild ID + different digest: divergent immutable replay refusal;
- new rebuild ID: insert complete new generation, atomically move account head, delete prior rebuild;
- any failure/crash: rollback, retaining the previous complete head.

Queries bind exact `market+ticker` and optional lane/version/campaign/leg; ticker without market is refused. Before the
bounded result query, the read snapshot scans the account generation and verifies head metadata, row count, contiguous
ordinals, rebuild identity, every indexed shadow field, status, canonical JSON and envelope digest. This is intentionally
O(account generation rows). Raw authoritative evidence is bounded at 1,000,000 records, while the projected account
generation is independently capped at 10,000 rows so corruption cannot hide outside a result window and every integrity
scan has a finite operational budget. KR and US rows with the same ticker cannot collide. The attribution-specific
10,000-row performance gate must pass before release; integrity is not weakened to meet a latency target.

Branch tests: complete persistence/reopen/read-only query, same replay, divergent replay, crash hook atomicity,
same-ticker KR/US isolation, explicit missing states, maximum input/query bounds, and raw-observation prune preservation.

## Function Logic Map: `performancejournal.Reader.AttributionEvidence`

Closed source → copy exact persisted identifiers → query exact position campaign lineage → if KNOWN, require exact
account/market/symbol/campaign and actual Position generation before copying campaign and lane/version; a mismatch is a
typed provenance conflict → because leg/fill/cost/FX facts are unavailable on the SELECT-only source, emit one unavailable attribution
row carrying only known identifiers/frozen outcome values and explicit missing-field lists. Legacy/none/error states never
become inferred campaign lineage. Source errors propagate and the adapter retains no writer authority.

Branch tests: known campaign enrichment, legacy/no campaign status, nil cost not zero, no FX invention, source error,
and reflection guard proving the source contains SELECT methods only.
