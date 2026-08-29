# Function Logic Map: `TestBothOrderReadsSendTheGroupTheyWereGiven`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L337-373). `status`는 다른 다섯 파라미터와 다르다 — 모양이 다른 두
답 중 하나를 고르고, 그중 하나(`OPEN`)는 이 엔드포인트에서 **절단될 수 없는 유일한 호출**이다.
그래서 "호출자가 준 그룹이 그대로 와이어에 실린다"는 자체 테스트값이 있는 주장이고,
두 읽기 어느 쪽이든 기본값을 넣거나 정규화하거나 떨어뜨리면 여기서 깨진다.

**additive 주장을 실측으로 만드는 테스트이기도 하다**: 같은 그룹으로 `Orders`와 `OrdersRaw`를
연달아 부르고, 서버가 본 `status` 값 두 개가 모두 같은지, 그리고 요청이 **정확히 2건**인지를
단언한다. 두 읽기가 같은 와이어 질문을 하고 계좌 해석을 공유한다는 뜻이다 —
클라이언트는 `ordersRawClient`로 하나만 만들고 `WithAccountSeq(3)`이라 `/api/v1/accounts`
왕복이 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `group` | `"OPEN"`, `"CLOSED"` | 이 테스트의 표 | 각 subtest |
| `seen` | 서버가 본 `status` 값들 | 핸들러 | 불일치 시 `t.Errorf` |
| 클라이언트 | `ordersRawClient(t, srv)` 하나 | 같은 파일 헬퍼 | — |

불변식: `len(seen) == 2` — 두 읽기가 **각각 한 번씩** 요청한다. 3이면 어딘가에서
계좌 해석 왕복이 늘었다는 뜻이고, 1이면 한쪽이 요청을 보내지 않았다는 뜻이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range, L338) | `OPEN`/`CLOSED` 순회 | 없음 | — | 자체 실행 |
| B2 (switch, L342) | 요청 경로 분기 | 없음 | — | 자체 실행 |
| B3 (case, L343) | `/oauth2/token` | 토큰 응답 | — | 자체 실행 |
| B4 (case, L345) | `/api/v1/orders` | `seen`에 `status` 기록 + 빈 페이지 | — | 자체 실행 |
| B5 (case, L348) | default | 없음 | `http.NotFound` | 자체 실행 |
| B6 (if, L355) | `Orders` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B7 (if, L358) | `OrdersRaw` 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B8 (range, L361) | `seen` 순회 | 없음 | — | 자체 실행 |
| B9 (if, L362) | 실린 그룹 ≠ 준 그룹 | 없음 | `t.Errorf` — 그룹이 답이 전량인지 이력 한 페이지인지를 정한다 | 자체 실행 |
| B10 (if, L368) | 요청 수 ≠ 2 | 없음 | `t.Fatalf` | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httptest.NewServer` | 브로커 대역, `status` 기록 | `defer srv.Close()` | ast.json calls/defers |
| `ordersRawClient` | 클라이언트 1개(계좌 seq 고정) | — | ast.json calls |
| `c.Orders` / `c.OrdersRaw` | 두 읽기 | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **yes** — 계좌 게이트웨이의 **와이어 요청이 동일함**을 재는 회귀 가드다.
  이 change의 "두 읽기가 같은 요청을 한다"는 additive 주장이 서는 실측 지점이고,
  요청 수 단언(B10)이 "두 번째 클라이언트도, 두 번째 계좌 해석도 생기지 않는다"를 잡는다.
