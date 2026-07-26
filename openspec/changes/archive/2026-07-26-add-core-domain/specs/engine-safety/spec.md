# engine-safety Specification (delta)

## MODIFIED Requirements

### Requirement: 결정 영속과 신뢰 경계
GuardianDecision은 발급자가 Gateway 호출 **전에** journal에 영속해야 한다(SHALL): class별 preimage 원문(EXPOSURE_RAISING → RiskIntent: 계좌·시장·심볼·방향·진입가·손절가·목표가·수량·정책 버전 / RISK_REDUCING → ReductionIntent: 계좌·시장·심볼·방향·상한 수량·사유), canonical 해시, generation, place 결정의 멱등키. 멱등키는 발급자가 소유한 값에서만 유도한다(SHALL — `f(decision_id, generation)`). generation의 전진 주체는 후속 change가 정의한다(SHALL).

EXPOSURE_RAISING 결정의 영속은 위험 예약 삽입과 **하나의 journal 트랜잭션**에서 수행되어야 하며(SHALL — 예약이 거부되면 결정도 함께 롤백되어 제출 가능한 결정이 남지 않는다), Gateway는 EXPOSURE_RAISING 결정의 제출 시 **HELD 상태의 예약 존재를 검증**한다(SHALL — 예약이 총계 한도의 권위라는 계약의 강제 지점; 예약 없는 진입 결정은 거부된다). RISK_REDUCING 결정은 예약을 요구하지 않는다.

attempt 기록은 결정 참조(decision_id·safety_class·generation)를 함께 영속하고(SHALL), Gateway는 제출 직전 **journal에서 읽은 preimage**로 해시를 재계산해 주문 파라미터·멱등키 일치를 대조한다(SHALL). 제출 호출자가 공급한 위험 데이터로 재검증해서는 안 된다(SHALL NOT — 검증이 순환한다).

수동 flatten은 청산·취소 결정을 ReductionIntent preimage와 함께 journal에 기록한 뒤 제출한다(SHALL — 비상 경로가 검증에 거부되어서도, 검증을 면제받아서도 안 된다).

Gateway는 브로커 호출 직전 결정의 만료 시각을 재검증하며 만료된 결정의 제출은 거부한다(SHALL).

#### Scenario: 손절 데이터 바꿔치기
- **WHEN** 결정 발급 시점과 다른 손절가로 주문이 제출되면
- **THEN** journal의 preimage와 불일치하여 Gateway가 거부한다

#### Scenario: 멱등키 불일치
- **WHEN** 결정에서 유도된 것과 다른 clientOrderId로 제출이 구성되면
- **THEN** Gateway가 거부한다

#### Scenario: 예약 없는 진입 결정 제출
- **WHEN** HELD 예약이 없는 EXPOSURE_RAISING 결정으로 제출을 시도하면
- **THEN** Gateway가 거부한다

#### Scenario: flatten의 청산 결정
- **WHEN** flatten saga가 청산 결정을 발급·기록하고 제출하면
- **THEN** ReductionIntent preimage 검증을 통과하며 한도·예약 요구 없이 수행된다

#### Scenario: 만료된 결정으로 제출
- **WHEN** 발급 후 만료 시각이 지난 결정으로 제출하면
- **THEN** Gateway가 브로커 호출 전에 거부하고 재발급을 요구한다

### Requirement: 자동화 게이트 기동 인터록
자동 주문 게이트는 기본 OFF이며(SHALL), 게이트 ON 설정 시 다음이 모두 검증되지 않으면 기동을 거부한다(SHALL):

1. 필수 한도 전부가 명시적으로 설정되고 양수·유한하며 통화 일치 — 주문 수량, 주문 notional, 총 개방 노출, 일일 손실 절대액, 일일 손실 자본 비율 중 **하나라도** 누락·0·NaN·Inf이면 거부(부분적으로 무제한인 게이트는 허가된 게이트가 아니다)
2. 유효한 capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치. attestation endpoint 집합은 엔진 자동 경로가 실제 사용하는 호출 전부와 drift 가드로 동기화한다(SHALL — 목록을 확장하는 change는 가드를 함께 갱신한다. 본 change는 exit 관측이 사용하는 가격 조회를 목록에 추가한다)
3. 거래 정책이 매도와 실주문 실행을 허용 — 매수는 가능한데 청산이 불가능한 조합으로는 기동할 수 없다(SHALL NOT)
4. Guardian이 인터록이 감사한 설정 한도와 같은 출처에서 구성됨 — 동등성 검증은 EXPOSURE_RAISING 결정의 한도에만 적용
5. 엔진 프로필에 ExecutionGateway가 구성됨(round-trip용 주문 조회 배선 포함)
6. **브로커측 보호 실행이 배선됨** — 보호주문 실행이 준비되지 않은 프로필(ProtectionReady 미충족)에서는 게이트 ON 기동을 거부한다(SHALL — 로컬 기준선은 프로세스 사망 시 무력하므로, 보호주문 도입 change가 이 표지를 배선하기 전에는 자동 진입이 켜질 수 없다. 산문 금지가 아니라 인터록 조항이다)

게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).

#### Scenario: attestation 만료 상태 기동
- **WHEN** 게이트 ON + attestation 만료 상태로 기동하면
- **THEN** 기동이 거부되고 재검증 안내가 출력된다

#### Scenario: 한도 일부만 설정
- **WHEN** 주문 수량 한도만 양수이고 총 개방 노출 한도가 설정되지 않은 상태로 기동하면
- **THEN** 기동이 거부된다

#### Scenario: 청산 불가 정책으로 기동
- **WHEN** 매도가 비활성인 거래 정책으로 게이트 ON 기동하면
- **THEN** 기동이 거부된다

#### Scenario: 한도 출처 불일치
- **WHEN** 주입된 Guardian이 인터록이 검증한 설정 한도와 다른 한도로 EXPOSURE_RAISING 결정을 찍으면
- **THEN** 기동(또는 해당 결정)이 거부된다

#### Scenario: 보호 실행 미배선 기동
- **WHEN** ProtectionReady가 충족되지 않은 프로필에서 게이트 ON 기동하면
- **THEN** 기동이 거부되고 보호주문 실행 선행이 안내된다
