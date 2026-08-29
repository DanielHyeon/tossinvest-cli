# Function Logic Map: `routeTableFindings`

- Source: `internal/console/static_test.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

`routeFindings`를 표 전체에 적용하고, 라우트가 아니라 **표**에 대한 주장 하나를 더한다: 상태변경 목록이 이름을 부른 경로가 실제로 등록돼 있다. 아무도 등록하지 않는 라우트를 부르는 목록은 이 콘솔을 더 이상 기술하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `routes` | 표 전체 | 추출기 | 해당 없음 |
| `consoleStateChanging` | 9경로 | static_test.go:606 | 등록되지 않은 이름이 있으면 발견 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, r := range routes` | `seen` 구성 + 발견 누적 | 없음 | `TestNoRouteNamesAnAccountMutation` |
| B2 | `for _, path := range consoleStateChanging` | 없음 | 없음 | 같은 위 |
| B3 | `!seen[path]` | 발견 추가 | 없음 | 목록에서 경로를 지우거나 등록을 지우는 변이 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `routeFindings` | 라우트별 판정 | 순수 | static_test.go:712 |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: 신설(추출). 표 수준 주장(등록 존재 확인)이 함께 들어왔다.
- High-risk impact: yes (계좌 라우트 부재 보증의 표 수준 절반)
