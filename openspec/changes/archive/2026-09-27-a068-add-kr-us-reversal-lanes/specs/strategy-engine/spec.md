## ADDED Requirements

### Requirement: KR와 US 단기 reversal lane은 함께 제공되고 독립 실행된다
change를 포함한 lane registry는 `kr_short_absorption_reversal_v1`과 `us_short_dislocation_reversal_v1`을 같은 release에 함께 등록해야 한다 (SHALL).
각 lane은 자신의 market scope와 strict versioned structural evidence만 평가해야 하며 다른 시장의
구현 완료, 운영 안정성, 세션 상태 또는 enablement를 선행 조건으로 삼아서는 안 된다
(MUST NOT). 각 lane은 순수 entry/add decision 또는 typed invalidation/refusal만 반환하고 broker,
journal writer, 운영 토글과 exit decision authority를 직접 소유해서는 안 된다 (MUST NOT).

#### Scenario: 두 reversal lane의 동시 제공
- **WHEN** change를 포함한 lane registry conformance를 검사한다
- **THEN** 두 고유 lane ID/version이 모두 존재하며 한쪽만 등록된 build는 검증에 실패한다

#### Scenario: 한 시장 장애의 독립성
- **WHEN** US reversal evidence가 실패하고 KR reversal lane은 유효한 KR 입력으로 평가 가능하다
- **THEN** US는 typed refusal과 신규 mutation 0건을 반환하고 KR 평가 및 공통 exit engine은 계속된다

#### Scenario: lane invalidation과 공통 exit
- **WHEN** reversal invalidation과 공통 exit engine의 stop 조건이 같은 interval에 관측된다
- **THEN** lane은 typed invalidation만 반환하고 exit decision/order를 만들지 않으며 공통 exit engine은 독립적으로 risk-reducing decision을 발급할 수 있다

### Requirement: reversal 입력은 시장별 strict schema와 시간 순서가 고정된다
Reversal evaluator는 시장별 strict versioned input schema와 exact integer arithmetic을 사용해야 한다 (SHALL).
공통 envelope는 schema version, market, symbol, source record/digest, units, `effective_at`,
`observed_at`, `ingested_at`, `evaluated_at`, `fresh_until`, threshold set, structural-window duration과
config digest를 요구해야 하고 (SHALL), `effective_at <= observed_at <= ingested_at <=
evaluated_at <= fresh_until`을 정확히 만족해야 한다 (SHALL). Timestamp ordering, unit, digest,
denominator 또는 overflow 오류는 typed refusal이어야 한다 (SHALL).
KR absorption schema는 non-negative `absorbed_notional_minor`, positive
`aggressive_sell_notional_minor`와 `absorption_ppm=floor(absorbed_notional_minor * 1_000_000 /
aggressive_sell_notional_minor)`를 사용해야 한다 (SHALL). US dislocation schema는 positive
`reference_price_minor`, non-negative `dislocation_low_price_minor`, non-negative
`dislocation_volume_shares`, positive `baseline_volume_shares`,
`drawdown_ppm=floor((reference_price_minor-dislocation_low_price_minor) * 1_000_000 /
reference_price_minor)`와 `relative_volume_ppm=floor(dislocation_volume_shares * 1_000_000 /
baseline_volume_shares)`를 사용해야 한다 (SHALL). Negative numerator, unknown field/unit 또는
schema/config digest mismatch를 보정해서는 안 된다 (MUST NOT).

#### Scenario: KR absorption denominator 부재
- **WHEN** aggressive sell notional이 0이거나 단위/config digest가 schema와 다르다
- **THEN** KR lane은 typed invalid evidence refusal을 반환하고 decision과 mutation은 0건이다

#### Scenario: US dislocation exact threshold
- **WHEN** drawdown/relative-volume ppm이 versioned inclusive threshold와 정확히 같다
- **THEN** integer comparison 순서에 따라 결정적으로 같은 결과를 반환하고 float rounding을 사용하지 않는다

