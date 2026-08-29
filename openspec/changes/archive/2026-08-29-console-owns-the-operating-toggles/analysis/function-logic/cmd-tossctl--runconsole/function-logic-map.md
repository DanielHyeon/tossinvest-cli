# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

console.Options에 TradingPolicy와 Gate seam 주입 두 줄이 추가됐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root | *rootOptions | CLI | config 경로 미해결 시 seam이 nil로 남고 화면이 설명한다 |

## Branches and early returns

분기는 전부 기존 기동 경로다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ cmd/tossctl/console.go:164 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ cmd/tossctl/console.go:173 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ cmd/tossctl/console.go:177 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ cmd/tossctl/console.go:181 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ cmd/tossctl/console.go:185 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ cmd/tossctl/console.go:190 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `if` @ cmd/tossctl/console.go:202 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `else` @ cmd/tossctl/console.go:204 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| consoleTradingPolicySeam | 구체 타입을 인터페이스로, typed-nil 회피 | 없음 | console.go |
| consoleGateSwitchSeam | 동상 | 없음 | console.go |

## State mutations and fallbacks

- 없음 — 구조체 리터럴 필드 둘.

## Safety conclusion

- Safe edit boundary: 주입 두 줄. 기존 기동 순서·오류 처리 무수정.
- High-risk impact: yes
