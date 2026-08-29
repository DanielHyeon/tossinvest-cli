## Context

Weekly value 판단은 filing revision의 point-in-time 재현성과 market-week uniqueness가 핵심이다.
OpenDART와 EDGAR는 a064에서 canonical evidence가 되며 a065/a066이 campaign state와 cap을 소유한다.
두 market lane은 같은 release에 추가되지만 source failure와 runtime enablement는 독립이다. 현재
runtime은 Parker-only OFF/UNOBSERVED이며 이 change는 자동 활성화하지 않는다.

## Goals / Non-Goals

**Goals:**

- filing/revision/dilution/model/fair-value strict schema와 fixed-point replay를 고정한다.
- official exchange calendar/IANA market-week의 atomic unique reservation을 정의한다.
- seven positive-fill leg count와 immutable planned allocation을 crash/retry에도 보존한다.
- capped target, minimum RR, typed invalidation과 common exit authority를 분리한다.

**Non-Goals:**

- raw OpenDART/EDGAR 호출, secret 발급 또는 lane 내부 model 학습
- broker dispatch, journal writer, activation, LIVE approval 또는 lane-generated exit
- server-local ISO week나 symbol/time 추정

## Decisions

### 1. Adapter output은 완전한 filing/model preimage다

Filing/revision chain, all point-in-time timestamps, currency/unit, diluted shares/dilution facts,
normalized input vector, model/config digest와 fixed-point fair value를 요구한다. Later revision을 과거
cutoff에 섞거나 float/local fallback을 사용하는 대안은 미래정보 누출과 replay drift 때문에
배제한다.

### 2. market-week allowance는 durable atomic reservation이다

Official exchange calendar가 IANA timezone에서 발급한 stable market-week identity를 사용하고
canonical reservation key를 `(campaign_id, market, stable_market_week_identity)`로 고정한다.
Calendar generation/digest는 identity 도출 evidence이지 key의 일부가 아니다. 동일한 official market
week의 calendar correction이 generation A→B로 바뀌어도 stable identity는 바뀌지 않으며 두 번째
allowance를 만들 수 없다. Any positive fill은 terminal CONSUMED, authoritative zero-fill
cancel/expiry만 RELEASED다. Duplicate idempotency retry는 기존 result를 반환한다. In-memory timer,
server-local ISO week나 단순 7일 duration은 crash/correction/holiday/DST에서 drift하므로 배제한다.

### 3. seven-leg count와 allocation은 planned basis다

Leg count는 positive-fill CONSUMED planned ordinal 수다. Zero-fill release와 retry는 count를 늘리지
않는다. Campaign risk budget을 account valuation currency minor units로, Q와 seven ceilings를
생성 시 고정하며 actual fill이나 unused quantity로 다른 ceiling을 키우지 않는다. 모든 q_leg는
planned remaining과 a066 q_final 이하이다. Positive fill의 filled risk는
`max(transferred_conservative_reservation_minor, ceil_minor(qty * max(entry_minor -
effective_stop_minor, 0) + entry_fees_minor + estimated_exit_fees_levies_minor))`로 고정한다.
US는 a066에 봉인된 동일한 공식 FX quote ID/as-of/rate direction/haircut으로 account
valuation minor units를 ceil한다. Filled usage는 transferred reserve 아래로 내려가지 않고
held/proposed는 a066 conservative reservations이며, checked 합은 immutable budget 이하여야 한다.
Overflow·provenance 부재는 신규 admission을 refusal하고, fill-time price·fees·FX가 overage/unknown을
만들면 fill/leg count는 보존하고 후속 market-week leg를 latch한다.

### 4. target/RR와 exit 권위는 분리한다

Entry target은 exact `min(staged_target_minor, fair_value_minor)`이다. Account valuation minor units에서
`reward_minor = max(target_minor-entry_minor, 0)*qty -
(entry_costs_minor+estimated_exit_costs_levies_minor)`, `risk_minor =
max(entry_minor-effective_stop_minor, 0)*qty +
(entry_costs_minor+estimated_exit_costs_levies_minor)`, `rr_ppm =
floor(reward_minor*1_000_000/risk_minor)`를 checked arithmetic으로 계산한다. `risk_minor <= 0`이거나
overflow/FX inconsistency면 refusal하고, `rr_ppm >= minimum_rr_ppm`일 때만 inclusive accept한다. US는
campaign risk와 동일한 a066 frozen official FX snapshot/direction/haircut/ceil을 reward, risk, 비용에
일관되게 적용한다. Lane은 invalidation/refusal만 emit하고 common exit engine이 stop/emergency
exit를 독립 발급한다. Fair value는 exit ceiling이지 protection 지연 조건이 아니다.

## Risks / Trade-offs

- [correction이 과거 판단을 달라 보이게 함] → append-only revision chain과 original cutoff/digest를 보존한다.
- [zero-fill release와 동시 retry 경쟁] → versioned reservation CAS와 idempotency key로 하나만 commit한다.
- [holiday/DST/correction week 경계 drift] → official calendar stable market-week identity 외 fallback을 금지하고 generation은 evidence로만 보존한다.
- [fair-value cap으로 minimum RR 실패] → entry를 거부하고 target을 임의로 높이지 않는다.
- [actual fill monetary risk가 budget 초과] → overage provenance를 영속하고 후속 entry만 차단하며 common exit를 계속한다.

## Migration Plan

1. Strict filing/model schema와 official market-week fixture를 추가한다.
2. Reservation/leg-count/allocation RED tests 뒤 두 lane을 같은 build에 OFF 등록한다.
3. correction, concurrency, crash/restart, holiday/DST, target/RR와 invalidation/common-exit separation을 검증한다.
4. 외부 credential 부재는 해당 market refusal로 격리하고 rollback은 lane 등록만 제거한다.

## Open Questions

없음. Source credential과 runtime approval은 a064/a072 운영 절차가 소유한다.
