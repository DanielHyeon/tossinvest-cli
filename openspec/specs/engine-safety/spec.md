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
2. 유효한 capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치. attestation endpoint 집합은 엔진 자동 경로가 실제 사용하는 호출 전부와 drift 가드로 동기화한다(SHALL — 목록을 확장하는 change는 가드를 함께 갱신한다)
3. 거래 정책이 **청산 경로가 실제로 요구하는 것 전부**를 허용 — `place`·`sell`·`cancel`·`allow_live_order_actions`(SHALL). 매수는 가능한데 청산이 불가능한 조합으로는 기동할 수 없다(SHALL NOT)는 이 절이 처음부터 주장한 것이고, 검사 대상이 그 주장에 미치지 못했다 — 매도 제출도 `place`이며(`internal/trading` 정책 검사가 side와 무관하게 요구한다), exit 관측 루프는 자기 발의를 취소하므로 `cancel`도 쓴다. 둘이 꺼진 채 기동하면 엔진은 뜨고 **첫 손절에서 거부된다**. 검사에 넣지 않은 요구는 요구가 아니다
4. Guardian이 인터록이 감사한 설정 한도와 같은 출처에서 구성됨 — 동등성 검증은 EXPOSURE_RAISING 결정의 한도에만 적용
5. 엔진 프로필에 ExecutionGateway가 구성됨(round-trip용 주문 조회 배선 포함)

게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).

**브로커측 보호 실행 배선(ProtectionReady)은 기동 조건이 아니라 진입 허가 조건이다**(SHALL).
조항 1~5가 기동을 결정하고, ProtectionReady는 기동한 런타임에서 **무엇이 허가되는지**를 결정한다.
이 분리의 근거는 실패의 비대칭이다 — 진입 후 프로세스가 죽으면 손절 없는 **새** 포지션이 남고,
청산 전에 죽으면 보호 미배선 상태의 **기존** 포지션이 남는다. 후자는 런타임을 거부해도 동일하게
발생하는 상태이므로, 그것을 이유로 청산 루프를 거부하는 것은 같은 결함을 더 넓은 범위에 유지한다.

ProtectionReady 미충족 프로필에서:

- 게이트 ON + 조항 1~5 충족이면 런타임은 **기동한다**(SHALL). 루프 집합은 변하지 않는다 —
  reconcile driver·exit observer·체결 감지.
- 노출을 **증가**시키는 mutation은 거부된다(SHALL). 집행은 mutation chokepoint에서 이루어지며,
  판정 근거는 호출자가 선언한 Safety Class가 아니라 **mutation 자체의 형태에서 계산된 사실**이다
  (매수 여부). 클래스를 잘못 붙여 우회할 수 없다.
- 집행 지점은 **하나**여야 한다(SHALL — 두 곳에서 판정하면 답을 틀릴 곳이 두 곳이 된다).
- ProtectionReady 표지는 config 키·Options 필드가 아닌 **컴파일 타임 상수**로만 존재한다
  (SHALL NOT — 설정으로 만족시킬 수 있는 표지는 짓지 않고 준비됐다고 주장하는 길이다).
  보호주문 도입 change가 자기 작업의 마지막 단계로 이 상수를 뒤집는다.
- 운영자에게 보호가 **프로세스 수명에 묶여 있다**는 사실이 기동 시점에 전달되어야 한다(SHALL —
  읽는 문장이며, 타이핑 확인·추가 승인 마찰이어서는 안 된다(SHALL NOT)).

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

#### Scenario: 보호 미배선 기동
- **WHEN** ProtectionReady 미충족 프로필에서 게이트 ON + 조항 1~5 충족으로 기동하면
- **THEN** 런타임이 기동하고, 보호가 프로세스 수명에 묶여 있다는 사실이 출력되며, 상태 보고의 protection은 UNWIRED로 남는다

#### Scenario: 보호 미배선에서 노출 증가 mutation
- **WHEN** ProtectionReady 미충족 상태에서 매수 mutation이 제출되면
- **THEN** 거부되고 사유가 보호 미배선으로 열거된다 — 결정이 EXPOSURE_RAISING으로 찍혀 있든 아니든 판정은 mutation의 형태에서 나온다

#### Scenario: 보호 미배선에서 노출 감소 mutation
- **WHEN** ProtectionReady 미충족 상태에서 매도 mutation이 제출되면
- **THEN** 통과한다 — 청산은 이 조항이 지목하는 실패를 만들지 않는다

#### Scenario: 청산에 필요한 토글 누락
- **WHEN** `trading.place` 또는 `trading.cancel`이 꺼진 채로 게이트 ON 기동하면
- **THEN** 기동이 거부되고 꺼진 토글이 이름으로 열거된다 — 손절을 낼 수 없는 구성으로 기동하지 않는다

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

**exit 관측 goroutine은 알림의 원격 전송을 동기로 기다려서는 안 된다(SHALL NOT — 손절 판정이 사는 goroutine이고 §0.3이 걸리는 자리다).** 그 goroutine 안에서 완료되어야 하는 것은 **그 등급이 갖는 내구 기록까지**이며(SHALL — critical은 outbox 행 하나, 일반 등급은 내구 기록이 없으므로 구조화 로그 줄 하나다. 기록은 전송보다 먼저라는 기존 요구는 완화되지 않고 강화된다), 원격 전송·시도 간 대기·시도별 실패 기록·전달 실패에 따른 진입 게이트 래치·운영 모드 승격은 그 goroutine **밖**에서 일어나야 한다(SHALL). critical 알림의 그 일은 정본 「배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」의 배달 실행자가 진다.

**exit 관측 goroutine에서 닿는 알림 입구는 전부 이 요구의 대상이다(SHALL).** 알림을 올리는 입구뿐 아니라 그 goroutine이 일으킨 **운영 모드 전이의 통지**(관측 두절에 따른 강화, 그 goroutine의 조회가 받은 자격증명 거절에 따른 강화, 청산 수량 상한을 읽는 조회에서의 같은 강화)도 같은 goroutine에서 원격 전송을 기다리게 하므로, 그 통지도 기록까지만 하고 반환해야 한다(SHALL). 통지를 없애서는 안 된다(SHALL NOT — 정본 risk-management는 자동 강화 때 알림 발송을 요구한다). 입구는 **주입 지점별로** 묶는다(SHALL — 같은 부품을 범위 밖 루프와 공유하면 exit goroutine에 주입된 인스턴스만 기록 전용이 되고, 범위 밖 호출자의 동작은 바뀌지 않는다). 입구를 셀 때는 호출 자리가 아니라 **그 goroutine에서 알림 경로에 도달하는 경로**를 센다(SHALL — 호출 자리 하나를 막고 「한 자리」라 적은 것이 21라운드의 P0였다).

**exit 관측 goroutine의 critical 기록은 발송 임차를 잡지 않아야 한다(SHALL NOT — 잡은 임차가 남으면 배달 실행자는 그 임차가 끝날 때까지 그 행을 집지 못하고, 첫 발송이 임차 길이만큼 늦는다).** 기록은 다른 발송자가 이미 쥔 임차를 풀거나 덮어서도 안 된다(SHALL NOT). exit 관측 goroutine의 기록이 재알림 창이 지난 정착 행을 다시 무장하는 동작은 그대로여야 한다(SHALL — 정본 「재무장된 outbox 행은 통째로 이번 에피소드를 말한다」가 그 기록 위에 서 있다). 이 재무장 요구는 exit 관측 goroutine의 기록에 대한 것이다 — 다른 기록자의 재알림 창은 그 기록자가 정하며 0(재무장 안 함)일 수 있다.

**exit 관측 goroutine이 전송하지 않고 반환한 것은 전달 실패가 아니다(SHALL NOT — 그 반환이 전달 실패 사유의 래치나 운영 모드 승격으로 이어져서는 안 된다).** 전달 실패의 판정은 배달 실행자의 시도에 선다. 기록 자체의 실패는 이 문장의 대상이 아니다 — 아래 「durable 기록의 실패」가 그것을 다룬다.

위의 원격 전송 비대기 문단들의 범위는 **exit 관측 goroutine**이다(SHALL — 사용자 결정 20-1의 문자. 이 요구 첫 문단의 등급·outbox·heartbeat 규범은 좁히지 않는다). 다른 호출자(대사 루프·런타임의 루프 비정상 반환 알림·주문 경로·운영자 시험 발송)가 자기 전송을 동기로 기다리는 것은 이 요구가 바꾸지 않으며(그 호출자가 쥐는 잠금의 범위는 아래 잠금 문단이 정한다), 이 요구를 그 호출자들의 체류 근거로 인용해서는 안 된다(SHALL NOT — 범위 해석은 `proposal.md` 「열린 질문」 Q1). 사람이 명령을 입력해 기다리는 대화형 시험 발송은 루프를 붙잡지 않으므로 여기에 들지 않는다(SHALL NOT — 범위를 넓히면 이 요구는 사람의 대기 시간까지 규정하게 되고, 그것은 §0.3이 지키려는 것이 아니다).

> **18판이 이 두 문장을 문단으로 갈랐다.** 앞 문단에 등급별 내구 기록을 적자 체류 낱말 둘이 한 문단에 모였고, 같은 문단의 범위 배제 문장이 `check_values.py`의 부정 표현 규칙에 걸렸다 — **한 문단이 열거와 배제를 동시에 하고 있었다.** 검사기를 고치지 않고 문단을 갈랐다. 검사기가 맞았다.
>
> 이 주석은 그 낱말들과 부정 표현을 **재현하지 않는다.** 재현하면 이 주석 자체가 같은 규칙에 걸린다.

**이 요구는 등급으로 좁혀지지 않는다(SHALL NOT).** critical만 exit 관측 goroutine 밖으로 옮기고 일반 등급의 전송을 그 goroutine 안에 남기는 구성은 이 요구를 만족하지 않는다 — 일반 등급의 동기 발행도 같은 네트워크 왕복을 같은 goroutine에서 기다린다. 일반 등급은 durable 기록을 갖지 않으므로 그 이관은 유실을 허용하는 형태여도 된다(SHALL — 버리는 것이 등급 강등이 아닌 이유는 이 등급의 정의가 best-effort이기 때문이다). 다만 **무엇을 버렸는지는 기록되어야 한다(SHALL — 기록 없는 유실은 전송된 것과 구별되지 않는다).** (20판 문장 유지 — Manager 판정 Q2. 이관 수단은 design D0.3g.)

**배달 실행자는 자기 사이클의 작업량에 상한을 가져야 한다(SHALL — 밀린 행 전부를 한 사이클에서 처리하는 구성은 사이클 자체가 무계다).** 배달 실행자가 보조 실행자라는 것과 런타임이 그것에 지는 의무는 정본 「엔진 런타임 수명주기」·「배달 실행자의 정지가 다른 루프를 내려서는 안 된다」가 싣는다. 같은 행이 두 번 나가지 않게 하는 것은 실행자의 수가 아니라 원장의 임차이고, 그 보장은 정본 「발송 권한은 원장이 준다」가 **유계 실행** 아래로 한정한다 — 이 델타는 배달 경로가 하나라는 사실을 배제의 근거로 쓰지 않는다.

**exit 관측 goroutine이 기다리는 잠금은 원격 전송을 덮어서는 안 된다(SHALL NOT — 덮으면 그 goroutine은 자기 전송 대신 남의 전송을 기다리고, 결함은 사라지지 않고 자리만 옮긴다).** 동기 알림 경로와 잠금을 공유하는 것 자체는 금지되지 않는다(SHALL — 로컬 원장 작업만 덮는 잠금은 원격 왕복만큼 붙잡지 않는다). 그러므로 엔진이 조립해 실행하는 발송 경로 가운데 그 잠금을 쥐는 **모든 보유자** — 이 요구의 범위 밖에서 동기 발송을 계속하는 호출자를 포함한다 — 는 대상 행을 고르고 결과를 정산할 때만 잠금을 쥐고, 원격 전송 동안에는 놓아야 한다(SHALL). 전송 동안의 행 단위 배제는 원장의 임차가 진다(SHALL — 정본 「배달 실행자는 잠금을 쥔 채 전송하지 않는다」·「발송 권한은 원장이 준다」). 기록 경로와 잠금을 전혀 공유하지 않고 원장 임차로 배제하는 배달 실행자는 이 요구를 만족한다(SHALL — a098의 배달 실행자가 그 형태다).

> **21판이 이 문단의 주어를 바꿨다.** 20판은 *"배달 경로는 원격 전송 위에서 어떤 잠금도 잡고 있어서는 안 된다"*였다. 배달 실행자는 동기 경로의 잠금을 잡지 않으므로 그 문장은 과녁이 아니었다. 과녁은 **exit 관측 goroutine이 줄 서는 잠금**이고, 그 잠금을 원격 전송 위에서 쥘 수 있는 것은 범위 밖 호출자의 동기 발송이다(측정과 좌표는 design D0.3e 5번 — 규범 문장에 코드 좌표를 두지 않는다).

**전송하는 동안 잠금을 놓는다는 것은 그 사이에 남이 그 행을 정착시킬 수 있다는 뜻이고, 그것은 허용되어야 한다(SHALL — 18라운드 A-P1).** 발송자가 결과를 정산할 때 정산이 **오류 없이 「이미 정산됨」 또는 「임차 상실」로** 돌아오면 그것은 오류가 아니라 **원장이 이름을 준 선점**으로 다루어야 하며(SHALL — 「이미 정산됨」은 전달 또는 운영자 승인이고, 「임차 상실」은 다른 발송자다. 행 없음·모르는 결과·원장 오류는 선점이 아니다(SHALL NOT) — 배달 실행자는 정본 「배달 실행자는 지속 실패를 …」의 판정을, 범위 밖 동기 발송자는 아래 「모든 발송자」 문단을 따른다), 발송자는 그 행을 덮지 않고(SHALL NOT — 승인을 미전달로 되돌리는 것은 사람이 이미 본 사건을 다시 세우는 것이다) 남은 행들의 처리를 계속해야 한다(SHALL — 한 행의 선점이 배치를 중단시켜서는 안 된다). **선점이 일어났다는 사실은 기록되어야 한다(SHALL — 조용히 갈라지는 분기는 다음 사람이 경합을 결함으로 오해하는 자리다).**

> **21판이 이 문단을 귀속만 고쳐 되살렸다 (Manager 판정 2026-09-28, Q5 조건).** 초안은 이 문단을 정본 사본으로 보고 지웠으나, 「남은 행의 처리를 계속한다」와 「운영자 승인에 의한 선점을 기록한다」를 소유하는 정본 문장이 없다(`review.md` §23.6). 20판의 *"대상 행이 더 이상 `PENDING`이 아니면 … 운영자가 먼저 정산했다는 사실로"*는 a099 이후 틀린 귀속이라(20라운드 B-4 = A-7) 원장의 결과 이름으로 바꿨다.

**기록 경로가 잠금 아래에서 하는 일은 로컬 원장 연산이어야 한다(SHALL).** 잠금 아래에 원격 호출·무기한 대기·사람의 입력을 넣는 구현은 이 요구를 만족하지 않는다(SHALL NOT). 다만 잠금 아래의 로컬 연산에도 **기한은 없다**(원장이 멈추면 관측도 멈춘다) — 이 요구가 만드는 것은 기한이 아니라 **그 항이 네트워크가 아니라는 것**이다.

**엔진 프로세스에서 critical 행을 기록하는 정식 경로는 알림기의 배제 잠금 아래에서 기록하는 경로다(SHALL)** — 알림기의 기록 전용 입구(발송 임차 없이 기록하고 재알림 창은 기록자가 정하며 0, 즉 재알림 창에 의한 재무장을 하지 않음도 허용된다)와, 범위 밖 호출자의 동기 발송이 그 잠금 아래에서 하는 claim 둘 다 이 부류다. 재알림 창에 의한 재무장은 이 부류를 통해서만 일어나야 한다(SHALL — 배제 잠금 밖에서 정착 행을 재알림 창으로 재무장하면 아래 셈과 해제 사이에 PENDING 행이 생긴다).

**미전달 수를 읽고 전달 실패 사유를 푸는 판단은 알림기의 배제 잠금 아래에서 기록하는 경로에 대해 나눌 수 없는 하나여야 한다(SHALL).** 그 사이에 그 경로로 새 PENDING 행이 생기면 그 판단은 틀린 것이고, 새 PENDING 행은 재무장뿐 아니라 **새 event key의 최초 기록으로도** 생긴다. 그 잠금 밖에서 원장에 직접 쓰는 기록자는 그 창을 덮지 않는 대신, 행을 넣기 **전에** 자기 진입 차단 사유를 세워야 한다(SHALL — 그러면 셈과 해제 사이에 그 행이 들어와 전달 실패 사유가 풀려도 진입은 그 사유로 막혀 있다). 세울 자기 사유가 없는 기록자(예: 운영자 완화의 통지)는 잠금 밖에서 쓰지 말고 알림기의 기록 전용 입구를 써야 한다(SHALL). 이 문단들은 **엔진 프로세스**에서, **생산 조립** — `tossctl` 실행 파일의 main에서 도달하는 비시험 경로 — 이 도달하는 기록에 대한 것이다(SHALL). 생산 조립이 그 기록자에게 엔진의 진입 게이트를 **항상** 넘긴다는 것은 조립 생성자 전수로 확인되어야 한다(SHALL — 한 자리의 존재 확인은 역할 확인이 아니다). 그 순서는 구조로 고정되어야 하며(SHALL), 금지 형태는 **삽입이 잠금보다 앞선 것, 그리고 잠금과 삽입 사이에 그 사유를 푸는 호출이 있는 것**이다. 무엇이 그 구간을 지키는지와 그 수단이 덮지 않는 기록자가 무엇인지는 change에 적혀야 한다(SHALL — 적히지 않은 조건은 다음 편집이 지우는 조건이다). 이 요구가 닫는 것은 **셈과 해제 사이의 유입**뿐이다 — 승인의 대상 집합이 정해진 **뒤**에 들어와 해제 전에 다른 발송자가 전달한 행은 이 요구로 덮이지 않으며, 이 요구를 그 행의 보호 근거로 인용해서는 안 된다(SHALL NOT — a124 design D10 (i)의 경계).

**운영자의 승인은 진입 게이트를 풀 수 있고, 전송의 성공 여부는 그 사유를 되살려서는 안 된다(SHALL NOT) — 모든 발송자에 대해 그렇다.** 배달 실행자에 대해서는 정본 「배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」의 「늦은 적용은 제때 적용과 같아야 한다」가 그 규칙이다. 범위 밖 동기 발송자도 같은 규칙을 따라야 한다(SHALL): 전달 실패의 근거가 확정된 순간(시도 기록·전달 기록·임차 반납의 결과가 돌아온 순간) 전달 실패 사유의 해제 세대를 읽고, 그 뒤에 해제가 있었으면 차단을 적용하지 않으며(SHALL NOT), 그 판정에 승격이 포함되면 승격은 적용한다(SHALL — 제때 된 승격은 해제로 풀리지 않는다). 근거가 확정된 순간과 해제 세대를 읽는 순간 사이의 해제를 「앞」으로 보는 것은 허용된다(SHALL — 차단이 더 서는 쪽으로만 틀린다). 승격이 포함된 판정에서 승격 쓰기가 실패하면, 조건부 차단의 결과와 무관하게 차단을 적용해야 한다(SHALL — 조건부 차단이 해제로 생략됐어도 둘 다 잃지 않는다. a124 정본의 같은 보수 조항이다). 승격을 포함하지 않는 판정(시도 기록이 행을 찾지 못함 · 승격할 계정이 없음)은 승격을 만들어서는 안 된다(SHALL NOT). 해제와 근거의 순서는 해제 세대로 가르며, 원장의 승인 시각으로 순서를 추정해서는 안 된다(SHALL NOT — 시각은 승인이 끝난 순간이 아니고 벽시계는 되감긴다). 해제의 두 조건(승인했고 미전달 수가 0)은 정본 「운영자가 밀린 알림을 읽고 승인하는 경로가 존재해야 한다」 그대로다.

> **22판이 20판 문장을 「모든 발송자」로 일반화해 되살렸다 (Manager 판정 C27).** 21판은 이 규칙을 a124 정본에 맡기고 지웠으나 그 정본의 주어는 배달 실행자뿐이다. 21판의 잠금 범위 변경은 범위 밖 동기 발송자가 운영자 승인 **뒤에** 래치를 세울 수 있게 만든다(21라운드 C2, 세 보이스). 이 문단이 그 발송자에 대한 규범의 주인이다.

