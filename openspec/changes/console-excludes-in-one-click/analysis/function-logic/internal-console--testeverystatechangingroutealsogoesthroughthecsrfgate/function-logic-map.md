# Function Logic Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `stateChanging` map | 상태변경 라우트 10개 | 스펙 본문의 열거 | 빠진 라우트는 "CSRF 뒤의 읽기 라우트"로 오판되어 실패한다 |
| `registeredRoutes(t)` | 실제 표 | AST 추출기 | 목록에만 있고 등록되지 않은 경로는 마지막 루프가 잡는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range routes` | 없음 | — | 이 테스트 |
| B2 | `switch` 진입 | 없음 | 아래 두 case | 이 테스트 |
| B3 | 상태변경인데 게이트 없음 | 없음 | 실패 | 이 테스트 |
| B4 | 읽기인데 게이트 뒤 | 없음 | 실패 — 열 수 없는 화면 | 이 테스트 |
| B5 | `range stateChanging` | 없음 | — | 이 테스트 |
| B6 | 목록에 있으나 미등록 | 없음 | 실패 | 이 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `registeredRoutes` | 라우트 표 + 게이트 판정 | — | CodeGraph + AST |

## State mutations and fallbacks

- 이 change의 변경은 map에 `"/settings/exclude": true` **1개 원소 추가**다.
- 이 목록과 `consoleStateChanging`은 같은 집합이어야 하며 서로 다른 질문을 한다 — 하나는 게이트 여부, 다른 하나는 행위 논증 여부.

## Safety conclusion

- Safe edit boundary: map 1개 원소.
- High-risk impact: no — 테스트다. 다만 이 목록이 실제보다 짧으면 CSRF 없는 config 쓰기 라우트가 통과한다.
