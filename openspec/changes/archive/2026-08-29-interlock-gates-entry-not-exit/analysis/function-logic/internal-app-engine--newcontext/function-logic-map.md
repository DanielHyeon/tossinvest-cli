# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

gatewayInputs에 `protectionOverride: opts.protectionOverride` 한 줄이 추가됐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opts.protectionOverride | nil 또는 포인터 | engine.Options — 내보내진 setter 없음 | 프로덕션은 항상 nil |

## Branches and early returns

분기는 전부 기존 기동 경로다. 추가된 대입은 분기 밖에 있다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/engine.go:390 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/engine.go:397 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/app/engine/engine.go:401 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/app/engine/engine.go:412 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/app/engine/engine.go:426 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/app/engine/engine.go:433 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `if` @ internal/app/engine/engine.go:442 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `if` @ internal/app/engine/engine.go:459 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B9 | `if` @ internal/app/engine/engine.go:477 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B10 | `if` @ internal/app/engine/engine.go:482 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| buildGateway | 실행 스택 구성 | 오류면 journal을 닫고 refuseStartup | 기존 |

## State mutations and fallbacks

- 없음 — 이 줄은 구조체 리터럴의 필드 하나다.

## Safety conclusion

- Safe edit boundary: 전달 한 줄. 기동 순서·오류 처리·journal 수명주기는 무수정.
- High-risk impact: yes
