# strategy-engine Specification

## Purpose
TBD - created by archiving change a047-add-strategy-engine. Update Purpose after archive.
## Requirements
### Requirement: 승인 후보는 독립 lane의 순수 결정으로 변환된다
strategy lane는 ApprovedCandidate와 versioned market inputs를 받아 EntryDecision 또는 명시적 refusal을 반환해야 하며 broker·journal·운영 토글을 직접 변경해서는 안 된다 (MUST NOT).

#### Scenario: 진입 결정
- **WHEN** 활성 lane가 유효 ApprovedCandidate를 평가해 진입 조건을 충족한다
- **THEN** candidate ID, lane ID/version, stop/target와 RiskIntent 입력을 가진 결정이 생성된다

#### Scenario: lane OFF
- **WHEN** lane desired/effective state가 OFF다
- **THEN** 신규 EntryDecision과 buy mutation은 0건이고 기존 exit loop는 계속된다

### Requirement: 첫 lane는 frozen KRX Parker VWAP conservative v1이다
첫 lane는 StockOS commit `d75113d3c338148606d86c8aedbbeb7ed446c0b8`와 source-set digest `09260ac29e50ed4d2a43d0e274f9a17465e00ee36fb61d759127f158985c23bd`의 `parker_vwap_trend_v1` conservative gate를 KRX regular-session closed 5-minute input에만 적용해야 한다 (SHALL). server-owned immutable constants와 gate order를 바꾸려면 새 lane version과 activation manifest 승인이 필요하다 (SHALL).
KRX open은 `09:00 KST`로 봉인하고 session open/close/evaluation은 같은 KST 거래일이어야 하며, caller가 이동한 session window를 받아서는 안 된다 (MUST NOT).
공식 1분봉의 `timestamp`는 봉이 **닫힌** 시각이므로 (2026-08-18 KR/US 라이브 실측), 라벨이 `t`인 봉은 `[t-1분, t)` 구간을 담아야 한다 (SHALL). 5분 버킷이 여는 시각은 첫 라벨보다 1분 이르며 5분 격자 정렬과 `closed_at`은 그 여는 시각에서 계산해야 한다 (SHALL). 정규장 편입 판정은 라벨 기준 `09:01`부터 `15:30`까지 양끝 포함이어야 하고 (SHALL), `09:00` 라벨(개장 전 `08:59`~`09:00`)을 정규장으로 받아들이거나 `15:30` 라벨(정규장 마지막 1분)을 버려서는 안 된다 (MUST NOT).

#### Scenario: StockOS translated accept arithmetic
- **WHEN** frozen fixture가 VWAP above/slope, EMA9 bullish pullback, LVN forward space, untangled/band/RR, age/drift를 모두 통과한다
- **THEN** `krx_parker_vwap_conservative_v1`은 source와 같은 entry, 0.7% stop, 3R target, expected RR와 accept provenance를 반환한다

#### Scenario: StockOS 세션 refusal 패리티
- **WHEN** frozen KRX calendar가 non-trading, 시가 동시호가, 종가 동시호가, 시간외, 시초 제외 또는 close-minus-45m cutoff를 나타낸다
- **THEN** lane은 source 순서와 같은 typed refusal reason을 반환하고 EntryDecision과 broker call은 0건이다

#### Scenario: 지원하지 않는 시장
- **WHEN** US 또는 pre/after-market candidate가 첫 lane를 요청한다
- **THEN** typed unsupported-scope refusal을 반환하고 EntryDecision과 broker call은 0건이다

#### Scenario: source digest 변경
- **WHEN** runtime lane source/constants digest가 manifest의 frozen digest와 다르다
- **THEN** desired state와 무관하게 effective entry는 OFF이고 새 manifest 승인 없이는 dispatch하지 않는다

#### Scenario: 불완전한 5분봉
- **WHEN** official 1분봉에 KST 정규장 밖 minute, 중간 누락 또는 아직 닫히지 않은 bucket이 있다
- **THEN** lane은 해당 5분봉을 만들지 않고 typed bar-integrity refusal과 broker call 0건을 반환한다

