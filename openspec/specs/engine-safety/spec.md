# engine-safety Specification

## Purpose
엔진 배선 안전(official-only)·ExecutionGateway 봉인·결정 계약(safety class·preimage·nonce)·기동 인터록·flatten saga·알림·audit 요구사항.

## Requirements

### Requirement: 엔진 배선의 구조적 official-only
엔진 프로필(`internal/app` engine wiring)은 공식 Open API 브로커를 **직접** 구성해야 하며(SHALL), hybrid 클라이언트·WTS mutator 타입은 엔진 의존 그래프에 존재해서는 안 된다(SHALL NOT — 정적 import 테스트로 검증). 엔진 프로필은 사용자 config의 `OpenAPI.Enabled/Prefer`를 무시하고, 유효한 공식 자격증명이 없으면 기동을 거부한다(SHALL). place/cancel/amend/조건주문 전 mutation matrix에서 WTS 호출 0회를 spy 테스트로 증명한다(SHALL).

#### Scenario: 자격증명 누락 기동
- **WHEN** 공식 자격증명 없이 엔진 프로필을 기동하면
- **THEN** WTS 폴백 없이 기동이 명시적으로 거부된다

#### Scenario: 전 mutation matrix WTS 미도달
- **WHEN** 엔진 배선으로 place/cancel/amend/조건주문을 각각 실행하면
- **THEN** WTS spy 호출 횟수는 모두 0이다

### Requirement: 엔진 브로커의 cancel/amend 사전 확인
공식 API에는 `GetOrderAvailableActions` 대응이 없으므로, 엔진 브로커 어댑터는 cancel/amend 사전 확인을 `OrderByID` 상태 파생으로 구현하거나 사전 확인을 브로커 선택적으로 만들어야 한다(SHALL). WTS 세션이 없거나 만료된 상태에서도 엔진의 cancel/amend는 동작해야 한다(SHALL — 테스트 필수).

#### Scenario: WTS 세션 만료 중 취소
- **WHEN** WTS 세션이 만료된 상태에서 엔진이 미체결 주문을 취소하면
- **THEN** 공식 API 경로만으로 취소가 완료된다

### Requirement: ExecutionGateway 봉인
엔진의 모든 주문 mutation은 단일 ExecutionGateway를 통해야 한다(SHALL). 엔진 프로필은 다음 순서로 구성한다(SHALL): 계좌 해석(게이트 상태와 무관) → journal 열기(파일시스템 allowlist·무결성 검사가 엔진 기동 조건이 된다 — P1 journal 계약의 의도된 상속) → Gateway 구성(journal·EntryGate의 journal 투영 재구성·해소기·durable NonceStore·예약 저장소) → 인터록. Gateway 없이 mutation을 낼 수 있는 엔진 구성은 존재해서는 안 된다(SHALL NOT).

Guardian 결정 없는 제출 경로는 컴파일·API 수준에서 존재하지 않아야 한다(SHALL NOT). 엔진 컨텍스트는 mutation 메서드를 가진 서비스 값을 외부에 노출해서는 안 되며(SHALL NOT — 확인 토큰은 호출자가 로컬에서 계산 가능하므로 봉인이 되지 못한다), 봉인은 정적 테스트로 증명한다(SHALL). 기존 소비자인 flatten은 엔진 컨텍스트가 아니라 자체 배선으로 구성하며 P1 동작 무변경을 고정한다(SHALL). 멱등 재생의 해소 전용 진입점은 attempt 식별자만 입력받고 저장된 wire body 외를 전송할 수 없다(SHALL NOT — 두 번째 제출 문 금지). Gateway는 멱등키를 실을 수 없는 transport로의 place를 거부한다(SHALL). 기존 CLI/MCP 표면은 upstream confirm token 게이트를 유지하며 이 계약의 대상이 아니다 — MCP 우회 리스크는 Phase 4(단일 writer 데몬)까지 문서화 유지.

#### Scenario: Guardian 결정 없는 제출 시도
- **WHEN** GuardianDecision 없이 Gateway 제출을 시도하면
- **THEN** 컴파일 오류 또는 즉시 거부된다

#### Scenario: 엔진 컨텍스트의 mutation 노출 부재
- **WHEN** 엔진 컨텍스트가 노출하는 값들에서 Gateway를 거치지 않는 mutation 경로를 찾으면
- **THEN** 그런 경로가 존재하지 않음이 정적 테스트로 증명된다

#### Scenario: 키 미지원 transport
- **WHEN** 멱등키를 실을 수 없는 브로커 경로로 place가 구성되면
- **THEN** Gateway가 제출 전에 거부한다

#### Scenario: Gateway 미구성 기동
- **WHEN** 게이트 ON 상태인데 엔진 프로필에 Gateway가 구성되지 않았으면
- **THEN** 기동이 거부된다

### Requirement: 결정의 Safety Class와 형태 일치
GuardianDecision은 mutation의 safety class를 명시 필드로 실어야 한다(SHALL): EXPOSURE_RAISING(진입 제출) / RISK_REDUCING(reduce-only 청산, 취소). PROTECTION_WEAKENING은 enum 값으로 예약되며 보호주문 도입 change가 발급·소비를 정의한다.

class 선언은 그것만으로 효력이 없다(SHALL NOT — 위조 가능한 표지는 한도 우회가 된다). Gateway는 mutation 형태에서 노출 증가 여부를 독립 계산해 **EXPOSURE_RAISING ⇔ 노출 증가** 일치를 검증하고 불일치를 거부한다(SHALL). 한도 면제는 mutation 종류 리터럴이 아니라 이 검증을 통과한 class 기준으로 판정한다(SHALL). EXPOSURE_RAISING 결정은 필수 한도가 모두 설정된 스냅샷 없이는 거부되고(SHALL), RISK_REDUCING 결정은 한도 스냅샷을 싣지 않으며 수량·금액 한도의 적용을 받지 않는다(SHALL).

