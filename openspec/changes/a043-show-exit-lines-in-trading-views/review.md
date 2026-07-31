# Review: a043-show-exit-lines-in-trading-views

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability, UI/UX

## Findings and decisions

1. `ExitLineView`는 console 소유가 아니라 `internal/operatorview`의 transport-neutral DTO다.
2. 화면은 a041/a042 snapshot을 재계산하지 않고 decision ID로만 order를 연결한다. symbol/time 근사 join은 금지한다.
3. StockOS의 contextual navigation을 참고하되 거래 화면은 read-only이며 입력 control을 0개로 유지한다.
4. 360px, keyboard/ARIA, CSP, stale/unknown/1주/broker-only order를 증거로 남긴다.

## Verification evidence

- OpenSpec strict validation: pass.
- Mutation capability: none by contract.
- Dependency baseline: implementation starts from `70aabdc`, after a041/a042 were
  integrated and gated. `base-commit.txt` was advanced from the portfolio-planning
  commit so a043's Function Logic Map and diff gate measure this change rather than
  attributing its prerequisite snapshots to this UI change.

## Verdict

a042 이후 구현을 승인한다. canonical DTO, no-recompute/no-fuzzy-link와 input-free 렌더 검증이 gate 조건이다.
