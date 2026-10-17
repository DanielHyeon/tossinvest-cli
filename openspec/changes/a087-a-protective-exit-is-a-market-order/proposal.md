# a087 · 보호 청산은 시장가 주문이다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a087`
- **Spec**: `order-execution`
- **위험 등급**: **High-risk** (손절 주문의 유형. §0.3 손절 즉시성 적용.)

> **이 change는 교체본이다.** 초안은 `a087-the-stop-lands-on-the-tick-grid`(호가 그리드
> 정본화)였고 proposal-freeze 리뷰에서 **FREEZE 거부**됐다(`review.md`). 교체본의 축은
> **호가 계산을 고치는 것이 아니라 호가 계산을 타지 않는 것**이다. (3판까지 「그 리뷰와 StockOS 대조가 같은 결론에
> 도달했다」고 적었으나 StockOS 는 1차 손절을 LIMIT 으로 낸다 — 아래 「StockOS 대조」 정정.) 그리드 정본화와
> 재가격 에스컬레이션은 **무소유**다(「a088」은 존재한 적 없고 a089 는 다른 내용으로 아카이브됨 — Non-goals).
>
> **2판 (사용자 결정 2026-09-28, a089 처분)**: **D2 를 분리해 선행한다.** 시장가 전환은
> 사람 실측(§0.7 — KR MARKET 매도 1회·세션 경계) 뒤에만 착지할 수 있으므로,
> 이 change 는 두 단계로 간다 — **Phase 1**(실측 불요): 보호 청산이 "읽을 가격이 없다"고
> 거절되지 않게 가격 사다리에 KR 당일 하한가 단을 더한다(design D2a). **Phase 2**(실측
> 게이트): 아래 What Changes 1~3(시장가 전환) — 착수 조건은 tasks 5.1·5.2 실측 기록이다.
> 재가격·에스컬레이션은 a089 불구현 아카이브(64a1b2b3)로 현재 무소유다.
>
> **3판 (Manager 판정 2026-09-30)**: Phase 1 은 **불구현 종결(전제 반증)** — 아래 §0 처분 절. 이 change 의
> 남은 범위는 Phase 2 뿐이며 §0.7 사람 실측 게이트 대기다.
>
> **4판 (2026-10-10, 2차 proposal-freeze 재리뷰 처분 + 사용자 결정)**: `review.md` 「2차 proposal-freeze 재리뷰」의
> P0 1·P1 4·P2 5·P3 3 을 문서로 처분했다(처분 대조표는 `review.md` 끝). 사용자 결정 두 건이 근거다 —
> ① **차단은 세 곳**이고 셋째 관문(`internal/trading` `placeIntentSupported`)을 **sell+market·가격 없음**에만 연다
> (a087 범위 확장 — 개방 범위는 아래 5판이 정정), ② 5.1 실측은 **별도 실측 도구**
> `tools/a087-market-sell-probe/` 로 production 변경 **전에** 한다. 반증된 근거(「StockOS 가 이미 검증」·「체결 보장」·
> 「정규장 기준」·「9분 무보호」)는 아래 본문에서 반증 사실과 함께 정정했다.
>
> **5판 (2026-10-10, 3차 proposal-freeze 재리뷰 처분 + 사용자 결정 2차)**: 3차 재리뷰 A-P1-1 이 ③ 무조건 개방을
> 불변식 3 과 양립 불가로 판정했고(사람 CLI 뿐 아니라 ops `place_order` → **MCP 에이전트 표면**까지 열린다), 사용자가
> **엔진 인스턴스 한정 개방**을 승인했다(`review.md` 「사용자 결정 (2026-10-10, 2차)」). ③의 sell+market 허용은 엔진이
> 만드는 `trading.Service` 인스턴스에만 생성자 옵션으로 서고, CLI·ops·MCP 가 쓰는 기본 생성 경로는 upstream 과
> 바이트 동일하다. 4판까지의 「사람 CLI 도 시장가 매도가 가능해진다」는 **철회된 1차 결정**이며 본문 전부 정정했다.
> 처분 대조표(3차분)는 `review.md` 끝.

## Why

**손절 결정 6회 중 5회가 제출되지 못했다(400 거부). 지정가였기 때문이고, 지정가일 이유가 없었다.**

