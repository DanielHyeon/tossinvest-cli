## ADDED Requirements

### Requirement: production runtime은 네 전략군의 여덟 lane instance를 독립 감독한다
Production strategy runtime은 continuation, reversal, weekly-value와 breakout-retest의 KR/US lane instance 정확히 8개를 `(market, family, lane_id, lane_version)` key로 감독해야 한다 (SHALL). 각 instance는 자체 cadence, bounded queue, deadline, health, consecutive-failure counter와 entry-only latch를 가져야 하며 (SHALL), 한 instance의 wait, timeout, panic, stale evidence 또는 latch가 peer instance의 evaluation cycle이나 state를 변경해서는 안 된다 (MUST NOT).

#### Scenario: KR breakout worker timeout
- **WHEN** KR breakout-retest worker가 cycle deadline을 초과한다
- **THEN** KR breakout failure counter만 증가하고 KR continuation/reversal/weekly-value와 모든 US worker는 자기 cadence로 계속 평가한다

#### Scenario: 정확한 worker cardinality
- **WHEN** canonical four-family production manifest로 runtime graph를 구성한다
- **THEN** 8개의 unique lane instance와 2개의 market coordinator가 생성되고 duplicate 또는 unknown instance는 0건이다

### Requirement: lane worker는 mutation 권위 없는 sealed proposal만 발행한다
각 lane worker는 immutable evidence snapshot을 pure evaluator에 전달하고 sealed proposal 또는 typed refusal만 발행해야 한다 (SHALL). Worker dependency closure에는 broker mutator, writable journal, Guardian issuer, activation/toggle writer가 없어야 하며 (MUST NOT), proposal seal은 family/lane/version, account/market/symbol/generation, candidate/setup, horizon, evidence/config digest, freshness, execution terms와 arbitration provenance를 포함해야 한다 (SHALL).

#### Scenario: mutation dependency 검사
- **WHEN** 8개 lane worker와 evaluator의 dependency closure를 정적으로 검사한다
- **THEN** broker mutation, writable journal, Guardian issuance와 operating-setting writer가 존재하지 않는다

#### Scenario: proposal seal 변조
- **WHEN** worker 출력의 lane version 또는 evidence digest가 seal 생성 뒤 변경된다
- **THEN** market coordinator가 proposal을 거부하고 Guardian 및 broker request는 0건이다

### Requirement: market coordinator는 owner와 calibrated score로 최대 한 proposal만 선택한다
KR/US market coordinator는 동일 `(account, market, symbol, position_generation)`의 current sealed proposals를 모아 active owner를 먼저 보존해야 한다 (SHALL). Family는 exact enum `{CONTINUATION, REVERSAL, WEEKLY_VALUE, BREAKOUT_RETEST}`이어야 한다 (SHALL). Active owner가 없을 때 eligible ON proposal은 동일한 approved arbitration score version/calibration digest의 integer `score_ppm` 범위 0..1,000,000에서만 비교해야 하며 (SHALL), highest unique proposal 최대 하나만 shared dispatch에 전달해야 한다 (SHALL). Singleton proposal도 approved score/calibration authority가 없으면 production dispatch에 전달해서는 안 된다 (MUST NOT). 최고점 tie, incomparable/uncalibrated score, multiple owner, stale owner revision은 fail-closed refusal이어야 한다 (SHALL).

#### Scenario: 세 short family가 같은 KR symbol을 제안
- **WHEN** owner가 없고 continuation, reversal과 breakout-retest가 같은 KR symbol에 comparable current scores를 제안한다
- **THEN** unique highest proposal 하나만 router/dispatch handoff로 전달되고 나머지는 arbitration lineage와 함께 선택되지 않는다

#### Scenario: cross-family score가 비교 불가능
- **WHEN** 두 eligible proposal의 score version 또는 calibration digest가 다르다
- **THEN** `ARBITRATION_UNCALIBRATED` refusal이 기록되고 Guardian 및 broker request는 0건이다

