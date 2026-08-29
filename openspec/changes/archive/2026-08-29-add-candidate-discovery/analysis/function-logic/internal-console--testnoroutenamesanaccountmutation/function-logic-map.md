# Function Logic Map: `TestNoRouteNamesAnAccountMutation`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

판정을 실제 표에 적용한다. 본문은 세 줄이고 판단은 전부 `routeTableFindings`에 있다 — 그 추출이 예외를 측정 가능하게 만든 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `registeredRoutes(t)` | 표 전체 | 추출기 | 0건이면 Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, finding := range routeTableFindings(registeredRoutes(t))` | `t.Error` | 없음 | 발견이 하나라도 있으면 실패 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles(t)` | 패키지의 **비테스트** `.go` 파일 전부를 읽는다 | 0개면 `t.Fatal` — 가드가 패키지를 보고 있지 않다는 카나리아 | static_test.go:36 |
| `parseFile` | `go/parser`로 파싱 | 실패하면 `t.Fatalf` | static_test.go:60 |
| (런타임 무접촉) | 가드는 **선언**을 읽는다. 실행되지 않는 코드가 정확히 이 가드가 잡아야 하는 것이다 | reflect·실행 경로를 쓰지 않는다 | static_test.go 파일 주석 |
| `routeTableFindings` | 표 판정 | 순수 | static_test.go:754 |

## State mutations and fallbacks

- 없음(판정 전용).

## Safety conclusion

- Safe edit boundary: 본문이 판정 추출로 대체됐다. 적용 대상(실제 표)은 무변경.
- High-risk impact: yes (계좌 라우트 부재 보증의 실제 적용 지점)