원장(`intents`, `pos-a578c51950ad24c05ded2f90`, 2026-08-05):

```text
00:54:00  LIMIT 245750  → 400 invalid-request "주문 가격이 호가 단위에 맞지 않습니다"
00:54:11  LIMIT 245750  → 400
00:55:02  LIMIT 245750  → 400
01:02:44  LIMIT 245750  → 400
01:02:51  LIMIT 245750  → 400
01:03:28  LIMIT 245500  → 통과 (시세가 우연히 그리드 위로 움직였다)
```

6회 결정 중 5회가 400 으로 미제출됐고, 6번째가 통과한 것은 코드가 아니라 시세가 그리드 위로 움직인 우연이다.

> **정정 (2026-10-10, 재리뷰 P2-1)**: 3판까지 이 자리는 「포지션은 9분간 손절 없이 있었다」였다. a089 감사
> (아카이브 `2026-09-28-a089-an-unserved-stop-is-counted/proposal.md` — `exit_events` 15행·`engine.log` 전수)가
> 「말할 수 없는 것: 그 9분 28초가 연속 무보호였다는 것」이라 적었고, 7분 42초 공백은 **주문 가능한 제안이 없었던
> 구간**이다(재리뷰 인용 `snapshot.go` `|| s.Orderable`). 척도를 시간에서 **횟수**로 바꾼다. 결론(5회 400 거부)은 그대로다.

### 지정가를 강제한 것은 잘못 인용된 규칙이다

[`exitloop.go`](../../../internal/app/engine/exitloop.go) `ExitObserver.sellIntent` 바로 위 주석
(좌표는 심볼 상대 — 재리뷰 P2-5):

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
적었다. 자본 보호를 원장 표기 편의와 맞바꾼 것이고, 그 대가가 위의 「6회 결정 중 5회 미제출」이다.
(3판까지 「그 대가가 위의 9분이다」 — 3차 재리뷰 A-P3-1 이 P2-1 정정의 잔재로 지적해 고쳤다.)

증거가 코드 구조에도 있다 — 위험 권위는 **이미 가격을 요구하지 않는다**:

```go
// exitloop.go ExitObserver.submit — IssueReduction에 넘기는 risk.Intent 리터럴
risk.Intent{ AccountRef, Market, Symbol, Side: risk.SideSell, Quantity }
//                                     ← Price 필드가 없다
```

### StockOS 대조 — 선례가 아니다 (정정 2026-10-10, 재리뷰 P1-2 · 2026-08-06 2차 리뷰 C2)

> **3판까지의 제목은 「StockOS가 이미 검증했다」였고 거짓이다(생략에 의한 오인용).** 반증 사실과 함께 남긴다.

`/mnt/D/project/axipient/stockos`, 실운영 중인 KIS 기반 자동매매에서 `_aggressive_exit_order_type`
(KRX → `MARKET`)는 **에스컬레이션 경로에서만** 불린다(2차 리뷰 C2 인용, `auto_exit_execution.py` `_apply_exit_order_style`):

```python
if _is_emergency_breach(plan) or EOD_FLATTEN or early_surge_fast_partial_profit:
    order_type = _aggressive_exit_order_type(...)      # KRX → MARKET
return replace(plan, order_type=BrokerOrderType.LIMIT, ...)   # ← 그 외 전부
# _is_emergency_breach: quote ≤ trigger × (1 − 0.5%)
```

StockOS 의 평범한 `STOP_LOSS`·`STOP_LOSS_LADDER` **1차 제출은 LIMIT** 이고, MARKET 은 0.5% 초과 이탈·EOD
에스컬레이션이다. a087 은 임계 없이 1차부터 MARKET 이므로 **인용한 실운영 시스템보다 엄격히 더 공격적**이다.
그러므로 StockOS 는 a087 의 근거가 아니다. a087 의 근거는 (1) 위 「잘못 인용된 규칙」 — 청산측 LIMIT 강제에
정당한 이유가 없다는 것, (2) 사용자 결정(2026-10-10 — 범위 확장으로 MARKET 방향 유지), (3) §5 실측이다.
거래소 하한가 지정가 원안과의 비교는 design D6(미채택 기록 — 원안 자체 평가는 미수행)에 있다.

