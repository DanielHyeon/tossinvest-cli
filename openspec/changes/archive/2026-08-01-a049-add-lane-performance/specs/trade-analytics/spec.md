## ADDED Requirements

### Requirement: 기존 outcome 집계는 lane와 policy version으로 분해 가능하다
trade analytics는 기존 전체 승률·PF·MDD·R을 보존하면서 결정적 lineage가 있는 결과를 lane/version과 policy/version으로 필터·집계할 수 있어야 한다 (SHALL).

#### Scenario: 전체 집계 보존
- **WHEN** lane performance 기능을 활성화한 뒤 동일 outcome 집합을 전체 집계한다
- **THEN** 기존 portfolio aggregate 결과는 변경되지 않는다

#### Scenario: 표본 부족
- **WHEN** 한 lane의 유효 closed trade가 최소 표본에 못 미친다
- **THEN** 관측값과 표본 수는 표시하되 추천 가능 상태로 해석하지 않는다
