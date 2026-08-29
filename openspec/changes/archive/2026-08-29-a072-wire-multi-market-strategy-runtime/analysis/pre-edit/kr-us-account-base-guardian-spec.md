# KR/US paired account-base Guardian pre-edit specification

## Decision

KR과 US는 정확히 하나의 account-base-currency `RiskGuardian`을 공유한다. KR의 quote 통화가
account base와 같더라도 명시적으로 봉인된 identity FX를 사용하고, US는 같은 요청 안에서
동결한 official quote-to-base FX와 보수 haircut을 사용한다. 두 시장을 각각 quote-currency
Guardian으로 나누는 방식은 account-wide exposure와 daily-loss cap을 두 번 허용할 수 있으므로
금지한다.

이 결정은 KR 운영 안정화를 US 구현의 선행조건으로 만들지 않는다. 아래 계약, RED/GREEN,
production adapter와 Gateway 검증은 항상 KR identity case와 US converted case를 같은 table과
같은 implementation wave에서 통과해야 한다.

## Frozen authority contract

하나의 요청에 사용되는 FX authority는 base currency, market quote currency, quote-to-base exact
decimal rate, conservative haircut, source/version/digest, observed-at와 fresh-until을 봉인한다.
production mint는 `officialfx.EvidenceAt`에서만 가능하고 caller의 raw rate, digest 또는 freshness
DTO를 authority로 받아서는 안 된다.

동일 authority는 다음 경계를 끝까지 지나야 한다.

```text
KR/US strategy result
  -> Guardian sizing and account-base limits
  -> aggregate plus five-bucket HELD reservations
  -> versioned decision limits envelope
  -> CLAIMED-to-SUBMITTING last-moment fence
  -> official ExecutionGateway validation
```

중간 단계가 FX를 다시 읽거나 다른 rate/haircut/digest를 사용하면 전체 admission은 fail-closed다.

## Unit and rounding invariants

- policy order-notional, open-exposure, daily-loss, equity와 risk budget은 account-base minor unit이다.
- available cash와 broker order cost는 market quote minor unit이다.
- KR도 identity authority를 생략하지 않는다. US는 USD-to-base official authority를 요구한다.
- base reservation은 불리한 방향으로 ceil하고 admissible quantity는 floor한다.
- FX 없음, 만료, 역방향, market/quote/base/digest 불일치 또는 haircut 완화는 exposure-raising
  entry만 거부한다. protection, reconciliation, fill, reduce-only exit는 계속한다.
- `limits_json`은 versioned envelope로 base limits와 exact FX evidence를 함께 저장한다. old/unknown
  envelope를 현재 base authority로 추정하지 않는다.

## Paired branch/test map

| Branch | KR | US | Shared assertion |
| --- | --- | --- | --- |
| Valid authority | KRW identity | USD-to-KRW official | same account Guardian and limits digest |
| Combined cap | KR hold first/second | US hold first/second | commit order와 무관하게 합산 base cap 초과 0 |
| Residual sizing | identity-converted residual | official-converted residual | second market downsizes or refuses |
| Cash | KRW quote cash | USD quote cash | cash is never compared in base units |
| FX refusal | stale/tampered identity | missing/stale/reversed official | zero journal rows and zero broker calls |
| TOCTOU | expires before issue | changes before issue | fresh recollection or atomic refusal, no authority reuse |
| Replay | exact identity replay | exact official replay | exact only idempotent; changed FX divergent |
| Gateway | exact envelope | exact envelope | swapped pair/digest/freshness produces zero broker calls |
| Isolation | KR authority failure | US authority failure | failed market does not cancel eligible peer or safety loops |
| Exit | no entry FX required | no entry FX required | risk-reducing path remains live |

## Existing high-risk functions requiring FLM before edits

- `risk.checkOrderSize`
- `risk.entryNotional`
- `risk.entryNotionalWithCosts`
- `risk.checkOpenExposure`
- `risk.checkDailyLoss`
- `risk.EntryExposureValue`
- `risk.StrategyEntryQuantity`
- `execgw.(*RiskGuardian).PrecheckQFinalEntry`
- the Gateway reservation/decision envelope validator used immediately before transport

No LIVE order, activation, approval or operating-toggle mutation is authorized by this contract.
