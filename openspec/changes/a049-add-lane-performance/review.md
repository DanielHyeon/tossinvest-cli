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
