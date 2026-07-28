# Function Logic Map: `TestOrdersFilterEmptyOmitsEveryParameterIncludingTheRequiredOne`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

base의 `TestOrdersFilterEmpty`(L187-221)를 **개명하고 단언 하나를 더한** 것이다
(HEAD L210-253). base의 단언 8개(B1-B8)는 그대로 남았고, B9이 새로 붙었다.

`git diff --unified=0`의 `@@ -183,8 +185,29 @@` hunk가 base 이름·주석 2줄을 새 이름·주석
23줄로 바꾼 부분이고, `@@ -218,6 +241,135 @@`이 함수 끝의 B9 삽입과 그 뒤의 신규 테스트다.

## 이 테스트가 고정하는 것과, 허가로 읽혀서는 안 되는 것

- **고정하는 것**: `OrdersFilter`의 zero value는 쿼리 파라미터를 하나도 만들지 않는다.
  클라이언트가 호출자가 대지 않은 값을 지어내지 않는다.
- **허가가 아닌 것**: 빈 필터가 쓸 수 있는 호출이라는 뜻이 아니다. openapi는 `status`를
  `required: true`로 표시하고, 이 저장소의 모든 호출자가 하나를 넘긴다 —
  `cmd/tossctl/soak.go`, `internal/verifylive/steps.go`(2곳),
  `internal/execgw/orders_source.go`, 콘솔의 주문 어댑터.
- **B9이 그 경계를 코드로 만든다**: 같은 입력을 `OrdersRaw`에 주면
  `ErrOrderStatusRequired`다. `Orders`가 행동을 유지하는 이유는 선행 호출자가 있어서고,
  새 읽기에는 그 변명이 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| httptest 서버 | `/oauth2/token`, `/api/v1/orders` | 이 테스트 | 그 밖은 404 |
| 클라이언트 | `WithAccountSeq(1)` | `New` + Option | — |
| 기대(기존) | 여섯 키 전부 부재, 결과 0건 | 이 테스트 | `t.Errorf`/`t.Fatalf` |
| 기대(신규) | `OrdersRaw(ctx, OrdersFilter{})`가 `ErrOrderStatusRequired` | 이 테스트 | `t.Fatalf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch, L212) | 요청 경로 분기 | 없음 | — | 자체 실행 |
| B2 (case, L213) | `/oauth2/token` | 토큰 응답 | — | 자체 실행 |
| B3 (case, L215) | `/api/v1/orders` | 빈 페이지 응답 | — | 자체 실행 |
| B4 (range, L217) | 여섯 키 순회 | 없음 | — | 자체 실행 |
| B5 (if, L218) | 키가 비어 있지 않음 | 없음 | `t.Errorf` | 자체 실행 |
| B6 (case, L223) | default | 없음 | `http.NotFound` | 자체 실행 |
| B7 (if, L238) | `Orders` 오류 | 없음 | `t.Fatal` | 자체 실행 |
| B8 (if, L241) | 결과가 0건이 아님 | 없음 | `t.Fatalf` | 자체 실행 |
| B9 (if, L249) | `OrdersRaw(빈 필터)`가 `ErrOrderStatusRequired`가 아님 | 없음 | `t.Fatalf` — 두 읽기의 차이가 의도된 것임을 고정 | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httptest.NewServer` | 브로커 대역 | `defer srv.Close()` | ast.json calls/defers |
| `c.Orders` | 기존 읽기의 행동 고정 | 오류 그대로 단언 | ast.json calls |
| `c.OrdersRaw` | 새 읽기의 거부 고정 | `errors.Is(err, ErrOrderStatusRequired)` | ast.json calls |

## Safety conclusion

- Safe edit boundary: 개명 + 근거 주석 + 단언 1건. 기존 단언 8개 무변경.
- High-risk impact: **yes** — 이것은 계좌 게이트웨이의 기존 읽기(`Orders`)가 **바뀌지 않았음**을
  고정하는 회귀 가드다. 이 change의 additive 주장이 서는 곳 중 하나가 여기다:
  `Orders`는 빈 필터를 그대로 보내고(선행 호출자 보호), 새 읽기만 거부한다.
  이름 변경이 위험을 더하지 않는 이유는 base의 단언이 하나도 삭제되지 않았기 때문이다.
