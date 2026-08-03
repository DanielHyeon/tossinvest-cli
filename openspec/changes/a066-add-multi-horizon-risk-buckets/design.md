## Context

기존 Guardian은 개별 주문, 총 노출과 일손실을 결정+예약 transaction으로 강제한다. 후속 lane는 KR/US와 short/medium horizon에서 여러 leg를 만들므로, 개별 주문이 기존 한도 안이어도 합산 집중도가 커질 수 있다. 같은 symbol을 두 lane가 동시에 획득하면 exit 책임과 scale-in 예약도 갈라진다.

이 change는 기존 Guardian 체인과 예약 권위를 약화하지 않고 그 뒤에 더 보수적인 다차원 cap을 추가한다. a065 campaign identity를 소비하지만 campaign core나 lane별 비율을 소유하지 않는다.

## Goals / Non-Goals

**Goals:**

- horizon, market, sector와 symbol별 exposure를 합산해 수량을 보수적으로 cap한다.
- 최종 수량 `q_final`이 lane가 제안한 `q_candidate`를 절대 초과하지 않게 한다.
- symbol/market의 owning lane와 campaign을 원자 예약한다.
- scale-in, partial fill, retry와 restart에도 예약·실노출을 중복 없이 재구성한다.
- entry loss lock과 bucket 장애가 위험 감소 경로를 막지 못하게 한다.

**Non-Goals:**

- 전략별 목표 수량·leg 비율 산출
- 운영 한도 수치의 자동 완화 또는 market 간 자본 자동 이전
- short, leverage 또는 비공식 FX source 도입
- live lane/automation 활성화

## Decisions

### D1. Bucket은 금액 단위의 직교 dimension 교집합이다

모든 exposure-raising 요청은 horizon(`SHORT|MEDIUM`), market(`KR|US`), strategy risk bucket ID/version, sector와 symbol bucket key를 가져야 한다. 각 bucket은 account base currency의 monetary limit, filled monetary exposure, HELD monetary reservation, valuation provenance, snapshot version과 freshness를 제공한다. unknown horizon/market/strategy/sector 또는 환산 불능은 별도 무제한 bucket으로 보내지 않고 진입을 거부한다. strategy bucket은 개별 lane ID와 같을 수도 있지만 여러 lane을 한 정책으로 묶을 수 있는 server-owned versioned risk identity이며 caller 문자열을 그대로 신뢰하지 않는다.

각 dimension을 가중 합산하는 대안은 한 dimension 초과가 다른 여유로 상쇄되므로 기각한다.

### D2. Monetary reservation은 worst price, fees와 FX haircut을 사용한다

buy quantity `q`의 예약 금액은 account base currency minor unit에서 다음 순수 decimal 함수로 계산한다.

`reserve(q) = ceil_minor(q × worst_executable_price_quote × fx_rate_quote_to_base × fx_haircut_multiplier + worst_case_fees_base(q))`

`worst_executable_price_quote`는 공식 주문 계약상 해당 intent가 체결될 수 있는 보수적 최고 가격이며 source/version/observed-at을 가진다. `fx_haircut_multiplier`는 1 이상이고 같은 통화이면 rate와 multiplier가 1이다. fee 함수와 FX policy도 version/digest를 가진다. 입력이 missing/stale/non-positive이거나 haircut이 1 미만, arithmetic overflow 또는 minor-unit ceil이 불가능하면 진입을 거부한다. caller price, mid-price 또는 fee 0 fallback은 사용하지 않는다.

각 bucket cap은 `reserve(q) <= monetary_remaining`을 만족하는 최대 non-negative 정수 q다. fee가 비선형일 수 있으므로 단순 per-share 나눗셈을 권위로 쓰지 않고 bounded monotone search 또는 동등한 exact algorithm을 사용한다.

### D3. q_final은 q_candidate와 모든 금액 cap의 최솟값이다

canonical field는 lane 제안 `q_candidate`와 최종 `q_final`이다. `q_final = min(q_candidate, q_existing_guardian, q_horizon, q_market, q_strategy, q_sector, q_symbol)`이며 각 cap은 위 monetary reservation 함수로 도출한 non-negative integer다. overflow, stale snapshot, price/fee/FX/currency 부재는 `q_final=0` typed refusal이다. 이 계층은 수량을 늘리는 multiplier를 갖지 않는다.

수량 0은 주문 intent가 아니라 typed refusal이다. `q_candidate`, `q_final`, reservation input/digest, 각 binding monetary cap과 산출 수량을 decision evidence로 저장한다.

### D4. q_final 확정 뒤 GuardianDecision과 예약을 원자 발급한다

기존 Guardian chain은 mutation 없는 precheck와 `q_existing_guardian` cap만 산출한다. bucket calculator가 `q_final`과 각 monetary reservation을 확정한 뒤에만 `(account, market, symbol, prospective-or-actual position generation)` unique owner row, 모든 HELD monetary reservation과 **q_final을 봉인한 GuardianDecision**을 expected snapshot version 아래 하나의 transaction으로 commit한다. q_candidate나 precheck quantity로 GuardianDecision을 먼저 기록해서는 안 된다. 같은 lane/campaign의 후속 leg만 기존 owner를 재사용할 수 있고 경쟁 lane는 전부 rollback된다.

사전 in-memory lock은 다중 process/restart를 막지 못하므로 권위로 사용하지 않는다. a065 prospective generation token은 first fill 전 owner key가 되고 실제 successor generation에 set-once 결합된다.

### D5. filled와 HELD monetary exposure를 함께 계상한다

