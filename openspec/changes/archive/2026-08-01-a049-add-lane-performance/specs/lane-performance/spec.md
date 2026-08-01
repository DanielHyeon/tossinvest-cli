## ADDED Requirements

### Requirement: lane 성과는 결정적 lineage만 사용한다
시스템은 candidate, lane, decision, attempt, order, fill, position과 close의 persisted identifier chain이 완전한 거래만 lane 성과에 귀속해야 한다 (SHALL).

#### Scenario: 완전한 lineage
- **WHEN** 하나의 closed trade가 전체 identifier chain을 가진다
- **THEN** 해당 lane/version과 policy/version에 비용 후 결과를 귀속한다

#### Scenario: 링크 누락
- **WHEN** fill에서 decision으로 가는 결정적 링크가 없다
- **THEN** `link_missing`으로 집계하고 symbol/time 근사로 lane를 선택하지 않는다

### Requirement: 시계열 성과는 측정 상태를 구분한다
시스템은 5/15/30분 markout, slippage와 가능한 MFE/MAE를 계산하고 데이터가 없을 때 `not_measured`를 반환해야 한다 (SHALL).
markout은 각 target 이후 첫 기존 관측을 최대 60초 tolerance 안에서 선택해야 하며 (SHALL), 이 change가 추가 quote polling을 만들어서는 안 된다 (MUST NOT).

#### Scenario: markout 관측 완료
- **WHEN** entry 뒤 세 window의 유효 가격 관측이 있다
- **THEN** 각 window의 비용 전·후 markout과 관측 source/time을 저장한다

#### Scenario: 관측 누락
- **WHEN** 15분 관측이 없다
- **THEN** 15분 값은 0이 아니라 `not_measured`다

### Requirement: derived 성과 저장소는 거래 원장과 격리되고 bounded pruning을 사용한다
high-volume observation은 별도 `performance.db`에 저장하고 raw row를 90일 보존해야 한다 (SHALL). pruning은 24시간마다 최대 500 rows/transaction이어야 하며 (SHALL), authoritative journal lineage/outcome/audit을 삭제해서는 안 된다 (MUST NOT).

#### Scenario: 대규모 retention
- **WHEN** 1,000,000 raw row fixture에서 90일 초과 row를 정리한다
- **THEN** 각 prune transaction은 500 row 이하이고 100ms lock 목표를 검증하며 최근 30일 query p95는 250ms 이하 목표를 검증한다

### Requirement: 성과 수집은 거래 권한이 없다
performance collector와 query는 order mutation, config write, lane toggle 또는 LIVE approval capability를 가져서는 안 된다 (MUST NOT).

#### Scenario: dependency 검사
- **WHEN** performance package의 dependency closure를 검사한다
- **THEN** broker mutation과 operating-setting writer가 존재하지 않는다
