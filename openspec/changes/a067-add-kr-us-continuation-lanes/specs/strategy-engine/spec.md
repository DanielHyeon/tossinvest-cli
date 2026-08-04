## ADDED Requirements

### Requirement: KR와 US 단기 continuation lane은 함께 제공되고 독립 실행된다
change를 포함한 lane registry는 `kr_short_flow_continuation_v1`과 `us_short_participation_continuation_v1`을 같은 release에 함께 등록해야 한다 (SHALL).
각 lane은 자신의 market scope, versioned evidence와 desired/effective state만 사용해야 하며 다른
시장의 구현 완료, 세션 상태, refusal 또는 운영 안정화를 조건으로 사용해서는 안 된다
(MUST NOT). 각 lane은 순수 entry/add decision 또는 typed invalidation/refusal만 반환하고 broker,
journal writer, 운영 토글과 exit decision authority를 직접 소유해서는 안 된다 (MUST NOT).

#### Scenario: 두 continuation lane의 동시 제공
- **WHEN** change를 포함한 lane registry conformance를 검사한다
- **THEN** 두 고유 lane ID/version이 모두 존재하며 한쪽만 등록된 build는 검증에 실패한다

#### Scenario: 한 시장 OFF의 독립성
- **WHEN** KR continuation lane이 OFF이고 US continuation lane은 유효한 US 입력으로 평가 가능하다
- **THEN** KR 신규 decision과 buy mutation은 0건이고 US 평가는 계속되며 공통 exit engine은 두 시장을 독립 감독한다

#### Scenario: lane invalidation과 공통 exit
- **WHEN** continuation invalidation과 공통 exit engine의 stop 조건이 같은 evaluation interval에 관측된다
- **THEN** lane은 typed invalidation만 반환하고 exit order/decision을 만들지 않으며 공통 exit engine은 lane 결과를 기다리지 않고 독립적으로 risk-reducing decision을 발급할 수 있다

### Requirement: continuation 입력은 시장별 strict versioned schema와 정수 산술을 사용한다
Continuation evaluator는 시장별 strict versioned input schema와 exact integer arithmetic만 사용해야 한다 (SHALL).
공통 envelope는 schema version, market, symbol, source record/digest, `effective_at`, `observed_at`,
`ingested_at`, `evaluated_at`, `fresh_until`, metric units, threshold set과 config digest를 요구해야 하고
(SHALL), `effective_at <= observed_at <= ingested_at <= evaluated_at <= fresh_until`을 만족하지 않거나
unknown field, 중복 JSON object key, unit, digest 또는 overflow가 있으면 typed refusal이어야 한다 (SHALL).
KR flow schema는 signed `net_flow_notional_minor`, positive `turnover_notional_minor`와
`flow_pressure_ppm = trunc_toward_zero(net_flow_notional_minor * 1_000_000 /
turnover_notional_minor)`를 사용해야 한다 (SHALL). US participation schema는 non-negative
`participating_volume_shares`, positive `baseline_volume_shares`, positive `reference_price_minor`,
`last_price_minor`, `participation_ppm = floor(participating_volume_shares * 1_000_000 /
baseline_volume_shares)`와 `price_change_ppm = trunc_toward_zero((last_price_minor -
reference_price_minor) * 1_000_000 / reference_price_minor)`를 사용해야 한다 (SHALL).
Evaluator는 versioned config에 저장된 integer ppm thresholds와 정의된 inclusive/exclusive 비교를
순서대로 적용하고 float, local recomputation 또는 누락값 fallback을 사용해서는 안 된다 (MUST NOT).

#### Scenario: KR flow 단위 또는 digest 불일치
- **WHEN** KR flow record의 notional unit, threshold config digest 또는 arithmetic preimage가 schema와 다르다
- **THEN** typed evidence refusal을 반환하고 EntryDecision과 mutation은 0건이다

#### Scenario: US participation 경계값
- **WHEN** US input의 integer ppm metric이 inclusive threshold와 정확히 같다
- **THEN** versioned comparison contract에 따라 결정적으로 동일한 accept/refusal이 생성되고 float rounding은 사용되지 않는다

#### Scenario: freshness와 시각 순서 위반
- **WHEN** evidence가 `evaluated_at > fresh_until`이거나 timestamp ordering을 위반한다
- **THEN** 해당 시장 lane만 typed stale/invalid refusal을 반환하고 peer market 평가는 계속된다

#### Scenario: duplicate JSON key
- **WHEN** KR 또는 US evidence의 top-level이나 nested object에 같은 key가 두 번 나타난다
- **THEN** 마지막 값으로 덮어쓰지 않고 strict decoder가 입력 전체를 거부한다

