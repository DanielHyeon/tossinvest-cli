## ADDED Requirements

### Requirement: 주문 화면의 청산 근거 연결은 무장 행만 센다
주문 화면이 브로커 주문을 청산 근거에 잇는 원장 리더는 intent 하나에 묶인 `exit_events` 행 가운데 무장 행만 세고 연결해야 한다(SHALL).
종결 훅 행(OPENED · READOPTED · PROPOSAL_REFUSED · PROPOSAL_CANCELLED · PROPOSAL_FILLED · COMPLETED · ADJUSTMENT_CLOSED)은 증거가 없는
수명주기 표지이므로 그 행 때문에 연결이 모호해져서는 안 된다(SHALL NOT). 훅 action 목록은 원장 패키지의 한 정의를 써야 한다(SHALL).
한 intent 에 무장 행이 둘 이상이면 훅 행 유무와 무관하게 지금처럼 `ambiguous_exit_evidence` 로 표시해야 한다(SHALL — fail-closed).
이 리더는 표시 전용이며 원장 쓰기 · 주문 · 손절 · 대사 판정을 바꾸어서는 안 된다(SHALL NOT).

#### Scenario: 체결된 청산 주문
- **WHEN** 청산 발의가 무장 행 하나를 쓰고 그 intent 의 매도 주문이 종결 체결되어 PROPOSAL_FILLED 훅 행이 더해진다
- **THEN** 주문 화면의 그 주문은 모호가 아니고 무장 행의 청산 근거(decision)를 보여 준다

#### Scenario: 거절 · 취소로 해제된 청산 주문
- **WHEN** 청산 발의의 무장 행 하나에 PROPOSAL_REFUSED 또는 PROPOSAL_CANCELLED 훅 행이 더해진다
- **THEN** 그 intent 의 주문은 모호가 아니고 무장 행의 청산 근거를 보여 준다

#### Scenario: 무장 행이 둘이다
- **WHEN** 한 intent 에 무장 행이 둘 있고 체결 훅 행도 있다
- **THEN** 그 주문은 지금처럼 `ambiguous_exit_evidence` 로 표시되고 근거 행을 고르지 않는다
