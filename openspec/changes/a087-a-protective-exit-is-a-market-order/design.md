# a087 설계 — 보호 청산에서 가격을 없앤다

## 문제의 형태

초안은 "제출가가 호가 그리드 위에 있게 하자"였다. proposal-freeze 리뷰가 그 설계를 무너뜨렸고
(`review.md`), 무너진 자리에서 더 단순한 질문이 남았다.

**보호 청산이 왜 가격을 갖는가?**

답은 없었다. 코드가 인용한 근거(`riskcalc`의 LIMIT 전용)는 진입 전용 규칙이었고, 코드가
실제로 적은 이유는 "원장에 적을 가격이 필요해서"였다. 체결 보장을 원장 표기와 맞바꾼 것이다.

가격을 없애면 초안이 풀려던 문제가 **전부 존재하지 않게 된다**:

| 초안이 풀려던 것 | 시장가에서 |
| --- | --- |
| 호가 그리드 이탈 (실측 5회 거부) | 가격이 없다 |
| 시세가 중간가일 가능성 (D1) | 가격을 안 쓴다 |
| ETF·우선주 override 갭 | 표를 안 탄다 |
| KOSDAQ 표 분기 | 〃 |
| `big.Rat` 정밀도 (US 48% 이동) | 〃 |
| 세 사본 정본화·AST 가드 | 이 경로에서는 불필요 |
| 가격 밴드 clamp | 브로커가 처리 |

## D1 — 유형은 제안의 성격이 정한다

`sellIntent`는 지금 유형을 상수로 쓴다. 제안을 인자로 받아 판정한다.

```go
orderType := "limit"
if isProtective(proposal) {   // BASELINE_BREACH · STOP_LOSS_LADDER
    orderType = "market"
}
```

**`isProtective`를 재사용하는 것이 핵심**이다. 새 술어를 만들면 "보호란 무엇인가"의 정의가
둘이 되고, 그 둘이 갈라지는 순간 한쪽만 시장가가 된다. [`exitloop.go:1217`](../../../internal/app/engine/exitloop.go)의
`isProtective`는 이미 §0.3 근거로 존재하고 주석이 그 이유를 적어 뒀다 — "Nothing may
withhold one of these: §0.3 forbids weakening or delaying the immediacy of a stop".
같은 §0.3이 이 change의 근거이므로 같은 술어여야 한다.

`isFullExit`는 쓰지 않는다. 그것은 익절을 **포함**하고(`ActionLadderTakeProfit`), 익절은
지정가로 남는다.

### 익절을 안 바꾸는 이유

§0.9는 보수 방향만 허용한다. 익절의 목적은 즉시성이 아니라 **가격 확보**다. 시장가로
바꾸면 체결가 불확실성이 커지고 그 불확실성은 이익을 깎는 쪽으로도 열린다 — 보수 방향이
아니다. 그리고 익절이 안 나가도 포지션은 여전히 보호받는다(보호 제안이 별도로 존재한다).
보호가 안 나가면 아무것도 남지 않는다. **비대칭이 설계다.**

## D2 — "가격 없음" 거부가 보호를 막지 않게 한다 (2상 — 사용자 결정 2026-09-28)

> a089 처분에서 사용자가 확정한 재작성 축: **D2 를 분리해 선행**한다. 시장가 전환(D1·D3)은
> 사람 실측(§0.7) 뒤에만 착지할 수 있으므로, 그 실측을 기다리는 동안 "가격 없음" 거부가
> 보호를 막는 상태를 두지 않는다.

현재 `sellIntent`(exitloop.go:1571-1578)의 가격 사다리:

```go
price := strings.TrimSpace(observed)
if price == "" { price = strings.TrimSpace(m.state.Baseline) }
if price == "" {
    return orderintent.PlaceIntent{}, fmt.Errorf(
        "position %s has no price to submit a liquidation at", m.position.ID)
}
```

관측가도 기준선도 없으면 **청산을 거부한다.** 시세를 못 읽은 순간 손절이 막힌다는 뜻이고,
그것은 §0.3이 금지하는 형태다.

### D2a — 선행 (실측 불요): 사다리에 셋째 단 — KR 당일 하한가

보호 제안(`isProtective`)에 한해 사다리를 늘린다: **관측가 → 기준선 → (KR) 당일 하한가 →
거부**. 하한가 LIMIT 매도는 "그 세션에서 가능한 가장 공격적인 유효 지정가"이고, 거래소
공표값이라 호가 그리드 위에 있다.

코드 영수증:

- 값의 출처는 공식 read `PriceLimits`(`internal/official/price_limits_reads.go` —
  `lowerLimitPrice`, nullable). 엔진 reader 인터페이스에 **이미 선언돼 있으나 비시험
  호출자 0** 이다(`internal/app/engine/reads.go:72`) — 배선만 없다.
- **KRX 전용**이다: `internal/client/marketdata.go:97` "미국장은 일일 가격제한 제도가
  없음". 그러므로 이 단은 KR 에만 선다.
- §0.4: 이 읽기는 관측 루프가 아니라 **거부 직전 폴백에서만** 1회 — 사건 한정이다.

