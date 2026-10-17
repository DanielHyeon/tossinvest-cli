## ADDED Requirements

### Requirement: entry polling capability는 market horizon family subscope에 결합된다
Scheduler는 low-priority strategy capability를 `(market, horizon, family, poll_class)` admission subscope에 결합해야 하며 (SHALL), continuation, reversal, weekly-value와 breakout-retest 사이에서 capability를 재사용하거나 replay해서는 안 된다 (MUST NOT). 모든 subscope는 같은 physical endpoint/reset-generation의 reported remaining, safety reserve, commitment set, issuance cap과 observation-cycle authority를 공유해야 하며 (SHALL), family 추가로 provider capacity를 복제하거나 곱해서는 안 된다 (MUST NOT).

#### Scenario: continuation capability replay
- **WHEN** KR SHORT continuation에 발급된 capability로 KR SHORT breakout poll을 완료하려 한다
- **THEN** family scope mismatch로 거부되고 shared commitment/capacity는 변하지 않는다

#### Scenario: 마지막 physical allowance 경쟁
- **WHEN** 8개 lane worker가 같은 endpoint/reset-generation의 마지막 low-priority slot을 동시에 acquire한다
- **THEN** shared authority에서 하나만 commit되고 나머지는 BUDGET_DEFERRED이며 safety reserve는 유지된다

### Requirement: lane cadence와 failure state는 peer lane 및 safety cadence와 독립이다
Scheduler는 각 canonical lane instance에 versioned cadence/deadline/backoff와 entry-only failure state를 적용해야 하며 (SHALL), 한 lane의 deadline, stale evidence, closed market, queue pressure 또는 repeated failure가 같은 시장/다른 시장 peer lane의 scheduler record를 변경해서는 안 된다 (MUST NOT). Entry/candidate cadence는 exit, fill detection, reconciliation과 protection supervision보다 낮은 priority로 동작해야 한다 (SHALL).

#### Scenario: US weekly evidence stale
- **WHEN** US weekly-value evidence만 stale이고 US regular session과 breakout evidence는 current다
- **THEN** US weekly cycle만 refusal/deferred되고 US breakout과 safety-class cadence는 계속될 수 있다

#### Scenario: lane failure latch recovery
- **WHEN** KR reversal failure latch가 설정된 뒤 fresh authority와 versioned recovery 조건이 충족되지 않았다
- **THEN** KR reversal entry만 OFF를 유지하고 restart 또는 peer lane 성공이 latch를 자동 해제하지 않는다
