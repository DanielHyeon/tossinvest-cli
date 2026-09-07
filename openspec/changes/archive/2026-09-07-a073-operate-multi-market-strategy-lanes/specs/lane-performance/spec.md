## MODIFIED Requirements

### Requirement: lane 성과는 결정적 lineage만 사용한다
시스템은 lane 성과에 결정적 lineage만 사용해야 한다 (SHALL). Market, candidate, lane/version,
campaign/leg, decision, attempt, order, fill, position/close와 policy/version의 persisted identifier
chain이 완전한 거래만 lane 성과에 귀속해야 한다 (SHALL). 같은 symbol 또는 ticker가 여러
market, lane 또는 campaign에 존재해도 symbol/time 근사로 누락 링크를 보정해서는 안 된다
(MUST NOT).

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
시스템은 partial fill과 staged close의 수량과 PnL을 보존해야 한다 (SHALL). Projector는
deduplicated fill event의 signed quantity delta와 explicit correction/bust lineage만 적용하고
cumulative order quantity를 다시 합산해서는 안 된다 (MUST NOT). Partial entry/exit fill과 각
close leg를 persisted identity로 분리하고 authoritative journal의 position cost-basis
policy/version 및 rounding rule로만 entry basis를 close delta에 배분해야 한다 (SHALL).

모든 시점에 acquired quantity는 attributed closed quantity와 authoritative residual position
quantity의 합과 같아야 하며 (SHALL), allocated entry basis와 close quantity의 합은 authoritative
position totals를 초과하거나 소실해서는 안 된다 (MUST NOT). Staged close는 체결된 delta만
realized PnL로 이동하고 잔여 수량을 open으로 유지해야 한다 (SHALL).

각 close delta는 원 currency의 entry basis, exit proceeds, gross PnL, broker/exchange entry·exit
fees, tax, persisted FX cost와 net PnL을 보존해야 한다 (SHALL). Reporting currency metric은
persisted FX source, rate, as-of, quote currency와 rounding version을 요구해야 하고 (SHALL),
`gross_pnl - entry_fees - exit_fees - taxes - fx_cost = net_pnl` 보존식을 만족해야 한다
(SHALL). Fee 또는 FX evidence가 누락되면 해당 metric을 `not_measured`로 분리하고 0, current FX
또는 다른 fill 값으로 대체해서는 안 된다 (MUST NOT).

#### Scenario: partial entry와 staged close
- **WHEN** entry가 세 fill로 체결되고 두 번의 staged close 뒤 residual quantity가 남는다
- **THEN** deduplicated fill delta 기준 closed quantity와 residual quantity의 합이 acquired quantity와 같고 각 close leg만 realized PnL에 포함된다

#### Scenario: duplicate cumulative update
- **WHEN** 같은 broker cumulative quantity update와 fill identity가 반복 수신된다
- **THEN** quantity, allocated basis, fees와 PnL은 한 번만 전진하고 conservation totals가 변하지 않는다

#### Scenario: fill correction
- **WHEN** broker bust/correction event가 원 fill identity와 signed negative delta를 포함한다
- **THEN** 해당 fill attribution만 역전하고 다른 market/campaign fill 또는 residual quantity를 symbol/time으로 조정하지 않는다

#### Scenario: 비용 후 손익 보존
- **WHEN** close delta에 entry/exit fees, tax와 persisted FX cost가 모두 있다
- **THEN** 원 currency와 reporting currency에서 gross에서 모든 비용을 뺀 값이 net PnL과 일치하고 rounding residual은 policy version대로 명시된다

#### Scenario: FX 또는 fee evidence 누락
- **WHEN** US close의 FX as-of 또는 broker fee observation이 누락된다
- **THEN** affected reporting metric은 `not_measured`이고 누락 값을 0이나 조회 시점 환율로 보정하지 않는다
