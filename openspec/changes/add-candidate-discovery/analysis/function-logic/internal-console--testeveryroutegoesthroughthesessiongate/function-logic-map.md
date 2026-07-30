# Function Logic Map: `TestEveryRouteGoesThroughTheSessionGate`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

모든 라우트가 `session0`을 거친다는 확인. 세션 토큰이 이 콘솔의 전부 — 터미널 점유가 인증이다. `session0` 없이 등록된 핸들러는 이 기계에서 소켓을 열 수 있는 무엇이든(개발자 노트북이면 모든 브라우저 탭과 모든 에이전트) 도달할 수 있다.

하한 19는 **카나리아**이고 완전성 증명이 아니다 — '추출기가 파싱을 멈췄다'를 잡는다. 한때 목록은 16개인데 단언은 17이었고(설정 화면이 목록에서 빠져 있었다) 그 뒤로 숫자와 목록을 주석에 함께 적어 어긋날 수 없게 했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `registeredRoutes(t)` | 표 전체 | 추출기 | 0건이면 추출기가 Fatal |
| 하한 19 | 실제 등록 수를 따라간다 | 주석의 열거 | 실제보다 낮으면 카나리아가 죽어도 알 수 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, r := range routes` | 없음 | 없음 | 표 19건 |
| B2 | `!r.Session` | `t.Errorf` | 없음 | 래퍼 제거 변이 |
| B3 | `len(routes) < 19` | `t.Errorf` | 없음 | 추출기 축소 변이 — 파일 범위를 `console.go`로 되돌리면 16으로 떨어진다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles(t)` | 패키지의 **비테스트** `.go` 파일 전부를 읽는다 | 0개면 `t.Fatal` — 가드가 패키지를 보고 있지 않다는 카나리아 | static_test.go:36 |
| `parseFile` | `go/parser`로 파싱 | 실패하면 `t.Fatalf` | static_test.go:60 |
| (런타임 무접촉) | 가드는 **선언**을 읽는다. 실행되지 않는 코드가 정확히 이 가드가 잡아야 하는 것이다 | reflect·실행 경로를 쓰지 않는다 | static_test.go 파일 주석 |
| `registeredRoutes` | 표 추출 | 0건이면 Fatal | static_test.go:127 |

## State mutations and fallbacks

- 없음(판정 전용).

## Safety conclusion

- Safe edit boundary: 하한을 실제 수(19)로 올리고 열거를 주석에 고정. 판정식은 무변경.
- High-risk impact: yes (인증 경로 감시)
