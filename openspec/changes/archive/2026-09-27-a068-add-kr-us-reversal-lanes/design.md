## Context

KR 수급 흡수와 US dislocation은 다른 evidence 의미를 가지며 단순 가격 하락과 구분되어야 한다.
현재 runtime은 Parker-only OFF/UNOBSERVED이고 production seam이 없다. 이 change는 a064–a066의
evidence/campaign/risk 계약 위에 두 reversal lane을 동시에 추가하되 activation은 만들지 않는다.

## Goals / Non-Goals

**Goals:**

- KR absorption/US dislocation strict schema, units, freshness, config digest와 integer arithmetic을 고정한다.
- sweep→break→reclaim의 causal order와 bounded window를 강제한다.
- immutable 2:4:8 allocation, a066 cap, stop non-retreat와 no-upward-reallocation을 보장한다.
- typed invalidation과 common exit authority를 분리하고 market lineage를 보존한다.

**Non-Goals:**

- 가격 하락만을 근거로 한 averaging down
- raw evidence adapter, portfolio risk owner, broker dispatch 또는 persistence 재구현
- lane의 exit decision, 운영 토글, activation 또는 LIVE approval

## Decisions

### 1. 시장별 schema는 metric과 structural event를 명시한다

KR은 absorbed/aggressive-sell notional minor units에서 absorption ppm을, US는 reference/low price와
share volume에서 drawdown/relative-volume ppm을 checked integer로 계산한다. Event와 metric은 같은
scope, schema/config digest와 `effective_at <= observed_at <= ingested_at <= evaluated_at <=
fresh_until`의 exact timestamp lineage를 가진다. 등호 경계는 유효하지만 evaluated_at이
fresh_until을 1 tick이라도 넘거나 중간 시각이 역전되면 진입을 refusal한다. Generic score map은
단위/threshold를 감추므로 사용하지 않는다.

### 2. final leg는 causal bounded conjunction이다

`sweep_at <= break_at <= reclaim_at <= evaluated_at`과 configured structural window를 모두
검증한다. 세 event 중 하나의 stale/missing/scope mismatch를 다른 score로 상쇄하지 않는다.
Weighted score 대안은 가격 하락만으로 final leg가 열릴 수 있어 배제한다.

### 3. 2:4:8은 immutable planned basis다

Campaign risk budget을 account valuation currency minor units로, Q를 함께 첫 leg 전에
고정하고 `floor(Q*2/14)`, `floor(Q*4/14)`, final remainder로 planned ceilings를 만든다.
Actual fill/cancel/retry는 ceiling을 다시 계산하거나 미사용 수량을 옮기지 않는다. q_leg는
planned remaining과 a066 q_final 이하이다.

Positive fill의 filled risk는 account valuation minor units의
`max(transferred_conservative_reservation_minor, ceil_minor(qty * max(entry_minor -
effective_stop_minor, 0) + entry_fees_minor + estimated_exit_fees_levies_minor))`로 고정한다.
US는 a066이 admission에 봉인한 공식 FX quote ID/as-of/rate direction과 conservative haircut을
동일하게 사용해 minor-unit ceil한다. Filled usage는 transferred reservation 이하로 내려가지
않고 held/proposed는 a066 conservative reservations를 사용하며, 세 금액의 checked sum이
immutable budget 이하여야 한다. Overflow·provenance 부재는 신규 admission refusal이고,
actual price·fees·FX overage/unknown은 fill을 보존한 채 후속 leg를 latch한다.

### 4. invalidation과 exit authority를 분리한다

Evaluator는 invalidation 시 add를 0건으로 만드는 typed result만 emit한다. Common exit engine은
lane과 독립적으로 stop/emergency reduction을 발급한다. Lane-generated exit 대안은 B안의 단일
exit authority를 깨므로 채택하지 않는다.

## Risks / Trade-offs

- [strict causal conjunction으로 기회 감소] → fail-closed를 유지하고 threshold/window 변경은 새 config/lane version으로만 수행한다.
- [feed latency로 false refusal 증가] → event별 observed/ingested/fresh-until과 bounded window를 lineage에 남긴다.
- [partial fill이 2:4:8 모양을 바꿈] → planned basis는 고정하고 actual fill은 exposure usage에만 반영한다.
- [actual fill monetary risk가 budget 초과] → overage/unknown-risk latch로 후속 exposure를 막고 common exit는 계속한다.
- [한 시장 adapter 결함] → market OFF/typed refusal로 격리하고 peer lane을 중단하지 않는다.

## Migration Plan

1. 두 strict schema와 causal-order fixtures를 먼저 추가한다.
2. 두 reversal evaluator와 immutable allocation을 같은 build에 구현하고 OFF 등록한다.
3. order/window, floor/remainder, cap, no-reallocation, invalidation/common-exit separation과 market isolation을 검증한다.
4. rollback은 두 lane 등록만 제거하고 common exit/reconciliation 경로는 유지한다.

## Open Questions

없음. 운영 activation은 후속 high-risk change가 사람 승인으로 소유한다.
