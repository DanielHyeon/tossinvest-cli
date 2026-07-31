## ADDED Requirements

### Requirement: strategy provenance는 주문까지 끊기지 않는다
entry decision, RiskIntent, mutation attempt, broker order와 fill은 candidate-life ID, threshold set/evidence digest와 lane ID/version을 결정적으로 연결해야 한다 (SHALL). 전략 provenance가 없는 legacy RiskIntent의 canonical bytes는 바뀌어서는 안 된다 (MUST NOT).

#### Scenario: 정상 체결
- **WHEN** strategy order가 체결된다
- **THEN** fill과 열린 position에서 원 candidate와 lane version을 역추적할 수 있다

#### Scenario: 재시작 중 중복 decision
- **WHEN** 같은 canonical decision이 재시작 뒤 다시 계산된다
- **THEN** deterministic identity와 duplicate guard가 두 번째 LIVE order를 차단한다
