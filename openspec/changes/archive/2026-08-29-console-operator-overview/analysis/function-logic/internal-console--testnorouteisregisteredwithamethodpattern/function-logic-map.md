# Function Logic Map: `TestNoRouteIsRegisteredWithAMethodPattern`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

Go 1.22의 `HandleFunc("GET /dashboard", …)`은 합법이고 이 화면이 무엇인지 말하는 자연스러운 방법이다. 여기서는 쓰면 안 된다 — 추출기가 리터럴을 **그대로 경로로** 읽으므로 표에 `GET /dashboard`가 들어가고 경로 키 대조를 하는 가드가 한꺼번에 조용히 어긋난다: 읽기 라우트 목록은 화면이 미등록이라고 보고하고, 계좌 동사 스캔은 경로가 아닌 문자열을 뒤지고, CSRF 짝 대조는 아무것도 일치하지 않을 이름과 비교한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Path` | `/`로 시작하고 공백·탭이 없어야 한다 | 추출기 | 위반이면 `t.Errorf`에 파일 이름을 함께 적는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, r := range registeredRoutes(t)` | 없음 | 없음 | 표 19건 |
| B2 | `strings.ContainsAny(r.Path, " \t") || !strings.HasPrefix(r.Path, "/")` | `t.Errorf` | 없음 | 메서드 패턴 등록 변이 |

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

- Safe edit boundary: 신설. 추출기가 메서드 패턴을 분해하게 되는 change가 올 때까지, 이 형태로 소리 내어 실패한다.
- High-risk impact: yes (라우트 게이트 감시)
