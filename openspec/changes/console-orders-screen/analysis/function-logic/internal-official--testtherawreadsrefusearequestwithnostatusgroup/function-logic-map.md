# Function Logic Map: `TestTheRawReadsRefuseARequestWithNoStatusGroup`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L267-328). 두 원문 읽기의 status 가드를 네 케이스
(`OrdersRaw` empty/blank, `ConditionalOrdersRaw` empty/blank)로 재고, 거부의 **성질** 셋을
따로 단언한다.

1. 타입이 있는 오류다 — `errors.Is(err, ErrOrderStatusRequired)`.
2. 오류가 파라미터 이름(`status`)과 어느 그룹이 살아 있는 질문에 답하는지(`OPEN`)를 말한다.
3. **요청이 나가기 전에** 거부한다 — 서버가 센 요청 수가 0이어야 한다.
   잊은 호출자가 rate-limit 슬롯을 써 가며 알아낼 일이 아니다.
4. 그리고 이것을 **장애로 오인하지 않는다** — `ShouldFallback(err)`가 false.
   이 클라이언트가 만들기를 거부한 요청을 웹 세션으로 보내는 것은 잘못된 질문을 다른
   곳에서 답하는 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `requests` 카운터 | 토큰 경로를 제외한 모든 요청을 센다 | 이 테스트의 핸들러 | 0이 아니면 `t.Errorf` |
| 케이스 표 | 4건(두 메서드 × empty/blank) | 이 테스트 | 각 subtest에서 단언 |
| 클라이언트 | `ordersRawClient`(`WithAccountSeq(3)`) | 같은 파일의 헬퍼 | — |

불변식: blank(`"   "`, `"  "`) 케이스가 있으므로 가드는 `TrimSpace` 기반이어야 한다 —
공백만인 문자열이 통과하면 `?status=+++`가 나간다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L270) | 요청 경로가 `/oauth2/token` | 토큰 응답 후 반환(카운트하지 않음) | — | 자체 실행 |
| B2 (range, L280) | 4개 케이스 순회 | 각 케이스를 `t.Run`으로 | — | 자체 실행 |
| B3 (if, L303) | `errors.Is(err, ErrOrderStatusRequired)` 실패 | 없음 | `t.Fatalf` | 자체 실행 |
| B4 (if, L306) | 오류에 `status`가 없음 | 없음 | `t.Errorf` | 자체 실행 |
| B5 (if, L309) | 오류에 `OPEN`이 없음 | 없음 | `t.Errorf` | 자체 실행 |
| B6 (if, L315) | `requests != 0` | 없음 | `t.Errorf` — 거부된 요청이 브로커에 닿았다 | 자체 실행 |
| B7 (if, L324) | `ShouldFallback(err)`가 true | 없음 | `t.Error` — 호출자 실수를 장애로 보고했다 | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

핸들러는 토큰 외의 어떤 요청이 와도 200을 돌려주도록 되어 있다 — **요청이 왔다는 사실
자체**를 실패로 잡기 위해서지, 응답 내용으로 잡는 것이 아니다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httptest.NewServer` | 요청 수 계수 | `defer srv.Close()` | ast.json calls/defers |
| `ordersRawClient` | 테스트 클라이언트 | — | ast.json calls |
| `c.OrdersRaw` / `c.ConditionalOrdersRaw` | 측정 대상 ×2 | 요청 전 거부 | ast.json calls |
| `errors.Is` / `strings.Contains` / `ShouldFallback` | 거부의 성질 확인 | — | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **yes** — 이 테스트가 지키는 성질(요청 전 거부, fallback 아님)은
  계좌 게이트웨이의 요청 발생 조건 그 자체다. 약해지면 "잊은 호출자"가 다시 조용히
  절단된 라이브 읽기를 보내고, 그 결과는 화면에서 잔여물이 사라지는 것이다.
  issues.md I-7에 기록된 대로, 이 테스트를 처음 변이 검증할 때 가드를 **삭제**하는 형태로
  넣었더니 미사용 import 때문에 빌드가 깨져 테스트가 실행조차 되지 않았다 —
  컴파일러가 잡은 것은 증명이 아니므로 조건을 항상 거짓이 되게 바꾸는 형태로 다시 넣었고,
  그때 이 테스트가 실제로 물었다.
