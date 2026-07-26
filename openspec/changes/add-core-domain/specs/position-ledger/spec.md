# position-ledger Specification (delta)

## ADDED Requirements

### Requirement: Position 투영과 단일 권위
Position은 journal의 체결 이벤트와 조정 이벤트의 **투영**이며(SHALL) 직접 변이 API를 노출하지 않는다(SHALL NOT). 투영은 심볼·시장 단위 집계, position-instance 식별자(CLOSED 후 재진입은 새 인스턴스), 평균단가, 수량을 산출한다(SHALL). 방향은 intent side에서 재도출한다(SHALL — 이 change의 범위에서 모든 체결은 로컬 intent가 있는 주문에서 온다; 발동 주문의 방향 출처는 보호주문 도입 change가 정의한다). 포지션 진실은 하나여야 하므로 reconciliation의 로컬 상태 파생은 이 투영을 소비하도록 재배선된다(SHALL — reconciliation delta 참조). float 누적이 아니라 decimal 문자열 산술을 사용한다(SHALL NOT float).

#### Scenario: 체결 반영으로만 포지션 변화
- **WHEN** 체결 delta가 반영되면
- **THEN** Position 투영이 갱신되고, 그 외 어떤 코드 경로도 Position 수량을 직접 쓰지 못한다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED된 심볼에 새 진입이 발생하면
- **THEN** 새 position-instance 식별자로 시작하고 이전 인스턴스 기록은 보존된다

### Requirement: 포지션 상태기계
Position은 FLAT → OPENING → OPEN → (SCALING | CLOSING) → CLOSED 상태를 가지며(SHALL) 전이는 `(현재 상태, 주문 역할, 누적 delta, lineage)`의 완전한 전이표로 결정론적으로 정의된다(SHALL). 표는 즉시 전량체결(FLAT→OPEN 직행 허용), OPENING 종료 판단(원주문 수량), SCALING 진입·종료, 정정 교체 주문의 lineage 승계, 매도 체결 귀속, CLOSED 종결성을 다룬다. 허용되지 않은 전이는 오류이며 RECONCILE로 전이한다(SHALL — 산식 보정 금지).

#### Scenario: 부분 청산 중 추가 체결
- **WHEN** CLOSING 상태에서 청산 주문의 부분체결이 반영되면
- **THEN** 잔여 수량이 감소하고 전량 체결 시 CLOSED로 전이한다

#### Scenario: 즉시 전량체결
- **WHEN** 진입 주문이 첫 관측에서 전량 체결로 나타나면
- **THEN** 전이표에 따라 OPEN에 도달하며 오류가 아니다

### Requirement: 조정 이벤트
reconciliation이 계좌와 로컬의 차이를 보고하면 Position은 append-only 조정 이벤트로 보정되어야 하며(SHALL), 계좌 값이 권위다. 조정 이벤트는 근거(브로커 스냅샷 시각·값)와 분류(외부 거래·수동 거래·원인 미상)를 기록하고(SHALL) Position 행을 직접 덮어쓰지 않는다(SHALL NOT). 조정 반영 중·후의 청산 수량은 order-execution 메인 스펙의 RECONCILE 확정 하한 공식을 따른다(SHALL — 이 스펙은 상한 규칙을 재정의하지 않는다).

#### Scenario: 계좌 수량이 로컬보다 적음
- **WHEN** reconciliation이 계좌 보유수량이 로컬보다 적다고 보고하면
- **THEN** 조정 이벤트가 기록되어 투영이 계좌 값으로 수렴하고 provenance에 조정이 구분되어 남는다

### Requirement: 체결 반영의 원자성
체결 스냅샷 갱신, Position 투영 갱신, exit 정책 상태(ladder·기준선) 갱신 트리거는 하나의 journal 트랜잭션에서 커밋되어야 한다(SHALL). 재처리 멱등성은 P1 누적 watermark가 담보하며, 같은 스냅샷의 재관측은 이중 반영되지 않는다(SHALL NOT).

#### Scenario: 체결 반영 직후 크래시
- **WHEN** 체결 반영 트랜잭션 커밋 직후 프로세스가 죽고 재시작되면
- **THEN** Position과 exit 상태가 모두 반영된 상태이거나 모두 미반영 상태이며, 중간 상태는 없다

### Requirement: Provenance Lineage
자동 거래의 전 과정은 결정 근거와 함께 추적 가능해야 한다(SHALL): 후보/신호 식별자(P3 예약, nullable) → GuardianDecision(preimage·한도 스냅샷·예약) → intent → MutationAttempt → Fill → Position 변화 → exit 판정 → 청산까지 journal 내 참조로 연결된다. 임의 시점에 "이 포지션은 왜 존재하는가"를 단일 질의로 재구성할 수 있어야 한다(SHALL).

#### Scenario: 포지션 provenance 질의
- **WHEN** OPEN 포지션의 provenance를 질의하면
- **THEN** 진입 결정(preimage)·주문·체결·exit 판정 이력이 시간순으로 반환된다

### Requirement: 원장 스키마 확장 규칙
Position 투영·조정 이벤트·exit 상태·성과 테이블은 기존 journal DB의 **단일 원자 additive 마이그레이션(v6)**으로 추가되어야 하며(SHALL — 버전별 immutable, 한 버전 번호에 하나의 스키마), P1·2a 내구성 계약(BEGIN IMMEDIATE·synchronous=FULL·손상 기동 거부·직전 자동 백업·구버전 ErrSchemaTooNew)을 그대로 상속한다. 별도 DB 파일을 만들지 않는다(SHALL NOT). 테이블마다 키·FK·unique 제약을 design에 명시하고 태스크는 전사한다(SHALL).

#### Scenario: 구버전 바이너리로 v6 열기
- **WHEN** v6 스키마 DB를 구버전 바이너리가 열면
- **THEN** 기동이 거부되고 복구 안내는 백업 복원 절차를 가리킨다