**전송 시도 횟수는 다시 계약이다(SHALL).** 시도를 줄이는 유일한 근거는 그 시도가 손절 루프를 붙잡는다는 것이었고, 루프 밖으로 나가면 그 근거가 없다. 그러므로 이 change는 시도 횟수를 **줄이지 않는다**(SHALL NOT — 줄이는 구성은 upstream이 이미 landed한 재시도 계약을 이 change가 거짓으로 만드는 구성이다). 시도 사이의 대기를 배달 루프의 주기로 대신하는 구현은 허용되며(SHALL), 그 경우 **한 사이클은 한 행을 한 번만 시도해야 한다**(SHALL — 한 사이클 안에서 재시도하면 사이클이 다시 무계가 된다).

**진입 게이트 래치가 늦어지는 것은 이 재설계의 대가이며 change에 적혀야 한다(SHALL — 적히지 않은 대가는 없는 것처럼 읽힌다).** exit 관측 goroutine이 동기 시도를 하지 않으면 전달 실패에 따른 래치는 배달 실행자의 판정으로만 선다. 그 시간은 사이클 시작 기준 `Q + (L−1)·C + I_list + (T + S + M)`이다(SHALL — 기호는 a124 design D6: `Q` 첫 사이클 시작까지의 큐 대기 · `L` 재시도 한도 · `C` 사이클 길이 · `I_list` 나열 비용 · `T` 한 행의 발행 실패 시간 · `S` 한 행의 원장 비용 · `M` 해제 세대 읽기와 울타리의 게이트 잠금 대기). 이 값은 a124 design D6의 **전제 H(동질 두절)** 아래의 조건부 상한이고, 전제가 깨지는 경우를 상한이라고 불러서는 안 된다(SHALL NOT). 20판의 식 *"다음 배달 사이클까지의 시간 + 시도 횟수 × 1회 상한 + (시도 횟수 − 1) × 사이클 주기"*는 「사이클 주기」를 `C`로 읽으면 발행 시간을 두 번 센다(a124 freeze Q6) — 그 식으로 돌아가서는 안 된다(SHALL NOT). 큐 대기를 포함해야 한다는 규칙은 정본 「배달 실행자의 행 선택은 굶주림을 만들지 않는다」가 싣는다. **이 지연이 손절을 늦추지 않는다는 것은 게이트가 진입만 막기 때문이며, 그 사실이 인용의 조건이다(SHALL — 게이트가 청산을 막게 되는 편집이 들어오면 이 대가는 다시 계산되어야 한다).**

**durable 기록의 실패는 여전히 동기로 다루어진다(SHALL).** outbox 기록 자체가 실패하면 행이 없고, 행이 없으면 배달 실행자가 찾을 것도 없다. 그 경우의 진입 게이트 래치와 운영 모드 승격은 기록을 시도한 그 자리에서 일어나야 하며(SHALL — 루프 밖으로 미룰 대상이 없다), 그 비용은 실패한 로컬 쓰기와 그 뒤의 게이트 래치·승격 트랜잭션이다. 이 경우의 승격은 **통지하지 않는다**(SHALL — 방금 기록에 실패한 원장에 통지를 다시 기록하려는 것은 같은 실패를 되풀이한다. 그 사실은 구조화 로그가 남긴다). 위의 「통지를 없애서는 안 된다」는 이 경우의 예외를 둔다.

**이 요구가 만드는 것은 여전히 보장이 아니다(SHALL NOT — 실제 체류가 주기를 넘지 않는다고 이 요구를 인용해서는 안 된다).** 루프에 남는 항은 구조화 로그 한 줄과 outbox 트랜잭션이고(22판: 임차를 잡지 않는 기록 트랜잭션 하나 — design D0.3g), 여기에 동기 알림 경로의 잠금을 기다리는 시간이 더해진다. **셋 다 기한이 없다.** 원장이 멈추면 관측도 멈춘다. 이 재설계가 바꾸는 것은 기한 없는 항의 **크기**이지 기한의 유무가 아니다 — 네트워크 왕복 × 시도 + 시도 간 대기였던 것이 로컬 쓰기가 된다. 그러므로 루프에 남는 몫은 이름을 갖고 편성되어야 하며(SHALL — 이름 없는 여유는 다음 편집이 말없이 먹는다), 그 몫이 관측 주기보다 작아야 한다(SHALL NOT — 같거나 크게 두는 편성은 이 요구를 위반한다).

**루프에 남는 몫의 근거가 그 경로를 직접 잰 것이 아니면 그렇게 적어야 한다(SHALL).** 축소된 경로의 실측이 없는 동안 상위집합의 실측을 대입하는 것은 보수적이므로 허용되며(SHALL), 그 대입이라는 사실과 실측 의무는 change에 남아야 한다(SHALL — 대입을 실측으로 적으면 다음 편집이 그 위에 짓는다).

**1회 전송 상한은 transport 구현이 소유한다(SHALL).** 조립부가 채운 값은 그 값을 읽는 transport에 대해서만 실효를 가지며, 그것을 읽지 않는 `Publisher` 구현에는 상한이 없다(SHALL NOT — 조립부에 적혔다는 사실이 모든 구현에 대한 보장은 아니다). **그 상한을 잘못 고르는 대가는 exit 관측 goroutine에서 내려간다(SHALL 근거 — 상한이 손절 경로 밖에 있으므로, 낮게 고른 상한은 손절을 늦추지 않고 배달 사이클 하나를 버린다).** 다만 실측 의무 자체는 남는다(SHALL — 실측 없이 고른 값은 여전히 실측 없이 고른 값이다).

예산 축소가 durability를 줄여서는 안 된다(SHALL NOT): outbox 기록은 전송 시도보다 먼저이며 예산과 무관하게 완료된다(SHALL). 전송 수단이 실제로는 메시지를 받았는데 응답이 상한 안에 오지 않아 실패로 기록되는 경우가 생긴다(SHALL — 그 경우 운영자는 중복을 받고 게이트는 래치되며, 이 손실은 인정되고 기록되어야 한다).

#### Scenario: critical 알림 전달 실패 지속
- **WHEN** critical 이벤트의 전송이 재시도 한도까지 실패하면
- **THEN** 신규 진입이 차단되고 outbox에 미전달 상태로 보존되며, 전달 복구 후 수동 확인으로 해제한다

#### Scenario: 엔진 루프가 전송을 기다리지 않는다
- **WHEN** exit 관측 사이클이 critical 알림을 올리고 전송 수단이 응답하지 않으면
- **THEN** 그 사이클은 **그 호출이 쓰는 구조화 로그 줄 하나와 outbox 트랜잭션**(임차를 잡지 않는 기록 하나)을 마치고 즉시 다음 포지션으로 넘어가며, 전송·시도별 실패 기록·게이트 래치·승격 트랜잭션은 그 사이클 밖에서 일어난다 — 사이클이 붙잡히는 시간에 원격 왕복이 포함되지 않는다

#### Scenario: exit 관측의 기록은 임차를 잡지 않는다
- **WHEN** exit 관측 사이클이 critical 알림을 기록하고 반환한 직후 배달 실행자의 사이클이 시작되면
- **THEN** 배달 실행자는 그 행을 남의 임차로 보지 않고 집을 수 있다
- **AND** 재알림 창이 지난 정착 행이었다면 그 행은 다시 무장되어 있다
- **AND** 그 기록 전에 다른 발송자가 그 행의 임차를 쥐고 있었다면 그 임차는 그대로다

#### Scenario: exit 관측이 일으킨 모드 전이의 통지도 기다리지 않는다
- **WHEN** 운영 모드가 NORMAL인 계정에서 exit 관측 사이클의 관측 두절 판정이나 가격 조회의 자격증명 거절이 운영 모드를 강화하고, 전송 수단이 응답하지 않으면
- **THEN** 그 사이클은 모드 전이와 그 통지의 기록까지만 하고 넘어가며, 통지의 전송은 배달 실행자가 한다
- **AND** 같은 부품을 쓰는 범위 밖 호출자(대사 루프 등)의 통지 동작은 바뀌지 않는다

#### Scenario: 승인 뒤 늦게 끝난 동기 전송 실패는 다시 잠그지 않는다
- **WHEN** 범위 밖 동기 발송자가 전달 실패의 근거를 확정한 뒤, 차단을 적용하기 전에 운영자가 밀린 행을 전부 승인해 전달 실패 사유를 푼다
- **THEN** 그 발송자는 전달 실패 사유로 다시 잠그지 않는다 — 해제가 그 발송자의 해제 세대 읽기 **뒤**일 때. 근거 확정과 세대 읽기 사이의 해제는 보수적으로 다시 잠글 수 있다
- **AND** 그 판정에 승격이 포함되면 운영 모드는 제때 적용했을 때처럼 승격되어 있다

#### Scenario: 해제로 생략된 차단 뒤 승격 쓰기가 실패한다
- **WHEN** 범위 밖 동기 발송자의 판정에서 해제 때문에 조건부 차단이 생략되고, 그 판정의 승격 쓰기가 원장 오류로 실패한다
- **THEN** 그 발송자는 전달 실패 사유로 진입을 잠근다 — 차단과 모드를 둘 다 잃지 않는다

#### Scenario: 원장에 직접 쓰는 기록자는 먼저 잠근다
- **WHEN** 알림기의 배제 잠금 밖에서 원장에 직접 쓰는 기록자가 critical 행을 넣고, 그 행이 운영자의 셈과 해제 사이에 들어온다
- **THEN** 전달 실패 사유가 풀려도 신규 진입은 그 기록자가 먼저 세운 자기 사유로 막혀 있다

#### Scenario: exit 관측의 반환은 전달 실패가 아니다
- **WHEN** exit 관측 사이클이 critical 알림을 기록만 하고 반환하면
- **THEN** 그 반환은 전달 실패 사유로 진입을 잠그지도 운영 모드를 승격하지도 않는다 — 그 판정은 배달 실행자의 시도에서 선다

#### Scenario: exit 관측의 기록은 남의 원격 전송을 기다리지 않는다
- **WHEN** 이 요구의 범위 밖 호출자가 동기 발송으로 원격 전송 중이고, 그때 exit 관측 사이클이 critical 알림을 올리면
- **THEN** 관측 사이클은 그 전송이 끝나기를 기다리지 않고 자기 기록을 마친다 — 기다리는 구성은 결함을 옮긴 것이지 없앤 것이 아니다

#### Scenario: 진입 차단이 늦게 걸린다
- **WHEN** 전송이 계속 실패해 시도가 소진되면
- **THEN** 진입 차단은 전송이 exit 관측 goroutine 안에 있던 구성보다 늦게 걸리며, 그 지연은 change에 전제와 함께 조건부 식으로 적혀 있고, 그동안 청산 경로는 아무 영향도 받지 않는다

#### Scenario: outbox 기록 자체가 실패한다
- **WHEN** critical 알림의 durable 기록이 실패하면
- **THEN** 진입 게이트 래치와 운영 모드 승격이 그 자리에서 동기로 일어난다 — 배달 실행자가 찾을 행이 없으므로 미룰 대상이 없다

#### Scenario: 일반 등급도 루프를 붙잡지 않는다
- **WHEN** exit 관측 사이클이 일반 등급 알림을 올리고 전송 수단이 응답하지 않으면
- **THEN** 그 사이클은 기다리지 않고 넘어가며, 이관 버퍼가 가득 차 있으면 그 알림은 버려지고 버렸다는 사실이 기록된다

#### Scenario: 사이클 총 체류는 여전히 약속하지 않는다
- **WHEN** 한 관측 사이클이 억제 없는 경로에서 알림을 여러 개 올리면
- **THEN** 사이클의 총 동기 체류는 알림 수에 비례해 늘어나며 이 요구는 그것을 유계로 만들지 않는다 — 비례하는 항이 원격 왕복에서 로컬 쓰기로 바뀔 뿐이다

#### Scenario: 상한을 읽지 않는 transport
- **WHEN** 조립부가 1회 상한을 채웠는데 배선된 `Publisher`가 그 값을 읽지 않으면
- **THEN** 그 구성에는 1회 상한이 없으며 배달 사이클의 작업량 상한이 그 구현에 대해 성립하지 않는다

#### Scenario: 예산을 줄여도 기록은 그대로다
- **WHEN** 전송 예산이 줄어든 구성에서 critical 알림이 올라오고 전송이 전부 실패하면
- **THEN** outbox 행은 이전과 동일하게 기록되어 있고 진입 차단도 이전과 동일하게 걸린다

#### Scenario: 다시 올릴 주기가 있어도 래치가 그것을 막는다
- **WHEN** 알림 산출 지점이 자기 래치를 세운 **뒤에** 알림을 올리는 구조이면
- **THEN** 다음 관측은 래치에서 early return하여 알림 호출에 닿지 않으므로, 그 경로의 재발송은 관측이 아니라 **배달 실행자**가 담당한다 — outbox 행이 PENDING으로 남는 것이 재발송의 근거가 된다

#### Scenario: 재강화는 새 게이트에 투영되지 않는다
- **WHEN** 원장의 운영 모드가 이미 강화된 상태에서 같은 트리거로 다시 강화하면
- **THEN** 변화 없음으로 조기 반환되어 모드 투영과 통지가 실행되지 않으므로, 재시작 후의 새 게이트는 이 재강화로부터 아무것도 받지 못한다 — 기록의 영속성을 강제의 영속성으로 인용해서는 안 된다(SHALL NOT). 모드를 재시작 너머로 집행하는 것은 재강화가 아니라 기동 복원이다(아래 ADDED 요구)

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

### Requirement: 엔진 런타임 수명주기

엔진 런타임(`tossctl engine run`)의 **감독 루프 집합**은 **reconcile driver·exit observer·체결 감지(= `filldetect.Detector` 폴링 루프)·전략 진입 외곽 루프(`strategy-entry`)**다(SHALL — 체결 감지 없는 런타임은 발의 pending이 영구 미해소로 남아 exit 수명주기 계약을 위반한다; **`strategy-entry`는 이 빌드에서 휴면이다** — 루프 자체는 기동하고 취소에 정상 배수하지만 두 시장이 모두 비활성이며(`cmd/tossctl/engine_strategy_entry_dormant_test.go:15-47`), 이 요구는 **진입이 활성이어야 한다고 요구하지 않는다**(SHALL NOT); 힌트 라우팅(`Hints`)을 포함하는 경우 Refresh 미배선은 감독이 아니라 **조립 시점 검증**으로 거부한다(SHALL); exit 관측의 SLO 양보 지점은 엔진 배선의 어댑터로 체결 감지 상태에 연결한다). 기동 순서(SHALL): ① **journal 디렉터리 flock 획득이 기동의 첫 동작이다** — 실패(다른 인스턴스 보유)면 즉시 거부한다(SHALL — journal은 단일 writer 설계이고, flock이 첫 동작이어야 journal open·마이그레이션 전체가 배타 안에 들어온다; 자문 마커로는 경합이 닫히지 않는다). ② 게이트 OFF면 기동할 루프 집합이 없으므로 거부한다 — 이는 이 change가 정의하는 규칙이다(기동 인터록은 게이트 ON에만 정의된다). ③ 게이트 ON이면 기동 인터록 검증을 소비하고, 미충족이면 인터록이 반환한 미충족 항목을 열거하며 거부한다(fail-closed). ④ verify runlock이 신선하면 거부한다. 이와 별도로 **엔진 활성 마커**(갱신 1분·stale 5분 — runlock 선례 수치)를 유지해 콘솔의 엔진 상태 표시·autostart의 사전 확인이 소비한다(SHALL — 자문 신호이며 배타는 flock이 담당함을 명시). verify 측이 이 마커를 검사해 엔진 실행 중 verify를 거부하는 것은 execution-verification change(2b)의 후속 태스크다 — 이 change는 엔진 측 검사(verify runlock 신선 시 기동 거부)만 소유한다.

**런타임은 감독 집합 밖에 보조 실행자를 기동할 수 있다**(SHALL). 감독 집합에 넣는다는 것은 **그 일의 비정상 반환으로 런타임 전체를 내린다**는 뜻이다(방어적 종료 계약). 알림 배달의 정지는 그런 실패가 아니다 — **배달이 멈춘 동안에도 손절·비상 청산은 계속되어야 한다**(안전 불변식 4). 보조 실행자에는 감독 계약이 걸리지 않는다(SHALL NOT — 방어적 종료도, 지속 열화 임계도 아니다). 그 대신 런타임이 보조 실행자에 대해 **셋을 져야 한다**(SHALL): ① 기동한 보조 실행자가 **반환할 때까지 기다린 뒤** 원장을 닫는다, ② 보조 실행자의 패닉이 **프로세스를 죽이지 않게 한다**, ③ 보조 실행자의 비정상 정지를 **관측하고 기록한다**. 셋 중 하나라도 없으면 「감독 밖」은 곧 **「아무도 안 본다」**와 같아진다.

감독 계약은 두 층이다(SHALL): ① **방어적 종료 계약** — **감독 루프가** 컨텍스트 취소 외의 사유로 반환하면 전체 런타임이 정지하고 critical 알림이 발송된다(현행 루프들은 그런 반환을 하지 않으므로 이는 방어선이다). 컨텍스트 취소에 의한 반환은 정상 종료이며 critical을 발송하지 않는다(SHALL NOT). ② **지속 열화 임계** — 루프가 살아 있으나 사이클이 연속 실패하는 상태를 각 루프에 정의한다: exit 관측은 landed 60초 두절 계약 유지, reconcile driver와 체결 감지는 연속 5주기 실패 시 critical 알림 + ENTRY_BLOCKED 자동 강화(SHALL — 자동 강화 트리거 열거는 risk-management delta가 확장한다; 루프는 계속 재시도한다 — landed "실패한 사이클은 다음 주기에 재시도" 결정과 양립).

종료 시그널은 루프 취소·완주 대기·journal 정합 close로 처리하고(SHALL), 두 번째 시그널은 즉시 종료한다. 재기동 복구는 landed 계약(pending 복원·편입 완결·nonce 재사용 금지)을 소비하며 새 복구 경로를 만들지 않는다(SHALL NOT).

> **이 `MODIFIED`가 고치는 것은 둘이고, 둘째는 a098이 만든 것이 아니다.**
>
> | | 정본이 적고 있던 것 | 고친 뒤 | 누가 만들었나 |
> |---|---|---|---|
> | 열거 | 셋 (`:170`) | **넷** | **a098이 아니다.** `strategy-entry`는 이미 착지해 있고(`cmd/tossctl/engine.go:377-398`) 테스트가 넷을 고정한다(`cmd/tossctl/engine_strategy_entry_dormant_test.go:50-58`). 그 확장을 기록한 `MODIFIED`가 `openspec/specs/`에 **없었다**(`rg strategy-entry openspec/specs/` → 0건) |
> | 보조 실행자 | **개념이 없다** | 있다 | **a098이다.** 결정 9-2 |
>
> **둘을 한 `MODIFIED`에 넣는 이유.** 하나만 고치면 나머지 하나가 남긴 거짓이
> 다음 편집의 근거가 된다. 열거만 고치면 보조 실행자는 여전히 *"루프 중 하나"*로
> 읽혀 방어적 종료의 대상이 되고, 보조 실행자만 더하면 **넷이 셋으로 적힌 문장이
> 그대로 남는다.** 그 문장이 4판의 방어가 서 있던 자리다.
>
> **a092 와의 어긋남은 정리됐다(a092 archive `2026-09-30-a092-an-alert-does-not-hold-the-stop`, tasks 24.5).**
> a092 의 옛 header note 는 *"「엔진 런타임 수명주기」는 루프를 **하나 더한다** — 알림 배달 루프"*라고 적었고
> 결정 9-2는 더하지 **않는다**(a099 §7.5). 아카이브된 a092 델타는 *"이 델타는 루프를 더하지 않고 「엔진 런타임 수명주기」를
> 건드리지 않는다"*로 끝났다 — a092 가 더한 보조 실행자(일반 등급 이관)는 이 요구의 보조 실행자 규칙을 따른다.

#### Scenario: 게이트 OFF 기동
- **WHEN** 게이트 OFF 상태로 `engine run`을 실행하면
- **THEN** "기동할 루프 집합이 없다(게이트 OFF)"로 거부되고 실패 종료한다 — 인터록 조항 열거는 없다