#### Scenario: freshness inclusive boundary
- **WHEN** reversal evidence의 모든 시각 순서가 유효하고 `evaluated_at == fresh_until`이다
- **THEN** freshness 경계는 유효하며 다른 schema/strategy 조건을 계속 평가한다

#### Scenario: freshness boundary 초과
- **WHEN** `evaluated_at` 이 `fresh_until`보다 1 tick 늦거나 `observed_at > ingested_at`이다
- **THEN** 해당 market lane은 typed stale/timestamp refusal을 반환하고 신규 decision과 mutation은 0건이다

### Requirement: reversal 구조 확인은 bounded causal order를 만족한다
최종 leg의 structural confirmation은 같은 account/market/symbol/position generation과 evidence version에서 causal order와 bounded window를 만족해야 한다 (SHALL).
각 event는 stable record ID/digest와 `sweep_at`, `break_at`, `reclaim_at`을 가져야 하고
`sweep_at <= break_at <= reclaim_at <= evaluated_at`이어야 하며 (SHALL),
`evaluated_at-sweep_at`은 versioned config의 non-negative `structural_window_ns` 이하이고 모든
event는 `fresh_until` 전이어야 한다 (SHALL). Price decline, 평균단가 변화, event 일부 또는 다른
scope의 event를 이용해 confirmation을 합성해서는 안 된다 (MUST NOT).

#### Scenario: 순서가 뒤바뀐 event
- **WHEN** reclaim_at이 break_at보다 이르거나 break_at이 sweep_at보다 이르다
- **THEN** 최종 leg는 typed structural-order refusal이고 신규 mutation은 0건이다

#### Scenario: bounded window 초과
- **WHEN** 세 event가 모두 있지만 evaluated_at-sweep_at이 structural_window_ns를 초과한다
- **THEN** stale structural refusal을 반환하고 final leg를 허용하지 않는다

#### Scenario: scope가 다른 reclaim
- **WHEN** sweep/break는 KR symbol generation에 속하지만 reclaim은 다른 market 또는 generation에 속한다
- **THEN** scope mismatch refusal을 반환하고 event를 결합하지 않는다

### Requirement: reversal progression은 immutable 2:4:8 planned allocation을 따른다
Reversal campaign은 immutable campaign risk budget과 planned quantity `Q`에서 세 leg ceiling을 계산해야 한다 (SHALL).
계획 수량은 `q1=floor(Q*2/14)`, `q2=floor(Q*4/14)`, `q3=Q-q1-q2`여야 하며 (SHALL),
campaign risk budget, per-share risk preimage, Q, weights, planned ceilings와 policy/config digest는
첫 leg 전에 영속되고 변경되어서는 안 된다 (MUST NOT). 각 `q_leg`는 해당 planned remaining과
a066 current `q_final` 이하이어야 한다 (SHALL). Partial fill, zero-fill cancel/expiry, retry 또는
이전 leg의 미사용량은 현재/후속 leg ceiling을 증가시키거나 옮겨서는 안 된다 (MUST NOT).
Scale-in은 effective stop을 불리하게 변경해서는 안 되며 (MUST NOT), final leg는 유효한 structural
confirmation과 공통 risk 승인을 모두 요구해야 한다 (SHALL).
신규 leg admission은 account valuation currency minor units에서 cumulative filled risk,
unresolved held risk와 proposed leg risk의 checked sum이 campaign risk budget 이하인 경우에만
허용되어야 한다 (SHALL). 각 positive fill의 filled risk는
`max(transferred_conservative_reservation_minor, ceil_minor(qty * max(entry_minor -
effective_stop_minor, 0) + entry_fees_minor + estimated_exit_fees_levies_minor))`여야 하며
(SHALL), held/proposed risk는 a066 conservative reservations여야 한다 (SHALL). Filled usage는
transferred conservative reservation보다 낮아져서는 안 된다 (MUST NOT). US risk는 a066
decision에 봉인된 동일한 공식 FX quote ID/as-of/quote-to-account rate direction과 1 이상
conservative haircut을 사용해 account valuation currency로 환산·minor-unit ceil해야 한다
(SHALL). 모든 산술은 checked이어야 하고 (SHALL), overflow·누락·FX provenance 불일치는
신규 admission typed refusal이어야 한다 (SHALL). 이미 발생한 fill의 계산 불가 또는
actual budget 초과는 fill 처리를 막지 않고 `CAMPAIGN_RISK_OVERAGE`나 unknown-risk
latch로 후속 exposure raising을 차단해야 한다 (SHALL). Overage를 unused planned quantity로
상쇄해서는 안 된다 (MUST NOT).

