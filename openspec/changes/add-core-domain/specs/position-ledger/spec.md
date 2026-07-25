# position-ledger Specification (delta)

## ADDED Requirements

### Requirement: Aggregate 경계
도메인 상태의 권위는 다음과 같이 분리되어야 한다(SHALL): OrderIntent/MutationAttempt(journal — P1), Fill(누적 스냅샷 — P1), **Position**(심볼·시장 단위 보유 상태 — 본 change), **ProtectionSaga**(진입-보호 연결 — protection-execution). Position은 Fill 반영과 reconcile 결과에서만 파생되며 직접 변이 API를 노출하지 않는다(SHALL NOT). 경계·이벤트 흐름은 `docs/aggregates.md`로 문서화한다(SHALL).

#### Scenario: 체결 반영으로만 포지션 변화
- **WHEN** 체결 스냅샷 delta가 반영되면
- **THEN** Position이 갱신되고, 그 외 어떤 코드 경로도 Position 수량을 직접 쓰지 못한다

### Requirement: 포지션 상태기계
Position은 FLAT → OPENING → OPEN → (SCALING | CLOSING) → CLOSED 상태를 가지며(SHALL) 허용되지 않은 전이는 오류다. 평균단가·수량은 체결 반영에서 재계산되고, reconcile이 계좌와의 차이를 보고하면 계좌 값이 우선한다(SHALL). 부분체결 중 상태는 OPENING/CLOSING을 유지하며 보호 수량 조정(protection-execution)에 이벤트를 발행한다.

#### Scenario: 부분 청산 중 추가 체결
- **WHEN** CLOSING 상태에서 청산 주문의 부분체결이 반영되면
- **THEN** 잔여 수량이 감소하고 전량 체결 시 CLOSED로 전이한다

### Requirement: Provenance Lineage
자동 거래의 전 과정은 결정 근거와 함께 추적 가능해야 한다(SHALL): 후보/신호 식별자(P3 예약) → GuardianDecision(reason 체인·한도 스냅샷) → intent → MutationAttempt → Fill → Position 변화 → 청산까지 journal 내 참조로 연결된다. 임의 시점에 "이 포지션은 왜 존재하는가"를 단일 질의로 재구성할 수 있어야 한다(SHALL).

#### Scenario: 포지션 provenance 질의
- **WHEN** OPEN 포지션의 provenance를 질의하면
- **THEN** 진입 결정(Guardian 스냅샷)·주문·체결 이력이 시간순으로 반환된다

### Requirement: 원장 스키마 확장 규칙
Position·ProtectionSaga·provenance·성과 테이블은 기존 journal DB의 additive 마이그레이션(v5+)으로 추가되어야 하며(SHALL), P1 내구성 계약(BEGIN IMMEDIATE·synchronous=FULL·손상 기동 거부·구버전 바이너리 ErrSchemaTooNew)을 그대로 상속한다. 별도 DB 파일을 만들지 않는다(SHALL NOT — 단일 writer 락·백업 단일화).

#### Scenario: 구버전 바이너리로 v5 DB 열기
- **WHEN** v5 스키마 DB를 구버전 바이너리가 열면
- **THEN** ErrSchemaTooNew로 기동이 거부된다
