# position-ledger Specification

## Purpose
포지션 투영(체결+조정 이벤트)·상태기계 전이표·조정 compare-and-append·체결 반영 원자성(apply hook)·provenance lineage·원장 스키마 확장 규칙.

## Requirements

### Requirement: Position 투영과 단일 권위
Position은 journal의 체결 이벤트와 조정 이벤트의 **투영**이며(SHALL) 직접 변이 API를 노출하지 않는다(SHALL NOT). 투영은 심볼·시장 단위 집계, position-instance 식별자(CLOSED 후 재진입은 새 인스턴스), 평균단가, 수량을 산출하고, 엔진 발 포지션은 진입 결정 참조(`entry_decision_id`)를 가진다(SHALL — 외부·수동 편입 포지션은 NULL이며 exit 정책 대상이 아니다). 방향은 intent side에서 재도출한다(SHALL — 이 change의 범위에서 모든 로컬 체결은 intent가 있는 주문에서 온다; 발동 주문의 방향 출처는 보호주문 도입 change가 정의한다). 포지션 진실은 하나이므로 reconciliation의 로컬 상태는 이 투영을 소비한다(SHALL — reconciliation delta). decimal 문자열 산술을 사용한다(SHALL NOT float 누적).

#### Scenario: 체결 반영으로만 포지션 변화
- **WHEN** 체결 delta가 반영되면
- **THEN** Position 투영이 갱신되고, 그 외 어떤 코드 경로도 Position 수량을 직접 쓰지 못한다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED된 심볼에 새 진입이 발생하면
- **THEN** 새 position-instance 식별자와 새 진입 결정 참조로 시작하고 이전 인스턴스 기록은 보존된다

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
이 change의 테이블은 기존 journal DB의 **단일 원자 additive 마이그레이션(v6)** 하나로 추가되며(SHALL — 한 버전 번호에 하나의 스키마), order-execution 메인 스펙의 Journal 내구성·백업·복원 계약을 상속한다(재서술하지 않음). 테이블 정의는 design D7 표가 정본이고 태스크는 전사다(SHALL).

#### Scenario: 구버전 바이너리로 v6 열기
- **WHEN** v6 스키마 DB를 구버전 바이너리가 열면
- **THEN** 기동이 거부되고 복구 안내는 백업 복원 절차를 가리킨다
