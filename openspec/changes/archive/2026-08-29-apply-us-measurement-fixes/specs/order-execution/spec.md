# order-execution Specification (delta)

## ADDED Requirements

### Requirement: 취소의 일시적 거절 처리

브로커가 취소를 **일시적 사유로 거절**하면(실측: HTTP 409 `already-processing`, `data.retryAfterSeconds` 동반 — measurements.md M16) 취소는 유한 재시도해야 한다(SHALL). 대기는 브로커가 제시한 `retryAfterSeconds`를 따르되 상한을 둔다(SHALL — 현행 5초). 재시도 횟수에는 상한이 있어야 한다(SHALL — 현행 추가 2회). 재시도가 모두 실패하면 실패로 기록하고 잔여 객체를 보고한다(SHALL — 조용히 성공으로 만들지 않는다).

이 재시도는 **취소에만** 적용된다(SHALL NOT — 주문 접수·정정·조건주문 생성의 자동 재시도는 계속 금지된다). 근거: 취소는 노출을 줄이는 방향이고, 재시도는 이미 승인된 동일 취소의 반복이며, 취소하지 못한 미체결 주문을 계좌에 남기는 것이 더 큰 위험이다. 재시도 사실과 횟수는 관측 기록에 남겨야 한다(SHALL — 재시도로 성공한 취소를 첫 시도 성공으로 표시해서는 안 된다(SHALL NOT)).

#### Scenario: 일시적 거절 후 성공

- **WHEN** 취소가 `already-processing`으로 거절되고 재시도가 성공하면
- **THEN** 주문은 취소되고, 기록에는 재시도가 있었다는 사실과 횟수가 남는다

#### Scenario: 재시도 상한 소진

- **WHEN** 상한까지 재시도해도 취소가 계속 거절되면
- **THEN** 실패로 기록되고 그 주문이 살아 있다는 사실이 잔여 객체로 보고된다

#### Scenario: 접수·정정은 재시도하지 않는다

- **WHEN** 주문 접수나 정정이 실패하면
- **THEN** 자동 재시도는 일어나지 않는다
