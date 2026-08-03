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
