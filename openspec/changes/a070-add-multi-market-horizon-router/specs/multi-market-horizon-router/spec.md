## ADDED Requirements

### Requirement: router는 position generation의 owning lane을 하나로 결정한다
Router는 `(account, market, canonical_symbol, position_generation)` ownership key마다 최대 하나의 owning lane 또는 typed refusal을 반환해야 한다 (SHALL).
Horizon은 lane eligibility, risk admission과 attribution에 포함되어야 하지만 ownership key에
포함되어서는 안 된다 (MUST NOT). Router는 해당 ownership key의 모든 horizon active owner를
조회해야 하고 (SHALL), durable open campaign owner가 하나 있으면 새 score나 다른 horizon으로
소유권을 바꿔서는 안 된다 (MUST NOT). Owner가 여러 개이거나 account/market/symbol/generation,
horizon 또는 eligibility가 missing/ambiguous/conflicting이면 fail closed해야 한다 (SHALL).

#### Scenario: cross-horizon 기존 owner
- **WHEN** short lane이 ownership key의 active owner이고 같은 key에 weekly lane candidate가 더 높은 score로 도착한다
- **THEN** router는 short owner를 보존하고 weekly 신규 owner/decision/reservation을 만들지 않으며 horizon을 바꿔 owner lookup을 우회하지 않는다

#### Scenario: 모든 horizon에서 중복 owner 발견
- **WHEN** 같은 ownership key에 short와 weekly active owner가 모두 발견된다
- **THEN** RECONSTRUCTION_MISMATCH refusal을 반환하고 어떤 lane도 선택하거나 owner를 자동 교체하지 않는다

#### Scenario: 새 position generation
- **WHEN** 이전 generation campaign이 CLOSED이고 새 position_generation에 owner가 없다
- **THEN** router는 새 generation의 eligible lane을 평가할 수 있지만 이전 generation owner를 재사용하지 않는다

#### Scenario: 두 lane이 같은 우선순위다
- **WHEN** owner가 없고 같은 scope에서 두 eligible lane이 동률이며 versioned tie-break가 없다
- **THEN** ambiguity refusal을 반환하고 owning lane과 entry decision은 0건이다

### Requirement: KR와 US routing lifecycle은 독립이다
Router는 KR와 US의 calendar, activation, evidence와 scheduler revision을 별도 lifecycle로 평가해야 한다 (SHALL).
한 시장의 disabled, OFF, closed, stale, CAS conflict 또는 failed 상태를 다른 시장 eligibility gate로
사용해서는 안 되고 (MUST NOT), KR 운영 안정화를 US routing 조건으로 삼거나 그 역을 수행해서도
안 된다 (MUST NOT). Router는 결합 KR+US activation manifest를 요구하거나 생성해서는 안 된다
(MUST NOT).

#### Scenario: KR disabled와 US eligible
- **WHEN** KR state가 disabled이고 US calendar, activation과 evidence는 유효하다
- **THEN** KR 신규 routing은 0건이고 US candidate는 KR revision/lock 상태와 무관하게 평가된다

#### Scenario: US calendar stale와 KR eligible
- **WHEN** US calendar가 WAIT_MARKET이고 KR calendar/activation은 유효하다
- **THEN** US 신규 routing은 0건이며 KR candidate routing은 계속된다

### Requirement: router는 순수하고 OFF 상태에서 mutation을 만들지 않는다
Router는 routing decision 또는 typed refusal만 반환해야 한다 (SHALL).
Broker, journal writer, campaign/owner state, activation manifest와 운영 토글을 직접 변경해서는 안
된다 (MUST NOT). 선택 대상 lane이 OFF이면 신규 entry routing과 buy mutation은 0건이어야 하며
(SHALL), common exit, reconciliation, fill detection과 protection supervision을 막아서는 안 된다
(MUST NOT).

#### Scenario: 선택 대상 lane이 OFF다
- **WHEN** existing owner 또는 deterministic selection 대상 lane이 desired/effective OFF다
- **THEN** typed disabled refusal을 반환하고 신규 entry와 mutation은 0건이며 common safety loops는 계속된다

#### Scenario: 정상 routing decision
- **WHEN** candidate scope, owner snapshot, horizon, calendar, activation, evidence와 lane eligibility가 모두 유효하다
- **THEN** account, market, canonical symbol, position generation, horizon, owning lane ID/version과 input digests를 가진 하나의 routing decision만 반환되고 외부 mutation은 0건이다