#### Scenario: 장 시작 첫 버킷
- **WHEN** 라벨 `09:01`~`09:05`인 다섯 개의 공식 1분봉을 `09:05` 이후에 집계한다
- **THEN** 여는 시각 `09:00`, 닫는 시각 `09:05`의 5분봉 하나를 만들고 정렬 거절은 0건이다

#### Scenario: 개장 전 1분을 담은 라벨
- **WHEN** 라벨 `09:00`인 봉이 5분 버킷에 포함된다
- **THEN** 그 봉은 `08:59`~`09:00`을 담으므로 typed outside-regular-session refusal을 반환하고 5분봉을 만들지 않는다

#### Scenario: 종목 상태 권위 부재
- **WHEN** HALT/LIMIT/MANAGED를 판정할 authoritative 상태가 없거나 30초보다 stale이다
- **THEN** quote나 price limit로 추측하지 않고 effective entry를 OFF로 유지한다

### Requirement: strategy entry는 공식 LIVE 경로만 사용한다
승인된 strategy entry는 Guardian, durable journal과 official Open API gateway를 순서대로 통과해야 하며 paper/shadow/canary order path를 가져서는 안 된다 (MUST NOT).

#### Scenario: 운영자 LIVE 승인
- **WHEN** 전체 gate가 통과하고 운영자가 lane와 automation을 명시적으로 승인한다
- **THEN** 다음 유효 결정은 공식 LIVE gateway를 사용한다

#### Scenario: Guardian refusal
- **WHEN** Guardian이 첫 실패 단계에서 거부한다
- **THEN** broker request는 0건이고 refusal과 provenance가 journal에 기록된다

### Requirement: KR와 US weekly value lane은 함께 제공되고 독립 실행된다
change를 포함한 lane registry는 `kr_weekly_disclosure_value_v1`과 `us_weekly_disclosure_value_v1`을 같은 release에 함께 등록해야 한다 (SHALL).
KR lane은 OpenDART, US lane은 SEC EDGAR에서 정규화된 point-in-time disclosure evidence만
사용해야 하며 (SHALL), 어느 시장도 다른 시장의 구현 완료, 운영 안정성, disclosure availability
또는 enablement를 조건으로 사용해서는 안 된다 (MUST NOT). 각 lane은 순수 entry/add decision
또는 typed invalidation/refusal만 반환하고 broker, journal writer, 운영 토글과 exit decision
authority를 직접 소유해서는 안 된다 (MUST NOT).
단순 caller boolean은 lane activation/evaluation authority가 아니며 (MUST NOT), a072 운영 adapter가
도입되기 전에는 package 외부에서 mint할 수 없는 dormant sealed authorization이 없으면 항상
`LANE_OFF`여야 한다 (SHALL).

#### Scenario: 두 weekly lane의 동시 제공
- **WHEN** change를 포함한 lane registry conformance를 검사한다
- **THEN** OpenDART KR lane과 EDGAR US lane의 고유 ID/version이 모두 존재하며 한쪽만 등록된 build는 검증에 실패한다

#### Scenario: 한 시장 disclosure 실패의 독립성
- **WHEN** OpenDART evidence가 missing/stale이고 EDGAR evidence는 유효하다
- **THEN** KR 신규 decision과 mutation은 0건이며 US 평가 및 공통 exit engine은 계속된다

#### Scenario: lane invalidation과 공통 exit
- **WHEN** value thesis invalidation과 공통 exit engine의 stop 조건이 함께 관측된다
- **THEN** lane은 typed invalidation만 반환하고 exit decision/order를 만들지 않으며 공통 exit engine은 독립적으로 risk-reducing decision을 발급할 수 있다

