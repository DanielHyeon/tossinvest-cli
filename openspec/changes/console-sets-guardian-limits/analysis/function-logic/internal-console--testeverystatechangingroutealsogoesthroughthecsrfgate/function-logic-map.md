# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `stateChanging` map | 상태변경 라우트 전부 | 스펙 본문의 열거 | 누락은 CSRF 없는 행위 라우트 |
| `registeredRoutes(t)` | 파싱된 표 | 패키지 소스 | 파싱 실패는 다른 테스트가 잡는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 라우트 순회 | 없음 | 없음 | 자기 자신 |
| B2 | switch 판정 | 없음 | 없음 | 자기 자신 |
| B3 | 상태변경인데 CSRF 없음 | 없음 | `t.Errorf` | 자기 자신 |
| B4 | 읽기인데 CSRF 있음 | 없음 | `t.Errorf` | 자기 자신 |
| B5 | 목록 순회 | 없음 | 없음 | 자기 자신 |
| B6 | 목록에 있는데 미등록 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | 라우트 표 파싱 | 실패는 t.Fatal | CodeGraph + AST |

## State mutations and fallbacks

- 이 change가 바꾼 것: `stateChanging`에 `/settings/limits`·`/settings/limits/preset` 두 항목 추가.
- B4가 중요하다. 두 라우트를 목록에 넣지 않으면 "읽기 라우트가 CSRF 게이트 뒤에 있다"로 실패했다 — 실제로 RED에서 그 문장으로 실패했고, 그것이 목록과 코드가 함께 움직이도록 강제하는 지점이다.

## Safety conclusion

- Safe edit boundary: map 항목 2개. 스펙 본문의 열거와 같은 커밋에서 움직여야 한다.
- High-risk impact: yes(가드) — CSRF 없는 한도 저장 라우트는 페이지에 박힌 이미지로 노출 상한을 바꿀 수 있게 만든다.
