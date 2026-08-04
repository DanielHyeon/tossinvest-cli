## ADDED Requirements

### Requirement: 모든 exposure-raising 요청은 다차원 bucket에 귀속된다

모든 자동 exposure-raising 요청은 account, horizon, market, server-owned strategy risk ID/version, sector와 symbol bucket key를 가져야 하며(SHALL), 각 bucket은 account base currency의 monetary limit, filled monetary exposure, HELD monetary reservation, valuation source/time 및 snapshot version을 제공해야 한다(SHALL). unknown dimension, caller-invented strategy, stale valuation, 누락 worst executable price/fee/FX 또는 해소되지 않은 currency는 무제한이나 0원 노출로 간주해서는 안 되며(MUST NOT) typed refusal로 fail closed해야 한다(SHALL).

#### Scenario: sector 분류 부재
- **WHEN** exposure-raising symbol의 authoritative sector를 evaluation cutoff에 결정할 수 없다
- **THEN** UNKNOWN_SECTOR로 거부되고 공용 무제한 sector bucket에 배치되지 않는다

#### Scenario: US 환산 근거 부재
- **WHEN** account limit currency와 US 요청 currency 사이의 authoritative fresh conversion이 없다
- **THEN** 최종 수량은 0이고 CURRENCY_UNRESOLVED가 기록된다

#### Scenario: Strategy bucket 부재
- **WHEN** lane/version에 대응하는 server-owned strategy risk bucket을 결정할 수 없다
- **THEN** UNKNOWN_STRATEGY_BUCKET으로 거부되고 market 또는 symbol bucket만으로 진입을 허용하지 않는다

### Requirement: Monetary reservation은 보수적 체결 비용을 사용한다

각 bucket의 HELD reservation은 `ceil_minor(q × worst_executable_price_quote × fresh_fx_rate × fx_haircut_multiplier + worst_case_fees_base(q))`로 계산되어야 한다(SHALL). worst executable price, fee policy, FX source/rate/haircut, observed-at과 digest를 보존해야 하며(SHALL), haircut은 1 미만일 수 없다(MUST NOT). missing/stale/non-positive input, fee/FX fallback, arithmetic overflow 또는 minor-unit 변환 실패는 `q_final=0` typed refusal이어야 한다(SHALL).

#### Scenario: Limit보다 낮은 mid price
- **WHEN** mid price가 낮지만 공식 주문 계약의 worst executable price가 더 높다
- **THEN** reservation은 worst executable price를 사용하고 mid price로 bucket 여유를 늘리지 않는다

#### Scenario: FX haircut 미적용
- **WHEN** cross-currency 요청의 FX haircut policy가 없거나 multiplier가 1 미만이다
- **THEN** q_final은 0이고 INVALID_FX_HAIRCUT으로 거부된다

#### Scenario: 비선형 fee cap
- **WHEN** fee policy가 minimum fee를 가져 단순 per-share 나눗셈과 정확한 reserve 함수 결과가 다르다
- **THEN** bucket cap은 reserve(q)가 monetary remaining 이하인 최대 정수 q로 계산된다

### Requirement: 최종 수량은 strategy 수량과 모든 cap을 초과하지 않는다

lane 제안 수량의 canonical field는 `q_candidate`, 최종 수량의 canonical field는 `q_final`이어야 한다(SHALL). `q_final`은 `min(q_candidate, existing Guardian cap, horizon monetary cap quantity, market monetary cap quantity, strategy monetary cap quantity, sector monetary cap quantity, symbol monetary cap quantity)`의 non-negative integer여야 한다(SHALL). `q_final`은 어떤 입력에서도 `q_candidate` 또는 기존 Guardian 허용 수량보다 커서는 안 된다(MUST NOT). 누락, 음수, overflow, division error 또는 stale snapshot은 수량 0의 typed refusal이어야 하며(SHALL), 수량을 증가시키는 fallback이나 multiplier를 사용해서는 안 된다(MUST NOT).

#### Scenario: 여유 bucket이 큰 경우
- **WHEN** 모든 bucket 금액 잔여가 q_candidate를 수용한다
- **THEN** q_final은 q_candidate와 같고 더 커지지 않는다

#### Scenario: 하나의 bucket이 더 작음
- **WHEN** monetary reserve 함수로 산출한 sector cap quantity만 3주이고 q_candidate와 다른 cap은 10주 이상이다
- **THEN** q_final은 3주이며 binding cap과 계산 preimage가 기록된다

