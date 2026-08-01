# Review: a049-add-lane-performance

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Findings and decisions

1. authoritative lineage/outcome은 journal에, high-volume derived observation은 별도 `performance.db`에 둔다.
2. a046 markout 계약과 기존 관측만 재사용하고 추가 polling은 0건이다.
3. raw retention 90일, 24시간마다 최대 500 rows/transaction, 100ms lock 목표를 고정했다. 1,000,000-row fixture의 최근 30일 query p95 목표는 250ms다.
4. collector/query는 broker mutation, config write, lane/LIVE capability를 갖지 않는다.

## Verification evidence

- OpenSpec strict validation: pass.
- Query semantics and missing states are versioned and explicit.

## Verdict

a048 이후 read-only analytics 구현을 승인한다. load/prune/no-extra-poll evidence가 gate 조건이다.

## 2026-08-01 implementation response to independent BLOCK

- SELL markout now applies the same explicit side convention as slippage and MFE/MAE before costs; 5/15/30 gross, cost-adjusted value and exact source/time/version are tested together.
- Trade, observation and snapshot identities use compare-and-append: exact replay is an idempotent skip and divergent bytes return `ErrImmutableConflict`. Restart, exact concurrency and divergent concurrency preserve one complete immutable collection.
- Migration hooks after DDL/version and collection hooks after trade/observations/snapshot are subprocess-SIGKILL tested as all-or-none.
- Retention deletes at most 500 rows per immediate transaction, performs at most four transactions per run, does not publish cadence while backlog remains, and immediately reschedules a late overdue append.
- Dashboard rejects periods over 90 days and more than 10,000 trades, resolves latest snapshots only for a 10,001-row sentinel filtered CTE, and applies identical market/lane/complete predicates to state counts.
- The 1,000,000-raw-row test now executes and explains the actual Dashboard SQL for 25 runs; it requires the trade-window and latest-snapshot indexes and refuses raw/global snapshot scans. The former COUNT proxy was removed.
- Normal, race, full, vet, Windows cross-build, strict OpenSpec and diff checks pass in the a049 worktree. Final independent re-review, post-a045+a047 journal adapter, approved sync/base-marker update and gate remain coordinator-owned.
