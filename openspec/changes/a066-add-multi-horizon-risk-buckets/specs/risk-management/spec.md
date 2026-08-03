## ADDED Requirements

### Requirement: Guardian은 q_candidate를 보수적인 q_final로 제한한다

기존 Guardian 판정 체인의 mutation-free precheck가 exposure-raising 의도를 허용한 뒤 multi-horizon calculator는 canonical `q_candidate`와 monetary bucket으로 `q_final`을 확정해야 한다(SHALL). GuardianDecision은 q_final 확정 뒤에만 발급되어야 하며(SHALL), decision과 intent quantity는 q_final과 정확히 같고 q_candidate 및 기존 Guardian 허용 수량을 초과해서는 안 된다(MUST NOT). q_candidate, q_final, worst executable price, fee/FX haircut digest, 각 monetary bucket snapshot/version, binding cap, owner와 refusal reason은 GuardianDecision provenance에 포함되어야 한다(SHALL).

#### Scenario: Bucket이 strategy 수량을 축소
- **WHEN** q_candidate가 10주이고 유효한 q_final이 4주다
- **THEN** GuardianDecision과 intent quantity는 정확히 4주이며 q_candidate 10주와 binding monetary bucket provenance를 보존한다

#### Scenario: Bucket 계산 실패
- **WHEN** required bucket snapshot 또는 valuation을 fail-closed로 계산할 수 없다
- **THEN** Guardian은 exposure-raising decision과 reservation을 발급하지 않고 안정 reason-code를 기록한다

#### Scenario: q_final 전 decision 기록 시도
- **WHEN** q_final과 monetary reservation preimage가 확정되기 전에 GuardianDecision을 기록하려 한다
- **THEN** 발급은 거부되고 제출 가능한 decision과 reservation은 0건이다

### Requirement: Guardian 원자 예약은 lane owner와 모든 bucket을 포함한다

q_final 확정 뒤 결정 영속과 예약의 단일 journal transaction은 q_final GuardianDecision, lane owner 및 horizon, market, strategy, sector, symbol monetary bucket reservation을 함께 commit해야 한다(SHALL). 어느 owner 또는 bucket 검사라도 실패하면 GuardianDecision, 기존 exposure reservation, monetary bucket reservation과 campaign owner가 모두 rollback되어야 한다(SHALL). Gateway는 제출 전에 decision quantity가 q_final과 일치하고 모든 HELD monetary reservation 및 owner가 여전히 유효한지 journal에서 재검증해야 한다(SHALL).

#### Scenario: Sector cap 경쟁
- **WHEN** 동시에 허용된 두 decision 중 하나가 sector 잔여 한도를 먼저 예약한다
- **THEN** 다른 transaction은 LIMIT_REACHED로 전체 rollback되고 제출 가능한 고아 decision이나 부분 reservation이 없다

#### Scenario: 제출 전 owner 상실
- **WHEN** decision 발급 뒤 Gateway 제출 전에 owner 또는 bucket reservation이 유효하지 않게 된다
- **THEN** Gateway는 broker request 전에 제출을 거부한다

### Requirement: 다중 horizon 손실 lock은 entry-only다

short/medium horizon 및 market별 손실 lock은 해당 범위의 신규 EXPOSURE_RAISING decision과 추가 leg만 차단해야 한다(SHALL). lock은 RISK_REDUCING decision 발급, stop, emergency exit, reconciliation, fill detection 또는 기존 reservation 정리에 적용되어서는 안 된다(MUST NOT). 보수 방향 lock 활성화는 즉시 영속할 수 있지만 완화는 사람 승인과 audit를 요구해야 한다(SHALL).

#### Scenario: Medium lock과 short entry
- **WHEN** KR medium-horizon loss lock만 활성이고 독립 한도가 정상인 KR short-horizon 진입이 평가된다
- **THEN** medium lock은 short 요청에 전파되지 않으며 short 요청은 나머지 Guardian 체인을 계속 평가한다

#### Scenario: Market lock 중 emergency exit
- **WHEN** US entry loss lock이 활성인 상태에서 US emergency exit가 발급된다
- **THEN** RISK_REDUCING decision은 lock 평가 없이 진행되고 broker exit 경로가 지연되지 않는다

### Requirement: fill 사실은 risk overage 또는 valuation unknown보다 우선 보존된다

Guardian과 bucket projection은 이미 발생한 authoritative fill과 Position apply를 cap overage, missing actual price/fee/FX 또는 stale risk snapshot을 이유로 거부해서는 안 된다(MUST NOT). Fill transaction은 모든 적용 bucket의 proportional HELD transfer, actual monetary exposure 또는 explicit unknown, `max(transfer, actual)` filled amount 및 overage를 Position과 원자적으로 기록해야 한다(SHALL). `RISK_OVERAGE`와 `UNKNOWN_ACTUAL_RISK`는 신규 exposure만 차단하고 stop, emergency exit, reconciliation, fill detection과 Position apply를 차단해서는 안 된다(MUST NOT).

#### Scenario: unknown actual risk인 authoritative fill
- **WHEN** authoritative fill delta가 있지만 actual fee 또는 FX를 확정할 수 없다
- **THEN** Guardian은 fill/Position을 적용하고 UNKNOWN_ACTUAL_RISK로 신규 exposure만 차단하며 missing 값을 0으로 대체하지 않는다

#### Scenario: 모든 bucket의 overage 기록
- **WHEN** 하나의 fill 뒤 여러 적용 bucket usage가 동시에 cap을 초과한다
- **THEN** fill/Position은 한 번 보존되고 horizon, market, strategy, sector와 symbol 각각의 overage와 RISK_OVERAGE latch가 같은 transaction에 기록된다