#### Scenario: 산술 overflow
- **WHEN** exposure 또는 valuation 계산이 지원 decimal 범위를 넘는다
- **THEN** q_final은 0이고 RISK_CALCULATION_INVALID로 거부되며 주문 intent는 생성되지 않는다

### Requirement: Scale-in은 filled와 held exposure를 모두 합산한다

동일 campaign의 모든 leg는 기존 filled monetary exposure와 unresolved HELD monetary reservation을 horizon, market, strategy, sector와 symbol bucket에 합산해야 한다(SHALL). 각 deduplicated positive fill transaction은 모든 적용 bucket에서 원 reservation policy/rounding에 따른 `transfer_delta = proportional_reserved_allocation(new_cumulative_fill) - previously_transferred`와 실제 fill price, allocated fee 및 persisted fill-time FX에 따른 `actual_delta`를 기록해야 한다(SHALL). 해당 fill의 `filled_delta`는 `max(transfer_delta, actual_delta)`여야 하고(SHALL), transfer delta만 HELD에서 차감해야 한다(SHALL). retry 또는 duplicate observation이 usage를 다시 증가시키거나 실제 체결가를 이유로 transfer보다 filled amount를 낮춰서는 안 된다(MUST NOT).

각 fill transaction은 모든 적용 bucket의 post-fill usage와 cap overage를 계산·영속해야 한다(SHALL). cap overage 또는 actual price/fee/FX unknown이어도 fill watermark와 authoritative Position apply를 거부, truncate 또는 rollback해서는 안 된다(MUST NOT). overage는 durable `RISK_OVERAGE`, unknown actual은 0원으로 간주하지 않고 durable `UNKNOWN_ACTUAL_RISK`로 모든 적용 bucket/owner에 latch해 신규 exposure만 차단해야 한다(SHALL). actual evidence가 나중에 확정되면 같은 fill identity의 amount를 `max(transfer, actual)`로 단조 보완하고 fill을 재적용해서는 안 된다(MUST NOT). crash는 fill/Position, HELD transfer, filled amount, all-bucket overage/latch를 전부 commit하거나 전부 rollback해야 한다(SHALL). cancel/expiry는 미체결 held 잔량에 대응하는 금액만 멱등 release해야 한다(SHALL).

#### Scenario: 두 번째 leg 예약
- **WHEN** 첫 leg의 일부가 filled이고 잔량 reservation이 HELD인 상태에서 두 번째 leg를 평가한다
- **THEN** 모든 bucket은 filled와 두 HELD 잔량을 합산한 뒤 cap을 판정한다

#### Scenario: 부분체결 재처리
- **WHEN** 같은 cumulative fill watermark가 다시 적용된다
- **THEN** held-to-filled 이동과 bucket 총 usage는 추가로 변하지 않는다

#### Scenario: 실제 체결 노출이 proportional transfer보다 큼
- **WHEN** partial fill의 실제 체결가격·allocated fee·FX monetary exposure가 원 예약의 proportional HELD transfer보다 크다
- **THEN** 모든 적용 bucket에서 proportional amount만 HELD에서 빠지고 filled amount는 실제 exposure이며 증가한 usage와 overage가 기록된다

#### Scenario: actual price fee 또는 FX unknown
- **WHEN** authoritative fill은 도착했지만 실제 price, allocated fee 또는 fill-time FX provenance 중 하나가 없다
- **THEN** fill/Position은 보존되고 계산 가능한 proportional amount는 provisional floor이며 actual을 0으로 만들지 않고 모든 적용 bucket에 UNKNOWN_ACTUAL_RISK가 latch되어 신규 exposure만 차단된다

#### Scenario: cap을 넘는 late replacement fill
- **WHEN** replacement 또는 cancelled predecessor의 late fill이 horizon, market, strategy, sector 또는 symbol cap을 초과한다
- **THEN** fill/Position과 watermark는 exactly once 보존되고 모든 적용 bucket의 overage가 기록되며 RISK_OVERAGE가 신규 exposure만 차단한다

#### Scenario: fill accounting crash와 retry
- **WHEN** fill transaction이 Position 또는 일부 bucket 갱신 뒤 commit 전에 crash하고 같은 observation이 재시도된다
- **THEN** 첫 transaction은 전부 rollback되고 retry가 Position, proportional transfer, filled max, all-bucket overage/latch를 정확히 한 번 commit한다

### Requirement: Symbol은 하나의 owning lane와 campaign만 가진다

