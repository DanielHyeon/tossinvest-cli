# a087 설계 — 보호 청산에서 가격을 없앤다

## 문제의 형태

초안은 "제출가가 호가 그리드 위에 있게 하자"였다. proposal-freeze 리뷰가 그 설계를 무너뜨렸고
(`review.md`), 무너진 자리에서 더 단순한 질문이 남았다.

**보호 청산이 왜 가격을 갖는가?**

답은 없었다. 코드가 인용한 근거(`riskcalc`의 LIMIT 전용)는 진입 전용 규칙이었고, 코드가
실제로 적은 이유는 "원장에 적을 가격이 필요해서"였다. 거부되지 않을 주문 형태를 원장 표기와 맞바꾼 것이다.
(3판까지 「체결 보장을 원장 표기와 맞바꿨다」였다 — **정정(재리뷰 P1-2 · 2차 리뷰 C3)**: 시장가는 체결을 보장하지
않는다. 경계 조건은 D8.) 그리고 2차 리뷰가 찾은 **셋째 이유**가 있다 — 이 fork 의 서비스 계층
(`internal/trading` `placeIntentSupported`)이 비분수 주문에 대해 지원하는 유일한 형태가 LIMIT 이다(D3).

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
둘이 되고, 그 둘이 갈라지는 순간 한쪽만 시장가가 된다. [`exitloop.go`](../../../internal/app/engine/exitloop.go)의
`isProtective`(패키지 함수)는 이미 §0.3 근거로 존재하고 주석이 그 이유를 적어 뒀다 — "Nothing may
withhold one of these: §0.3 forbids weakening or delaying the immediacy of a stop".
같은 §0.3이 이 change의 근거이므로 같은 술어여야 한다.

**소비자가 셋이 된다 (재리뷰 P3-2).** base 이후 `isProtective` 는 재판정 보류 예외(`record`)와 a091 floor 의미
(`submit`)에 쓰인다(CodeGraph 1.6.0 비시험 호출자 = `record`·`submit` — `analysis/gate-0.4-codegraph.md`). a087 의
유형 결정이 셋째다. 그러므로 a087 은 **술어 자체를 바꾸지 않는다** — 바꾸면 세 소비자가 함께 움직인다. tasks 2.1 의
risk-pattern-report 가 세 소비자를 적고, 2.2 가 술어 불변을 시험으로 고정한다. (각 소비자 안의 분기 위치는 검증
필요 — AST 미작성, task 2.1 소관.)

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

`sellIntent`(base `102d4e99` 좌표 exitloop.go:1571-1578 — HEAD 에서는 함수 머리 뒤 가격 사다리, 본문 바이트 동일
— 재리뷰 P2-5)의 가격 사다리:

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

열거 결과 모집단은 **0** 이다(이 절의 줄 좌표는 전부 base `102d4e99` 기준 기록이다). 위 사다리의 폴백(B1 `:1573`)과
거부(B2 `:1576`)는 생산 경로에서 도달할 수 없다.
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
245,750 그리드 불일치 사건(옛 표기 「9분 사건」 — 척도는 「6회 결정 중 5회 미제출」, 관측가는 **있었다**)을 고치는 것은 D2a 가 아니라 D2b 다:
D2a 는 가격 부재 거부만 다룬다.

## D3 — 게이트는 축소에 한해서만 연다 (두 관문 — 2026-10-10 개정)

> **개정 (재리뷰 P0-1 · 사용자 결정 2026-10-10)**: 3판의 D3 는 `checkOrderShape`(②) 하나만 다뤘다. 엔진 주문은
> 그 뒤에 `trading.Service.Place` → `placeIntentSupported`(③)를 지나고, ③은 비분수 비지정가를 `ErrPlaceUnsupported`
> 로 거부한다. **②만 열면 보호 청산은 100% 로컬 거부된다.** 사용자 결정으로 ③도 같은 모양으로 연다.
>
> **재개정 (3차 재리뷰 A-P1-1 · 사용자 결정 2차 2026-10-10)**: ③ 개방은 **엔진 인스턴스 한정**이다(아래 ③). CLI·ops·MCP
> 기본 경로는 upstream 과 바이트 동일. 1차 결정의 「사람 CLI upstream 동작 변경 수용」은 철회됐다.

### ② `checkOrderShape` (`internal/execgw/failclosed.go`, 호출자 `CheckPlace` 1)

```go
if orderType != "limit" {
    return reject(ReasonUnsupportedOrderType,
        "only limit orders (and US fractional market orders) are supported, got %q", …)
}
```

이 거부가 진입과 청산을 구분하지 않는다. 구분을 넣는다.

```text
fractional          → 기존 분기 그대로 (US market only)
sell + market       → 통과. 단 Price != 0 이면 거부
buy  + market       → 거부 (현행 유지)
그 외 non-limit     → 거부 (현행 유지)
```

**새 분기의 위치 (재리뷰 P2-2 — 검증 필요, AST 미작성: task 1.0.1)**: 재리뷰의 손 읽기는 `orderType != "limit"` 거부
**뒤에** KR 통화 검사와 수량 양수 검사가 온다고 적었다. 그 자리에 `sell+market` 의 **조기 통과 반환**을 넣으면 그 검사들을
건너뛴다. 그러므로 새 분기는 **통과를 반환하지 않는다** — 비지정가 거부의 예외 조건으로만 서고, 뒤의 통화·수량 검사는
`sell+market` 에도 그대로 적용돼야 한다(또는 같은 검사를 반복). 실제 순서는 task 1.0.1 의 AST 열거로 확정하고, RED 는
「`sell+market` + 비 KRW 통화(KR)」·「`sell+market` + 수량 0」이 여전히 거부되는 행을 포함한다(tasks 1.1).

**마지막 가격 검사는 limit 전용으로 남는다 (3차 재리뷰 A-P2-2 — 검증 필요, AST 미작성: task 1.0.1)**: 손 읽기로는
`checkOrderShape` 의 맨 끝 거부 「a limit order needs a positive price」(`intent.Price <= 0`)가 수량 검사 뒤에 선다. 이 검사가
`sell+market` 에 적용되면 가격 없는 시장가가 전부 거부된다. 그러므로 `sell+market` 에서는 이 검사 대신 위의 `Price != 0`
거부가 선다 — 「limit 이면 `Price <= 0` 거부, sell+market 이면 `Price != 0` 거부」로 갈래가 나뉘고, 통화·수량 검사는 둘 다에
공통이다. RED 1.1 의 「`sell+market`·가격 없음 → 통과」 행이 이 갈래를 잡지만 설계가 침묵하지 않도록 여기 적는다.

