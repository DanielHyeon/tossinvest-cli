# Function Logic Map: `TestOrdersFilterEmpty`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `base` — 이 이름은 HEAD에 존재하지 않는다)
- Risk scan: `risk-pattern-report.md` (현재 파일 기준. base에는 매칭 0건, HEAD도 0건)

기존 테스트 — 이 change가 **개명**했다. base(`137cc8d`)의 `TestOrdersFilterEmpty`(L187-221)는
HEAD에서 `TestOrdersFilterEmptyOmitsEveryParameterIncludingTheRequiredOne`(L210-253)이 됐다.
그래서 이 target의 evidence는 **base revision**이며, HEAD의 후신에는 별도의 target 디렉터리가
있다(`internal-official--testordersfilteremptyomitseveryparameterincludingtherequiredone`).

## 왜 이름을 바꿨나

base의 이름과 주석("verifies that Orders() with zero filter omits all params")은 **부재만**
말했고, 어떤 구현이 그것을 **허가로 읽었다**. 콘솔의 라이브 읽기가 `?limit=100`에 status 없이
나갔고, 그것은 브로커가 거부할 수 있는 요청이거나 계좌 전 이력 한 페이지다 — 후자면
101번째 살아 있는 주문이 화면에서도 건수에서도 빠지고 "0건 이상"으로 렌더된다.

본문 자체의 단언은 base와 HEAD가 같다(B1-B8이 그대로 남아 있다). HEAD가 더한 것은
이름·설명과, `OrdersRaw`가 같은 입력을 **거부한다**는 단언 한 건(HEAD의 B9)이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| httptest 서버 | `/oauth2/token`, `/api/v1/orders` | 이 테스트 | 그 밖은 404 |
| 클라이언트 | `WithAccountSeq(1)` — 지연 해석 생략 | `New` + Option | — |
| 호출 | `Orders(ctx, OrdersFilter{})` | 이 테스트 | — |
| 기대 | `status/symbol/from/to/cursor/limit` 여섯 키가 전부 부재 | 이 테스트 | `t.Errorf` |

불변식: 이 테스트가 고정하는 것은 **하나**다 — `OrdersFilter`의 zero value가 쿼리
파라미터를 하나도 만들지 않는다. `status`는 답의 **모양**을 바꾸므로, 클라이언트가 조용히
하나를 고르면 기존 호출자 전부를 대신해 다른 질문을 고르는 것이 된다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch, base L189) | 요청 경로 분기 | 없음 | — | 자체 실행 |
| B2 (case, base L190) | `/oauth2/token` | 토큰 응답 | — | 자체 실행 |
| B3 (case, base L192) | `/api/v1/orders` | 빈 페이지 응답 | — | 자체 실행 |
| B4 (range, base L194) | 여섯 키 순회 | 없음 | — | 자체 실행 |
| B5 (if, base L195) | 키가 비어 있지 않음 | 없음 | `t.Errorf` — 클라이언트가 값을 지어냈다 | 자체 실행 |
| B6 (case, base L200) | default | 없음 | `http.NotFound` | 자체 실행 |
| B7 (if, base L215) | `Orders` 오류 | 없음 | `t.Fatal` | 자체 실행 |
| B8 (if, base L218) | 결과가 0건이 아님 | 없음 | `t.Fatalf` | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httptest.NewServer` | 브로커 대역 | `defer srv.Close()` | ast.json calls/defers |
| `New` + Option 3종 | 테스트 클라이언트 | — | ast.json calls |
| `c.Orders` | 측정 대상(기존 읽기) | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: 개명 + 설명 추가 + 단언 1건 추가. base의 단언 8개는 그대로 남았다.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉) — 다만 **이 이름 자체가 사고의
  원인이었다.** 재는 대상이 High-risk이고, 이름이 "빈 필터는 쓸 수 있는 호출"이라는
  허가로 읽혔다. 이름을 고친 것이 이 change의 안전 조치 중 하나다.
