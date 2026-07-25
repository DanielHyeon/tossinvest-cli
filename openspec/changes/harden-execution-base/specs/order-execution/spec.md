# order-execution Specification (delta)

## ADDED Requirements

### Requirement: Intent-Mutation-주문 분리 모델
주문 실행 모델은 세 노드를 분리해야 한다(SHALL): 불변 OrderIntent(의도), MutationAttempt(PLACE/CANCEL/AMEND 시도 — 각각 독립 수명주기), 브로커 주문 노드(주문번호 단위). 토스 공식 API의 cancel/modify는 새 주문번호를 발급하므로, 주문 노드 간 관계는 lineage edge(`replaces`, 원주문 체결수량, 요청 잔여수량, 새 주문번호)로 기록해야 한다(SHALL). 엔진 경로의 lineage 기록은 journal DB 안에서 트랜잭션으로 수행한다(SHALL).

#### Scenario: 부분체결 중 정정
- **WHEN** 부분체결된 주문에 AMEND가 성공해 새 주문번호가 발급되면
- **THEN** 원주문 노드는 체결수량과 함께 종결되고, 새 주문 노드가 `replaces` edge·잔여수량과 함께 생성되며, 두 기록은 동일 트랜잭션에서 커밋된다

#### Scenario: 다단계 정정 체인
- **WHEN** 정정이 2회 연속 수행되면
- **THEN** lineage 체인으로 최초 주문번호에서 현재 주문번호를 결정적으로 해소할 수 있다

### Requirement: MutationAttempt 수명주기
각 MutationAttempt는 RECORDED → DISPATCH_STARTED → (ACKED | IN_DOUBT) → 종결(CONFIRMED | NOT_DISPATCHED | FAILED_CONFIRMED | UNRESOLVED_IN_DOUBT) 단계를 가져야 한다(SHALL). RECORDED는 fsync 완료 후에만 DISPATCH_STARTED로 진행하며(SHALL), 재시작 시 RECORDED 단계에서 멈춘 attempt는 NOT_DISPATCHED로 안전 종결하고, DISPATCH_STARTED 이후 단계는 해소 절차 완료 전까지 차단 대상으로 취급한다(SHALL).

#### Scenario: 전송 시작 전 크래시
- **WHEN** RECORDED까지만 기록된 attempt가 재시작 시 발견되면
- **THEN** NOT_DISPATCHED로 종결되고 어떤 차단도 발생하지 않는다

#### Scenario: 전송 중 크래시
- **WHEN** DISPATCH_STARTED로 기록된 attempt가 재시작 시 발견되면
- **THEN** IN_DOUBT 해소 절차가 시작되고 완료 전까지 해당 심볼 신규 mutation이 차단된다

### Requirement: IN_DOUBT 해소
IN_DOUBT 해소는 다음을 모두 만족해야 한다(SHALL): (1) 자동 재제출 절대 금지 — 브로커 멱등성 키가 없으므로 무조건, (2) journal에 저장된 fingerprint(계좌·심볼·방향·수량·가격·제출 시각 창)로 OPEN과 CLOSED **양쪽** 목록을 pagination 완주하며 대조, (3) 부재 판정은 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 + 매수가능금액/보유수량 delta 교차 확인 후에만, (4) 증명 불가 시 UNRESOLVED_IN_DOUBT로 영구 차단하고 운영자 해소만 허용. 유일 매칭을 보장하기 위해 엔진은 심볼당 in-flight mutation을 항상 1개로 제한한다(SHALL).

#### Scenario: 제출 응답 유실 후 주문이 2페이지에 존재
- **WHEN** IN_DOUBT 해소 조회에서 대상 주문이 목록 2페이지 이후에 있으면
- **THEN** pagination 완주로 발견되어 CONFIRMED로 종결된다

#### Scenario: 단발 부재 조회
- **WHEN** 첫 조회에서 주문이 보이지 않으면
- **THEN** FAILED로 판정하지 않고 안정화 재조회를 계속한다

#### Scenario: 해소 불능
- **WHEN** 관찰 기간 내 존재도 부재도 증명되지 않으면
- **THEN** UNRESOLVED_IN_DOUBT로 표기되어 해당 심볼이 영구 차단되고 운영자 알림이 발송된다