**`Price != 0` 검사를 넣는 이유**: openapi가 `MARKET`에 `price` 전달을 금지하고 전달 시
`400 invalid-request`를 준다. ~~`orderintent`가 이미 정규화하므로 게이트는 두 번째 확인이다~~ — **정정(재리뷰 P2-2)**:
`orderintent` 의 MARKET → `Price = 0` 정규화는 `NormalizePlace` 안에 있고, `sellIntent` 는 `PlaceIntent` 를 **구조체
리터럴로 조립**하므로 그 정규화를 지나지 않는다. 그러므로 `checkOrderShape` 의 `Price != 0` 거부는 「두 번째 확인」이 아니라
**이 경로의 유일한 로컬 확인**이다. (`checkOrderShape` 주석의 「strict subset filter」는 이중 확인의 근거가 아니라
「진짜 검사는 서비스에 있다」는 뜻이었다 — 그것이 ③이다.) 결론(거부를 넣는다)은 같다.

**비분수 MARKET 의 정수 수량 (재리뷰 P2-2)**: openapi 는 소수 수량을 US 시장가 매도에만 허용하고 그 외는 400 이다.
현재 어느 관문이 비분수 MARKET 의 정수성을 검사하는지는 **미확인**(검증 필요 — task 1.0.1·1.5.1 AST). 처분은 그 열거
뒤 tasks 1.1a 에서 정한다 — 로컬 거부를 넣거나, 브로커 400 에 맡기고 그 사실을 시험 이름으로 고정하거나. 어느 쪽이든
침묵한 생략은 없다.

### ③ `placeIntentSupported` (`internal/trading/service.go`, 비시험 호출자 `PreviewPlace`·`Place`) — 엔진 인스턴스 한정

```go
// 비분수 주문 (재리뷰 인용, 검증 필요 — AST 미작성: task 1.5.1)
if intent.OrderType != "limit" { return false }   // Place → ErrPlaceUnsupported
// (손 읽기) 그 뒤 KR → CurrencyMode == "KRW", US → "KRW" | "USD"
```

**누가 이 술어를 지나는가 (3차 재리뷰 A-P1-1)**: `trading.NewService` 로 만든 인스턴스 셋 — 엔진
(`internal/app/engine/engine.go` `OrderPath.Trading`, 소비자 `execgw` 게이트웨이·엔진 `tradingService`), CLI 앱
(`internal/app/app.go` → `tossctl order place`·`tossctl ops`), **MCP 서버**(`cmd/tossctl/mcp.go` — ops `place_order` 가
`internal/mcp/catalog.go` 로 노출되는 에이전트 표면). 셋은 같은 `cfg.Trading` 을 공유하므로 config 토글로는 엔진과 사람·에이전트
경로를 가를 수 없다. 갈라지는 것은 **인스턴스**다.

**처분 (사용자 결정 2차 2026-10-10)**: 개방은 엔진 인스턴스에만 생성자 옵션으로 선다. 모양은 **제안**이고 코드는 tasks 1.5.3
몫이다 — 기존 빌더 관례(`WithLineage`·`WithConditional`)를 따라:

```go
// 제안 (tasks 1.5.3 에서 확정)
func (s *Service) WithProtectiveMarketSell() *Service { s.protectiveMarketSell = true; return s }

// engine.go — 엔진 인스턴스에서만 켠다
trading.NewService(cfg.Trading, broker).WithConditional(broker).WithProtectiveMarketSell()

// 술어는 인스턴스 값을 인자로 받는다 — PreviewPlace·Place 두 호출자가 같은 값을 넘긴다
func placeIntentSupported(intent orderintent.PlaceIntent, protectiveMarketSell bool) bool
```

- **기본값(옵션 없음)** — `app.go`·`mcp.go`·시험의 모든 `NewService` 는 옵션을 켜지 않으므로 비분수 non-limit 은 오늘처럼
  `false` → `Place` 가 `ErrPlaceUnsupported`, `PreviewPlace` 의 `LiveReady`·경고 문구도 upstream 과 같다. 이 동일성을 시험으로
  고정한다(tasks 1.5.5).
- **엔진 인스턴스(옵션 켬)** — ②와 **같은 모양**으로만 연다: `sell + market + 가격 없음` → 지원. `buy + market` → 비지원
  (시험 고정). `sell + market + 가격 있음` → 비지원. fractional·limit 분기는 무변화.
- **두 호출자가 같은 값을 써야 한다** — `Place` 는 지원 판정 뒤 `guard(…, s.PreviewPlace(intent), …)` 를 부르고(손 읽기,
  검증 필요 — AST 미작성: task 1.5.1), 엔진 게이트웨이도 `PreviewPlace(i).ConfirmToken` 을 쓴다. 한쪽만 옵션을 보면
  `Place` 는 지원·`PreviewPlace` 는 비지원으로 갈라진다. 그러므로 옵션은 술어의 인자로 들어가고, 1.5.2 RED 가 엔진 인스턴스의
  `PreviewPlace(sell+market).LiveReady` 를 함께 잰다.
- **조기 통과 금지 (3차 재리뷰 A-P2-2 — P2-2 를 ③ 에 적용)**: 손 읽기로는 `OrderType != "limit"` 거부 **뒤에** 시장별 통화
  검사가 온다. `sell+market` 예외를 조기 `return true` 로 넣으면 그 통화 검사를 건너뛴다. 새 분기는 비지정가 거부의 예외 조건으로만
  서고, 뒤의 통화 검사(KR → KRW, US → KRW|USD)는 `sell+market` 에도 그대로 적용된다. 엔진 경로에는 ② 가 앞에 있지만 ③ 의 단독
  시험(1.5.2)은 ② 없이 이 규칙을 잰다 — 통화 행을 포함한다.
- **문구** — `PreviewPlace` 의 「Live place supports … limit orders … and US fractional (market) orders.」는 기본 인스턴스에서
  바이트 그대로 둔다. 엔진 인스턴스에서 그 문장을 바꿀지(또는 한 줄 덧붙일지)는 1.5.4 가 정하되, 기본 인스턴스의 문구 바이트는
  바뀌지 않는다. `checkOrderShape` 주석 「internal/trading is not ours to change — design D1」(execgw 측 설계의 D1)은 이 change 가
  `internal/trading` 에 옵션을 더하므로 「기본 동작은 upstream 그대로, 엔진 인스턴스만 옵션」으로 고친다(1.5.4).
