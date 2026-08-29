# Function Logic Map: `TestTheRawOrderReadDerivesTheMarketFromTheCurrency`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L554-574). `/orders` 응답에는 market도 name도 없다. 기존 읽기가
currency를 디코딩한 뒤 버리므로, 시장 열은 여기서 유도되거나 존재하지 않는다 —
그리고 시장이 없는 화면은 "이 중 어느 것이 지금 열려 있는 세션의 것인가"에 답하지 못한다.

세 번째 단언이 요점이다: `marketFromCurrency("JPY")`는 빈 문자열이어야 한다.
이 빌드가 모르는 통화는 **추측된 시장이 아니라 모르는 시장**이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 서버 | `ordersRawServer(t, false, "")` — KRW 1건, USD 1건 | 같은 파일 헬퍼 | `defer srv.Close()` |
| 호출 | `OrdersRaw(ctx, OrdersFilter{Status: "OPEN"})` | 이 테스트 | — |
| 기대 | KRW→KR, USD→US, JPY→`""` | 이 테스트 | `t.Errorf` |

`Currency`도 함께 단언한다 — 유도된 값만 남기고 원문을 버리면 다음 통화가 추가될 때
근거가 사라진다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L559) | 읽기 오류 | 없음 | `t.Fatal` | 자체 실행 |
| B2 (if, L562) | KRW 행의 currency/market 불일치 | 없음 | `t.Errorf` | 자체 실행 |
| B3 (if, L566) | USD 행의 currency/market 불일치 | 없음 | `t.Errorf` | 자체 실행 |
| B4 (if, L570) | `marketFromCurrency("JPY")`가 빈 문자열이 아님 | 없음 | `t.Errorf` — 모르는 통화에 시장을 추측하면 운영자가 그 열로 거른다 | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ordersRawServer` / `ordersRawClient` | 픽스처 | `defer srv.Close()` | ast.json calls/defers |
| `Client.OrdersRaw` | 측정 대상 | 오류 그대로 단언 | ast.json calls |
| `marketFromCurrency` | 미인식 통화 직접 확인 | 순수 | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉). 재는 대상은 화면 필터 축이고,
  이 테스트의 핵심 가치는 "모를 때 추측하지 않는다"를 코드로 고정하는 것이다.