남는 거부(정직한 열거 — fail-closed 는 거부하는 정상 입력을 말한다):

- US 포지션에서 관측가·기준선이 모두 빈 경우(하한가 제도 없음), 그리고 KR 하한가 읽기
  자체가 실패·null 인 경우. 이때는 거부하되 **critical 로 알린다**(무음 거부 금지 —
  a090 의 미관측 계수와 합집합, 교차 인용). US 무가격은 D2b 가 착지하면 자연 소멸한다
  (시장가는 가격이 필요 없다).
- 두 값이 모두 비는 모집단이 언제 생기는지(기준선 없는 포지션의 유형)는 phase 1 RED 가
  열거한다.

#### 반증 (2026-09-30, P1.2 ① 모집단 열거) — D2a 는 세우지 않는다

열거 결과 모집단은 **0** 이다. 위 사다리의 폴백(B1 `:1573`)과 거부(B2 `:1576`)는 생산 경로에서 도달할 수 없다.
Manager 가 독립 대조해 확인했다(2026-09-30). 영수증(`issues.md` I-P1, `analysis/`):

- **호출 사슬 (CodeGraph 1.6.0)** — `sellIntent` 의 생산 호출 1곳 = `submit` `:1382`, `submit` ← `record` `:1301`,
  `record` ← `judgeRatchet`·`judgeLadder`. `observed` = `snapshot.ObservedPrice`(`record` `:1190`), 도중 재대입 없음.
- **가드 좌표 (AST, `analysis/p1-population/`)** — 스냅숏은 평가 성공 **뒤에만** 생성된다(`snapshot.go:189`→`:191`,
  `:120`→`:122` 오류 반환). 평가기는 성공 반환 전에 `positive("observed price", …)` 를 무조건 지난다
  (`ratchet.go:347` · `ladder.go:333`, 그 앞 반환은 전부 오류). `positive` 는 파싱 실패·`Sign() <= 0` 거부.
- **커버리지 0** — 엔진 스위트(`go test ./internal/app/engine -coverprofile`) B1 본문 `1573.17-1575.3` = 0,
  B2 본문 `1576.17-1579.3` = 0 (`analysis/p1-population/sellintent-coverage.txt`).

그러므로 **D2a 는 세우지 않는다.** 도달 불가 분기 뒤의 High-risk 코드는 생산 경로로 변이가 닿지 않고, 하한가 자동
제출은 미검증 가상 호출자에게 주문을 내주는 방향이라 보수 방향도 아니다 — 오늘의 B2 거부가 이미 그 가상 경로의
fail-closed 방어다. **B1·B2 는 fail-closed 방어 반환으로 남고, 닫힌 이유는 핀 시험이 고정한다**:
`internal/exitpolicy/a087_observed_price_pin_test.go`(빈·공백·"0"·파싱 불가 관측가 → 두 평가기 모두 거부, 변이 4/4 CAUGHT).
가드가 빠지는 편집은 그 시험을 깨뜨리고, 그때 B1·B2 가 다시 열린 문이 된다는 사실이 드러난다.

**"무가격 → 손절 없음"의 실제 자리는 `sellIntent` 가 아니라 판정 이전이다** — `observe` 가 `Last <= 0` 을 거르고
(`exitloop.go:765`) `ObserveOnce` 가 응답 없는·만료된 종목을 무음 `continue` 한다. 가격이 없으면 이탈을 판정할 수 없어
보호 제안 자체가 생기지 않는다. 그 자리는 **a090**(`a090-an-unobserved-position-is-counted`, freeze 완료) 소관이다.

### D2b — 시장가 전환 (실측 §0.7 뒤): 거부의 구조적 소멸

시장가 경로는 가격을 필요로 하지 않으므로 **보호 제안에서는 이 거부가 도달 불가가 된다.**
초안 리뷰의 H1이 "새 거부 경로를 만들지 말라"고 한 것의 반대편 — 있던 거부 경로 하나가
사라진다.

가격 읽기 자체를 보호 분기에서 **건너뛴다**. 읽어서 버리는 것이 아니라 읽지 않는다.
읽으면 실패할 수 있고, 실패가 거부로 이어지는 경로가 다시 생긴다.

~~D2b 가 착지하면 D2a 의 하한가 단은 보호 경로에서 도달 불가가 된다 — phase 2 GREEN 이
그 단을 **제거**하고 도달 불가 증명을 함께 남긴다.~~ (2026-09-30: D2a 불구현 — 위 「반증」. 제거할 단이 없다.
D2b 가 보호 분기에서 가격 읽기를 건너뛰면 B1·B2 는 보호 경로에서 구조적으로 사라지고, 익절 경로에는 오늘처럼
평가기 가드 뒤의 fail-closed 방어로 남는다.)
9분 사건(그리드 불일치 400 — 관측가는 **있었다**)을 고치는 것은 D2a 가 아니라 D2b 다:
D2a 는 가격 부재 거부만 다룬다.

## D3 — 게이트는 축소에 한해서만 연다

[`failclosed.go:84`](../../../internal/execgw/failclosed.go):