#### Scenario: singleton score가 uncalibrated
- **WHEN** 한 owner scope에 eligible proposal이 하나뿐이지만 approved score version 또는 calibration digest가 없다
- **THEN** production coordinator는 `ARBITRATION_UNCALIBRATED`로 거부하고 SHADOW counterfactual 외 dispatch handoff는 0건이다

#### Scenario: active weekly owner
- **WHEN** 같은 symbol/generation에 active weekly-value owner가 있고 breakout proposal이 더 높은 score를 제시한다
- **THEN** 기존 weekly-value owner만 유지되고 breakout proposal은 owner를 교체하거나 새 campaign을 만들지 않는다

### Requirement: coordinator intake는 owner-scope별 bounded queue를 사용한다
Coordinator intake는 `(account,market,symbol,position_generation,family,lane_id,lane_version,snapshot_digest)` exact dedup key를 사용해야 하며 (SHALL), server-owned positive finite capacity, latest-per-key coalescing과 deterministic owner-scope ordering을 적용해야 한다 (SHALL). 서로 다른 symbol/owner scope를 market-wide proposal 하나로 접어서는 안 되며 (MUST NOT), overflow 또는 drop은 typed refusal과 bounded counter로 관측되어야 하고 active-owner scope를 silent eviction해서는 안 된다 (MUST NOT).

#### Scenario: market 안의 두 owner scope
- **WHEN** KR의 두 symbol owner scope가 동시에 current proposal을 가진다
- **THEN** 두 scope는 독립 arbitration되고 각 scope마다 최대 한 selected proposal이 bounded handoff를 기다린다

#### Scenario: same-key burst
- **WHEN** 같은 dedup key의 snapshot revision이 capacity를 넘는 속도로 도착한다
- **THEN** newest current envelope 하나로 coalesce되고 drop count가 증가하며 다른 owner scope와 active-owner record는 손실되지 않는다

### Requirement: 모든 전략군은 하나의 owner risk Guardian dispatch 권위를 공유한다
Runtime은 모든 family에 대해 horizon/family가 없는 symbol owner key, account당 하나의 Guardian, 하나의 account-wide exposure/loss domain, 하나의 journal/dispatch owner와 official ExecutionGateway를 사용해야 한다 (SHALL). Family/horizon/market risk bucket은 a066의 `q_final` 교집합에 참여하되 physical account capacity를 복제해서는 안 되며 (MUST NOT), 어떤 lane도 `q_final > q_candidate`를 만들 수 없다 (MUST NOT).

#### Scenario: 두 family의 동시 owner 획득
- **WHEN** 같은 owner scope의 두 selected proposal이 journal admission을 동시에 시도한다
- **THEN** atomic owner/q_final transaction 하나만 성공하고 다른 proposal은 broker request 전 conflict로 전체 rollback된다

#### Scenario: 한 family risk bucket 고갈
- **WHEN** breakout family bucket cap이 0이고 continuation family 및 account-wide cap은 유효하다
- **THEN** breakout entry만 q_final 0/refusal이고 continuation은 나머지 공용 Guardian chain을 계속 평가하며 account cap은 복제되지 않는다

### Requirement: entry worker 장애는 safety lifecycle을 지연하지 않는다
Lane OFF, market close, evidence failure, worker queue pressure와 low-priority budget exhaustion은 exposure-raising evaluation만 대기·거부해야 한다 (SHALL). Fill detection, reconciliation, broker-resident protection supervision, exit observation과 emergency reduction은 별도 context/queue와 reserved API budget으로 계속되어야 하며 (SHALL), lane worker가 safety loop를 취소하거나 safety reserve를 소비해서는 안 된다 (MUST NOT).

#### Scenario: 모든 entry worker latch OFF
- **WHEN** 8개 entry worker가 모두 failure latch OFF다
- **THEN** 신규 proposal/entry는 0건이지만 fill, reconcile, protection, exit와 emergency reduction cycle은 기존 cadence로 계속된다

