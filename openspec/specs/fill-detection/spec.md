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

### Requirement: Tracked order ownership is proven

The durable tracked-order source SHALL return a non-terminal fill snapshot only when journal evidence proves exactly one local engine owner through a confirmed mutation attempt or recorded replacement lineage in the same account, market-local trading day, symbol, and side. A broker-only observation without that evidence MUST remain external and MUST NOT be followed as a local engine order after it leaves the open list. Ambiguous same-scope ownership MUST fail closed as an identifier conflict without projecting a position, invoking fill hooks, or releasing a reservation.

#### Scenario: External open order disappears
- **WHEN** a broker open order with no local confirmed attempt or lineage is observed, stored, and later absent from the broker open list
- **THEN** it is not returned as a locally tracked order and no OrderByID follow-up is attributed to engine ownership

#### Scenario: Confirmed engine order leaves the open list
- **WHEN** a confirmed engine order or its recorded replacement leaves the broker open list before a terminal snapshot is stored
- **THEN** it remains in the tracked set and is read by identifier until a broker-derived terminal state is durably recorded

#### Scenario: Broker reuses an order identifier on a later trading day
- **WHEN** a previously terminal broker order identifier is observed again on a later market-local trading day in the same account
- **THEN** the later observation starts a new cumulative fill sequence and cannot inherit the prior order's terminal or filled quantity

#### Scenario: Canonical ownership is ambiguous
- **WHEN** more than one local intent claims the same account, market-local trading day, symbol, side, and broker order identifier
- **THEN** the journal records an identifier conflict atomically and performs no local fill projection, hook, or reservation release

#### Scenario: Replacement identifiers are reused outside the selected scope
- **WHEN** another account, market, or trading day has a confirmed amendment with the same parent or child order identifier
- **THEN** tracked orders, reconciliation local state, and live-order cancellation follow only the confirmed amendment in the selected canonical scope

#### Scenario: Legacy and scoped lineage disagree
- **WHEN** validated legacy lineage and schema-v16 scoped lineage name different successors in one canonical scope
- **THEN** resolution records an account-wide identifier conflict and refuses to choose either successor

#### Scenario: Legacy empty-scope evidence meets reused identifiers
- **WHEN** a schema-v15 fill snapshot has no account or trading-day scope and its order or lineage endpoint is reused on another trading day in the same account and market
- **THEN** the snapshot is not attributed to either reused tracked lineage scope, while a terminal snapshot in another market remains authoritative only for its matching market

#### Scenario: Two canonical orders share one opaque identifier
- **WHEN** two engine-owned orders in different canonical market, trading-day, symbol, or side scopes share the same broker order identifier
- **THEN** their cumulative snapshots coexist durably and the detector polls, derives lineage, and applies each scope independently without overwriting or skipping either order

#### Scenario: Reused identifier has fills and corrections in both scopes
- **WHEN** two canonical scopes sharing one opaque identifier each append fills or execution corrections
- **THEN** their event and correction streams, provenance, and trade outcomes remain separated by the complete canonical identity, and an order-id-only compatibility read fails closed as ambiguous

#### Scenario: New observation has only part of canonical scope
- **WHEN** a new observation supplies some but not all of account, market-local trading day, symbol, and side
- **THEN** persistence rejects it without overwriting an existing snapshot or re-emitting a cumulative fill delta

#### Scenario: External evidence predates a future exact owner
- **WHEN** a broker-only snapshot or event is stored before the first confirmed local intent and a later order reuses the same complete canonical identity
- **THEN** the earlier evidence remains external and cannot terminate, track, project, explain, or contribute P&L to the later order

#### Scenario: Two intents claim the same canonical identity
- **WHEN** two confirmed intents name the same account, market-local trading day, symbol, side, and opaque order identifier
- **THEN** legacy migration binds no event or correction, provenance/outcomes attribute no fill, and the existing identifier-conflict gate remains fail closed

#### Scenario: Legacy event has one preexisting confirmed owner
- **WHEN** an append-only blank-scope fill event predates schema v19 and exactly one owner's confirmed transition predates that event
- **THEN** migration leaves the original event unchanged, writes one additive canonical binding, and scoped reads, provenance, and outcomes consume the composed evidence