### Requirement: weekly value 입력은 point-in-time filing과 model schema를 완전하게 보존한다
Weekly value evaluator는 strict versioned filing/revision/model schema와 checked fixed-point arithmetic만 사용해야 한다 (SHALL).
입력은 market, symbol/issuer identity, source, filing/report ID, revision ID, superseded revision ID,
filing `as_of`, `observed_at`, `ingested_at`, `evaluated_at`, `fresh_until`, ISO-4217 currency, monetary
unit/scale, diluted shares와 unit, dilution event facts, normalized financial input vector와 units,
model ID/version, model config/threshold digest, canonical evidence digest, fair-value minor units와 exact
arithmetic preimage를 포함해야 한다 (SHALL). `as_of <= observed_at <= ingested_at <= evaluated_at <=
fresh_until`을 위반하거나 revision chain, currency/unit, diluted shares, dilution facts, model/config
digest 또는 checked arithmetic이 누락·unknown·overflow이면 typed refusal이어야 한다 (SHALL).
Evaluator는 later revision이나 future-ingested fact를 이전 cutoff에 사용하거나 float/local fallback으로
fair value를 재구성해서는 안 된다 (MUST NOT).
Strict decoder가 만든 evidence는 전체 immutable revision/PIT/dilution/financial/model preimage의
private seal을 가져야 하고 (SHALL), caller literal 또는 decode 뒤 field mutation은 fail-closed
`STRICT_SCHEMA_INVALID`여야 한다 (SHALL).

#### Scenario: 미래 revision 누출
- **WHEN** evaluation cutoff 뒤 ingested된 correction을 과거 campaign leg 평가에 사용하려 한다
- **THEN** point-in-time refusal을 반환하고 이전 decision/fair value를 다시 쓰지 않는다

#### Scenario: currency 또는 diluted shares 누락
- **WHEN** monetary input currency/unit 또는 current diluted-share evidence를 확정할 수 없다
- **THEN** 해당 market weekly entry는 typed schema refusal이고 quantity와 broker mutation은 0건이다

#### Scenario: model arithmetic replay
- **WHEN** 같은 canonical filing vector, diluted shares와 model/config digest를 replay한다
- **THEN** checked fixed-point fair-value minor units와 decision digest가 동일하고 float rounding은 사용되지 않는다

### Requirement: weekly market-week reservation은 공식 calendar 기준으로 원자 관리된다
Weekly lane은 official exchange calendar와 market IANA timezone이 발급한 `stable_market_week_identity`에 원자 unique reservation을 사용해야 한다 (SHALL).
Server-local time, 단순 7일 duration 또는 비공식 ISO-week fallback을 사용해서는 안 된다
(MUST NOT). Canonical reservation key는 `(campaign_id, market, stable_market_week_identity)`여야
하고 (SHALL), next planned leg ordinal, calendar generation/digest evidence와 idempotency key를 보존해야
한다 (SHALL). Calendar generation은 identity 도출 evidence일 뿐 key의 일부가 아니며
(MUST NOT), 동일한 official market week의 calendar correction이 generation A→B로 바뀌어도
stable identity와 allowance는 변하지 않아야 한다 (SHALL). 같은 canonical key의 concurrent
명령 중 하나만 commit되어야 한다 (SHALL). Cumulative positive fill이 한 주의 reservation을
영구 `CONSUMED`로 만들고 (SHALL), fill 0이 authoritative하게 증명된 cancel/expiry만 `RELEASED`로
전환할 수 있다 (SHALL). Duplicate retry는 기존 reservation/result를 반환하고 새 leg나 새 weekly
allowance를 만들 수 없다 (MUST NOT).
Stable identity는 market/exchange prefix와 IANA-zone Monday `session_date`의 exact ISO year/week에서
도출한 값과 byte-for-byte 같아야 하고 (SHALL), reservation CAS version과 count는
`(campaign_id, market)` scope별로 독립이어야 한다 (SHALL). Reserve freshness는 command에 봉인된
trusted `evaluated_at`을 기준으로 검사해야 하며 calendar 자신의 `observed_at`으로 대체해서는 안 된다
(MUST NOT).

#### Scenario: 같은 market week 동시 예약
- **WHEN** 같은 campaign/week에 두 worker가 next leg를 동시에 예약한다
- **THEN** 하나만 commit되고 다른 요청은 기존 idempotent result 또는 conflict를 받으며 두 번째 active reservation은 없다

#### Scenario: positive partial fill 뒤 cancel
- **WHEN** weekly leg가 한 주에 1주 이상 fill된 뒤 잔량이 cancel된다
- **THEN** 해당 week는 CONSUMED로 유지되고 같은 market/stable_market_week_identity에 다른 leg를 예약할 수 없다

