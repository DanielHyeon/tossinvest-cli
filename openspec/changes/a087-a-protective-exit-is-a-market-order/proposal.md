# a087 · 보호 청산은 시장가 주문이다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a087`
- **Spec**: `order-execution`
- **위험 등급**: **High-risk** (손절 주문의 유형. §0.3 손절 즉시성 적용.)

> **이 change는 교체본이다.** 초안은 `a087-the-stop-lands-on-the-tick-grid`(호가 그리드
> 정본화)였고 proposal-freeze 리뷰에서 **FREEZE 거부**됐다(`review.md`). 그 리뷰와
> StockOS 대조가 같은 결론에 도달했다 — **호가 계산을 고치는 것이 아니라 호가 계산을
> 타지 않는 것이 답이다.** 그리드 정본화는 a088로, 재가격 에스컬레이션은 a089로 옮긴다.

## Why

**손절 주문이 9분 동안 다섯 번 거부됐다. 지정가였기 때문이고, 지정가일 이유가 없었다.**

원장(`intents`, `pos-a578c51950ad24c05ded2f90`, 2026-08-05):

```text
00:54:00  LIMIT 245750  → 400 invalid-request "주문 가격이 호가 단위에 맞지 않습니다"
00:54:11  LIMIT 245750  → 400
00:55:02  LIMIT 245750  → 400
01:02:44  LIMIT 245750  → 400
01:02:51  LIMIT 245750  → 400
01:03:28  LIMIT 245500  → 통과 (시세가 우연히 그리드 위로 움직였다)
```

포지션은 9분간 손절 없이 있었고, 빠져나온 것은 코드가 아니라 운이다.

### 지정가를 강제한 것은 잘못 인용된 규칙이다

[`exitloop.go:1452`](../../../internal/app/engine/exitloop.go):

```go
// The limit price is the **observed price** … Automated orders are LIMIT only
// (riskcalc's rule, kept on the exit side because a market sell has no price
// the ledger can record an intent against)
```

**riskcalc의 규칙이 아니다.** 그 규칙은 진입 전용이고, 근거는 노출 평가다.

```go
// internal/riskcalc/riskcalc.go:106
ErrMarketEntry = errors.New("riskcalc: automated entries are LIMIT only")
// internal/riskcalc/aggregate.go:27
"automated entries are LIMIT only — a non-limit entry has no defined exposure valuation"
```

승인된 `order-execution` spec도 같다 — "자동 **진입**은 LIMIT 전용이다(SHALL — 시장가
**진입**의 노출 평가가는 정의되지 않는다)", Scenario는 "시장가 **진입** 시도" 하나뿐이다.

**시장가 매수는 노출을 묶을 가격이 없다. 시장가 매도는 노출을 줄인다.** 근거가 전이되지
않는다. 청산 경로가 진입 규칙을 자기에게 잘못 적용했고, 실제 이유로는 **원장 편의**를
적었다. 자본 보호를 원장 표기 편의와 맞바꾼 것이고, 그 대가가 위의 9분이다.

증거가 코드 구조에도 있다 — 위험 권위는 **이미 가격을 요구하지 않는다**:

```go
// exitloop.go:1251 — IssueReduction에 넘기는 intent
risk.Intent{ AccountRef, Market, Symbol, Side: risk.SideSell, Quantity }
//                                     ← Price 필드가 없다
```

### StockOS가 이미 검증했다

`/mnt/D/project/axipient/stockos`, 실운영 중인 KIS 기반 자동매매:

```python
# apps/api/stockos_api/auto_exit_execution.py:2952
def _aggressive_exit_order_type(market: str) -> BrokerOrderType:
    if market.upper() == "KRX":
        return BrokerOrderType.MARKET          # ← 호가 그리드를 타지 않는다
    return BrokerOrderType.MARKETABLE_LIMIT
```

그리고 긴급 청산은 지연 게이트를 **전부 우회**한다. 주석이 이유를 적어 뒀다:

> A stop-loss / emergency-breach exit bypasses the age + price-drop gates —
> deferring it behind a still-young, non-urgent pending order would leave the
> position unprotected (**security review HIGH**).

TossOS의 2c 기본 가설도 이미 **SINGLE+MARKET 손절**이다(measurements M12). 브로커측
보호주문이 MARKET이라면, 그것이 배선되기 전의 앱측 합성 보호도 MARKET이어야 일관된다.

### 브로커는 막지 않는다 (openapi 정본)

`OrderCreateRequest.oneOf[0]` (`OrderCreateQuantityBased`):

| 필드 | 계약 |
| --- | --- |
| `orderType` | `enum ["LIMIT","MARKET"]` — **시장 제한 없음** |
| `side` | `enum ["BUY","SELL"]` |
| `price` | "`LIMIT`일 때만 사용. `MARKET`: **전달 불가**" |
| `quantity` | "기본: **양의 정수만**. 소수점은 미국 시장가 매도에만 — 그 외(매수/지정가/**국내**)는 400" |

