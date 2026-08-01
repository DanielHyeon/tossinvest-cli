## Context

`internal/clock`은 KR/US 세션 primitive를 제공하지만 candidate cadence와 entry runtime을 통합한 scheduler가 없다.

## Goals / Non-Goals

**Goals:** 시장 시간, desired state와 API 예산을 한 scheduler decision으로 제공한다.

**Non-Goals:** 전략 판단, 주문 가격, exit 중단, 자동 승인.

## Decisions

1. scheduler는 `ENTRY_ALLOWED`, `WAIT_MARKET`, `DISABLED`, `BUDGET_DEFERRED`의 typed decision을 반환한다.
2. official `MarketCalendar` response를 typed adapter로 검증하고 canonical response digest를 calendar version으로 사용한다. fetched-at은 요청 시작이 아니라 응답 완료 시각이다. fetched-at 6시간 이내이고 대상 session 시작 전 refresh에 성공한 경우만 entry에 사용하며 missing/stale/parse 실패는 `WAIT_MARKET`로 fail-closed한다. IANA timezone을 사용하고 로컬 머신 timezone에 의존하지 않는다.
3. endpoint budget-key별 reported remaining/reset을 중앙 coordinator가 소비한다. 각 reset window에서 overflow-safe `max(5 calls, ceil(remaining의 50%))`를 exit/fill/reconcile/protection safety reserve로 남기며 음수 또는 remaining>limit 산술은 fail-closed한다. 허용된 low-priority request에는 `crypto/rand` 기반 capability를 발급하고 coordinator, endpoint key digest, poll class와 reset generation에 결합한다. 호출자는 성공/오류/취소와 무관하게 completion을 명시하지만 completion은 capacity를 즉시 복원하지 않고 completed/unreconciled로 유지한다. 각 budget request는 시작 전에 별도의 불투명 one-shot observation cycle을 발급받으며, 이 cycle은 coordinator, endpoint key digest, reset generation과 발급 시점의 monotonic completion watermark에 결합한다. 같은 window response는 자신의 request cycle이 시작되기 전에 이미 완료된 commitment만 reconcile한다. wall `ObservedAt`/`completed-at`, 처리 순서, 초기/manual `Observe`, 위조·재생·cross-key/coordinator/generation cycle은 reconciliation authority가 아니다. 신뢰 가능한 다음 reset도 valid nonnil cycle이 모든 기존 commitment를 account한 뒤에만 generation을 전환하며, commitment가 이미 비어 있어도 manual `Observe`는 generation, issued capability 기억 또는 observation-cycle 상한을 초기화할 권한이 없다. commitment capability 발급은 reported limit와 독립된 endpoint/generation 절대 상한을 가지며, reconcile은 issued 기억을 지우지 않고 proven reset만 지운다. reset raw/kind/instant 검증은 official parser의 단일 read-only helper를 재사용하며, exact epoch threshold `1_000_000_000`, observed-at 기준 inclusive `[-1m,+24h]` plausibility, overflow-safe delta conversion/addition과 raw-kind 일치를 요구한다. delta reset은 첫 derived instant를 고정 anchor로 삼아 1초 이내 subsecond/quantization drift를 같은 window로 분류하되 grant deadline은 가장 이른 값으로만 줄이고, tolerance 비교는 `MinInt`를 negation하지 않는 ordered bound를 사용한다. 다음 generation은 anchor+tolerance 이후의 response가 새 reset을 보고할 때만 만든다. epoch reset identity는 exact instant 비교를 유지한다. 우선순위는 emergency exit > reconcile > fill detection > protection supervision > candidate/entry > analytics이고, budget header가 없거나 stale하면 candidate/entry/analytics 추가 poll은 0건이다.
4. restart desired state에는 monotonic revision, actor, approved-at, market scope와 config version을 저장한다. 저장은 같은 프로필의 cross-process flock 아래 현재 revision과 expected revision을 비교한 뒤 revision+1을 atomic rename하며, 이미 커밋된 운영자 OFF를 stale ON writer가 덮어쓸 수 없다. flock 대기는 caller cancellation을 따르고 최대 2초로 제한한다.
5. UI 소유 카테고리는 `strategy-runtime`, 하위 section은 `시장·일정`이다. scheduler desired/effective, auto-start desired/effective, market scope, session scope, calendar version/updated-at과 typed decision reason을 표시한다. 범위와 reason은 server-defined choice이며 자유 입력을 받지 않는다.
6. auto-resume은 a047 activation manifest의 scheduler/calendar/approval binding이 현재 값과 정확히 일치할 때만 허용한다. desired state만으로 승인이나 manifest를 재구성하지 않는다.
7. 초기 defaults는 scheduler OFF, auto-start OFF, market none, regular-session only다. calendar는 authoritative adapter가 제공하는 read-only 값이며 수동 휴장일 편집 control은 최초 범위에 두지 않는다.
8. `PollClass`와 official reset kind는 닫힌 enum으로 처리한다. 알 수 없는 값은 fail-closed이고, 동일 observed-at의 budget correction은 모든 provenance 필드가 같을 때 remaining을 낮추는 방향만 반영한다. reported/reset/reset-kind/reset-raw/limit 등 provenance가 충돌하면 strictly newer observation 전까지 low-priority poll을 거부한다.
9. official calendar의 선택 date는 빈 값 또는 정확하고 실제 존재하는 Gregorian `YYYY-MM-DD`만 허용한다. production console은 단일 공유 official client의 typed calendar read에서 canonical digest/source/fetched-at을 얻되 입력 control이나 저장 권한을 추가하지 않는다. market scope는 `none`, `KR`, `US`만 허용한다. 정확한 per-market calendar/activation binding이 없는 결합 `KR+US` scope는 descriptor, persistence와 display에서 지원하지 않고 fail-closed한다.

## Risks / Trade-offs

- [휴장 데이터 stale] → calendar version/updated-at을 노출하고 불확실하면 entry fail-closed다.
- [clock jump] → monotonic ticker와 exchange-time 재평가를 분리한다.
- [OFF·장닫힘·예산대기를 같은 중지로 오인] → desired/effective와 `DISABLED`, `WAIT_MARKET`, `BUDGET_DEFERRED` 설명을 같은 상태 카드에 분리한다.

## Migration Plan

scheduler를 entry-disabled 상태로 배포한 뒤 a047 runtime에 연결한다. rollback은 scheduler loop만 중지하고 exit/reconcile supervisor는 유지한다.

## Open Questions

장전·장후 신규 진입은 최초 범위에서 금지로 고정한다.
