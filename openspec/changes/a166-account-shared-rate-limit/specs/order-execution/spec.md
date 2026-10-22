> proposal 단계 초안 delta. 사용자 결정(2026-10-22)을 요구 문장으로 옮긴 것이며, design 단계에서 기존 요구
> 「Retry Matrix 산출물」의 MODIFIED 로 다시 쓴다(429 처리 문장이 기존 요구의 「Retry-After 상한 준수」와 겹친다 — archive 전에 해소한다).

## ADDED Requirements

### Requirement: 계좌 공유 요청 속도 상한

같은 계좌로 공식 API 를 치는 모든 TossOS 프로세스(엔진·soak·verify·프로브·CLI·콘솔)의 요청 합은 초당 3건을 넘어서는 안 된다(SHALL NOT).
예산의 키는 계좌 식별이어야 하고 config 디렉터리여서는 안 된다(SHALL). 상한은 손절·비상 청산 요청을 지연시켜서는 안 된다(SHALL NOT).

#### Scenario: 두 프로세스의 합산

- **WHEN** 엔진과 soak 이 같은 계좌로 동시에 조회하면
- **THEN** 두 프로세스의 요청 합이 초당 3건을 넘지 않는다

### Requirement: 엔진 폴링 GET 의 429 처리

엔진 폴링 경로의 GET 이 429 를 받으면 같은 주기 안에서 즉시 재시도해서는 안 된다(SHALL NOT). `Retry-After` 가 있고 30초 이하일 때만
그만큼 기다린 뒤 다시 시도할 수 있고, 없거나 30초를 넘으면 이번 조회를 포기하고 staleness 규칙에 맡긴다(SHALL).

#### Scenario: Retry-After 없음

- **WHEN** 엔진의 미체결 조회가 `Retry-After` 없는 429 를 받으면
- **THEN** 그 주기에는 다시 보내지 않고, 조회 실패가 staleness 임계를 넘으면 신규 진입이 차단된다