### Requirement: continuation progression은 immutable 8:4:2 planned allocation을 따른다
Continuation campaign은 immutable campaign risk budget과 planned quantity `Q`에서 세 leg ceiling을 계산해야 한다 (SHALL).
계획 수량은 `q1=floor(Q*8/14)`, `q2=floor(Q*4/14)`, `q3=Q-q1-q2`여야 하며 (SHALL),
campaign risk budget, per-share risk preimage, `Q`, weights, planned ceilings와 policy/config digest는
첫 leg 전에 영속되고 변경되어서는 안 된다 (MUST NOT). 각 요청 `q_leg`는 해당 leg의 planned
remaining과 a066이 같은 snapshot에서 반환한 `q_final` 이하이어야 한다 (SHALL). Partial fill,
zero-fill cancel/expiry, retry와 이전 leg의 미사용량은 현재 또는 후속 leg planned ceiling을
증가시키거나 옮겨서는 안 된다 (MUST NOT). 두 번째와 세 번째 leg는 fresh continuation
confirmation과 공통 risk 승인을 다시 통과해야 하고 (SHALL), scale-in은 effective stop을 더
불리하게 만들 수 없다 (MUST NOT).
신규 leg admission은 account valuation currency minor units에서 cumulative filled risk,
unresolved held risk와 proposed leg risk의 checked sum이 campaign risk budget 이하인 경우에만
허용되어야 한다 (SHALL). 각 positive fill의 filled risk는
`max(transferred_conservative_reservation_minor, ceil_minor(qty * max(entry_minor -
effective_stop_minor, 0) + entry_fees_minor + estimated_exit_fees_levies_minor))`여야 하며
(SHALL), filled usage는 held에서 이전한 a066 conservative reservation보다 낮아져서는
안 된다 (MUST NOT). Held와 proposed risk는 a066이 봉인한 conservative reservation을
사용해야 한다 (SHALL). US 금액은 a066 decision에 고정된 동일한 공식 FX quote ID,
as-of, quote-to-account rate direction과 1 이상의 conservative haircut을 사용해 account
valuation currency로 환산한 뒤 minor-unit ceil해야 하며 (SHALL), 다른 FX snapshot을
섞어서는 안 된다 (MUST NOT). 모든 곱셈·덧셈·환산은 checked arithmetic이어야 하고
(SHALL), overflow나 입력/FX provenance 부재는 신규 admission을 typed refusal해야 한다
(SHALL). 이미 발생한 fill의 actual risk가 budget을 넘거나 계산 불가이면 fill 처리를
막지 않고 `CAMPAIGN_RISK_OVERAGE`나 unknown-risk latch를 영속해 모든 후속
exposure-raising leg를 차단해야 한다 (SHALL). Overage를 unused planned quantity로
상쇄해서는 안 된다 (MUST NOT).
A066 cap seal은 exact immutable campaign-plan digest, plan policy digest와 proposed reservation
quantity를 포함해야 하며 (SHALL), policy가 다르거나 같은 market/policy의 다른 campaign plan에서
replay해서는 안 된다 (MUST NOT).
Fill/cancel risk event는 exact plan/risk state, campaign ID, leg ordinal, order reference, source digest와
RFC3339Nano observation time에 결속되어야 한다 (SHALL). 이 scope가 불일치하거나 provenance가
비어 있으면 held/filled accounting을 변경하지 않고 non-applied evidence와 unknown-risk latch를
남겨야 한다 (SHALL). FillID/CancelID가 없으면 full non-ID preimage digest를 synthetic identity로
사용해 exact raw retry만 idempotent하게 처리하고 다른 raw preimage를 합쳐서는 안 된다 (MUST NOT).
Campaign ID, order reference와 source digest는 각각 최대 256 bytes여야 한다 (SHALL).
Quantity가 0인 FillRiskEvent는 positive fill로 적용하거나 held를 release하거나 filled risk를
증가시켜서는 안 되며 (MUST NOT), non-applied evidence와 unknown-risk latch를 남겨야 한다 (SHALL).
Stop candidate는 price/source/policy/version/digest/observed-at/fresh-until을 함께 봉인해야 하며
(SHALL), `observed_at <= evaluated_at <= fresh_until`을 만족하는 경우에만 admission에 사용할 수
있다 (SHALL). Cancelled 또는 expired leg는 모두 terminal이어야 한다 (SHALL).

#### Scenario: floor와 마지막 remainder
- **WHEN** immutable planned quantity Q로 8:4:2 plan을 만든다
- **THEN** 앞의 두 leg는 exact integer floor이고 세 번째 leg가 remainder를 받아 세 planned ceiling 합은 정확히 Q다

