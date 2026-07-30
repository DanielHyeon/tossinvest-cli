# Function Logic Map: `marketFromCurrency`

- Source: `internal/official/orders_reads.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 순수 함수**다(HEAD L303-312). `GET /api/v1/orders` 응답에는 market 필드도 name 필드도
없다. `Client.Orders`는 currency를 디코딩한 뒤 버리므로, 시장 열은 여기서 유도되거나
존재하지 않는다 — 그리고 시장이 없는 화면은 "이 중 어느 것이 지금 열려 있는 세션의 것인가"에
답하지 못한다.

모르는 통화는 **빈 문자열**이다. 행에 틀린 시장이 붙는 것이 시장이 없는 것보다 나쁘다 —
운영자가 그 열로 거른다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `currency` | 임의 문자열; `TrimSpace`+`ToUpper` 후 비교 | `/orders` 응답의 `currency` | 미인식 → `""` |

불변식: 순수 함수 — 부작용 없음, 전역 없음, 오류 없음. 조건주문 쪽은 payload가 market을
직접 실어 보내므로 이 유도를 쓰지 않는다(`ConditionalOrdersRaw`는 `o.Market`을 그대로 옮긴다).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch, L304) | `ToUpper(TrimSpace(currency))` 분기 | 없음 | — | `TestTheRawOrderReadDerivesTheMarketFromTheCurrency` |
| B2 (case, L305) | `"KRW"` | 없음 | `"KR"` | 동상(KRW 주문 행) |
| B3 (case, L307) | `"USD"` | 없음 | `"US"` | 동상(USD 주문 행) |
| B4 (case, L309) | default | 없음 | `""` | 동상(`marketFromCurrency("JPY")` 직접 단언) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ToUpper` / `strings.TrimSpace` | 대소문자·공백 흔들림 흡수 | 순수 | ast.json calls |

네트워크·파일·계좌 접촉 0.

## State mutations and fallbacks

- 없음. 순수 함수.
- "추측하지 않는다"가 곧 fallback 정책이다 — 모르는 통화에 기본 시장을 넣지 않는다.

## Safety conclusion

- Safe edit boundary: 신규 leaf 함수 가산. 호출자는 `Client.OrdersRaw` 하나(그리고 테스트).
- High-risk impact: **no** — 계좌·주문·원장에 닿지 않는 순수 문자열 사상이다.
  다만 그 출력이 High-risk 화면의 **필터 축**이 되므로, 잘못된 값을 만들지 않는 것보다
  **모를 때 빈 값을 내는 것**이 이 함수의 실제 안전 성질이다. 그 성질은 B4가 소유하고
  테스트가 직접 단언한다.
