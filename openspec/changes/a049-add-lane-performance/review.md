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
- Normal, focused race, full, vet, Windows cross-build, strict OpenSpec and diff checks passed for the pre-integration implementation; the final post-a045+a047 evidence is recorded below.

## 2026-08-01 final integration and independent re-review

- Upstream a045/a047 was merged at `a286c898cb362f2bf86bf378d4051e250c403b57`; `base-commit.txt` was refreshed to that exact post-integration baseline so a049 evidence does not claim earlier changes as its own.
- Journal schema v15 adds nullable `cost_total` without a default or legacy backfill, freezes future exact costs, rejects authority-row UPDATE while retaining bounded DELETE, and covers migration rollback plus the before-version, after-version and post-commit SIGKILL boundaries.
- `internal/performancejournal` is the single-method SELECT-only bridge from `journal.ReadOnly`. Outcome-anchored joins preserve legacy rows, require exact mutation/order/fill/position/close cardinality, reject cross-account or corrupted risk bindings, and label zero/multiple fills `link_missing` without proximity guessing.
- The derived store distinguishes unknown SQL NULL cost from measured literal zero, fails closed on malformed persisted decimals, binds a store to one server-selected journal account, and keeps `performance` free of journal/broker/config write dependencies.
- The performance-history route remains GET/HEAD-only. Regression tests reject input, textarea, select, form, button, contenteditable, POST/action/hx-post/data-method, order-place, lane-toggle, LIVE approval and settings-apply controls.
- Implementation ownership was split across the journal v15 worker (`66797ff`, `d6090d0`), logic-map evidence worker (`14a7c40`) and coordinator bridge/read-model integration (`7d4f361`). Diffs were reviewed against `a286c898` before integration.
- Independent security review: PASS, no actionable findings. Independent test/quality review: PASS, adequate acceptance coverage and no actionable findings. Independent backend/maintainability review: PASS, no blockers.
- Non-blocking follow-up constraints are recorded: production wiring must use `performancejournal.New(*journal.ReadOnly)` rather than low-level derived writes; a future cursor must document per-trade partial success; any post-release derived schema change requires v2 or an explicit rebuild path. a051 wiring tests will pin the first constraint.

### Verification evidence

- `go test ./... -count=1`: PASS after adding the documented read-only lineage SQL exception to the centralized exit-eligibility spelling guard.
- Focused normal tests for journal/performance/performancejournal/console/position: PASS.
- Focused race for all a049 journal migration/lineage/outcome paths and full race for performance/performancejournal/console/position: PASS.
- Package-wide journal race exceeded its fixed 10-minute package timeout with no race diagnostic; it is not represented as a pass. Changed paths pass focused race.
- `go vet ./...`: PASS.
- Windows cross-compile via `GOOS=windows CGO_ENABLED=0 go test -exec=true ...`: PASS.
- strict OpenSpec validation, Function Logic Map analysis, `git diff --check`, `make sdd-sync` and `make sdd-check`: PASS.

## Final verdict

Approved for the a049 change gate. LIVE/order/config mutation authority remains absent and no arbitrary UI input was introduced.
