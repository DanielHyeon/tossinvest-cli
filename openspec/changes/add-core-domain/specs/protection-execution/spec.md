# protection-execution Specification (delta)

## ADDED Requirements

### Requirement: 네이티브 조건주문 우선
보호주문(손절, 손절+익절)은 토스 공식 조건주문(SINGLE/OCO)을 우선 사용해야 한다(SHALL) — 브로커측에 상주하므로 프로세스 사망·네트워크 단절에도 보호가 유지된다. 조건주문이 불가한 케이스(지원 범위 밖 시장·유형)만 로컬 synthetic 감시로 폴백하며, 폴백 활성 시 critical 알림과 함께 해당 사실이 journal에 기록된다(SHALL). 조건주문 경로는 P1의 `trading.Service` Conditional 메서드를 경유한다(SHALL — upstream 직접 호출 금지).

#### Scenario: 진입 체결 후 OCO 등록
- **WHEN** 진입이 전량 체결되면
- **THEN** stop(필수)과 target(선택)을 담은 조건주문이 브로커에 등록되고 saga가 ACTIVE로 전이한다

#### Scenario: 조건주문 미지원 케이스
- **WHEN** 조건주문이 거부되는 시장·유형이면
- **THEN** synthetic 폴백이 활성화되고 critical 알림이 발송된다

### Requirement: 진입-보호 Saga
진입 체결부터 보호 완료까지는 durable saga로 관리되어야 한다(SHALL): 진입 체결 감지 → 보호주문 제출(stop-first) → 등록 확인 → ACTIVE. 각 단계는 journal에 영속되고 재시작 시 미완 saga는 복구 절차(보호 재확인·재제출)를 거친다(SHALL). 진입 체결 후 보호 등록까지의 무보호 노출 시간 SLO(기본 10초)를 측정하고 위반 시 critical 알림을 발송한다(SHALL). 보호 제출이 IN_DOUBT면 해소 전까지 해당 심볼 신규 진입이 차단된다.

#### Scenario: 보호 제출 전 크래시
- **WHEN** 진입 체결 후 보호주문 제출 전에 프로세스가 죽고 재시작되면
- **THEN** 복구가 미보호 포지션을 감지해 보호주문을 제출하고 critical 알림을 발송한다

### Requirement: 부분체결 수량 정합
보호주문 수량은 실제 체결 수량과 항상 일치해야 한다(SHALL): 진입 부분체결 시 체결분만큼 보호를 등록·정정하고, 청산·손절 체결 시 잔여 보호 수량을 조정하며, oversell이 될 수 있는 보호 정정은 취소-후-재등록으로 수행한다. 포지션 CLOSED 시 잔여 보호주문은 취소된다(SHALL).

#### Scenario: 진입 50% 체결 상태의 보호
- **WHEN** 진입 주문이 절반만 체결된 상태로 유지되면
- **THEN** 보호주문 수량은 체결분과 같고, 추가 체결 시 보호 수량이 따라 조정된다

### Requirement: 손절 즉시성 보존
보호 경로(손절 발동·청산)는 Guardian 진입 판정·staleness 차단·심볼 진입 차단의 영향을 받지 않아야 하며(SHALL NOT — §0.3), kill switch·EXIT_ONLY·영구 불일치 상태에서도 계속 동작한다(SHALL). 네이티브 조건주문의 발동은 브로커 책임이므로 로컬 상태와 무관하게 유지된다.

#### Scenario: 진입 차단 latch 중 손절 발동
- **WHEN** 계정이 진입 차단 상태에서 브로커측 stop 조건주문이 발동되면
- **THEN** 체결이 정상 감지·반영되고 어떤 로컬 차단도 이를 방해하지 않는다
