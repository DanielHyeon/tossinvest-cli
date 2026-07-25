# order-execution Specification (delta)

## ADDED Requirements

### Requirement: 조건주문 mutation 계약
조건주문의 등록·취소·정정은 일반 주문 mutation과 동일한 계약 아래 수행되어야 한다(SHALL): journal 선기록 → DISPATCH_STARTED → 결과 분류 → 필요 시 IN_DOUBT 해소. 엔진 경로에서 확인 토큰만 검사하고 브로커를 직접 호출하는 제출은 존재해서는 안 된다(SHALL NOT). 조건주문 fingerprint는 계좌·심볼·방향·트리거 가격·수량·주문 유형(SINGLE/OCO)·제출 시각 창으로 구성되고, 해소 조회는 조건주문 목록의 미체결·종결 양쪽을 pagination 완주하며 대조한다(SHALL). 유일 매칭이 불가능하면 재제출하지 않고 UNRESOLVED_IN_DOUBT로 처리한다(SHALL NOT 재제출). CLI/MCP 표면의 기존 확인 토큰 경로는 변경 대상이 아니다.

#### Scenario: 조건주문 제출 응답 유실
- **WHEN** 조건주문 등록 요청 후 응답이 유실되고 프로세스가 재시작되면
- **THEN** DISPATCH_STARTED 상태의 attempt가 발견되어 IN_DOUBT 해소가 시작되고, 해소 전에는 같은 보호 의도의 재제출이 수행되지 않는다

#### Scenario: 해소 조회에서 조건주문 발견
- **WHEN** IN_DOUBT 해소 조회가 fingerprint와 일치하는 조건주문을 종결 목록 2페이지에서 발견하면
- **THEN** 해당 attempt는 CONFIRMED로 종결되고 중복 등록은 발생하지 않는다

### Requirement: 브로커 발동 주문의 체결 귀속
조건주문 발동으로 생성된 브로커 주문의 체결은 해당 포지션에 귀속되어야 한다(SHALL). 조건주문 등록이 확정되면 예상 주문 레코드(조건주문 식별자·심볼·방향·최대 수량·연결 대상)를 journal에 기록하고(SHALL), 체결 감지가 로컬 intent 없는 브로커 주문을 만나면 먼저 예상 주문과 대조한 뒤 매칭 실패 시에만 외부 주문으로 분류한다(SHALL). 발동 주문에 대해 intent를 소급 생성해서는 안 된다(SHALL NOT — 내지 않은 주문에 의도를 부여하면 provenance가 거짓이 된다). 순 보유수량 계산은 이 귀속 규칙을 반영하며, 기존 reconciliation의 외부 주문 판정은 회귀 없이 유지되어야 한다(SHALL).

#### Scenario: 브로커측 손절 발동으로 포지션 청산
- **WHEN** 브로커에 등록된 stop 조건주문이 발동해 전량 체결되면
- **THEN** 그 체결이 예상 주문을 통해 포지션에 귀속되어 순 보유수량이 0이 되고, 영구 불일치로 보고되지 않는다

#### Scenario: 정말로 외부에서 낸 주문
- **WHEN** 예상 주문과 매칭되지 않는 브로커 주문의 체결이 관측되면
- **THEN** 외부 주문으로 분류되고 reconciliation의 기존 처리가 그대로 적용된다

### Requirement: Mutation Safety Class와 직렬화
모든 mutation은 두 safety class 중 하나로 분류되어야 한다(SHALL): EXPOSURE_RAISING(진입 제출, 보호 없는 수량 증가)과 RISK_REDUCING(보호주문 생성·증량, reduce-only 청산, 모든 취소). 직렬화 규칙은 클래스별로 분리된다(SHALL): EXPOSURE_RAISING은 심볼당 1건으로 제한하고, RISK_REDUCING은 EXPOSURE_RAISING의 in-flight 또는 IN_DOUBT 상태에 의해 차단되어서는 안 된다(SHALL NOT — §0.3). RISK_REDUCING끼리는 대상 식별자(조건주문 ID 또는 대상 주문번호) 단위로 직렬화한다. oversell은 차단이 아니라 수량 상한으로 방지한다(SHALL): 미해소 진입이 존재하면 RISK_REDUCING 수량은 확정 보유수량과 계좌 조회 매도가능수량 중 작은 값을 넘지 못한다. 계좌 조회가 staleness 임계를 넘으면 RISK_REDUCING은 계속 허용하되 수량을 확정 보유수량으로만 제한한다(SHALL — 보수 방향).

#### Scenario: 진입 IN_DOUBT 중 손절 제출
- **WHEN** 같은 심볼의 진입 attempt가 IN_DOUBT인 상태에서 보호주문 제출이 요청되면
- **THEN** 제출이 차단되지 않고 수행되며, 수량은 확정 보유수량 이하로 제한된다

#### Scenario: 모호한 진입 중 청산 수량 상한
- **WHEN** 미해소 진입이 있는 심볼을 청산하면
- **THEN** 청산 수량이 확정 보유수량과 매도가능수량 중 작은 값을 넘지 않아 oversell이 발생하지 않는다

