# Function Logic Map: `Runner.stepOrderAmend`

- Source: `internal/verifylive/steps.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 함수 — 바뀐 것은 **기록된 문장 한 개와 그 문장을 만드는 3줄**이다.

```
- sr.observe("order.amend.ok", "true", EndpointModifyOrder+" accepted a KR price+quantity amend")
+ amended := "price+quantity"
+ if !SameMarket(spec.Market, MarketKR) {
+     amended = "price-only"
+ }
+ sr.observe("order.amend.ok", "true",
+     EndpointModifyOrder+" accepted a "+spec.Market+" "+amended+" amend")
```

이 편집은 이 change가 한 것이 아니다. 이 브랜치 구간의 **앞선 커밋** `f62457c`
(`fix(verify): 관측 detail을 시장에 연동 + 발굴 임계 확정 [apply-us-measurement-fixes]`)의
것이고, `add-candidate-discovery`의 base(`583772c`)가 그 커밋보다 앞서기 때문에
이 change의 diff에 잡혀 evidence가 요구된다. 나머지 두 change의 base는 그 뒤라 요구되지 않는다.

**요청은 이미 시장을 따르고 있었다.** `amendOrder`가 US에서 수량을 싣지 않는다는 것은
이 편집 이전부터 참이고 `TestTheUSAmendSendsNoQuantity`가 그것을 잰다
(openapi: `OrderModifyRequest.quantity` — "US 주식: 전달 불가. 제공 시
`400 us-modify-quantity-not-supported`"). 틀렸던 것은 **그 옆에 적히는 문장**이었다:
US 실행이 "브로커가 KR price+quantity 정정을 받았다"고 기록했고, 그것은 브로커가
400으로 거절했을 요청을 서술한 문장이다.

## 이 단계는 실주문을 낸다 — 노출과 정리 의무

- `placeOrder`가 라이브 매수 지정가 1건을 만들고(`checkOrderCap` → `gate` → `PlaceOrder` →
  `sr.created`), `amendOrder`가 그것을 정정한다. **공식 API의 정정은 새 주문번호를 만든다** —
  그래서 `order.amend.issues_new_id` 관측이 있고, 원본 id와 현재 id 둘을 모두 읽는다.
- 꼬리에 `cancelLiveOrders(…, "the amend check is complete")`가 있다 —
  `stepOrderCancel`과 달리 여기서는 취소가 측정이 아니라 **정리**다.
- **실패가 계좌에 남기는 것**:
  - B1(`farBuySpec`) — 없음.
  - B2(`placeOrder`) — 정상 경로에서는 없음. 응답 유실 IN_DOUBT면 기록에 없는 주문이 생길 수 있다.
  - B3(`amendOrder` 실패) — **주문이 남는다.** 정정이 새 번호를 만드는 API라 이때 남는 것이
    원본인지 자식인지가 확정적이지 않고, 그것이 `order.amend.original_status`/
    `current_status` 관측(B5-B8)이 존재하는 이유다. 꼬리의 `cancelLiveOrders`에 도달하지
    못하므로 정리는 **다음 실행의 프롤로그**(`StepCleanup`)로 넘어가고, 그때까지 그 주문은
    노출 상한을 채워 이후 mutating 단계를 요청 전에 거부시킨다.
  - 꼬리 `cancelLiveOrders` 실패 — 같은 결과이며 오류로 보고된다.

이 change(문장 수정)는 위 어느 것도 건드리지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `spec` | 시장에서 먼 매수 지정가, 최소 수량 | `farBuySpec` | 실패 즉시 반환 |
| `newPrice` | `OneTickFurther(spec.Price, "buy", spec.Market)` — 더 멀어지는 방향 | `pricing.go` | 체결 위험을 늘리지 않는다 |
| `spec.Market` | `MarketKR` \| `MarketUS` | `Options.Market` | 요청(수량 포함 여부)과 **기록 문장** 둘 다 여기서 갈린다 |
| 승인 배치 | `Plan.Authorises` | `plan.go` | 목록에 없으면 요청 자체가 없다 |

불변식: 정정은 **더 멀어지는 방향**으로만 한다 — 측정을 위해 체결 확률을 올리지 않는다.
그리고 이제 요청과 기록이 **같은 조건**(`SameMarket(spec.Market, MarketKR)`)에서 갈린다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L426) | `farBuySpec` 실패 | 없음 | err | `runner_test.go` |
| B2 (if, L430) | `placeOrder` 실패 | 정상 경로에서는 없음 | err | `plan_test.go`, `confirm_test.go` |
| B3 (if, L436) | `amendOrder` 실패 | **주문이 계좌에 남는다**(꼬리 정리에 도달하지 못함) | err | `cleanup_test.go`(다음 실행이 정리) |
| B4 (if, L443) | `!SameMarket(spec.Market, MarketKR)` — **이 change가 넣은 분기** | 기록 문자열만 `price-only`로 | — | `TestTheRecordDoesNotCallAUSRequestAKROne` |
| B5 (if, L451) | `currentID != orderID`(정정이 새 번호를 냈다) | — | — | `record_test.go`, `us_market_test.go` |
| B6 (if, L452) | 원본 id 조회 성공 | `order.amend.original_status` 기록 | — | 동상 |
| B7 (else, L455) | 원본 id 조회 실패 | `original_status = "unreadable"` + 오류 요약 | — | 동상 |
| B8 (if, L459) | 현재 id 조회 성공 | `current_status`·`current_price` 기록 | — | 동상 |

B4가 유일한 신규 분기이고, 계좌에 나가는 요청이 아니라 **기록 문자열**을 고른다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.farBuySpec` | 체결되지 않을 스펙 | 실패 즉시 반환 | ast.json calls |
| `r.placeOrder` | **라이브 주문 생성** | 상한 → 승인 게이트 → 전송 | ast.json calls |
| `OneTickFurther` | 더 먼 가격 | 순수 | ast.json calls |
| `r.amendOrder` | **라이브 정정** — 시장별로 수량 포함 여부가 갈린다 | 새 주문번호 반환 가능 | ast.json calls |
| `SameMarket` (신규 호출) | 기록 문장의 시장 분기 | 순수 | ast.json calls |
| `r.readOrder` ×2 | 원본·현재 id 상태 | 실패는 `unreadable`로 기록 | ast.json calls |
| `r.cancelLiveOrders` | **정리** — 이 단계가 만든 것을 남기지 않는다 | 오류는 그대로 반환 | ast.json calls |

