# position-ledger Specification (delta)

## ADDED Requirements

### Requirement: Aggregate 경계
도메인 상태의 권위는 다음과 같이 분리되어야 한다(SHALL): OrderIntent/MutationAttempt(journal — P1), Fill(누적 스냅샷 — P1), **Position**(심볼·시장 단위 보유 상태 — 본 change), **ProtectionSaga**(진입-보호 연결 — protection-execution). Position은 체결 이벤트와 조정 이벤트의 투영이며 직접 변이 API를 노출하지 않는다(SHALL NOT). 포지션 진실은 하나여야 한다(SHALL): 기존 reconciliation의 로컬 상태 파생과 Position aggregate는 같은 질의를 권위로 삼고, 두 개의 독립 계산이 공존해서는 안 된다(SHALL NOT). 경계·이벤트 흐름은 `docs/aggregates.md`로 문서화한다(SHALL).

#### Scenario: 체결 반영으로만 포지션 변화
- **WHEN** 체결 스냅샷 delta가 반영되면
- **THEN** Position이 갱신되고, 그 외 어떤 코드 경로도 Position 수량을 직접 쓰지 못한다

#### Scenario: reconcile와 Position의 일치
- **WHEN** 같은 시점에 reconciliation의 로컬 상태와 Position aggregate를 각각 조회하면
- **THEN** 두 값이 같다

### Requirement: 포지션 상태기계
Position은 FLAT → OPENING → OPEN → (SCALING | CLOSING) → CLOSED 상태를 가지며(SHALL) 허용되지 않은 전이는 오류다. 전이는 `(현재 상태, 주문 역할, 누적 스냅샷 delta, lineage, 브로커 포지션)`의 완전한 전이표로 결정론적으로 정의되어야 하며(SHALL), 표는 다음을 모두 다룬다: 즉시 전량체결(FLAT→OPEN 직행), OPENING 종료 판단에 필요한 원주문 수량, SCALING 진입·종료 조건, 정정으로 교체된 주문의 lineage 승계, 외부 포지션 발견, 매도 체결의 귀속. 체결 delta는 부호가 없으므로 방향은 intent side에서 재도출한다(SHALL). CLOSED는 종결 상태이며 같은 심볼의 다음 거래는 새 position-instance 식별자로 시작한다(SHALL). 평균단가·수량은 이벤트 반영에서 재계산된다.

#### Scenario: 부분 청산 중 추가 체결
- **WHEN** CLOSING 상태에서 청산 주문의 부분체결이 반영되면
- **THEN** 잔여 수량이 감소하고 전량 체결 시 CLOSED로 전이한다

#### Scenario: 즉시 전량체결
- **WHEN** 진입 주문이 첫 체결 관측에서 전량 체결로 나타나면
- **THEN** 전이표에 따라 OPEN에 도달하며 OPENING 상태가 누락되어도 오류가 아니다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED된 심볼에 새 진입이 발생하면
- **THEN** 새로운 position-instance 식별자로 FLAT에서 시작하고 이전 인스턴스의 기록은 보존된다

### Requirement: 조정 이벤트
reconciliation이 계좌와 로컬의 차이를 보고하면 Position은 append-only 조정 이벤트로 보정되어야 하며(SHALL), 계좌 값이 권위다. 조정 이벤트는 Position 행을 직접 덮어쓰지 않고(SHALL NOT) 근거(브로커 스냅샷 시각·값)와 분류(외부 거래·수동 거래·원인 미상)를 함께 기록한다(SHALL). 청산·보호 수량은 로컬 값과 계좌 값 중 작은 쪽을 상한으로 사용한다(SHALL — oversell 방지). 불일치 상태의 진입 차단 해제 조건은 조정 반영 후 재조회 일치로 정의한다(SHALL).

#### Scenario: 계좌 수량이 로컬보다 적음
- **WHEN** reconciliation이 계좌 보유수량이 로컬보다 적다고 보고하면
- **THEN** 조정 이벤트가 기록되어 Position이 계좌 값으로 수렴하고, 그 사이의 청산 수량은 계좌 값을 상한으로 한다

#### Scenario: 조정의 provenance 보존
- **WHEN** 조정 이벤트로 보정된 Position의 이력을 조회하면
- **THEN** 체결 이벤트와 조정 이벤트가 구분되어 시간순으로 반환된다

### Requirement: 체결 반영의 원자성
체결 스냅샷 갱신, Position 투영 갱신, 보호 saga의 목표 수량 갱신·작업 등록은 하나의 journal 트랜잭션에서 커밋되어야 한다(SHALL). 체결만 커밋된 뒤 크래시해 Position이나 saga 생성이 누락되거나, 반대 순서로 중복 반영되는 것은 허용되지 않는다(SHALL NOT). 재처리 멱등성 키를 정의해 같은 스냅샷의 재관측이 이중 반영되지 않게 한다(SHALL).

#### Scenario: 체결 반영 직후 크래시
- **WHEN** 체결 반영 트랜잭션 커밋 직후 프로세스가 죽고 재시작되면
- **THEN** Position과 보호 작업이 모두 반영된 상태이거나 모두 미반영 상태이며, 중간 상태는 존재하지 않는다

### Requirement: Provenance Lineage
자동 거래의 전 과정은 결정 근거와 함께 추적 가능해야 한다(SHALL): 후보/신호 식별자(P3 예약) → GuardianDecision(reason 체인·RiskIntent 해시·한도 스냅샷·위험 예약) → intent → MutationAttempt → Fill → Position 변화 → 보호주문·발동 주문 → 청산까지 journal 내 참조로 연결된다. 임의 시점에 "이 포지션은 왜 존재하는가"를 단일 질의로 재구성할 수 있어야 한다(SHALL).

#### Scenario: 포지션 provenance 질의
- **WHEN** OPEN 포지션의 provenance를 질의하면
- **THEN** 진입 결정(Guardian 스냅샷)·주문·체결·보호주문 이력이 시간순으로 반환된다

### Requirement: 원장 스키마 확장 규칙
Position·ProtectionSaga·성과 테이블은 기존 journal DB의 additive 마이그레이션(v6)으로 추가되어야 하며(SHALL), P1 내구성 계약(BEGIN IMMEDIATE·synchronous=FULL·손상 기동 거부·구버전 바이너리 거부)을 그대로 상속한다. 별도 DB 파일을 만들지 않는다(SHALL NOT — 단일 writer 락·백업 단일화). 마이그레이션은 버전별로 분리되고 각각 immutable하며(SHALL), 테이블마다 키·외래키·unique 제약·append-only 여부를 명시하고 스키마 계약 테스트를 동반한다(SHALL). 마이그레이션 직전 자동 백업을 수행하고 복원 절차를 문서화·테스트한다(SHALL) — 구버전 바이너리 실행은 기동 거부이므로 롤백 수단이 아니다(SHALL NOT 롤백으로 간주).

#### Scenario: 구버전 바이너리로 신버전 DB 열기
- **WHEN** v6 스키마 DB를 구버전 바이너리가 열면
- **THEN** 기동이 거부되고, 복구 안내는 백업 복원 절차를 가리킨다

#### Scenario: 마이그레이션 실패
- **WHEN** v5→v6 마이그레이션이 중간에 실패하면
- **THEN** 직전 백업으로 복원할 수 있고 DB는 손상 상태로 남지 않는다
