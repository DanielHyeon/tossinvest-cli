# Function Logic Map: `Runner.stepOrderCancel`

- Source: `internal/verifylive/steps.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 함수 — 바뀐 것은 **기록된 문장 한 개**다.

```
- sr.observe("order.place.ok", "true", EndpointPlaceOrder+" accepted a minimum-quantity KR limit order")
+ sr.observe("order.place.ok", "true",
+     EndpointPlaceOrder+" accepted a minimum-quantity "+spec.Market+" limit order")
```

이 편집은 이 change가 한 것이 아니다. 이 브랜치 구간의 **앞선 커밋** `f62457c`
(`fix(verify): 관측 detail을 시장에 연동 + 발굴 임계 확정 [apply-us-measurement-fixes]`)의
것이고, `add-candidate-discovery`의 base(`583772c`)가 그 커밋보다 앞서기 때문에
이 change의 diff에 잡혀 evidence가 요구된다. 나머지 두 change의 base는 그 뒤라 요구되지 않는다.

**요청은 바뀌지 않았다.** `farBuySpec` → `placeOrder` → `readOrder` → `cancelOrder` →
`readOrder`의 호출 순서·인자·조건이 전부 그대로이고, 파일 diff는 `+11 -2`로 그 2줄이
위 문장 하나다. 바뀐 것은 **기록**이며, 기록이 바뀐 이유는 그것이 틀린 문장이었기
때문이다: US 실행이 "브로커가 KR 최소수량 지정가 주문을 받았다"고 적고 있었다.

## 이 단계는 실주문을 낸다 — 노출과 정리 의무

- `placeOrder`는 `checkOrderCap`(노출 상한) → `gate`(승인된 배치의 `Plan.Authorises`와
  사람 확인) → `broker.PlaceOrder` → `sr.created("order", …)` 순서다.
  즉 **실제 매수 지정가 주문 1건**이 계좌에 얹히고, 그 사실이 기록(Artifact)에 남는다.
- 이 단계의 **취소가 곧 측정**이므로 꼬리에 `cancelLiveOrders`가 없다.
  `stepOrderAmend`와 다른 점이고, 실패했을 때의 결과가 여기서 갈린다.
- **실패가 계좌에 남기는 것**:
  - B1(`farBuySpec` 실패) — 남는 것 없음. 주문을 만들기 전이다.
  - B2(`placeOrder` 실패) — 정상 경로에서는 남는 것 없음. 단, 브로커가 받아들이고 응답이
    유실된 IN_DOUBT 경우에는 계좌에 주문이 있을 수 있고 그때는 `sr.created`가 실행되지
    않았으므로 **기록에 없는 주문**이 된다.
  - B4(`cancelOrder` 실패, measurements.md M16) — **살아 있는 매수 지정가 주문 1건이
    계좌에 남는다.** 기록에는 남아 있으므로 (a) 노출 상한이 그것을 세어 이후 모든 mutating
    단계를 요청 전에 거부하고, (b) 다음 실행의 정리 프롤로그(`StepCleanup`,
    `cleanupTargets` — 주문은 **항상** 대상이다)가 승인 목록에 취소 한 줄로 올린다.
    `cleanup.go`가 존재하는 이유가 정확히 이 조합(취소 실패 + 종결 판정은 resume이 건너뜀 +
    상한은 기록을 읽음)이었다.

이 change(문장 수정)는 위 어느 것도 건드리지 않았다 — 상한·승인·정리 경로 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `spec` | `farBuySpec`이 만든 시장에서 먼 매수 지정가, 최소 수량 | `pricing.go` + `Options.Market` | 실패 시 즉시 반환 |
| `spec.Market` | `MarketKR` \| `MarketUS` | `Options.Market`/`MarketOf` | 기록 문장에 그대로 들어간다 |
| 승인 배치 | `Plan.Authorises(step, kind, symbol, side, quantity)` | `plan.go` | 목록에 없으면 요청 자체가 없다 |
| 노출 상한 | `checkOrderCap` | `mutate.go` + 기록 | 초과면 요청 전 거부 |

불변식: 이 단계가 만드는 라이브 객체는 **한 번에 하나**다(두 개를 동시에 두는 것은
`stepIdempotencyTTLEdge`뿐이고 그쪽은 opt-in이다).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L395) | `farBuySpec` 실패 | 없음 | err | `runner_test.go`의 거부·중단 케이스 |
| B2 (if, L399) | `placeOrder` 실패(상한·승인·전송·브로커 거부) | 정상 경로에서는 없음 | err | `plan_test.go`(미승인 거부), `confirm_test.go` |
| B3 (if, L405) | 배치 후 `readOrder` 성공 | 관측 2건 기록 | — | `us_market_test.go`의 기록 단언 |
| B4 (if, L410) | `cancelOrder` 실패 | **주문이 계좌에 남는다** | err | `cleanup_test.go`(M16: 실패한 취소가 남긴 것을 다음 실행이 정리) |
| B5 (if, L415) | 취소 후 `readOrder` 성공 | 관측 3건(상태·`canceledAt` 유무·체결수량) | — | `us_market_test.go` |

B3·B5는 **읽기 실패를 삼킨다**(`err == nil`일 때만 기록). 의도된 것이다 — 조회가 실패해도
취소는 이미 성공했고, 관측이 없다는 사실 자체가 기록의 부재로 남는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.farBuySpec` | 체결되지 않을 가격·최소 수량 | 실패 즉시 반환 | ast.json calls |
| `r.placeOrder` | **라이브 주문 생성** | 상한 → 승인 게이트 → 전송, 재시도 없음(`replayPlaceOrder`는 별도 단계) | ast.json calls |
| `r.readOrder` ×2 | 배치 후·취소 후 상태 관측 | 실패는 기록 생략으로 흡수 | ast.json calls |
| `r.cancelOrder` | **라이브 취소** — 이 단계의 측정 대상 | `already-processing` 재시도는 `retry` 경로가 소유 | ast.json calls |
| `sr.observe` ×6 | 기록 | — | ast.json calls |
| `orDash` / `strconv.FormatBool` / `strings.TrimSpace` | 기록 문자열 | 순수 | ast.json calls |

## State mutations and fallbacks

- **계좌 변경 있음**: 매수 지정가 1건 생성 후 취소. 이것이 이 단계의 목적이다.
- 기록 변경: `sr.created("order", …)`가 Artifact를 남기고, 그것이 노출 상한과 다음 실행
  정리 프롤로그의 입력이 된다.
- fallback 없음. 취소 실패는 삼켜지지 않고 오류로 올라간다.

## Safety conclusion

- Safe edit boundary: 기록 문장 1건(`spec.Market` 삽입). 요청·상한·승인·정리 무변경.
- High-risk impact: **yes** — 라이브 주문 코드다. 사람이 승인한 배치 안에서만 실행되지만,
  실제 계좌에 주문을 얹고 취소한다.
  이 편집이 안전한 이유: 바뀐 값은 `sr.observe`의 세 번째 인자(사람이 읽는 detail)뿐이고
  브로커로 나가는 `intent`에는 닿지 않는다. 그리고 방향이 보수적이다 — 틀린 문장을
  없앤다. 기록은 이후 change(2c의 귀속 규칙)가 읽는 증거이고, **거기 있는 거짓 문장은
  없는 문장보다 나쁘다.**
