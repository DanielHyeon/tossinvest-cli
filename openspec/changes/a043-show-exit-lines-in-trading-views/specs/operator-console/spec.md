## ADDED Requirements

### Requirement: positions는 현재와 다음 exit 동작을 함께 표시한다
콘솔은 `/positions`에서 entry, initial stop, current protection, next target, next protection, rung, projected exit quantity와 evaluated-at을 권위 snapshot 그대로 표시해야 한다 (SHALL).

#### Scenario: 관리 중 포지션
- **WHEN** 완전한 exit snapshot이 있는 포지션을 조회한다
- **THEN** 운영자는 현재 보호와 다음 가격 도달 시 기준선·수량 동작을 한 화면에서 읽을 수 있다

#### Scenario: 1주 포지션
- **WHEN** 보유 수량이 1주이고 다음 rung이 partial이다
- **THEN** 화면은 `중간 매도 없음 · 보호선 승격`과 최종/손절 시 1주 전량을 표시한다

#### Scenario: stale snapshot
- **WHEN** snapshot이 stale 또는 일부 unknown이다
- **THEN** 값은 0이 아니라 `—`와 stale/unknown 사유로 표시된다

### Requirement: orders는 exit 주문 근거를 결정적으로 연결한다
콘솔은 broker order의 명시적 mutation-attempt intent lineage로 exit event의 decision ID와 기준선 snapshot을 연결하고 trigger line, observation, policy, rung과 reason을 표시해야 한다 (SHALL).

#### Scenario: 연결된 손절 주문
- **WHEN** broker order의 mutation attempt intent가 protection breach exit event를 참조한다
- **THEN** 주문 상세는 당시 현재가와 보호선, 전량 사유를 표시한다

#### Scenario: 연결 식별자 부재
- **WHEN** broker order에 결정적 snapshot 링크가 없다
- **THEN** 화면은 근거 미연결로 표시하고 symbol/time으로 추정하지 않는다

### Requirement: 거래 화면과 설정 화면의 역할은 분리된다
`/positions`와 `/orders`는 exit 상태를 읽기 전용으로 설명해야 하며 (SHALL), 정책 설정 control을 복제해서는 안 된다 (MUST NOT). 설정이 필요한 문맥에는 a050의 canonical category deep link를 제공해야 한다 (SHALL).

#### Scenario: positions에서 종목 정책 확인
- **WHEN** 운영자가 포지션의 정책을 변경하려 한다
- **THEN** 화면은 `/optimization?category=position-management` 링크를 제공하고 현재 표 안에서 즉시 변경하지 않는다

#### Scenario: orders에서 보호주문 확인
- **WHEN** 운영자가 exit 주문의 보호 설정을 확인하려 한다
- **THEN** 화면은 `/optimization?category=exit-protection` 링크를 제공한다

#### Scenario: 입력 없는 거래 화면
- **WHEN** 운영자가 `/positions` 또는 `/orders`를 연다
- **THEN** 화면에는 form/input/textarea/select/button/contenteditable이 없고 해당 경로의 POST는 405로 거부된다
