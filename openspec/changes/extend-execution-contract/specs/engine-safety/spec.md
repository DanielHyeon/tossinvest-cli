# engine-safety Specification (delta)

## ADDED Requirements

### Requirement: RiskIntent 결합
GuardianDecision은 주문 파라미터뿐 아니라 위험 판정의 입력에도 결합되어야 한다(SHALL). 불변 `RiskIntent`(계좌·시장·심볼·방향·진입가·손절가·목표가·수량·정책 버전)의 canonical hash가 결정에 실리고, Gateway는 제출 직전 실제 주문 파라미터와 동반 위험 데이터에서 해시를 재계산해 대조한다(SHALL). 재계산 값이 다르면 제출은 거부된다(SHALL). Phase 1의 주문 intent 해시는 유지되며 RiskIntent 해시가 추가된다 — 전자는 "이 주문", 후자는 "이 위험 판단"을 묶는다.

#### Scenario: 손절 데이터 바꿔치기
- **WHEN** 결정 발급 시점과 다른 손절가로 주문이 제출되면
- **THEN** RiskIntent 해시 불일치로 Gateway가 거부한다

### Requirement: 위험 감소 mutation의 한도 면제 계약
보호주문과 청산 등 위험 감소 mutation의 결정은 빈 한도 스냅샷을 실어야 한다(SHALL). 주문 수량·notional 한도가 청산을 거부하는 상황은 §0.3 위반이므로, 발급자는 이 클래스에 대해 한도 스냅샷을 채우지 않는다(SHALL NOT). Gateway의 빈 스냅샷 처리는 기존 규칙을 유지한다 — 빈 스냅샷은 노출을 증가시키는 mutation만 거부한다.

#### Scenario: 주문 한도를 초과하는 청산
- **WHEN** 주문당 최대 수량을 초과하는 보유 포지션을 전량 청산하면
- **THEN** 한도 초과로 거부되지 않고 제출된다

### Requirement: 결정 nonce의 durable 저장
one-shot nonce 저장소는 journal 기반이어야 한다(SHALL). 프로세스 재시작이 소비 기록을 잃어서는 안 되며(SHALL NOT), 영속된 결정 스냅샷을 재제출에 사용하려는 시도는 nonce 재사용으로 거부되어야 한다(SHALL).

#### Scenario: 재시작 후 결정 재사용 시도
- **WHEN** 재시작 후 journal에 보존된 GuardianDecision 스냅샷으로 제출을 시도하면
- **THEN** 이미 소비된 nonce로 판정되어 거부된다

## MODIFIED Requirements

### Requirement: ExecutionGateway 봉인
엔진의 모든 주문 mutation은 단일 ExecutionGateway를 통해야 하며(SHALL), 여기에는 조건주문의 등록·취소·정정이 포함된다(SHALL). Gateway는 GuardianDecision(주문 intent 해시, RiskIntent 해시, 한도 스냅샷, 만료 시각, one-shot nonce)을 요구하고 브로커 호출 직전 재검증한다(SHALL). Guardian 결정 없는 제출 경로는 컴파일·API 수준에서 존재하지 않아야 한다(SHALL NOT: raw mutator 미노출, 엔진 경로의 조건주문 직접 호출 미노출). 기존 CLI/MCP 표면은 upstream의 confirm token 게이트를 유지하며 이 변경의 대상이 아니다 — MCP 표면이 Gateway를 우회한다는 잔존 리스크는 Phase 4(단일 writer 데몬)까지 문서화된 상태로 유지된다.

#### Scenario: Guardian 결정 없는 제출 시도
- **WHEN** GuardianDecision 없이 Gateway 제출을 시도하면
- **THEN** 컴파일 오류 또는 즉시 거부된다

#### Scenario: nonce 재사용
- **WHEN** 이미 사용된 GuardianDecision으로 재제출을 시도하면
- **THEN** one-shot nonce 검증 실패로 거부된다

#### Scenario: 엔진의 조건주문 직접 호출 부재
- **WHEN** 엔진 의존 그래프에서 Gateway를 거치지 않는 조건주문 제출 경로를 찾으면
- **THEN** 그런 경로가 존재하지 않음이 정적 테스트로 증명된다

### Requirement: 자동화 게이트 기동 인터록
자동 주문 게이트는 기본 OFF이며(SHALL), 게이트 ON 설정 시 다음이 모두 검증되지 않으면 기동을 거부한다(SHALL): (1) 필수 한도 전부가 명시적으로 설정되고 양수·유한하며 통화가 일치 — 주문 수량, 주문 notional, 총 개방 노출, 일일 손실 절대액, 일일 손실 자본 비율 중 **하나라도** 누락·0·NaN·Inf이면 거부한다(SHALL — 부분적으로 무제한인 게이트는 허가된 게이트가 아니다), (2) 유효한 capability attestation(만료 시각·계좌 식별·성공 endpoint 집합을 담은 로컬 durable 기록 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치, (3) attestation의 endpoint 집합이 조건주문의 등록·취소·정정을 포함, (4) 거래 정책이 매도와 조건주문과 실주문 실행을 모두 허용 — 매수는 가능한데 손절·청산이 불가능한 조합으로는 기동할 수 없다(SHALL NOT), (5) Guardian의 한도 스냅샷이 인터록이 감사한 설정 한도와 같은 출처에서 구성됨. 게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).

#### Scenario: attestation 만료 상태 기동
- **WHEN** 게이트 ON + attestation 만료 상태로 기동하면
- **THEN** 기동이 거부되고 재검증 안내가 출력된다

#### Scenario: 한도 일부만 설정
- **WHEN** 주문 수량 한도만 양수이고 총 개방 노출 한도가 설정되지 않은 상태로 기동하면
- **THEN** 기동이 거부된다

#### Scenario: 청산 불가 정책으로 기동
- **WHEN** 매도 또는 조건주문이 비활성인 거래 정책으로 게이트 ON 기동하면
- **THEN** 기동이 거부된다

#### Scenario: 조건주문 미검증 attestation
- **WHEN** attestation의 endpoint 집합에 조건주문이 없는 상태로 게이트 ON 기동하면
- **THEN** 기동이 거부된다
