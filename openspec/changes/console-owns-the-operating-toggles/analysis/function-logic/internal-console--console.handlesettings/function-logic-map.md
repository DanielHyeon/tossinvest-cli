# Function Logic Map: `Console.handleSettings`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

두 seam의 상태를 page에 싫는 블록이 추가됐다. 기존 세 블록은 무수정.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| c.opts.TradingPolicy | nil 또는 seam | 주입 | nil이면 TradingWired=false로 화면이 설명한다 |
| c.opts.Gate | nil 또는 seam | 주입 | nil이면 GateWired=false |

## Branches and early returns

추가된 분기는 seam nil 검사 둘과 Load 오류 하나다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/console/settings.go:131 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/console/settings.go:134 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/console/settings.go:139 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/console/settings.go:142 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/console/settings.go:147 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/console/settings.go:150 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| TradingPolicy.Load | 거래 정책 원문 | 오류는 TradingLoadErr로 격리 | settings_operating.go |

## State mutations and fallbacks

- 없음 — 읽기만 한다. 게이트 seam은 write-only라 Load를 부르지 않는다.

## Safety conclusion

- Safe edit boundary: 블록 두 개 추가. 오류 격리는 기존 세 섹션과 같은 모양이다.
- High-risk impact: no
