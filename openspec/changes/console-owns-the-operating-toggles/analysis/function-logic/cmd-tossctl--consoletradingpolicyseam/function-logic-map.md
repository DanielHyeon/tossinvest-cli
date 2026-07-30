# Function Logic Map: `consoleTradingPolicySeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

신규 어댑터. 구체 포인터가 nil일 때 인터페이스 nil을 돌려준다 — typed-nil은 배선된 것처럼 보이고 화면이 panic하는 컨트롤을 그린다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root | *rootOptions | CLI | config 경로 미해결 → nil |

## Branches and early returns

분기 하나 — 구체 포인터의 nil 검사.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ cmd/tossctl/console.go:324 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| new*Seam | 구체 seam 생성 | nil 반환 | operatingsettings.go |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 신규 함수. 선례(consoleLimitSettingsSeam)와 동일한 모양.
- High-risk impact: no