#### Scenario: zero-fill expiry
- **WHEN** reservation의 주문이 fill 0으로 authoritative expiry되고 pending attempt도 없다
- **THEN** reservation은 멱등 RELEASED되고 leg count는 증가하지 않으며 동일 retry가 새 leg를 만들지 않는다

#### Scenario: holiday와 US DST 경계
- **WHEN** holiday-shortened KR week 또는 US DST 전환 주의 session_date를 평가한다
- **THEN** official calendar가 도출한 stable market-week identity로 한도를 판정하고 process timezone 변화는 결과를 바꾸지 않는다

#### Scenario: calendar generation correction A→B
- **WHEN** reservation 후 official calendar correction이 generation A를 B로 대체하지만 동일한 market의 동일한 official week이다
- **THEN** stable_market_week_identity는 같고 B evidence는 lineage에 append될 뿐이며 concurrent/replay 요청은 두 번째 weekly slot을 얻지 못한다

#### Scenario: crash/restart 예약 replay
- **WHEN** reservation commit 또는 fill-consume 전후에 crash한 뒤 journal을 replay한다
- **THEN** 동일 campaign/market/stable-week/leg reservation state와 idempotency result가 복원되고 calendar generation이 바뀌어도 중복 weekly allowance가 없다

### Requirement: weekly campaign은 최대 일곱 planned leg와 immutable allocation을 지킨다
Weekly campaign은 생성 시 immutable campaign risk budget, planned quantity `Q`, 일곱 leg ceiling과 allocation policy/config digest를 영속해야 한다 (SHALL).
Leg count는 positive cumulative fill로 `CONSUMED`된 서로 다른 planned ordinal의 수이며 최대 7이어야
한다 (SHALL). Zero-fill RELEASED reservation, submit retry와 duplicate fill은 leg count를 늘려서는
안 된다 (MUST NOT). 각 q_leg는 해당 ordinal의 immutable planned remaining과 a066 current q_final
이하이어야 하며 (SHALL), actual fill, cancel, expiry 또는 이전 ordinal의 미사용량으로 현재/후속
planned ceiling을 상향 재배분해서는 안 된다 (MUST NOT). 각 신규 reservation 전 value thesis,
latest eligible revision, dilution, structure와 projected risk를 다시 검증해야 한다 (SHALL).
신규 reservation/admission은 account valuation currency minor units에서 cumulative filled risk,
unresolved held risk와 proposed leg risk의 checked sum이 campaign risk budget 이하인 경우에만
허용되어야 한다 (SHALL). 각 positive fill의 filled risk는
`max(transferred_conservative_reservation_minor, ceil_minor(qty * max(entry_minor -
effective_stop_minor, 0) + entry_fees_minor + estimated_exit_fees_levies_minor))`여야 하며
(SHALL), held/proposed risk는 a066 conservative reservations여야 한다 (SHALL). Filled usage는
transferred conservative reservation보다 낮아져서는 안 된다 (MUST NOT). US는 a066 decision에
봉인된 동일한 공식 FX quote ID/as-of/quote-to-account rate direction과 conservative haircut을
사용해 account valuation currency로 환산한 뒤 minor-unit ceil해야 하며 (SHALL), reward/risk와
다른 FX snapshot을 섞어서는 안 된다 (MUST NOT). 모든 산술은 checked이어야 하고 (SHALL),
overflow·누락·FX provenance 불일치는 신규 admission을 typed refusal해야 한다 (SHALL).
이미 발생한 fill의 계산 불가 또는 actual budget 초과는 weekly fill/CONSUMED 처리를 막지
않고 `CAMPAIGN_RISK_OVERAGE`나 unknown-risk latch로 모든 후속 신규 leg를 차단해야 한다
(SHALL). Overage를 released/unused planned quantity로 상쇄해서는 안 된다 (MUST NOT).
Positive fill의 reservation consume과 actual-risk transfer/latch는 하나의 pure aggregate transition만
제공해야 하며 (SHALL), reservation-only API는 positive fill consume을 거부해 두 상태 사이 crash
window를 만들 수 없어야 한다 (SHALL).
Risk state의 plan digest, filled/held minor balances, latches와 applied-fill receipts는 private canonical
seal로 결속해야 하며 (SHALL), caller field mutation이나 copied-map clear 뒤 admission/fill apply는
fail-closed risk refusal이어야 한다 (SHALL). 외부에는 scalar/read-only 조회만 제공해야 한다 (SHALL).