account, market, symbol, prospective-or-actual position generation에는 하나의 owning lane와 active campaign만 존재해야 한다(SHALL). first fill 전에는 a065 prospective generation token이 owner identity에 사용되고 실제 successor generation에 set-once 결합되어야 한다(SHALL). `q_final` 확정 뒤 최초 owner 획득, q_final을 봉인한 GuardianDecision과 모든 monetary bucket reservation은 하나의 journal transaction에서 commit되어야 하며(SHALL), 경쟁 lane의 owner conflict는 결정과 reservation을 모두 rollback해야 한다(SHALL). 같은 owner의 후속 leg만 기존 ownership을 재사용할 수 있다(SHALL).

#### Scenario: 두 lane의 동시 최초 진입
- **WHEN** 서로 다른 lane가 같은 account/market/symbol generation의 owner와 bucket을 동시에 예약한다
- **THEN** 한 lane만 commit되고 다른 lane는 OWNER_CONFLICT로 거부되며 두 번째 campaign이나 HELD reservation은 없다

#### Scenario: 소유 lane의 scale-in
- **WHEN** 기존 owning lane와 campaign이 다음 leg를 예약한다
- **THEN** ownership은 바뀌지 않고 새 수량만 동일 campaign의 bucket usage에 합산된다

### Requirement: Ownership과 bucket은 journal에서 결정적으로 재구성된다

시스템은 decision, prospective/actual generation owner, monetary reservation, fill, cancel/expiry와 Position generation event를 replay해 ownership과 bucket usage를 결정적으로 재구성해야 한다(SHALL). snapshot drift, orphan reservation 또는 중복 owner가 발견되면 해당 account/symbol의 새 exposure를 차단해야 하며(SHALL), 임의 합산이나 자동 owner 교체를 수행해서는 안 된다(MUST NOT). owner release는 generation CLOSED/수량 0, pending exposure-raising mutation/HELD reservation 부재, clean reconciliation, 그리고 이전 generation의 active/pending protection order·replace/recovery saga·sell/reduce-only claim·sell mutation·unresolved fill observation 부재가 증명될 때만 멱등 수행되어야 한다(SHALL). release가 해당 claim을 자동 취소·삭제해서는 안 된다(MUST NOT).

#### Scenario: 재시작 후 replay
- **WHEN** process restart 후 동일 ordered journal evidence를 replay한다
- **THEN** restart 전과 동일한 owner, filled/held usage와 snapshot digest가 복원된다

#### Scenario: 고아 reservation
- **WHEN** replay가 유효한 decision 또는 campaign이 없는 HELD reservation을 발견한다
- **THEN** 새 exposure는 RECONSTRUCTION_MISMATCH로 차단되고 reservation을 자동 삭제하지 않는다

#### Scenario: 종결 generation release
- **WHEN** Position generation이 CLOSED이고 pending entry/HELD, protection saga/order, sell/reduce-only claim/mutation과 unresolved fill이 모두 없으며 reconciliation이 broker 수량 0을 증명한다
- **THEN** owner release event가 한 번 기록되며 retry는 같은 결과를 반환한다

#### Scenario: 이전 protection claim 잔존
- **WHEN** Position은 CLOSED지만 이전 generation의 broker protection order 또는 replacement saga가 active, pending, stale 또는 unknown이다
- **THEN** owner는 release되지 않고 새 generation exposure는 차단되며 claim을 자동 삭제하지 않는다

#### Scenario: 이전 sell claim 잔존
- **WHEN** Position은 CLOSED지만 이전 generation의 sell/reduce-only mutation 또는 unresolved fill observation이 남아 있다
- **THEN** owner는 release되지 않고 late sell이 새 generation에 착지하지 않도록 신규 entry를 차단한다

### Requirement: Bucket과 loss lock은 위험 감소를 차단하지 않는다

bucket limit, owner conflict, loss lock, stale snapshot, missing FX/sector 및 reconstruction mismatch는 EXPOSURE_RAISING 경로에만 적용되어야 한다(SHALL). stop, emergency exit, reduce-only mutation, reconciliation과 fill detection은 bucket 계산, owner 획득, 외부 evidence 수집 또는 entry lock 해제를 기다려서는 안 된다(MUST NOT).

#### Scenario: loss lock 중 stop
- **WHEN** short 또는 medium horizon loss lock이 활성인 동안 stop exit가 요청된다
- **THEN** stop은 bucket admission 없이 기존 risk-reducing 경로로 즉시 진행한다

#### Scenario: bucket 재구성 오류 중 fill 관측
- **WHEN** bucket reconstruction mismatch가 있는 동안 기존 주문의 fill observation이 도착한다
- **THEN** fill detection과 Position/reconciliation 투영은 계속되고 새 exposure만 차단된다