### Requirement: 원자적 위험 예약
계좌 전체에 걸친 한도(총 개방 노출, 일일 손실, 현금)의 판정과 그 결과의 예약은 하나의 `BEGIN IMMEDIATE` 트랜잭션 안에서 수행되어야 한다(SHALL). 서로 다른 심볼에 대한 동시 결정이 같은 스냅샷을 각각 통과해 합산 한도를 초과하는 것은 허용되지 않는다(SHALL NOT). 예약은 다음 중 하나에서만 해제된다(SHALL): nonce 소비 후 체결 또는 취소 확정, 결정 만료, 제출 실패 확정. IN_DOUBT 상태의 예약은 해소 전까지 유지된다(SHALL — fail-closed).

#### Scenario: 동시 다심볼 결정
- **WHEN** 총 개방 노출 한도의 잔여분이 1건분만 남은 상태에서 서로 다른 두 심볼의 결정이 동시에 요청되면
- **THEN** 하나만 예약에 성공하고 다른 하나는 한도 초과로 거부된다

#### Scenario: IN_DOUBT 중 예약 유지
- **WHEN** 예약을 소비한 제출이 IN_DOUBT로 남으면
- **THEN** 해소 완료 전까지 예약이 해제되지 않아 그만큼의 노출이 이중 사용되지 않는다

## MODIFIED Requirements

### Requirement: MutationAttempt 수명주기
각 MutationAttempt는 RECORDED → DISPATCH_STARTED → (ACKED | IN_DOUBT) → 종결(CONFIRMED | NOT_DISPATCHED | FAILED_CONFIRMED | UNRESOLVED_IN_DOUBT) 단계를 가져야 한다(SHALL). 이 수명주기는 PLACE/CANCEL/AMEND뿐 아니라 조건주문의 등록·취소·정정에도 동일하게 적용된다(SHALL). RECORDED는 fsync 완료 후에만 DISPATCH_STARTED로 진행하며(SHALL), 재시작 시 RECORDED 단계에서 멈춘 attempt는 NOT_DISPATCHED로 안전 종결하고, DISPATCH_STARTED 이후 단계는 해소 절차 완료 전까지 차단 대상으로 취급한다(SHALL). 다만 차단의 범위는 mutation safety class 규칙을 따르며, 미해소 EXPOSURE_RAISING attempt가 같은 심볼의 RISK_REDUCING mutation을 차단해서는 안 된다(SHALL NOT).

#### Scenario: 전송 시작 전 크래시
- **WHEN** RECORDED까지만 기록된 attempt가 재시작 시 발견되면
- **THEN** NOT_DISPATCHED로 종결되고 어떤 차단도 발생하지 않는다

#### Scenario: 전송 중 크래시
- **WHEN** DISPATCH_STARTED로 기록된 attempt가 재시작 시 발견되면
- **THEN** IN_DOUBT 해소 절차가 시작되고 완료 전까지 해당 심볼의 신규 EXPOSURE_RAISING mutation이 차단된다

#### Scenario: 조건주문 attempt의 동일 수명주기
- **WHEN** 조건주문 등록 attempt가 DISPATCH_STARTED로 기록된 뒤 재시작되면
- **THEN** 일반 주문과 동일한 IN_DOUBT 해소 절차가 적용된다

### Requirement: IN_DOUBT 해소
IN_DOUBT 해소는 다음을 모두 만족해야 한다(SHALL): (1) 자동 재제출 절대 금지 — 브로커 멱등성 키가 없으므로 무조건, (2) journal에 저장된 fingerprint(계좌·심볼·방향·수량·가격·제출 시각 창, 조건주문은 트리거 가격·주문 유형 포함)로 미체결과 종결 **양쪽** 목록을 pagination 완주하며 대조, (3) 부재 판정은 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 + 매수가능금액/보유수량 delta 교차 확인 후에만, (4) 증명 불가 시 UNRESOLVED_IN_DOUBT로 영구 차단하고 운영자 해소만 허용. 유일 매칭을 보장하기 위해 엔진은 심볼당 in-flight EXPOSURE_RAISING mutation을 항상 1개로 제한한다(SHALL). RISK_REDUCING mutation의 유일 매칭은 대상 식별자(조건주문 ID 또는 대상 주문번호)로 보장하며, 이 클래스는 심볼 단위 제한을 받지 않는다(SHALL NOT).

#### Scenario: 제출 응답 유실 후 주문이 2페이지에 존재
- **WHEN** IN_DOUBT 해소 조회에서 대상 주문이 목록 2페이지 이후에 있으면
- **THEN** pagination 완주로 발견되어 CONFIRMED로 종결된다

#### Scenario: 단발 부재 조회
- **WHEN** 첫 조회에서 주문이 보이지 않으면
- **THEN** FAILED로 판정하지 않고 안정화 재조회를 계속한다

#### Scenario: 해소 불능
- **WHEN** 관찰 기간 내 존재도 부재도 증명되지 않으면
- **THEN** UNRESOLVED_IN_DOUBT로 표기되어 해당 심볼의 신규 진입이 영구 차단되고 운영자 알림이 발송된다 (보호·청산 경로는 계속 동작)
