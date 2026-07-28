# Function Logic Map: `TestOrderByIDIntegration`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `base` — base 쪽 hunk에만 걸린다)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 테스트 — **본문 무변경**이다. 이 change는 이 함수 **뒤**(base L267 이후)에 원문 보존
읽기의 테스트 묶음을 삽입했고(`@@ -265,3 +417,181 @@`), 그 hunk가 함수 끝과 인접해
evidence가 요구됐다. base(`137cc8d`) L228-267과 HEAD L380-419는 **바이트 동일**
(함수 구간 sha256 `cb92bf2f3f66ebe8…` 일치, 본 세션 확인). 152칸 이동은 앞쪽 삽입 때문이다.

이 테스트가 지키는 것: `OrderByID`가 `/api/v1/orders/{id}`를 부르고, 계좌 헤더를 실으며,
응답을 `domain.Order`로 사상한다는 것. 이 change는 그 계약을 고치지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| httptest 서버 | `/oauth2/token`, `/api/v1/orders/abc123` | 이 테스트 | 그 밖은 404 |
| 클라이언트 | `WithAccountSeq(5)` | `New` + Option | — |
| 기대 | ID/Symbol/Price + `X-Tossinvest-Account: 5` | 이 테스트 | `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch, base L231) | 요청 경로 분기 | 없음 | — | 자체 실행 |
| B2 (case, base L232) | `/oauth2/token` | 토큰 응답 | — | 자체 실행 |
| B3 (case, base L234) | `/api/v1/orders/abc123` | 헤더 기록 + 주문 응답 | — | 자체 실행 |
| B4 (case, base L237) | default | 없음 | `http.NotFound` | 자체 실행 |
| B5 (if, base L252) | `OrderByID` 오류 | 없음 | `t.Fatal` | 자체 실행 |
| B6 (if, base L255) | ID 불일치 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 (if, base L258) | Symbol 불일치 | 없음 | `t.Fatalf` | 자체 실행 |
| B8 (if, base L261) | Price 불일치 | 없음 | `t.Fatalf` | 자체 실행 |
| B9 (if, base L264) | 계좌 헤더 불일치 | 없음 | `t.Fatalf` | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httptest.NewServer` | 브로커 대역 | `defer srv.Close()` | ast.json calls/defers |
| `New` + Option 3종 | 테스트 클라이언트 | — | ast.json calls |
| `c.OrderByID` | 측정 대상 | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: **본문 0줄 변경**. 인접 삽입만.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉) — 다만 재는 대상이 High-risk이고,
  이 테스트가 계좌 헤더가 실린다는 것을 잡는 유일한 지점 중 하나다. 무변경임을 증명하는
  것이 이 evidence의 요점이다.
