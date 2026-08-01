# a049 performance data plan

## Authority and access pattern

- Authoritative order, fill, position, close, and audit rows stay in `journal.db`.
- `performance.db` is rebuildable derived data. It cannot delete or update journal rows.
- The a047 handoff is the typed `JournalLineageReader.ClosedStrategyTrades` seam. Its future adapter must join persisted IDs only. Symbol/time proximity is not part of the interface.
- a049 adds zero quote/API polls. `CollectClosedStrategyTrades` receives caller-owned `map[position_id][]Observation` values.
- Read/write shape is append-heavy raw observation ingestion plus read-heavy 30-day aggregates. SQLite WAL permits readers during bounded writes; four connections allow concurrent reads while SQLite serializes writes.

## Entity relationship

```text
performance_trades (1) ──< measurement_snapshots (1) ──< metric_observations
        │
        └── position_id ──< price_observations

journal.db (authoritative, separate file)
candidate → lane provenance → decision → attempt → opaque broker order
          → entry fill event → position → trade_outcome close row
                    │ exact typed adapter after a045+a047 schema integration
                    └──────────────────────────────> performance_trades
```

`performance_trades` keeps each optional lineage component nullable and records `complete` or `link_missing`. A missing component is never filled from symbol or time. `price_observations` and measurement/metric rows are append-only through public APIs.

## Query and index justification

| Query | Index | Reason |
|---|---|---|
| 90-day bounded raw prune | `idx_price_observations_at(observed_at,id)` | covering range scan and deterministic 500-row transaction order |
| One position's 5/15/30 markout and MFE/MAE | `idx_price_observations_position_at(position_id,observed_at,id)` | exact position equality followed by time range/order |
| Default 30-day complete-lineage dashboard | `idx_performance_trades_window(closed_at,lineage_status,market,lane_id,...)` | bounded close window is always present; remaining fixed/default filters are covered |
| Lane/policy drilldown | `idx_performance_trades_lane(lane_id,lane_version,policy_id,policy_version,closed_at)` | equality on immutable versions followed by close range |
| Exact replay / latest append-only measurement | `UNIQUE(trade_id,calculated_at,semantics_version)` plus `idx_measurement_snapshots_trade(trade_id,id)` | compare-and-append identity and highest append identity per already-filtered trade without updating a prior snapshot |

The prune suite verifies deterministic bounded deletes. Dashboard performance is measured through the actual joined read path described below, not a raw-row COUNT proxy.

### Automated schema/index review

The database-designer analyzers were run against the checked-in DDL and the three hot query patterns. Their raw outputs are `performance-schema-analysis.json` and `performance-index-analysis.json`; `performance-index-schema.json` is the optimizer's required cardinality-bearing input.

- The DDL parser reports no foreign keys and misses the composite primary key on `metric_observations`, although the production `schemaV1` has both `REFERENCES` clauses and `PRIMARY KEY(snapshot_id, metric_key)`. This is a parser limitation, not an accepted production gap.
- The reported lane/version, decision fields, and policy/version 3NF issues are deliberate immutable provenance snapshots in a rebuildable cross-database read model. Normalizing them into local authority tables would create a second source of truth and complicate deterministic replay.
- Positive `entry_price`, `quantity`, observation `price`, non-negative costs, side, lineage status, metric key, and metric status are enforced by typed Go validation and/or production SQLite `CHECK` constraints. Decimal strings remain `TEXT` so SQLite floating-point coercion cannot silently change accounting values.
- The index optimizer proposed four indexes. Three are exact-prefix or exact-column duplicates of `idx_price_observations_at`, `idx_price_observations_position_at`, and `idx_performance_trades_window`; adding them would only increase append cost. The fourth `(position_id,observed_at)` is a strict prefix of the existing covering `(position_id,observed_at,id)` index. All four are rejected as redundant. The authoritative test runs the actual bounded Dashboard SQL—not a COUNT proxy—against a database containing 1,000,000 raw rows, asserts both the trade-window and per-filtered-trade latest-snapshot indexes in `EXPLAIN QUERY PLAN`, refuses global snapshot/raw scans, and measures p95 <=250ms.
- The generic schema analyzer reports the deliberate denormalized lane/decision/policy columns as 3NF candidates. They remain embedded because this is a rebuildable immutable read model keyed by exact source versions; normalizing them would add joins without creating new authority. Its composite-primary-key parser also reports `metric_observations` as missing a key even though the DDL and runtime schema define `PRIMARY KEY(snapshot_id,metric_key)`. Exact rational quantity/price positivity is enforced before every public append; SQLite TEXT is retained to avoid lossy REAL checks.

## Retention, migration, and recovery

- Raw derived observations: 90 days.
- Prune cadence: once per 24 hours after the cutoff backlog is drained. Each `BEGIN IMMEDIATE` transaction deletes no more than 500 rows; a run performs at most four transactions and reports remaining backlog for immediate reschedule. `last_pruned_at` is written only in the transaction that proves no expired row remains. A later append older than that completed cutoff clears the marker in its compare-and-append transaction, so sustained backfill cannot lock cleanup out for 24 hours. Tests enforce the per-transaction 100ms writer-lock target.
- Derived schema v1 creation and `user_version` update share one transaction. A failed migration leaves version 0 and no partial tables. Subprocess SIGKILL hooks after DDL and after version assignment prove all-or-none migration; append hooks after trade, observation and snapshot phases prove all-or-none collection.
- Derived DB rollback is stop collector → preserve/move file for diagnosis → rebuild from authoritative evidence. It never rolls journal or audit rows back.
- Journal schema wiring is deliberately deferred until a045's v13 and a047 provenance schema are integrated. No placeholder v14/v15 is reserved here. The integration must allocate the actual next version and implement the exact adapter before task 2.1 is complete.

## Consistency, security, and monitoring

- Exact immutable replay is an idempotent skip. A duplicate identity with divergent canonical bytes fails closed with `ErrImmutableConflict`; restart and concurrent tests prove one complete collection, no overwrite, and no partial observation/snapshot rows.
- Dashboard periods are bounded to 90 days and the filtered CTE is bounded to 10,001 sentinel trades. More than 10,000 fails closed instead of silently truncating or allocating unbounded memory. The latest-snapshot lookup is correlated only over that filtered set.
- DB/WAL/SHM are chmod 0600 and the directory 0700.
- Decimal source values are stored as text; metrics use `big.Rat` and versioned `lane-performance/v1` semantics.
- Monitor: database bytes, raw row count/age, last prune timestamp, prune duration/deleted rows, dashboard p50/p95, `link_missing` and `not_measured` rates, and migration/crash reopen failures.
- Alert when no successful prune occurs for 48h, p95 exceeds 250ms, or missing-state ratios change materially. These are analytical alerts only and must not gate exits or trigger orders.

## Scale path

At the specified 1M-row fixture SQLite remains the chosen local store. If raw volume grows beyond a single-host working set, retain the public append/query contracts and move only the rebuildable derived store; authoritative journal IDs and versioned semantics remain unchanged. No sharding is justified before measured query or retention pressure exceeds the current targets.
