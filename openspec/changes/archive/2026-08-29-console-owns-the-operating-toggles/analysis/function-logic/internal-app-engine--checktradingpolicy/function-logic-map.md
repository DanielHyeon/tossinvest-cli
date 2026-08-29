# Function Logic Map: `checkTradingPolicy`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

검사 대상에 `place`와 `cancel`을 더했다. 이 절이 처음부터 주장하던 것(청산 불가 조합 금지)을 실제로 검사하게 만든 것이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| policy.Place | true/false | config.Trading | false면 trading.place를 열거하고 거부 |
| policy.Cancel | true/false | config.Trading | false면 trading.cancel을 열거하고 거부 |

## Branches and early returns

추가된 분기 둘은 !policy.Place와 !policy.Cancel이다. 각각 interlock_test.go의 새 케이스가 덩는다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/interlock.go:568 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/interlock.go:571 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/app/engine/interlock.go:574 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/app/engine/interlock.go:577 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/app/engine/interlock.go:580 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| strings.Join | 미충족 목록을 한 줄로 | 오류 없음 | 기존 |

## State mutations and fallbacks

- 없음 — 순수 함수다.

## Safety conclusion

- Safe edit boundary: 분기 둘을 더한 것. 기존 둘의 조건·순서·문구는 무수정이고, 방향은 fail-closed(요구를 늘림)이다.
- High-risk impact: yes
