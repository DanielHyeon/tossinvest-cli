# lane-performance Specification

## Purpose

lane 성과 귀속의 계약이다. 어떤 거래를 어떤 lane/campaign 에 귀속하고, 무엇을 측정하지
않았다고 말하며, 파생 저장소가 거래 원장과 어떻게 격리되는지를 정의한다.

**a073 이 올린 문장은 좁힌 판본이다.** 아래 두 Requirement 의 규칙은 구현돼 있고
`internal/performance` 시험이 지키지만, 파생 저장소를 **채우는 것은 자동이 아니다**
(2026-09-08 측정): 투영은 `tossctl performance project-attribution` 이라는 별도 CLI
하위명령(`cmd/tossctl/root.go`)이고, 엔진 사이클은 그것을 부르지 않는다. 읽는 쪽
(`internal/httpapi/read.go`·`internal/console/performance_history.go`)은 배선돼 있다.
따라서 이 계약은 **투영된 것에 대해** 참이며, 투영을 언제 돌릴지는 운영자의 행위다.
## Requirements
### Requirement: lane 성과는 결정적 lineage만 사용한다

시스템은 market, candidate, lane/version, campaign/leg, decision, attempt, order, fill, position/close와 policy/version의 persisted identifier chain이 완전한 거래만 lane 성과에 귀속해야 한다 (SHALL). 같은 symbol 또는 ticker가 여러 market, lane 또는 campaign에 존재해도 symbol/time 근사로 누락 링크를 보정해서는 안 된다 (MUST NOT).

lineage 상태는 저장소 스키마가 강제한다: `attribution` 행의 `lineage_status`는
`complete` 또는 `link_missing` 만 허용하는 CHECK 아래에 있고, 제3의 값이나 침묵한
귀속은 쓸 수 없다.

#### Scenario: 완전한 lineage
- **WHEN** 하나의 closed trade가 market부터 campaign/leg와 close까지 전체 identifier chain을 가진다
- **THEN** 해당 market, lane/version, campaign/leg와 policy/version에 비용 후 결과를 귀속한다

#### Scenario: 링크 누락
- **WHEN** fill에서 decision으로 가는 결정적 링크가 없다
- **THEN** `link_missing`으로 집계하고 symbol/time 근사로 lane를 선택하지 않는다

#### Scenario: campaign identifier 누락
- **WHEN** lane/version은 있지만 campaign 또는 leg identifier가 없는 closed trade가 있다
- **THEN** attributed sample에서 제외하고 누락 identifier를 `link_missing` provenance로 기록한다

#### Scenario: 시장 간 동일 ticker
- **WHEN** KR과 US에 동일한 ticker 문자열의 거래가 존재한다
- **THEN** persisted market identifier로만 분리하고 한 시장의 결과를 다른 시장 lane/campaign에 귀속하지 않는다

### Requirement: partial fill과 staged close는 수량과 PnL을 보존한다

Projector는 deduplicated fill event의 signed quantity delta와 explicit correction/bust lineage만 적용해야 하고 cumulative order quantity를 다시 합산해서는 안 된다 (MUST NOT). 모든 시점에 acquired quantity는 attributed closed quantity와 authoritative residual position quantity의 합과 같아야 하며 (SHALL), staged close는 체결된 delta만 realized PnL로 옮기고 잔여 수량을 open으로 유지해야 한다 (SHALL).

각 close delta는 원 currency의 entry basis, exit proceeds, gross PnL, entry·exit fee, tax,
persisted FX cost와 net PnL을 보존해야 하며 (SHALL), `gross_pnl - entry_fees -
exit_fees - taxes - fx_cost = net_pnl` 보존식을 만족해야 한다 (SHALL). Fee 또는 FX
evidence가 누락되면 해당 metric은 `not_measured`이고 0, 조회 시점 환율 또는 다른 fill
값으로 대체해서는 안 된다 (MUST NOT) — source currency와 reporting currency가 같아도
그렇다.

#### Scenario: partial entry와 staged close
- **WHEN** entry가 세 fill로 체결되고 두 번의 staged close 뒤 residual quantity가 남는다
- **THEN** deduplicated fill delta 기준 closed quantity와 residual quantity의 합이 acquired quantity와 같고 각 close leg만 realized PnL에 포함된다

#### Scenario: duplicate cumulative update
- **WHEN** 같은 broker cumulative quantity update와 fill identity가 반복 수신된다
- **THEN** quantity, allocated basis, fee와 PnL은 한 번만 전진하고 conservation totals가 변하지 않는다

#### Scenario: fill correction
- **WHEN** broker bust/correction event가 원 fill identity와 signed negative delta를 포함한다
- **THEN** 해당 fill attribution만 역전하고 다른 market/campaign fill 또는 residual quantity를 symbol/time으로 조정하지 않는다

#### Scenario: FX 또는 fee evidence 누락
- **WHEN** US close의 FX as-of 또는 broker fee observation이 누락된다
- **THEN** affected reporting metric은 `not_measured`이고 누락 값을 0이나 조회 시점 환율로 보정하지 않는다

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