#### Scenario: 일곱 positive-fill leg 이후 요청
- **WHEN** campaign에 일곱 distinct planned ordinal이 positive fill로 CONSUMED되었다
- **THEN** 다음 reservation은 PLAN_EXHAUSTED refusal이고 quantity와 mutation은 0건이다

#### Scenario: zero-fill은 leg를 소비하지 않는다
- **WHEN** ordinal 3 reservation이 authoritative zero-fill RELEASED다
- **THEN** positive-fill leg count는 변하지 않고 ordinal 4로 건너뛰거나 unused quantity를 다른 ordinal에 더하지 않는다

#### Scenario: q_leg cap property
- **WHEN** 임의 planned ceilings, q_final, partial fills, retries, cancel/expiry와 restart sequence를 적용한다
- **THEN** 모든 q_leg는 immutable ordinal ceiling과 a066 q_final 이하이고 admitted filled/held/proposed monetary risk는 immutable campaign budget 이하이며 duplicate fill은 risk를 다시 늘리지 않고 zero-fill cancel/expiry는 미체결 held만 release하며 upward reallocation이 없다

#### Scenario: weekly conservative reservation floor
- **WHEN** partial fill의 exact stop-distance와 비용 산식이 held에서 이전된 conservative reservation보다 작다
- **THEN** filled risk는 transferred reservation을 유지하고 duplicate fill replay는 usage를 중복 증가시키지 않는다

#### Scenario: weekly US frozen FX mismatch
- **WHEN** US fill risk가 a066에 봉인된 FX quote/as-of/direction/haircut과 다른 snapshot으로 환산된다
- **THEN** 신규 admission은 refusal되고 이미 발생한 fill은 보존된 채 unknown-risk latch가 후속 weekly leg를 차단한다

#### Scenario: weekly actual fill risk overage
- **WHEN** positive fill의 actual price, fees 또는 FX 때문에 recalculated cumulative filled risk가 immutable campaign budget을 초과한다
- **THEN** fill과 weekly CONSUMED state는 유지되고 overage latch가 후속 market-week reservation/leg를 차단하며 common exit는 계속된다

#### Scenario: fill-time monetary risk 근거 부재
- **WHEN** positive fill의 authoritative fee 또는 FX evidence가 없어 campaign currency risk를 계산할 수 없다
- **THEN** fill과 leg count는 적용하되 risk를 0으로 추정하지 않고 unknown-risk latch로 후속 신규 leg를 차단한다

### Requirement: weekly value target과 RR은 공통 안전 출구를 약화시키지 않는다
Weekly entry target은 `min(staged_target_minor, fair_value_minor)`로 계산해야 한다 (SHALL).
Account valuation currency minor units에서 `costs_minor = entry_costs_minor +
estimated_exit_costs_levies_minor`, `reward_minor = max(target_minor-entry_minor, 0)*qty -
costs_minor`, `risk_minor = max(entry_minor-effective_stop_minor, 0)*qty + costs_minor`,
`rr_ppm = floor(reward_minor*1_000_000/risk_minor)`로 계산해야 한다 (SHALL). 모든
곱셈·덧셈·뺄셈은 checked fixed-point여야 하고 (SHALL), `risk_minor <= 0`이거나
overflow, cost/FX provenance 부재는 typed refusal이어야 한다 (SHALL). `rr_ppm >=
minimum_rr_ppm`이면 inclusive accept하고 그보다 작으면 refusal해야 한다 (SHALL). US RR은
filled/held/proposed campaign risk와 동일하게 a066 decision에 봉인된 공식 FX quote ID/as-of,
quote-to-account rate direction, conservative haircut과 minor-unit ceil snapshot을 reward, risk, entry/exit
cost에 모두 사용해야 하며 (SHALL), 서로 다른 FX snapshot을 섞어서는 안 된다 (MUST NOT).
신규 leg는 effective stop을 불리하게 이동시켜서는 안 되고
(MUST NOT), 필요한 structural stop이 a066 cap보다 넓으면 stop을 임의로 좁히지 말고 거부해야 한다
(SHALL). Lane target/invalidation은 common stop, emergency exit, reconciliation, protection 또는
공통 exit engine을 지연하거나 억제해서는 안 된다 (MUST NOT).
Structural stop candidate는 version, source/policy digest, observed/fresh-until과 private seal을
보존해야 하고 (SHALL), evaluation cutoff에서 stale하거나 변조된 후보는 거부해야 한다 (SHALL).