TossOS 의 2c 기본 가설은 **SINGLE+MARKET 손절**이다(measurements M12 — 조건주문 등록 실측). 이것은 일관성
논거일 뿐 `/api/v1/orders` MARKET 접수의 증거가 아니다(조건주문은 다른 endpoint — 재리뷰 P1-1).

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

**차단은 세 곳이다.** (정정 2026-10-10, 재리뷰 P0-1 — 3판까지 「차단은 두 줄이다」였고 거짓이다.
셋째 관문은 2026-08-06 2차 리뷰 C1 이 critical 로 적었으나 3판 본문에 반영되지 않았다.)

```go
// ① internal/app/engine/exitloop.go — ExitObserver.sellIntent 의 PlaceIntent 리터럴
OrderType: "limit",

// ② internal/execgw/failclosed.go — checkOrderShape (호출자 CheckPlace 1)
if orderType != "limit" { return reject(ReasonUnsupportedOrderType, …) }

// ③ internal/trading/service.go — placeIntentSupported (비분수 주문)
if intent.OrderType != "limit" { return false }   // → Place 가 ErrPlaceUnsupported
```

②와 ③은 **독립된 관문**이다(CodeGraph 1.6.0 — `placeIntentSupported` 비시험 호출자 = `PreviewPlace`·`Place`,
`checkOrderShape` 호출자 = `CheckPlace`). 엔진의 모든 주문은 `execgw` `Gateway.place` 에서
`g.trading.Place(...)`(구체 타입 `*trading.Service`)를 지난다. ①②만 고치면 보호 청산은 **100% 로컬 거부**된다
(재리뷰 인용 경로, 검증 필요 — AST 미작성: `ErrPlaceUnsupported` → `ReasonUnsupportedOrderType` → `submit` 의
`default` 갈래 → `alertProposalRefused` dedup 1건 뒤 침묵).

③은 upstream 상속 코드이고(upstream/main 본문 동일 — 3차 재리뷰), 이 술어를 거치는 `trading.Service` 인스턴스는
셋이다 — 엔진(`internal/app/engine/engine.go` `OrderPath.Trading`), CLI 앱(`internal/app/app.go` → `tossctl order place`·
`tossctl ops`), **MCP 서버**(`cmd/tossctl/mcp.go`). ops `place_order`(`internal/ops/write_operations.go`)는 MCP 카탈로그
(`internal/mcp/catalog.go`)로 노출되는 **에이전트 표면**이다. `checkOrderShape` 주석은 「internal/trading is not ours to
change」라 적는다.

**사용자 결정 (2026-10-10, 2차 — 1차 결정을 대체)**: ③을 ②와 같은 모양 — **sell+market·가격 없음** — 으로 열되
**엔진이 만드는 인스턴스에만** 생성자 옵션으로 연다. CLI·ops·MCP 의 기본 생성 경로는 upstream 과 바이트 동일
(`ErrPlaceUnsupported` 유지)이고, 그 동일성을 시험으로 고정한다. 엔진이 꺼져 있으면 엔진 인스턴스가 없으므로 흔적 0 —
불변식 3 보존(`issues.md` I-R1 종결). 1차 결정(2026-10-10)이 수용했던 「사람 CLI 에서도 시장가 매도 가능」은 이 처분으로
**철회**됐다 — 사람의 측정 수단은 `tools/a087-market-sell-probe/` 다. `buy+market` 거부는 엔진 인스턴스에서도 시험으로 고정한다.

## What Changes

### 0. Phase 1 (선행, 실측 불요) — 보호 청산의 가격 사다리에 KR 하한가 단

> **처분 (2026-09-30, Manager 판정): 불구현 종결 — 전제 반증.** P1.2 ① 모집단 열거에서 이 단이 앞에 서려던
> 거부(`sellIntent` B2)와 그 앞 폴백(B1)이 생산 경로에서 도달 불가로 측정됐다(관측가는 평가기의 `positive` 가드를
> 지난 값뿐, 엔진 스위트 커버리지 0). 하한가 단은 세우지 않고, B1·B2 가 닫힌 이유를 평가기 핀 시험
> (`internal/exitpolicy/a087_observed_price_pin_test.go`)이 고정한다. 영수증: design D2a 「반증」, `issues.md` I-P1.
> 무가격 → 손절 없음의 실제 자리(판정 이전 무음 skip)는 a090 소관이다. **Phase 2(아래 1~3, §0.7 실측 게이트)는 유지.**
> 아래 원문은 기록으로 남긴다(줄 좌표는 base `102d4e99` 기준 — 현재 HEAD 와 다르다).

