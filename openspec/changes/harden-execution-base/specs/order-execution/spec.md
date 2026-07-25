# order-execution Specification (delta)

## ADDED Requirements

### Requirement: 주문 상태기계
주문 생명주기는 명시적 상태기계로 관리되어야 하며(SHALL), 상태 정의의 정본은 공식 Open API 주문 status 스키마다. 상태는 최소 INTENT_RECORDED → SUBMITTING → (SUBMITTED | IN_DOUBT) → PENDING → PARTIALLY_FILLED → (FILLED | CANCELLED | AMENDED | REJECTED)를 포함하고, 정정 결과 불명은 AMEND_IN_DOUBT로 표현해야 한다(SHALL). 허용되지 않은 전이는 오류다.

#### Scenario: 허용되지 않은 상태 전이 시도
- **WHEN** FILLED 상태의 주문에 SUBMITTING 전이를 적용하면
- **THEN** 전이가 거부되고 오류가 반환된다

#### Scenario: 공식 API 응답 fixture 계약
- **WHEN** 공식 API 주문 조회 응답 fixture로 상태 매핑 테스트를 실행하면
- **THEN** 모든 공식 status 값이 상태기계의 상태로 결정적으로 매핑된다

### Requirement: Durable Intent Journal
모든 주문 mutation(신규·취소·정정)은 브로커 제출 **이전에** 영속 journal에 의도(intent key, canonical 문자열, 시각, 상태)를 기록해야 한다(SHALL). journal 기록에 실패하면 제출하지 않는다(SHALL NOT). journal 저장 경로는 저장소 밖 로컬 데이터 디렉터리이며, fuseblk 파일시스템 경로가 지정되면 기동을 거부해야 한다(SHALL).

#### Scenario: 제출 직전 크래시 후 재시작
- **WHEN** journal에 SUBMITTING으로 기록된 intent가 있는 상태로 프로세스가 재시작되면
- **THEN** 해당 intent는 IN_DOUBT로 표기되고 브로커 조회로 확정되기 전까지 동일 intent의 재제출이 차단된다

#### Scenario: fuseblk 경로 지정
- **WHEN** journal 경로가 fuseblk 마운트를 가리키면
- **THEN** 기동이 명시적 오류와 함께 거부된다

### Requirement: 주문 mutation 재시도 금지
주문 mutation 호출이 타임아웃·5xx·연결 오류 등 결과 불명으로 실패하면 자동 재시도해서는 안 되며(SHALL NOT), intent를 IN_DOUBT로 표기하고 주문·체결·거래내역 조회로 실제 결과를 확정한 뒤에만 후속 조치를 해야 한다(SHALL). IN_DOUBT가 해소되기 전까지 해당 심볼·레인의 신규 진입은 차단된다(SHALL).

#### Scenario: 제출 타임아웃
- **WHEN** 주문 제출 요청이 타임아웃되면
- **THEN** 재제출 없이 IN_DOUBT로 기록되고, 조회로 주문 존재가 확인되면 SUBMITTED로, 부재가 확인되면 FAILED로 확정된다

#### Scenario: 시장가 주문의 즉시 체결 확인
- **WHEN** IN_DOUBT 확정 조회를 수행하면
- **THEN** 미체결 목록뿐 아니라 체결·거래내역도 함께 조회하여 이미 체결된 주문을 탐지한다

### Requirement: Retry Matrix
HTTP 재시도 정책은 endpoint 유형별로 분리 정의되어야 한다(SHALL): 조회는 bounded jitter backoff로 제한 재시도, 주문 mutation은 재시도 금지, rate limit(429)은 유형별 대기 후 조회만 재개. 필수 상태 조회의 staleness가 임계값을 초과하면 신규 진입을 차단해야 한다(SHALL).

#### Scenario: 조회 429 응답
- **WHEN** 시세 조회가 429를 받으면
- **THEN** bounded jitter로 재시도하고 주문 경로에는 영향을 주지 않는다

### Requirement: Fail-closed 분기
공식 주문 경로에서 interactive auth challenge가 요구되거나, USD 주문에 필요한 통화 잔고가 부족하거나, 지원하지 않는 주문 유형이 감지되면 주문을 거부하고 알림을 발송해야 한다(SHALL). 자동 환전·자동 승인 시도는 금지된다(SHALL NOT).

#### Scenario: USD 잔고 부족 매수
- **WHEN** USD 매수 intent에 대해 사전 통화 잔고 확인이 부족을 반환하면
- **THEN** 주문은 제출되지 않고 거부 사유가 기록·통지된다
