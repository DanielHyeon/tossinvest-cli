# Change: add multi-horizon risk buckets

## Why

기존 Guardian은 주문·총 노출·일손실 한도를 강제하지만, 단기/중기 전략과 KR/US 시장이 동시에 scale-in할 때 horizon·market·sector·symbol별 집중도를 합산하는 계약이 없다. 또한 같은 symbol을 여러 lane가 동시에 소유하면 각각의 주문이 개별 한도를 통과해도 포지션 위험과 exit 책임이 분리될 수 있다. 다중 lane 연결 전에 보수적인 수량 cap과 단일 소유권을 원자적으로 강제해야 한다.

## What Changes

- short/medium horizon, market, strategy, sector 및 symbol 차원의 금액 위험 bucket과 명시적 unknown 분류를 추가한다.
- lane가 제안한 canonical `q_candidate`를 상한으로 두고 worst executable price, worst-case fees와 fresh FX haircut을 적용한 account-base-currency 예약 금액으로 각 bucket cap을 계산한다. canonical `q_final`은 `q_candidate`를 절대 초과하지 않으며 누락·stale·통화 환산 불가·산술 오류는 수량 0 또는 typed refusal로 fail closed한다.
- 최초 campaign 예약 시 symbol/market의 owning lane을 원자적으로 획득하고, 다른 lane의 경쟁 결정이나 scale-in이 두 번째 campaign 또는 별도 위험 예약을 만들지 못하게 한다.
- leg 추가·부분체결·재시작 시에도 동일 campaign의 filled monetary exposure와 HELD monetary reservation을 합산한다. owner는 이전 generation의 protection 및 sell claim까지 clean하다는 증거가 있어야 release된다.
- 각 fill transaction은 모든 적용 bucket에서 원 예약의 proportional HELD transfer와 실제 체결가격·fee·FX 기반 monetary exposure를 비교해 filled amount를 둘 중 큰 값으로 기록한다. cap overage 또는 실제 price/fee/FX unknown이어도 fill/Position은 보존하고 durable `RISK_OVERAGE` 또는 `UNKNOWN_ACTUAL_RISK` latch로 신규 exposure만 차단한다.
- Guardian precheck는 기록 권한을 만들지 않으며, `q_final`이 확정된 뒤에만 그 수량을 봉인한 GuardianDecision, lane owner와 모든 금액 reservation을 하나의 journal transaction으로 발급한다.
- bucket 초과와 진입 손실 latch는 exposure-raising 경로만 차단한다. stop, emergency exit, reduce-only 주문, reconciliation 및 fill detection은 평가·대기시키지 않는다.

## Capabilities

### New Capabilities

- `multi-horizon-risk-buckets`: horizon/market/strategy/sector/symbol 금액 노출 cap, conservative price/fee/FX 예약, scale-in 합산, 단일 owning-lane 예약 및 fail-closed `q_candidate → q_final` 계약

### Modified Capabilities

- `risk-management`: 기존 Guardian 고정 판정 체인과 예약 권위를 확장해 `q_final <= q_candidate`, final-quantity GuardianDecision, 금액 bucket 예약·해제 및 entry-only loss lock을 강제

## Impact

- Guardian 입력·reason code·journal 예약 스키마와 exposure snapshot/reconstruction, fill별 proportional transfer/actual exposure/overage evidence가 확장된다.
- a065의 전략 중립 campaign identity를 소비하지만 전략별 leg 비율은 알지 않는다. KR과 US lane는 같은 규칙 아래 독립 market bucket으로 동시에 개발·검증할 수 있다.
- 위험 감소와 관측 루프의 우선순위는 보존된다. 운영 한도 완화, lane ON, live 주문은 사람 승인 없이는 수행하지 않는다.