US 전용은 **금액 기반 주문**(`oneOf[1]`)과 **소수점 수량** 둘뿐이다. `quantity` 설명이
국내 정수 수량을 명시적으로 상정하고, 엔드포인트 설명도 "지정가·시장가 주문 생성"이다.

**전송 계층은 이미 옳다** — [`orders_write.go`](../../../internal/official/orders_write.go)의
`buildOrderCreate`가 `orderType`을 그대로 올리고 `LIMIT`일 때만 `price`·`timeInForce`를
붙인다(`omitempty`). [`orderintent`](../../../internal/orderintent/intent.go)도 이미
`"market"`을 받고 MARKET이면 `Price = 0`으로 정규화한다.

**차단은 두 줄이다.**

```go
// internal/execgw/failclosed.go:84
if orderType != "limit" { return reject(ReasonUnsupportedOrderType, …) }

// internal/app/engine/exitloop.go:1486
OrderType: "limit",
```

## What Changes

### 1. 보호 청산의 주문 유형을 시장으로 결정한다

`sellIntent`가 유형을 하드코딩하지 않고 **제안의 성격과 시장**으로 정한다.

| 제안 | KR | US |
| --- | --- | --- |
| 보호 (`isProtective` — `BASELINE_BREACH`·`STOP_LOSS_LADDER`) | **MARKET** | **MARKET** |
| 익절 (`LADDER_TAKE_PROFIT`) | LIMIT (현행 유지) | LIMIT (현행 유지) |

익절을 그대로 두는 것이 §0.9다. 익절은 즉시성이 목적이 아니고, 시장가로 바꾸면 체결가
불확실성이 이익 쪽으로 열린다 — 보수 방향이 아니다. 보호만 바꾼다.

US도 MARKET인 근거: 실측 M12가 US 조건주문 SINGLE+MARKET 등록을 확인했고, openapi가
US 온주 MARKET 매도에 제약을 두지 않는다. StockOS가 US에 MARKETABLE_LIMIT을 쓴 것은
KIS 계약 때문이며 토스 계약에는 그 제약이 없다. `[미측정 — US MARKET 매도 실주문 없음]`.

### 2. 게이트를 축소 주문에 한해 연다

`checkOrderShape`의 `orderType != "limit"` 거부에 **축소(매도) 시장가** 분기를 추가한다.
진입(매수) 시장가는 계속 거부한다 — 그것이 riskcalc의 실제 규칙이고 노출 평가가 없다.

```text
side == "sell" && orderType == "market"  → 허용 (price가 없음을 확인)
side == "buy"  && orderType == "market"  → 거부 (현행 유지, US fractional 예외 존치)
```

### 3. 원장이 잃는 것을 명시한다

시장가는 제출 시점에 가격이 없다. 원장의 `intents.price`는 비고, 체결가는 체결 조회로
확정된다. **그것이 정직한 기록이다** — 지금은 "받을 것으로 기대한 가격"을 적고 있고,
그 값은 체결가가 아니며, 위 5건은 아예 주문이 되지도 못했다.

관측·화면이 가격 없는 청산 의도를 표시할 수 있어야 한다.

## Impact

- **Specs**: `order-execution` (ADDED 2 — 보호 청산의 주문 유형, 축소 시장가의 게이트 통과)
- **Code**: `internal/app/engine/exitloop.go`(`sellIntent`), `internal/execgw/failclosed.go`
  (`checkOrderShape`), 관측·원장 표시부
- **Schema**: 없음 (`intents.price`는 이미 빈 값 허용)
- **§0.4**: 요청 수 변화 없음
- **§0.3**: 손절을 **빠르게** 한다. 거부 자체가 사라지고 체결이 보장된다
- **§0.9**: 익절은 무변경. 보호만 즉시성 방향으로

## Non-goals

- **호가 그리드 정본화** → **a088**. 익절 지정가·`flatten` fallback이 여전히 필요로 한다.
  StockOS 형태(십진 문자열·symbol 인자·direction 인자·거부 금지)로 이식
- **재가격 에스컬레이션** → **a089**. 7분 43초 공백(지연 이벤트 0건)은 이 change와 독립
- **진입 주문 유형.** riskcalc의 규칙이 실제로 적용되는 곳이고 바꾸지 않는다
- **조건주문(브로커측 보호).** 2c 범위
- **`flatten`.** 이미 거래소 하한가를 1순위로 쓰며 별도 논거를 갖는다

## 실측 필요 (사용자 승인 항목, §0.7)

1. **KR MARKET 매도 1회 실주문.** 스키마는 지원하지만 실접수는 미측정이다. 최소 수량으로
   장중 1회. 성공 시 `measurements.md`에 기록
2. **세션 경계 거동.** KRX 시간외단일가에는 시장가가 없다. 정규장 밖 MARKET 매도의 응답
   (`422 order-hours-closed`인지 다른 코드인지)이 미측정. 엔진 청산 루프는 정규장 기준이라
   차단 요소는 아니지만, 장 마감 직후 발동 시 거동을 알아야 한다
