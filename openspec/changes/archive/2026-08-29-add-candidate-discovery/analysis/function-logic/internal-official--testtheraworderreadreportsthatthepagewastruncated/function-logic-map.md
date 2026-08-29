# Function Logic Map: `TestTheRawOrderReadReportsThatThePageWasTruncated`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L582-597). `Client.Orders`는 `nextCursor`와 `hasNext`를 디코딩한 뒤
**둘 다 버린다**. 그래서 첫 페이지를 센 호출자는 두 번째 페이지가 있는지 알 방법이 없다.
잘린 페이지에서 센 건수는 자신 있게 짧은 숫자이고, 이 화면에서 그것은 노출 상한을 채우고
있는 잔여물이 사라진다는 뜻이다.

서버를 `ordersRawServer(t, true, "cursor-2")`로 세워 브로커가 후속 페이지를 말한 경우를
만든다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 서버 | `hasNext=true`, `nextCursor="cursor-2"` | 같은 파일 헬퍼 | `defer srv.Close()` |
| 호출 | `OrdersRaw(ctx, OrdersFilter{Status: "OPEN"})` | 이 테스트 | — |
| 기대 | `HasNext == true`, `NextCursor == "cursor-2"` | 이 테스트 | `t.Error`/`t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L587) | 읽기 오류 | 없음 | `t.Fatal` | 자체 실행 |
| B2 (if, L590) | `HasNext`가 false | 없음 | `t.Error` — 건수가 확정된 숫자로 렌더된다 | 자체 실행 |
| B3 (if, L594) | `NextCursor` 불일치 | 없음 | `t.Errorf` | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ordersRawServer(t, true, "cursor-2")` | 잘린 페이지 픽스처 | `defer srv.Close()` | ast.json calls/defers |
| `Client.OrdersRaw` | 측정 대상 | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉). 재는 성질은 High-risk다 —
  페이지 경계 유실은 잔여물을 조용히 숨기는 실패이고, 이 화면의 건수는 "N건 이상"으로
  렌더될 수 있어야 한다.
