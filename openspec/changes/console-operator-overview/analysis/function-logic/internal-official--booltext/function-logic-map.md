# Function Logic Map: `boolText`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트 헬퍼**다(HEAD L459-464). `ordersRawServer`가 조립하는 JSON 문자열에
`hasNext`의 리터럴을 넣기 위한 2분기 순수 함수다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `b` | `true`/`false` | `ordersRawServer`의 파라미터 | 해당 없음 |

불변식: 순수 함수. JSON 리터럴이므로 `"true"`/`"false"` 소문자여야 한다
(`strconv.FormatBool`과 같은 결과이며, 이 파일은 `strconv`를 import하지 않는다).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L460) | `b == true` | 없음 | `"true"`; 아니면 꼬리에서 `"false"` | `TestTheRawOrderReadReportsThatThePageWasTruncated`(true) / 나머지 raw 테스트(false) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls = null |

## State mutations and fallbacks

- 없음. 순수 함수이고 테스트 파일 안에만 있다.

## Safety conclusion

- Safe edit boundary: 신규 leaf 헬퍼 가산.
- High-risk impact: **no** — 테스트 전용 순수 함수. 계좌·주문·원장 무접촉.
