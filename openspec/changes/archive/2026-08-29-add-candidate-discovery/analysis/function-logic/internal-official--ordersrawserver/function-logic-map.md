# Function Logic Map: `ordersRawServer`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트 헬퍼**다(HEAD L427-457). 한 페이지에 주문 두 건을 담은 `httptest` 서버를
만든다 — 브로커가 null을 보낸 미체결 시장가 주문 하나와, 평균가가 **진짜로** 0인 체결 주문
하나. 이 대비가 원문 보존 읽기 테스트 전부의 기반이다: 부재와 0을 같은 응답 안에 나란히
두어야 "구분한다"를 잴 수 있다.

`hasNext`/`nextCursor`를 파라미터로 받아 잘린 페이지도 만들 수 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `hasNext` | bool | 호출 테스트 | `boolText`로 JSON 리터럴화 |
| `nextCursor` | 빈 문자열이면 JSON `null` | 호출 테스트 | — |
| 응답 픽스처 | 미체결(`price`/`execution` null) + 체결(`price:"0"`, `averageFilledPrice:"0"`) | 이 헬퍼 | — |

불변식: `execution`을 **객체 전체 null**로 보낸다 — API가 살아 있는 주문에 대해 실제로 하는
일이고, 필드별 null만으로는 그 경우를 재지 못한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L430) | `nextCursor != ""` | JSON 커서를 인용 문자열로 | — | `TestTheRawOrderReadReportsThatThePageWasTruncated` |
| B2 (switch, L434) | 요청 경로 분기 | 없음 | — | 호출 테스트 4건 |
| B3 (case, L435) | `/oauth2/token` | 토큰 응답 | — | 동상 |
| B4 (case, L437) | `/api/v1/orders` | 두 건 페이지 응답 | — | 동상 |
| B5 (case, L453) | default | 없음 | `http.NotFound` | 동상 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

서버 수명은 호출 테스트가 `defer srv.Close()`로 소유한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` | 실패 위치를 호출부로 | — | ast.json calls |
| `httptest.NewServer` / `http.HandlerFunc` | 브로커 대역 | — | ast.json calls |
| `boolText` | `hasNext`의 JSON 리터럴 | 순수 | ast.json calls |
| `http.NotFound` | 예상 밖 경로 | — | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 헬퍼 가산.
- High-risk impact: **no** — 테스트 전용, 실계좌·실브로커 무접촉. 다만 이 픽스처의
  **충실도**가 위 테스트들의 값을 결정한다: `execution`을 객체 전체 null로 보내는 것과
  진짜 0 행을 나란히 두는 것이 그 충실도의 두 축이다.
