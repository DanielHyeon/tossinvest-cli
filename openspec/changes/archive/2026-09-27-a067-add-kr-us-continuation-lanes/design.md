## Context

기존 strategy engine은 KR Parker lane만 가지며 production seam은 미배선, runtime은
`UNOBSERVED`, lane은 OFF다. 이 change는 a064 evidence, a065 campaign/leg와 a066 risk cap 위에
KR flow와 US participation continuation을 같은 release로 추가한다. 두 시장의 구현·평가·상태는
독립이며 lane 등록은 runtime activation이나 LIVE authority를 만들지 않는다.

## Goals / Non-Goals

**Goals:**

- strict versioned KR flow/US participation schema와 exact integer arithmetic을 고정한다.
- immutable campaign risk budget에서 8:4:2 planned allocation을 결정적으로 만든다.
- partial fill/retry/cancel이 allocation을 위로 재배분하지 못하게 한다.
- lane invalidation과 common exit authority를 분리하고 KR/US lineage를 보존한다.

**Non-Goals:**

- raw evidence 수집, campaign/risk persistence 재구현 또는 broker dispatch
- lane에서 exit decision, journal mutation, 운영 토글이나 LIVE approval 발급
- KR 결과를 US readiness 조건으로 사용하거나 그 역을 수행

## Decisions

### 1. 두 시장은 별도 strict schema와 lane version을 가진다

KR flow는 notional minor units와 turnover에서 signed ppm을, US participation은 share volume과
price minor units에서 participation/price-change ppm을 만든다. 모든 timestamp, unit, threshold와
config digest를 입력에 포함하고 checked integer arithmetic만 허용한다. Generic metric map이나
float comparison은 source/unit drift를 숨기므로 배제한다.

### 2. 8:4:2는 immutable planned basis에서 한 번만 계산한다

Campaign 생성 시 account valuation currency minor units의 risk budget, per-share risk와 a066이
허용한 planned quantity Q를 고정한다. `floor(Q*8/14)`, `floor(Q*4/14)`, final remainder로 ceiling을
만든다. Actual fill은 risk usage에는 반영되지만 planned ceiling을 다시 계산하거나 unfilled 수량을
후속 leg에 더하지 않는다. 각 요청은 planned remaining과 current a066 q_final의 min을 넘지 않는다.

각 positive fill의 filled risk는 account valuation minor units에서
`max(transferred_conservative_reservation_minor, ceil_minor(qty * max(entry_minor -
effective_stop_minor, 0) + entry_fees_minor + estimated_exit_fees_levies_minor))`로 동결한다. US
금액은 a066 admission에 봉인된 동일한 공식 FX quote ID/as-of/rate direction과 conservative haircut을
사용해 account valuation currency로 환산하고 minor-unit ceil한다. Filled usage는 이전한 conservative
reservation보다 작아질 수 없으며 held/proposed usage는 a066 conservative reservations다. Admission은
이 세 checked 합이 immutable budget 이내인지 검사한다. Overflow나 입력/FX provenance 부재는 신규
admission을 refusal하고, 이미 발생한 fill의 actual risk가 budget을 넘거나 계산 불가면 fill은 적용하되
overage/unknown-risk latch가 후속 leg를 막는다. Fill 기반 재최적화나 unused quantity 상계는 뒤 leg
노출을 키우므로 제외한다.

### 3. lane은 invalidation만 emit하고 common exit engine이 exit를 소유한다

Lane은 entry/add decision 또는 typed invalidation/refusal만 반환한다. Stop, emergency exit와
risk-reducing decision은 공통 exit engine이 독립적으로 판단한다. Lane이 exit order를 만들거나
exit engine이 lane cycle을 기다리는 구조는 두 권위와 지연 경로를 만들므로 배제한다.

### 4. 상태와 mutation은 lane 밖에 둔다

Evaluator는 immutable campaign/evidence snapshot만 받고 결과를 반환한다. a065가 idempotency와
fill history를, a066이 quantity cap을 소유한다. 두 lane은 같은 registry unit에 OFF로 등록한다.

## Risks / Trade-offs

- [strict schema가 provider drift 때 진입을 줄임] → typed refusal로 해당 market만 fail closed하고 peer evaluation은 유지한다.
- [integer floor가 작은 Q에서 앞 leg를 0으로 만듦] → 0 leg를 합성 확대하지 않고 versioned plan refusal/skip 규칙으로 처리한다.
- [slippage/fees/FX가 admitted risk를 초과] → actual fill을 보존하고 overage/unknown-risk latch로 후속 exposure만 차단하며 common exit를 유지한다.
- [common exit와 invalidation이 동시에 발생] → lane은 add 0건만 보장하고 exit engine이 독립 권위로 진행한다.
- [동시 release가 결함 범위를 넓힘] → runtime enablement는 시장별 OFF로 분리하고 peer fixture를 같은 gate에서 검증한다.

## Migration Plan

1. 두 schema/evaluator와 allocation fixtures를 같은 build에 추가한다.
2. 두 lane을 registry에 desired/effective OFF로 함께 등록한다.
3. schema arithmetic, floor/remainder, no-reallocation, cap, invalidation/exit separation과 KR/US isolation을 검증한다.
4. a070/a072 전에는 dispatch에 연결하지 않으며 rollback은 두 registry entry만 제거하고 common exit engine을 변경하지 않는다.

## Open Questions

없음. Activation과 protection readiness는 a071/a072가 소유한다.