#### Scenario: fair value가 staged target보다 낮다
- **WHEN** staged target은 1200 minor units이고 fair value는 1100이다
- **THEN** entry target은 정확히 1100이며 더 높은 staged target으로 RR을 부풀리지 않는다

#### Scenario: capped target의 minimum RR 실패
- **WHEN** fair-value cap과 비용을 적용한 expected RR이 configured minimum보다 작다
- **THEN** typed RR refusal을 반환하고 entry/add decision과 mutation은 0건이다

#### Scenario: minimum RR inclusive boundary
- **WHEN** exact checked formula의 `rr_ppm` 값이 `minimum_rr_ppm`과 정확히 같다
- **THEN** RR 경계는 통과하며 float 또는 다른 rounding mode로 refusal하지 않는다

#### Scenario: zero or negative risk denominator
- **WHEN** checked `risk_minor` 결과가 0 이하이거나 산술 overflow가 발생한다
- **THEN** typed RR_CALCULATION_INVALID refusal을 반환하고 entry/add decision은 0건이다

#### Scenario: reward/risk FX snapshot 불일치
- **WHEN** US reward와 risk 또는 entry/exit cost가 서로 다른 FX snapshot으로 환산된다
- **THEN** RR을 계산하지 않고 typed provenance refusal로 신규 leg를 차단한다

#### Scenario: common emergency exit
- **WHEN** fair-value target 전 common stop 또는 emergency exit 조건이 참이다
- **THEN** lane은 exit decision을 만들지 않고 신규 leg를 0건으로 만들며 공통 exit engine은 독립적으로 즉시 진행한다

### Requirement: weekly value 결과는 disclosure와 시장 lineage를 보존한다
모든 weekly value decision과 refusal은 disclosure와 market-scoped lineage를 보존해야 한다 (SHALL).
결과는 market, lane ID/version, candidate ID, filing/revision/model/config/evidence digests, calendar
generation/digest evidence, stable_market_week_identity, reservation ID, campaign ID, position generation, immutable risk budget,
planned leg ordinal/ceiling을 포함해야 하고 (SHALL), correction이 원본 lineage를 덮어쓰거나 KR와
US attribution을 결합해서는 안 된다 (MUST NOT).

#### Scenario: 정정 공시 재평가
- **WHEN** 이전 filing correction이 새 evidence version으로 수집된다
- **THEN** 새 결과는 correction과 원본 filing lineage를 연결하고 이전 point-in-time decision을 변경하지 않는다

### Requirement: production strategy dispatch는 router와 campaign lineage를 보존한다
Strategy engine은 production dispatch에서 router와 campaign lineage를 보존해야 한다 (SHALL). ApprovedCandidate와 immutable evidence를 market/horizon router에 전달하고
선택된 lane의 순수 EntryDecision 또는 typed refusal을 production runtime에 반환해야 한다
(SHALL). 결정은 market, candidate/evidence digest, router/version, lane/version과 campaign/leg
identifier를 포함해야 하며 (SHALL), lane과 router는 Guardian, broker, journal mutation 또는
운영 토글 writer를 직접 호출해서는 안 된다 (MUST NOT).

#### Scenario: US reversal lane 선택
- **WHEN** current US ApprovedCandidate가 router를 거쳐 reversal lane에서 수락된다
- **THEN** EntryDecision은 US market, router/lane version과 campaign/leg lineage를 포함하고 broker call은 아직 0건이다

#### Scenario: router refusal
- **WHEN** candidate horizon과 활성 lane binding이 일치하지 않는다
- **THEN** typed router refusal을 반환하고 lane evaluation, Guardian과 broker request는 0건이다

#### Scenario: 순수 lane의 mutation 의존성
- **WHEN** strategy lane 또는 router dependency closure를 검사한다
- **THEN** broker mutator, journal writer와 operating-setting writer가 존재하지 않는다
