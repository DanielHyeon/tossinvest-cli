# strategy-runtime — a127 delta

## ADDED Requirements

### Requirement: 생산 전략 권한 적재기는 엔진이 연 현재 원장 스키마만 읽는다

생산 전략 권한 적재기(위험 snapshot 권한 · route 권한)는 원장 `user_version` 이 engine 이 주입한 현재 journal 스키마 버전과 정확히 같을 때만 원장을 판독해야 한다 (SHALL). 판독 조건은 마이그레이션과 함께 움직이지 않는 고정 리터럴이어서는
안 되며(MUST NOT), 주입 값이 없거나 0 이하이면 원장을 열기 전에 거절해야 한다(SHALL). 더 새 원장과 더 옛 원장은 둘 다 거절하되 거절 사유가
방향을 구분해야 하고(SHALL), 거절은 범위 국소 거절이 아니라 그 시장 권한의 결함으로 남아야 한다(SHALL). 적재기의 수락은 `journal.Open` 으로 연
현재 원장으로 시험되어야 한다(SHALL).

#### Scenario: 엔진이 연 현재 원장

- **WHEN** engine 이 `journal.Open` 으로 연(마이그레이션된) 원장 경로와 현재 스키마 버전을 두 적재기에 넘긴다
- **THEN** 스키마 조건은 통과하고 적재기는 기존 판독 · 판정을 그대로 수행한다

#### Scenario: 이 빌드보다 새 원장

- **WHEN** 원장 `user_version` 이 주입된 현재 버전보다 크다
- **THEN** 적재기는 원장 행을 판독하지 않고 더 새 원장임을 말하는 사유로 거절하며 exposure-raising 요청은 0건이다

#### Scenario: 마이그레이션되지 않은 원장

- **WHEN** 원장 `user_version` 이 주입된 현재 버전보다 작다
- **THEN** 표와 열이 모두 있더라도 적재기는 더 옛 원장임을 말하는 사유로 거절하며 exposure-raising 요청은 0건이다

#### Scenario: 주입 누락

- **WHEN** 호출자가 현재 스키마 버전을 넘기지 않았거나 0 이하를 넘긴다
- **THEN** 적재기는 원장 파일을 열기 전에 거절한다

#### Scenario: 읽기 집합의 열이 없는 원장

- **WHEN** 버전은 같지만 적재기가 읽는 열 하나가 원장에 없다(그 열을 읽는 질의가 해당 범위에서 조건부로만 실행되는 경우 포함)
- **THEN** 적재기는 판독 전에 실패를 드러내 fail-closed 로 거절하고 권한을 만들지 않는다
