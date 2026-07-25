# protection-execution Specification (delta)

## ADDED Requirements

### Requirement: 네이티브 조건주문 우선
보호주문(손절, 손절+익절)은 토스 공식 조건주문(SINGLE/OCO)을 우선 사용해야 한다(SHALL) — 브로커측에 상주하므로 프로세스 사망·네트워크 단절에도 보호가 유지된다. 이 안전성은 검증된 속성에 한해 주장할 수 있으며(SHALL), 다음이 실계좌 능력 attestation으로 확인되지 않은 시장·주문 유형에서는 자동 진입을 하지 않는다(SHALL NOT): 프로세스 사망 후 존속, 시장별 지원, 트리거 기준가, 정규장 밖 동작, 만료·timezone, OCO sibling 취소, 부분체결 잔량 처리, 정정 원자성. 모든 보호주문 mutation은 ExecutionGateway를 경유한다(SHALL — 실행 계약).

#### Scenario: 진입 체결 후 보호주문 등록
- **WHEN** 진입이 체결되면
- **THEN** 검증된 유형의 조건주문이 Gateway를 경유해 등록되고 saga가 ACTIVE로 전이한다

#### Scenario: 미검증 시장에서의 진입 시도
- **WHEN** 조건주문 능력이 검증되지 않은 시장의 진입 의도가 평가되면
- **THEN** 자동 진입이 거부된다

### Requirement: 진입-보호 Saga 상태기계
진입 체결부터 보호 완료까지는 durable saga로 관리되어야 하며(SHALL), 상태는 DESIRED → RECORDED → DISPATCHED → (ACTIVE | AMBIGUOUS) → (REPLACING | CANCELING | DEGRADED) → CLOSED로 정의된다(SHALL). 각 상태의 허용 mutation·timeout·재시작 행동을 전이표로 명시하고(SHALL), 각 journal 커밋 전후의 crash point를 열거한 fault-injection 수용 매트릭스를 산출물로 만든다(SHALL). 진입 체결 후 보호 등록까지의 무보호 노출 시간 SLO(기본 10초)를 측정하고 위반 시 critical 알림을 발송한다(SHALL). 보호 제출이 AMBIGUOUS면 실행 계약의 해소 절차를 거치며 해소 전 재제출은 금지된다(SHALL NOT).

#### Scenario: 보호 제출 전 크래시
- **WHEN** 진입 체결 후 보호주문 제출 전에 프로세스가 죽고 재시작되면
- **THEN** 복구가 미보호 포지션을 감지해 보호주문을 제출하고 critical 알림을 발송한다

#### Scenario: 보호 제출 응답 유실 후 크래시
- **WHEN** 보호주문 요청 전송 후 응답을 받기 전에 프로세스가 죽고 재시작되면
- **THEN** saga는 AMBIGUOUS로 복구되어 조회로 해소하며, 해소 전에는 재제출하지 않아 중복 보호주문이 생기지 않는다

#### Scenario: HALT_ALL 중 미보호 포지션 복구
- **WHEN** HALT_ALL 상태에서 재시작 복구가 미보호 포지션을 발견하면
- **THEN** 보호주문 제출이 위험 감소 mutation으로 허용된다

### Requirement: 보호 수량 정합
보호주문 수량은 체결 수량을 따라야 한다(SHALL). 동시 체결 환경에서 순간적 일치는 보장할 수 없으므로, 계약은 측정 가능한 transient invariant로 정의한다(SHALL): 보호 수량이 보유 수량과 어긋난 상태의 최대 허용 지속 시간을 정하고, 초과 시 critical 알림과 함께 진입을 차단한다. 수량 정정은 원자적 정정을 우선 사용하며(SHALL), 정정 원자성이 능력 검증에서 확인되지 않은 경우에만 취소-후-재등록으로 폴백한다. 어떤 경우에도 보호 수량 합계가 보유 수량을 초과해서는 안 된다(SHALL NOT — oversell 방지). 포지션 CLOSED 시 잔여 보호주문은 취소된다(SHALL).

#### Scenario: 진입 추가 체결에 따른 보호 증량
- **WHEN** 부분체결 상태에서 추가 체결이 반영되면
- **THEN** 보호 수량이 원자적 정정으로 증가하고, 어긋난 시간이 허용치를 넘지 않는다

#### Scenario: 정정 원자성 미검증
- **WHEN** 정정 원자성이 확인되지 않은 상태에서 수량을 줄여야 하면
- **THEN** 취소-후-재등록으로 수행되고 무보호 구간이 측정·기록된다

### Requirement: 폴백 전환 조건
로컬 synthetic 감시로의 폴백은 브로커가 해당 주문 유형·시장을 **명시적으로 미지원**한다고 응답한 경우에만 허용된다(SHALL). timeout·5xx·응답 유실 같은 모호한 결과에서는 폴백을 활성화해서는 안 되며(SHALL NOT — 모호한 native 등록 뒤 synthetic을 추가하면 이중 청산이 가능하다), 조회로 해소한 뒤 결정한다. 폴백 활성 시 critical 알림과 journal 기록이 남는다(SHALL). 보호 확인이 끝내 실패하면 미체결 진입 잔량을 취소하고 운영자 승인 기반 긴급 청산 상태로 전환한다(SHALL). synthetic 청산은 시장가가 아닌 공격적 limit을 사용하며, 급락·거래정지에서 체결을 보장하지 못한다는 한계를 문서화한다(SHALL).

#### Scenario: 명시적 미지원 응답
- **WHEN** 조건주문이 미지원 유형으로 명시적으로 거부되면
- **THEN** synthetic 폴백이 활성화되고 critical 알림이 발송된다

#### Scenario: 보호 제출 timeout
- **WHEN** 조건주문 등록이 timeout으로 끝나면
- **THEN** 폴백이 활성화되지 않고 해소 조회가 먼저 수행된다

### Requirement: 손절 즉시성 보존
보호 경로(보호주문 제출·발동·청산)는 Guardian 진입 판정·staleness 차단·심볼 진입 차단·진입 IN_DOUBT의 영향을 받지 않아야 하며(SHALL NOT — §0.3), kill switch·EXIT_ONLY·HALT_ALL·영구 불일치 상태에서도 계속 동작한다(SHALL). 이는 실행 계약의 위험 감소 mutation 클래스로 구현된다. 네이티브 조건주문의 발동은 브로커 책임이므로 로컬 상태와 무관하게 유지된다.

#### Scenario: 진입 차단 latch 중 손절 발동
- **WHEN** 계정이 진입 차단 상태에서 브로커측 stop 조건주문이 발동되면
- **THEN** 체결이 정상 감지·귀속되고 어떤 로컬 차단도 이를 방해하지 않는다

#### Scenario: 영구 불일치 상태의 보호 제출
- **WHEN** UNRESOLVED_IN_DOUBT로 심볼 진입이 영구 차단된 상태에서 보호주문 제출이 필요하면
- **THEN** 제출이 수행되며 수량은 확정 보유수량 이하로 제한된다