`sellIntent`의 사다리(관측가 → 기준선 → 거부, exitloop.go:1571-1578)를 보호 제안에 한해
**관측가 → 기준선 → (KR) 당일 하한가 → 거부**로 늘린다. 하한가는 거래소 공표값이라 호가
그리드 위에 있고, 그 세션에서 가능한 가장 공격적인 유효 지정가다. 값의 출처는 이미 엔진
reader 인터페이스에 선언된 `PriceLimits`(reads.go:72, 비시험 호출자 0 — 배선만 없다)이고
**KRX 전용**이다(marketdata.go:97). 남는 거부(US 무가격·KR 하한가 읽기 실패)는 critical 로
알린다(무음 금지, a090 교차). 상세는 design D2a. Phase 2 가 착지하면 이 단은 보호 경로에서
도달 불가가 되어 제거된다(도달 불가 증명과 함께).

> 9분 사건(그리드 불일치 400)은 Phase 1 이 고치지 않는다 — 관측가는 **있었다**. 그것을
> 고치는 것은 Phase 2 의 시장가다. Phase 1 은 가격 부재 거부만 없앤다.

### 1. Phase 2 — 보호 청산의 주문 유형을 시장으로 결정한다 (착수 조건: §0.7 실측)

`sellIntent`가 유형을 하드코딩하지 않고 **제안의 성격과 시장**으로 정한다.

| 제안 | KR | US |
| --- | --- | --- |
| 보호 (`isProtective` — `BASELINE_BREACH`·`STOP_LOSS_LADDER`) | **MARKET** | **MARKET** |
| 익절 (`LADDER_TAKE_PROFIT`) | LIMIT (현행 유지) | LIMIT (현행 유지) |

익절을 그대로 두는 것이 §0.9다. 익절은 즉시성이 목적이 아니고, 시장가로 바꾸면 체결가
불확실성이 이익 쪽으로 열린다 — 보수 방향이 아니다. 보호만 바꾼다.

US도 MARKET인 근거: openapi가 US 온주 MARKET 매도에 제약을 두지 않는다. StockOS가 US에
MARKETABLE_LIMIT을 쓴 것은 KIS 계약 때문이며 토스 계약에는 그 제약이 없다. M12 는 US **조건주문**
SINGLE+MARKET 등록 실측이라 `/api/v1/orders` MARKET 접수의 증거가 아니다. `[미측정 — US MARKET 매도 실주문 없음]`.

> **US 한계 (2026-10-10, 재리뷰 P1-4 — 처분: US 를 Phase 2 에 유지하고 한계를 공개)**: Phase 2 는 US 에서
> **실측 없이 착지한다.** §5 실측은 KR 뿐이다(a100 runbook 도 「US 는 범위 밖」). 따라서 US 보호 청산의
> (가) 정규장 MARKET 매도 접수, (나) 프리/애프터 세션 응답, (다) 소수 수량 잔고의 시장가 매도 거동은 착지 시점에
> **전부 미측정**이고, `[미측정]` 태그는 게이트가 아니다. 첫 US 보호 청산이 사실상 첫 실측이 된다 — 그 응답이
> 거부이면 US 보호 청산은 오늘(LIMIT, 가끔 통과)보다 **나빠질 수 있다**. 이 위험을 알고 착지하는지는 §1~§3 착수 전
> 사람 확인 항목이다(tasks §5 표 아래 「US」 행). 세션 교집합은 design D7.

### 2. 게이트를 축소 주문에 한해 연다 — 두 관문 (②·③)

`checkOrderShape`(②)와 `placeIntentSupported`(③ — 엔진 인스턴스 한정)의 `orderType != "limit"` 거부에 **축소(매도)
시장가·가격 없음** 분기를 추가한다. **비분수** 진입(매수) 시장가는 두 관문 모두 계속 거부한다 — 그것이 riskcalc의 실제
규칙이고 노출 평가가 없다. (US 분수 시장가 매수 — 금액 기반 — 는 오늘처럼 기존 fractional 분기가 통과시킨다. 무변경.)

