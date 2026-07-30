# Function Logic Map: `consoleSettingsSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | nil 가능 | 호출자 | 구체 포인터가 nil이면 인터페이스 nil을 돌려준다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `newAdoptionSettingsSeam(root) != nil` | 없음 | seam 또는 nil | `TestWithoutASeamNeitherControlRenders`(콘솔 측) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newAdoptionSettingsSeam` | 편입 seam 생성 | nil 가능 | CodeGraph + AST |

## State mutations and fallbacks

- 무변경. 이 함수는 이 change에서 한 글자도 바뀌지 않았고, 바로 아래에 같은 모양의 `consoleLimitSettingsSeam`이 추가되면서 diff hunk에 함께 들어왔다.
- typed-nil 회피가 이 함수의 전부다: 구체 포인터를 인터페이스에 넣은 뒤 nil 비교하면 항상 false가 되어 화면이 배선된 것처럼 보인다.

## Safety conclusion

- Safe edit boundary: 없음(무변경).
- High-risk impact: no.