#### Scenario: 게이트 ON + 인터록 미충족 기동
- **WHEN** 게이트 ON + ProtectionReady 미충족 상태로 `engine run`을 실행하면
- **THEN** 루프가 하나도 시작되지 않고 인터록의 미충족 항목이 열거되며 실패 종료한다

#### Scenario: 두 번째 인스턴스 기동
- **WHEN** 엔진이 실행 중인 머신에서 `engine run`(또는 autostart·콘솔 버튼)이 다시 실행되면
- **THEN** 실행 중 인스턴스가 안내되고 기동이 거부된다

#### Scenario: 루프의 비정상 반환
- **WHEN** 기동된 **감독 루프** 중 하나가 컨텍스트 취소가 아닌 사유로 반환하면
- **THEN** 나머지 루프도 정지하고 critical 알림이 발송되며 프로세스가 실패로 종료한다

#### Scenario: 보조 실행자의 비정상 반환
- **WHEN** 기동된 **보조 실행자**가 컨텍스트 취소가 아닌 사유로 반환하거나 패닉하면
- **THEN** 감독 루프는 **하나도 정지하지 않고** 프로세스도 죽지 않는다
- **AND** 그 정지가 **관측되어 기록된다**
- **AND** 런타임은 종료할 때 그 실행자의 반환을 **기다린 뒤** 원장을 닫는다

#### Scenario: 프로덕션 감독 루프 집합은 넷이다
- **WHEN** `tossctl engine run`이 감독 루프를 기동한다
- **THEN** 감독 루프의 이름 집합은 reconcile·exit·체결 감지·`strategy-entry` **넷과 정확히 같다**
- **AND** 그 검사는 **부분 일치가 아니라 집합 동일성**이어야 한다 — 부분 일치는 다섯 번째가 늘어도 초록이다
- **AND** 배달 실행자의 이름은 **거기 나타나지 않는다** — 나타나면 보조 실행자가 아니다

> **이 Scenario는 하한이다.** 오늘 저장소의 핀은 이보다 강하다 —
> `TestProductionRuntimeIncludesOneDormantStrategyEntryOuterLoop`
> (`cmd/tossctl/engine_strategy_entry_dormant_test.go:50-58`)이
> `reflect.DeepEqual`로 **순서까지** 고정한다. 기존 핀을 이 하한까지 낮추지 말 것 —
> 그 금지는 spec이 아니라 **구현 task가 진다**(a098 tasks §5.2c). spec이 테스트의
> 강도를 규범으로 정하면 승인된 정본이 테스트 구현에 묶인다.
>
> 같은 파일 계열에 **부분 일치만 하는 테스트**가 하나 있다 —
> `TestTheLoopSetIsTheSpecifiedThree`(`cmd/tossctl/engine_test.go:347-361`)는
> 소스 문자열 포함만 보므로 **루프가 넷이어도 통과한다.** 그것은 이 Scenario의
> 증거가 **아니다**.

#### Scenario: 정상 종료는 critical이 아니다
- **WHEN** SIGTERM으로 런타임이 graceful 종료하면
- **THEN** 루프가 취소·완주되고 journal이 정합하게 닫히며 critical 알림은 발송되지 않는다

#### Scenario: reconcile driver 지속 실패
- **WHEN** reconcile 사이클이 연속 5회 실패하면
- **THEN** critical 알림과 함께 ENTRY_BLOCKED로 자동 강화되고 루프는 재시도를 계속한다

#### Scenario: 검증 실행 중 기동 시도
- **WHEN** verify runlock이 신선한 상태에서 `engine run`을 실행하면
- **THEN** 기동이 거부되고 검증 종료 후 재시도가 안내된다

### Requirement: ProtectionReady는 attestation 범위에서만 WIRED다
엔진은 ProtectionReady를 exact attestation 범위에서만 WIRED로 계산해야 한다 (SHALL). KR과
US 각각에 대해 현재 계좌·profile·시장·주문유형·수량·세션·trigger source,
atomic/continuous replace semantics 및 exact broker identity/query/dedup/idempotency capability가
strict versioned signed attestation과 일치하고 protection supervisor/saga가 production
assembly에 배선된 경우에만 해당 시장의 exposure-raising mutation을 허용해야 한다 (SHALL).
Attestation은 tool/build와 evidence digest, issued-at, expires-at, monotonic serial, key ID,
allowlisted algorithm과 signature에 묶이고 pin된 trust root, revocation, bounded rotation overlap,
maximum lifetime와 durable trusted-time floor를 통과해야 한다 (SHALL). legacy/unknown field,
capability 누락, 만료, signature·digest·경로·소유자·권한 불일치, serial/time rollback은 해당
시장에서 fail-closed해야 한다 (SHALL). 한 시장의 불일치는 다른 시장의 유효한 readiness를
낮춰서는 안 되지만 (MUST NOT), 실패한 시장의 신규 진입을 허용해서도 안 된다 (MUST NOT).
Readiness 판정은 운영 토글, lane, autostart, automation gate 또는 LIVE approval을
생성·변경해서는 안 된다 (MUST NOT). 외부 readiness enum은 정확히 `WIRED`와 `UNWIRED`만
허용해야 하며 (SHALL), 모든 실패 상세는 enum 변형 대신 stable typed refusal과 provenance로
제공해야 한다 (SHALL).

#### Scenario: 미검증 시장
- **WHEN** KR capability만 attested된 상태에서 US 자동 진입을 시도한다
- **THEN** entry는 protection_unwired로 거부되고 기존 US 보유의 protection, reconciliation과 reduce-only exit는 계속된다

#### Scenario: 유효 capability
- **WHEN** 현재 KR profile이 KR attestation과 일치하고 Guardian/gate가 유효하다
- **THEN** KR protection readiness clause는 `WIRED`지만 운영자가 승인하지 않은 KR lane는 여전히 OFF다

#### Scenario: US readiness와 KR 실패의 독립성
- **WHEN** US attestation과 wiring은 유효하지만 KR attestation은 만료됐다
- **THEN** KR 신규 진입만 protection_unwired로 거부되고 US readiness verdict는 유지되며 두 시장의 exit와 reconciliation은 계속된다

#### Scenario: decision 뒤 attestation 만료
- **WHEN** EntryDecision 생성 뒤 durable dispatch 직전에 해당 시장 attestation이 만료된다
- **THEN** broker request는 0건이고 해당 시장 readiness를 `UNWIRED`로 낮추며 기존 protection과 reduce-only exit를 유지한다

#### Scenario: signed evidence 부재
- **WHEN** 보호 saga 코드는 배선됐지만 current signed attestation이 없다
- **THEN** 두 시장 readiness는 `UNWIRED`이고 테스트 fixture 외의 live 보호주문으로 자동 검증하지 않는다

#### Scenario: readiness enum 계약
- **WHEN** attestation 또는 wiring 검증이 실패한다
- **THEN** readiness는 정확히 `UNWIRED`이고 `READY`, `DEGRADED`, `UNKNOWN` 같은 제3 enum 대신 typed refusal을 반환한다

#### Scenario: trust state 실패의 시장 격리
- **WHEN** KR attestation key는 revoked됐지만 US attestation은 별도 active key와 증가한 serial로 유효하다
- **THEN** KR만 `UNWIRED`와 typed key refusal이고 US는 `WIRED`를 유지하며 두 시장의 기존 protection과 exit는 계속된다

### Requirement: 자동 진입은 모든 안전 권한의 교집합이다
엔진은 automation gate, operating mode, lane state, Guardian, reconciliation health와 ProtectionReady가 모두 허용할 때만 strategy entry를 제출해야 한다 (SHALL).
이 조건들은 동일한 immutable activation manifest에 version/digest/expiry로 결합돼야 하며 (SHALL), durable dispatch 직전에 전부 재검증되지 않으면 신규 진입을 제출해서는 안 된다 (MUST NOT).

#### Scenario: protection 미배선
- **WHEN** 다른 조건이 허용돼도 ProtectionReady가 UNWIRED다
- **THEN** buy는 거부되고 reduce-only exit는 계속된다

#### Scenario: kill switch
- **WHEN** kill switch가 활성화된다
- **THEN** 신규 entry를 즉시 중지하고 기존 보호·청산 감독은 유지한다

#### Scenario: 승인 manifest 불일치
- **WHEN** decision 뒤 dispatch 전에 threshold, settings, attestation, Guardian, scheduler 또는 build digest가 바뀐다
- **THEN** 신규 attempt는 제출되지 않고 effective entry는 OFF와 구체적 refusal reason을 기록한다

#### Scenario: manifest 만료 뒤 재시작
- **WHEN** 저장된 desired state는 ON이지만 activation manifest가 만료됐다
- **THEN** 재시작은 승인 상태를 재구성하지 않고 entry OFF를 유지하며 exit/reconcile은 계속한다

### Requirement: Production startup constructs one Guardian after durable scope exists
The engine profile SHALL, when `engine.automation_gate.enabled` is true and no
explicit test Guardian is injected, resolve the official account and open the
writable engine journal before constructing exactly one production
`RiskGuardian`. It SHALL construct no Guardian and no loop set when the gate is
off. A Guardian construction failure SHALL close the journal, record/refuse
startup, and start no loop.

#### Scenario: Gate-on production assembly
- **WHEN** the real CLI assembler loads a valid gate, resolves an account, and opens the engine journal
- **THEN** it constructs exactly one `RiskGuardian` scoped to that account and journal before running the interlock

#### Scenario: Gate-off production assembly
- **WHEN** the real CLI assembler loads an automation gate that is off
- **THEN** it does not construct a Guardian and the command starts no engine loop

#### Scenario: Guardian construction fails
- **WHEN** the configured gate cannot produce a valid production Guardian after the journal is open
- **THEN** startup is refused, the journal is closed, and no loop or order side effect begins

### Requirement: Interlock and runtime share the Guardian identity
The engine SHALL pass the same Guardian instance to the startup interlock and,
after verification, publish that instance on `Context.Guardian`. The exit
observer SHALL obtain its `ReductionIssuer` from that field and SHALL NOT
construct, substitute, or bypass another Guardian.

#### Scenario: Verified context constructs exit observation
- **WHEN** the startup interlock verifies a production Guardian
- **THEN** the context publishes that exact instance and the exit observer uses it for reduce-only issuance

#### Scenario: Guardian cannot issue reductions
- **WHEN** the verified context carries a Guardian that does not implement `ReductionIssuer`
- **THEN** exit-observer construction fails before any observation loop starts

### Requirement: Command regression exercises production assembly
The command package SHALL have a regression test that invokes the actual
production assembly helper with an isolated config directory, a real SQLite
journal on an allowlisted test filesystem, and an `httptest` official broker.
The test SHALL inject no Guardian, SHALL contact no live endpoint, and SHALL
prove verified Guardian and exit-observer wiring while protection remains the
shipped `UNWIRED` value. The durability test override SHALL exist only under a
dedicated Go test build tag and SHALL have no flag, environment variable,
config key, or ordinary production symbol.

#### Scenario: Isolated USD CLI assembly
- **WHEN** the command regression supplies valid USD limits, credentials, attestation, account response, and no Guardian override
- **THEN** the actual CLI assembly constructs exactly one real `RiskGuardian`, returns the configured USD snapshot, records an isolated reduce-only decision through the context journal, and constructs an exit observer from that same Guardian

### Requirement: 엔진이 소유한 모든 런타임 endpoint는 자기 잔재에서 회복한다

엔진이 소유한 모든 런타임 endpoint(position policy command·position policy runtime·alert control 포함)의 기동 시 잔재 회수는 자기 생성·종료·회수 시퀀스가 control 디렉터리 안 파일에 만들 수 있는 모든 부분 상태(descriptor·socket·현행 및 구버전 staging 잔재)를 소유자 사망 검증(connect probe — PID 불사용) 후 사람 개입 없이 회수해야 하며(SHALL), socket 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 수락 중인 socket 위에 두 번째 서버가 올라서서는 안 된다(SHALL NOT).

#### Scenario: pre-chmod socket 잔재에서의 재기동

- **WHEN** listen과 chmod 사이에 죽어 group/other 비트 없는 비-0600 socket이 남은
  상태에서 엔진이 기동하면
- **THEN** 두 socket endpoint 모두 잔재를 회수하고 기동을 계속한다

#### Scenario: 산 주인의 endpoint는 탈취되지 않는다

- **WHEN** 살아 있는 주인이 수락 중인 socket 위에서 두 번째 기동이 시도되면
- **THEN** 두 번째 기동은 그 socket을 unlink하지 않고 거부된다

#### Scenario: staging 잔재는 우리 잔재다

- **WHEN** 발행 전 임시 이름(신규 `.s-` staging 또는 현행·구버전 공통 CreateTemp
  이름)의 정규 파일·socket만 남은 상태에서 엔진이 기동하면
- **THEN** 잔재를 회수하고 기동을 계속하며 회수 후 control 디렉터리에 잔재가 없다

#### Scenario: 낯선 엔트리는 건드리지 않는다

- **WHEN** socket을 발행하는 endpoint의 control 디렉터리에 그 endpoint가 만들 수 없는
  이름 또는 모양의 엔트리가 있는 상태에서 기동하면
- **THEN** 회수는 아무것도 제거하지 않고 그 endpoint의 기동을 거부한다

### Requirement: 엔진 기동은 자기 endpoint 표면의 실패로 죽지 않는다

position policy command·position policy runtime·alert control endpoint의 기동 실패는 엔진 부팅을 실패시키지 않고 해당 표면 없이 계속해야 하며(SHALL), 그 강등은 stderr 안내와 obs 이벤트로 보고하되(SHALL) 그 보고가 critical 등급·obs 등급표 등재·원장 outbox 적재 중 어느 것도 사용해서는 안 되고(SHALL NOT — 미전달 outbox 행은 다음 부팅의 진입 게이트를 잠근다, a108 D3-2), 강등된 표면의 소비자 메시지가 강등을 엔진 부재로 단정해서는 안 된다(SHALL NOT).

#### Scenario: endpoint 하나의 실패가 보호 루프를 세우지 않는다

- **WHEN** 세 endpoint 중 어느 하나의 Start가 어떤 이유로든 실패한 채 엔진이 기동하면
- **THEN** 엔진은 그 표면 없이 부팅을 완료하고 손절·청산 루프는 정상 가동한다

#### Scenario: 강등 보고는 진입 게이트를 잠그지 않는다

- **WHEN** 강등 보고가 발행된 뒤 엔진이 재시작하면
- **THEN** 그 보고로 인해 잠긴 진입 게이트가 없다

#### Scenario: 소비자는 강등을 엔진 부재로 오귀속하지 않는다

- **WHEN** 엔진이 alert control 또는 격리 해제 표면 없이 강등 부팅한 상태에서
  운영자가 그 표면의 CLI·콘솔 명령을 실행하면
- **THEN** 표시되는 메시지는 엔진 부재를 단정하지 않고 강등 가능성과 엔진 로그
  확인을 안내한다

### Requirement: attestation endpoint 집합의 증거원

attestation의 성공 endpoint 집합은 **각 항목이 그것을 증명할 수 있는 증거원에서만** 와야 한다(SHALL).
증거원은 둘이고 역할이 겹치지 않는다:

- **무인 read-only soak** — 읽기 endpoint를 증명한다. soak 기록의 비-GET 항목은 attestation에
  실려서는 안 된다(SHALL NOT) — 아무것도 접수하지 않는 도구의 기록에 mutation이 있다는 것은
  측정이 아니라 기록 오염이다.
- **사람이 승인한 감독 검증** — 무인 도구가 구조적으로 실행할 수 없는 mutation endpoint를
  증명한다.

감독 검증 기록은 **무인 도구가 실행할 수 없다고 선언된 endpoint 목록에 있는 것만** 기여할 수
있다(SHALL). 그 목록 밖의 endpoint는 감독 검증 기록이 성공을 증명하더라도 attestation에
실려서는 안 된다(SHALL NOT). 근거: 감독 하 1회 성공은 여러 날의 무인 운전이 증명하는 것과
같은 속성이 아니며, 읽기를 감독 검증으로 대신 증명하면 soak 결함이 조용히 덮인다.

감독 검증 기록이 endpoint를 기여하려면 다음이 **모두** 참이어야 한다(SHALL):

1. 그 endpoint의 호출이 오류 없이 **성공**했다
2. 그 기록의 계좌가 attestation의 계좌와 같다
3. 성공 시각이 attestation의 유효 기간 안이다
4. 그 endpoint가 무인 도구 실행 불가 목록에 있다

기록의 계좌가 attestation의 계좌와 **다르면** 조용히 건너뛰지 않고 발급을 거부해야 한다(SHALL) —
기대 경로에 다른 계좌의 기록이 있다는 것은 설정 오류이고, 무시하면 그 오류가 "증거 없음"과
구별되지 않는다.

attestation은 각 mutation endpoint를 **무엇이 증명했는지** 기록해야 한다(SHALL) — 최소한
endpoint, 성공 시각, 증거 기록의 출처. 근거: 인터록 거부 메시지는 무엇이 빠졌는지만 말하므로,
게이트가 통과한 뒤 "무엇을 근거로 켜졌나"에 답할 수 있는 곳은 attestation 자신뿐이다.

요구 endpoint 중 하나라도 어느 증거원으로도 채워지지 않으면 attestation은 그것을 **싣지 않은
채** 발급되고, 기동 인터록이 그 부족을 근거로 거부한다(SHALL) — 이 요구는 인터록이 요구하는
집합을 바꾸지 않는다(SHALL NOT).

#### Scenario: 감독 검증이 mutation endpoint를 증명한다

- **WHEN** 사람이 승인한 감독 검증이 주문 접수와 취소를 성공시킨 기록이 있고 soak이 완료된 상태에서 attestation을 발급하면
- **THEN** 그 두 mutation endpoint가 attestation의 성공 집합에 실리고, 각각을 무엇이 증명했는지가 함께 기록된다

#### Scenario: 감독 검증은 읽기를 증명하지 못한다

- **WHEN** 감독 검증 기록이 어떤 읽기 endpoint의 성공을 담고 있어도
- **THEN** 그 읽기는 감독 검증을 근거로 attestation에 실리지 않는다 — 읽기는 무인 soak이 증명한다

#### Scenario: 실패한 호출은 증거가 아니다

- **WHEN** 감독 검증 기록의 mutation 호출이 오류로 끝났으면
- **THEN** 그 endpoint는 실리지 않는다

#### Scenario: 계좌가 다른 증거는 발급을 거부시킨다

- **WHEN** 감독 검증 기록의 계좌가 soak의 계좌와 다르면
- **THEN** attestation은 발급되지 않고 사유가 보고된다

#### Scenario: 유효 기간 밖의 증거는 증거가 아니다

- **WHEN** mutation 성공 시각이 attestation 유효 기간보다 오래됐으면
- **THEN** 그 endpoint는 실리지 않는다

#### Scenario: 부족한 채로 발급되고 인터록이 거부한다

- **WHEN** 감독 검증이 아직 없어 mutation endpoint가 하나도 채워지지 않았으면
- **THEN** attestation은 읽기만 담은 채 발급되고, 게이트 ON 기동은 그 부족을 근거로 거부된다

### Requirement: 공통 정책 설정은 기동 시 fail-closed 검증된다
엔진은 non-empty 공통 policy ID가 registry에 없거나 policy가 ordering/ratio/runner 조건을 위반하면 exit observer 기동을 거부해야 한다 (SHALL).

#### Scenario: 손상된 common policy 설정
- **WHEN** config가 알 수 없는 common policy ID를 포함한다
- **THEN** 엔진은 조용히 RATCHET으로 후퇴하지 않고 이유를 포함해 기동을 거부한다

### Requirement: 공통 정책은 위험 축소 권한만 사용한다
공통 정책이 만드는 모든 주문 proposal은 기존 Guardian의 reduce-only issuance와 execution gateway를 거쳐야 하며 설정 승인이 LIVE order 승인이나 exposure 증가 권한으로 해석되어서는 안 된다 (MUST NOT).

#### Scenario: HYBRID_50 부분익절
- **WHEN** HYBRID_50 rung이 부분익절을 제안한다
- **THEN** 기존 reduction decision, reservation, idempotency, submit 경로를 사용하고 신규 buy 권한을 만들지 않는다

#### Scenario: automation gate OFF
- **WHEN** 공통 정책이 저장돼 있지만 automation gate가 OFF다
- **THEN** 설정은 보존되지만 unattended exit observer는 기존 interlock에 따라 기동하지 않는다

### Requirement: 승인된 엔진 자동 기동

