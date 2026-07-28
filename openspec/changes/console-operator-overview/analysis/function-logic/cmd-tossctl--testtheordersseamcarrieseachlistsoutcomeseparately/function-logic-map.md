# Function Logic Map: `TestTheOrdersSeamCarriesEachListsOutcomeSeparately`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (revision=current, L739–803, 분기 13개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

세 콜 중 하나를 잃어도 나머지를 데려가지 않는다는 것을 고정한다. 조건주문에 429를 주입하고, 미체결·종결이 살아남고 `ConditionalError`만 채워지는지 본다.

세 실패를 하나의 에러로 접으면 답한 부분까지 화면에서 사라지고, 화면에는 정직하게 말할 것이 남지 않는다. 마지막 단언은 별개의 주장이다 — `price:null`·`execution:null`이 **0이 아니라 부재**로 도착해야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 조건주문 엔드포인트 | 429 | httptest 핸들러 | `ConditionalError`가 비면 FAIL |
| 미체결/종결 응답 | 각 1건, 종결은 `hasNext:true` | 동일 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 라우팅 switch | — | — | 이 테스트 |
| B2 | 토큰 엔드포인트 | — | 토큰 응답 | 동일 |
| B3 | `status=OPEN` | — | 미체결 1건 | 동일 |
| B4 | 그 외 `/api/v1/orders` (=CLOSED) | — | 종결 1건 + `hasNext:true` | 동일 |
| B5 | 조건주문 | — | 429 | 동일 |
| B6 | 그 외 | — | 404 | 동일 |
| B7 | `Orders`가 에러 반환 | — | `t.Fatalf` — 반쪽 판독에 에러를 주면 안 된다 | 동일 |
| B8 | 미체결이 살아남지 못함 | — | `t.Fatalf` | 동일 |
| B9 | `OpenError`가 채워짐 | — | `t.Errorf` | 동일 |
| B10 | 종결이 살아남지 못함 | — | `t.Fatalf` | 동일 |
| B11 | `ClosedTruncated`가 거짓 | — | `t.Error` | 동일 |
| B12 | `ConditionalError`가 빔 | — | `t.Error` — 0건이 측정값처럼 렌더된다 | 동일 |
| B13 | null price/execution이 숫자로 도착 | — | `t.Errorf` | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleOrdersSeam(newConsoleBroker(&rootOptions{})).Orders(ctx)` | 실제 seam | 3콜, 결과 개별 귀속 | console.go L469 |

## State mutations and fallbacks

- 테스트 — httptest 서버, 실계좌 접촉 없음.

## Safety conclusion

- Safe edit boundary: 개별 귀속 기대값. 세 결과를 하나의 에러로 접는 편집이 이 테스트를 깬다.
- High-risk impact: yes (주문 경로) — 조건주문 실패가 삼켜지면 노출 상한을 채우고 있는 잔여물이 "0건"으로 보인다. 실계좌 부작용은 없다.