- **불변식 3** — 엔진이 꺼져 있으면 엔진 인스턴스가 만들어지지 않으므로 이 개방의 흔적은 0 이다. CLI·ops·MCP 는 엔진 토글과
  무관하게 upstream 동일. `issues.md` I-R1 종결.
- **엔진 인스턴스 안의 다른 생산자** — 엔진 인스턴스를 지나는 매도 조립은 `sellIntent`·`flatten` 둘이고(2차 재리뷰 P3-3,
  `flatten` 은 LIMIT) a112 전략 레인은 매수 전용이다. 그 열거를 1.5.1 risk-pattern-report 가 CodeGraph 로 다시 적는다(엔진
  `tradingService` 소비자 포함). 엔진 인스턴스를 쓰는 사람 조작 표면이 발견되면 그 표면은 엔진 토글 뒤에 있으므로 불변식 3 은
  유지되지만, 그 사실을 1.5.0 Pre-Edit 에 적는다.

**②와 ③의 판정이 갈라질 위험**: 같은 규칙이 두 자리에 산다. 한쪽만 바뀌면 다른 쪽이 그 시험을 대신 통과시킨다
([[two-judgements-cover-for-each-other]]). 그러므로 RED 는 두 관문을 **각각 단독으로** 잰다(한 관문을 우회한 직접 호출
시험 각 1벌) — tasks 1.1·1.5.2.

### ②·③ 손 읽기 주장의 AST 대조 (2026-10-28, HEAD `0c9bfd272` — 0.10 리뷰 증거)

위의 「검증 필요 — AST 미작성」 주장 넷을 HEAD AST(`analysis/evidence/execgw--checkordershape.ast.json` ·
`trading--placeintentsupported.ast.json` · `trading--service.place.ast.json` · `trading--service.previewplace.ast.json`, 열거
`analysis/evidence/enumerate-output.md`)로 대조했다. 이 대조는 구현 전 FLM(1.0.1·1.5.1)을 **대체하지 않는다** — 그 번들은 착수 시 HEAD 에서
다시 만든다. base `102d4e99` 이후 `failclosed.go` 의 변경은 `AllReasonCodes` 의 6줄(a112 `3260f4eb6` · a094 `766a8456c`)뿐이고
`checkOrderShape`·`trading/service.go` 본문은 base 와 같다(`git diff 102d4e99 HEAD`).

- **② 위치**: `checkOrderShape` B10(`:84` `orderType != "limit"` 거부) 뒤에 B11(`:88` KR 통화) · B12(`:91` 수량 `<= 0`) · B13(`:94` 가격 `<= 0`)이
  온다. 손 읽기와 같다 — `sell+market` 예외를 B10 의 조기 `return nil` 로 넣으면 B11·B12 를 건너뛰고, B13 은 `sell+market` 에서 `Price != 0` 거부로
  갈라야 한다.
- **③ 위치**: `placeIntentSupported` B3(`:281` 비지정가 `return false`) 뒤에 B4(`:284` KR → KRW) 와 마지막 반환(`:287` US → KRW|USD)이 온다. 손 읽기와 같다.
  ③ 에는 **수량·가격 검사가 없다**(분기 4개 전부 시장·분수·유형·통화).
- **정수 수량 (1.1a 입력)**: ② B12 는 `<= 0` 만, ③ 은 수량을 보지 않는다 — **어느 관문도 비분수 MARKET 의 정수성을 검사하지 않는다.** 1.1a 의
  처분(로컬 거부 추가 / 브로커 400 위임)은 이 사실에서 출발한다.
- **두 호출자**: `Service.Place` B1(`:188`)이 술어를 부르고 B4(`:200`)에서 `s.guard(…, s.PreviewPlace(intent), …)` 를 부른다. `PreviewPlace`(:109-143,
  분기 6)는 `:115` 에서 술어를 다시 부른다. 한쪽만 옵션을 보면 갈라진다는 위 「두 호출자가 같은 값」 주장과 같다.
- **문자열 정규화 차이**: ② 는 `ToLower`·`TrimSpace` 로 정규화하고(`:54-57`), ③ 은 `"kr"`·`"market"` 을 그대로 비교한다. 엔진 `sellIntent` 는
  `m.position.Market` 을 그대로 싣는다. 오늘 LIMIT 보호 청산이 ③ 을 지나므로 생산 값은 소문자다 — 새 판정(D7 「구간 입력」 조건 1)은 `clock.ParseMarket`
  으로 정규화하고, 실패는 LIMIT 이다.

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

> **정정 (재리뷰 P1-2 · 2차 리뷰 C2)**: 3판의 첫 행은 「KRX 긴급 청산 = MARKET → **승계**」였다. StockOS 의 평범한
> 손절 1차 제출은 LIMIT 이고 MARKET 은 0.5% 초과 이탈·EOD 에스컬레이션에서만 쓰인다(`_apply_exit_order_style`).
> a087 은 1차부터 MARKET 이라 그 행은 승계가 아니라 **이탈**이다. StockOS 는 a087 의 근거가 아니다(proposal 「StockOS 대조」).
> 「a088로」·「a089로」 행은 대상이 없다(재리뷰 P3-1) — **무소유**로 고친다.

| StockOS | 위치 | a087 판정 |
| --- | --- | --- |
| 1차 손절 = LIMIT, 0.5% 초과 이탈·EOD 에서만 KRX MARKET | `auto_exit_execution.py` `_apply_exit_order_style`·`_aggressive_exit_order_type` | **이탈** — a087 은 임계 없이 1차부터 MARKET. 근거는 사용자 결정·§5 실측(D6) |
| 긴급 청산은 지연 게이트 전부 우회 | `auto_exit_execution.py:3130` | **무소유** |
| N회 후 MARKETABLE_LIMIT → true MARKET 승격 | `auto_exit_execution.py:3160` | **무소유** |
| US = MARKETABLE_LIMIT | 같은 파일 | **미승계** — KIS 계약 제약이고 토스에는 없다 |
| 호가 정렬 `Decimal(str(price))` | `auto_exit_execution.py:3001` | **무소유** |
| 정렬 실패 시 fallback, 거부 없음 | `auto_exit_execution.py:3005-3017` | **무소유** |
| `TickSizeProvider` + symbol override | `tick_size_provider.py` | **무소유** |

StockOS가 US에 marketable limit을 쓰는 것은 `protective_order_capability`가 기록한 KIS의
한계 때문이다 — `kis_us_stock_rest_stop_oco_not_supported`. 토스 openapi에는 그 제약이
없고 US 온주 MARKET 매도를 금지하는 문장이 없다. 그래서 양 시장 모두 MARKET으로 간다.
`[미측정 — US MARKET 매도 실주문 없음]` — 이 한계의 처분은 proposal What Changes §1 「US 한계」.