`engine.autostart`는 기본 OFF여야 한다(SHALL). 콘솔 프로세스가 시작될 때 이
설정이 ON인 경우에만 기존 엔진 시작 경로를 정확히 한 번 호출해야 한다(SHALL).
자동 기동은 수동 [엔진 시작]과 동일한 journal flock, automation gate, Guardian,
capability attestation, 거래 정책, ExecutionGateway startup interlock을 사용해야
하며(SHALL), 그 중 어느 조건도 대신 설정하거나 우회해서는 안 된다(SHALL NOT).
설정 읽기 실패는 자동 기동을 생략하는 fail-closed 결과여야 한다(SHALL).

#### Scenario: 기본 설정과 구버전 설정
- **WHEN** 새 기본 config를 만들거나 `engine.autostart`가 없는 기존 config를 읽으면
- **THEN** autostart는 OFF이고 엔진 시작 호출이 발생하지 않는다

#### Scenario: 승인된 부팅 자동 기동
- **WHEN** `engine.autostart`가 ON인 config로 콘솔 프로세스가 시작되면
- **THEN** 기존 엔진 시작 seam이 정확히 한 번 호출되고 그 seam의 startup interlock 결과가 최종 기동 여부를 결정한다

#### Scenario: 자동 기동의 인터록 거부
- **WHEN** autostart는 ON이지만 automation gate 또는 기존 startup interlock 조건이 충족되지 않으면
- **THEN** 엔진은 기존 사유로 거부되고 콘솔은 계속 실행되며 실제 주문 경로는 열리지 않는다

#### Scenario: 설정 읽기 실패
- **WHEN** 콘솔 시작 시 autostart 설정을 읽을 수 없거나 JSON이 잘못되었으면
- **THEN** 엔진 시작 seam은 호출되지 않고 오류가 운영자에게 표시된다

#### Scenario: 부팅 중복 인스턴스
- **WHEN** autostart ON인 콘솔이 시작될 때 동일 journal의 엔진이 이미 실행 중이면
- **THEN** 기존 marker·process 검사와 journal flock이 두 번째 엔진을 허용하지 않는다

### Requirement: 자문 마커 단독으로 엔진 기동을 거부하지 않는다

엔진 활성 마커는 자문 신호이므로 기동 경로(`engine run`·autostart·콘솔 기동 버튼)는 마커의 신선도만을 근거로 기동을 거부해서는 안 된다(SHALL NOT — 배타는 journal 디렉터리 flock이 담당한다는 것이 `엔진 런타임 수명주기`의 기존 규정이다).

마커가 신선한데 엔진 프로세스가 관측되지 않으면 그 마커를 유령으로 판정하고 기동을 진행해야 한다(SHALL — 컨테이너 재생성·SIGKILL·호스트 재부팅은 프로세스를 지우지만 마커 파일을 지우지 않는다). 실제로 다른 인스턴스가 살아 있다면 flock 획득 실패가 정본 거부가 된다(SHALL).

프로세스 열거가 실패하면 부재를 주장할 수 없으므로 기존 거부 동작을 유지해야 한다(SHALL — 알 수 없을 때는 보수적으로 거부한다. 잘못 거부하면 운영자가 다시 시도하면 되지만, 잘못 허용하면 flock 하나에만 기대게 된다).

거부 안내는 약해져서는 안 된다(SHALL NOT): 엔진 프로세스가 실제로 관측되면 실행 중 인스턴스를 안내하며 거부하고, flock이 거부하면 flock의 사유를 안내한다(SHALL). 마커의 PID와 갱신 시각은 계속 안내 문구의 재료로 쓸 수 있으나 그것이 거부의 근거가 되어서는 안 된다(SHALL NOT).

이 요구사항은 stale 창(5분)·마커 갱신 주기(1분)·flock 배타를 바꾸지 않는다(SHALL NOT).

#### Scenario: 컨테이너 재생성이 남긴 유령 마커
- **WHEN** 엔진 프로세스가 없는 상태에서 마커가 stale 창 안의 갱신 시각과 이제는 존재하지 않는 PID를 담고 있고 autostart가 실행되면
- **THEN** 기동이 진행되고 "이미 실행 중"으로 거부되지 않는다

#### Scenario: 실제로 실행 중인 두 번째 인스턴스
- **WHEN** 엔진 프로세스가 관측되는 상태에서 기동이 다시 시도되면
- **THEN** 실행 중 인스턴스가 안내되고 기동이 거부된다

#### Scenario: 프로세스 열거 실패
- **WHEN** 마커가 신선하고 프로세스 열거가 오류를 반환하면
- **THEN** 기동이 거부되고 부재로 단정하지 않는다

#### Scenario: 마커는 없지만 flock을 다른 인스턴스가 쥐고 있다
- **WHEN** 마커가 없거나 stale인데 다른 인스턴스가 journal flock을 보유한 상태로 기동하면
- **THEN** flock 획득 실패가 거부 사유로 안내되고 두 번째 런타임이 기동하지 않는다

#### Scenario: 거부 근거는 마커가 아니다
- **WHEN** 기동 경로의 소스를 검사하면
- **THEN** 마커 신선도만으로 거부하는 경로가 존재하지 않는다

### Requirement: 엔진 프로세스 발견은 실제 명령줄과 소유 프로필에 일치한다

엔진 프로세스를 찾는 패턴은 이 바이너리가 실제로 spawn하는 명령줄에 일치해야 한다(SHALL — 콘솔은 자신의 `--config-dir`·`--session-file`을 자식 argv 앞에 붙이므로, 하위 명령만을 연속 문자열로 가정한 패턴은 콘솔이 띄운 엔진을 스스로 찾지 못한다). 패턴은 argv 토큰 경계를 지켜야 하며 다른 하위 명령(`console`·`httpapi`·`soak`)의 명령줄에 일치해서는 안 된다(SHALL NOT).

발견된 프로세스에 종료 시그널을 보내기 전에 그 프로세스가 이 콘솔이 소유한 엔진인지 판정해야 한다(SHALL — 소유의 기준은 journal 디렉터리다. 이 spec은 이미 인스턴스 배타를 journal 디렉터리 flock으로 정의하므로 같은 journal을 여는 프로세스만 같은 인스턴스다). 소유를 증명할 수 없는 프로세스에는 시그널을 보내서는 안 된다(SHALL NOT — 잘못 보낸 SIGTERM은 포지션을 지키던 다른 프로필의 엔진을 멈춘다).

소유 판정은 명령줄에서 되뽑은 설정 디렉터리를 콘솔 자신이 쓰는 것과 **같은 해석 경로**로 journal 디렉터리로 바꾼 뒤 비교해야 한다(SHALL — 기본 경로를 명시한 콘솔과 생략한 autostart는 같은 인스턴스이며, 플래그 문자열 비교는 그 경우를 다르다고 판정한다).

프로세스 열거 자체가 실패하면 부재를 주장할 수 없으므로 기존 거부 동작을 유지해야 한다(SHALL — a056이 정한 규칙을 바꾸지 않는다). 빈 목록과 열거 실패는 계속 구분되어야 한다(SHALL).

autostart 스크립트는 같은 후보 패턴을 사용하되 소유 판정을 수행하지 않아도 된다(MAY — 셸에서 판정을 재구현하면 Go와 어긋날 수 있고, 스크립트의 오탐이 만드는 결과는 "기동하지 않는다"뿐이라 보수적 방향이다). 이 요구사항은 flock 배타·stale 창·마커 갱신 주기를 바꾸지 않는다(SHALL NOT).

#### Scenario: 콘솔이 띄운 엔진을 콘솔이 찾는다
- **WHEN** 콘솔이 `--config-dir`와 `--session-file`을 갖고 엔진을 spawn한 뒤 프로세스를 조회하면
- **THEN** 그 엔진이 발견된다

#### Scenario: 정지 버튼이 도는 엔진을 세운다
- **WHEN** 그렇게 spawn된 엔진이 실행 중인 상태에서 정지를 요청하면
- **THEN** 그 프로세스에 종료 시그널이 가고 무엇을 세웠는지 안내된다 — "실행 중인 엔진을 찾지 못했다"로 끝나지 않는다

#### Scenario: 다른 프로필의 엔진은 건드리지 않는다
- **WHEN** 다른 journal 디렉터리로 실행 중인 엔진이 함께 관측되는 상태에서 정지를 요청하면
- **THEN** 그 프로세스에는 시그널이 가지 않는다

#### Scenario: 다른 하위 명령은 엔진이 아니다
- **WHEN** `console`·`httpapi`·`soak` 명령줄을 같은 패턴으로 검사하면
- **THEN** 어느 것도 엔진으로 발견되지 않는다

#### Scenario: 열거 실패는 부재가 아니다
- **WHEN** 프로세스 열거가 오류를 반환하고 마커가 신선하면
- **THEN** 기동이 거부되고 부재로 단정하지 않는다

### Requirement: strategy dispatch lease는 모든 안전 권한을 fenced 제출 직전에 재검증한다
Engine profile은 strategy dispatch lease의 모든 안전 권한과 owner fence를 제출 직전에 재검증해야 한다 (SHALL). Exposure-raising strategy attempt마다 candidate/evidence,
router/lane/version, campaign/leg, activation/calendar generations, exact `WIRED` ProtectionReady
attestation, reconciliation, risk reservation, Guardian decision/generation, build digest와
monotonic owner epoch/fencing token을 하나의 durable lease에 결합해야 한다 (SHALL).
ExecutionGateway 직전 검증은 journal의 current authority만 사용해야 하며 (SHALL), caller가
제공한 복제 상태나 생성 시점 검증으로 대체해서는 안 된다 (MUST NOT). Claim/validation은
lease를 비가역 소비해야 한다 (SHALL). Current `ISSUED` lease의 authority 누락·변경·만료,
stale epoch/token, scope mismatch와 pre-transport cancel은 lease/attempt `REFUSED`와 그 lease의
exact reservation `RELEASED`를 같은 journal transaction에서 영속하고 broker request를 0건으로
만들어야 한다 (SHALL). 이미 소비된 terminal lease replay는 retry attempt만 `REFUSED`하고 원래
lease/disposition을 변경하지 않으며, retry attempt의 별도 exact HELD reservation만 release해야
한다 (SHALL). 이
pre-transport failure들에 `AMBIGUOUS` 또는 `HELD`를 사용해서는 안 된다 (MUST NOT).

#### Scenario: activation manifest drift
- **WHEN** lane decision 뒤 dispatch 전에 해당 시장 activation manifest digest 또는 generation이 바뀐다
- **THEN** lease/attempt `REFUSED`와 exact reservation `RELEASED`를 원자 기록하고 broker request는 0건이며 effective entry를 OFF로 낮춘다

#### Scenario: 다른 시장 lease
- **WHEN** KR decision을 US calendar 또는 ProtectionReady generation에 결합된 lease로 제출한다
- **THEN** scope 불일치로 lease와 exact reservation을 `REFUSED + RELEASED` 처리하고 broker 호출 전에 거부한다

#### Scenario: durable lease 없는 제출
- **WHEN** GuardianDecision은 있지만 strategy dispatch lease가 없는 exposure-raising 제출을 시도한다
- **THEN** Engine profile과 Gateway는 typed refused attempt를 영속하고 해당 attempt에 결합된 exact reservation이 있으면 원자 RELEASED하며 broker request와 합성 lease는 0건이다

#### Scenario: validation failure 뒤 원상 복구
- **WHEN** 한 validation이 generation mismatch로 실패한 뒤 current 값이 lease preimage 값으로 돌아온다
- **THEN** terminal lease는 부활하지 않고 fresh decision과 fresh lease만 새 claim을 허용한다

#### Scenario: 만료 또는 stale fence
- **WHEN** lease가 만료됐거나 owner epoch/fencing token이 current durable state보다 stale이다
- **THEN** `REFUSED + RELEASED`를 원자 기록하고 broker request는 0건이며 `AMBIGUOUS`로 분류하지 않는다

### Requirement: market worker 장애는 그 시장 entry만 격리한다
Engine supervisor는 market worker 장애를 그 시장 entry scope에 격리해야 한다 (SHALL). KR 또는
US entry worker의 OFF, market wait, stale evidence, budget defer, cycle failure, panic, abnormal
return, watchdog expiry와 반복 crash를 해당 시장의 effective entry OFF latch와 bounded restart로
한정해야 한다 (SHALL). Peer market evaluation과 Reconcile driver, fill detector, protection
supervisor, exit observer, emergency reduction loop는 계속 실행해야 한다 (SHALL). Market worker
장애만으로 전체 process 또는 peer market을 종료해서는 안 된다 (MUST NOT).

#### Scenario: US entry worker abnormal return
- **WHEN** US entry worker가 비정상 반환하지만 safety loop와 KR worker는 정상이다
- **THEN** US entry만 typed OFF latch로 강화하고 bounded restart하며 KR evaluation과 모든 safety loop를 유지한다

#### Scenario: automation OFF
- **WHEN** automation effective state가 OFF로 전환된다
- **THEN** KR·US 신규 entry와 scale-in은 0건이고 fill, reconciliation, protection과 reduce-only exit는 계속된다

### Requirement: central integrity fault는 외부 fenced safety fallback으로 복구된다
Engine deployment는 central integrity fault를 외부 fenced safety fallback으로 복구해야 한다 (SHALL). Journal corruption, Gateway invariant violation, owner epoch/fence CAS 불능 또는 복수
current owner가 감지되면 모든 신규 entry를 즉시 차단하고 critical alert를 발행해야 한다
(SHALL). 별도 deployment domain의 external supervisor는 이전 owner token을 fence한 새 epoch로
entry capability가 없는 safety-only fallback을 versioned `safety_fallback_rto` 안에 기동해야
하며 (SHALL), 그 RTO는 60초를 초과해서는 안 된다 (MUST NOT). Fallback은
fill/reconciliation/protection/reduce-only exit/emergency reduction만 수행하고 entry lease를
발급해서는 안 된다 (MUST NOT).

#### Scenario: central dispatch owner integrity 상실
- **WHEN** current owner fence가 손상되거나 두 owner가 current라고 주장한다
- **THEN** 모든 entry를 차단하고 stale token을 broker 전에 거부하며 external supervisor가 60초 이하의 frozen RTO 안에 fenced safety-only fallback을 시작한다

#### Scenario: fallback 기동 실패
- **WHEN** external supervisor가 frozen RTO 안에 safety-only fallback을 기동하지 못한다
- **THEN** broker-resident protection을 자동 취소하지 않고 `SAFETY_FALLBACK_UNAVAILABLE` critical state를 지속 발행하며 신규 entry는 0건이다

### Requirement: 사람이 읽는 알림은 한국어로 어느 종목인지 말한다

알림의 제목과 본문은 한국어여야 한다(SHALL). 종목을 가리키는 알림은 그 종목의
이름과 코드를 함께 제시해야 한다(SHALL — `이름(코드)` 형식).

종목 이름은 계좌 보유 조회가 이미 반환하는 값에서 얻어야 하며, 이름을 얻기 위해
별도의 브로커 요청을 추가해서는 안 된다(SHALL NOT — §0.4). 이름을 알 수 없으면 코드만
제시해야 하며(SHALL), 이름을 추정하거나 다른 출처로 대체해서는 안 된다(SHALL NOT).

알림의 구조화 payload와 구조화 로그 필드는 기계 판독 표면이므로 영문 키와 원문 값을
유지해야 한다(SHALL). 이 요구사항은 사람이 읽는 제목·본문에만 적용된다.

알림에 계좌번호·잔고·세션·자격증명을 포함해서는 안 된다(SHALL NOT — §0.8, 기존 규칙 유지).

#### Scenario: 보유 종목에 대한 알림
- **WHEN** 보유 중인 종목에 대해 알림이 발송되고 그 종목의 이름이 계좌 보유 조회로 알려져 있다
- **THEN** 제목과 본문은 한국어이고 종목은 이름과 코드를 함께 제시한다

#### Scenario: 이름을 알 수 없는 종목
- **WHEN** 알림 대상 종목의 이름이 알려져 있지 않다
- **THEN** 코드만 제시하고 이름을 추정하지 않으며, 이름을 얻기 위한 추가 브로커 요청을 하지 않는다

#### Scenario: 기계 판독 표면
- **WHEN** 알림이 outbox에 기록되고 구조화 로그가 남는다
- **THEN** payload와 로그 필드의 키와 값은 영문·원문 그대로이고 한국어화되지 않는다

### Requirement: 같은 조건의 critical 알림은 재알림 창 안에서 한 번만 전송한다

critical 알림이 전달된 뒤 재알림 창이 지나기 전에 같은 event key의 조건이 다시 관측되면 그 알림을 다시 전송해서는 안 된다(SHALL NOT).

outbox의 중복 제거가 **행에만** 적용되고 전송에는 적용되지 않으면, 같은 조건을 관측하는
매 사이클이 새 push를 만든다. 그때 "한 조건은 한 알림"이라는 계약은 원장 안에서만 참이고
운영자의 기기에서는 거짓이다 — 그리고 그 상태에서 알림 채널은 신호가 아니라 소음이 되어,
정작 다른 critical 알림이 그 사이에 묻힌다.

억제는 **창**이어야 하며 영구적이어서는 안 된다(SHALL NOT). 창이 지나면 같은 조건은 다시
전송되어야 하고(SHALL), 그 재전송은 최초 전달과 같은 경로를 걸어야 한다(SHALL) — 재시도
예산, 전달 실패 시의 진입 차단, 운영 모드 승격이 모두 그대로 적용된다.

영구 억제가 금지되는 이유는 둘이며 둘 다 안전 문제다. 첫째, 반복되는 조건의 전송은 알림
경로가 **아직 살아 있다는 유일한 주기적 증거**이므로, 영구히 억제하면 transport가 죽어도
아무도 모르는 채 엔진이 계속 거래한다. 둘째, event key는 조건을 담고 **원인을 담지
않으므로**, 한 원인으로 전달된 알림이 같은 key의 다른 원인을 영구히 가린다.

아직 전달되지 않은(PENDING) 행은 창과 무관하게 계속 재시도해야 한다(SHALL).
전송 실패는 중복이 아니라 미완이다.

인식되지 않는 outbox 상태는 전송이 필요한 것으로 취급하고, 정상 `PENDING` 전달 경로로
재무장해야 한다(SHALL) — 상태 열에 CHECK 제약이 없으므로 모르는 값은 이 빌드가 이해하지
못하는 행이고, "운영자가 받았는지 모른다"의 안전한 해석은 보내는 것이다. 재무장 없이
발행만 하면 전달 완료 표시가 실패해 다음 관측이 다시 발행하므로 허용되지 않는다(SHALL NOT).

durable 기록은 그대로 유지해야 한다(SHALL): 조건이 다시 관측되면 outbox는 같은 행을
돌려주고, 구조화 로그는 관측마다 한 줄을 남긴다. 억제되는 것은 **전송뿐**이며 관측
사실의 기록이 아니다.

#### Scenario: 전송에 성공한 조건이 창 안에서 다시 관측된다
- **WHEN** 같은 event key의 critical 이벤트가 전송 성공 뒤 재알림 창 안에서 다시 관측되면
- **THEN** outbox 행은 하나로 유지되고 새 전송은 발생하지 않으며, 구조화 로그에는 관측마다 한 줄이 남는다

#### Scenario: 전송에 성공한 조건이 창을 넘겨 다시 관측된다
- **WHEN** 재알림 창이 지난 뒤 같은 조건이 다시 관측되면
- **THEN** 같은 행에 대해 다시 전송하고, 그 전송이 재시도 한도까지 실패하면 신규 진입이 차단된다

#### Scenario: 전달된 뒤 알림 경로가 죽는다
- **WHEN** 알림이 한 번 전달된 뒤 transport가 죽고 같은 조건이 창을 넘겨 다시 관측되면
- **THEN** 그 재전송이 실패하며 진입 차단과 운영 모드 승격이 최초 전달 실패와 똑같이 일어난다

#### Scenario: 같은 key의 다른 원인이 나중에 발생한다
- **WHEN** 한 원인으로 전달된 알림과 같은 event key를 갖는 다른 원인의 조건이 창을 넘겨 발생하면
- **THEN** 그 발생도 운영자에게 전송된다

#### Scenario: 전송에 실패한 조건이 다시 관측된다
- **WHEN** 첫 전송이 재시도 한도까지 실패해 행이 PENDING으로 남은 뒤 같은 조건이 다시 관측되면
- **THEN** 창과 무관하게 같은 행에 대해 전송을 다시 시도하고, 성공하면 전달 완료로 표시한다