#### Scenario: floor와 마지막 remainder
- **WHEN** immutable planned quantity Q로 2:4:8 plan을 만든다
- **THEN** 첫 두 leg는 exact floor이고 세 번째 leg가 remainder를 받아 planned ceiling 합은 정확히 Q다

#### Scenario: partial fill upward reallocation
- **WHEN** 첫 leg가 일부만 fill되고 미사용량을 final leg에 추가하려 한다
- **THEN** 추가분은 거부되고 최초 planned ceilings와 campaign Q는 유지된다

#### Scenario: price decline만 발생한다
- **WHEN** 가격은 하락했지만 causal/bounded sweep, break, reclaim 중 하나가 유효하지 않다
- **THEN** final leg는 typed refusal이고 entry/add mutation은 0건이다

#### Scenario: q_leg와 a066 cap property
- **WHEN** 임의 Q, q_final, fills, retries와 cancellations를 생성해 transition을 반복한다
- **THEN** 모든 q_leg는 해당 planned ceiling과 a066 q_final 이하이고 admitted filled/held/proposed monetary risk는 immutable campaign budget 이하이며 duplicate fill은 risk를 다시 늘리지 않고 cancel/expiry는 미체결 held만 release하며 unfilled allocation을 상향 재배분하지 않는다

#### Scenario: conservative reservation floor
- **WHEN** partial fill의 exact stop-distance와 비용 산식이 held에서 이전된 conservative reservation보다 작다
- **THEN** filled usage는 transferred reservation을 유지하고 duplicate observation은 그 금액을 다시 늘리지 않는다

#### Scenario: US frozen FX mismatch
- **WHEN** US fill risk 환산이 a066에 봉인된 FX quote/as-of/rate direction/haircut과 다른 preimage를 사용한다
- **THEN** 신규 admission은 typed refusal이며 이미 발생한 fill은 보존하고 unknown-risk latch로 후속 leg를 차단한다

#### Scenario: actual fill risk overage
- **WHEN** slippage, fees 또는 FX로 recalculated cumulative filled risk가 immutable campaign budget을 초과한다
- **THEN** fill/position projection과 common exit는 계속되고 overage latch가 모든 후속 신규 leg를 차단한다

#### Scenario: actual risk evidence 부재
- **WHEN** fill-time fee 또는 authoritative FX가 없어 actual monetary risk를 계산할 수 없다
- **THEN** risk를 0으로 추정하지 않고 fill을 적용한 뒤 unknown-risk overage latch로 후속 신규 leg를 차단한다

#### Scenario: invalidation과 add가 동시에 성립한다
- **WHEN** 같은 evaluation에서 structural invalidation과 scale-in 조건이 함께 참이다
- **THEN** lane은 typed invalidation/refusal만 반환하고 신규 leg는 0건이며 공통 exit engine 권한은 계속된다

### Requirement: reversal 결과는 시장별 lineage를 보존한다
모든 reversal decision과 refusal은 market-scoped lineage를 보존해야 한다 (SHALL).
결과는 market, lane ID/version, candidate ID, schema/config/structural evidence digests, campaign ID,
position generation, immutable risk-budget digest와 leg ordinal/planned ceiling을 포함해야 하고
(SHALL), KR와 US attribution을 결합해서는 안 된다 (MUST NOT).

#### Scenario: KR와 US 동시 evaluation
- **WHEN** KR와 US reversal lane이 같은 scheduler cycle에서 각각 결과를 생성한다
- **THEN** 두 결과는 account/market/symbol/position-generation-scoped lane과 campaign attribution key로 독립 유지된다
