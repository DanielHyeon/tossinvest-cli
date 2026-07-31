# Review: a048-add-market-aware-scheduler

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Findings and decisions

1. official MarketCalendar typed response의 canonical digest를 version으로 사용하고 6시간 freshness/session 전 refresh 실패는 `WAIT_MARKET`다.
2. budget-key별 remaining의 50% 올림과 최소 5 calls 중 큰 값을 safety reserve로 둔다. provenance가 없거나 stale하면 추가 entry/candidate/analytics poll은 0건이다.
3. priority는 emergency exit > reconcile > fill detect > protection > candidate/entry > analytics다.
4. auto-resume은 exact activation manifest 일치 때만 desired를 복원하며 desired field에서 승인을 재구성하지 않는다.

## Verification evidence

- OpenSpec strict validation: pass.
- Default: scheduler OFF, auto-start OFF, market none, regular session only.

## Verdict

dormant scheduler와 deterministic budget/calendar logic 구현을 승인한다. a047 activation prerequisite 전 runtime entry 연결은 금지한다.