#### Scenario: 인식되지 않는 outbox 상태를 만난다
- **WHEN** 이 빌드가 아는 상태가 아닌 값을 가진 행에 대해 전송 필요 여부를 물으면
- **THEN** 전송이 필요한 것으로 답하고 행을 PENDING으로 재무장해 전달 완료 표시가 성공한다

### Requirement: 한 조건의 동시 관측은 한 번만 전송한다

같은 event key를 동시에 관측한 둘 이상의 경로가 각각 전송해서는 안 된다(SHALL NOT).

전송 필요 여부의 판정과 그에 따른 전송은 **하나의 배타 구간** 안에서 일어나야 한다(SHALL).
판정만 원장 트랜잭션으로 감싸고 전송을 그 밖에 두면, 두 관측이 아직 전달되지 않은 같은
행을 읽고 둘 다 "전송 필요"로 판정한 뒤 차례로 전송한다. 두 번째 전송은 이미 전달된 행을
표시하려다 실패하며, 그 실패 로그가 폭주의 유일한 흔적이 된다.

이 배타 구간은 outbox 백로그를 비우는 경로에도 같이 적용되어야 한다(SHALL) — 그 경로와
관측 경로가 같은 행을 동시에 발행할 수 있기 때문이다.

#### Scenario: 한 조건이 동시에 관측된다
- **WHEN** 같은 event key의 critical 이벤트를 여러 경로가 동시에 관측하면
- **THEN** 전송은 한 번만 발생하고, 이미 전달된 행에 전달 표시를 다시 시도하는 일이 없다

### Requirement: 중복 제거 계약은 전송 횟수로 검증한다

"한 조건은 한 알림" 계약의 자동 검증은 **전송 횟수**를 조건으로 삼아야 한다(SHALL).

outbox 행 수만 세는 검사는 이 계약을 검증하지 못한다(SHALL NOT — 행 수는 `event_key`의
UNIQUE 제약이 이미 보장하므로, 그 검사는 스키마를 확인할 뿐 전송 경로를 통과하지 않는다.
전송이 사이클마다 발생하는 동안에도 그런 검사는 통과한다).

분기 커버리지를 그 근거로 삼아서도 안 된다(SHALL NOT — `covermode=set`은 블록의 실행
여부를 기록하고 횟수를 세지 않으므로, 폭주하는 코드와 한 번만 도는 코드가 같은 값을 낸다).

#### Scenario: 같은 조건이 여러 번 관측된다
- **WHEN** 전송이 성공하는 상태에서 같은 event key의 critical 이벤트를 한 창 안에서 여러 번 관측하면
- **THEN** 전송 횟수는 1이고 outbox 행은 1이다

### Requirement: 재무장된 outbox 행은 통째로 이번 에피소드를 말한다

재무장되는 행은 **정체성을 제외한 모든 값**이 이번 관측의 것으로 다시 쓰여야 한다(SHALL).
정체성은 event key·event type·등급·최초 생성 시각이며, 그 밖의 제목·본문·payload·전달 시도
횟수·마지막 시도 시각·전달 시각은 전부 에피소드의 속성이다.

이전 에피소드의 값을 남겨서는 안 된다(SHALL NOT). event key는 조건을 담고 원인을 담지
않으므로, 같은 key로 재무장된 행이 이전 원인의 본문을 그대로 들고 있으면 그 행은 **아직
전달되지 않은 알림의 내용이 지금 일어나는 일과 다르다**고 말한다.

전달 시각과 마지막 시도 시각도 함께 지워야 한다(SHALL). 본문을 이번 관측의 것으로 바꾼
행이 이전 전달 시각을 들고 있으면, 그 행은 **지금 담고 있는 내용이 그때 전달됐다**고
말하며 그것은 거짓이다. 한 행이 증거이려면 모든 칸이 같은 사건을 가리켜야 한다.
두 에피소드를 한 행에 담을 수는 없다.

과거 에피소드의 기록은 구조화 로그가 보관한다. outbox 행은 감사 로그가 아니라 **전달
과제**이며, 전달 시각의 기능적 독자는 재알림 창 계산 하나뿐이고 그것은 settled 상태에서만
읽힌다.

#### Scenario: 다른 원인으로 같은 조건이 창을 넘겨 다시 관측된다
- **WHEN** 한 원인으로 전달된 뒤 재알림 창이 지나고, 같은 event key를 갖는 다른 원인의 관측이 그 행을 재무장하면
- **THEN** 그 행의 제목·본문·payload는 이번 관측의 것이고, 전달 시도 횟수는 0이며, 전달 시각과 마지막 시도 시각은 비어 있다

#### Scenario: 재무장된 행을 backlog 비우기 경로가 보낸다
- **WHEN** 재무장된 행을 outbox backlog를 비우는 경로가 전송하면
- **THEN** 운영자가 받는 제목과 본문은 그 행을 재무장한 관측의 것이다

### Requirement: 기록하지도 전송하지도 못한 critical 알림은 재시작을 넘겨 진입을 막는다

critical 알림의 outbox 기록 시도 자체가 실패하면, 신규 진입을 차단해야 한다(SHALL).
그 실패를 구조화 로그로 남겨야 한다(SHALL).

그 차단은 **프로세스 재시작을 넘겨 살아남아야 한다**(SHALL). 진입 게이트의 래치는 메모리에만
있고, 기록이 실패한 경우에는 원장에 알림 행조차 없다. 따라서 메모리 래치만 남기면 재시작
한 번으로 차단과 알림이 함께 사라지고, 운영자가 아무것도 받지 못한 채 신규 진입이 다시
열린다. 그 상태는 이 요구사항이 막으려는 바로 그 상태다.

내구적 차단을 남기려는 시도 자체가 실패할 수 있다는 이유로 그 시도를 생략해서는 안 된다
(SHALL NOT). 기록 실패의 원인이 원장 장애라면 내구적 차단도 실패하지만, 그 실패는 로그로
남아 "재시작하면 차단이 풀린다"는 사실 자체의 기록이 된다. 실패한 시도가 침묵보다 낫다.

전송 실패가 진입을 막는 이유는 "운영자가 이 사건을 못 받았다"이며, 기록 실패는 그보다
이르고 더 나쁘다 — 전송되지 않았고 나중에 재시도할 근거조차 남지 않았다.

호출자가 그 오류를 검사하는지에 결과가 의존해서는 안 된다(SHALL NOT). 오류를 반환하는
것으로 충분하다고 보면 그 오류를 버리는 호출자 하나가 이 요구사항 전체를 무효로 만든다.

이 요구사항이 미치는 범위는 **배선된 것까지다.** 진입 게이트가 배선되지 않은 알림 경로는
무엇도 차단할 수 없고, 계정 참조가 없는 경로는 내구적 전환을 기록할 곳이 없다. 그런 조립은
이 요구사항을 위반하는 것이 아니라 **이 요구사항의 보호를 받지 못하는 것**이며, 그 사실은
조립 지점의 책임이다. 다만 어떤 조립에서도 그 실패가 호출자에게 오류로 반환되는 것은
멈춰서는 안 된다(SHALL NOT) — 그것이 배선과 무관하게 남는 마지막 통지 경로다.

청산 경로는 이 차단의 영향을 받아서는 안 된다(SHALL NOT — 차단은 신규 진입 전용이며
손절·비상 청산의 즉시성은 어떤 알림 실패로도 약해지지 않는다).

#### Scenario: outbox 기록 트랜잭션이 실패한다
- **WHEN** critical 이벤트의 outbox 기록이 실패하면
- **THEN** 신규 진입이 차단되고, 그 차단을 재시작 뒤에도 남기려는 시도가 이루어지며, 실패가 구조화 로그에 남는다

#### Scenario: 원장 장애로 내구적 차단마저 실패한다
- **WHEN** outbox 기록과 내구적 차단이 같은 원장 장애로 모두 실패하면
- **THEN** 그 사실이 오류 수준으로 기록되고, 이 프로세스의 진입 차단은 그대로 유지된다

#### Scenario: 호출자가 알림 오류를 버린다
- **WHEN** critical 알림의 오류를 검사하지 않는 호출자에서 outbox 기록이 실패하면
- **THEN** 신규 진입은 그대로 차단된다

#### Scenario: 게이트도 로거도 계정 참조도 배선되지 않은 알림 경로에서 기록이 실패한다
- **WHEN** 선택 배선이 모두 비어 있는 알림 경로에서 outbox 기록이 실패하면
- **THEN** 패닉하지 않고 그 실패를 오류로 반환하며, 차단할 게이트가 없다는 사실은 조립 지점의 책임으로 남는다

### Requirement: 시한 재알림과 상태 복구는 서로 다른 규칙이다

인식된 settled 행은 재알림 창이 비활성인 호출자에 대해 재무장되어서는 안 된다(SHALL NOT).
기록만 하고 전달하지 않는 호출자가 그런 호출자다.

같은 호출자에 대해서도, **인식되지 않는 상태**의 행은 재무장해야 한다(SHALL). 그것은
시한 재알림이 아니라 상태 복구다 — 모르는 상태는 전달됐다는 증거가 아니며, 복구하지
않으면 이후의 전달 완료 표시가 그 행을 꺼내지 못해 관측할 때마다 다시 발행된다.

두 규칙은 문서에 함께 적혀야 한다(SHALL). 한쪽만 적힌 문서는 다른 쪽 동작을 결함으로
보이게 하고, 그 오독을 근거로 안전한 복구 동작이 제거될 수 있다.

#### Scenario: 기록만 하는 호출자가 전달된 행을 다시 기록한다
- **WHEN** 재알림 창이 비활성인 호출자가 이미 전달된 행과 같은 event key를 기록하면
- **THEN** 그 행은 재무장되지 않고 전송도 필요하지 않다고 답한다

#### Scenario: 기록만 하는 호출자가 인식되지 않는 상태의 행을 만난다
- **WHEN** 재알림 창이 비활성인 호출자가 이 빌드가 모르는 상태의 행과 같은 event key를 기록하면
- **THEN** 그 행은 PENDING으로 복구된다

### Requirement: 동시성 계약의 검증은 시간이 아니라 사건을 근거로 한다

배타 구간을 검증하는 자동 테스트는 경과 시간을 통과 조건으로 삼아서는 안 된다(SHALL NOT).
부하가 걸린 기계에서 거짓 통과하며, 거짓 통과는 그 구간이 사라진 뒤에도 초록이다.

배타 구간의 검증은 **관측 가능한 사건**을 근거로 해야 한다(SHALL) — 동시 진입이 실제로
일어났는가, 배타여야 할 두 작업의 효과가 원장에 섞였는가.

동시 관측을 검증하는 테스트는 참여 goroutine을 공통 출발점에서 동시에 출발시켜야 한다
(SHALL). 그러지 않으면 스케줄러가 직렬화한 실행도 통과시키며, 그때 그 테스트는 배타
구간이 아니라 스케줄러를 검증한 것이다.

알림 전달의 배타 구간을 지키는 잠금에는 **그것을 제거하면 실패하는 테스트**가 있어야
한다(SHALL). 잠금은 분기가 아니므로 커버리지로는 그 부재를 볼 수 없다 — `n.mu.Lock`은
필요하든 아니든 실행된 것으로 기록된다.

그 요구는 **뮤테이션 실측으로** 충족해야 한다(SHALL): 잠금을 제거하고 테스트를 반복
실행해 탐지 횟수를 세고, 그 수를 근거로 남긴다. "이 테스트가 그 잠금을 지킨다"는 주장을
실행 없이 적어서는 안 된다(SHALL NOT).

리뷰가 제안한 테스트 수정이 탐지율을 개선하지 않으면 **개선하지 않았다고 기록해야
한다**(SHALL). 제안을 반영했다는 사실이 그 제안이 들었다는 증거가 아니다.

잠금의 필요성을 어떤 방법으로도 재현 가능하게 보일 수 없으면, 보이지 못했다고 기록해야
한다(SHALL). 개선된 테스트를 완결된 증명으로 세어서는 안 된다(SHALL NOT).

#### Scenario: 배타 구간의 잠금을 제거한다
- **WHEN** 전송 경로와 backlog 비우기 경로가 공유하는 잠금을 제거하면
- **THEN** 적어도 하나의 테스트가 실패한다

#### Scenario: 단일 코어에서 동시 관측 테스트를 돌린다
- **WHEN** 병렬 실행이 사실상 불가능한 환경에서 동시 관측 테스트를 돌리면
- **THEN** 그 테스트는 경합 창을 스케줄러의 우연이 아니라 구성으로 열어, 직렬 실행만으로는 통과하지 않는다

#### Scenario: 리뷰가 제안한 수정이 탐지율을 올리지 않는다
- **WHEN** 리뷰가 지시한 테스트 수정을 반영하고 뮤테이션 탐지율을 재면 개선이 없으면
- **THEN** 개선이 없었다는 사실을 수치와 함께 남기고, 실제로 탐지율을 올리는 다른 수정을 찾는다

#### Scenario: 잠금의 필요성을 재현 가능하게 보일 수 없다
- **WHEN** 어떤 잠금의 제거를 반복 실행으로 잡아내는 테스트를 만들 수 없으면
- **THEN** 그 잠금은 미검증으로 기록되고, 개선된 테스트는 완결로 세지 않는다

### Requirement: 내구 알림은 발송 주체를 가져야 한다

엔진은 outbox에 PENDING으로 남은 critical 알림을 **주기적으로 재시도하는 실행자**를 가져야 한다(SHALL).
기록만 하고 발송하지 않는 구성은 **운영자에게 도달하지 않는
경보를 내구화한 것**이고, 그것은 경보가 아니라 기록이다.

이 요구는 발송의 **성공**을 요구하지 않는다. 전송 수단이 죽어 있으면 행은 PENDING으로
남고 진입 게이트는 잠긴 채로 있다 — 그것이 기존 동작이며 a098은 그것을 바꾸지 않는다.
요구하는 것은 **시도하는 주체의 존재**다.

#### Scenario: 발송자 없이 기록만 하는 구성은 요구를 만족하지 않는다

- **WHEN** critical 알림이 outbox에 기록되고 진입 게이트가 잠긴다
- **THEN** 그 행을 재시도하는 실행자가 **있어야 한다**
- **AND** 전송 수단이 살아 있으면 그 행은 유한한 시간 안에 DELIVERED가 되어야 한다

#### Scenario: 운영자는 밀린 것을 보고 해제할 수 있어야 한다

- **WHEN** 진입 게이트가 미전달 알림 때문에 잠겨 있다
- **THEN** 운영자가 밀린 행을 **읽을 수 있어야 하고** 승인으로 게이트를 풀 수 있어야 한다
- **AND** 그 해제는 **타이핑 확인이나 추가 승인 마찰을 요구하지 않는다**

### Requirement: 재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다

기동 시 미전달 critical 알림이 남아 있으면 신규 진입을 **다시 차단해야 한다**(SHALL).
그 복원은 **어떤 루프보다 먼저** 끝나야 한다 —
진입이 가능해진 뒤에 차단이 걸리면 그 사이가 열려 있다.

래치만으로 차단을 들고 있으면 **재시작이 그것을 지운다.** 원장에는 아직 아무에게도
도달하지 않은 경보가 남아 있는데 새 프로세스는 그것을 모른 채 진입을 연다.
그 상태에서 운영자는 **아무것도 안 했는데 차단이 풀렸다는 사실조차 모른다**.

**복원은 차단하는 방향으로만 한다**(SHALL). 기동 시 미전달 수가 0이라는 사실을
근거로 **차단을 푸는 일은 하지 않는다**(SHALL NOT) — 푸는 근거는 운영자의 승인이고,
그 규범은 이 델타가 안 바꾼다.

#### Scenario: 미전달 알림을 남긴 채 엔진을 재시작한다

- **WHEN** 미전달 critical 알림이 원장에 남은 상태에서 프로세스가 재시작된다
- **THEN** 신규 진입은 **다시 차단된 상태로 올라온다**
- **AND** 그 차단은 **어떤 진입 판정보다 먼저** 걸려 있다
- **AND** 그 차단은 **운영자가 승인할 때** 풀린다 — 발송이 성공하는 것만으로는 안 풀린다
- **AND** **재시작만으로는 풀리지 않는다**

### Requirement: 운영자가 밀린 알림을 읽고 승인하는 경로가 존재해야 한다

운영자는 밀린 critical 알림을 **읽고 승인할 수 있어야 한다**(SHALL).
승인 경로가 없으면 진입 차단을 푸는 수단이 **재시작뿐**이고, 위 Requirement가
그 재시작을 막으므로 **차단이 영구가 된다**.

그 표면은 **타이핑 확인이나 추가 승인 마찰을 요구해서는 안 된다**(SHALL NOT).
승인 기록은 **운영자의 이름을 남겨야 한다**(SHALL) — 기계가 증명할 수 없던 것을
사람이 단언하는 자리이고, audit trail이 요점이다.

표면이 보여 주는 것에 **알림 본문·계좌 식별자·임차 토큰 원문이 들어가서는 안 된다**(SHALL NOT).

#### Scenario: 운영자가 밀린 알림을 해제한다

- **WHEN** 미전달 critical 알림 때문에 진입이 차단되어 있다
- **THEN** 운영자가 그 목록을 **읽을 수 있다** — 행 id · 나이 · 시도 수 · 임차 보유자
- **AND** 행을 승인하면 **그 이름이 원장에 남는다**
- **AND** **승인했고 미전달 수가 0일 때** **전달 실패로 걸린** 진입 차단이 풀린다 —
  두 조건이 다 필요하다. **다른 사유로 걸린 차단은 이 승인이 안 푼다**
- **AND** 그 해제는 **엔진 프로세스 안에서** 일어난다 — 원장만 고치면 게이트는 안 풀린다
- **AND** 승인 자체에 **추가 확인 절차가 없다**

#### Scenario: 발송 중인 행도 운영자가 승인할 수 있다

- **WHEN** 어떤 발송자가 임차를 들고 그 행을 보내는 중이다
- **THEN** 운영자의 승인은 **성공한다** — 임차가 그것을 막지 않는다
- **AND** 그 발송자의 정산은 **거부된다**
- **AND** 남은 임차 잔재는 그 행이 다시 무장될 때 **지워진다**

### Requirement: 배달 실행자의 정지가 다른 루프를 내려서는 안 된다

배달 실행자의 `Run`이 반환해도 **다른 감독 루프는 계속 돌아야 한다**(SHALL).
알림 경로의 결함이 exit 관측 루프를 멈추면 **알림 버그가 손절 부재가 된다.**

> **a092 22판이 이 정본 요구를 MODIFIED로 싣는다 (21라운드 C10, Manager 판정).** 규범 문장은 한 글자도 바꾸지 않았다. 바꾼 것은 근거 ①(투영 배선 뒤 거짓이 된다) · 낡은 좌표(당시 좌표를 두고 지금 이름을 병기 — 새 코드 좌표는 넣지 않는다) · 그리고 a098 작성 당시의
> 「a092가 진다」 서술 둘에 단 23판 정정 표지다(22라운드 K9) — 델타 밖 편집은 archive가 보장하지 않는다.

**배달 실행자는 감독 루프가 아니어야 한다**(SHALL NOT — 사용자 결정 9-2).
그 성질을 **판정으로 만들어서는 안 된다**(SHALL NOT): 감독 루프로 등록한 뒤
*"이 이름이면 다르게 다룬다"*로 거르는 구성은 **판정이 틀리면 무너진다.**
등록하지 않으면 **틀릴 판정이 없다.**

> **⛔ 6판이 이 문단의 두 번째 근거를 지웠다.** 5판까지는 여기에
> *"엔진 런타임의 루프 집합을 **셋**으로 못 박은 승인된 정본과도 어긋난다"*가
> 붙어 있었다. **그 근거는 쓸 수 없다** — 정본의 셋은 프로덕션의 넷과 이미 다르고
> (당시 `cmd/tossctl/engine.go:377-398` — 지금은 생산 조립의 `Loops:` 목록), 이 델타의 `MODIFIED`가 그것을 **넷으로 고친다**
> (사용자 결정 10-1). 고치는 문장을 동시에 근거로 인용할 수는 없다.
>
> 남은 근거 하나로 충분하다: **틀릴 판정을 안 만든다.** 그것은 정본이 무엇을
> 적고 있든 참이다.

**패닉도 정지로 다뤄야 한다**(SHALL). 배달 실행자의 패닉이 프로세스를 죽이면
그 순간 손절 루프도 함께 죽고, 위 첫 문장이 **패닉 경로에서 거짓**이 된다.