#### Scenario: Ownership and evidence share a released second
- **WHEN** legacy ownership and evidence have the same second-resolution timestamp
- **THEN** the evidence remains unbound because equality is not causal proof; new confirmations and fill commits serialize a strict durable order so owner-before-evidence remains attributable after restart

#### Scenario: Confirmed cancellation echoes the target identifier
- **WHEN** a confirmed CANCEL response repeats the broker identifier of the PLACE it cancelled
- **THEN** the cancellation is audit history but never a second order owner, and live-order lookup remains available to the emergency exit path

#### Scenario: External cumulative snapshot precedes a reused local order
- **WHEN** an external exact-scope snapshot exists before one local order-creating confirmation and a later observation reports that local order's cumulative fill
- **THEN** the new local sequence starts from zero and applies the full cumulative quantity rather than subtracting the external snapshot

### Requirement: 목록을 떠난 추적 주문은 종결 목록에서 찾는다

체결 감지는 미체결(OPEN) 목록에 없는 추적 주문의 단건 조회 파생이 fail-closed 가 아닌 비종결이면, 그 주문을 종결(CLOSED) 목록에서 찾아야
한다(SHALL). 단건 파생이 종결이거나 fail-closed 이면 종결 목록을 조회해서는 안 된다(MUST NOT). 질의는 종목과 그 주문 거래일 앞뒤 하루의 주문
생성일 창으로 좁히고 페이지 상한과 커서 반복 거절을 지켜야 한다(SHALL). 한 주기의 종결 목록 조회는 최대 한 건이어야 하고 그 조회 전체(페이지
합계)에는 정해진 실시간 시한의 취소 요청이 걸려야 하며(SHALL), 같은 주문은 최소 간격을 지켜야 하고(SHALL) 같은 주기 안에서 재시도해서는 안 된다(MUST NOT).

#### Scenario: 미체결 목록에 있는 주문
- **WHEN** 추적 주문이 미체결 목록에 있거나, 단건 조회가 이미 종결을 돌려주거나, 단건 파생이 fail-closed 다
- **THEN** 종결 목록 질의는 일어나지 않고, fail-closed 단건은 오늘과 같이 처리된다

#### Scenario: 목록 이탈 주문이 여럿이다
- **WHEN** 한 주기에 한 번도 조회되지 않은 후보가 셋이고 후보 집합이 그대로이며 주기가 성공한다
- **THEN** 그 주기에는 하나만 조회되고 나머지는 다음 주기로 미뤄지며, 세 주기에 걸쳐 하나씩 조회된다

#### Scenario: 종결 목록 응답이 멈춘다
- **WHEN** 종결 목록 질의가 실시간 시한 안에 끝나지 않고 목록 조회기가 취소를 따른다
- **THEN** 질의는 시한에서 끊기고 같은 주기의 다른 주문 스냅숏은 그대로 적용된다

### Requirement: 종결 목록 행은 상태만 바꾸며 좁은 조건에서만 받는다

주문번호가 바이트 동일한 종결 목록 행은 raw 바이트가 전부 같을 때만 하나로 세야 한다(SHALL). 그렇게 센 행이 정확히 하나이고, 정규 범위가
추적 범위와 같고, 누적 체결 수량(`execution.filledQuantity`)이 있고 읽히며, 그 행의 파생이 fail-closed 가 아닌 종결이고, 주문 수량과 누적 체결
수량이 단건 조회와 같을 때에만 그 주기 단건 스냅숏의 상태 필드(status·취소 여부·취소 시각·파생)를 그 행의 것으로 바꿔야 한다(SHALL).
수량·평균가·체결금액·체결 시각은 단건 조회의 것을 유지해야 한다(SHALL).

