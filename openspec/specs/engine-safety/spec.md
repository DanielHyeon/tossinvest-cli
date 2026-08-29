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
> **a092의 header note(`a092/specs/engine-safety/spec.md:24-25`)와 어긋난다.**
> 그것은 *"「엔진 런타임 수명주기」는 루프를 **하나 더한다** — 알림 배달 루프"*라고
> 적는다. 결정 9-2는 더하지 **않는다.** 그 문장의 정리는 a092가 진다 (a099 §7.5).

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
엔진은 현재 계좌·profile·시장·주문유형·수량·세션·trigger source와 atomic/continuous replace semantics가 strict versioned attestation과 일치하고 protection saga가 배선된 경우에만 exposure-raising mutation을 허용해야 한다 (SHALL). attestation은 tool/build와 evidence digest에 묶이고 legacy/unknown field, 만료, 경로·소유자·권한 불일치는 fail-closed해야 한다 (SHALL).

#### Scenario: 미검증 시장
- **WHEN** KR capability만 attested된 상태에서 US 자동 진입을 시도한다
- **THEN** entry는 protection_unwired로 거부되고 기존 US 보유의 reduce-only exit는 계속된다

#### Scenario: 유효 capability
- **WHEN** 현재 profile이 attestation과 일치하고 Guardian/gate가 유효하다
- **THEN** protection readiness clause는 충족되지만 운영자가 승인하지 않은 lane는 여전히 OFF다

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

**배달 실행자는 감독 루프가 아니어야 한다**(SHALL NOT — 사용자 결정 9-2).
그 성질을 **판정으로 만들어서는 안 된다**(SHALL NOT): 감독 루프로 등록한 뒤
*"이 이름이면 다르게 다룬다"*로 거르는 구성은 **판정이 틀리면 무너진다.**
등록하지 않으면 **틀릴 판정이 없다.**

> **⛔ 6판이 이 문단의 두 번째 근거를 지웠다.** 5판까지는 여기에
> *"엔진 런타임의 루프 집합을 **셋**으로 못 박은 승인된 정본과도 어긋난다"*가
> 붙어 있었다. **그 근거는 쓸 수 없다** — 정본의 셋은 프로덕션의 넷과 이미 다르고
> (`cmd/tossctl/engine.go:377-398`), 이 델타의 `MODIFIED`가 그것을 **넷으로 고친다**
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

> **근거 둘.** ① `journal.SetModeProjector`가 프로덕션에서 bind되지 않으므로
> 그 승격은 **산 프로세스의 진입 게이트에 닿지 않는다** — a092가 발견했고 배선은
> 미배정 후속이다. **a098은 그 배선을 기다리지 않는다.**
> ② 승격은 원장에 남고 완화에 **사람 승인**이 필요하므로
> (`openspec/specs/risk-management/spec.md:102-108`), 실행자가 살아난 뒤에도
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
> `EntryGate.Block`은 `g.mu`와 맵뿐이라(`internal/execgw/retry.go:498-505`)
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
> 모드로 승격하면 사유가 사라진 뒤에도 **사람 승인 없이는 안 풀린다**(`risk-management` `:102-108`) —
> 그것은 *"복구는 재시작"*을 **거짓으로 만든다.**
>
> **덤으로 얻은 것 셋.** `internal/journal` diff가 **비어 있는 채로 남고**(§5.2),
> 승인된 `risk-management`에 **`MODIFIED`가 필요 없고**,
> 자동 트리거의 닫힌 열거(`operating_mode.go:75-96`·`:513-547`)를 **안 건드린다.**
>
> **잃은 것 하나는 a092가 진다** — 결정 12-2. a092의 같은 Scenario
> (`a092/specs/engine-safety/spec.md:126-130`)는 모드 승격을 SHALL로 적고 있고,
> **a098은 그것을 안 진다.** 인계는 a099 §7.5.

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
> | `EntryGate.Block` | `internal/execgw/retry.go:498-505` — **`g.mu`와 맵뿐, 원장에 안 닿는다** | **사라진다** |
> | 운영 모드 승격 | `TransitionOperatingMode`가 `tx.Commit()`(`operating_mode.go:468`) | **남는다** |
>
> 배달 실행자가 죽고 게이트만 잠긴 상태에서 프로세스를 재시작하면
> **미전달 critical 알림이 그대로인데 신규 진입이 다시 열린다.**
> 그것이 a092의 셋째 조항이 있던 이유다.
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

엔진의 strategy projection endpoint(control 디렉터리·descriptor·socket)의 기동 시 잔재 회수는 그 생성·종료·회수 시퀀스가 만들 수 있는 모든 부분 상태(빈 디렉터리, descriptor만, socket만, 둘 다, 쓰다 만 산출물과 staging 잔재)를 소유자 사망 검증 후 사람 개입 없이 회수해야 하며(SHALL), 산출물 발행은 부분 상태가 최종 이름에 나타나지 않도록 stage+rename으로 해야 하고(SHALL), 소유자 생존 판정은 프로세스 ID 재사용에 오판되지 않는 수단이어야 하며(SHALL) kill-0 단독 판정은 금지되고(SHALL NOT), 소유권·symlink 검증과 낯선 엔트리의 거부는 유지되어야 한다(SHALL).

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