정지는 조용해서는 안 된다(SHALL NOT). 정지 사실은 구조화 로그로 남아야 하고,
**진입 게이트는 잠겨야 한다.**

**정상 종료를 정지로 오인해서는 안 된다**(SHALL NOT). 런타임 취소로 실행자가
반환하는 것은 죽음이 아니다. 그것으로 게이트를 잠그면 **다음 기동이 아무 이유 없이
운영자 승인을 요구한다.**

게이트 래치는 `EntryGate.Block`으로 **직접** 세워야 한다(SHALL).
운영 모드 승격(`EscalateOperatingMode`)에 **기대서도 안 되고, 함께 하지도 않는다**(SHALL NOT — 사용자 결정 11-1).

> **근거 둘.** ① (a092 22판 정정) 이 요구가 처음 쓰일 때는 운영 모드 투영이 생산에 배선되지 않아
> 그 승격이 산 프로세스의 진입 게이트에 닿지 않았다. a092가 투영을 배선한 뒤에는 승격이 **실제로 진입을 막는다** —
> 그러므로 이 요구의 규범(승격하지 않는다)은 근거 ①이 아니라 아래 ②로 선다. **a098은 그 배선을 기다리지 않았다.**
> ② 승격은 원장에 남고 완화에 **사람 승인**이 필요하므로
> (정본 risk-management 「운영 모드 전환은 방향 비대칭이다」 — 완화는 사람 승인, 23판 좌표 `:133`), 실행자가 살아난 뒤에도
> 진입이 막힌 채가 된다 — **아래 Scenario의 「복구는 재시작」이 거짓이 된다.**

#### Scenario: 배달 실행자가 죽는다

- **WHEN** 배달 실행자의 `Run`이 취소가 아닌 사유로 반환하거나 패닉한다
- **THEN** exit 관측 루프는 **계속 돈다**
- **AND** 진입 게이트가 잠긴다
- **AND** 구조화 로그 한 줄이 남는다
- **AND** 죽은 실행자는 **자동으로 되살아나지 않는다**

> **✅ 6판 — 사용자 결정 11-1이 네 번째 조항을 지웠다.** 5판까지 여기에는
> *"운영 모드가 `ENTRY_BLOCKED`로 승격된다"*가 있었다. **없앤다.**
>
> **지운 것이 무엇을 잃게 하는지 먼저 적는다 — 5라운드가 옳았던 부분이다.**
> `EntryGate.Block`은 `g.mu`와 맵뿐이라(당시 `retry.go:498-505` — 지금은 `EntryGate.Block`)
> **재시작이 그 래치를 지운다.** 운영 모드는 `tx.Commit()`으로 남는다
> (`operating_mode.go:468`). 그래서 *"래치만으로는 부족하다"*는 **참이다.**
>
> **그런데 그 구멍은 이 델타의 다른 요구가 이미 막는다.** 위의
> 「재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다」가 기동 시 **원장의
> 미전달 행을 보고 다시 차단한다**(4.6). 실행자가 죽어 있었다면 그 죽음이
> 남긴 것은 **미전달 행**이고, 재시작은 그 행을 보고 잠근다.
>
> | 재시작 시점 | 5판(모드 승격 포함) | **6판(래치만)** |
> |---|---|---|
> | 미전달 행이 남아 있다 | 막힌다 | **막힌다** — 4.6의 원장 복원 |
> | 미전달 행이 0인데 실행자만 죽어 있었다 | **막힌 채로 남는다 — 사람 승인 전까지** | **열린다** |
>
> **아래 칸이 결정 11-1이 고른 것이다.** 새 프로세스의 배달 실행자는 **살아 있다.**
> 차단의 사유가 *"보낼 주체가 없다"*였으므로 주체가 생긴 시점에 사유가 소멸한다.
> 모드로 승격하면 사유가 사라진 뒤에도 **사람 승인 없이는 안 풀린다**(정본 risk-management 「운영 모드 전환은 방향 비대칭이다」 — 완화는 사람 승인, 당시 좌표 `:102-108`) —
> 그것은 *"복구는 재시작"*을 **거짓으로 만든다.**
>
> **덤으로 얻은 것 셋.** `internal/journal` diff가 **비어 있는 채로 남고**(§5.2),
> 승인된 `risk-management`에 **`MODIFIED`가 필요 없고**,
> 자동 트리거의 닫힌 열거(`operating_mode.go:75-96`·`:513-547`)를 **안 건드린다.**
>
> **잃은 것 하나는 a092가 진다** — 결정 12-2. a092의 같은 Scenario
> (`a092/specs/engine-safety/spec.md:126-130`)는 모드 승격을 SHALL로 적고 있고,
> **a098은 그것을 안 진다.** 인계는 a099 §7.5.
>
> > **⛔ a092 23판 정정 (22라운드 K9)** — 위 문단은 a098 작성 당시의 기록이다. a092는 20판에서 결정 11-1을 적용해 그 Scenario의 모드 승격을 지웠고,
> > 22판에서 실행자 사망 Scenario를 이 정본에 맡겼다. **a092가 지는 것은 없다** — 실행자 사망의 잔여는 a092 design D0.3g 7이 보상 통제와 함께 수용했다.

#### Scenario: 런타임 취소로 실행자가 반환한다

- **WHEN** 런타임이 취소되어 배달 실행자가 그 취소를 반환한다
- **THEN** 진입 게이트는 **안 잠긴다**
- **AND** critical 알림이 **안 나간다**
- **AND** 런타임은 그 실행자가 **반환할 때까지 기다린 뒤** 원장을 닫는다

> ## ⚠⚠ 19라운드 B-P5 — 같은 제목의 Scenario가 두 change에서 조항이 달랐다
>
> a092의 같은 Scenario(`a092/specs/engine-safety:126-130`)는 **다섯 조항**을
> 요구한다. a098은 **셋만** 적고 있었고, 빠진 둘이 위에 더한 것이다.
> 그리고 a092 §6.11이 이 Scenario를 **통째로 a098의 R2·R3에 매핑**한다 —
> 즉 **두 조항이 어느 change도 안 지는 상태였다.**
>
> §6.11이 22 = 22의 **제목 대칭**을 기계로 확인했는데, 그 검사는
> **조항을 안 본다.** 대칭이 초록이면서 요구가 새는 것이 가능하고,
> 이 자리가 그 실증이다(a092 task 10.4.4가 검사를 조항 단위로 확장한다).
>
> **왜 게이트 래치만으로 부족한가 — 측정했다.**
>
> | | 어디 | 재시작 후 |
> |---|---|---|
> | `EntryGate.Block` | `EntryGate.Block`(당시 `retry.go:498-505`) — **`g.mu`와 맵뿐, 원장에 안 닿는다** | **사라진다** |
> | 운영 모드 승격 | `TransitionOperatingMode`가 `tx.Commit()`(`operating_mode.go:468`) | **남는다** |
>
> 배달 실행자가 죽고 게이트만 잠긴 상태에서 프로세스를 재시작하면
> **미전달 critical 알림이 그대로인데 신규 진입이 다시 열린다.**
> 그것이 a092의 셋째 조항이 있던 이유다.
>
> > **⛔ a092 23판 정정 (22라운드 K9)** — 이 블록이 인용한 a092 Scenario(`a092/specs/engine-safety:126-130`)는 a092 델타에서 사라졌다(22판).
> > 「두 조항이 어느 change도 안 지는 상태」는 a098 작성 당시의 기록이다.
>
> > **⛔ 6판이 이 블록의 결론을 뒤집었다 — 사용자 결정 11-1.**
> >
> > 19판은 여기서 *"**둘 다 한다**: 게이트는 `EntryGate.Block`으로 직접,
> > 모드는 원장에 남긴다"*로 끝냈다. **뒤엣것을 안 한다.**
> >
> > 위 두 줄의 **측정은 그대로 유효하다** — 래치는 재시작에 사라지고 모드는 남는다.
> > 틀린 것은 측정이 아니라 **거기서 끌어낸 결론**이다. *"미전달 행이 그대로인데
> > 진입이 열린다"*를 막는 것은 모드 승격이 아니라 **기동 시 원장 복원**이고,
> > 이 델타는 그것을 이미 별도 요구로 진다 — 「재시작이 진입 차단을 푸는
> > 우회로가 되어서는 안 된다」(4.6).
> >
> > 즉 19판은 **이미 막혀 있는 구멍을 두 번째 수단으로 다시 막고 있었고**,
> > 그 두 번째 수단의 대가가 *"복구는 재시작"*의 거짓화였다.
> > 자세한 것은 바로 위 Scenario의 ✅ 블록.
>
> **design D2의 「`EscalateOperatingMode`에 기대지 않는다」는 유효하다.**
> 그 문장의 근거는 *"승격 경로가 announcer nil로 되돌아오지 않는다"*였고
> **래치를 직접 거는 것**을 요구했다. 6판은 그 요구를 그대로 두고
> *"모드도 승격한다"*만 뺀다.

#### Scenario: 배달 실행자는 exit 사이클을 붙잡지 않는다

- **WHEN** 전송 수단이 응답하지 않는다
- **THEN** exit 관측 사이클의 체류 시간은 **영향을 받지 않는다**

### Requirement: 발송 주체의 부재는 발송 실패와 다른 차단 사유다

배달 실행자의 정지로 걸리는 진입 차단은 **자기 사유 코드를 가져야 한다**(SHALL — 사용자 결정 8-1).
전달 실패로 걸리는 차단과 **같은 코드를 써서는 안 된다**.

**이 요구는 오늘 없던 차단을 하나 만든다.** 오늘 진입을 막는 자리는 전부
*"실제로 보내려다 실패했다"*가 조건이고, 이것은 **아무도 안 보내고 있다**가 조건이다.
시도의 실패가 아니라 **시도할 주체의 부재**이므로 종류가 다르다.
**이 델타는 그 하나만 더하고, 기존 차단·해제 자리는 한 줄도 바꾸지 않는다**(SHALL NOT).

두 사유를 한 코드로 합쳐서는 안 된다(SHALL NOT). 합치면 **운영자가 밀린 알림을 전부
승인하는 순간 「보낼 주체가 없다」는 차단도 함께 풀린다** — 보낼 주체는 여전히 없는데
진입만 열린다. 그것은 이 change의 전제가 그 자리에서 지워지는 것이다.

이 차단을 **자동으로 푸는 경로를 만들어서는 안 된다**(SHALL NOT).
죽은 실행자는 되살아나지 않으므로 복구는 **재시작**이고, 새 프로세스에는
산 실행자가 있다. 이것은 위의 *"재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다"*와
충돌하지 않는다 — 그 요구의 대상은 **원장에 남은 미전달 행**이고 재시작이 그 사실을
안 바꾸는 반면, 이 차단의 대상은 **죽은 프로세스의 실행자**이고 재시작이 그것을 실제로 바꾼다.

#### Scenario: 밀린 알림을 전부 승인해도 발송 주체는 돌아오지 않는다

- **WHEN** 배달 실행자가 죽어 진입이 차단된 상태에서 운영자가 밀린 행을 **전부 승인한다**
- **THEN** 미전달 수는 **0이 된다**
- **AND** 전달 실패로 걸린 차단은 **풀린다**
- **AND** **발송 주체 부재로 걸린 차단은 안 풀린다** — 사유가 다르기 때문이다
- **AND** 그 차단이 **왜 남아 있는지가 운영자에게 보인다**

### Requirement: 배달 실행자는 잠금을 쥔 채 전송하지 않는다

배달 실행자는 **동기 알림 경로와 공유하는 잠금을 원격 전송 위에서 쥐어서는 안 된다**(SHALL NOT).
쥐면 밀린 알림의 개수가 곧 정지 알림의 대기 시간이 되고, **그 개수에 상한이 없다.**

배제는 **원장이 져야 한다**(SHALL). 프로세스 안의 잠금은 그 잠금을 잡는 발송자들만
가르고, 배달 실행자를 잠금 밖으로 빼는 순간 아무것도 안 가른다.

#### Scenario: 밀린 양이 정지를 늦추지 않는다

- **WHEN** outbox에 미전달 critical 알림이 N개 쌓여 있고 배달 실행자가 돌고 있다
- **THEN** 동기 정지 알림 경로의 체류 시간은 **N에 비례하지 않는다**
- **AND** N을 키워도 그 체류 시간은 **늘지 않는다**

#### Scenario: 배달 실행자와 동기 발송 경로가 같은 행을 동시에 보내지 않는다

- **WHEN** 배달 실행자와 동기 알림 경로가 같은 미전달 행을 같은 순간에 집는다
- **THEN** **하나만** 그 행을 전송한다
- **AND** 그 배제는 **둘이 같은 잠금을 공유하지 않아도** 성립한다

### Requirement: strategy projection endpoint의 잔재 회수는 자기 수명주기가 만드는 모든 상태를 다룬다

엔진의 strategy projection endpoint(control 디렉터리·descriptor·socket)의 기동 시 잔재 회수는 그 생성·종료·회수 시퀀스가 만들 수 있는 모든 부분 상태(빈 디렉터리, descriptor만, socket만, 둘 다, 쓰다 만 산출물과 staging 잔재)를 소유자 사망 검증 후 사람 개입 없이 회수해야 하며(SHALL), 산출물 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 소유자 생존 판정은 프로세스 ID 재사용에 오판되지 않는 수단이어야 하며(SHALL) kill-0 단독 판정은 금지되고(SHALL NOT), 최종 이름 socket의 소유자 사망 판정은 관측된 권한 비트로 추정해서는 안 되며(SHALL NOT) 검증한 그 socket에 발행 계약 권한(0600)을 복원한 뒤의 connect probe로 증명해야 하고(SHALL), 그 권한 복원은 회수 경로에서만 일어나야 하며 조회 클라이언트는 endpoint의 권한을 바꿔서는 안 되고(SHALL NOT), 소유권·symlink 검증과 낯선 엔트리의 거부는 유지되어야 한다(SHALL).

#### Scenario: 반쪽 잔재에서의 재기동 (2026-08-13 사고)

- **WHEN** control 디렉터리에 descriptor만 남은 상태(graceful shutdown이 socket을 unlink한
  뒤 프로세스가 죽음)에서 엔진이 기동하면
- **THEN** 잔재를 회수하고 기동을 계속한다 — 어떤 재시도 루프도 같은 상태에 영구히
  막히지 않는다

#### Scenario: 쓰다 만 잔재도 잔재다

- **WHEN** 0바이트·잘린 descriptor, 또는 chmod 전에 죽어 group/other 비트 없는 비-0600
  권한으로 남은 socket이 잔재로 남은 상태에서 엔진이 기동하면
- **THEN** 소유자 사망이 입증되는 한 회수하고 기동을 계속한다

#### Scenario: 재사용된 PID는 주인이 아니다

- **WHEN** 잔재 descriptor의 PID 자리에 무관한 생존 프로세스가 있고 socket은 수락하지
  않으면
- **THEN** 소유자 사망으로 판정하고 회수한다

#### Scenario: 살아 있는 주인은 건드리지 않는다

- **WHEN** 잔재의 socket이 연결을 수락하면
- **THEN** 회수하지 않고 이번 기동 시도를 거부한다

#### Scenario: 쓰기 비트가 깎인 산 socket은 죽은 것이 아니다

- **WHEN** 수락 중인 최종 이름 socket의 권한에서 소유자 쓰기 비트가 외부 chmod로 깎인
  상태에서 엔진이 기동하면
- **THEN** 회수는 권한 비트로 사망을 추정하지 않고 0600 복원 후 probe로 생존을 확인해
  그 socket을 제거하지 않고 기동 시도를 거부한다
- **AND** 그 socket의 group/other 권한은 넓어지지 않는다

#### Scenario: 선임자의 늦은 정리가 후계자를 지우지 않는다

- **WHEN** 종료 중인 선임 프로세스의 지연된 정리와 후계 프로세스의 endpoint 발행이
  겹치면
- **THEN** 선임자는 자신이 발행한 경로만 제거할 수 있고 후계자의 socket은 사라지지
  않는다

### Requirement: 조회 전용 endpoint의 실패는 엔진을 죽이지 않는다

조회 전용 export endpoint(strategy projection 등)의 기동 실패는 엔진 기동을 중단시켜서는 안 되며(SHALL NOT), 엔진은 해당 endpoint 없이 보호·대사 루프를 계속하고 기동 경고와 관측 이벤트로 그 사실을 보고해야 하며(SHALL), 그 보고가 알림 outbox·알림 전달 상태·entry gate에 연결되어서는 안 되고(SHALL NOT — 미전달 행은 다음 부팅의 진입을 잠근다), 강등 기동 후 같은 프로세스 안에서 endpoint 재시도를 해서는 안 되며(SHALL NOT), 엔진 싱글턴 보장은 journal flock이 단독으로 소유한다(SHALL).

#### Scenario: projection 기동 실패에서의 엔진 기동

- **WHEN** strategy projection endpoint 기동이 실패하면
- **THEN** 엔진은 projection 없이 루프를 시작하고 기동 경고·관측 이벤트를 남기며,
  손절·대사 판정 경로는 영향받지 않는다

#### Scenario: 강등 기동은 진입 상태를 바꾸지 않는다

- **WHEN** 알림 전달이 불가능한 배포에서 강등 기동한 엔진을 재기동하면
- **THEN** 강등이 만든 미전달 알림 행은 존재하지 않고 entry gate는 그것 때문에 잠기지
  않는다

#### Scenario: 강등 기동의 싱글턴 불변

- **WHEN** projection 없이 강등 기동한 엔진이 살아 있는 동안 두 번째 엔진이 기동을
  시도하면
- **THEN** journal flock이 두 번째 기동을 거부한다

### Requirement: 알림 채널 식별자는 기계가 만든다

전송 경로의 채널 식별자가 그 자체로 접근 제어인 서비스를 대상으로 할 때, 시스템은 그 식별자를 암호학적 난수로 생성해야 하며 (SHALL), 사람이 고른 이름을 요구해서는 안 된다 (SHALL NOT — 사람이 고른 이름은 계좌 별명·제품명·기본값으로 수렴하고, 그것은 계좌 이벤트를 공개 채널에 놓는 것과 같다).

생성된 식별자는 최소 128비트의 엔트로피를 가져야 한다 (SHALL). 식별자에 붙는 사람이 읽는 접두어는 엔트로피로 계산되어서는 안 된다 (SHALL NOT).

식별자의 생성과 저장은 audit 로그에 시각·주체와 함께 기록되어야 하며 (SHALL), 기록되는 것은 채널이 설정되었다는 사실이지 식별자의 값이 아니다 (SHALL NOT — §0.8, a074가 정한 계약과 같다).

알림을 끄는 것은 식별자를 폐기해서는 안 된다 (SHALL NOT — 폐기하면 다시 켤 때 기존 구독이 무효가 되고, 운영자가 알림을 잠시 끄는 행위가 재구독 비용을 갖게 된다). 꺼진 설정에 남아 있는 식별자가 전송 경로를 다시 구성해서는 안 된다 (SHALL NOT — §0.7).

#### Scenario: 채널 생성
- **WHEN** 채널이 없는 상태에서 알림이 켜진다
- **THEN** 최소 128비트 엔트로피의 식별자가 생성되어 저장되고, audit에는 채널이 설정되었다는 사실만 남는다

#### Scenario: 이미 채널이 있는 상태에서 다시 켠다
- **WHEN** 식별자가 이미 저장된 상태에서 알림이 다시 켜진다
- **THEN** 기존 식별자가 유지되고 새 식별자가 생성되지 않는다

#### Scenario: 끄고 다시 켠다
- **WHEN** 알림을 끈 뒤 다시 켠다
- **THEN** 같은 식별자로 켜지고, 끄기와 켜기 사이에 전송 경로는 구성되지 않는다

### Requirement: 발송 권한은 원장이 준다

원장은 발송을 넘겨줄 때 그 행에 **임차를 표시해야 한다**(SHALL). 임차를 얻지 못한
발송자는 그 행을 **발송하지 않아야 한다**(SHALL NOT).

배제의 근거는 발송자 프로세스 안의 잠금이어서는 안 된다(SHALL NOT). 잠금은 그것을
잡는 한 프로세스 안의 발송자들만 가른다. **원장이 배제를 져야** 발송자가 둘이 되어도
같은 요구가 유지된다.

전달 성공을 기록하는 CAS는 이 요구를 만족하지 않는다. 그 CAS는 **네트워크 발송이 끝난
뒤에** 돌기 때문에 이중 **정산**만 막고 이중 **발송**은 막지 못한다.