bucket usage는 authoritative Position/fill projection에 귀속된 filled monetary exposure와 미해소 HELD monetary reservation의 합이다. 각 deduplicated positive fill delta의 Position apply transaction은 horizon, market, strategy, sector와 symbol 모든 적용 bucket에 대해 원 reservation policy/rounding으로 계산한 `transfer_delta = proportional_reserved_allocation(new_cumulative_fill) - previously_transferred`와 실제 fill price, allocated fee 및 fill-time persisted FX에 따른 `actual_delta`를 함께 기록한다. 해당 fill의 `filled_delta = max(transfer_delta, actual_delta)`이고 `transfer_delta`만 HELD에서 차감한다. actual이 더 크면 usage는 그 차이만큼 증가하며, 실제 가격이 낮다는 이유로 transfer보다 낮춰 여유를 만들 수 없다.

실제 fill price, fee 또는 FX provenance가 unknown이면 fill, watermark와 Position apply를 거부하거나 rollback하지 않는다. 계산 가능한 proportional HELD amount는 provisional filled floor로 이동하되 actual amount를 0으로 간주하지 않고 durable `UNKNOWN_ACTUAL_RISK`를 모든 적용 bucket/owner에 latch해 evidence가 authoritative하게 보완될 때까지 신규 exposure를 차단한다. actual이 나중에 확정되면 같은 fill identity에서 filled amount를 `max(transfer, actual)`로 단조 보완하고 이미 반영된 fill을 다시 적용하지 않는다.

각 transaction은 `filled + remaining HELD - monetary limit`의 positive delta를 horizon, market, strategy, sector와 symbol 각각에 overage로 저장한다. 어느 cap을 초과해도 fill/Position과 risk-reducing 경로는 보존하고 durable `RISK_OVERAGE` latch로 신규 exposure만 차단한다. partial, replacement 및 predecessor late fill도 같은 rule을 사용한다. retry는 fill identity/cumulative watermark로 delta 0이고 crash는 Position, HELD transfer, filled amount, all-bucket overage/latch를 모두 commit하거나 모두 rollback한다. cancel/expiry는 미체결 held 잔량에 대응하는 금액만 release한다. event replay는 owner, monetary reservation, actual provenance, overage와 usage snapshot을 결정적으로 재구성하며 불일치는 entry를 차단한다.

### D6. Owner release는 이전 generation의 protection/sell claim clean을 요구한다

owner release는 Position generation CLOSED/수량 0, pending exposure-raising mutation 및 HELD reservation 부재, reconciliation의 broker 수량 0 확인뿐 아니라 이전 generation에 귀속된 active/pending broker protection order, protection replace/recovery saga, pending sell/reduce-only claim, sell mutation attempt와 unresolved fill observation이 모두 없다는 journal/broker attestation을 요구한다. 하나라도 unknown, stale 또는 남아 있으면 owner를 유지하고 새 generation entry를 차단한다. release가 protection이나 sell claim을 자동 취소·삭제해서는 안 된다.

### D7. 손실 lock과 bucket failure는 entry admission에만 적용한다

daily/horizon loss lock과 bucket snapshot 장애는 EXPOSURE_RAISING decision/leg만 차단한다. RISK_REDUCING, stop, emergency exit, reconciliation과 fill observation은 bucket 계산, owner 획득 또는 FX 수집을 기다리지 않고 기존 경로로 진행한다.

별도 entry-only port를 둬 위험 감소 호출자가 실수로 bucket admission을 거칠 수 없게 한다.

## Risks / Trade-offs

- [다차원 예약 deadlock] → 정규화된 bucket key 정렬 후 하나의 journal transaction에서 획득한다.
- [price/fee/FX/strategy/sector evidence stale] → 진입만 fail closed하고 risk-reducing 경로는 독립시킨다.
- [partial/replacement/late fill 이중 계상 또는 cap 초과] → fill identity watermark와 tx-scoped proportional transfer/actual max를 사용하고 overage·unknown은 fill을 버리지 않고 신규 entry latch로 보존한다.
- [owner가 영구 잔존] → CLOSED generation과 protection/sell claim clean을 증명하는 idempotent release를 제공하며 unknown이면 보수적으로 유지한다.
- [보수 cap으로 기회 감소] → 안전한 의도이며 완화는 별도 사람 승인·audit change로만 수행한다.

## Migration Plan

1. monetary bucket policy, owner, reservation과 event 구조를 additive migration으로 추가한다.
2. reservation/cap 계산기를 pure function으로 구현해 price, nonlinear fee, FX haircut, minor-unit ceil, boundary, overflow, stale/missing과 `q_final <= q_candidate` property를 검증한다.
3. journal 원자 예약·owner race·partial/replacement/predecessor-late fill, actual price/fee/FX unknown, all-bucket overage와 restart/crash replay를 fixture로 검증한다.
4. Guardian에 shadow evaluation을 연결해 기존 승인 수량과 cap 차이를 기록하되 주문 수량은 변경하지 않는다.
5. 후속 runtime change에서 보호 readiness와 사람 승인 조건을 충족한 뒤에만 authoritative entry admission으로 연결한다.

Rollback은 authoritative binding을 OFF로 유지/제거하고 기존 Guardian 경로를 보존한다. 신규 스키마를 구버전이 열면 ErrSchemaTooNew로 fail closed하며 owner/예약 행을 임의 삭제하지 않는다.

## Open Questions

- KRW account에서 US exposure를 평가할 공식 FX source, haircut과 freshness budget은 authoritative binding 전 별도 운영 설정으로 확정해야 한다. 확정 전 US exposure-raising `q_final`은 0으로 fail closed한다.
