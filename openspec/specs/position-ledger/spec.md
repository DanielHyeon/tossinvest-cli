# position-ledger Specification

## Purpose
포지션 투영(체결+조정 이벤트)·상태기계 전이표·조정 compare-and-append·체결 반영 원자성(apply hook)·provenance lineage·원장 스키마 확장 규칙.

## Requirements

### Requirement: Position 투영과 단일 권위

Position은 journal의 체결 이벤트와 조정 이벤트의 **투영**이며(SHALL) 직접 변이 API를 노출하지 않는다(SHALL NOT). 투영은 심볼·시장 단위 집계, position-instance 식별자(CLOSED 후 재진입은 새 인스턴스), 평균단가, 수량을 산출하고, 엔진 발 포지션은 진입 결정 참조(`entry_decision_id`)를 가진다(SHALL — 설정 후 인스턴스 수명 동안 변경·NULL화되지 않으며(SHALL NOT), 외부·수동 취득 포지션은 NULL로 남는다). 외부 취득 포지션의 편입은 `positions.adoption_id`(additive v7, `position_adoptions` 단방향 참조 — 편입 행의 심볼·수량 필드는 편입 시점 스냅샷이고 권위는 positions 투영이다)로 기록되며 **set-once**다(SHALL — 전용 tx API로만 기입하고 그 외 UPDATE의 언급은 정적 스캔이 거부한다). exit 정책 대상 자격은 `entry_decision_id 또는 adoption_id`가 설정된 포지션이며(SHALL — 자격 판정은 단일 술어 함수로 모은다), reconcile fold의 "entry 결정 상속 금지" 가드는 자격 술어가 아닌 `entry_decision_id` 명시 비교로 좁혀 유지한다(SHALL — 편입 포지션에 fold가 착지하는 것은 정상 재대사 경로다). adoption_id 부여는 수량·평균단가를 변경하지 않는다(SHALL NOT). 방향은 intent side에서 재도출한다(SHALL — 이 change의 범위에서 모든 로컬 체결은 intent가 있는 주문에서 온다; 발동 주문의 방향 출처는 보호주문 도입 change가 정의한다). 포지션 진실은 하나이므로 reconciliation의 로컬 상태는 이 투영을 소비한다(SHALL — reconciliation delta). decimal 문자열 산술을 사용한다(SHALL NOT float 누적 — 편입 원가(cost_basis)는 브로커 원문 decimal 문자열을 보존하며, 편입 관측가는 엔진 공통 가격 경로를 따른다 `[기존 제약 — 가격 경로 float64]`).

편입 포지션의 lineage 형태(SHALL 명시): `ADOPTION → POSITION → EXIT_EVENT …`이며 intent·MutationAttempt·Fill arm이 비는 것이 정상이다(엔진이 실행한 청산 fill은 그 시점부터 정상 arm으로 나타난다). 조정 이벤트로 포지션 수량이 0이 되면 exit_state는 completed 처리되고 ADJUSTMENT_CLOSED exit_event가 기록되며 알림된다(SHALL — 고아 exit_state 금지). 이 경우 trade_outcome 행은 만들지 않는다(SHALL NOT — 매도가 엔진 밖에서 일어나 매도 leg이 없다; adopt-external-positions design A7).

#### Scenario: 체결 반영으로만 포지션 변화
- **WHEN** 체결 delta가 반영되면
- **THEN** Position 투영이 갱신되고, 그 외 어떤 코드 경로도 Position 수량을 직접 쓰지 못한다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED된 심볼에 새 진입이 발생하면
- **THEN** 새 position-instance 식별자와 새 진입 결정 참조로 시작하고 이전 인스턴스 기록은 보존된다

#### Scenario: 편입 기록의 set-once
- **WHEN** adoption_id가 이미 설정된 포지션에 재기입이 시도되면
- **THEN** 전용 API가 거부하고 기존 참조는 변하지 않는다

#### Scenario: 외부 매도로 수량 0
- **WHEN** 편입 포지션의 수량이 조정 이벤트로 0이 되면
- **THEN** exit_state가 completed 처리되고 ADJUSTMENT_CLOSED exit_event와 알림이 발생하며 trade_outcome 행은 생성되지 않는다

### Requirement: 포지션 상태기계
Position은 FLAT → OPENING → OPEN → (SCALING | CLOSING) → CLOSED 상태를 가지며(SHALL) 전이는 `(현재 상태, 주문 역할, 누적 delta, lineage)`의 **완전한 전이표**로 결정론적으로 정의된다(SHALL — 전이표 전체는 design 산출물이며 태스크는 표의 전 행을 테스트한다). 표는 즉시 전량체결(FLAT→OPEN 직행 허용), OPENING 종료 판단(원주문 수량), SCALING 진입·종료, 정정 교체 주문의 lineage 승계, 매도 체결 귀속, CLOSED 종결성을 다룬다. 허용되지 않은 전이는 오류이며 RECONCILE로 전이한다(SHALL — 산식 보정 금지).

