# fill-detection Specification

## Purpose
체결 감지의 권위 소스(공식 Open API 주기 폴링)와 SSE 힌트 규약, per-fill ID가 없는 브로커 모델에 대응하는 누적 스냅샷 멱등 반영 요구사항을 정의한다.

## Requirements

### Requirement: 폴링이 체결 감지의 권위
체결·미체결·잔고 상태의 권위 소스는 공식 Open API 주기 폴링이어야 하며(SHALL), WTS 세션이 만료되어도 체결 감지는 중단 없이 동작해야 한다(SHALL). 폴링 대상은 최소 미체결 주문 목록(pagination 완주), 주문 상세(OrderByID), 잔고·매수가능금액이다. 신선도 SLO는 "브로커에서 관측 가능해진 체결 → 로컬 durable 반영 커밋"으로 측정점을 정의하고(SHALL), 측정 window·percentile과 위반 시 신규 진입 차단 조건을 수치로 정한다. 429·장애 구간은 outage 상태로 별도 분류한다.

#### Scenario: WTS 세션 만료 중 체결 발생
- **WHEN** WTS 세션이 만료된 상태에서 주문이 체결되면
- **THEN** 공식 API 폴링이 SLO 이내에 체결을 감지하고 journal 상태를 갱신한다

#### Scenario: SLO 위반 지속
- **WHEN** 체결 감지 지연이 정의된 임계를 초과하면
- **THEN** 신규 진입이 차단되고 복구 시 자동 해제된다

### Requirement: 누적 스냅샷 기반 멱등 반영
공식 API는 per-fill 식별자를 제공하지 않으므로(누적 filledQuantity·평균가·filledAt만), 체결 반영은 주문(lineage 노드) 단위 누적 스냅샷 모델을 사용해야 한다(SHALL): 직전 관측 대비 양(+)의 filledQuantity delta만 신규 체결로 반영하고, 감소·역순 스냅샷은 UNKNOWN_BROKER_STATE로 fail-closed 처리한다(SHALL). 평균가 갱신은 스냅샷 교체로 처리하며 중복 반영을 만들지 않는다. 동일 스냅샷의 중복 수신(폴링·SSE 재조회·재시작 후)은 상태를 변경하지 않는다(SHALL).

#### Scenario: 동일 스냅샷 중복 수신
- **WHEN** 같은 filledQuantity 스냅샷이 폴링과 SSE 재조회에서 각각 도착하면
- **THEN** 체결 반영은 한 번만 일어난다

#### Scenario: filledQuantity 감소 관측
- **WHEN** 이전 관측보다 작은 filledQuantity가 도착하면
- **THEN** UNKNOWN_BROKER_STATE로 처리되어 해당 심볼이 차단되고 알림이 발송된다

### Requirement: SSE는 지연 단축 힌트
WTS SSE 이벤트는 즉시 재조회를 촉발하는 힌트로만 사용되어야 하며(SHALL), 이벤트 페이로드를 상태 변경 근거로 사용해서는 안 된다(SHALL NOT). 이벤트 기반 재조회는 토픽별 coalescing(single-flight)과 최소 간격 제한을 적용한다(SHALL).

#### Scenario: 이벤트 폭주
- **WHEN** 같은 토픽의 SSE 이벤트가 짧은 시간에 다수 도착하면
- **THEN** 재조회는 진행 중 1건으로 합쳐지고 최소 간격이 보장된다
