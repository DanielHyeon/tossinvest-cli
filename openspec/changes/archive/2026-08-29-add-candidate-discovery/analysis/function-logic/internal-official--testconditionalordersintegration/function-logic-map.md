# Function Logic Map: `TestConditionalOrdersIntegration`

- Source: `internal/official/conditional_reads_test.go`
- AST evidence: `ast.json` (revision `base` — base 쪽 hunk에만 걸린다)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 테스트 — **본문 무변경**이다. 이 change는 이 함수 **뒤**(base L81 이후)에 원문 보존
읽기의 테스트를 삽입했고(`@@ -79,3 +79,70 @@`, `+67 -0`), 그 hunk가 함수 끝과 인접해
evidence가 요구됐다. base(`137cc8d`) L48-81과 HEAD L48-81은 **바이트 동일**
(함수 구간 sha256 `24f8e79734d0bc40…` 일치, 본 세션 확인). 줄 번호도 그대로다.

이 테스트가 지키는 것: `ConditionalOrders`가 `status=OPEN`을 실어 보내고, 응답을
`domain.ConditionalOrderList`로 사상하며, SINGLE의 `second`가 nil로 남는다는 것.
이 change는 그 계약을 **고치지 않았고**, 옆에 원문 보존 읽기를 세웠을 뿐이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| httptest 서버 | `/oauth2/token`, `/api/v1/conditional-orders` 두 경로만 | 이 테스트 | 그 밖의 경로는 `t.Errorf` |
| 클라이언트 | `WithAccountSeq(1)`로 지연 해석 생략 | `New` + Option | — |
| 응답 픽스처 | SINGLE 1건, `targetProfitRate`/`triggeredOrderId` null | 이 테스트 | 단언 실패 시 `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch, base L50) | 요청 경로 분기 | 없음 | — | 자체 실행 |
| B2 (case, base L51) | `/oauth2/token` | 토큰 응답 | — | 자체 실행 |
| B3 (case, base L53) | `/api/v1/conditional-orders` | 조건주문 페이지 응답 | — | 자체 실행 |
| B4 (if, base L54) | `status != "OPEN"` | 없음 | `t.Errorf` — 그룹이 와이어에 안 실렸다 | 자체 실행 |
| B5 (case, base L58) | default | 없음 | `t.Errorf("unexpected path")` | 자체 실행 |
| B6 (if, base L72) | `ConditionalOrders` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 (if, base L75) | 건수·id·`second`·`HasNext` 중 하나라도 어긋남 | 없음 | `t.Fatalf` | 자체 실행 |
| B8 (if, base L78) | 첫 다리 trigger가 70000이 아님 | 없음 | `t.Fatalf` | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httptest.NewServer` | 브로커 대역 | `defer srv.Close()` | ast.json calls/defers |
| `New` + `WithBaseURL`/`WithHTTPClient`/`WithAccountSeq` | 테스트 클라이언트 | — | ast.json calls |
| `c.ConditionalOrders` | 측정 대상 | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: **본문 0줄 변경**. 인접 삽입만.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉) — 다만 이 테스트가 재는 대상은
  High-risk(계좌 게이트웨이의 조건주문 조회)이고, 그것이 무변경임을 증명하는 것이
  이 evidence의 요점이다.
