## Context

`internal/clock`은 KR/US 세션 primitive를 제공하지만 candidate cadence와 entry runtime을 통합한 scheduler가 없다.

## Goals / Non-Goals

**Goals:** 시장 시간, desired state와 API 예산을 한 scheduler decision으로 제공한다.

**Non-Goals:** 전략 판단, 주문 가격, exit 중단, 자동 승인.

## Decisions

1. scheduler는 `ENTRY_ALLOWED`, `WAIT_MARKET`, `DISABLED`, `BUDGET_DEFERRED`의 typed decision을 반환한다.
2. official `MarketCalendar` response를 typed adapter로 검증하고 canonical response digest를 calendar version으로 사용한다. fetched-at 6시간 이내이고 대상 session 시작 전 refresh에 성공한 경우만 entry에 사용하며 missing/stale/parse 실패는 `WAIT_MARKET`로 fail-closed한다. IANA timezone을 사용하고 로컬 머신 timezone에 의존하지 않는다.
3. endpoint budget-key별 reported remaining/reset을 중앙 coordinator가 소비한다. 각 reset window에서 `max(5 calls, ceil(remaining의 50%))`를 exit/fill/reconcile/protection safety reserve로 남긴다. 우선순위는 emergency exit > reconcile > fill detection > protection supervision > candidate/entry > analytics이고, budget header가 없거나 stale하면 candidate/entry/analytics 추가 poll은 0건이다.
4. restart desired state에는 actor, approved-at, market scope와 config version을 저장한다.
5. UI 소유 카테고리는 `strategy-runtime`, 하위 section은 `시장·일정`이다. scheduler desired/effective, auto-start desired/effective, market scope, session scope, calendar version/updated-at과 typed decision reason을 표시한다. 범위와 reason은 server-defined choice이며 자유 입력을 받지 않는다.
6. auto-resume은 a047 activation manifest의 scheduler/calendar/approval binding이 현재 값과 정확히 일치할 때만 허용한다. desired state만으로 승인이나 manifest를 재구성하지 않는다.
7. 초기 defaults는 scheduler OFF, auto-start OFF, market none, regular-session only다. calendar는 authoritative adapter가 제공하는 read-only 값이며 수동 휴장일 편집 control은 최초 범위에 두지 않는다.

## Risks / Trade-offs

- [휴장 데이터 stale] → calendar version/updated-at을 노출하고 불확실하면 entry fail-closed다.
- [clock jump] → monotonic ticker와 exchange-time 재평가를 분리한다.
- [OFF·장닫힘·예산대기를 같은 중지로 오인] → desired/effective와 `DISABLED`, `WAIT_MARKET`, `BUDGET_DEFERRED` 설명을 같은 상태 카드에 분리한다.

## Migration Plan

scheduler를 entry-disabled 상태로 배포한 뒤 a047 runtime에 연결한다. rollback은 scheduler loop만 중지하고 exit/reconcile supervisor는 유지한다.

## Open Questions

장전·장후 신규 진입은 최초 범위에서 금지로 고정한다.
