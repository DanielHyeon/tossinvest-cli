# Function Logic Map: `TestEveryRouteGoesThroughTheSessionGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `registeredRoutes(t)` | 패키지 전 파일에서 추출한 라우트 | AST 추출기 | 추출기가 멈추면 개수 하한이 카나리아다 |
| 라우트 수 하한 | 실제 등록 수(현재 20) | 주석의 열거와 합이 같아야 한다 | 하한이 실제보다 낮으면 카나리아가 죽어도 알 수 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range routes` | 없음 | — | 이 테스트 |
| B2 | `!r.Session` | 없음 | 실패 — 세션 게이트 없는 라우트 | 이 테스트 |
| B3 | `len(routes) < 20` | 없음 | 실패 — 추출기가 표를 다 못 읽었다 | 이 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | 라우트 표 추출 | 파싱 실패는 B3이 잡는다 | CodeGraph + AST |

## State mutations and fallbacks

- 이 change의 변경은 하한 19→20과 주석 열거의 "2 its two edits"→"3 its three edits"다.
- 숫자와 열거를 함께 고치는 것이 규칙이다 — 합이 맞지 않는 주석은 다음 독자가 대조를 그만두게 만든다.

## Safety conclusion

- Safe edit boundary: 상수 1개와 그 근거 주석.
- High-risk impact: no — 테스트다.