#### Scenario: class 위조 시도
- **WHEN** 매수 주문이 RISK_REDUCING class의 결정으로 제출되면
- **THEN** 형태 불일치로 거부되어 한도 우회가 발생하지 않는다

#### Scenario: 주문 한도를 초과하는 청산
- **WHEN** 주문당 최대 수량을 초과하는 포지션을 전량 청산하면
- **THEN** RISK_REDUCING 결정이므로 한도 초과로 거부되지 않는다

#### Scenario: 한도 없는 진입 결정
- **WHEN** 한도 스냅샷이 비었거나 항목이 누락된 EXPOSURE_RAISING 결정으로 제출하면
- **THEN** 거부된다

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

### Requirement: 결정 nonce의 durable 저장
one-shot nonce 저장소는 journal 기반이어야 한다(SHALL). 프로세스 재시작이 소비 기록을 잃어서는 안 되며(SHALL NOT), 영속된 결정 스냅샷을 새 제출에 사용하려는 시도는 nonce 재사용으로 거부된다(SHALL). 소비 기록은 전송 시작 기록과 같은 트랜잭션에서 남긴다(SHALL). 소비 기록의 보존 기간은 최대 결정 유효 시간 이상이어야 한다(SHALL). 멱등 재생(해소 절차)은 nonce 소비가 아니며 재사용 거부의 대상이 아니다(SHALL NOT).

#### Scenario: 재시작 후 결정 재사용 시도
- **WHEN** 재시작 후 journal에 보존된 GuardianDecision 스냅샷으로 새 제출을 시도하면
- **THEN** 이미 소비된 nonce로 판정되어 거부된다

#### Scenario: 재시작 후 해소 재생
- **WHEN** 재시작 후 IN_DOUBT attempt의 멱등 재생이 수행되면
- **THEN** nonce 재사용 거부의 대상이 되지 않고 해소 절차로 진행된다

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

### Requirement: Flatten Saga
`tossctl` flatten-all은 durable saga로 구현되어야 한다(SHALL): (1) 신규 진입 차단 → (2) 미체결 각각 취소 + 결과 확정(IN_DOUBT 규칙 적용) → (3) 계좌 재조회 안정화 → (4) 최신 매도가능수량 기준 reduce-only 청산 주문(비-fractional은 공격적 limit) → (5) 반복 reconcile로 잔여 확인. `--dry-run` 모드는 제출할 주문 목록을 mutation 0건으로 출력해야 한다(SHALL). 확인 문자열은 마스킹된 계좌 식별·포지션 수·예상 청산 수량·만료 nonce를 포함하고, TTY 직접 입력만 허용하며 자동화 플래그는 금지된다(SHALL NOT). 크래시 후 재실행 시 saga는 journal 기록에서 안전하게 재개된다(SHALL).

#### Scenario: 취소 결과 불명 중 청산 단계 진입 시도
- **WHEN** 미체결 취소가 IN_DOUBT 상태인데 청산 단계로 진행하려 하면
- **THEN** 해당 심볼 청산은 취소 확정 시까지 보류되고 oversell이 방지된다

#### Scenario: dry-run
- **WHEN** flatten-all --dry-run을 실행하면
- **THEN** 취소·청산 대상 목록이 출력되고 어떤 mutation도 발생하지 않는다

### Requirement: 등급화된 알림
알림은 등급화되어야 한다(SHALL): critical(IN_DOUBT·UNRESOLVED_IN_DOUBT 발생, 자격증명 만료 임박, 영구 불일치, UNKNOWN_BROKER_STATE)은 로컬 durable outbox에 기록 후 전송하고, 전달 실패가 지속되면 신규 진입을 차단한다(SHALL). 일반(체결·상태 전이)은 best-effort. 죽은 프로세스는 스스로 통지할 수 없으므로 heartbeat(예상 주기 초과 시 ntfy 측에서 경보) 방식을 사용한다(SHALL). 알림 이벤트 타입은 확장 가능한 enum으로 정의하고 Phase 2 이벤트(kill switch·운영 모드)는 예약만 한다.

#### Scenario: critical 알림 전달 실패 지속
- **WHEN** critical 이벤트의 전송이 재시도 한도까지 실패하면
- **THEN** 신규 진입이 차단되고 outbox에 미전달 상태로 보존되며, 전달 복구 후 수동 확인으로 해제한다

### Requirement: 테스트·도구의 실 endpoint 기계적 차단
테스트 바이너리와 검증 도구는 실 Toss hostname에 대한 mutation 요청이 구조적으로 불가능해야 한다(SHALL): 테스트는 격리 config 디렉터리 + httptest transport만 사용하고, 실 hostname으로의 POST 시도를 hard fail시키는 transport 가드 테스트를 완료 게이트에 포함한다(SHALL).

#### Scenario: 테스트에서 실 hostname POST 시도
- **WHEN** 테스트 중 실 Toss hostname으로 mutation 요청이 구성되면
- **THEN** transport 가드가 즉시 실패시키고 테스트가 실패한다

### Requirement: 운영 설정 audit
게이트 토글·한도 변경 등 운영 설정 변경은 변경 전후 값·시각·주체를 audit 로그로 기록해야 한다(SHALL).

#### Scenario: 게이트 토글 변경
- **WHEN** 자동화 게이트 설정이 변경되면
- **THEN** audit 로그에 이전 값·새 값·시각이 기록된다
