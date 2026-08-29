# Function Logic Map: `TestTheCeilingIsNotTheInterlock`

- Source: `internal/config/limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 상한의 10배 블록 | 상한 초과, 그러나 양수·유한 | 상한에서 파생 | 위반 목록이 비면 `t.Errorf` |
| 같은 블록의 인터록 판정 | 통과여야 한다 | `GuardianLimits.Validate` | 거부하면 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | USD 상한 조회 오류 | 없음 | `t.Fatalf` | `TestAnUnregisteredCurrencyFailsClosed` |
| B2 | 위반 목록이 비어 있음 | 없음 | `t.Errorf` | 자기 자신 |
| B3 | 인터록이 거부함 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GuardianCeiling` | 기준 블록 | 미등록 통화는 error | CodeGraph + AST |
| `CeilingViolations` | 콘솔 쓰기 경로 판정 | 목록을 돌려준다 | CodeGraph + AST |
| `GuardianLimits.Validate` | 기동 인터록 판정 | error를 돌려준다 | CodeGraph + AST |

## State mutations and fallbacks

- 상태 변경 없음. 신규 테스트다.
- 고정하는 사실은 design D4다: 상한은 **콘솔 쓰기 경로 전용**이고 인터록에는 상한 개념이 없다. 두 판정이 같은 블록에 반대 답을 내는 것이 정상이라는 것을 코드로 남긴다.
- 이것을 적어 두지 않으면 다음 사람이 "상한이 낮으니 시스템 전체가 그 아래"라고 읽고, 손으로 편집한 config.json에도 같은 보호가 있다고 결론낸다. 없다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 10배 계수(임의의 초과량이면 된다).
- High-risk impact: yes — 두 판정이 수렴하면 하나는 반드시 틀린 일을 하게 된다.
