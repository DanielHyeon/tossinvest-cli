## ADDED Requirements

### Requirement: strategy dispatch lease는 모든 안전 권한을 fenced 제출 직전에 재검증한다
Engine profile은 strategy dispatch lease의 모든 안전 권한과 owner fence를 제출 직전에 재검증해야 한다 (SHALL). Exposure-raising strategy attempt마다 candidate/evidence,
router/lane/version, campaign/leg, activation/calendar generations, exact `WIRED` ProtectionReady
attestation, reconciliation, risk reservation, Guardian decision/generation, build digest와
monotonic owner epoch/fencing token을 하나의 durable lease에 결합해야 한다 (SHALL).
ExecutionGateway 직전 검증은 journal의 current authority만 사용해야 하며 (SHALL), caller가
제공한 복제 상태나 생성 시점 검증으로 대체해서는 안 된다 (MUST NOT). Claim/validation은
lease를 비가역 소비해야 한다 (SHALL). Current `ISSUED` lease의 authority 누락·변경·만료,
stale epoch/token, scope mismatch와 pre-transport cancel은 lease/attempt `REFUSED`와 그 lease의
exact reservation `RELEASED`를 같은 journal transaction에서 영속하고 broker request를 0건으로
만들어야 한다 (SHALL). 이미 소비된 terminal lease replay는 retry attempt만 `REFUSED`하고 원래
lease/disposition을 변경하지 않으며, retry attempt의 별도 exact HELD reservation만 release해야
한다 (SHALL). 이
pre-transport failure들에 `AMBIGUOUS` 또는 `HELD`를 사용해서는 안 된다 (MUST NOT).

#### Scenario: activation manifest drift
- **WHEN** lane decision 뒤 dispatch 전에 해당 시장 activation manifest digest 또는 generation이 바뀐다
- **THEN** lease/attempt `REFUSED`와 exact reservation `RELEASED`를 원자 기록하고 broker request는 0건이며 effective entry를 OFF로 낮춘다

#### Scenario: 다른 시장 lease
- **WHEN** KR decision을 US calendar 또는 ProtectionReady generation에 결합된 lease로 제출한다
- **THEN** scope 불일치로 lease와 exact reservation을 `REFUSED + RELEASED` 처리하고 broker 호출 전에 거부한다

#### Scenario: durable lease 없는 제출
- **WHEN** GuardianDecision은 있지만 strategy dispatch lease가 없는 exposure-raising 제출을 시도한다
- **THEN** Engine profile과 Gateway는 typed refused attempt를 영속하고 해당 attempt에 결합된 exact reservation이 있으면 원자 RELEASED하며 broker request와 합성 lease는 0건이다

#### Scenario: validation failure 뒤 원상 복구
- **WHEN** 한 validation이 generation mismatch로 실패한 뒤 current 값이 lease preimage 값으로 돌아온다
- **THEN** terminal lease는 부활하지 않고 fresh decision과 fresh lease만 새 claim을 허용한다

#### Scenario: 만료 또는 stale fence
- **WHEN** lease가 만료됐거나 owner epoch/fencing token이 current durable state보다 stale이다
- **THEN** `REFUSED + RELEASED`를 원자 기록하고 broker request는 0건이며 `AMBIGUOUS`로 분류하지 않는다

### Requirement: market worker 장애는 그 시장 entry만 격리한다
Engine supervisor는 market worker 장애를 그 시장 entry scope에 격리해야 한다 (SHALL). KR 또는
US entry worker의 OFF, market wait, stale evidence, budget defer, cycle failure, panic, abnormal
return, watchdog expiry와 반복 crash를 해당 시장의 effective entry OFF latch와 bounded restart로
한정해야 한다 (SHALL). Peer market evaluation과 Reconcile driver, fill detector, protection
supervisor, exit observer, emergency reduction loop는 계속 실행해야 한다 (SHALL). Market worker
장애만으로 전체 process 또는 peer market을 종료해서는 안 된다 (MUST NOT).

#### Scenario: US entry worker abnormal return
- **WHEN** US entry worker가 비정상 반환하지만 safety loop와 KR worker는 정상이다
- **THEN** US entry만 typed OFF latch로 강화하고 bounded restart하며 KR evaluation과 모든 safety loop를 유지한다

#### Scenario: automation OFF
- **WHEN** automation effective state가 OFF로 전환된다
- **THEN** KR·US 신규 entry와 scale-in은 0건이고 fill, reconciliation, protection과 reduce-only exit는 계속된다

### Requirement: central integrity fault는 외부 fenced safety fallback으로 복구된다
Engine deployment는 central integrity fault를 외부 fenced safety fallback으로 복구해야 한다 (SHALL). Journal corruption, Gateway invariant violation, owner epoch/fence CAS 불능 또는 복수
current owner가 감지되면 모든 신규 entry를 즉시 차단하고 critical alert를 발행해야 한다
(SHALL). 별도 deployment domain의 external supervisor는 이전 owner token을 fence한 새 epoch로
entry capability가 없는 safety-only fallback을 versioned `safety_fallback_rto` 안에 기동해야
하며 (SHALL), 그 RTO는 60초를 초과해서는 안 된다 (MUST NOT). Fallback은
fill/reconciliation/protection/reduce-only exit/emergency reduction만 수행하고 entry lease를
발급해서는 안 된다 (MUST NOT).

#### Scenario: central dispatch owner integrity 상실
- **WHEN** current owner fence가 손상되거나 두 owner가 current라고 주장한다
- **THEN** 모든 entry를 차단하고 stale token을 broker 전에 거부하며 external supervisor가 60초 이하의 frozen RTO 안에 fenced safety-only fallback을 시작한다

#### Scenario: fallback 기동 실패
- **WHEN** external supervisor가 frozen RTO 안에 safety-only fallback을 기동하지 못한다
- **THEN** broker-resident protection을 자동 취소하지 않고 `SAFETY_FALLBACK_UNAVAILABLE` critical state를 지속 발행하며 신규 entry는 0건이다