### Requirement: 브로커 상태 파생
주문 종결 상태는 공식 API의 원시 status만이 아니라 `(status, canceledAt, execution.filledQuantity, quantity, lineage)`에 대한 우선순위 파생 함수로 결정해야 한다(SHALL). 관측된 공식 status는 OPEN/CLOSED 수준임을 전제로 하고, 모순 조합·미지의 status 값은 UNKNOWN_BROKER_STATE로 fail-closed 처리해야 한다(SHALL: 해당 심볼 신규 진입 차단 + 알림).

#### Scenario: CLOSED + canceledAt 존재
- **WHEN** status=CLOSED이고 canceledAt이 설정된 주문을 파생하면
- **THEN** CANCELLED로 판정된다 (filledQuantity>0이면 부분체결 후 취소로 기록)

#### Scenario: 미지의 status 값
- **WHEN** 파생 함수가 알 수 없는 status 문자열을 받으면
- **THEN** UNKNOWN_BROKER_STATE로 fail-closed 처리되고 알림이 발송된다

### Requirement: Journal 내구성
Journal은 SQLite 단일 파일로 하되(SHALL): 저장 경로는 `$TOSSOS_DATA_DIR` > `$XDG_DATA_HOME/tossos` > `~/.local/share/tossos` 순으로 해석하고, 로컬 저널링 파일시스템 allowlist(ext4/xfs/btrfs) 밖이면 기동을 거부한다(SHALL). intent 기록은 `BEGIN IMMEDIATE` + `synchronous=FULL`로 커밋 성공 후에만 제출을 진행하고(SHALL), 스키마는 버전 필드와 additive migration 규칙을 가지며, 손상 감지 시 기동을 거부한다(SHALL). Phase 2 원장이 이 journal을 import할 수 있도록 안정 primary key·불변 intent 필드를 문서화한다.

#### Scenario: 커밋 실패 시 제출 차단
- **WHEN** journal 트랜잭션 커밋이 실패하면
- **THEN** 브로커 제출은 수행되지 않는다

#### Scenario: DB 손상
- **WHEN** 기동 시 journal 무결성 검사가 실패하면
- **THEN** 기동이 거부되고 복구 안내가 출력된다

### Requirement: Retry Matrix 산출물
재시도 정책은 endpoint×method×오류 클래스 표로 스펙 산출물화해야 한다(SHALL). 최소 규정: 주문 mutation은 어떤 오류에도 자동 재시도 금지, 조회는 재시도 예산(횟수·총 시간)과 bounded jitter, 429는 Retry-After 상한 준수, 401/403은 즉시 신규 진입 차단 + 알림, 필수 조회(잔고·미체결·가격)의 staleness가 임계를 넘으면 신규 진입 차단. 표의 수치는 구현 시 확정하고 표 없이 구현하지 않는다(SHALL NOT).

#### Scenario: 필수 조회 장기 실패
- **WHEN** 잔고 조회가 재시도 예산을 소진하고 staleness 임계를 초과하면
- **THEN** 신규 진입이 차단되고 조회 복구 후 자동 해제된다

### Requirement: 시간 규율
모든 시간 판정(제출 시각 창, staleness, 안정화 간격, 거래일 경계)은 주입 가능한 clock을 사용해야 하며(SHALL), 시장별 timezone(KST/ET)과 DST 전환을 명시적으로 처리해야 한다(SHALL). 거래일 경계는 시장별 규칙으로 정의한다.

#### Scenario: DST 전환일의 US 세션 판정
- **WHEN** 미국 DST 전환일에 세션 판정을 수행하면
- **THEN** ET 기준 정확한 장 시간이 사용된다

### Requirement: Fail-closed 분기
공식 주문 경로에서 interactive auth challenge 요구, USD 주문의 통화 잔고 부족, 미지원 주문 유형은 제출 없이 거부하고 사유 코드와 함께 기록·통지해야 한다(SHALL). 자동 환전·자동 승인은 금지된다(SHALL NOT). 차단·거부 사유는 안정적 reason-code enum으로 정의한다(SHALL).

#### Scenario: USD 잔고 부족 매수
- **WHEN** USD 매수 intent의 사전 잔고 확인이 부족을 반환하면
- **THEN** 제출 없이 reason code와 함께 거부·통지된다