```text
side == "sell" && orderType == "market" && price 없음  → 허용
side == "sell" && orderType == "market" && price 있음  → 거부
side == "buy"  && orderType == "market"                → 거부 (현행 유지, US fractional 예외 존치)
```

③의 개방은 엔진 인스턴스에만 선다 — CLI·ops·MCP·그 인스턴스들의 `PreviewPlace` 는 upstream 과 동일하다(위 「차단은
세 곳이다」, 사용자 결정 2차 2026-10-10). 새 분기의 **위치**(기존 통화·수량 검사 뒤 — ②·③ 모두 조기 통과 금지)와
옵션의 모양은 design D3 이 정한다.

### 3. 원장이 잃는 것을 명시한다

시장가는 제출 시점에 가격이 없다. 원장의 `intents.price`는 비고, 체결가는 체결 조회로
확정된다. **그것이 정직한 기록이다** — 지금은 "받을 것으로 기대한 가격"을 적고 있고,
그 값은 체결가가 아니며, 위 5건은 아예 주문이 되지도 못했다.

관측·화면이 가격 없는 청산 의도를 표시할 수 있어야 한다.

## Impact

- **Specs**: `order-execution` (ADDED 3 — 보호 청산의 주문 유형, 축소 시장가의 게이트
  통과, 가격 없는 청산 의도의 원장·화면). **Phase 1 은 델타에 기여하지 않는다** — 하한가
  단은 Phase 2 가 제거하는 과도 상태라, 아카이브 시점의 정본에 남기면 낡은 요구가 된다.
  Phase 1 의 계약은 코드·시험·design D2a 로만 산다. (2026-09-30: Phase 1 불구현 — 남는 것은 평가기 핀 시험 1개)
- **Code**: `internal/app/engine/exitloop.go`(`sellIntent`·`submit` 호출부), `internal/execgw/failclosed.go`
  (`checkOrderShape` + 「internal/trading is not ours to change」 주석), **`internal/trading/service.go`
  (`placeIntentSupported` + 엔진 전용 생성자 옵션 — 기본값은 upstream 동일, 2026-10-10 사용자 결정 2차)**,
  **`internal/app/engine/engine.go`(엔진 `trading.Service` 생성에서만 옵션을 켠다)**, 관측·원장 표시부.
  ~~Phase 1 은 `PriceLimits` 배선 추가~~
  (불구현 — Phase 1 의 코드 산출물은 `internal/exitpolicy/a087_observed_price_pin_test.go` 시험 1개)
- **영향받는 표면 (무변경이어도 시험 대상)**: 전송 `official.buildOrderCreate` 와 그 사본 `execgw.PlaceWireBody`
  (`wirebody.go` — 재생 경로가 쓰는 wire body; 재리뷰 P3-3: 둘 다 MARKET 에 가격·TIF 를 싣지 않음 확인),
  **upstream 동일성 고정 대상** — 사람 CLI `tossctl order place`(`internal/app/app.go` 인스턴스), `tossctl ops`·**MCP
  `place_order`**(ops 쓰기 `internal/ops/write_operations.go` → MCP 카탈로그 `internal/mcp/catalog.go`; MCP 서버 인스턴스
  `cmd/tossctl/mcp.go`), 그 인스턴스들의 `PreviewPlace` 문구. 셋 다 기본 생성 경로라 비분수 market 매도는 계속
  `ErrPlaceUnsupported` 다(3차 재리뷰 A-P1-1)
- **Schema**: 없음 (`intents.price`는 이미 빈 값 허용). 청산 슬리피지(tasks 3.8)는 **기존 칸에서 유도**한다 — 결정 시
  관측가 `exit_events.observed_price`(같은 행 `proposed_intent_id`) 와 체결 평균가 `fill_snapshots.average_price`(그 intent 의
  브로커 `order_id`) — 새 열·표 없음. 그 결속(intent → order_id)이 기존 칸으로 성립하지 않으면 3.8 착수 전에 이 줄을
  schema 변경으로 고치고 재리뷰를 받는다(3차 재리뷰 A-P2-4)