#### Scenario: 앱에서 취소된 주문이 목록을 떠난다
- **WHEN** 추적 중인 엔진 주문이 미체결 목록에 없고 단건 조회는 `PENDING`·취소 시각 없음·체결 0 을 돌려주며, 종결 목록의 같은 주문번호·같은 범위 행 하나가 `CANCELED`·체결 0 이다
- **THEN** 그 주기의 스냅숏은 종결 `CANCELLED` 로 기록되고, 그 주문번호에 빈 범위 legacy 스냅숏 행이 없으면 추적 집합과 대사의 로컬 열린 주문에서 빠진다

#### Scenario: 같은 주문번호가 서로 다른 두 행으로 실린다
- **WHEN** 종결 목록 창에 같은 주문번호를 가진 행이 둘 있고, 정규 범위가 다르거나 범위는 같아도 status·수량이 다르다
- **THEN** 어느 행도 받아들여지지 않고 사유 `closed_list_ambiguous` 가 남는다

#### Scenario: 종결 목록 행에 체결 수량이 없다
- **WHEN** 종결 목록 행은 `CANCELED` 이지만 `execution` 이나 `execution.filledQuantity` 가 없다
- **THEN** 그 행은 받아들여지지 않고 사유 `closed_list_execution_missing` 가 남는다

### Requirement: 종결 목록이 확인하지 못하면 단건 스냅숏을 유지한다

종결 목록이 앞 요구의 조건을 채우지 못하거나 조회되지 않은 주기에는 체결 감지가 단건 조회 스냅숏을 그대로 써야 하며(SHALL), 종결 목록
행으로 체결 수량이나 종결을 기록해서는 안 되고(MUST NOT), 종결 목록 조회의 실패로 주기 전체를 실패시켜서는 안 된다(MUST NOT).

#### Scenario: 종결 목록에서 찾지 못한다
- **WHEN** 미체결 목록에 없는 추적 주문의 단건 조회가 비종결이고 종결 목록 창에도 그 주문번호가 없다
- **THEN** 단건 조회 스냅숏이 그대로 쓰이고 종결 목록 조회가 더하는 종결·체결 기록·진입 래치는 없으며(단건 자체의 체결 증가는 오늘과 같이 처리됨), 주기는 성공하고 사유 `closed_list_not_found` 가 남는다

#### Scenario: 종결 목록과 단건 조회의 체결 수량이 다르다
- **WHEN** 종결 목록 행은 `FILLED` 이고 단건 조회는 `PENDING`·체결 0 이다
- **THEN** 체결은 기록되지 않고 단건 조회 스냅숏이 그대로 쓰이며, 사유 `closed_list_quantity_disagrees` 가 남는다

#### Scenario: 종결 목록을 읽지 못한다
- **WHEN** 종결 목록 질의가 429 또는 전송 실패로 끝난다
- **THEN** 그 주기는 실패하지 않고 미체결 목록·잔고 신선도는 기록되며, 같은 주기 안의 재조회는 없다

### Requirement: 해결되지 않는 목록 이탈 주문은 지속될 때 한 번 알린다

종결 목록 조회 후보인 목록 이탈 추적 주문이 정해진 시간 이상 확인되지 않으면 그 주문(내부 식별자)과 사유를 운영자 알림 통로에 한 번
기록·전송 시도해야 하며(SHALL — 전달 보장은 없음, 일반 등급), 그 알림에 브로커 주문번호·계좌·종목·수량·가격을 실어서는 안 되고(MUST NOT), 그 전에 미체결 목록에 다시 나타나거나
종결되면 알리지 않고 상태를 지워야 한다(SHALL).

#### Scenario: 방금 접수된 주문이 다음 주기에 미체결 목록에 나타난다
- **WHEN** 한 주기에서 종결 목록에서 찾지 못한 주문이 다음 주기 미체결 목록에 나타난다
- **THEN** 운영자 알림은 나가지 않고 그 주문의 미해결 상태는 지워진다

#### Scenario: 해결되지 않은 상태가 이어진다
- **WHEN** 같은 추적 주문이 정해진 시간 이상 종결 목록으로 확인되지 않는다
- **THEN** 그 주문(내부 식별자)과 시장·사유·지속 시간이 알림 통로에 한 번 넘겨지고, 해결될 때까지 다시 넘기지 않는다
