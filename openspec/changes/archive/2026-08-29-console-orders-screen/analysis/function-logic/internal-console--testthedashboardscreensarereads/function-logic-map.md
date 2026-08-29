# Function Logic Map: `TestTheDashboardScreensAreReads`

- Source: `internal/console/static_test.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

포지션·이력·개요·주문·발굴 다섯 화면이 등록돼 있고, 세션 게이트를 거치며, CSRF 게이트 **뒤에 있지 않다**는 확인. POST만 답하는 대시보드는 아무도 열 수 없는 페이지이고, 이 테스트가 그 실패를 반대편에서 잡는다. `want` 맵은 이 change들에서 `/dashboard`·`/orders`·`/signals` 셋이 더해졌다.

**경계**: `want`에 없는 화면은 여기서 검사되지 않는다. 그쪽은 `TestEveryRouteGoesThroughTheSessionGate`와 `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`가 표 전체에 대해 덮는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `want` | 5경로 → CSRF 기대값 false | 테스트 본문 | 누락 경로는 `found` 검사가 잡는다 |
| `registeredRoutes(t)` | 표 전체 | 추출기 | 0건이면 Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, r := range registeredRoutes(t)` | 없음 | 없음 | 표 19건 |
| B2 | `gated, ok := want[r.Path]; !ok` | 없음 | continue | 관심 밖 라우트 |
| B3 | `r.CSRFGated != gated` | `t.Errorf` | 없음 | 읽기 화면을 `mutating`으로 감싸는 변이 |
| B4 | `!r.Session` | `t.Errorf` | 없음 | `session0` 제거 변이 |
| B5 | `for path := range want` | 없음 | 없음 | 다섯 화면 |
| B6 | `!found[path]` | `t.Errorf` | 없음 | 화면 등록 제거 변이 — 추출기가 파일을 못 읽으면 여기서도 실패한다 |

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

- Safe edit boundary: `want` 맵에 세 경로 추가. 판정 세 갈래는 무변경.
- High-risk impact: yes (라우트 게이트 감시 — 읽기 화면이 CSRF 게이트 뒤로 들어가면 예외 판정의 전제가 흔들린다)
