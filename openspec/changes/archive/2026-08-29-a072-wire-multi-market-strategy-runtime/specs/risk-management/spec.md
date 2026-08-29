## ADDED Requirements

### Requirement: KR과 US는 하나의 account-base Guardian 권위를 공유한다

Production strategy runtime은 계좌마다 정확히 하나의 account-base-currency Guardian을 사용해야 한다 (SHALL).
그 Guardian은 하나의 account-wide exposure/daily-loss 권위를 소유한다. KR과 US에 quote-currency별
Guardian 또는 독립 account-wide cap을 만들어서는 안 된다 (MUST NOT). `limit_currency`는
account base currency를 뜻하고, 주문 notional·총 노출·일일 손실·equity·risk budget은 그 base
minor unit으로 비교해야 한다 (SHALL). Market cash와 broker order cost는 market quote minor
unit으로 비교해야 하며 base와 quote 숫자를 직접 비교해서는 안 된다 (MUST NOT).

각 exposure-raising 요청은 official source가 mint한 request-scoped frozen FX authority를 사용해야
한다 (SHALL). Authority는 base/quote pair, exact decimal rate, conservative haircut,
source/version/digest, observed-at/fresh-until을 봉인해야 하며 (SHALL), caller가 제공한 raw rate나
digest를 authority로 승격해서는 안 된다 (MUST NOT). KR same-currency path도 봉인된 identity FX를
사용하고 US path는 official quote-to-base FX를 사용해야 한다 (SHALL). 같은 authority가 Guardian
sizing, aggregate와 다섯 bucket reservation, decision limits envelope와 Gateway last-moment
validation에 끝까지 사용되어야 하며 중간 재조회나 다른 환율 사용은 금지한다 (MUST NOT).

Base reservation은 exposure를 작게 계산하지 않도록 ceil하고 admissible quantity는 floor해야 한다
(SHALL). FX 없음, stale, 역방향, market/quote/base/digest 불일치 또는 완화된 haircut은 해당
시장의 신규 exposure만 거부하고 (SHALL), protection, fill, reconciliation과 reduce-only exit를
차단해서는 안 된다 (MUST NOT).

#### Scenario: KR identity와 US official conversion의 동시 admission

- **WHEN** 같은 account-base Guardian에 KR identity-FX 요청과 US official quote-to-base 요청이 동시에 들어온다
- **THEN** 두 요청은 commit 순서와 무관하게 하나의 base exposure 잔여를 공유하며 합산 cap을 초과하지 않고 어느 시장도 peer 안정화를 기다리지 않는다

#### Scenario: US FX 만료

- **WHEN** US q_final precheck에 사용한 official FX가 issue 또는 Gateway 직전 만료된다
- **THEN** US exposure-raising authority는 broker request와 부분 reservation 없이 거부되고 KR eligible work와 양 시장 safety lifecycle은 계속된다

#### Scenario: quote cash와 base limit 분리

- **WHEN** US 주문 비용은 USD이고 Guardian exposure limit은 account base currency다
- **THEN** cash는 USD order cost와, exposure는 동일 frozen FX로 환산한 base amount와 비교되며 USD 숫자를 base limit 숫자와 직접 비교하지 않는다

#### Scenario: cross-currency concurrent cap race

- **WHEN** KR과 US 요청이 동시에 마지막 account-wide capacity를 예약하려 한다
- **THEN** aggregate와 다섯 bucket의 base reservation transaction은 합산 cap 안의 요청만 commit하고 loser는 bounded fresh recollection 또는 atomic refusal이며 중복 권한은 0건이다

#### Scenario: FX outage 중 exit

- **WHEN** entry용 official FX authority를 만들 수 없지만 기존 US 포지션의 reduce-only exit가 필요하다
- **THEN** 신규 US entry만 닫히고 exit, protection, reconciliation과 fill 처리는 계속된다