#### Scenario: partial fill 뒤 upward reallocation 시도
- **WHEN** 첫 leg가 planned ceiling보다 적게 fill되고 남은 수량을 두 번째 또는 세 번째 leg에 더하려 한다
- **THEN** 추가분은 거부되고 각 leg ceiling과 campaign Q는 최초 planned basis로 유지된다

#### Scenario: a066 cap이 planned remaining보다 작다
- **WHEN** 다음 leg planned remaining은 5주이고 current a066 q_final은 3주다
- **THEN** q_leg는 최대 3주이며 retry나 fill 차이로 5주를 초과하거나 cap을 우회하지 않는다

#### Scenario: property와 replay 불변식
- **WHEN** 임의 Q, partial fills, duplicate retry와 cancel/expiry sequence를 replay한다
- **THEN** 모든 q_leg는 해당 planned ceiling과 a066 cap 이하이고 admitted filled/held/proposed monetary risk는 immutable campaign budget 이하이며 duplicate fill은 risk를 다시 늘리지 않고 zero-fill cancel/expiry는 미체결 conservative held만 release한다

#### Scenario: foreign 또는 incomplete fill/cancel event
- **WHEN** 다른 campaign ID, leg 99, empty order/source 또는 invalid observation time의 risk event가 도착한다
- **THEN** 기존 held/filled 값은 그대로 유지되고 non-applied evidence와 unknown-risk latch만 남으며 exact raw retry만 duplicate다

#### Scenario: fill risk accounting precondition failure
- **WHEN** transferred reservation이 held보다 크거나 held/filled 값이 corrupt하거나 `filled + risk`가 overflow한다
- **THEN** fill evidence와 unknown-risk latch는 보존되지만 held/filled는 둘 다 기존 값 그대로이며 부분 accounting update는 0건이다

#### Scenario: zero quantity fill observation
- **WHEN** FillRiskEvent quantity가 0이거나 기존 positive FillID와 zero-quantity raw event가 충돌한다
- **THEN** positive fill application은 0건이고 held/filled는 불변이며 unknown-risk latch와 non-applied evidence가 보존된다

#### Scenario: transferred reservation보다 낮은 actual risk
- **WHEN** positive fill의 가격·stop·비용 산식값이 held에서 이전된 conservative reservation보다 낮다
- **THEN** filled risk는 transferred conservative reservation을 유지하고 사용량을 낮추지 않는다

#### Scenario: US fill의 frozen FX 일관성
- **WHEN** US fill risk를 admission에 봉인된 FX quote가 아닌 더 유리한 새 rate로 환산하려 한다
- **THEN** 환산을 거부하고 동일한 quote ID/as-of/rate direction/haircut으로 ceil한 account valuation minor units만 사용한다

#### Scenario: actual fill risk overage
- **WHEN** actual fill price, fees 또는 FX로 recalculated cumulative filled risk가 immutable campaign budget을 초과한다
- **THEN** fill/position projection과 common exit는 계속되고 overage latch가 모든 후속 신규 leg를 차단한다

#### Scenario: actual risk evidence 부재
- **WHEN** fill-time fee 또는 authoritative FX가 없어 actual monetary risk를 계산할 수 없다
- **THEN** risk를 0으로 추정하지 않고 fill을 적용한 뒤 unknown-risk overage latch로 후속 신규 leg를 차단한다

#### Scenario: invalidation과 add가 동시에 성립한다
- **WHEN** 같은 evaluation에서 continuation add 조건과 exit/structural invalidation이 함께 관측된다
- **THEN** lane은 typed invalidation/refusal만 반환하고 신규 leg decision은 0건이며 공통 exit engine 권한은 계속된다

#### Scenario: invalidation code가 비어 있다
- **WHEN** structural 또는 exit invalidation flag가 true지만 typed invalidation code가 비어 있다
- **THEN** lane은 invalidation을 합성하지 않고 typed invalidation-invalid refusal을 반환한다

### Requirement: continuation 결과는 시장별 lineage를 보존한다
모든 continuation decision과 refusal은 market-scoped lineage를 보존해야 한다 (SHALL).
결과는 market, lane ID/version, candidate ID, evidence/schema/config digest, campaign ID, position
generation, immutable risk-budget digest와 leg ordinal/planned ceiling을 포함해야 하고 (SHALL), KR와
US lineage 또는 한 시장의 성과를 다른 시장 lane에 결합해서는 안 된다 (MUST NOT).

#### Scenario: 동일 symbol text의 시장별 평가
- **WHEN** KR와 US 평가가 같은 symbol text 또는 같은 시각에 결과를 생성한다
- **THEN** account/market/symbol/position-generation-scoped campaign과 lane lineage가 서로 다른 attribution key로 유지된다
