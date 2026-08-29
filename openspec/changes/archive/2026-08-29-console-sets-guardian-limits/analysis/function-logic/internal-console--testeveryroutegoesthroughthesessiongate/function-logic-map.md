# Function Logic Map: `TestEveryRouteGoesThroughTheSessionGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `registeredRoutes(t)` | 패키지 소스에서 파싱한 라우트 표 | `console.go` + `overview.go` 등 | 파싱이 멈추면 하한이 카나리아 |
| 라우트 수 하한 | 22 | 이 테스트 | 실제보다 낮으면 카나리아가 죽어도 모른다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 라우트 순회 | 없음 | 없음 | 자기 자신 |
| B2 | `!r.Session` | 없음 | `t.Errorf` | 자기 자신 |
| B3 | `len(routes) < 22` | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | 라우트 표 파싱 | 실패는 t.Fatal | CodeGraph + AST |

## State mutations and fallbacks

- 이 change가 바꾼 것: 하한 20→22와 열거 주석에 한도 편집 2줄 추가.
- 하한은 "검사가 파싱을 멈췄다"의 카나리아다. 실제 등록 수를 따라가지 않으면 스캐너가 절반만 읽어도 통과한다.

## Safety conclusion

- Safe edit boundary: 하한 숫자와 주석. 라우트를 늘리는 change는 반드시 함께 올려야 한다.
- High-risk impact: yes(가드) — 이 하한이 낮으면 세션 게이트 없는 라우트가 침묵으로 통과한다.