## State mutations and fallbacks

- **계좌 변경 있음**: 매수 지정가 1건 생성 → 정정(새 번호 가능) → 정리 취소.
- 기록 변경: `sr.created`의 Artifact와 관측 5~6건. 그중 하나가 이 change가 고친 문장이다.
- fallback: 원본 id 조회 실패만 `unreadable`로 흡수하고(B7), 나머지 실패는 오류로 올라간다.

## Safety conclusion

- Safe edit boundary: 기록 문장 1건 + 그것을 만드는 분기 3줄. 요청 조립(`amendOrder`),
  상한, 승인, 정리 경로 무변경.
- High-risk impact: **yes** — 라이브 주문·정정 코드이고, 정정이 새 주문번호를 만드는 API라
  실패 지점에 따라 계좌에 남는 객체가 달라진다.
  이 편집이 안전한 이유: 새 분기는 `sr.observe`의 detail 문자열만 고르고 브로커로 나가는
  요청에 닿지 않는다. 그리고 **요청과 기록이 같은 술어를 쓰게 만든 것**이 이 편집의 요점이다 —
  `amendOrder`는 이미 `SameMarket(…, MarketKR)`로 수량 포함 여부를 정하고 있었고,
  이제 문장도 같은 조건으로 갈린다. 두 정본이 어긋나 있던 상태가 없어졌다.