#### Scenario: 부분 청산 중 추가 체결
- **WHEN** CLOSING 상태에서 청산 주문의 부분체결이 반영되면
- **THEN** 잔여 수량이 감소하고 전량 체결 시 CLOSED로 전이한다

#### Scenario: 즉시 전량체결
- **WHEN** 진입 주문이 첫 관측에서 전량 체결로 나타나면
- **THEN** 전이표에 따라 OPEN에 도달하며 오류가 아니다

### Requirement: 조정 이벤트
reconciliation이 보고한 차이는 append-only 조정 이벤트로 보정되며(SHALL) 계좌 값이 권위다. 조정 이벤트는 근거(브로커 스냅샷 시각·값)·분류(외부/수동/미상)·**기대 이전 값**을 기록하고(SHALL), 커밋 트랜잭션 안에서 기대 이전 값·체결 watermark의 불변을 재검증해 어긋나면 폐기·재수집한다(SHALL — reconciliation delta의 compare-and-append). Position 행 직접 덮어쓰기는 금지된다(SHALL NOT). 조정 중·후의 청산 수량은 order-execution 메인 스펙의 RECONCILE 확정 하한을 따른다(SHALL — 재정의하지 않음).

#### Scenario: 계좌 수량이 로컬보다 적음
- **WHEN** reconciliation이 계좌 보유수량이 로컬보다 적다고 보고하면
- **THEN** 조정 이벤트가 기대 이전 값 검증을 통과해 기록되고 투영이 계좌 값으로 수렴한다

### Requirement: 체결 반영의 원자성 — tx-scoped apply hook
체결 스냅샷 갱신, Position 투영 갱신, exit 상태 갱신(pending 해소·taken_ratio 이동 포함)은 **journal이 소유한 원자 apply 지점**에서 하나의 트랜잭션으로 커밋되어야 한다(SHALL). 설계 형태는 tx-scoped hook이다(SHALL — 체결 반영 트랜잭션이 주입된 투영·exit 적용 함수를 같은 트랜잭션 스코프에서 호출한다; journal 밖의 후속 별도 커밋은 원자성 요구를 만족하지 않고, journal이 도메인 상태기계를 직접 소유하는 것도 아니다). `taken_ratio_total`은 이 지점에서만 이동한다(SHALL — exit-policy의 체결 시점 필드). 재처리 멱등성은 P1 누적 watermark가 담보한다.

#### Scenario: 체결 반영 직후 크래시
- **WHEN** 체결 반영 트랜잭션 커밋 직후 프로세스가 죽고 재시작되면
- **THEN** 투영·exit 상태·pending 해소가 모두 반영됐거나 모두 미반영이며, 중간 상태는 없다

### Requirement: Provenance Lineage
자동 거래의 전 과정은 결정 근거와 함께 추적 가능해야 한다(SHALL): 후보/신호 식별자(P3 예약, nullable) → GuardianDecision(preimage·한도·예약) → intent → MutationAttempt → Fill → Position 변화(`entry_decision_id` 조인) → **exit 판정 이벤트(append-only `exit_events`)** → 청산까지 journal 내 참조로 연결된다. "이 포지션은 왜 존재하는가"를 단일 질의로 재구성할 수 있어야 하며(SHALL), 그 조인 경로는 스키마의 명시적 참조 컬럼이다(SHALL — 시간창 휴리스틱 매칭 금지).

#### Scenario: 포지션 provenance 질의
- **WHEN** OPEN 포지션의 provenance를 질의하면
- **THEN** 진입 결정(preimage)·주문·체결·exit 판정 이벤트가 명시적 참조 조인으로 시간순 반환된다

### Requirement: 원장 스키마 확장 규칙

원장 스키마 변경은 additive만 허용된다(SHALL — 기존 테이블 재작성·컬럼 삭제·CHECK 완화 금지). 이 규칙 하에서 core domain 테이블은 단일 원자 additive 마이그레이션 v6으로 추가되었고(테이블 정의는 add-core-domain design D7 표가 정본), **편입 테이블·컬럼은 단일 원자 additive 마이그레이션 v7로 추가된다**(SHALL — 정의는 adopt-external-positions design A1이 정본; positions.adoption_id는 DEFAULT 없는 nullable ADD COLUMN). 스키마 골든 목록 테스트(wantTables·버전별 테이블 목록)는 각 마이그레이션과 함께 갱신된다(SHALL). 구버전 바이너리는 더 새로운 on-disk 스키마를 ErrSchemaTooNew로 거부한다(SHALL — §0.6 rollback 계약).

#### Scenario: v7 적용 후 구버전 기동
- **WHEN** v7이 적용된 journal을 v6 바이너리가 열면
- **THEN** ErrSchemaTooNew로 거부되고 어떤 쓰기도 일어나지 않는다

