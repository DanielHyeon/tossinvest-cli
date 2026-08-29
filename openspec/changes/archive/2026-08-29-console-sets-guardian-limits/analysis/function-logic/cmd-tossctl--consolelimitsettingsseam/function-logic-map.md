# Function Logic Map: `consoleLimitSettingsSeam`

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
| B1 | `newLimitSettingsSeam(root) != nil` | 없음 | seam 또는 nil | `TestWithoutASeamTheLimitEditorRefusesRatherThanPretends` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newLimitSettingsSeam` | 한도 편집 seam 생성 | nil 가능 | CodeGraph + AST |

## State mutations and fallbacks

- 신규 함수. `consoleSettingsSeam`과 같은 모양이고 같은 이유로 존재한다 — 구체 포인터에서 nil을 판정해 typed-nil이 인터페이스에 들어가지 않게 한다.
- 돌려주는 seam의 Save는 `config.GuardianLimits`를 받는다. 그 타입에 `enabled` 필드가 없다는 것이 콘솔이 게이트를 못 여는 이유이고, 이 배선이 그 타입을 바꾸지 않는다.

## Safety conclusion

- Safe edit boundary: 함수 전체(신규). 반환 타입이 `console.LimitSettings`여야 하고, 그 인터페이스의 Save 인자 타입은 정적 테스트가 못박는다.
- High-risk impact: yes(간접) — 이 배선이 없으면 편집 표면이 없고, 잘못 배선하면 다른 config 파일을 편집한다. `configServiceFor`를 공유하므로 개요·편입·한도가 같은 파일을 본다.
