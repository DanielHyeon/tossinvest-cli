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

## 2026-08-01 exact review remediation (implementation evidence; rereview pending)

- HIGH chronology finding accepted: wall `ObservedAt > completedAt` comparison을
  제거하고 request 시작 시점의 opaque one-shot observation cycle과 monotonic
  completion watermark를 도입했다. held pre-completion response, wall rollback,
  manual observation, forge/replay/cross-scope/generation을 fail-closed로 고정했다.
- MED issued-memory finding accepted: reported `MaxInt`와 독립된 endpoint/reset
  generation당 256 commitment 발급 상한을 두고 same-window reconcile은 issued
  set을 보존하며 proven reset만 초기화한다. safety class는 상한 밖이다.
- MED delta-reset finding accepted: delta raw/derived semantics, fixed anchor
  1초 tolerance, conservative earliest deadline과 latest-plausible-boundary proof를
  분리했다. epoch identity는 exact다.
- Focused scheduler tests: GREEN. 전체/race/vet/strict/sdd 결과와 독립 exact
  rereview는 이 implementation pass 뒤 기록한다. 이 절은 3.4/3.5 승인이나
  production activation authority를 만들지 않는다.

## 2026-08-01 second exact review remediation (implementation evidence; rereview pending)

- HIGH generation-authority finding accepted: `budgetWindowNext`는 commitment
  map이 비어 있어도 valid nonnil observation cycle 없이는 generation, issued
  commitment 기억 또는 observation-cycle 상한을 초기화하지 않는다. 256개
  issued가 모두 same-window reconcile된 뒤 manual 새-window observation을
  주입하는 우회 회귀를 RED/GREEN으로 고정했다.
- HIGH parser-equivalence finding accepted: official reset parsing을 pure
  `ParseRateBudgetReset` helper로 단일화하고 scheduler가 canonical raw, exact
  `1_000_000_000` kind threshold, derived instant, inclusive `[-1m,+24h]`
  plausibility를 그대로 검증한다. wrapping integer, raw-kind mismatch,
  boundary와 implausible epoch를 고정하고 delta tolerance의 `MinInt` absolute
  negation을 ordered bounds로 제거했다.
- Focused official/scheduler tests: GREEN. 전체/race/vet/strict/sdd 결과와 동일
  reviewer exact-SHA rereview는 이 implementation pass 뒤 기록한다. tasks
  3.4/3.5와 gate 상태는 바꾸지 않는다.
