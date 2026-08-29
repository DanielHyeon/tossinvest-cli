# Function Logic Map: `New`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

생성자에 `protectionOverride` 한 필드를 전달하는 줄이 추가됐다. 검증 분기와 기본값 채우기는 무수정.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opts.ProtectionOverrideForTest | nil 또는 WIRED/UNWIRED 포인터 | 호출자 (프로덕션은 항상 nil) | nil이면 defaultProtection이 답한다 |

## Branches and early returns

분기는 전부 기존 생성자 검증이며 이 change가 만들지 않았다. 추가된 것은 분기 없는 대입 한 줄이다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` @ internal/execgw/gateway.go:146 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `case` @ internal/execgw/gateway.go:147 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `case` @ internal/execgw/gateway.go:149 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `case` @ internal/execgw/gateway.go:151 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/execgw/gateway.go:170 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/execgw/gateway.go:173 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `if` @ internal/execgw/gateway.go:176 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음 — 신규 호출 없음) | 필드 대입만 추가 | 해당 없음 | AST |

## State mutations and fallbacks

- Gateway.protectionOverride에 opts 값을 복사한다. 기존 필드 대입은 무수정.

## Safety conclusion

- Safe edit boundary: 필드 추가와 그 전달. 기존 검증 분기(journal/trading/accountRef nil)는 손대지 않았다.
- High-risk impact: yes