**이 요구의 경계를 정직하게 적는다.** 임차는 **이미 나간 원격 발송을 취소하지 못한다.**
발송자가 정지·긴 GC·VM 일시정지로 임차를 넘겨 잃으면 그 사실은 **정산이 거부되는
시점에야** 드러나고, 그때 발송은 이미 끝나 있다. 위 CAS를 반증한 논리가 임차에도 걸린다.

| 조건 | 규범 |
|---|---|
| **유계 실행** 아래의 경합 | 정확히 하나만 발송한다 (SHALL) |
| 임차를 넘겨 잃은 발송자의 **정산** | 반드시 거부되어야 한다 (SHALL) |
| 그 밖의 경우 | 같은 알림이 **두 번 나갈 수 있다** — 이 요구는 그것을 막지 않는다 |

**이중 정산은 절대 허용하지 않고, 이중 발송은 유계 실행 밖에서 허용한다.**
둘을 한 문장에 합치면 이 저장소가 지킬 수 없는 규범이 된다.

**그리고 이 요구는 전달 성공을 보장하지 않는다.** 전송 수단이 계속 죽어 있거나
발송 주체가 없으면 전달 횟수는 **0**이다. 이 요구가 정하는 것은 **누가 보낼 수 있고
누가 정산할 수 있는가**이지 **몇 번 도달하는가**가 아니다. 도달을 지는 것은 배달
주체와 진입 차단이고, 그것은 다른 요구다.

#### Scenario: 두 발송자가 같은 행을 동시에 집는다

- **WHEN** 유계 실행 아래에서 발송자 둘이 같은 미전달 critical 알림 행을 같은 순간에 claim한다
- **THEN** 정확히 **하나만** 발송 권한을 얻는다
- **AND** 다른 하나는 그 행을 **발송하지 않는다**
- **AND** 그 배제는 두 발송자가 **같은 잠금을 공유하지 않아도** 성립한다

#### Scenario: 임차를 넘겨 잃은 발송자가 뒤늦게 깨어난다

- **WHEN** 발송자가 임차를 든 채 멈춘 사이 임차가 만료되고 다른 발송자가 같은 행을 집어
  발송한 뒤, 먼저 있던 발송자가 깨어나 자기 발송을 실행한다
- **THEN** 그 알림은 **두 번 나갈 수 있다** — 이 시스템은 그것을 침묵보다 낫다고 본다
- **AND** 먼저 있던 발송자의 **정산은 거부된다**
- **AND** 그 거부는 **구조화 로그로 남는다**

#### Scenario: claim을 거치지 않는 발송 경로가 없다

- **WHEN** 어떤 코드 경로가 outbox 행을 전송 수단으로 내보낸다
- **THEN** 그 경로는 **그 행의 임차를 먼저 얻어야 한다**

### Requirement: 발송 중인 행은 여전히 미전달이다

미전달 수와 밀린 목록은 **발송 중인 행을 계속 포함해야 한다**(SHALL).
발송이 진행 중이라는 사실은 그 행이 **전달되었다는 뜻이 아니다**.

임차 표시는 알림의 **상태**가 아니다(SHALL NOT). 상태로 표시하면 발송 중인 행이 미전달
집계에서 빠지고, **미전달 critical 알림이 남아 있는데 신규 진입 차단이 스스로 풀린다.**

#### Scenario: 발송 중에 진입 차단이 풀리지 않는다

- **WHEN** 미전달 critical 알림 하나가 발송자에게 claim되어 발송 중이다
- **THEN** 미전달 수는 **그 행을 계속 센다**
- **AND** 운영자가 승인해도 **미전달 수가 0이 아니면 진입 게이트는 안 풀린다**
- **AND** 운영자의 밀린 목록에 그 행이 **계속 보인다**

### Requirement: 임차 경합은 「이미 전달됨」과 같은 결과가 아니다

발송 권한을 못 얻은 것과 알림이 이미 전달된 것은 **서로 다른 결과여야 한다**(SHALL).
둘을 한 값으로 합치면, 임차를 든 발송자가 죽었을 때 **미전달 critical 알림이 조용히
억제되고 진입이 열린다** — 억제의 근거가 「운영자가 이미 받았다」로 잘못 읽히기 때문이다.

**임차 경합을 관측했다는 사실만으로 진입 게이트를 움직여서는 안 된다**(SHALL NOT).
잠그는 쪽도, 푸는 쪽도 같다. 정상 발송 중에도 경합은 일어나고, 그렇게 걸린 잠금을
**성공한 발송은 풀 수 없다**. 원장은 「살아서 보내는 중」과 「죽었다」를 구분해 주지
않으므로 **경합자가 추측하면 안 된다.**

**진입 차단이 걸리고 풀리는 근거는 이 델타가 바꾸지 않는다**(SHALL NOT). 발송 시도의
실패가 잠그고 운영자의 승인이 푼다 — `openspec/specs/engine-safety/spec.md`의 정본
그대로다. 임차는 **누가 보내는가**를 정하고, **언제 진입을 막는가**는 안 정한다.

> 3판은 이 자리에 *"미전달 수가 0이 되면 사람 개입 없이 풀려야 한다"*를 적었다.
> **승인된 정본보다 덜 보수적이고 `MODIFIED` 표시도 없었다** — 3라운드 A-P1.
> 사용자 결정 5-1로 되돌렸다.

#### Scenario: 임차를 든 발송자가 죽고 다른 관측이 들어온다

- **WHEN** 발송자가 claim한 뒤 전송하지 못하고 사라지고, 그 임차가 살아 있는 동안
  같은 조건이 다시 관측된다
- **THEN** 그 관측은 **「이미 전달됨」으로 다뤄지지 않는다**
- **AND** 그 행은 **미전달로 계속 세어진다** — 어떤 읽기 경로에서도 안 사라진다
- **AND** 그 관측은 진입 게이트를 **잠그지도 풀지도 않는다**
- **AND** 그 사실이 **구조화 로그로 남는다**

> 3판은 위 두 번째 AND를 *"신규 진입은 차단된 상태다"*라고 적었다. **거짓이다** —
> 아무도 아직 시도하지 않은 행은 오늘도 진입을 안 막고, 결정 5-1이 그것을 안 바꾼다.
> 4라운드가 잡았다. 그 행이 진입을 막게 되는 것은 **누군가 실제로 보내려다 실패할 때**
> 이고, 그 자리는 `deliver`의 오늘 코드다.

#### Scenario: 정상 발송 중의 경합이 진입을 영구히 잠그지 않는다

- **WHEN** 한 발송자가 유효한 임차로 발송 중이고, 다른 관측이 같은 행을 claim하려다 실패한다
- **THEN** 그 관측은 발송을 **건너뛴다**
- **AND** 그 관측은 진입 게이트를 **잠그지도 풀지도 않는다**
- **AND** 원래 발송자가 성공적으로 정산하면 그 행은 미전달 수에서 **빠진다**

### Requirement: 임차의 움직임은 관측 가능해야 한다

임차 경합·만료 탈취·소유자 불일치는 **각각 구조화 로그로 남아야 한다**(SHALL).
전송 실패와 **같은 이벤트로 합쳐서는 안 된다** — 원인이 다르고 운영자가 할 일이 다르다.

운영자가 읽는 밀린 목록은 각 행의 **임차 상태를 보여야 한다**(SHALL).
PENDING인데 아무도 안 보내고 있는 행과, 지금 누군가 보내는 중인 행은 다른 상황이다.

#### Scenario: 임차를 잃은 발송자의 기록 시도

- **WHEN** 임차가 만료되어 다른 발송자가 집은 뒤, 먼저 있던 발송자가 결과를 기록한다
- **THEN** 그 거부는 **전송 실패와 다른 이벤트로 남는다**
- **AND** 새 보유자의 임차는 **그대로 있는다**
- **AND** 먼저 있던 발송자는 **남은 전송을 중단한다**

#### Scenario: 한 번의 시도가 실패해도 임차는 유지된다

- **WHEN** claim한 발송자의 전송 시도 하나가 실패하고 그 발송자에게 재시도 예산이 남아 있다
- **THEN** 그 행은 **PENDING으로 남는다**
- **AND** 임차는 **그 발송자에게 그대로 있는다** — 재시도 사이에 다른 발송자가 들어오면
  같은 알림이 두 번 나간다

#### Scenario: 발송자가 예산을 다 쓰면 임차를 놓는다

- **WHEN** claim한 발송자가 재시도 예산을 전부 소진하고 포기한다
- **THEN** 그 행은 **PENDING으로 남는다**
- **AND** 임차가 풀려 **다음 발송자가 집을 수 있다**

#### Scenario: 보냈는지 원장이 모르면 임차를 놓지 않는다

- **WHEN** 전송은 성공했는데 그것을 원장에 기록하지 못한다
- **THEN** 임차는 **풀리지 않는다** — 만료가 풀 때까지 억제 표시로 남는다
- **AND** 신규 진입은 **차단된 상태로 있는다**

> 이 세 Scenario는 1판에서 하나였고 *"실패하면 임차가 풀린다"*라고 적었다.
> **1라운드 A-P3 = B-P7이 그것을 규범과 구현 계획의 모순으로 잡았다.**
> 셋째는 1라운드 B-P3이 더한 것이다 — 그 자리에서 풀면 다음 관측이 다시 보내
> **2026-08-08의 폭풍(한 행에 예순 번)이 돌아온다.**

### Requirement: 죽은 발송자가 알림을 가두지 못한다

임차는 **만료되어야 한다**(SHALL). 발송자가 claim한 뒤 죽으면 그 행은 만료 뒤에 다시
claim 가능해져야 한다.

만료의 실패 방향은 **재발송이다**(SHALL). 운영자가 같은 경보를 두 번 보는 것과 한 번도
못 보는 것 중, 후자가 이 시스템이 존재하는 이유의 실패다.

#### Scenario: 발송자가 claim한 채 죽는다

- **WHEN** 발송자가 행을 claim한 뒤 발송을 끝내지 못하고 사라진다
- **THEN** 만료 뒤에 다른 발송자가 그 행을 **claim할 수 있다**
- **AND** 그 사이에도 그 행은 **미전달로 세어진다**
- **AND** 그 행은 어떤 읽기 경로에서도 **사라지지 않는다**

#### Scenario: 임차를 잃은 발송자가 새 보유자를 밀어내지 못한다

- **WHEN** 임차가 만료되어 다른 발송자가 같은 행을 claim한 뒤,
  먼저 있던 발송자가 자기 시도의 결과를 기록한다
- **THEN** 새 보유자의 임차는 **그대로 있는다**

### Requirement: 배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다

critical 알림의 전달 실패가 지속되면 신규 진입을 차단하고 운영 모드를 승격하는 주체는 배달 실행자여야 한다(SHALL).
그 판정은 원장에 남는 행의 시도 수에 서야 하며(SHALL), 프로세스 메모리의 사이클 계수에만 서서는 안 된다
(SHALL NOT — 재시작이 계수를 지운다).

차단 사유는 정본의 전달 실패 사유(`critical_alert_undelivered`)여야 하고(SHALL), 승격 트리거는 정본의
`CRITICAL_ALERT_UNDELIVERED` 여야 한다(SHALL). 새 사유·새 트리거를 만들어서는 안 된다(SHALL NOT). 그래서 해제는
정본 「운영자가 밀린 알림을 해제한다」의 두 조건 그대로이고, 모드 완화는 사람 승인 그대로다.

이 요구와 그 시나리오에서 「운영 모드 승격」은 **원장에 남는 모드 전이**를 말한다. 그 모드를 진입 게이트에 투영해 신규 진입을 거절하게 하는 것과
기동 때 그 투영을 복원하는 것은 이 요구가 보장하지 않는다. 이 요구의 신규 진입 차단은 전달 실패 사유의 게이트 래치로 성립해야 하며(SHALL),
운영 모드의 진입 집행에 기대서는 안 된다(SHALL NOT — 모드 투영이 배선되지 않은 빌드에서도 이 요구는 참이어야 한다). 모드 투영이 배선되기 전에는
원장의 모드 행이 진입 집행을 더하지 않으며, 그 사실은 change 에 적혀야 한다(SHALL). 이 요구가 허용하거나 요구하는 재잠금(아래 「늦은 적용」의 예외)은
이 문장과 무관하게 그대로다.

동기 알림 경로가 전송을 시도하지 않는 구성에서도 이 요구는 성립해야 한다(SHALL) — 동기 경로의 시도에 의존하는
차단은 그 경로가 루프 밖으로 옮겨지는 순간 사라진다. 전송 수단이 설정되지 않은 것은 응답하지 않는 것과 같은
실패 시도로 세어야 한다(SHALL).

판정이 읽는 시도 수는 그 시도를 기록한 원장 쓰기가 **커밋한 값**이어야 한다(SHALL) — 기록 전에 읽어 둔 목록의 값은 그 사이
다른 발송자나 재무장이 바꿀 수 있다. 시도 기록이 적용되지 않은 결과(행이 이미 정산됨 · 남의 임차)에서는 잠그거나 승격해서는
안 된다(SHALL NOT).

**늦은 적용은 제때 적용과 같아야 한다(SHALL).** 배달 실행자가 판정을 늦게 적용한 결과(진입 게이트와 운영 모드)는, 그 판정을 근거가 확정된 순간에
원자적으로 적용했을 때의 결과와 같아야 한다(SHALL) — 이것은 「운영자의 승인은 전송의 성공 여부로 되살아나지 않는다」를 실행자의 지연에 옮긴 것이다.
제때 적용은 **차단과 승격을 함께** 하고, 그 뒤의 운영자 해제는 **차단만** 지운다(운영 모드는 사람의 완화 승인으로만 풀린다). 그러므로:

- 근거가 확정된 **뒤**에 전달 실패 사유의 해제가 있었으면, 차단은 적용하지 않아야 하고(SHALL — 제때 선 차단도 그 해제가 지웠다) 그 판정에 승격이
  포함되면 승격은 적용해야 한다(SHALL — 제때 된 승격은 그 해제로 풀리지 않는다). 승격을 포함하지 않는 판정(시도 기록이 행을 찾지 못함 · 모르는
  결과)은 승격을 만들어서는 안 된다(SHALL NOT).
- 승격이 포함된 판정에서 승격 쓰기가 실패하면, 조건부 차단의 결과와 무관하게 차단을 적용해야 한다(SHALL — 조건부 차단이 섰어도 그 뒤 해제가
  지웠을 수 있다; 둘 다 잃지 않는다, 보수 방향).
- 근거가 확정되기 **앞**의 해제는 판정을 바꿔서는 안 된다(SHALL NOT — 아직 없던 차단을 지울 수 없었다).
- 이 구별은 행의 원장 상태가 아니라 **해제와 근거의 순서**로 해야 한다(SHALL — 전달된 행은 승인 상태가 되지 않으므로 원장 상태로는 「전달 뒤 수동
  해제」를 알아볼 수 없다). 원장의 승인 시각으로 순서를 추정해서는 안 된다(SHALL NOT — 시각은 승인이 끝난 순간이 아니고 벽시계는 되감긴다).

근거가 확정된 순간은: 시도 기록이 적용된 판정이면 그 커밋이 돌아온 순간, 원장 오류로 끝난 판정이면 그 오류가 돌아온 순간, 연속 기록 실패 판정이면
한도째 오류가 돌아온 순간이다. 연속 기록 실패에서 이번 오류가 한도째이면 한도 판정이 해제 세대에 따른 리셋보다 **먼저**다(SHALL — 그 해제가 한도째 오류 뒤였을 수 있다): 승격을
적용하고, 차단은 직전 증가 이후 해제가 없을 때만 적용한다(SHALL). 한도 아래의 오류에서만 해제 세대 변화가 연속을 끊는다(SHALL). 확정 순간과 해제 세대를 읽는 순간 사이의 해제를 「앞」으로 보는 것은 허용된다(SHALL — 차단이 더 서는 쪽으로만
틀린다). 원장 오류로 끝난 판정은 그 오류가 운영자의 승인을 가렸을 수 있어도 원장 상태로 버려서는 안 된다(SHALL NOT — 보수 방향; 남는 차단은 빈 목록
승인이 푼다). 운영자의 해제로 이어지지 않은 승인(다른 행만 승인해 미전달이 남음, 또는 도중에 실패한 승인)은 판정을 바꿔서는 안 된다(SHALL NOT).

**배달 실행자는 손절 경로의 동기 알림·비상 청산이 기다리는 잠금 안에서 다른 무엇도 기다려서는 안 된다(SHALL NOT — 그 대기는 취소되지 않는다).**
배달 실행자가 잡는 잠금은 진입 게이트 잠금뿐이어야 하며(SHALL), 그 구간에서는 게이트 상태를 읽고 쓰는 일만 해야 하고 원격 전송·원장 트랜잭션·
로그·다른 잠금을 해서는 안 된다(SHALL NOT). 그래서 손절 경로가 배달 실행자 때문에 기다리는 시간은 그 게이트 구간 하나로 경계가 있다. 승격이
게이트 잠금 밖이므로 사람이 모드를 완화한 직후 이미 적용된 판정의 승격이 뒤따를 수 있고, 그 결과(모드가 다시 막힘)는 허용된다(SHALL — 보수 방향).

**시도를 원장에 기록하지 못하는 것도 지속 실패다(SHALL).** 한 행의 시도 기록이나 임차가 원장 오류로 끝나는 일이 그 행에서 **연속으로**
재시도 한도만큼 일어나면, 또는 미전달 나열이 연속으로 그만큼 실패하면 신규 진입을 차단해야 한다(SHALL) — 기록되지 않는 시도는 시도 수를
올리지 못하므로, 이 규칙이 없으면 원장 결함이 차단의 면제가 된다. 한 행의 연속은 그 행에 대한 정산 쓰기가 **적용될 때만** 끊긴다(SHALL —
다른 행의 성공이나 아무것도 쓰지 않은 결과는 그 행의 기록 능력의 증거가 아니다). 전달 실패 사유의 해제 세대가 바뀌면 그 전의 연속은 끊긴다
(SHALL — 운영자가 backlog 를 비운 신호다; 그 신호가 틀려도 결과는 다시 세는 지연뿐이다) — 단 이번 오류가 한도째이면 위의 한도 판정이 먼저다. 해제로 이어지지 않은 승인은 연속을 끊어서는 안 된다(SHALL NOT).
엔진 종료로 인한 취소는 세지 않는다(SHALL NOT). 그 밖에 한 행의 연속이 끊기는 것은 그 행이 PENDING 을 떠났음이 관측될 때뿐이다(SHALL) — 임차가
「이미 정산됨」을 돌려주거나, 배치로 잘리지 않은(완전한) 나열에 그 행이 없을 때. 잘린 나열에 없다는 것은 그 증거가 아니다(SHALL NOT). 임차 반납의
성공은 정산 쓰기가 아니다(SHALL NOT — 연속을 끊지 않는다). 나열 실패의 연속은 나열이 성공하면 끊긴다(SHALL).

**발행은 됐는데 전달 기록이 실패하면 즉시 신규 진입을 차단하고 운영 모드를 승격해야 하며(SHALL), 배달 실행자는 그 행의 임차를 스스로
놓아서는 안 된다(SHALL NOT — 놓으면 이미 나간 알림이 곧장 다시 나간다).** 이 억제는 임차가 살아 있고 교체되지 않은 동안만이다 — 만료 뒤의
재발행은 허용되며 그 사실은 change 에 적혀야 한다(SHALL). 남이 먼저 정산한 것(승인·남의 임차)은 이 사건이 아니다(SHALL NOT). 이 판정도
위의 「늦은 적용은 제때 적용과 같아야 한다」를 따른다(SHALL).

차단 사유의 설명에는 행의 제목·본문·payload·원문 오류·계좌·임차 토큰을 넣어서는 안 된다(SHALL NOT — 게이트 상태는 그것을
읽는 모든 곳에 노출된다). 이 요구가 새로 만드는 로그 줄에는 행의 제목·본문·payload·임차 토큰·계좌 참조·원문 오류 문자열을 넣어서는 안 된다
(SHALL NOT).

#### Scenario: 전송이 재시도 한도까지 실패한다
- **WHEN** 배달 실행자가 어떤 행의 시도 수를 재시도 한도까지 올렸는데 전달되지 않았다
- **THEN** 신규 진입이 차단되고 운영 모드가 ENTRY_BLOCKED 로 승격되며 그 행은 미전달로 보존된다
- **AND** 동기 알림 경로는 그 사이 한 번도 전송을 시도하지 않았다