#### Scenario: physical API allowance 고갈
- **WHEN** lane evidence polling이 low-priority allowance를 모두 commitment로 보유한다
- **THEN** 추가 lane poll은 deferred되고 safety-class 호출의 reserved budget은 감소하지 않는다

### Requirement: 새 runtime은 OFF 기본값과 first-leg 경계를 유지한다
모든 8개 descriptor, worker desired/effective state, automation과 autostart는 새 설치·migration·restart에서 기본 OFF 또는 UNOBSERVED여야 한다 (SHALL). Legacy 3-family approval은 4-family activation으로 자동 승격되어서는 안 되며 (MUST NOT), breakout v1은 a066 scale-in 완료와 별도 승인 전 production first-leg proposal 이외의 추가 exposure leg를 만들지 않아야 한다 (MUST NOT). ProtectionReady 또는 다른 required activation authority가 missing이면 broker exposure-raising request는 0건이어야 한다 (SHALL).

#### Scenario: legacy three-family restart
- **WHEN** 기존 3-family manifest가 저장된 상태에서 4-family runtime binary가 시작된다
- **THEN** breakout과 새 runtime activation은 OFF이고 기존 approval을 복제하거나 확대하지 않는다

#### Scenario: protection readiness missing
- **WHEN** lane proposal, score와 Guardian 입력은 유효하지만 current market ProtectionReady가 UNWIRED다
- **THEN** exposure-raising broker request는 0건이고 reduce-only safety lifecycle은 계속된다

### Requirement: SHADOW는 OFF state를 승격하지 않는 read-only runtime이다
새 설치·migration·restart는 `desired=OFF`, `effective=OFF`, `runtime=UNOBSERVED`를 유지해야 한다 (SHALL). `SHADOW`는 server-owned signed shadow manifest가 있을 때만 pure evaluation과 counterfactual projection을 허용하며 (SHALL), desired/effective/activation을 ON으로 만들거나 dispatch capability를 소유해서는 안 된다 (MUST NOT). Process-local shadow 상태를 restart에서 자동 복구해서는 안 된다 (MUST NOT).

#### Scenario: signed shadow manifest가 없는 restart
- **WHEN** 이전 process가 SHADOW를 관측한 뒤 signed shadow manifest 없이 restart한다
- **THEN** 모든 lane는 OFF/OFF/UNOBSERVED이고 proposal dispatch 또는 activation write는 0건이다

### Requirement: runtime lineage와 health는 lane 단위로 결정적으로 관측된다
Runtime은 worker key, cycle generation, snapshot/evidence/config digest, queue depth/drop count, deadline, health, first refusal, arbitration outcome와 selected dispatch lineage를 bounded-cardinality status로 노출해야 한다 (SHALL). 기존 market-level projection fields를 보존하고 fixed deterministic order의 additive `lanes[8]`와 `coordinators[2]` child collections를 제공해야 하며 (SHALL), older readers가 additive unknown fields를 무시할 수 있어야 한다 (SHALL). 이 surface는 read-only이며 activation/mutation endpoint를 추가해서는 안 된다 (MUST NOT). Setup/candidate/symbol처럼 unbounded 값은 metric label로 사용해서는 안 되며 (MUST NOT), journal attribution은 candidate부터 close까지 exact identifier chain을 유지해야 한다 (SHALL).

#### Scenario: lane failure attribution
- **WHEN** US reversal은 stale evidence로 거부되고 US breakout은 proposal을 발행한다
- **THEN** 두 worker의 독립 health/refusal과 breakout arbitration lineage가 symbol/time 추정 없이 조회된다

#### Scenario: legacy projection reader
- **WHEN** 8-lane runtime response를 legacy market-level reader가 조회한다
- **THEN** 기존 fields의 의미와 형식은 유지되고 additive lane/coordinator fields를 무시해도 read가 실패하지 않는다