## D6 — 거래소 하한가 지정가 원안 (미채택 기록 — 재리뷰 P1-3, 3차 A-P2-3 정정)

**원안은 평가된 적이 없다.** 2026-08-06 2차 리뷰 C3·「3차 교정 방향」 2 는 보호 청산의 가격으로 **거래소 하한가
지정가**를 쓰라고 했다(`flatten` 이 이미 그렇게 한다 — `internal/flatten/liquidate.go` 「The exchange's own floor」).
D2a 는 그것을 관측가·기준선 **뒤의 셋째 폴백 단**으로 좁혔고, 그 폴백이 도달 불가로 반증됐다(D2a 「반증」). 원안 —
보호 제안에서 관측가 **대신** 하한가 — 은 그 반증과 무관하게 비교되지 않은 채 남았다(proposal 이 「그리드 불일치 사건(옛 표기 「9분 사건」)은 Phase 1 이
고치지 않는다」고 적은 것이 그 증거다).

재리뷰가 적은 원안의 성질(비교 기록 — 측정하지 않은 값은 표기대로):

| 축 | 하한가 지정가 원안 | 시장가 (a087) |
| --- | --- | --- |
| 세 관문 | 전부 통과(여전히 LIMIT — ③ 변경 불요) | ②·③ 변경 필요(D3) — ③ 은 엔진 인스턴스 옵션, upstream 동작 변화 0(2차 결정) |
| 245,750 사건 | 해소(거래소 공표값 = 온그리드) | 해소(가격 없음) |
| 원장 가격 | 남는다 | 비운다(D4·§3 화면 작업 필요) |
| 시장 | **KR 전용**(`client/marketdata.go` 「미국장은 일일 가격제한 제도가 없음」) | KR 만(7판 — US 는 실측 전까지 LIMIT, D7 Q2) |
| 브로커 호출 | 보호 청산마다 `PriceLimits` GET 1회(§0.4) | 추가 0 |
| 시간외단일가 | 가격 범위 밖일 가능성 `[미측정]` | MARKET 불가 → 오늘과 같은 LIMIT(D7 처분 2026-10-22) |
| 체결 품질 | 하한가 도달·매수 호가 공백 시 미체결 | 같음(D8) |

**처분: 미채택 — 시장가 방향 사용자 결정(2026-10-10)에 따른 것이며, 원안 자체 평가는 미수행.** (정정 — 3차 재리뷰
A-P2-3: 4판은 「기각 (사용자 결정)」이라 적었으나 사용자 결정 기록에 원안은 나오지 않는다. 「(가)를 골랐으니 (다)는 기각」은
추론이었다.) 재리뷰가 P0-1 의 처분으로 (가) ③ 수정 (나) 우회 (다) 이 원안을 제시했고, 사용자는 (가) — **시장가 방향 유지** —
를 택했다(2차 결정으로 ③ 개방은 엔진 인스턴스 한정, D3). 이 표는 그 결정이 무엇을 받아들였는지의 기록이다(원장 가격 대신
③ 의 엔진 전용 옵션·원장/화면 작업, 대신 US 동시 해소). 「세 관문」 행의 비용은 2차 결정으로 줄었다 — upstream 동작 변화가
없어졌고 남는 비용은 `internal/trading` 의 옵션 하나다. 재개 조건: §5.1 이 MARKET 매도를 거부하면(tasks §5 표) 이 원안이
재검토 후보가 된다 — 그 판단은 사람 결정이다. 원안의 문언 기각을 받을지는 사람 판단으로 남긴다(`issues.md` I-R3).

**계기 (재리뷰 P1-3 · 2차 리뷰 「3차 교정」 3·4)**: 현재 `slippagePct` 는 진입 전용이고 청산 슬리피지를 재는 코드가
없다(재리뷰 인용). §5.1 은 접수만 재고 체결 품질을 재지 않는다. 그러므로 시장가가 원안보다 나은지는 착지 뒤 데이터로만
답할 수 있고, 그 데이터를 남기는 계기를 tasks 3.8 에 둔다(보호 청산 체결가 vs 결정 시 관측가, bps). bps 로 두 방식을
다시 비교할지와 그 임계는 이 change 에서 정하지 않는다(`issues.md` I-R3).

## D7 — 세션 교집합: MARKET 불가 ∩ LIMIT 금지 = 제출 불가 (재리뷰 P1-4)

spec Requirement 1 은 보호 청산의 지정가를 **금지**한다(SHALL NOT). 브로커가 MARKET 은 받지 않고 LIMIT 은 받는
구간에서는 두 조건의 교집합으로 보호 청산의 **제출 가능성이 0** 이 된다:

| 구간 | MARKET | LIMIT | a087 후 보호 청산 |
| --- | --- | --- | --- |
| KRX 정규장 | §5.1 측정 대상 | 가능(오늘) | MARKET — §5.1 결과에 달림 |
| KRX 시간외단일가 | **없음**(거래소 제도 — proposal 스스로 인정) | 가능 | **제출 불가** |
| KRX 장 마감 후·SOR/확장 세션 | `[미측정]` — §5.2 | `[미측정]` | §5.2 결과에 달림 |
| US 정규장 | `[미측정]` | 가능(오늘) | 미측정 착지(proposal 「US 한계」) |
| US 프리/애프터 | `[미측정]` | `[미측정]` | 미측정 — MARKET 불가이면 **제출 불가** |

**엔진의 세션 게이트 유무: 미해결, §1~§3 착수 전 확인.** 재리뷰는 HEAD 에서 `InRegularSession`(`clock/market.go`)의
비시험 호출자가 `verifylive/hours.go` 하나뿐이고 `internal/app/engine` 에는 0 이라고 적었다(CodeGraph·`rg`). 그러나
엔진이 다른 수단(관측 루프 기동 시간대·시세 만료 판정 등)으로 정규장 밖 판정을 사실상 막는지는 **이 문서가 모른다**
(검증 필요 — 새 분기 주장 아님). 착수 전 확인 방법: 엔진의 청산 판정 상류(`ObserveOnce`·`judge`)를 CodeGraph 로 따라가
세션·시각 입력이 있는지 열거한다(tasks 0.5).

처분은 그 확인과 §5.2 결과 **둘을 입력으로** 사람이 정한다. 문서가 미리 정하지 않는다. 다만 무효화 대상은 지금 명시한다 —
교집합 구간이 실재하면(엔진이 그 구간에 판정하고, 브로커가 MARKET 을 거부하면):

- spec Requirement 1 의 「지정가로 제출해서는 안 된다(SHALL NOT)」는 그 구간에서 **보호 청산을 0 으로 만든다** → 세션별
  유형(정규장 MARKET, 그 밖 LIMIT) 같은 축소가 필요해지고, 그것은 새 설계(세션 입력 배선)라 이 개정이 하지 않는다.
- tasks 2.4 의 「보호 제안일 때 가격을 읽지 않는다」는 그 구간에서 LIMIT 가격이 필요하므로 무효가 된다.
- 엔진이 그 구간에 판정하지 않음이 확인되면 교집합은 공집합이고 위 두 줄은 유지된다.

### 처분 (사용자 결정 2026-10-22) — 시장가를 받지 않는 구간은 오늘처럼 지정가

입력 둘이 나왔다. ① tasks 0.5(2026-10-20, `analysis/gate-0.5-session-inputs.md`): 청산 판정 상류 AST 12개에서 **세션 입력 0 ·
시각 입력 40(전부 경과 시간)** — 엔진에 세션 게이트가 없다. 정규장 밖에도 시세가 오면 판정·제출 시도가 일어난다. 그래서 위
「교집합 공집합」 줄은 성립하지 않는다. ② §5.1(2026-10-19, M50): KRX 정규장 MARKET 매도 1주 접수·체결완료. §5.2 는 미측정이다.

Manager 질문 「시장가가 안 되는 시간대에 손절을 어떻게 낼지. 권장: 그 시간대에는 지금처럼 지정가」에 사용자가 그 권장대로 답했다.
그래서 spec Requirement 1 을 다음처럼 개정했다(requirement 수준 변경 — WORKFLOW 「리뷰 게이트」에 따라 수정분 리뷰를 다시 받는다, tasks 0.10).

> **7판 개정 (사용자 결정 2026-10-28 「권장대로」 — `issues.md` 「사용자 결정 (2026-10-28)」)**: 아래 표는 Q2·Q3·Q4 를 반영한
> 판본이다. 바뀐 칸은 넷이다 — ① **US 정규장 MARKET → LIMIT**(Q2, 실측 전까지), ② KRX 정규장 MARKET 칸을 **브로커 달력이 정상 영업일로
> 보이는 날**로 좁혔다(Q3 — 휴장·조기 폐장·개장 지연일은 종일 LIMIT), ③ 「§5.2 가 접수를 보이기 전까지」 문구를 **LIMIT 확정**으로 바꿨다
> (Q4 — §5.2 not-applicable), ④ 종가단일가(15:20~15:30)를 LIMIT 으로 뺐다(Q1 원칙 적용 — `review.md` 0.10 P2-4). 6판 표의 US 정규장 행 「MARKET — 미측정 착지」는 **철회**됐다. 구간 판정 규칙은 아래 「구간 입력」 절.

| 구간 | 보호 청산 유형 | 근거 |
| --- | --- | --- |
| KRX 정규장 접속매매 — 브로커 달력이 그날을 **정상 영업일**로 보이고(아래 「구간 입력」 조건 1~5) 그 시각이 09:00~15:20 KST 안 | MARKET | §5.1 접수 실측(M50, 14:59 접속매매) |
| KRX 정규장 종가단일가(15:20~15:30) | **LIMIT(오늘과 같음)** | Q1 원칙(접수 미측정 → LIMIT)을 이 하위 구간에 적용 — §5.1 은 이 구간을 재지 않았다. 문서 유도, Manager 확인(`review.md` 0.10 P2-4) |
| KRX 휴장일 · 조기 폐장일 · 개장 지연일 (달력의 그날 정규장이 정상 영업일 모양이 아님) | **LIMIT — 그날 종일** | Q3(조기 폐장일 13:00~15:30 MARKET 누출 방지) + 아래 「구간 입력」 해석 |
| KRX 시간외단일가 | **LIMIT(오늘과 같음)** | 거래소 제도상 시장가 없음 |
| KRX 장 마감 후·장 개시 전 그 밖의 구간 | **LIMIT 확정** | 사용자 결정 2026-10-22 · Q4(§5.2 not-applicable — 이 칸을 옮길 실측을 하지 않는다) |
| US 정규장 | **LIMIT — 실측 전까지** | Q2(2026-10-28). US MARKET 접수 실측 0(M12 는 조건주문). 6판의 「미측정 착지」 철회 |
| US 프리/애프터 | **LIMIT** | 사용자 결정 2026-10-22 · Q2 |
| 구간 판정 입력을 못 읽음 — 달력 오류·결측·낡음·시계 역행·시장 문자열 판정 불가 | **LIMIT** | Q1(불변식 4 유도 수락) · Q3(달력 오류·결측이면 LIMIT) |

LIMIT 의 가격은 오늘 `sellIntent` 사다리 그대로다(관측가 → 기준선 → 거부 — `analysis/function-logic/internal-app-engine--exitobserver.sellintent/`
base `102d4e99` 산출물. HEAD `0c9bfd272` AST `analysis/evidence/engine--exitobserver.sellintent.ast.json` 도 분기 4·반환 4 로 같다. 구현 때 HEAD 에서 FLM 으로 다시 만든다 — tasks 2.1).

**불변식 4 근거 (약화 없음).** 기준선은 오늘의 동작(보호 청산은 언제나 LIMIT)이다. 이 처분은 오늘보다 나빠지는 칸을 만들지 않는다.

- MARKET 칸은 접수가 실측된 구간 하나 — 정상 영업일의 KRX 정규장 접속매매(09:00~15:20) — 뿐이다(7판: US 정규장 칸 삭제 Q2 · 종가단일가 제외 P2-4).
- 그 밖의 칸은 LIMIT 이라 오늘과 바이트 같은 주문을 낸다.
- 접수가 실측되지 않은 구간과 판정 입력이 없는 경우를 LIMIT 으로 둔 것은 이 문서의 유도다. 사용자 문언은 「시장가가 안 되는 시간대」다.
  모르는 구간에 MARKET 을 내면 거부될 수 있고, 같은 구간에 LIMIT 은 오늘 접수될 수 있다. 그러면 그 칸은 오늘보다 약해진다(§0.3 위반).
  그래서 모름은 LIMIT 으로 기운다. **이 유도는 사용자가 수락했다(2026-10-28 Q1).**
- 반대 방향(실제로는 MARKET 을 받는 구간인데 LIMIT 을 냄)은 오늘과 같은 동작이다. 그러므로 약화가 아니다. 이득을 덜 얻을 뿐이다.

**무효화된 줄.**

- 위 「spec Requirement 1 … 보호 청산을 0 으로 만든다」 — 개정으로 해소. 교집합 구간에서 보호 청산은 LIMIT 으로 나간다.
- tasks 2.4 「보호 제안일 때 가격을 읽지 않는다」 — **MARKET 구간에서만** 참이다. LIMIT 구간에서는 오늘처럼 가격을 읽는다. tasks 2.4·2.5 를 그에 맞춰 고쳤다.

**새 설계 의무 (구현 task 2.6).** 엔진은 오늘 세션 입력을 하나도 읽지 않는다(0.5). 그래서 구현은 MARKET/LIMIT 을 가를 **구간·접수
입력**을 새로 배선해야 한다. ~~이 입력의 출처는 이 문서가 정하지 않는다.~~ 출처는 7판에서 정해졌다(아래 「구간 입력」, Q3).
`sellIntent` 시그니처에 시각 또는 구간 인자가 더해지면 2.1 FLM 의 분기 집합이 바뀐다. 그래서 2.1·2.6 은 같은 로트에서 재생성한다.

### 구간 입력 (7판 — 사용자 결정 2026-10-28 Q3)

**출처 = 브로커 시장 달력 + 고정 시각.** 달력 읽기는 `official.Client.TypedMarketCalendar`(`internal/official/typed_calendar_reads.go:37`)이고,
엔진 진입 쪽 `strategyScheduleAuthorityLoader.prepare`(`strategy_schedule_authority.go:220`)가 이미 같은 함수를 부른다. **달력 오류·결측이면 LIMIT.**
분기 근거는 손 읽기가 아니라 HEAD `0c9bfd272` 의 AST 다 — `analysis/evidence/*.ast.json`, 열거 `analysis/evidence/enumerate-output.md`
(재생성 `python3 -I …/analysis/evidence/enumerate.py`, `source_sha256` 대조 stale 0).

**`clock` 술어만으로는 안 되는 이유 (AST).** `Market.RegularSession`(`clock/market.go:122-143`)의 분기는 B1(`Location` 오류)·B2(시장 표 없음)
둘뿐이고 성공 반환(:136)은 고정 표 `regularHours`(:55-58, KR 09:00~15:30)로 창을 만든다. `Session.Contains`(:113-118)는 B1 에서 주말·0 값만 거른다.
휴장일·조기 폐장·개장 지연을 가르는 분기는 **없다** — 파일 주석(:49-54)도 휴일은 브로커 달력 몫이라 적는다. 그래서 `InRegularSession` 만 쓰면
조기 폐장일 13:00~15:30 에 MARKET 이 나간다(Q3 의 누출).

**달력 함수의 분기 (AST).**

- `TypedMarketCalendar`: B1 나라 문자열(KR·US 외 오류) · B2 날짜 검증 · B4 GET 오류 → 오류 반환. 응답을 해석하지 않고 그대로 돌려준다.
- `scheduler.adaptCalendarDay`(`scheduler/calendar.go:103-134`): KR 은 B4·B5 에서 `Integrated.RegularMarket` 만 읽는다. B8 — 세션이 없으면
  `Regular == nil` 인 **정상 반환**(오류 아님 — 휴장일 모양). B9 — 시작·끝 0 값·역순 오류, B10 — 거래소 현지 날짜 밖 오류.
- `scheduler.AdaptOfficialCalendar`(`:56-101`): B6 이전 < 오늘 < 다음 영업일 순서, B7 이전·다음 영업일의 정규장 존재를 요구한다. **오늘의
  `Regular == nil` 은 통과한다**(휴장일). `EarlyClose`(:132)는 길이가 6h30m 미만인지만 본다 — 개장 지연(길이 같고 시각이 밀림)은 못 가른다.
- `CalendarSnapshot.ValidityAt`(`:138-149`): B1 시계 역행 · B2 6시간 신선도 · **B3 `REFRESH_TOO_LATE` — 정규장 중에 장 개시 뒤 읽은 달력을 거부**.

**openapi 의 KR 정규장 의미 (`docs/migration/openapi.latest.json` 스키마 `RegularMarketSession`·`IntegratedHour`)**: 달력의 KR 정규장은
「**KRX·NXT 정규장의 합집합**(가장 이른 시작 ~ 가장 늦은 종료). 종가단일가 구간을 포함」이다. 같은 스키마의 `singlePriceAuctionStartTime` 은
「정규장 내 종가단일가 구간 시작 (**KRX 기준**). KRX 휴장이면 null」이다. 즉 합집합 창은 **KRX 자신의 조기 폐장을 숨길 수 있다**(NXT 가 정상
운영하면 합집합은 정상 모양). 그리고 현재 Go 형 `official.MarketCalendarSession`(`typed_calendar_reads.go:11-14`)은 `startTime`·`endTime` 둘만
읽는다 — KRX 기준 필드가 없다. openapi 예시에 조기 폐장일 표본은 없다(예시 요약: 영업일 · 휴장일 · 부분 휴장 — `[미측정]`).

**판정 규칙 (MARKET 은 다섯 조건이 모두 참일 때만, 하나라도 거짓·불명이면 LIMIT).**

1. 시장이 KR 이다 — `clock.ParseMarket` 으로 정규화. 실패 → LIMIT. US 는 Q2 로 언제나 LIMIT 이므로 US 달력은 읽지 않는다.
2. 엔진 메모리에 KR 달력 응답이 있고, 그 응답이 `AdaptOfficialCalendar` 검증을 통과한다(위 B1~B8 — 오류면 LIMIT).
3. 응답의 신선도: 읽은 시각이 `now` 보다 미래가 아니고 읽은 지 6시간 미만(`ValidityAt` B1·B2 와 같은 값·같은 의미). **`ValidityAt` 을 그대로
   쓰지 않는다** — B3 `REFRESH_TOO_LATE` 는 진입 활성화의 규칙이고, 청산에 쓰면 6시간 상한과 겹쳐 KRX 정규장 끝 구간이 매일 LIMIT 으로 강등된다
   (장 개시 전에 읽은 달력은 늦어도 장 개시 6시간 뒤 낡는데 정규장은 6시간 30분이다). 강등은 안전 약화가 아니라 이득 손실이지만, 규칙을 가져오는
   이유가 청산에 없다.
4. 응답의 오늘 날짜가 `now` 의 KST 날짜와 같고, 그날 KR 정규장이 **정상 영업일 모양**이다 — 시작 09:00 · 끝 15:30 KST(합집합) **그리고**
   KRX 기준 종가단일가 시작 15:20 KST. 이 셋 중 하나라도 다르거나 없으면(휴장 `integrated` null · 조기 폐장 · 개장 지연 · KRX 휴장) **그날 종일 LIMIT**.
   KRX 기준 필드를 읽으려면 `official.MarketCalendarSession` 에 `singlePriceAuctionStartTime` 을 더해야 한다(타입 확장 — 2.6.4).
5. 고정 시각: `now` 가 `[09:00, 15:20)` KST 안이고 주말이 아니다. 시작과 주말은 `clock.MarketKR.RegularSession(now)`(`Open` · `Weekend`)를 쓰고, 끝 15:20 은
   종가단일가 시작 상수다 — `Session.Contains` 의 끝(15:30)은 종가단일가를 포함하므로 그대로 쓰지 않는다(위 표 종가단일가 행).

「고정 시각」의 이 해석(정상 영업일이 아니면 종일 LIMIT, 정상일이면 고정 창)은 이 문서의 유도다 — 사용자 문언은 「브로커 달력 + 고정 시각」이다.
대안(달력 창과 고정 창의 교집합)은 조기 폐장일 09:00~13:00 에 MARKET 이득을 더 얻지만, 합집합 의미 때문에 KRX 조기 폐장을 NXT 창이 가리면
교집합이 13:00~15:30 을 다시 연다. 이 해석은 MARKET 칸을 대안보다 좁히기만 하므로 불변식 4 에 대해 보수 방향이다. Manager 확인 항목으로 남긴다
(`review.md` 0.10 P2-2).

**즉시성 제약 (불변식 4 — 이 절의 가장 무거운 요구).**

- **제출 경로에서 브로커를 부르지 않는다.** 판정은 메모리의 마지막 달력 응답만 읽는다. 응답이 없거나 낡았으면 그 자리에서 LIMIT 이다 —
  달력 GET 을 기다리지 않는다. 오늘의 LIMIT 보호 청산이 겪지 않는 대기를 더하면 그것이 지연이다.
- **판정 실패는 거부가 아니라 LIMIT 이다.** 판정 함수는 오류를 반환해 `sellIntent`·`submit` 의 거부 갈래(`submit` B5 `:1533` → 해제)로 보내지
  않는다. 달력 문제로 보호 청산이 **안 나가는** 갈래가 생기면 오늘보다 약해진다.
- **달력 갱신은 청산 관측 주기 밖에서 한다.** 관측 루프(`Run` B3 `o.clk.Sleep(ctx, o.Interval())` — 0.5 AST)가 달력 GET 에 묶이면 판정이 늦어진다.
  갱신 주기는 6시간 상한보다 짧아야 하고 KR 만 읽는다(§0.4 — 하루 수 회의 GET). 구조(별도 루프·기존 엔진 루프 목록 등록 등)는 2.6.4 가 정한다.
- 엔진이 꺼져 있으면 이 갱신·판정이 없다(불변식 3 — CLI·ops·MCP 경로는 달력을 읽지 않는다).

**잔존 `[미측정]`**: ① 브로커 달력이 KRX 조기 폐장·개장 지연을 위 필드로 실제로 보이는지(openapi 예시 없음). 거짓으로 정상 모양을 보이면 그날
조기 폐장 뒤 MARKET 이 나갈 수 있다 — 그 시각 MARKET 이 거부되면 오늘보다 약하다. ② 정규장 칸 안의 종가단일가(15:20~15:30)와 VI 단일가는 §5.1
(14:59 접속매매) 밖이다 — 종가단일가는 LIMIT 으로 뺐고(위 표), VI 단일가는 예측 입력이 없어 D8 잔존 행이다.

### §5.2 처분 (7판 — 사용자 결정 2026-10-28 Q4): not-applicable

§5.2(정규장 밖 KR MARKET 매도 1주 실측)는 하지 않는다. 사유: D7 이 모름을 LIMIT 으로 고정했으므로 실측의 이득은 「나중에 어떤 칸을 MARKET 으로
옮길 근거」뿐이다. 그 칸은 오늘 LIMIT 이고 바뀌지 않는다. 반대로 실측은 실주문이 다음 세션에 체결될 위험이 있다(tasks §5 표 「세션 밖인데 200」 행).
이득 0 · 위험 1 이다. 그래서 tasks 5.2 는 not-applicable 로, 5.3 은 5.1 분(M50)만으로 닫는다. LIMIT 칸을 MARKET 으로 옮기려면 새 실측 · spec 개정 ·
리뷰를 거친다(이 change 밖). `issues.md` I-R2(`order-type-not-allowed` 422 분류)는 「§5.2 결과 뒤 판정」이었으므로 **판정할 입력이 없는 채 무소유로 남는다**
— 정규장 밖 MARKET 을 내지 않으므로 그 코드를 받을 생산 경로가 없다.

## D8 — 시장가가 지정가보다 낫지 않거나 나빠지는 경계 조건 (재리뷰 P2-3)

| 경계 조건 | 처분 | 사유 |
| --- | --- | --- |
| 하한가 도달·매수 호가 공백 | **범위 밖 — 잔존 위험** | MARKET 매도도 체결되지 않는다. 하한가 지정가 대비 이득 0. 체결 보장 주장 철회(proposal §0.3)가 이 행의 기록 |
| VI(변동성 완화장치) 단일가 구간 | **범위 밖 — `[미측정]`** | 단일가 매매 중 MARKET 접수·체결 거동 실측 없음. §5 측정 항목도 아니다. `issues.md` I-R4 |
| KRX 종가단일가(15:20~15:30 — 7판) | **다룬다 — LIMIT(오늘과 같음)** | openapi `RegularMarketSession` 이 정규장을 「종가단일가 구간을 포함」으로 정의한다. §5.1 은 14:59 접속매매 한 건이라 이 구간을 재지 않았다. D7 Q1 원칙(접수 미측정 → LIMIT)을 적용해 MARKET 칸에서 뺐다 — Manager 확인(`review.md` 0.10 P2-4) |
| 부분체결 잔량 (1차 리뷰 A14 의 행방) | **다룬다 — tasks 3.6** → 갭 I-R9 는 **새 change a167 로 이연**(사용자 결정 2026-10-22: 경보만이 아니라 자동 재가격·재주문) | MARKET 잔량이 남는지·`filldetect` 가 그 잔량을 어떻게 닫는지 확인하고, 갭이면 issues 에 기록. 3.6 결과: 잔량이 비종결로 남는 동안 시간 상한·경보 없음(`analysis/partial-fill-3.6/`). 처분은 아래 D10 |
| 1억 이상 주문 | **범위 밖 — 유형 무관 기존 위험** | `orders_write.go` `ConfirmHighValueOrder: false` → `400 confirm-high-value-required`(재리뷰 인용)는 LIMIT 에도 같다. MARKET 의 금액 평가 기준 `[미측정]` — `issues.md` I-R4 |
| 갭다운·거래정지 | **범위 밖 — 잔존 위험** | 접수 ≠ 체결. 거래정지 중 제출 응답 `[미측정]` |

## D9 — a100 D8 이중 매도 계약과의 상호작용 (재리뷰 P2-4 — a100 으로 명시 이연)

a100 design D8 은 계약 1 을 「a087 Phase 2 가 착지하면 창은 줄어든다」로 개정해 a087 을 전제로 삼았고, 계약 2 는 「상주
조건주문 취소 확인 실패는 인프로세스 매도를 막지 않는다」이다. 인프로세스 보호 매도가 MARKET 이 되면 즉시 체결되고,
취소가 확인되지 않은 상주 SINGLE+MARKET 이 같이 발동하는 창(M13)의 의미가 바뀐다. 브로커 매도가능수량 예약(M29)이 둘째
주문을 막는지는 `[미측정]`.

**처분: a100 으로 이연한다.** a100 은 HOLD(`f3a061e5`)라 착지 순서가 열려 있고, 이중 매도 계약의 소유자는 a100 이다.
a087 은 이 상호작용을 바꾸는 코드를 갖지 않는다(조건주문 경로 무변화 — tasks 4.3). 다만 a087 이 먼저 착지하면 a100 의
계약 1 전제가 참이 되는 순간이 a100 착지보다 앞서므로, **a100 착수 시 이 절을 재대조**해야 한다. a100 문서 쪽 반영은
이 개정의 디렉터리 범위 밖이다 — `issues.md` I-R5 에 Manager 반영 요청으로 남긴다.

## D10 — 비종결 잔량의 자동 재가격·재주문은 새 change a167 이 소유한다 (사용자 결정 2026-10-22, I-R9)

Manager 질문 「남은 잔량 경보. 권장: 가격을 자동으로 다시 정해 재주문」에 사용자가 그 권장대로 답했다. 고른 것은 **경보만이 아니라
자동 재가격·재주문**이다. 이것은 청산 경로에 새 주문 행동(취소 → 재제출 또는 정정)을 더하는 High-risk 변경이다. a087 의 범위
(보호 청산 1차 제출의 유형)에 없다. 그래서 a087 은 이것을 구현하지 않는다. 소유자는 새 change
`a167-a-protective-remainder-is-re-priced`(`STORY-TOS-a167`, proposal 수준 정의만)다.

a087 이 남기는 것:

- 갭의 기록 — `issues.md` I-R9 와 `analysis/partial-fill-3.6/`(AST 근거, HEAD `0686c0f0a`).
- 이연 사실 — tasks 「후속 change」 표의 a167 행.
- 상호작용 하나 — a167 의 재제출 유형은 이 문서 D7 의 구간 표를 따라야 한다. LIMIT 구간이면 재제출도 LIMIT 이다. 그래서 a167 은
  a087 Phase 2 의 구간 입력(tasks 2.6) 뒤에 착지한다. a167 proposal 「선행」에 적었다.

## 건드리지 않는 것

- **진입 주문 유형.** riskcalc 규칙이 실제로 적용되는 곳
- **`flatten`.** 이미 거래소 하한가를 1순위로 쓴다(`liquidate.go` 「The exchange's own floor」). 무변경. 하한가를 보호 청산에
  쓰는 원안은 D6(미채택 기록)
- **`verifylive`.** "체결되면 안 되는 지정가"가 그 도구의 안전 성질이다
- **조건주문.** 2c 범위
- **호가 그리드 정본화.** **무소유**(옛 표기 「a088」은 존재한 적 없다 — 3차 재리뷰 A-P3-1, proposal Non-goals)

## 검증

- `sellIntent` 분기: 보호 2종 → market·익절 → limit·시장가에 가격 없음·**보호가 가격
  없음으로 거부되지 않음**(MARKET 구간) · **D7 LIMIT 구간·구간 판정 불가에서 보호 2종 → 오늘과 같은 limit**(7판: US 정규장 포함 ·
  휴장·조기 폐장·개장 지연일 종일 · 달력 오류·결측·낡음 — `PlaceIntent` 전체가 오늘 출력과 같음) · **판정 실패가 거부가 아니라 LIMIT** ·
  **제출 경로의 브로커 호출 0**(D7 「구간 입력」 즉시성 제약)
- `checkOrderShape` 표: `sell+market` 통과 / `buy+market` 거부 / 가격 실린 시장가 거부 /
  `sell+market` 의 통화·수량 검사 생존 / fractional·limit 기존 분기 무변화
- `placeIntentSupported` 표(③, 단독 시험, **엔진 인스턴스 옵션 켬**): `sell+market`·가격 없음 지원 / `buy+market` 비지원 /
  가격 실린 시장가 비지원 / `sell+market` + 시장별 통화 위반 비지원 / fractional·limit 무변화 / `PreviewPlace` `LiveReady` 일치
- **upstream 동일성 (기본 인스턴스)**: 옵션 없는 `NewService` 가 비분수 market 매도를 `ErrPlaceUnsupported` 로 계속 거부,
  `PreviewPlace` 경고 문구 바이트 동일, ops `place_order`(MCP 노출 표면) 가 market 매도를 계속 거부 — 사람·에이전트 경로 무변화
- 엔진 경로 종단 시험: 보호 청산 MARKET 이 ①②③ 을 **모두** 지나 전송 계층에 닿는다(P0-1 의 회귀 시험)
- 진입 시장가 거부가 **살아 있음**을 별도 테스트로 고정(②·③ 각각)
- `isProtective` 술어 불변 + 세 소비자(D1)
- 원장: `price` 빈 값, `order_type` market, 관측가·기준선이 제출가로 새지 않음, 청산 슬리피지 계측(D6)
- 전송 표류 가드: KR market sell 골든 행(`decision_test.go` — 현재 US 만)
- `flatten`·`verifylive`·조건주문 무변화
- 전체 `go test ./... -race` 회귀 0, upstream 650 green 유지
- **실계좌 1회**(사용자 승인, `tools/a087-market-sell-probe/`): KR MARKET 매도 접수(5.1 — 2026-10-19 M50 수용) — 세션 경계(5.2)는
  not-applicable(7판 Q4, D7 「§5.2 처분」)