```go
if orderType != "limit" {
    return reject(ReasonUnsupportedOrderType,
        "only limit orders (and US fractional market orders) are supported, got %q", …)
}
```

이 한 줄이 진입과 청산을 구분하지 않는다. 구분을 넣는다.

```text
fractional          → 기존 분기 그대로 (US market only, 이미 line 70-82)
sell + market       → 통과. 단 Price != 0 이면 거부
buy  + market       → 거부 (현행 유지)
그 외 non-limit     → 거부 (현행 유지)
```

**`Price != 0` 검사를 넣는 이유**: openapi가 `MARKET`에 `price` 전달을 금지하고 전달 시
`400 invalid-request`를 준다. `orderintent`가 이미 `intent.Price = 0`으로 정규화하지만
(`intent.go:140`), 게이트는 정규화를 신뢰하지 않는 자리다 — `checkOrderShape`의 주석이
스스로 "a strict subset filter"라고 말한다. 두 번 확인하는 것이 이 함수의 일이다.

### 왜 진입은 계속 막는가

`riskcalc`의 근거가 진입에서는 여전히 옳다 — 시장가 매수는 노출 평가가가 없고, Guardian이
예약할 금액을 계산할 수 없다. 그 규칙을 이 change가 건드리지 않는다는 것을 **별도 테스트로
고정**한다. 안 그러면 다음 사람이 "a087이 시장가를 열었다"로 읽는다.

## D4 — 원장이 무엇을 기록하는가

시장가는 제출 시점에 가격이 없다. 원장의 `intents.price`는 빈 문자열이 된다.

**이것이 정보 손실처럼 보이지만 아니다.** 지금 그 칸에 들어가는 값은 관측가이고, 그것은
체결가가 아니며, 실측 5건에서는 **주문이 되지도 못한 값**이었다. 없는 가격을 적는 것보다
비우는 것이 정직하다.

체결가는 `filldetect`가 체결 조회로 확정한다 — 이미 그렇게 동작한다면 변경 없음을
테스트로 고정하고, 아니면 갭으로 기록한다(tasks 3.5).

초안 리뷰 M3이 "정렬된 제출가를 원장에 남겨라"고 한 것과 방향이 반대로 보이지만 같은
요구다: **원장은 실제로 보낸 것을 적어야 한다.** 초안에서는 정렬된 가격이 실제로 보낸
것이었고, 여기서는 가격이 없는 것이 실제로 보낸 것이다.

화면·알림은 빈 값을 결측으로 그리면 안 된다. `—`가 아니라 "시장가"다.

## D5 — StockOS 대조 (승계와 미승계)

| StockOS | 위치 | a087 판정 |
| --- | --- | --- |
| KRX 긴급 청산 = `BrokerOrderType.MARKET` | `auto_exit_execution.py:2952` | **승계** — D1 |
| 긴급 청산은 지연 게이트 전부 우회 | `auto_exit_execution.py:3130` | **a089로** |
| N회 후 MARKETABLE_LIMIT → true MARKET 승격 | `auto_exit_execution.py:3160` | **a089로** |
| US = MARKETABLE_LIMIT | 같은 파일 | **미승계** — KIS 계약 제약이고 토스에는 없다 |
| 호가 정렬 `Decimal(str(price))` | `auto_exit_execution.py:3001` | **a088로** |
| 정렬 실패 시 fallback, 거부 없음 | `auto_exit_execution.py:3005-3017` | **a088로** |
| `TickSizeProvider` + symbol override | `tick_size_provider.py` | **a088로** |

StockOS가 US에 marketable limit을 쓰는 것은 `protective_order_capability`가 기록한 KIS의
한계 때문이다 — `kis_us_stock_rest_stop_oco_not_supported`. 토스 openapi에는 그 제약이
없고 US 온주 MARKET 매도를 금지하는 문장이 없다. 그래서 양 시장 모두 MARKET으로 간다.
`[미측정 — US MARKET 매도 실주문 없음]`.

## 건드리지 않는 것

- **진입 주문 유형.** riskcalc 규칙이 실제로 적용되는 곳
- **`flatten`.** 이미 거래소 하한가를 1순위로 쓴다(`liquidate.go:372-378`). 별도 논거
- **`verifylive`.** "체결되면 안 되는 지정가"가 그 도구의 안전 성질이다
- **조건주문.** 2c 범위
- **호가 그리드 정본화.** a088

## 검증

- `sellIntent` 분기: 보호 2종 → market·익절 → limit·시장가에 가격 없음·**보호가 가격
  없음으로 거부되지 않음**
- `checkOrderShape` 표: `sell+market` 통과 / `buy+market` 거부 / 가격 실린 시장가 거부 /
  fractional·limit 기존 분기 무변화
- 진입 시장가 거부가 **살아 있음**을 별도 테스트로 고정
- 원장: `price` 빈 값, `order_type` market, 관측가·기준선이 제출가로 새지 않음
- `flatten`·`verifylive`·조건주문 무변화
- 전체 `go test ./... -race` 회귀 0, upstream 650 green 유지
- **실계좌 1회**(사용자 승인): KR MARKET 매도 접수 + 세션 경계 응답