- **§0.4**: 관측 루프 무변화. ~~Phase 1 이 더하는 것은 가격 부재 폴백 사건에서만 KR
  `PriceLimits` 읽기 1회~~ (불구현 — 새 브로커 호출 0)
- **§0.3**: 보호 청산이 호가 그리드·가격 밴드·시세 신선도 때문에 **로컬·브로커에서 거부되는 클래스**를 없앤다.
  ~~거부 자체가 사라지고 체결이 보장된다~~ — **정정(재리뷰 P1-2 · 2차 리뷰 C3)**: 시장가는 **접수**를 보장할 수
  있을 뿐(그것도 §5 실측 전에는 미확인) 갭다운·하한가·거래정지·얇은 호가에서 **체결**을 보장하지 않는다. 세션 밖
  거부 클래스는 다른 시계 구간으로 옮겨질 수 있다(design D7). 경계 조건 열거는 design D8
- **§0.9**: 익절은 무변경. 보호만 즉시성 방향으로

## Non-goals

- **호가 그리드 정본화** — **무소유**(재리뷰 P3-1: 「a088」은 `openspec/changes`·`archive` 어디에도 없다).
  익절 지정가·`flatten` fallback이 여전히 필요로 한다. StockOS 형태(십진 문자열·symbol 인자·direction 인자·거부 금지)
- **재가격 에스컬레이션** — **무소유**(a089 는 다른 내용 「나가지 못한 손절은 세어진다」로 2026-09-28 불구현
  아카이브됐다 — `64a1b2b3`). 7분 43초 공백은 주문 가능한 제안이 없던 구간이다(Why 정정)
- **1차 리뷰 A1~A14 · I2(400 분류)** — **무소유**. a094 R1 이 가져간 것은 `opposite-pending-order-exists` 하나다
- **진입 주문 유형.** riskcalc의 규칙이 실제로 적용되는 곳이고 바꾸지 않는다
- **조건주문(브로커측 보호).** 2c 범위
- **`flatten`.** 이미 거래소 하한가를 1순위로 쓴다(무변경). 같은 하한가 지정가를 **보호 청산의 가격으로** 쓰는
  원안은 평가되지 않은 채 남아 있었다 — design D6 에 **미채택** 기록으로 남긴다(시장가 방향 사용자 결정 2026-10-10 에 따른
  미채택 — 사용자가 원안을 보고 기각한 기록은 없고, 원안 자체 평가는 미수행. 3차 재리뷰 A-P2-3)

## 실측 필요 (사용자 승인 항목, §0.7)

1. **KR MARKET 매도 1회 실주문.** 스키마는 지원하지만 실접수는 미측정이다. 최소 수량으로
   장중 1회. 성공·실패 모두 `measurements.md`에 기록. **실행 수단은 별도 실측 도구
   `tools/a087-market-sell-probe/`** 다(사용자 결정 2026-10-10 — `tossctl order place` 는 ③에서 브로커에 닿기 전에
   로컬 거부되므로 그 결과는 거짓 음성이다, 재리뷰 P1-1). 실행은 사람 터미널 + 주문별 승인. 결과 → 처분 표는 tasks §5
2. **세션 경계 거동.** KRX 시간외단일가에는 시장가가 없다. 정규장 밖 MARKET 매도의 응답이 미측정이고, 예상
   응답 집합은 `422 order-hours-closed` **와 `422 order-type-not-allowed`**(openapi `POST /api/v1/orders` 422 예시
   「현재 사용할 수 없는 호가 유형」, 재리뷰 P1-4) 그리고 그 밖의 코드다.
   ~~엔진 청산 루프는 정규장 기준이라 차단 요소는 아니지만~~ — **정정(재리뷰 P1-2 · 2차 리뷰 H1)**: 엔진에
   `InRegularSession` 호출자가 없다(비시험 호출자는 `verifylive` 하나). 엔진이 정규장 밖에서 보호 청산을 제출하는지와
   그때의 처분은 design D7 — **미해결, §1~§3 착수 전 확인**
3. **US** — 실측 없음. 한계는 위 What Changes §1 「US 한계」
