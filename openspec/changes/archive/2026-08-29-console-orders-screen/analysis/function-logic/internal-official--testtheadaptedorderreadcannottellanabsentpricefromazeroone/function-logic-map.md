# Function Logic Map: `TestTheAdaptedOrderReadCannotTellAnAbsentPriceFromAZeroOne`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L484-503). 원문 보존 읽기가 왜 존재하는지를 **측정으로** 고정한다:
같은 픽스처를 기존 `Client.Orders`로 읽으면 null 가격과 null 체결이 **둘 다 0**으로 온다.

`Orders`에 대한 불평이 아니다 — float64는 포트폴리오를 인쇄하는 CLI에 맞는 모양이고
기존 호출자 전부가 정확히 그것에 의존한다. 이 테스트는 "브로커가 아무 값도 보내지 않았다"를
0이 아닌 무엇으로 렌더해야 하는 화면은 그 위에 지을 수 없다는 **사실**을 남긴다.
동시에 `Orders`의 현재 행동을 고정하는 회귀 가드이기도 하다 — 그것이 의도적으로 바뀌면
원문 읽기의 존재 이유도 함께 바뀐다는 문장이 실패 메시지에 들어 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 서버 | `ordersRawServer(t, false, "")` — 미체결 1건(price/execution null) + 체결 1건(price "0") | 같은 파일 헬퍼 | `defer srv.Close()` |
| 클라이언트 | `ordersRawClient(t, srv)` | 같은 파일 헬퍼 | — |
| 호출 | `Orders(ctx, OrdersFilter{})` | 이 테스트 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L489) | `Orders` 오류 | 없음 | `t.Fatal` | 자체 실행 |
| B2 (if, L492) | 건수 ≠ 2 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 (if, L495) | `Price`가 둘 다 0이 아님 | 없음 | `t.Fatalf` — `Orders`가 더 이상 null을 0으로 접지 않는다면 원문 읽기의 존재 이유가 함께 바뀌었다 | 자체 실행 |
| B4 (if, L499) | `AverageExecutionPrice`가 둘 다 0이 아님 | 없음 | `t.Fatalf` | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ordersRawServer` / `ordersRawClient` | 픽스처 | `defer srv.Close()` | ast.json calls/defers |
| `Client.Orders` | 측정 대상(기존 읽기) | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **yes** — 계좌 게이트웨이의 기존 읽기가 **바뀌지 않았음**을 고정하는
  회귀 가드다. 이 change의 additive 주장(기존 `Orders`의 사상 동작 무변경)이 서는 지점이며,
  방향도 옳다: 여기서 실패한다는 것은 누군가 `Orders`의 십진수 사상을 건드렸다는 뜻이고
  그것은 CLI·MCP 직렬화 계약의 변경이다.