#### Scenario: 운영자가 승인하면 차단이 풀리고 모드는 남는다
- **WHEN** 운영자가 밀린 행을 전부 승인해 미전달 수가 0 이 된다
- **THEN** 전달 실패 사유의 진입 차단은 풀린다
- **AND** 원장의 운영 모드는 사람 승인 전까지 ENTRY_BLOCKED 로 남는다

#### Scenario: 발송 중에 운영자가 승인한다
- **WHEN** 배달 실행자가 한도 직전의 행을 전송하는 동안 운영자가 그 행을 승인하고, 그 뒤 전송이 실패한다
- **THEN** 그 실패는 진입을 차단하지도 운영 모드를 승격하지도 않는다

#### Scenario: 한도 도달 직후 운영자가 승인한다
- **WHEN** 배달 실행자가 한도에 이른 실패를 기록하고 차단한 직후 운영자가 밀린 행을 전부 승인한다
- **THEN** 승인은 차단을 풀고, 그 뒤 늦은 처리가 차단을 다시 거는 것은 허용된 보수 예외(승격 쓰기 실패 뒤의 무조건 차단)뿐이다

#### Scenario: 시도 기록이 계속 실패한다
- **WHEN** 전송은 실패하고 그 시도를 원장에 기록하는 쓰기가 재시도 한도만큼 연속으로 실패하며 미전달 행이 남아 있다
- **THEN** 신규 진입이 차단된다
- **AND** 엔진 종료로 취소된 기록은 그 수에 들어가지 않는다

#### Scenario: 래치가 서기 전에 운영자가 backlog 를 비운다
- **WHEN** 배달 실행자가 한도에 이른 실패를 기록한 직후, 차단을 적용하기 전에 운영자가 밀린 행을 전부 승인한다
- **THEN** 차단은 걸리지 않는다 — 빈 backlog 위에 다시 잠기지 않는다(해제가 실행자의 해제 세대 읽기 **뒤**일 때; 기록 확정과 세대 읽기 사이의 해제는
  보수적으로 다시 잠글 수 있다)
- **AND** 원장의 운영 모드는 제때 적용했을 때처럼 승격되어 있다(승인은 모드를 풀지 않는다)

#### Scenario: 근거 확정 앞의 해제는 판정을 버리게 하지 않는다
- **WHEN** 운영자의 해제가 있은 뒤에 새 행이 한도에 이르는 실패를 기록한다
- **THEN** 차단이 걸리고 운영 모드가 승격된다

#### Scenario: 근거 확정 뒤의 해제는 제때 적용한 것과 같은 결과를 남긴다
- **WHEN** 행이 한도에 이르는 실패를 기록한 뒤, 차단이 적용되기 전에 그 행이 다른 경로로 전달되고 운영자가 해제한다
- **THEN** 차단은 걸리지 않고 원장의 운영 모드는 승격된다
- **AND** 게이트와 원장의 운영 모드는 판정을 실패 기록 순간에 적용하고 그 뒤 해제가 차단을 지운 경우와 같다 — 해제가 실행자의 해제 세대 읽기 뒤일 때;
  그 앞이면 차단이 보수적으로 남을 수 있다

#### Scenario: 오류로 끝난 판정은 원장의 승인 상태로 버리지 않는다
- **WHEN** 배달 실행자가 행을 임차한 뒤 운영자가 그 행만 승인하고(다른 행이 남아 해제 없음), 실행자의 전달 기록이 원장 오류로 끝난다
- **THEN** 차단이 걸리고 운영 모드가 승격된다 — 원장의 승인 시각으로 순서를 추정하지 않는다

#### Scenario: 다른 행만 승인해도 판정은 선다
- **WHEN** 배달 실행자가 한 행의 한도 도달을 판정하는 사이 운영자가 다른 행만 승인해 미전달이 남는다
- **THEN** 차단이 걸리고 운영 모드가 승격된다

#### Scenario: 기록 실패가 이어지는 동안 운영자가 승인한다
- **WHEN** 한 행의 시도 기록이 연속으로 실패하는 사이 운영자가 밀린 행을 승인하고, 그 뒤 새 critical 행이 기록되며, 옛 행의 기록 실패가 한 번 더 난다
- **THEN** 옛 행의 실패는 차단을 다시 걸지 않는다 — 새 행은 자기 시도로 판정된다

#### Scenario: 발행은 됐는데 전달 기록이 실패한다
- **WHEN** 배달 실행자가 행을 발행했고 그 전달을 원장에 기록하는 쓰기가 실패한다
- **THEN** 신규 진입이 즉시 차단되고 운영 모드가 승격되며, 배달 실행자는 그 행의 임차를 놓지 않아 임차가 살아 있는 동안 같은 알림이 다시 나가지 않는다

#### Scenario: 재시작은 판정을 지우지 않는다
- **WHEN** 한도에 이른 미전달 행이 있는 채로 엔진이 재시작한다
- **THEN** 기동 복원이 미전달 수로 다시 차단하고 운영 모드는 원장에 남아 있다

### Requirement: 진입 게이트는 사유별 해제 세대를 가진다

진입 게이트는 사유마다 **해제 세대**를 가져야 한다(SHALL). 그 사유의 해제 요청이 올 때마다 세대는 정확히 1 증가해야 하며(SHALL — 그 사유의 래치가
있었는지와 무관하게), 세대는 줄어들어서는 안 된다(SHALL NOT — 단조 증가). 그 사유의 해제 요청 없이는 세대가 바뀌어서는 안 된다(SHALL NOT) —
다른 사유의 해제·어떤 사유의 잠금·종목 단위 잠금과 해제·운영 모드 투영·대사 투영 재구성은 이 세대를 바꾸지 않는다.

「세대가 그대로일 때만 잠근다」는 연산은 게이트 잠금 하나 안에서 세대 비교와 잠금을 함께 해야 하며(SHALL — 둘 사이에 해제가 끼면 안 된다),
세대가 다르면 아무것도 바꾸지 않아야 하고(SHALL), 같으면 일반 잠금과 같은 규칙(없을 때만 삽입, 처음 설명 유지)을 따라야 한다(SHALL).

해제 세대의 도입이 기존 게이트 상태 세대(전략 진입 봉인에 쓰는 것)의 의미를 바꿔서는 안 된다(SHALL NOT) — 그 세대는 지금처럼 실제로 상태가
바뀐 때만 오른다.

전달 실패 사유를 해제하는 호출자는 운영자의 승인 경로뿐이어야 하며(SHALL), 그 경로는 미전달 0 을 확인한 뒤에만 해제를 요청한다(SHALL — 그래서 해제
요청 자체가 사람 승인의 표식이고, 그 순간 래치가 있었는지는 우연이다). 새 호출자를 더하는 편집은 세대의 의미가 그대로인지 다시 증명해야 한다(SHALL).

#### Scenario: 무관한 해제는 세대를 바꾸지 않는다
- **WHEN** 다른 사유가 해제되거나, 어떤 사유가 잠기거나, 운영 모드가 투영된다
- **THEN** 전달 실패 사유의 해제 세대는 그대로다

#### Scenario: 래치가 없어도 해제 요청은 세대를 올린다
- **WHEN** 전달 실패 사유의 래치가 없는 게이트에 그 사유의 해제 요청이 온다
- **THEN** 그 사유의 해제 세대는 1 오르고, 게이트 상태 세대는 그대로다

#### Scenario: 세대가 바뀐 뒤의 조건부 잠금은 아무것도 하지 않는다
- **WHEN** 배달 실행자가 세대를 읽은 뒤 해제 요청이 있었고, 그 뒤 실행자가 그 세대로 조건부 잠금을 요청한다
- **THEN** 게이트는 잠기지 않고 요청은 거짓을 돌려준다

### Requirement: 배달 실행자의 행 선택은 굶주림을 만들지 않는다

한 사이클의 선택은 아직 재시도 한도에 이르지 않은 행을 먼저 골라야 하며(SHALL), 한도에 이른 행은 그 사이클의
잔여 자리에만 들어간다(SHALL). 한도에 이른 행을 버려서는 안 된다(SHALL NOT — 내구성 완화이고, 미전달의 결과인
진입 차단도 함께 사라진다). 빠지는 것은 우선순위뿐이다.

진입 게이트 래치가 걸리기까지의 최악 시간은 change 에 적혀야 한다(SHALL). 그 값은 첫 시도 전의 큐 대기를
포함해야 하며(SHALL), 큐를 뺀 값을 상한이라고 불러서는 안 된다(SHALL NOT).

#### Scenario: 한도 행 열이 배치를 채우고 있다
- **WHEN** 시도 한도에 이른 미전달 행이 배치 상한 이상 쌓여 있고 새 critical 행이 기록된다
- **THEN** 다음 사이클은 새 행을 먼저 시도한다

#### Scenario: 한도에 이른 행도 사라지지 않는다
- **WHEN** 한도에 이른 행이 잔여 자리로 밀린다
- **THEN** 그 행은 미전달로 남고 미전달 수에 계속 세어진다

### Requirement: 운영 모드는 산 프로세스의 진입 게이트에 닿는다

원장의 운영 모드 전이는 생산 엔진의 진입 게이트에 투영되어야 한다(SHALL — 정본 risk-management 「모드의 강제 지점은 EntryGate 투영이다」의 요구를 생산 조립에서 성립시키는 문장이다. 모드×클래스 표와 트리거 열거는 그 정본이 소유하며 이 요구는 다시 적지 않는다). 엔진 기동은 원장의 현재 모드를 **첫 진입 점검보다 먼저** 게이트에 복원해야 한다(SHALL — 복원이 진입 허용보다 늦으면 그 사이의 진입은 모드를 보지 못한다).

> **21판 범위 — 사용자 결정(Q3 확정, 2026-09-28).** *"투영기 배선(SetModeProjector 생산 배선 + 기동 RestoreOperatingModeProjection +
> AC2 수리)은 21판 범위 포함, 모드 완화 경로는 승인 원칙(자동은 조이기만 / 완화는 OPERATOR + 승인 참조 + commit 전 audit / journal API +
> tossctl mutating 명령 / 콘솔 없음)으로 설계한다."* 설계는 design D0.3f.

모드 사유의 교체는 게이트 잠금 **한 번 안에서** 원자적이어야 한다(SHALL — 지운 뒤 잠금을 놓고 다시 세우는 창에서는 진입 점검이 모드 사유를 보지 못한다). 게이트에 적용된 모드는 **커밋 순서에서 역행해서는 안 된다**(SHALL NOT — 커밋 뒤 잠금 없이 부르는 투영은 늦게 커밋된 전이의 투영이 먼저 도착할 수 있다. 늦게 도착한 옛 전이의 투영은 적용하지 않는다). 전이 호출이 반환한 뒤의 진입 점검은 그 전이 이상으로 새로운 모드를 보아야 한다(SHALL). 커밋과 투영 사이의 짧은 창의 진입 점검은 직전 모드를 볼 수 있다 — 이 요구의 보장 시점은 전이 호출의 반환이며(SHALL), 그 창을 닫으려고 게이트 잠금 안에 원장 입출력을 넣어서는 안 된다(SHALL NOT). 정본 risk-management 「모드 전환은 journal 영속과 **동시에** EntryGate 계좌 latch로 투영」의 「동시에」는 **같은 전이 호출 안에서 커밋 뒤 투영까지**로 읽는다(SHALL — 영수증은 design D0.3h 2: 전이 함수가 커밋한 뒤 같은 호출 안에서 묶인 투영기를 부른다). 구현이 커밋과 투영을 다른 호출·다른 실행 흐름으로 가르면 이 해석을 위반한다(SHALL NOT).

원장의 「현재 모드」는 **커밋 순서** 하나로 정해져야 한다(SHALL — 벽시계로 정하면 시각을 먼저 얻고 늦게 커밋된 전이가 「과거」가 되어, 산 게이트와 다른 모드를 재시작 복원이 세울 수 있다. 그것은 fail-open 방향이다). 전이의 방향 판정, 기동 복원, 게이트의 투영 적용 순서, 그리고 운영자가 읽는 전이 이력의 순서는 같은 순서 하나를 써야 한다(SHALL).

모드 투영을 생산에 배선하는 빌드는 **사람이 모드를 완화하는 생산 경로**를 함께 가져야 한다(SHALL — 완화 경로 없이 투영하면 모든 자동 강화가 원장을 직접 고치기 전까지 풀리지 않는 진입 차단이 된다). 그 경로는 다음을 지켜야 한다:

- 자동 전이는 조이기만 한다(SHALL — 자동 완화는 거절된다). 완화는 **OPERATOR**가 **승인 참조**를 달고 하며(SHALL), 그 audit 기록은 전이의 **commit 전에** 쓰이고 audit 쓰기가 실패하면 전이는 일어나지 않는다(SHALL — 기록 없는 완화가 이 경로가 막으려는 상태다).
- 완화는 원장 API(`TransitionOperatingMode`)를 거치는 **tossctl 명령**이어야 하고(SHALL), 그 명령은 `mutating: true`로 표지되어야 한다(SHALL — 계좌의 진입 허용을 바꾼다). 대화형 에이전트는 그것을 자동 실행하지 않는다(SHALL NOT). 콘솔 표면은 두지 않는다(SHALL NOT).
- 완화는 **엔진 프로세스 안에서** 일어나야 한다(SHALL — 모드 사유 래치는 엔진 프로세스의 게이트에 살고, 원장만 고치면 산 게이트는 안 풀린다. 정본 「운영자가 밀린 알림을 읽고 승인하는 경로」와 같은 이유다).
- 완화 명령은 운영자 이름·승인 참조·사유를 전부 요구하고 기본값을 두지 않는다(SHALL). 타이핑 확인이나 추가 승인 마찰은 요구하지 않는다(SHALL NOT).
- 완화는 모드 사유의 차단만 푼다(SHALL — 다른 사유의 차단은 그대로이며, 명령은 남은 사유를 보여야 한다).
- 완화 명령은 전이 뒤의 모드와 남은 진입 차단 사유를 다시 읽어 보여야 한다(SHALL — 같은 호출 안의 다른 경로가 방금 푼 모드를 다시 조일 수 있다).
- 완화도 운영자에게 통지되어야 한다(SHALL — 누군가 진입을 다시 연 것은 알려야 하는 사건이다). 원장이 커밋한 뒤 통지만 실패하면 명령은 그 완화를 **성공**으로 보고하고 통지 실패를 함께 보여야 한다(SHALL — 커밋된 완화를 실패로 보고하면 운영자가 다시 시도한다). 여기서 「통지 실패」는 **통지의 기록 실패**다 — 전송은 배달 실행자의 일이라 명령의 호출 안에서 드러나지 않는다. 명령은 대신 다시 읽은 통지 행의 상태(전달 전인지)를 보인다(SHALL).

**통지자(announcer)를 받는 운영 모드 전이의 통지는 상태를 바꾼 전이마다 한 번 나가야 한다(SHALL)** — 사람의 완화, 관측 두절·자격증명 거절 같은 자동 강화가 여기에 든다. **전달 실패에 따른 강화**(`CRITICAL_ALERT_UNDELIVERED` — 동기 발송 경로와 배달 실행자 둘 다)와 위의 durable 기록 실패에 따른 강화는 **통지하지 않는다**(SHALL NOT — 방금 전송이나 기록에 실패한 수단으로 그 사실을 다시 알리지 않는다. 그 사실은 진입 차단 사유와 구조화 로그가 남긴다). 통지의 중복 제거 신원은 **전이 하나**(전이 행의 고유 신원)여야 하며, 계정과 목적 모드만으로 두어서는 안 된다(SHALL NOT — 완화 경로가 생기면 「강화 → 완화 → 재알림 창 안의 재강화」가 도달 가능해지고, 그 재강화·재완화의 통지가 옛 정착 행에 흡수되어 나가지 않는다). 통지는 전이가 상태를 바꾼 때만 일어나므로 이 신원이 통지 수를 늘리지 않는다.

기존에 강화된 모드 행의 처분은 **배포하는 사람의 결정**이다(SHALL — 배선이 착지한 첫 기동은 원장의 현재 모드를 집행한다. 운영 원장에 강화된 행이 있으면 그 기동에서 신규 진입이 막히며, 그것을 풀지 여부는 사람이 판단하고 사람이 실행한다). 배포는 운영 원장의 현재 모드를 사람이 먼저 확인한 뒤에만 한다(SHALL — 의도된 효과이지만 예고 없이 일어나서는 안 된다).

#### Scenario: 자동 강화가 산 프로세스의 진입을 막는다
- **WHEN** 엔진이 도는 중에 자동 트리거가 운영 모드를 ENTRY_BLOCKED로 강화한다
- **THEN** 그 전이 호출이 반환한 뒤의 진입 점검은 모드 사유로 신규 진입을 거절한다
- **AND** 청산은 영향을 받지 않는다

#### Scenario: 기동 복원이 진입 허용보다 먼저다
- **WHEN** 원장의 운영 모드가 ENTRY_BLOCKED인 채로 엔진이 기동한다
- **THEN** 첫 진입 점검은 이미 모드 사유를 본다

#### Scenario: 모드 사유 교체에 빈 창이 없다
- **WHEN** 모드 투영이 진행되는 동안 진입 점검이 겹친다
- **THEN** 그 점검은 옛 모드 사유 또는 새 모드 사유 중 하나를 보며, 둘 다 없는 순간을 보지 않는다

#### Scenario: 사람이 모드를 완화한다
- **WHEN** 운영자가 엔진이 도는 중에 완화 명령으로 ENTRY_BLOCKED를 NORMAL로 되돌리며 이름·승인 참조·사유를 준다
- **THEN** audit 기록이 commit 전에 남고, 원장에 OPERATOR 행이 그 승인 참조와 함께 남으며, 산 엔진의 모드 사유 차단이 풀린다
- **AND** 다른 사유의 차단은 그대로이고 명령은 남은 사유를 보인다

#### Scenario: 자동 완화는 없다
- **WHEN** 자동 경로가 운영 모드를 NORMAL로 되돌리려 한다
- **THEN** 원장은 그 전이를 거절하고 모드는 그대로다

#### Scenario: audit를 쓰지 못한 완화는 일어나지 않는다
- **WHEN** 완화 요청의 audit 기록 쓰기가 실패한다
- **THEN** 원장에 완화 행이 없고 게이트의 모드 사유 차단도 그대로다

#### Scenario: 원장만 고친 완화는 산 게이트를 풀지 않는다
- **WHEN** 엔진 프로세스 밖에서 원장에 완화 행이 쓰인다
- **THEN** 산 엔진의 모드 사유 차단은 그대로이고, 다음 기동의 복원이 그 행을 투영한다 — 그래서 완화 명령은 엔진 프로세스를 거친다

#### Scenario: 겹친 전이의 투영이 뒤바뀌어 도착한다
- **WHEN** 두 전이가 차례로 커밋되고 그 투영이 반대 순서로 게이트에 도착한다
- **THEN** 게이트의 모드 사유는 나중에 커밋된 전이의 것이다

#### Scenario: 재알림 창 안의 재강화도 통지된다
- **WHEN** 운영자가 모드를 완화한 뒤 재알림 창 안에 관측 두절 트리거가 같은 목적 모드로 다시 강화한다
- **THEN** 그 재강화의 통지가 새로 기록되어 운영자에게 간다 — 옛 강화 통지의 정착 행에 흡수되지 않는다

#### Scenario: 전달 실패에 따른 강화는 통지하지 않는다
- **WHEN** critical 알림의 전달 실패가 지속되어 운영 모드가 `CRITICAL_ALERT_UNDELIVERED` 트리거로 강화된다
- **THEN** 그 전이의 통지는 기록되지 않고, 진입 차단 사유와 구조화 로그가 그 사실을 남긴다

#### Scenario: 벽시계가 뒤로 가도 복원은 커밋 순서의 최신이다
- **WHEN** 나중에 커밋된 전이의 기록 시각이 먼저 커밋된 전이의 시각보다 이르고, 그 뒤 엔진이 재시작한다
- **THEN** 기동 복원은 나중에 **커밋된** 전이의 모드를 세운다

#### Scenario: 기동 복원 직후의 첫 강화도 투영된다
- **WHEN** 기동 복원이 원장의 최신 모드를 세운 직후 자동 트리거가 운영 모드를 강화한다
- **THEN** 그 강화는 게이트에 적용된다

