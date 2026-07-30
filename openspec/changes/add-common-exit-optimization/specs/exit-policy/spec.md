## ADDED Requirements

### Requirement: 포지션별 등록 ladder 평가
exit 관측 루프는 LADDER 포지션마다 저장된 policy ID로 등록 정책을 해석하고, observer 전역 공통표로 서로 다른 활성 포지션을 재해석해서는 안 된다 (SHALL NOT).

#### Scenario: 서로 다른 활성 정책
- **WHEN** 같은 계좌의 두 활성 포지션이 각각 BALANCED와 HYBRID_50 snapshot을 가진다
- **THEN** 각 포지션은 자신의 목표·보호선·부분익절 비율로 평가된다

#### Scenario: 저장된 정책이 등록되지 않음
- **WHEN** LADDER exit state의 non-empty policy ID가 현재 registry에 없다
- **THEN** 그 포지션의 판정과 주문 제출은 보류되고 운영 경고가 기록된다

### Requirement: 기존 청산 우선순위와 단조성 보존
공통 정책 평가는 기존 baseline breach의 위험 축소 우선순위, pending take-profit cancel-first, journal arm-before-submit, baseline/high-water 단조성 규칙을 그대로 적용해야 한다 (SHALL).

#### Scenario: rung 승격과 보호선 이탈 동시 발생
- **WHEN** 한 관측에서 high-water가 새 rung/trailing 보호선을 승격시키고 현재가가 그 보호선 아래다
- **THEN** 더 높은 보호선을 먼저 기록하고 부분익절보다 잔량 전량 보호 청산을 우선한다

#### Scenario: pending 부분익절 중 보호선 이탈
- **WHEN** 부분익절 proposal이 pending인 동안 보호선이 이탈된다
- **THEN** 청산을 보류하지 않고 기존 주문을 먼저 취소한 뒤 잔량 청산 proposal을 처리한다
